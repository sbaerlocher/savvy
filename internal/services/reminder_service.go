// Package services contains business logic.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"savvy/internal/email"
	"savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ReminderServiceInterface defines expiry reminder operations.
type ReminderServiceInterface interface {
	CheckAndSendReminders(ctx context.Context) error
}

// ReminderService implements ReminderServiceInterface.
type ReminderService struct {
	reminderRepo      repository.ReminderRepository
	voucherRepo       repository.VoucherRepository
	giftCardRepo      repository.GiftCardRepository
	voucherShareRepo  repository.VoucherShareRepository
	giftCardShareRepo repository.GiftCardShareRepository
	notifRepo         repository.NotificationRepository
	pushService       PushServiceInterface
	emailService      email.ServiceInterface
	emailTokenService EmailTokenServiceInterface
	daysBefore        []int
	location          *time.Location // Timezone for date calculations
	frontendURL       string         // Base URL for resource links in emails
}

// NewReminderService creates a new reminder service.
func NewReminderService(
	reminderRepo repository.ReminderRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
	voucherShareRepo repository.VoucherShareRepository,
	giftCardShareRepo repository.GiftCardShareRepository,
	notifRepo repository.NotificationRepository,
	pushService PushServiceInterface,
	emailService email.ServiceInterface,
	emailTokenService EmailTokenServiceInterface,
	daysBefore []int,
	location *time.Location,
	frontendURL string,
) ReminderServiceInterface {
	if location == nil {
		location = time.UTC
	}
	return &ReminderService{
		reminderRepo:      reminderRepo,
		voucherRepo:       voucherRepo,
		giftCardRepo:      giftCardRepo,
		voucherShareRepo:  voucherShareRepo,
		giftCardShareRepo: giftCardShareRepo,
		notifRepo:         notifRepo,
		pushService:       pushService,
		emailService:      emailService,
		emailTokenService: emailTokenService,
		daysBefore:        daysBefore,
		location:          location,
		frontendURL:       strings.TrimSuffix(frontendURL, "/"),
	}
}

// CheckAndSendReminders checks for expiring vouchers and gift cards and sends reminders.
func (s *ReminderService) CheckAndSendReminders(ctx context.Context) error {
	slog.InfoContext(ctx, "Checking for expiring items", "days_before", s.daysBefore)

	var totalSent int
	for _, days := range s.daysBefore {
		sent, err := s.checkVouchers(ctx, days)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to check expiring vouchers", "days_before", days, "error", err)
		}
		totalSent += sent

		sent, err = s.checkGiftCards(ctx, days)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to check expiring gift cards", "days_before", days, "error", err)
		}
		totalSent += sent
	}

	// Check vouchers becoming valid tomorrow
	sent, err := s.checkVoucherValidityStart(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to check voucher validity starts", "error", err)
	}
	totalSent += sent

	slog.InfoContext(ctx, "Expiry reminder check complete", "reminders_sent", totalSent)
	return nil
}

// calculateDaysLeft calculates the number of days until expiry in the configured timezone.
// It uses calendar days (start of day) instead of exact hours to avoid timezone issues.
func (s *ReminderService) calculateDaysLeft(expiryTime time.Time) int {
	// Get today's calendar date in configured timezone
	now := time.Now().In(s.location)
	ty, tm, td := now.Date()

	// Extract date components from UTC since dates are stored as end-of-day UTC
	// (T23:59:59Z). Converting to local timezone first can shift the date forward
	// by one day (e.g., 23:59:59 UTC → 00:59:59+01:00 = next day in Europe/Zurich).
	utc := expiryTime.UTC()
	ey, em, ed := utc.Date()

	// Compare as pure UTC dates to avoid DST hour-shift issues.
	// Constructing both dates in UTC with the same time component ensures
	// the Sub() result is always an exact multiple of 24 hours.
	todayUTC := time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
	expiryUTC := time.Date(ey, em, ed, 0, 0, 0, 0, time.UTC)

	daysLeft := int(expiryUTC.Sub(todayUTC).Hours() / 24)

	// Ensure we return at least 0 (for items expiring today)
	if daysLeft < 0 {
		return 0
	}
	return daysLeft
}

func (s *ReminderService) checkVouchers(ctx context.Context, daysBefore int) (int, error) {
	vouchers, err := s.voucherRepo.GetExpiringVouchers(ctx, daysBefore)
	if err != nil {
		return 0, fmt.Errorf("get expiring vouchers: %w", err)
	}

	sent := 0
	for i := range vouchers {
		v := &vouchers[i]
		if v.UserID == nil {
			continue
		}

		daysLeft := s.calculateDaysLeft(v.ValidUntil)

		// Only send reminder if daysLeft matches the current daysBefore window
		if daysLeft != daysBefore {
			continue
		}

		merchantName := "Unknown"
		if v.Merchant != nil {
			merchantName = v.Merchant.Name
		}

		emailData := email.ExpiryReminderData{
			MerchantName: merchantName,
			ResourceType: "voucher",
			DaysLeft:     daysLeft,
			ExpiresAt:    s.formatDate(v.ValidUntil, v.User),
			Code:         v.Code,
			Value:        s.formatVoucherValue(v),
			ResourceURL:  s.buildResourceURL("vouchers", v.ID),
		}

		// Send reminder to owner
		if s.sendReminderToUser(ctx, *v.UserID, v.User, "voucher", v.ID, daysBefore, daysLeft, merchantName, emailData) {
			sent++
		}

		// Send reminders to share recipients
		if s.voucherShareRepo != nil {
			shares, err := s.voucherShareRepo.GetByVoucherID(ctx, v.ID)
			if err != nil {
				slog.WarnContext(ctx, "Failed to get voucher shares for reminder", "voucher_id", v.ID, "error", err)
			} else {
				for j := range shares {
					share := &shares[j]
					if share.SharedWithUser == nil {
						continue
					}
					shareEmailData := emailData
					shareEmailData.ExpiresAt = s.formatDate(v.ValidUntil, share.SharedWithUser)
					if s.sendReminderToUser(ctx, share.SharedWithID, share.SharedWithUser, "voucher", v.ID, daysBefore, daysLeft, merchantName, shareEmailData) {
						sent++
					}
				}
			}
		}
	}

	return sent, nil
}

func (s *ReminderService) checkGiftCards(ctx context.Context, daysBefore int) (int, error) {
	giftCards, err := s.giftCardRepo.GetExpiringGiftCards(ctx, daysBefore)
	if err != nil {
		return 0, fmt.Errorf("get expiring gift cards: %w", err)
	}

	sent := 0
	for i := range giftCards {
		gc := &giftCards[i]
		if gc.UserID == nil || gc.ExpiresAt == nil {
			continue
		}

		daysLeft := s.calculateDaysLeft(*gc.ExpiresAt)

		// Only send reminder if daysLeft matches the current daysBefore window
		if daysLeft != daysBefore {
			continue
		}

		merchantName := "Unknown"
		if gc.Merchant != nil {
			merchantName = gc.Merchant.Name
		}

		emailData := email.ExpiryReminderData{
			MerchantName: merchantName,
			ResourceType: "gift_card",
			DaysLeft:     daysLeft,
			ExpiresAt:    s.formatDate(*gc.ExpiresAt, gc.User),
			Code:         gc.CardNumber,
			Value:        s.formatCurrency(gc.CurrentBalance, gc.Currency),
			ResourceURL:  s.buildResourceURL("gift-cards", gc.ID),
		}

		// Send reminder to owner
		if s.sendReminderToUser(ctx, *gc.UserID, gc.User, "gift_card", gc.ID, daysBefore, daysLeft, merchantName, emailData) {
			sent++
		}

		// Send reminders to share recipients
		if s.giftCardShareRepo != nil {
			shares, err := s.giftCardShareRepo.GetByGiftCardID(ctx, gc.ID)
			if err != nil {
				slog.WarnContext(ctx, "Failed to get gift card shares for reminder", "gift_card_id", gc.ID, "error", err)
			} else {
				for j := range shares {
					share := &shares[j]
					if share.SharedWithUser == nil {
						continue
					}
					shareEmailData := emailData
					shareEmailData.ExpiresAt = s.formatDate(*gc.ExpiresAt, share.SharedWithUser)
					if s.sendReminderToUser(ctx, share.SharedWithID, share.SharedWithUser, "gift_card", gc.ID, daysBefore, daysLeft, merchantName, shareEmailData) {
						sent++
					}
				}
			}
		}
	}

	return sent, nil
}

// sendReminderToUser sends a reminder (notification + push + email) to a single user.
// Returns true if the reminder was sent, false if it was already sent or failed.
// In-app notifications are always created. Push and email are gated by user preferences.
func (s *ReminderService) sendReminderToUser(
	ctx context.Context,
	userID uuid.UUID,
	user *models.User,
	resourceType string,
	resourceID uuid.UUID,
	daysBefore int,
	daysLeft int,
	merchantName string,
	emailData email.ExpiryReminderData,
) bool {
	already, err := s.reminderRepo.HasBeenSent(ctx, userID, resourceType, resourceID, daysBefore)
	if err != nil {
		slog.WarnContext(ctx, "Failed to check reminder status", "user_id", userID, "resource_type", resourceType, "resource_id", resourceID, "error", err)
		return false
	}
	if already {
		return false
	}

	// Create in-app notification (always)
	notification := &models.Notification{
		UserID:       userID,
		Type:         models.NotificationTypeExpiryReminder,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata: models.NotificationMetadata{
			"merchant_name": merchantName,
			"days_left":     daysLeft,
			"expires_at":    emailData.ExpiresAt,
		},
		IsRead: false,
	}
	if err := s.notifRepo.Create(ctx, notification); err != nil {
		slog.WarnContext(ctx, "Failed to create expiry notification", "user_id", userID, "resource_type", resourceType, "resource_id", resourceID, "error", err)
		return false
	}

	lang := ""
	if user != nil {
		lang = user.Language
	}

	// Send push notification (gated by category + channel preferences)
	if user == nil || (user.PushRemindersEnabled && user.PushNotificationsEnabled) {
		s.sendPush(ctx, userID, merchantName, resourceType, daysLeft, lang)
	}

	// Send email notification (gated by category + channel preferences)
	if user == nil || (user.EmailRemindersEnabled && user.EmailNotificationsEnabled) {
		s.sendEmail(ctx, user, emailData)
	}

	// Mark reminder as sent
	reminder := &models.ExpiryReminderSent{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		DaysBefore:   daysBefore,
	}
	if err := s.reminderRepo.MarkSent(ctx, reminder); err != nil {
		slog.WarnContext(ctx, "Failed to mark reminder as sent", "user_id", userID, "resource_type", resourceType, "resource_id", resourceID, "error", err)
	}

	return true
}

func (s *ReminderService) sendPush(ctx context.Context, userID uuid.UUID, merchantName, resourceType string, daysLeft int, lang string) {
	if s.pushService == nil {
		return
	}

	lctx := i18nCtx(lang)
	resource := i18n.T(lctx, "push.resource."+resourceType)
	title := i18n.T(lctx, "push.expiry.title")

	var body string
	if daysLeft == 1 {
		body = i18n.T(lctx, "push.expiry.body_tomorrow", map[string]any{"Resource": resource, "Merchant": merchantName})
	} else {
		body = i18n.T(lctx, "push.expiry.body", map[string]any{"Resource": resource, "Merchant": merchantName, "Days": daysLeft})
	}

	url := "/" + resourceType + "s"
	if err := s.pushService.SendPushToUser(ctx, userID, title, body, url); err != nil {
		slog.WarnContext(ctx, "Failed to send expiry push notification", "user_id", userID, "error", err)
	}
}

func (s *ReminderService) sendEmail(ctx context.Context, user *models.User, data email.ExpiryReminderData) {
	if s.emailService == nil || user == nil {
		return
	}

	userName := user.FirstName
	if userName == "" {
		userName = user.Email
	}

	lang := user.Language
	if lang == "" {
		lang = "en"
	}

	unsubscribeURL := s.generateUnsubscribeURL(ctx, user.ID)

	if err := s.emailService.SendExpiryReminder(ctx, user.Email, userName, data, unsubscribeURL, lang); err != nil {
		slog.WarnContext(ctx, "Failed to send expiry reminder email", "email", user.Email, "error", err)
	}
}

// generateUnsubscribeURL creates a one-click unsubscribe URL for expiry reminders.
func (s *ReminderService) generateUnsubscribeURL(ctx context.Context, userID uuid.UUID) string {
	if s.emailTokenService == nil {
		return ""
	}

	token, err := s.emailTokenService.CreateUnsubscribeReminderToken(ctx, userID)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create unsubscribe reminder token", "user_id", userID, "error", err)
		return ""
	}

	return s.frontendURL + "/unsubscribe?token=" + token + "&type=reminders"
}

// formatDate formats a date according to the user's language preference.
// Uses UTC date components since dates are stored as end-of-day UTC (T23:59:59Z).
// Converting to local timezone can shift the date forward by one day.
func (s *ReminderService) formatDate(t time.Time, user *models.User) string {
	utc := t.UTC()
	lang := ""
	if user != nil {
		lang = user.Language
	}
	switch lang {
	case "de":
		return fmt.Sprintf("%d. %s %d", utc.Day(), germanMonth(utc.Month()), utc.Year())
	case "fr":
		return fmt.Sprintf("%d %s %d", utc.Day(), frenchMonth(utc.Month()), utc.Year())
	default:
		return utc.Format("January 2, 2006")
	}
}

func germanMonth(m time.Month) string {
	months := [...]string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"}
	return months[m-1]
}

func frenchMonth(m time.Month) string {
	months := [...]string{"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"}
	return months[m-1]
}

// formatVoucherValue formats the voucher value based on its type.
func (s *ReminderService) formatVoucherValue(v *models.Voucher) string {
	switch v.Type {
	case "percentage":
		return fmt.Sprintf("%.0f%%", v.Value)
	case "fixed_amount":
		return s.formatCurrency(v.Value, v.Currency)
	case "bonus_points":
		return fmt.Sprintf("+%.0f Punkte", v.Value)
	default:
		return ""
	}
}

// formatCurrency formats a monetary value with currency.
func (s *ReminderService) formatCurrency(amount float64, currency string) string {
	if currency == "" {
		currency = "CHF"
	}
	return fmt.Sprintf("%s %.2f", currency, amount)
}

// buildResourceURL builds a full URL to view the resource in Savvy.
func (s *ReminderService) buildResourceURL(resourcePath string, resourceID uuid.UUID) string {
	if s.frontendURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", s.frontendURL, resourcePath, resourceID.String())
}

// checkVoucherValidityStart checks for vouchers becoming valid tomorrow and sends notifications.
func (s *ReminderService) checkVoucherValidityStart(ctx context.Context) (int, error) {
	vouchers, err := s.voucherRepo.GetVouchersStartingTomorrow(ctx)
	if err != nil {
		return 0, fmt.Errorf("get vouchers starting tomorrow: %w", err)
	}

	sent := 0
	for i := range vouchers {
		v := &vouchers[i]
		if v.UserID == nil {
			continue
		}

		merchantName := "Unknown"
		if v.Merchant != nil {
			merchantName = v.Merchant.Name
		}

		emailData := email.ValidityStartData{
			MerchantName: merchantName,
			ValidFrom:    s.formatDate(v.ValidFrom, v.User),
			Code:         v.Code,
			Value:        s.formatVoucherValue(v),
			ResourceURL:  s.buildResourceURL("vouchers", v.ID),
		}

		// Send to owner
		if s.sendValidityStartToUser(ctx, *v.UserID, v.User, v.ID, merchantName, emailData) {
			sent++
		}

		// Send to share recipients
		if s.voucherShareRepo != nil {
			shares, err := s.voucherShareRepo.GetByVoucherID(ctx, v.ID)
			if err != nil {
				slog.WarnContext(ctx, "Failed to get voucher shares for validity start", "voucher_id", v.ID, "error", err)
			} else {
				for j := range shares {
					share := &shares[j]
					if share.SharedWithUser == nil {
						continue
					}
					shareEmailData := emailData
					shareEmailData.ValidFrom = s.formatDate(v.ValidFrom, share.SharedWithUser)
					if s.sendValidityStartToUser(ctx, share.SharedWithID, share.SharedWithUser, v.ID, merchantName, shareEmailData) {
						sent++
					}
				}
			}
		}
	}

	return sent, nil
}

// sendValidityStartToUser sends a validity start notification to a single user.
// In-app notifications are always created. Push and email are gated by user preferences.
func (s *ReminderService) sendValidityStartToUser(
	ctx context.Context,
	userID uuid.UUID,
	user *models.User,
	voucherID uuid.UUID,
	merchantName string,
	emailData email.ValidityStartData,
) bool {
	// Dedup via expiry_reminder_sents with resource_type="voucher_start"
	already, err := s.reminderRepo.HasBeenSent(ctx, userID, "voucher_start", voucherID, 1)
	if err != nil {
		slog.WarnContext(ctx, "Failed to check validity start reminder status", "user_id", userID, "voucher_id", voucherID, "error", err)
		return false
	}
	if already {
		return false
	}

	// Create in-app notification (always)
	notification := &models.Notification{
		UserID:       userID,
		Type:         models.NotificationTypeValidityStart,
		ResourceType: "voucher",
		ResourceID:   voucherID,
		Metadata: models.NotificationMetadata{
			"merchant_name": merchantName,
			"valid_from":    emailData.ValidFrom,
		},
		IsRead: false,
	}
	if err := s.notifRepo.Create(ctx, notification); err != nil {
		slog.WarnContext(ctx, "Failed to create validity start notification", "user_id", userID, "voucher_id", voucherID, "error", err)
		return false
	}

	lang := ""
	if user != nil {
		lang = user.Language
	}

	// Send push notification (gated by category + channel preferences)
	if user == nil || (user.PushRemindersEnabled && user.PushNotificationsEnabled) {
		s.sendValidityStartPush(ctx, userID, merchantName, lang)
	}

	// Send email notification (gated by category + channel preferences)
	if user == nil || (user.EmailRemindersEnabled && user.EmailNotificationsEnabled) {
		s.sendValidityStartEmail(ctx, user, emailData)
	}

	// Mark as sent
	reminder := &models.ExpiryReminderSent{
		UserID:       userID,
		ResourceType: "voucher_start",
		ResourceID:   voucherID,
		DaysBefore:   1,
	}
	if err := s.reminderRepo.MarkSent(ctx, reminder); err != nil {
		slog.WarnContext(ctx, "Failed to mark validity start reminder as sent", "user_id", userID, "voucher_id", voucherID, "error", err)
	}

	return true
}

func (s *ReminderService) sendValidityStartPush(ctx context.Context, userID uuid.UUID, merchantName, lang string) {
	if s.pushService == nil {
		return
	}

	lctx := i18nCtx(lang)
	title := i18n.T(lctx, "push.validity_start.title")
	body := i18n.T(lctx, "push.validity_start.body", map[string]any{"Merchant": merchantName})

	if err := s.pushService.SendPushToUser(ctx, userID, title, body, "/vouchers"); err != nil {
		slog.WarnContext(ctx, "Failed to send validity start push notification", "user_id", userID, "error", err)
	}
}

func (s *ReminderService) sendValidityStartEmail(ctx context.Context, user *models.User, data email.ValidityStartData) {
	if s.emailService == nil || user == nil {
		return
	}

	userName := user.FirstName
	if userName == "" {
		userName = user.Email
	}

	lang := user.Language
	if lang == "" {
		lang = "en"
	}

	unsubscribeURL := s.generateUnsubscribeURL(ctx, user.ID)

	if err := s.emailService.SendValidityStart(ctx, user.Email, userName, data, unsubscribeURL, lang); err != nil {
		slog.WarnContext(ctx, "Failed to send validity start email", "email", user.Email, "error", err)
	}
}
