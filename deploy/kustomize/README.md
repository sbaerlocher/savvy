# Savvy Kustomize

Kustomize-basierte Deployment-Konfiguration für das Savvy mit Base + Overlays für Development, Staging und Production.

## 📁 Struktur

```
kustomize/
├── base/                          # Base Konfiguration
│   ├── deployment.yaml           # Deployment Manifest
│   ├── service.yaml              # Service Manifest
│   ├── serviceaccount.yaml       # ServiceAccount
│   ├── configmap.yaml            # ConfigMap
│   ├── externalsecret.yaml       # External Secrets
│   └── kustomization.yaml        # Base Kustomization
│
├── overlays/
│   ├── development/              # Development Overlay
│   │   ├── kustomization.yaml   # Dev Kustomization
│   │   ├── deployment-patch.yaml # Dev Patches
│   │   ├── configmap-patch.yaml # Dev ConfigMap
│   │   ├── ingress.yaml         # Dev Ingress
│   │   └── postgres.yaml        # Internal PostgreSQL
│   │
│   ├── staging/                  # Staging Overlay
│   │   ├── kustomization.yaml   # Staging Kustomization
│   │   ├── deployment-patch.yaml # Staging Patches
│   │   ├── configmap-patch.yaml # Staging ConfigMap
│   │   ├── ingress.yaml         # Staging Ingress
│   │   └── hpa.yaml             # HPA (2-5 replicas)
│   │
│   └── production/               # Production Overlay
│       ├── kustomization.yaml   # Prod Kustomization
│       ├── deployment-patch.yaml # Prod Patches (3+ replicas, Anti-Affinity)
│       ├── configmap-patch.yaml # Prod ConfigMap + OAuth Config
│       ├── ingress.yaml         # Prod Ingress
│       ├── hpa.yaml             # HPA (3-10 replicas)
│       ├── servicemonitor.yaml  # Prometheus ServiceMonitor
│       └── oauth-externalsecret.yaml # OAuth Secret
│
└── README.md                     # This file
```

## 🚀 Quick Start

### Prerequisites

```bash
# Install Kustomize
brew install kustomize

# Or use kubectl (built-in)
kubectl version --client
```

### Development

```bash
# Build manifests
kustomize build kustomize/overlays/development

# Apply to cluster
kustomize build kustomize/overlays/development | kubectl apply -f -

# Or use kubectl directly
kubectl apply -k kustomize/overlays/development

# Port forward
kubectl port-forward -n savvy-dev svc/savvy 8080:80
```

### Staging

```bash
# Build and review
kustomize build kustomize/overlays/staging

# Apply
kubectl apply -k kustomize/overlays/staging

# Check status
kubectl get pods -n savvy-staging -w
```

### Production

```bash
# ALWAYS review before applying to production!
kustomize build kustomize/overlays/production > production-manifests.yaml
less production-manifests.yaml

# Apply
kubectl apply -k kustomize/overlays/production

# Watch rollout
kubectl rollout status deployment/savvy -n savvy
```

## 🔧 Configuration

### Base Configuration

**Defaults** (aus `base/configmap.yaml`):

- `GO_ENV=production`
- `ENABLE_CARDS=true`
- `ENABLE_VOUCHERS=true`
- `ENABLE_GIFT_CARDS=true`
- `ENABLE_LOCAL_LOGIN=true`
- `ENABLE_REGISTRATION=false`
- `TIMEZONE=UTC`
- `OTEL_ENABLED=false`

**Security**:

- Non-root User (UID 65532)
- seccomp Profile (RuntimeDefault)
- Drop ALL Capabilities
- External Secrets Operator

### Development Overlay

**Changes**:

- Namespace: `savvy-dev`
- Name Prefix: `dev-`
- Image Tag: `dev`
- Image Pull Policy: `Always`
- Replicas: `1`
- Resources: Minimal (50m CPU, 64Mi RAM)
- Local Login: Enabled
- Registration: Enabled
- Database: Internal PostgreSQL (no persistence)
- Ingress: `savvy-dev.local` (Staging Cert)

**Quick Deploy**:

```bash
kubectl apply -k kustomize/overlays/development
```

### Staging Overlay

**Changes**:

- Namespace: `savvy-staging`
- Name Prefix: `staging-`
- Image Tag: `staging`
- Replicas: `2` (HPA: 2-5)
- Resources: Medium (250m CPU, 256Mi RAM)
- Local Login: Disabled (OAuth only)
- Registration: Disabled
- OTEL: Enabled (Grafana Alloy)
- Database: External PostgreSQL (`savvy_staging`)
- Ingress: `savvy-staging.example.com` (Production Cert)

**Quick Deploy**:

```bash
kubectl apply -k kustomize/overlays/staging
```

### Production Overlay

**Changes**:

- Namespace: `savvy`
- Name Prefix: None
- Image Tag: `1.1.0`
- Replicas: `3` (HPA: 3-10)
- Resources: High (500m CPU, 512Mi RAM)
- Pod Anti-Affinity: Preferred (spread across nodes)
- Local Login: Disabled (OAuth only)
- Registration: Disabled
- OTEL: Enabled (Grafana Alloy)
- OAuth: Full Configuration
- Database: External PostgreSQL (`savvy`)
- Ingress: `savvy.example.com` (Production Cert, Rate Limit 100)
- ServiceMonitor: Prometheus Metrics

**Quick Deploy**:

```bash
# Review first!
kustomize build kustomize/overlays/production > production.yaml
less production.yaml

# Apply
kubectl apply -k kustomize/overlays/production
```

## 📊 Environment Comparison

| Feature            | Development | Staging         | Production    |
| ------------------ | ----------- | --------------- | ------------- |
| **Namespace**      | `savvy-dev` | `savvy-staging` | `savvy`       |
| **Replicas**       | 1           | 2 (HPA: 2-5)    | 3 (HPA: 3-10) |
| **Resources**      | 50m/64Mi    | 250m/256Mi      | 500m/512Mi    |
| **Database**       | Internal    | External        | External      |
| **Local Login**    | ✅ Enabled  | ❌ Disabled     | ❌ Disabled   |
| **Registration**   | ✅ Enabled  | ❌ Disabled     | ❌ Disabled   |
| **Timezone**       | EU/Zurich   | EU/Zurich       | EU/Zurich     |
| **OAuth**          | ❌ Disabled | ✅ Enabled      | ✅ Enabled    |
| **OTEL**           | ❌ Disabled | ✅ Enabled      | ✅ Enabled    |
| **ServiceMonitor** | ❌ Disabled | ❌ Disabled     | ✅ Enabled    |
| **Anti-Affinity**  | ❌ Disabled | ❌ Disabled     | ✅ Enabled    |
| **Rate Limit**     | ❌ Disabled | ❌ Disabled     | ✅ 100/min    |

## 🔐 Secret Management

### External Secrets Operator

Alle Secrets werden via External Secrets Operator aus Bitwarden synchronisiert:

**Required Secrets**:

1. **savvy-session-secret** - Session Secret
2. **savvy-db-password** - Database Password
3. **savvy-oauth-secret** - OAuth Client Secret (Production only)

**Setup**:

```bash
# 1. Create secrets in Bitwarden
# 2. Ensure ClusterSecretStore exists
kubectl get clustersecretstore bitwarden-secret-store

# 3. External Secrets werden automatisch erstellt
kubectl get externalsecret -n savvy
```

## 🔄 Customization

### Overlay-spezifische Änderungen

**Beispiel: Custom Domain in Production**:

```yaml
# kustomize/overlays/production/ingress.yaml
spec:
  rules:
    - host: my-domain.com # Ändern
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: savvy
                port:
                  number: 80
  tls:
    - secretName: savvy-tls
      hosts:
        - my-domain.com # Ändern
```

**Beispiel: Custom Resources**:

```yaml
# kustomize/overlays/production/deployment-patch.yaml
spec:
  template:
    spec:
      containers:
        - name: savvy
          resources:
            limits:
              cpu: "4" # Erhöhen
              memory: 2Gi # Erhöhen
            requests:
              cpu: "1" # Erhöhen
              memory: 1Gi # Erhöhen
```

### ConfigMap Generator

Kustomize unterstützt ConfigMap Generation:

```yaml
# kustomization.yaml
configMapGenerator:
  - name: savvy-version
    literals:
      - VERSION=1.2.0
      - BUILD_DATE=2026-01-28
```

## 🔍 Testing & Validation

### Dry-Run

```bash
# Development
kustomize build kustomize/overlays/development | kubectl apply --dry-run=client -f -

# Staging
kustomize build kustomize/overlays/staging | kubectl apply --dry-run=client -f -

# Production
kustomize build kustomize/overlays/production | kubectl apply --dry-run=server -f -
```

### Diff

```bash
# Vergleiche current state mit neuen manifests
kustomize build kustomize/overlays/production | kubectl diff -f -
```

### Validate

```bash
# Validiere YAML Syntax
kustomize build kustomize/overlays/production | kubectl apply --validate=true --dry-run=client -f -
```

## 🚀 Deployment Workflow

### CI/CD Integration

**GitHub Actions** (Beispiel):

```yaml
- name: Build Kustomize
  run: |
    kustomize build kustomize/overlays/${{ env.ENVIRONMENT }} > manifests.yaml

- name: Deploy
  run: |
    kubectl apply -f manifests.yaml
    kubectl rollout status deployment/savvy -n savvy
```

### GitOps (Rancher Fleet)

**fleet.yaml** (für `applications/` Repository):

```yaml
defaultNamespace: savvy
kustomize:
  dir: ./savvy/kustomize/overlays/production
```

### Manual Deployment

```bash
# 1. Review
kustomize build kustomize/overlays/production

# 2. Apply
kubectl apply -k kustomize/overlays/production

# 3. Watch
kubectl get pods -n savvy -w

# 4. Check rollout
kubectl rollout status deployment/savvy -n savvy

# 5. Verify
curl https://savvy.example.com/health
```

## 🐛 Troubleshooting

### Kustomize Build Fails

```bash
# Check syntax
kustomize build kustomize/overlays/production --enable-alpha-plugins

# Verbose output
kustomize build kustomize/overlays/production -v 10
```

### ConfigMap/Secret not found

```bash
# Check External Secrets
kubectl get externalsecret -n savvy
kubectl describe externalsecret savvy-session -n savvy

# Check generated secrets
kubectl get secret -n savvy
```

### Patches not applied

```bash
# Check patch paths in kustomization.yaml
# Ensure patch files exist
ls -la kustomize/overlays/production/*.yaml

# Test build
kustomize build kustomize/overlays/production | grep -A 10 "kind: Deployment"
```

### Image tag not updated

```bash
# Check images section in kustomization.yaml
kustomize build kustomize/overlays/production | grep "image:"
```

## 📚 Kustomize Resources

- **Official Docs**: https://kustomize.io
- **Kubectl Integration**: https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/
- **Best Practices**: https://github.com/kubernetes-sigs/kustomize/blob/master/docs/FIELDS.md

## 🔗 Related Documentation

- **Helm Charts**: [../helm/README.md](../helm/README.md)
- **Architecture**: [../ARCHITECTURE.md](../ARCHITECTURE.md)
- **Operations**: [../OPERATIONS.md](../OPERATIONS.md)
- **Main README**: [../README.md](../README.md)

## 📝 Changelog

### Version 1.0.0 (2026-01-27)

- ✅ Initial Kustomize Setup
- ✅ Base + 3 Overlays (Dev, Staging, Prod)
- ✅ External Secrets Integration
- ✅ HPA Configuration
- ✅ ServiceMonitor (Prometheus)
- ✅ Pod Anti-Affinity (Production)
- ✅ OAuth Configuration (Production)

## 📄 License

MIT - See [LICENSE](../LICENSE)
