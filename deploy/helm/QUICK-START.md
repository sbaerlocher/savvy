# Savvy - Quick Start Guide

Schnelleinstieg für lokales Testing und Production Deployment.

## 🚀 Development (5 Minuten)

Schnellstes Setup mit interner PostgreSQL:

```bash
# 1. Chart installieren
helm install savvy-dev ./helm/savvy \
  -f helm/savvy/values-development.yaml \
  --namespace savvy-dev \
  --create-namespace

# 2. Warten bis Pods ready
kubectl wait --for=condition=ready pod \
  -l app.kubernetes.io/name=savvy \
  -n savvy-dev \
  --timeout=120s

# 3. Port Forward
kubectl port-forward -n savvy-dev svc/savvy-dev-savvy 8080:80

# 4. Browser öffnen
open http://localhost:8080
```

**Default Login** (Development):
- Email: `admin@example.com`
- Password: `admin` (oder registriere neuen User)

## 🏭 Production (30 Minuten)

### Prerequisites

```bash
# 1. Cert-Manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml

# 2. External Secrets Operator
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets-system --create-namespace

# 3. Ingress-Nginx
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx --create-namespace
```

### Secrets Setup

```bash
# 1. Session Secret generieren
openssl rand -base64 32

# 2. In Bitwarden speichern als:
#    - savvy-session-secret
#    - savvy-db-password
#    - savvy-oauth-secret (optional)

# 3. ClusterSecretStore erstellen
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: bitwarden-secret-store
spec:
  provider:
    bitwarden:
      apiUrl: "https://vault.bitwarden.com"
      identityUrl: "https://identity.bitwarden.com"
      organizationId: "YOUR-ORG-ID"
      auth:
        secretRef:
          credentials:
            name: bitwarden-credentials
            namespace: external-secrets-system
            key: token
EOF
```

### Deployment

```bash
# 1. values-custom.yaml erstellen
cat > values-custom.yaml <<EOF
image:
  tag: "1.1.0"

ingress:
  enabled: true
  hosts:
    - host: savvy.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: savvy-tls
      hosts:
        - savvy.example.com

database:
  external:
    enabled: true
    host: "postgres.database.svc.cluster.local"
    name: "savvy"
    user: "savvy"

oauth:
  enabled: true
  clientId: "your-client-id"
  issuer: "https://auth.example.com/application/o/savvy/"
  redirectUrl: "https://savvy.example.com/callback"

config:
  enableLocalLogin: false
  enableRegistration: false
EOF

# 2. Install
helm install savvy ./helm/savvy \
  -f helm/savvy/values-production.yaml \
  -f values-custom.yaml \
  --namespace savvy \
  --create-namespace

# 3. Watch Pods
kubectl get pods -n savvy -w

# 4. Get Ingress IP
kubectl get ingress -n savvy
```

## ⚙️ Configuration Cheat Sheet

### Feature Toggles

```yaml
config:
  enableCards: true          # Kundenkarten
  enableVouchers: true       # Gutscheine
  enableGiftCards: true      # Geschenkkarten
  enableLocalLogin: true     # Email/Password Login
  enableRegistration: false  # Neue User Registration
```

### Resource Sizing

**Small (1-10 Users)**:
```yaml
replicaCount: 1
resources:
  limits:
    cpu: "500m"
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

**Medium (10-100 Users)**:
```yaml
replicaCount: 2
resources:
  limits:
    cpu: "1"
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
```

**Large (100+ Users)**:
```yaml
replicaCount: 3
resources:
  limits:
    cpu: "2"
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
```

## 🔍 Debugging Commands

```bash
# Pod Status
kubectl get pods -n savvy

# Logs (Follow)
kubectl logs -n savvy -l app.kubernetes.io/name=savvy -f

# Describe Pod
kubectl describe pod -n savvy <pod-name>

# Exec into Pod
kubectl exec -it -n savvy <pod-name> -- /bin/sh

# Port Forward
kubectl port-forward -n savvy svc/savvy 8080:80

# Events
kubectl get events -n savvy --sort-by='.lastTimestamp'

# Resource Usage
kubectl top pods -n savvy
```

## 🔧 Common Issues

### "ImagePullBackOff"

```bash
# Check image tag
helm get values savvy -n savvy | grep tag

# Check registry credentials
kubectl get secret -n savvy
```

### "CrashLoopBackOff"

```bash
# Check logs
kubectl logs -n savvy <pod-name> --previous

# Check database connection
kubectl run -it --rm debug --image=postgres:16 -n savvy -- \
  psql -h postgres.database.svc.cluster.local -U savvy
```

### "ExternalSecret not syncing"

```bash
# Check ExternalSecret status
kubectl get externalsecret -n savvy
kubectl describe externalsecret savvy -n savvy

# Check ClusterSecretStore
kubectl get clustersecretstore
kubectl describe clustersecretstore bitwarden-secret-store

# Check ESO logs
kubectl logs -n external-secrets-system -l app.kubernetes.io/name=external-secrets
```

## 📊 Health Checks

```bash
# Liveness (basic health)
curl http://localhost:8080/health

# Readiness (database + deps)
curl http://localhost:8080/ready

# Metrics
curl http://localhost:8080/metrics
```

## 🔄 Update

```bash
# Update image tag
helm upgrade savvy ./helm/savvy \
  -f values-production.yaml \
  -f values-custom.yaml \
  --set image.tag=1.2.0 \
  --namespace savvy

# Rollback
helm rollback savvy --namespace savvy

# History
helm history savvy --namespace savvy
```

## 🗑️ Cleanup

```bash
# Uninstall
helm uninstall savvy --namespace savvy

# Delete namespace
kubectl delete namespace savvy
```

## 📚 Next Steps

- **Full Documentation**: [helm/DEPLOYMENT.md](DEPLOYMENT.md)
- **Architecture**: [ARCHITECTURE.md](../../ARCHITECTURE.md)
- **Operations**: [OPERATIONS.md](../../OPERATIONS.md)
- **Chart README**: [helm/savvy/README.md](README.md)

## 🆘 Support

- **GitHub Issues**: https://github.com/sbaerlocher/loyalty-system/issues
- **Documentation**: [README.md](../../README.md)
