// Package email provides SMTP email sending functionality.
package email

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

// ExpiryReminderData holds the data for an expiry reminder email.
type ExpiryReminderData struct {
	MerchantName string // e.g. "IKEA"
	ResourceType string // "voucher" or "gift_card"
	DaysLeft     int    // e.g. 3
	ExpiresAt    string // formatted expiry date, e.g. "26. Februar 2026"
	Code         string // voucher code or gift card number
	Value        string // formatted value, e.g. "20%" or "CHF 50.00"
	ResourceURL  string // link to view in Savvy, e.g. "https://savvy.example.com/vouchers/uuid"
}

// ValidityStartData holds the data for a validity start notification email.
type ValidityStartData struct {
	MerchantName string // e.g. "IKEA"
	ValidFrom    string // formatted date, e.g. "1. März 2026"
	Code         string // voucher code
	Value        string // formatted value, e.g. "20%" or "CHF 50.00"
	ResourceURL  string // link to view in Savvy
}

// ServiceInterface defines the contract for sending emails.
type ServiceInterface interface {
	SendPasswordReset(ctx context.Context, toEmail, toName, resetURL, expiresIn, language string) error
	SendEmailVerification(ctx context.Context, toEmail, toName, verifyURL, language string) error
	SendAccountDeletionConfirmation(ctx context.Context, toEmail, toName, language string) error
	SendExpiryReminder(ctx context.Context, toEmail, toName string, data ExpiryReminderData, unsubscribeURL, language string) error
	SendValidityStart(ctx context.Context, toEmail, toName string, data ValidityStartData, unsubscribeURL, language string) error
	SendShareNotification(ctx context.Context, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language string) error
	SendTransferNotification(ctx context.Context, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language string) error
	SendTestEmail(ctx context.Context, toEmail, toName, language string) error
	// CheckConnection verifies SMTP server connectivity without sending an email.
	CheckConnection(ctx context.Context) error
}

// SMTPConfig holds SMTP configuration.
type SMTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string // #nosec G117 -- struct field name, not a hardcoded secret
	FromEmail   string
	FromName    string
	UseTLS      bool
	FrontendURL string
}

// SMTPEmailService sends emails via SMTP.
type SMTPEmailService struct {
	config    SMTPConfig
	templates *template.Template
	ehloHost  string // EHLO hostname derived from FromEmail domain
}

// NewSMTPEmailService creates a new SMTP email service.
func NewSMTPEmailService(cfg SMTPConfig) (*SMTPEmailService, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	// Derive EHLO hostname from sender email domain (RFC 5321).
	// EHLO must identify the sending host, not the target SMTP server.
	ehloHost := "localhost"
	if parts := strings.SplitN(cfg.FromEmail, "@", 2); len(parts) == 2 {
		ehloHost = parts[1]
	}

	return &SMTPEmailService{
		config:    cfg,
		templates: tmpl,
		ehloHost:  ehloHost,
	}, nil
}

// emailStrings holds localized strings for email templates.
type emailStrings struct {
	Subject string
	Data    map[string]string
}

// getLogoURL returns the full URL to the logo image based on FrontendURL.
func (s *SMTPEmailService) getLogoURL() string {
	// Use FrontendURL as base, fallback to placeholder if not set
	if s.config.FrontendURL == "" {
		return "https://via.placeholder.com/150x50?text=Savvy"
	}
	// Append /logo.png to the base URL
	baseURL := strings.TrimSuffix(s.config.FrontendURL, "/")
	return baseURL + "/logo.png"
}

// passwordResetStrings returns localized strings for password reset emails.
func passwordResetStrings(lang, userName, resetURL, expiresIn string) emailStrings {
	ctx := emailCtx(lang)
	ud := map[string]any{"UserName": userName}
	return emailStrings{
		Subject: et(ctx, "email.password_reset.subject"),
		Data: map[string]string{
			"UserName":     userName,
			"ResetURL":     resetURL,
			"ExpiresIn":    expiresIn,
			"Lang":         lang,
			"Title":        et(ctx, "email.password_reset.title"),
			"Greeting":     et(ctx, "email.common.greeting", ud),
			"Message":      et(ctx, "email.password_reset.message"),
			"ButtonText":   et(ctx, "email.password_reset.button"),
			"ExpiresText":  et(ctx, "email.password_reset.expires", map[string]any{"ExpiresIn": expiresIn}),
			"IgnoreText":   et(ctx, "email.password_reset.ignore"),
			"FallbackText": et(ctx, "email.common.fallback_text"),
			"Footer":       et(ctx, "email.common.footer"),
		},
	}
}

// emailVerificationStrings returns localized strings for email verification emails.
func emailVerificationStrings(lang, userName, verifyURL string) emailStrings {
	ctx := emailCtx(lang)
	ud := map[string]any{"UserName": userName}
	return emailStrings{
		Subject: et(ctx, "email.verification.subject"),
		Data: map[string]string{
			"UserName":     userName,
			"VerifyURL":    verifyURL,
			"Lang":         lang,
			"Title":        et(ctx, "email.verification.title"),
			"Greeting":     et(ctx, "email.common.greeting", ud),
			"Message":      et(ctx, "email.verification.message"),
			"ButtonText":   et(ctx, "email.verification.button"),
			"ExpiresText":  et(ctx, "email.verification.expires"),
			"FallbackText": et(ctx, "email.common.fallback_text"),
			"Footer":       et(ctx, "email.common.footer"),
		},
	}
}

// accountDeletedStrings returns localized strings for account deletion emails.
func accountDeletedStrings(lang, userName string) emailStrings {
	ctx := emailCtx(lang)
	ud := map[string]any{"UserName": userName}
	return emailStrings{
		Subject: et(ctx, "email.account_deleted.subject"),
		Data: map[string]string{
			"UserName":   userName,
			"Lang":       lang,
			"Title":      et(ctx, "email.account_deleted.title"),
			"Greeting":   et(ctx, "email.common.greeting", ud),
			"Message":    et(ctx, "email.account_deleted.message"),
			"IgnoreText": et(ctx, "email.account_deleted.ignore"),
			"Footer":     et(ctx, "email.common.footer"),
		},
	}
}

// testEmailStrings returns localized strings for test emails.
func testEmailStrings(lang, userName string) emailStrings {
	ctx := emailCtx(lang)
	ud := map[string]any{"UserName": userName}
	return emailStrings{
		Subject: et(ctx, "email.test.subject"),
		Data: map[string]string{
			"UserName":         userName,
			"Lang":             lang,
			"Title":            et(ctx, "email.test.title"),
			"Greeting":         et(ctx, "email.common.greeting", ud),
			"Message":          et(ctx, "email.test.message"),
			"StatusTitle":      et(ctx, "email.test.status_title"),
			"StatusLine1":      et(ctx, "email.test.status_line1"),
			"StatusLine2":      et(ctx, "email.test.status_line2"),
			"StatusLine3":      et(ctx, "email.test.status_line3"),
			"SuccessMessage":   et(ctx, "email.test.success"),
			"AutomatedMessage": et(ctx, "email.test.automated"),
			"Footer":           et(ctx, "email.common.footer"),
		},
	}
}

// expiryReminderStrings returns localized strings for expiry reminder emails.
func expiryReminderStrings(lang, userName, unsubscribeURL string, data ExpiryReminderData) emailStrings {
	ctx := emailCtx(lang)
	ud := map[string]any{"UserName": userName}
	resourceName := et(ctx, "push.resource."+data.ResourceType)
	var daysText string
	if data.DaysLeft == 0 {
		daysText = et(ctx, "email.expiry.days_left_today")
	} else {
		daysText = etc(ctx, "email.expiry.days_left", data.DaysLeft)
	}

	// Determine code/value labels based on resource type
	codeLabelKey := "email.expiry.code_label.voucher"
	valueLabelKey := "email.expiry.value_label.voucher"
	if data.ResourceType == "gift_card" {
		codeLabelKey = "email.expiry.code_label.gift_card"
		valueLabelKey = "email.expiry.value_label.gift_card"
	}

	return emailStrings{
		Subject: et(ctx, "email.expiry.subject", map[string]any{"Resource": resourceName, "Merchant": data.MerchantName}),
		Data: map[string]string{
			"UserName":        userName,
			"Lang":            lang,
			"Title":           et(ctx, "email.expiry.title"),
			"Greeting":        et(ctx, "email.common.greeting", ud),
			"Message":         et(ctx, "email.expiry.message", map[string]any{"Resource": resourceName, "Merchant": data.MerchantName, "DaysText": daysText}),
			"MerchantName":    data.MerchantName,
			"ResourceType":    resourceName,
			"ExpiresAt":       data.ExpiresAt,
			"Code":            data.Code,
			"Value":           data.Value,
			"ResourceURL":     data.ResourceURL,
			"DetailLabel":     et(ctx, "email.expiry.detail_label"),
			"TypeLabel":       et(ctx, "email.expiry.type_label"),
			"MerchantLabel":   et(ctx, "email.expiry.merchant_label"),
			"CodeLabel":       et(ctx, codeLabelKey),
			"ValueLabel":      et(ctx, valueLabelKey),
			"ExpiresLabel":    et(ctx, "email.expiry.expires_label"),
			"DaysLeftText":    daysText,
			"DaysLeftLabel":   et(ctx, "email.expiry.days_left_label"),
			"ButtonText":      et(ctx, "email.common.button_view"),
			"IgnoreText":      et(ctx, "email.expiry.ignore"),
			"UnsubscribeURL":  unsubscribeURL,
			"UnsubscribeText": et(ctx, "email.common.unsubscribe_reminders"),
			"Footer":          et(ctx, "email.common.footer"),
		},
	}
}

// validityStartStrings returns localized strings for validity start emails.
func validityStartStrings(lang, userName, unsubscribeURL string, data ValidityStartData) emailStrings {
	ctx := emailCtx(lang)
	ud := map[string]any{"UserName": userName}

	return emailStrings{
		Subject: et(ctx, "email.validity_start.subject", map[string]any{"Merchant": data.MerchantName}),
		Data: map[string]string{
			"UserName":        userName,
			"Lang":            lang,
			"Title":           et(ctx, "email.validity_start.title"),
			"Greeting":        et(ctx, "email.common.greeting", ud),
			"Message":         et(ctx, "email.validity_start.message", map[string]any{"Merchant": data.MerchantName}),
			"MerchantName":    data.MerchantName,
			"ValidFrom":       data.ValidFrom,
			"Code":            data.Code,
			"Value":           data.Value,
			"ResourceURL":     data.ResourceURL,
			"MerchantLabel":   et(ctx, "email.expiry.merchant_label"),
			"CodeLabel":       et(ctx, "email.expiry.code_label.voucher"),
			"ValueLabel":      et(ctx, "email.expiry.value_label.voucher"),
			"ValidFromLabel":  et(ctx, "email.validity_start.valid_from_label"),
			"BadgeText":       et(ctx, "email.validity_start.badge"),
			"ButtonText":      et(ctx, "email.common.button_view"),
			"IgnoreText":      et(ctx, "email.validity_start.ignore"),
			"UnsubscribeURL":  unsubscribeURL,
			"UnsubscribeText": et(ctx, "email.common.unsubscribe_reminders"),
			"Footer":          et(ctx, "email.common.footer"),
		},
	}
}

// shareNotificationStrings returns localized strings for share notification emails.
func shareNotificationStrings(lang, userName, fromName, resourceType, resourceURL, unsubscribeURL string) emailStrings {
	ctx := emailCtx(lang)
	ud := map[string]any{"UserName": userName}
	resourceName := et(ctx, "push.resource."+resourceType)

	return emailStrings{
		Subject: et(ctx, "email.share.subject", map[string]any{"FromName": fromName}),
		Data: map[string]string{
			"Lang":            lang,
			"Title":           et(ctx, "email.share.title"),
			"Greeting":        et(ctx, "email.common.greeting", ud),
			"Message":         et(ctx, "email.share.message", map[string]any{"FromName": fromName, "Resource": resourceName}),
			"ButtonText":      et(ctx, "email.common.button_view"),
			"ResourceURL":     resourceURL,
			"IgnoreText":      et(ctx, "email.share.ignore"),
			"UnsubscribeURL":  unsubscribeURL,
			"UnsubscribeText": et(ctx, "email.common.unsubscribe_notifications"),
			"Footer":          et(ctx, "email.common.footer"),
		},
	}
}

// transferNotificationStrings returns localized strings for transfer notification emails.
func transferNotificationStrings(lang, userName, fromName, resourceType, resourceURL, unsubscribeURL string) emailStrings {
	ctx := emailCtx(lang)
	ud := map[string]any{"UserName": userName}
	resourceName := et(ctx, "push.resource."+resourceType)

	return emailStrings{
		Subject: et(ctx, "email.transfer.subject", map[string]any{"FromName": fromName}),
		Data: map[string]string{
			"Lang":            lang,
			"Title":           et(ctx, "email.transfer.title"),
			"Greeting":        et(ctx, "email.common.greeting", ud),
			"Message":         et(ctx, "email.transfer.message", map[string]any{"FromName": fromName, "Resource": resourceName}),
			"ButtonText":      et(ctx, "email.common.button_view"),
			"ResourceURL":     resourceURL,
			"IgnoreText":      et(ctx, "email.transfer.ignore"),
			"UnsubscribeURL":  unsubscribeURL,
			"UnsubscribeText": et(ctx, "email.common.unsubscribe_notifications"),
			"Footer":          et(ctx, "email.common.footer"),
		},
	}
}

// normalizeLanguage ensures the language is valid, falling back to "en".
func normalizeLanguage(lang string) string {
	switch lang {
	case "de", "en", "fr":
		return lang
	default:
		return "en"
	}
}

// SendPasswordReset sends a password reset email.
func (s *SMTPEmailService) SendPasswordReset(_ context.Context, toEmail, toName, resetURL, expiresIn, language string) error {
	lang := normalizeLanguage(language)
	strs := passwordResetStrings(lang, toName, resetURL, expiresIn)

	body, err := s.renderTemplate("password_reset.html", strs.Data)
	if err != nil {
		return fmt.Errorf("failed to render password reset template: %w", err)
	}

	return s.sendMail(toEmail, strs.Subject, body)
}

// SendEmailVerification sends an email verification email.
func (s *SMTPEmailService) SendEmailVerification(_ context.Context, toEmail, toName, verifyURL, language string) error {
	lang := normalizeLanguage(language)
	strs := emailVerificationStrings(lang, toName, verifyURL)

	body, err := s.renderTemplate("email_verification.html", strs.Data)
	if err != nil {
		return fmt.Errorf("failed to render email verification template: %w", err)
	}

	return s.sendMail(toEmail, strs.Subject, body)
}

// SendAccountDeletionConfirmation sends an account deletion confirmation email.
func (s *SMTPEmailService) SendAccountDeletionConfirmation(_ context.Context, toEmail, toName, language string) error {
	lang := normalizeLanguage(language)
	strs := accountDeletedStrings(lang, toName)

	body, err := s.renderTemplate("account_deleted.html", strs.Data)
	if err != nil {
		return fmt.Errorf("failed to render account deleted template: %w", err)
	}

	return s.sendMail(toEmail, strs.Subject, body)
}

// SendExpiryReminder sends an expiry reminder email.
func (s *SMTPEmailService) SendExpiryReminder(_ context.Context, toEmail, toName string, data ExpiryReminderData, unsubscribeURL, language string) error {
	lang := normalizeLanguage(language)
	strs := expiryReminderStrings(lang, toName, unsubscribeURL, data)

	body, err := s.renderTemplate("expiry_reminder.html", strs.Data)
	if err != nil {
		return fmt.Errorf("failed to render expiry reminder template: %w", err)
	}

	return s.sendMail(toEmail, strs.Subject, body)
}

// SendValidityStart sends a validity start notification email.
func (s *SMTPEmailService) SendValidityStart(_ context.Context, toEmail, toName string, data ValidityStartData, unsubscribeURL, language string) error {
	lang := normalizeLanguage(language)
	strs := validityStartStrings(lang, toName, unsubscribeURL, data)

	body, err := s.renderTemplate("validity_start.html", strs.Data)
	if err != nil {
		return fmt.Errorf("failed to render validity start template: %w", err)
	}

	return s.sendMail(toEmail, strs.Subject, body)
}

// SendShareNotification sends a share notification email.
func (s *SMTPEmailService) SendShareNotification(_ context.Context, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language string) error {
	lang := normalizeLanguage(language)
	strs := shareNotificationStrings(lang, toName, fromName, resourceType, resourceURL, unsubscribeURL)

	body, err := s.renderTemplate("share_notification.html", strs.Data)
	if err != nil {
		return fmt.Errorf("failed to render share notification template: %w", err)
	}

	return s.sendMail(toEmail, strs.Subject, body)
}

// SendTransferNotification sends a transfer notification email.
func (s *SMTPEmailService) SendTransferNotification(_ context.Context, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language string) error {
	lang := normalizeLanguage(language)
	strs := transferNotificationStrings(lang, toName, fromName, resourceType, resourceURL, unsubscribeURL)

	body, err := s.renderTemplate("share_notification.html", strs.Data)
	if err != nil {
		return fmt.Errorf("failed to render transfer notification template: %w", err)
	}

	return s.sendMail(toEmail, strs.Subject, body)
}

// CheckConnection verifies SMTP server connectivity without sending an email.
// Only tests TCP connectivity and SMTP banner/EHLO to avoid mail server logging
// empty transactions ("Client Quit Before Message") from health check probes.
func (s *SMTPEmailService) CheckConnection(ctx context.Context) error {
	addr := net.JoinHostPort(s.config.Host, fmt.Sprintf("%d", s.config.Port))

	// Use DialTimeout with context for timeout support
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Attempt to establish SMTP session
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer func() { _ = client.Close() }()

	// Send EHLO with our own domain (not the SMTP server's hostname)
	if err := client.Hello(s.ehloHost); err != nil {
		return fmt.Errorf("SMTP EHLO failed: %w", err)
	}

	// Gracefully close connection
	return client.Quit()
}

// SendTestEmail sends a test email to verify SMTP configuration.
func (s *SMTPEmailService) SendTestEmail(_ context.Context, toEmail, toName, language string) error {
	lang := normalizeLanguage(language)
	strs := testEmailStrings(lang, toName)

	body, err := s.renderTemplate("test_email.html", strs.Data)
	if err != nil {
		return fmt.Errorf("failed to render test email template: %w", err)
	}

	return s.sendMail(toEmail, strs.Subject, body)
}

func (s *SMTPEmailService) renderTemplate(name string, data any) (string, error) {
	// Inject LogoURL into template data if it's a map[string]string
	if dataMap, ok := data.(map[string]string); ok {
		dataMap["LogoURL"] = s.getLogoURL()
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// sanitizeHeader removes CR/LF characters to prevent email header injection.
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

func (s *SMTPEmailService) sendMail(to, subject, htmlBody string) error {
	// Sanitize header values to prevent email header injection
	to = sanitizeHeader(to)
	subject = sanitizeHeader(subject)

	from := s.config.FromEmail
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", sanitizeHeader(s.config.FromName), sanitizeHeader(s.config.FromEmail))
	}

	// Build email message with RFC 5322 compliant headers.
	// Message-ID and Date are required by RFC 5322 and their absence
	// triggers spam filters (e.g. MailChannels 550 5.7.1 [CS] blocks).
	randBytes := make([]byte, 8)
	_, _ = rand.Read(randBytes)
	msgID := fmt.Sprintf("<%d.%x@%s>", time.Now().UnixNano(), randBytes, s.ehloHost)

	var msg strings.Builder
	fmt.Fprintf(&msg, "Message-ID: %s\r\n", msgID)
	fmt.Fprintf(&msg, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	fmt.Fprintf(&msg, "From: %s\r\n", from)
	fmt.Fprintf(&msg, "To: %s\r\n", to)
	fmt.Fprintf(&msg, "Subject: %s\r\n", subject)
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	addr := net.JoinHostPort(s.config.Host, fmt.Sprintf("%d", s.config.Port))

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	if s.config.UseTLS {
		return s.sendMailTLS(addr, auth, s.config.FromEmail, to, []byte(msg.String()))
	}

	return smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(msg.String()))
}

func (s *SMTPEmailService) sendMailTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	// Connect via plain TCP first (required for STARTTLS on port 587)
	conn, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer func() { _ = client.Close() }()

	// Send EHLO with our own domain (not the SMTP server's hostname)
	if err := client.Hello(s.ehloHost); err != nil {
		return fmt.Errorf("SMTP EHLO failed: %w", err)
	}

	// Upgrade to TLS using STARTTLS (required for port 587)
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
		MinVersion: tls.VersionTLS12,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close email body: %w", err)
	}

	return client.Quit()
}

// LogEmailService logs emails instead of sending them (for development).
type LogEmailService struct{}

// NewLogEmailService creates a new log-based email service.
func NewLogEmailService() *LogEmailService {
	return &LogEmailService{}
}

// SendPasswordReset logs the password reset email.
func (s *LogEmailService) SendPasswordReset(_ context.Context, toEmail, toName, resetURL, expiresIn, language string) error {
	slog.Info("Email: Password Reset",
		"to", toEmail,
		"name", toName,
		"reset_url", resetURL,
		"expires_in", expiresIn,
		"language", language,
	)
	return nil
}

// SendEmailVerification logs the email verification email.
func (s *LogEmailService) SendEmailVerification(_ context.Context, toEmail, toName, verifyURL, language string) error {
	slog.Info("Email: Email Verification",
		"to", toEmail,
		"name", toName,
		"verify_url", verifyURL,
		"language", language,
	)
	return nil
}

// SendAccountDeletionConfirmation logs the account deletion confirmation email.
func (s *LogEmailService) SendAccountDeletionConfirmation(_ context.Context, toEmail, toName, language string) error {
	slog.Info("Email: Account Deletion Confirmation",
		"to", toEmail,
		"name", toName,
		"language", language,
	)
	return nil
}

// SendExpiryReminder logs the expiry reminder email.
func (s *LogEmailService) SendExpiryReminder(_ context.Context, toEmail, toName string, data ExpiryReminderData, unsubscribeURL, language string) error {
	slog.Info("Email: Expiry Reminder",
		"to", toEmail,
		"name", toName,
		"merchant", data.MerchantName,
		"resource_type", data.ResourceType,
		"days_left", data.DaysLeft,
		"expires_at", data.ExpiresAt,
		"code", data.Code,
		"value", data.Value,
		"unsubscribe_url", unsubscribeURL,
		"language", language,
	)
	return nil
}

// SendValidityStart logs the validity start notification email.
func (s *LogEmailService) SendValidityStart(_ context.Context, toEmail, toName string, data ValidityStartData, unsubscribeURL, language string) error {
	slog.Info("Email: Validity Start",
		"to", toEmail,
		"name", toName,
		"merchant", data.MerchantName,
		"valid_from", data.ValidFrom,
		"code", data.Code,
		"value", data.Value,
		"unsubscribe_url", unsubscribeURL,
		"language", language,
	)
	return nil
}

// SendShareNotification logs the share notification email.
func (s *LogEmailService) SendShareNotification(_ context.Context, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language string) error {
	slog.Info("Email: Share Notification",
		"to", toEmail,
		"name", toName,
		"from", fromName,
		"resource_type", resourceType,
		"resource_url", resourceURL,
		"unsubscribe_url", unsubscribeURL,
		"language", language,
	)
	return nil
}

// SendTransferNotification logs the transfer notification email.
func (s *LogEmailService) SendTransferNotification(_ context.Context, toEmail, toName, fromName, resourceType, resourceURL, unsubscribeURL, language string) error {
	slog.Info("Email: Transfer Notification",
		"to", toEmail,
		"name", toName,
		"from", fromName,
		"resource_type", resourceType,
		"resource_url", resourceURL,
		"unsubscribe_url", unsubscribeURL,
		"language", language,
	)
	return nil
}

// SendTestEmail logs the test email.
func (s *LogEmailService) SendTestEmail(_ context.Context, toEmail, toName, language string) error {
	slog.Info("Email: Test Email",
		"to", toEmail,
		"name", toName,
		"language", language,
	)
	return nil
}

// CheckConnection always returns nil for log-based email service (no real SMTP).
func (s *LogEmailService) CheckConnection(_ context.Context) error {
	slog.Debug("Log-based email service check (always healthy)")
	return nil
}
