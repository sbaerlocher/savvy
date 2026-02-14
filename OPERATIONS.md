# Operations Guide

**Last Updated**: 2026-03-02
**Project**: Savvy
**Version**: 1.2.0

---

## Overview

This document covers operational aspects for running Savvy in production:

- Production deployment
- Monitoring and observability
- Incident response
- Performance tuning
- Security operations
- Admin operations
- Scaling

For development setup, see [DEVELOPMENT.md](DEVELOPMENT.md).

---

## 🚀 Production Deployment

### Architecture

```
Client (HTTPS:443) → Traefik (TLS) → Savvy App (HTTP:3000) → PostgreSQL
```

### Environment Variables

```bash
# Application
GO_ENV=production
SERVER_PORT=3000

# Database
DATABASE_URL=postgres://user:pass@host:5432/savvy?sslmode=require

# Session (CRITICAL: Change in production!)
SESSION_SECRET=<generate-with-openssl-rand-base64-32>
SESSION_MAX_AGE=604800                # Session duration in seconds (default: 7 days)
# Note: Sessions are stored in PostgreSQL (server-side) with SHA-256 hashed tokens
# Expired sessions are automatically cleaned up every hour

# OAuth (Optional)
OAUTH_CLIENT_ID=your-client-id
OAUTH_CLIENT_SECRET=your-client-secret
OAUTH_ISSUER=https://auth.example.com

# Feature Toggles
ENABLE_CARDS=true                    # Enable Cards feature
ENABLE_VOUCHERS=true                 # Enable Vouchers feature
ENABLE_GIFT_CARDS=true               # Enable Gift Cards feature
ENABLE_LOCAL_LOGIN=true              # Enable email/password authentication
ENABLE_REGISTRATION=true             # Enable user self-registration

# Timezone
TIMEZONE=Europe/Zurich               # IANA timezone for date calculations (default: UTC)

# SMTP Configuration (Email Notifications)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=smtp-user
SMTP_PASSWORD=smtp-password
SMTP_FROM_EMAIL=noreply@example.com
SMTP_FROM_NAME=Savvy
SMTP_USE_TLS=true

# Web Push Notifications (VAPID)
VAPID_PUBLIC_KEY=<public-key>        # Generate with: npx web-push generate-vapid-keys
VAPID_PRIVATE_KEY=<private-key>
VAPID_SUBJECT=mailto:admin@example.com

# Expiry Reminders
ENABLE_EXPIRY_REMINDERS=true
REMINDER_DAYS_BEFORE=7,3,1           # Days before expiry
REMINDER_CHECK_TIME=08:00

# Two-Factor Authentication (2FA/TOTP)
ENABLE_2FA=false
TOTP_ISSUER=Savvy
TOTP_ENCRYPTION_KEY=<32-byte-base64> # AES-256 key

# Rate Limiting
RATE_LIMIT_GLOBAL_RATE=50               # Global rate limit (requests/second)
RATE_LIMIT_GLOBAL_BURST=100             # Global burst size
RATE_LIMIT_AUTH_RATE=5                   # Auth endpoint rate (requests/second)
RATE_LIMIT_AUTH_BURST=10                 # Auth endpoint burst
RATE_LIMIT_PASSWORD_RESET_RATE=2         # Password reset rate
RATE_LIMIT_PASSWORD_RESET_BURST=5        # Password reset burst
RATE_LIMIT_USER_RATE=10                  # Per-user rate
RATE_LIMIT_USER_BURST=20                 # Per-user burst

# Application
LOG_LEVEL=INFO                           # DEBUG, INFO, WARN, ERROR
METRICS_PORT=9090                        # Separate Prometheus metrics port
SHUTDOWN_TIMEOUT_SECONDS=10              # Graceful shutdown timeout
CSP_REPORT_URI=                          # Optional CSP report-uri endpoint

# Observability
OTEL_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
```

### Feature Toggles

The application supports runtime feature toggles to enable/disable functionality without code changes.

**Resource Features**:

- `ENABLE_CARDS` - Controls Cards feature (default: `true`)
  - When disabled: Cards menu hidden, API endpoints return 404
  - Use case: Gradual rollout or feature deprecation

- `ENABLE_VOUCHERS` - Controls Vouchers feature (default: `true`)
  - When disabled: Vouchers menu hidden, API endpoints return 404
  - Use case: Seasonal feature (e.g., only during promotional periods)

- `ENABLE_GIFT_CARDS` - Controls Gift Cards feature (default: `true`)
  - When disabled: Gift Cards menu hidden, API endpoints return 404
  - Use case: Compliance requirements or regional restrictions

**Authentication Features**:

- `ENABLE_LOCAL_LOGIN` - Controls email/password authentication (default: `true`)
  - When disabled: Only OAuth/OIDC login available
  - Use case: Enterprise SSO-only deployments

- `ENABLE_REGISTRATION` - Controls user self-registration (default: `true`)
  - When disabled: Only admins can create users via admin panel
  - Use case: Invite-only or closed beta deployments

**Configuration**:

Feature toggles are read at application startup. Changes require application restart. The client fetches feature
configuration from `/api/v1/config` on page load.

### Pre-Deployment Checklist

- [ ] Generate strong SESSION_SECRET (min 32 chars)
- [ ] Configure DATABASE_URL with SSL enabled
- [ ] Set up TLS certificates (Let's Encrypt via Traefik)
- [ ] Configure backup strategy
- [ ] Set up monitoring (Prometheus + Grafana)
- [ ] Configure log aggregation
- [ ] Test rollback procedure

---

## 📊 Monitoring & Observability

### Health Checks

- `/health` - Liveness probe
- `/ready` - Readiness probe
- `/metrics` - Prometheus metrics

### Key Metrics

**HTTP**:

- `http_request_duration_seconds` - Request latency histogram
- `http_requests_total` - Total HTTP requests counter

**Database**:

- `db_connections_active` - Active connections
- `db_connections_idle` - Idle connections

**Resources**:

- `cards_total`, `vouchers_total`, `gift_cards_total` - Resource counts
- `vouchers_by_status` - Vouchers by status (active, expired)
- `gift_cards_by_status` - Gift cards by status (active, expired, redeemed)
- `shares_total` - Active shares by resource type

**Notifications**:

- `push_subscriptions_total`, `push_subscribed_users_total`, `email_verified_users_total`
- `push_notifications_enabled_total`, `email_notifications_enabled_total`
- `push_reminders_enabled_total`, `push_sharing_enabled_total`
- `email_reminders_enabled_total`, `email_sharing_enabled_total`

**Authentication**:

- `login_attempts_total` - Login attempts by result

### Logging

Structured JSON logs with trace IDs:

```json
{
  "time": "2026-02-09T10:00:00Z",
  "level": "info",
  "msg": "HTTP request",
  "method": "GET",
  "path": "/api/v1/cards",
  "status": 200,
  "duration_ms": 45,
  "trace_id": "abc123"
}
```

---

## 🚨 Incident Response

### Common Issues

#### High Error Rate

```bash
# Check logs
kubectl logs -l app=savvy --tail=100 | grep ERROR

# Check database connectivity
kubectl exec -it deployment/savvy -- nc -zv postgres 5432

# Check pod status
kubectl get pods -l app=savvy

# Check pod events
kubectl describe pod -l app=savvy
```

#### Database Connection Pool Exhausted

```sql
-- Check active connections
SELECT count(*) FROM pg_stat_activity WHERE state = 'active';

-- Kill long-running query
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE pid = <pid>;
```

#### Unauthorized Access

```sql
-- Check failed logins
SELECT ip_address, COUNT(*)
FROM audit_logs
WHERE action = 'login_failed'
AND created_at >= NOW() - INTERVAL '1 hour'
GROUP BY ip_address HAVING COUNT(*) > 5;
```

---

## ⚡ Performance Tuning

### Database Optimization

```sql
-- Check slow queries
SELECT calls, mean_exec_time, query
FROM pg_stat_statements
WHERE mean_exec_time > 100
ORDER BY mean_exec_time DESC LIMIT 20;

-- Check unused indexes
SELECT indexname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0;
```

### Connection Pool

```bash
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

---

## 🔐 Security Operations

### Security Monitoring

```sql
-- Suspicious deletion activity
SELECT user_id, COUNT(*) AS deletions
FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '1 hour'
AND action IN ('delete', 'hard_delete')
GROUP BY user_id HAVING COUNT(*) > 10;
```

### Security Checklist

- [ ] Strong SESSION_SECRET (32+ chars)
- [ ] HTTPS enabled (TLS 1.2+)
- [ ] Database SSL (sslmode=require)
- [ ] Regular security updates
- [ ] Rate limiting enabled
- [ ] CSRF protection active
- [ ] Audit logging enabled
- [ ] Server-side sessions enabled (PostgreSQL-backed)
- [ ] Stale session invalidation active (password change invalidates old sessions)

---

## 👤 Admin Operations

### Accessing the Admin Panel

Admin users can access the admin interface at `/admin` after logging in with an admin account.

**Creating the first admin user**:

```sql
-- Set existing user as admin (one-time setup)
UPDATE users SET is_admin = true WHERE email = 'admin@example.com';
```

### Admin Panel Features

The admin panel provides five sections accessible via dropdown menu in navigation:

**1. Users Management** (`/admin/users`)

- View all users with search and filtering (by role, auth provider, email/name)
- Create new users (if local login is enabled)
- Edit user details (name, email, role)
- Promote/demote admin status
- Expandable user detail rows
- User impersonation for support troubleshooting

**2. Merchants Management** (`/admin/merchants`)

- View all merchants/brands
- Create new merchants with name, color, and logo
- Edit existing merchants
- Manage merchant branding across cards, vouchers, and gift cards

**3. Audit Log Viewer** (`/admin/audit-log`)

- View all system actions (deletions, ownership transfers)
- Filter by user, action type, resource type, date range
- Search by resource ID or user email
- Export logs for compliance reporting

**4. Email Templates** (`/admin/email-templates`)

- Preview all email templates

**5. System Health** (`/admin/system-health`)

- Real-time monitoring of all services (database, SMTP, OAuth, VAPID, 2FA)
- Auto-refresh every 30 seconds (toggleable)
- Test email functionality for SMTP validation
- SVG status indicators with color coding

### User Management

**Creating a new user**:

1. Navigate to `/admin/users`
2. Click "Create User" button
3. Fill in email, name, password, and role (user/admin)
4. Submit to create account

**Editing a user**:

1. Find user in admin panel
2. Click "Edit" on user row
3. Update name, email, or role
4. Save changes

**Promoting to admin**:

1. Edit the user
2. Change role from "user" to "admin"
3. User gains admin privileges immediately

### User Impersonation

Admins can impersonate users to troubleshoot issues while maintaining audit trails.

**How to impersonate**:

1. Navigate to `/admin/users`
2. Find the target user in the list
3. Click "Impersonate" button
4. Redirected to dashboard as that user
5. Banner shows "Viewing as [User Name]"
6. Click "Stop Impersonation" to return to admin account

**Security notes**:

- Only admins can impersonate
- Cannot impersonate other admins
- All actions are logged with admin's ID
- Session tracks impersonation state

### Resource Restoration

Soft-deleted resources can be restored through the admin API (UI coming soon):

**API Endpoints**:

- `POST /api/v1/admin/resources/cards/{id}/restore`
- `POST /api/v1/admin/resources/vouchers/{id}/restore`
- `POST /api/v1/admin/resources/gift-cards/{id}/restore`

**Manual restoration** (emergency fallback):

```sql
-- Restore soft-deleted card
UPDATE cards SET deleted_at = NULL WHERE id = 'uuid-here';
UPDATE card_shares SET deleted_at = NULL WHERE card_id = 'uuid-here';
```

---

## 📈 Scaling

### Horizontal Scaling

Scaling configuration is managed through deployment manifests:

- **Helm Charts**: See `deploy/helm/templates/hpa.yaml` for HorizontalPodAutoscaler configuration
- **Kustomize**: See `deploy/kustomize/overlays/*/kustomization.yaml` for environment-specific replica counts

Key scaling parameters:

- Minimum replicas: 2 (production)
- Maximum replicas: 10
- Target CPU utilization: 70%

### Database Scaling

For database scaling considerations:

- Use read replicas for read-heavy workloads
- Configure PgBouncer for connection pooling
- Adjust connection pool settings in environment variables

---

## 📞 Support

- **Security**: [security@sbaerlo.ch](mailto:security@sbaerlo.ch)
- **Documentation**: [SUPPORT.md](SUPPORT.md)

---

For development setup, see [DEVELOPMENT.md](DEVELOPMENT.md).

For system architecture, see [ARCHITECTURE.md](ARCHITECTURE.md).
