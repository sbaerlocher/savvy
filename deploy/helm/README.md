# Savvy Helm Chart

Helm chart für das Savvy - Digitale Verwaltung von Kundenkarten, Gutscheinen und Geschenkkarten.

## Features

- ✅ Full-Stack Architecture (Go + Echo + SvelteKit + TypeScript)
- ✅ PostgreSQL Database (external oder internal)
- ✅ External Secrets Operator Integration (Bitwarden)
- ✅ Prometheus Metrics + ServiceMonitor
- ✅ Health Checks (Liveness + Readiness)
- ✅ Feature Toggles via Environment Variables
- ✅ OAuth/OIDC Support
- ✅ Horizontal Pod Autoscaling
- ✅ Security Best Practices (non-root, seccomp, drop ALL capabilities)

## Prerequisites

- Kubernetes 1.24+
- Helm 3.8+
- PostgreSQL 14+ (external oder via Sub-Chart)
- External Secrets Operator (optional, für Production empfohlen)
- Cert-Manager (für TLS)
- Ingress Controller (nginx empfohlen)

## Installation

### Testing & Validation

For local testing, linting, and CI/CD pipelines, use the test values file:

```bash
# Lint the Helm chart
helm lint deploy/helm --values deploy/helm/values-test.yaml

# Render templates
helm template savvy deploy/helm --values deploy/helm/values-test.yaml

# Or use just recipes
just helm-lint      # Lint with test values
just helm-template  # Render with test values
```

> **Note**: The test values file (`values-test.yaml`) provides placeholder values for required secrets like `database.existingSecret`, allowing chart validation without actual secrets.

### Quick Start (Development)

```bash
# Create a secret for the database URL
kubectl create secret generic savvy-db-secret \
  --from-literal=DATABASE_URL="postgres://user:pass@postgres:5432/savvy?sslmode=disable" \
  --namespace savvy-dev

# Install with custom values
helm install savvy deploy/helm \
  --namespace savvy-dev \
  --create-namespace \
  --set database.existingSecret=savvy-db-secret \
  --set config.enableLocalLogin=true \
  --set config.enableRegistration=true
```

### Production Deployment

```bash
# 1. Create secrets in Kubernetes (or use External Secrets Operator)
kubectl create secret generic savvy-db-secret \
  --from-literal=DATABASE_URL="postgres://user:pass@postgres.example.com:5432/savvy?sslmode=require" \
  --namespace savvy

# 2. Install with production values
helm install savvy deploy/helm \
  -f deploy/helm/values-production.yaml \
  --namespace savvy \
  --create-namespace \
  --set database.existingSecret=savvy-db-secret \
  --set image.tag=1.1.0 \
  --set ingress.hosts[0].host=savvy.example.com \
  --set oauth.issuer=https://auth.example.com/application/o/savvy/
```

## Configuration

### Required Values

The following values **must** be provided for deployment:

| Parameter                 | Description                                      | Example           |
| ------------------------- | ------------------------------------------------ | ----------------- |
| `database.existingSecret` | Kubernetes secret name containing `DATABASE_URL` | `savvy-db-secret` |

When OAuth is enabled, these are also required:

| Parameter                      | Description                         | Example                                         |
| ------------------------------ | ----------------------------------- | ----------------------------------------------- |
| `secrets.oauth.existingSecret` | Secret containing OAuth credentials | `savvy-oauth-secret`                            |
| `oauth.issuer`                 | OAuth/OIDC issuer URL               | `https://auth.example.com/...`                  |
| `oauth.redirectUrl`            | OAuth callback URL                  | `https://savvy.example.com/auth/oauth/callback` |

> **CI/CD Note**: For testing and validation without actual secrets, use `values-test.yaml` which provides placeholder values.

### Key Values

| Parameter                     | Description                   | Default                     |
| ----------------------------- | ----------------------------- | --------------------------- |
| `replicaCount`                | Number of replicas            | `1`                         |
| `image.repository`            | Image repository              | `ghcr.io/sbaerlocher/savvy` |
| `image.tag`                   | Image tag                     | `Chart.appVersion`          |
| `config.enableCards`          | Enable Cards feature          | `true`                      |
| `config.enableVouchers`       | Enable Vouchers feature       | `true`                      |
| `config.enableGiftCards`      | Enable Gift Cards feature     | `true`                      |
| `config.enableLocalLogin`     | Enable Email/Password login   | `true`                      |
| `config.enableRegistration`   | Enable user registration      | `false`                     |
| `config.timezone`             | IANA timezone for date calc.  | `UTC`                       |
| `config.smtpHost`             | SMTP server host              | `""`                        |
| `config.enable2FA`            | Enable 2FA/TOTP               | `false`                     |
| `secrets.smtp.existingSecret` | Secret with SMTP credentials  | `""`                        |
| `database.external.enabled`   | Use external PostgreSQL       | `true`                      |
| `externalSecrets.enabled`     | Use External Secrets Operator | `true`                      |
| `oauth.enabled`               | Enable OAuth/OIDC             | `false`                     |
| `ingress.enabled`             | Enable Ingress                | `false`                     |
| `monitoring.enabled`          | Enable Prometheus metrics     | `false`                     |
| `autoscaling.enabled`         | Enable HPA                    | `false`                     |

### Database Configuration

#### External PostgreSQL (Production)

```yaml
database:
  external:
    enabled: true
    host: "postgres.database.svc.cluster.local"
    port: 5432
    name: "savvy"
    user: "savvy"
    sslMode: "require"
    existingSecret: "savvy-db"
    existingSecretKey: "password"
```

#### Internal PostgreSQL (Development)

```yaml
database:
  external:
    enabled: false
  postgresql:
    enabled: true
    auth:
      username: savvy
      password: changeme
      database: savvy
    primary:
      persistence:
        enabled: true
        size: 8Gi
```

### External Secrets Configuration

```yaml
externalSecrets:
  enabled: true
  refreshInterval: 1h
  secretStoreRef:
    name: bitwarden-secret-store
    kind: ClusterSecretStore
  secrets:
    - name: savvy-session
      key: "savvy-session-secret"
      property: password
      target:
        name: savvy-session
        key: SESSION_SECRET
```

### OAuth/OIDC Configuration

```yaml
oauth:
  enabled: true
  clientId: "savvy-production"
  issuer: "https://auth.example.com/application/o/savvy/"
  redirectUrl: "https://savvy.example.com/callback"
  existingSecret: "savvy-oauth"
  existingSecretKey: "client-secret"
```

### SMTP Configuration (Email Notifications)

```yaml
config:
  smtpHost: "smtp.gmail.com"
  smtpPort: 587
  smtpFromEmail: "noreply@example.com"
  smtpFromName: "Savvy"
  smtpUseTLS: true

secrets:
  smtp:
    existingSecret: "savvy-smtp-secret"
    usernameKey: "SMTP_USERNAME"
    passwordKey: "SMTP_PASSWORD"
```

### Web Push Notifications (VAPID)

```yaml
config:
  vapidPublicKey: "BKxT..." # Generate with: npx web-push generate-vapid-keys
  vapidSubject: "mailto:admin@example.com"

secrets:
  vapid:
    existingSecret: "savvy-vapid-secret"
    privateKeyKey: "VAPID_PRIVATE_KEY"
```

### Expiry Reminders & 2FA

```yaml
config:
  enableExpiryReminders: true
  reminderDaysBefore: "7,3,1"
  reminderCheckTime: "08:00"
  enable2FA: false
  totpIssuer: "Savvy"

secrets:
  totp:
    existingSecret: "savvy-totp-secret" # Required if 2FA enabled
    encryptionKeyKey: "TOTP_ENCRYPTION_KEY"
```

### Feature Toggles

```yaml
config:
  enableCards: true # Cards feature
  enableVouchers: true # Vouchers feature
  enableGiftCards: true # Gift Cards feature
  enableLocalLogin: false # Email/Password (false = OAuth only)
  enableRegistration: false # User registration
  timezone: "Europe/Zurich" # IANA timezone (default: UTC)
```

## Monitoring

### Prometheus Metrics

```yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s
```

**Metrics Endpoint**: `/metrics`

**Key Metrics**:

- `http_request_duration_seconds` - HTTP Request Duration
- `http_requests_total` - Total HTTP Requests
- `cards_total`, `vouchers_total`, `gift_cards_total` - Resource Counts
- `db_connections_active`, `db_connections_idle` - Database Connections
- `login_attempts_total` - Login Attempts

### Health Checks

- **Liveness**: `/health` - Basic health check
- **Readiness**: `/ready` - Database + Dependencies check

## Upgrade

```bash
# Upgrade to new version
helm upgrade savvy deploy/helm \
  -f deploy/helm/values-production.yaml \
  --set image.tag=1.2.0
```

## Uninstall

```bash
helm uninstall savvy --namespace savvy
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n savvy -l app.kubernetes.io/name=savvy
```

### View Logs

```bash
kubectl logs -n savvy -l app.kubernetes.io/name=savvy -f
```

### Database Connection Issues

```bash
# Check External Secret
kubectl get externalsecret -n savvy
kubectl describe externalsecret savvy -n savvy

# Check Secret
kubectl get secret savvy-db -n savvy -o yaml

# Test Database Connection
kubectl run -it --rm debug --image=postgres:16 --restart=Never -n savvy -- \
  psql -h postgres.database.svc.cluster.local -U savvy -d savvy
```

### Ingress Issues

```bash
# Check Ingress
kubectl get ingress -n savvy
kubectl describe ingress savvy -n savvy

# Check Certificate
kubectl get certificate -n savvy
kubectl describe certificate savvy-tls -n savvy
```

## Security Considerations

- ✅ **Non-root User**: Runs as UID 65532 (nonroot)
- ✅ **seccomp Profile**: Runtime/Default
- ✅ **Capabilities**: ALL dropped
- ✅ **Read-only Root Filesystem**: Disabled (Session-Storage benötigt)
- ✅ **Network Policy**: Empfohlen in Production
- ✅ **Secrets Management**: External Secrets Operator (Bitwarden)
- ✅ **HTTPS Enforcement**: Ingress mit TLS + Force Redirect

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Kubernetes                            │
│                                                              │
│  ┌──────────────┐         ┌──────────────┐                  │
│  │   Ingress    │────────▶│   Service    │                  │
│  │  (nginx)     │         │  ClusterIP   │                  │
│  └──────────────┘         └──────────────┘                  │
│                                  │                           │
│                           ┌──────▼────────┐                  │
│                           │  Deployment   │                  │
│                           │  (HPA: 1-10)  │                  │
│                           └───────────────┘                  │
│                                  │                           │
│                 ┌────────────────┼────────────────┐          │
│                 │                │                │          │
│         ┌───────▼─────┐  ┌──────▼──────┐  ┌─────▼──────┐   │
│         │ PostgreSQL  │  │  External   │  │  Grafana   │   │
│         │  (external) │  │  Secrets    │  │   Alloy    │   │
│         └─────────────┘  │  Operator   │  │  (OTEL)    │   │
│                          └─────────────┘  └────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Links

- **GitHub**: https://github.com/sbaerlocher/loyalty-system
- **Documentation**: See [ARCHITECTURE.md](../../ARCHITECTURE.md)
- **Operations Guide**: See [OPERATIONS.md](../../OPERATIONS.md)

## License

MIT - See [LICENSE](../../LICENSE)
