# Savvy System - Deployment

Deployment resources for the Savvy System - Digital Cards, Vouchers & Gift Cards Management.

## 📁 Structure

```
deploy/
├── grafana/                       # Grafana Dashboards
│   └── savvy-overview.json # Overview Dashboard
│
├── helm/                          # Helm Charts
│   ├── savvy/             # Chart Directory
│   │   ├── Chart.yaml            # Chart Metadata
│   │   ├── values.yaml           # Default Values
│   │   ├── values-*.yaml         # Environment Presets
│   │   ├── templates/            # Kubernetes Manifests
│   │   └── README.md             # Chart Documentation
│   ├── DEPLOYMENT.md             # Full Deployment Guide
│   ├── README.md                 # Helm Overview
│   └── savvy-1.1.0.tgz    # Packaged Chart
│
└── kustomize/                     # Kustomize Configurations
    ├── base/                      # Base Resources
    │   ├── namespace.yaml         # Namespace (with Pod Security Standards)
    │   ├── deployment.yaml        # Deployment
    │   ├── service.yaml           # Service
    │   ├── configmap.yaml         # ConfigMap
    │   ├── serviceaccount.yaml    # ServiceAccount
    │   ├── externalsecret.yaml    # External Secrets
    │   └── kustomization.yaml     # Base Kustomization
    │
    ├── overlays/                  # Environment Overlays
    │   ├── development/           # Dev Environment
    │   │   ├── kustomization.yaml
    │   │   ├── deployment-patch.yaml
    │   │   ├── configmap-patch.yaml
    │   │   ├── ingress.yaml
    │   │   └── postgres.yaml      # Internal PostgreSQL
    │   │
    │   ├── staging/               # Staging Environment
    │   │   ├── kustomization.yaml
    │   │   ├── deployment-patch.yaml
    │   │   ├── configmap-patch.yaml
    │   │   ├── ingress.yaml
    │   │   └── hpa.yaml           # HPA (2-5 replicas)
    │   │
    │   └── production/            # Production Environment
    │       ├── kustomization.yaml
    │       ├── deployment-patch.yaml
    │       ├── configmap-patch.yaml
    │       ├── ingress.yaml
    │       ├── hpa.yaml           # HPA (3-10 replicas)
    │       ├── servicemonitor.yaml # Prometheus
    │       ├── oauth-configmap.yaml # OAuth Config
    │       └── oauth-externalsecret.yaml
    │
    ├── README.md                  # Kustomize Documentation
    └── QUICK-START.md             # Quick Reference
```

## 🚀 Quick Start

### Helm (Recommended for Complex Deployments)

```bash
# Development
helm install savvy-dev ./deploy/helm/savvy \
  -f deploy/helm/savvy/values-development.yaml \
  --namespace savvy-dev \
  --create-namespace

# Production
helm install savvy ./deploy/helm/savvy \
  -f deploy/helm/savvy/values-production.yaml \
  --namespace savvy \
  --create-namespace
```

**See**: [helm/README.md](helm/README.md) for full documentation.

### Kustomize (Recommended for GitOps)

```bash
# Development
kubectl apply -k deploy/kustomize/overlays/development

# Production
kubectl apply -k deploy/kustomize/overlays/production
```

**See**: [kustomize/README.md](kustomize/README.md) for full documentation.

## 📊 Monitoring

### Grafana Dashboards

Import dashboards from `deploy/grafana/`:

1. **savvy-overview.json** - Main dashboard with:
   - HTTP Request Rate
   - Request Latency (p95, p99)
   - Total Users, Active Sessions
   - Resource Counts (Cards, Vouchers, Gift Cards)
   - Database Connections

### Prometheus Metrics

Metrics endpoint: `/metrics` (Port 8080)

**Key Metrics**:
- `http_request_duration_seconds` - Request latency histogram
- `http_requests_total` - Total HTTP requests counter
- `users_total` - Total users gauge
- `active_sessions` - Active user sessions gauge
- `cards_total` - Total cards gauge
- `vouchers_total` - Total vouchers gauge
- `gift_cards_total` - Total gift cards gauge
- `db_connections_active` - Active DB connections gauge
- `db_connections_idle` - Idle DB connections gauge

## 🔐 Security Features

### Pod Security Standards

Namespaces enforce **restricted** Pod Security Standards:
```yaml
pod-security.kubernetes.io/enforce: restricted
pod-security.kubernetes.io/audit: restricted
pod-security.kubernetes.io/warn: restricted
```

### Container Security

- ✅ Non-root User (UID 65532)
- ✅ seccomp Profile: RuntimeDefault
- ✅ Drop ALL Capabilities
- ✅ Read-only Root Filesystem: false (Session storage needed)
- ✅ No Privilege Escalation

### Secret Management

All secrets managed via **External Secrets Operator** (Bitwarden):
- `savvy-session` - Session secret
- `savvy-db` - Database password
- `savvy-oauth` - OAuth client secret (production)

## 🌍 Environment Comparison

| Feature | Development | Staging | Production |
|---------|------------|---------|------------|
| **Namespace** | savvy-dev | savvy-staging | savvy |
| **Replicas** | 1 | 2 (HPA: 2-5) | 3 (HPA: 3-10) |
| **Resources** | 50m/64Mi | 250m/256Mi | 500m/512Mi |
| **Database** | Internal | External | External |
| **Local Login** | ✅ Enabled | ❌ Disabled | ❌ Disabled |
| **Registration** | ✅ Enabled | ❌ Disabled | ❌ Disabled |
| **OAuth** | ❌ Disabled | ✅ Enabled | ✅ Enabled |
| **OTEL** | ❌ Disabled | ✅ Enabled | ✅ Enabled |
| **ServiceMonitor** | ❌ Disabled | ❌ Disabled | ✅ Enabled |
| **Anti-Affinity** | ❌ Disabled | ❌ Disabled | ✅ Enabled |
| **TLS** | Staging Cert | Prod Cert | Prod Cert |

## 🔄 Deployment Methods

### 1. Helm (Template-based)

**Pros**:
- Values-based configuration
- Package management (tgz)
- Helm hooks (post-install, pre-upgrade)
- Dependencies (sub-charts)
- Release management

**Cons**:
- Templating complexity
- State management (Helm secrets)

**Use Case**: Complex applications, package distribution

### 2. Kustomize (Overlay-based)

**Pros**:
- Pure YAML, no templating
- Native kubectl integration
- GitOps-friendly
- Overlay pattern (base + patches)
- No state management

**Cons**:
- Limited templating capabilities
- No dependency management

**Use Case**: GitOps workflows (Rancher Fleet, ArgoCD), simple overlays

## 🔧 Configuration

### Feature Toggles

```yaml
# All environments support these toggles:
ENABLE_CARDS: "true"          # Cards feature
ENABLE_VOUCHERS: "true"       # Vouchers feature
ENABLE_GIFT_CARDS: "true"     # Gift Cards feature
ENABLE_LOCAL_LOGIN: "false"   # Email/Password (false = OAuth only)
ENABLE_REGISTRATION: "false"  # User registration
```

### Observability

```yaml
# Staging + Production:
OTEL_ENABLED: "true"
OTEL_EXPORTER_OTLP_ENDPOINT: "grafana-alloy.observability.svc.cluster.local:4318"
```

### OAuth/OIDC

```yaml
# Staging + Production:
OAUTH_CLIENT_ID: "savvy-production"
OAUTH_ISSUER: "https://auth.example.com/application/o/savvy/"
OAUTH_REDIRECT_URL: "https://savvy.example.com/callback"
```

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [helm/README.md](helm/README.md) | Helm Charts Overview |
| [helm/DEPLOYMENT.md](helm/DEPLOYMENT.md) | Full Deployment Guide (30 min) |
| [helm/savvy/README.md](helm/savvy/README.md) | Chart Documentation |
| [kustomize/README.md](kustomize/README.md) | Kustomize Overview |
| [kustomize/QUICK-START.md](kustomize/QUICK-START.md) | Quick Reference Guide |

## 🎯 GitOps Integration

### Rancher Fleet

**fleet.yaml** (for `applications/` repository):

**Kustomize**:
```yaml
defaultNamespace: savvy
kustomize:
  dir: ./savvy/overlays/production
```

**Helm**:
```yaml
defaultNamespace: savvy
helm:
  chart: ./savvy/helm/savvy
  valuesFiles:
    - values-production.yaml
```

### ArgoCD

**Application**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: savvy
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/sbaerlocher/loyalty-system
    targetRevision: main
    path: deploy/kustomize/overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: savvy
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

## 🐛 Troubleshooting

### Common Issues

**Pods not starting**:
```bash
kubectl get pods -n savvy
kubectl describe pod <pod-name> -n savvy
kubectl logs <pod-name> -n savvy
```

**External Secrets not syncing**:
```bash
kubectl get externalsecret -n savvy
kubectl describe externalsecret savvy-session -n savvy
```

**Database connection failed**:
```bash
kubectl run -it --rm debug --image=postgres:16 -n savvy -- \
  psql -h postgres.database.svc.cluster.local -U savvy -d savvy
```

### Health Checks

```bash
# Liveness
kubectl exec -n savvy <pod-name> -- curl http://localhost:8080/health

# Readiness
kubectl exec -n savvy <pod-name> -- curl http://localhost:8080/ready

# Metrics
kubectl exec -n savvy <pod-name> -- curl http://localhost:8080/metrics
```

## 🔗 Links

- **GitHub**: https://github.com/sbaerlocher/loyalty-system
- **Main README**: [../README.md](../README.md)
- **Architecture**: [../ARCHITECTURE.md](../ARCHITECTURE.md)
- **Operations**: [../OPERATIONS.md](../OPERATIONS.md)

## 📝 Changelog

### Version 1.1.0 (2026-01-27)

- ✅ Helm Charts (Base + Dev/Prod Presets)
- ✅ Kustomize (Base + Dev/Staging/Prod Overlays)
- ✅ Grafana Dashboard (Overview)
- ✅ Pod Security Standards (restricted)
- ✅ External Secrets Operator Integration
- ✅ ServiceMonitor (Prometheus)
- ✅ HPA Configuration
- ✅ OAuth/OIDC Support

## 📄 License

MIT - See [LICENSE](../LICENSE)
