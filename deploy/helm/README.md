# Savvy Helm Chart

Helm chart für das Savvy - Digitale Verwaltung von Kundenkarten, Gutscheinen und Geschenkkarten.

## Features

- ✅ Clean Architecture (Go + Echo + Templ + HTMX)
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

### Quick Start (Development)

```bash
# Mit interner PostgreSQL Datenbank
helm install savvy ./helm/savvy \
  -f helm/savvy/values-development.yaml \
  --namespace savvy-dev \
  --create-namespace
```

### Production Deployment

```bash
# 1. Secrets in Bitwarden erstellen:
#    - savvy-session-secret
#    - savvy-db-password
#    - savvy-oauth-secret (optional)

# 2. External Secrets konfigurieren
helm install savvy ./helm/savvy \
  -f helm/savvy/values-production.yaml \
  --namespace savvy \
  --create-namespace \
  --set image.tag=1.1.0 \
  --set ingress.hosts[0].host=savvy.example.com \
  --set oauth.issuer=https://auth.example.com/application/o/savvy/
```

## Configuration

### Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `ghcr.io/sbaerlocher/savvy` |
| `image.tag` | Image tag | `Chart.appVersion` |
| `config.enableCards` | Enable Cards feature | `true` |
| `config.enableVouchers` | Enable Vouchers feature | `true` |
| `config.enableGiftCards` | Enable Gift Cards feature | `true` |
| `config.enableLocalLogin` | Enable Email/Password login | `true` |
| `config.enableRegistration` | Enable user registration | `false` |
| `database.external.enabled` | Use external PostgreSQL | `true` |
| `externalSecrets.enabled` | Use External Secrets Operator | `true` |
| `oauth.enabled` | Enable OAuth/OIDC | `false` |
| `ingress.enabled` | Enable Ingress | `false` |
| `monitoring.enabled` | Enable Prometheus metrics | `false` |
| `autoscaling.enabled` | Enable HPA | `false` |

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

### Feature Toggles

```yaml
config:
  enableCards: true          # Cards feature
  enableVouchers: true       # Vouchers feature
  enableGiftCards: true      # Gift Cards feature
  enableLocalLogin: false    # Email/Password (false = OAuth only)
  enableRegistration: false  # User registration
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
- `active_sessions` - Active User Sessions
- `db_connections_active`, `db_connections_idle` - Database Connections

### Health Checks

- **Liveness**: `/health` - Basic health check
- **Readiness**: `/ready` - Database + Dependencies check

## Upgrade

```bash
# Upgrade to new version
helm upgrade savvy ./helm/savvy \
  -f helm/savvy/values-production.yaml \
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
