# Kustomize Quick Start

Schnelleinstieg für Kustomize-basiertes Deployment des Savvy Systems.

## 🚀 5-Minuten Setup (Development)

```bash
# 1. Apply Development Overlay
kubectl apply -k kustomize/overlays/development

# 2. Wait for pods
kubectl wait --for=condition=ready pod \
  -l app.kubernetes.io/name=savvy \
  -n savvy-dev \
  --timeout=120s

# 3. Port Forward
kubectl port-forward -n savvy-dev svc/dev-savvy 8080:80

# 4. Open Browser
open http://localhost:8080
```

**Features**:
- ✅ Internal PostgreSQL (no persistence)
- ✅ Local Login enabled
- ✅ Registration enabled
- ✅ 1 Replica
- ✅ Minimal resources

## 🏭 Production Deployment

### Prerequisites

```bash
# 1. Secrets in Bitwarden
# - savvy-session-secret
# - savvy-db-password
# - savvy-oauth-secret

# 2. Verify ClusterSecretStore
kubectl get clustersecretstore bitwarden-secret-store

# 3. Verify External Database
kubectl run -it --rm test --image=postgres:16 -- \
  psql -h postgres.database.svc.cluster.local -U savvy -d savvy
```

### Customize

Edit production overlay files:

```bash
# Domain ändern
vim kustomize/overlays/production/ingress.yaml

# OAuth konfigurieren
vim kustomize/overlays/production/configmap-patch.yaml
```

### Deploy

```bash
# 1. Review manifests
kustomize build kustomize/overlays/production | less

# 2. Apply
kubectl apply -k kustomize/overlays/production

# 3. Watch rollout
kubectl rollout status deployment/savvy -n savvy

# 4. Check pods
kubectl get pods -n savvy

# 5. Test
curl https://savvy.example.com/health
```

## 📊 Environments

### Development

```bash
# Deploy
kubectl apply -k kustomize/overlays/development

# Logs
kubectl logs -n savvy-dev -l app.kubernetes.io/name=savvy -f

# Delete
kubectl delete -k kustomize/overlays/development
```

### Staging

```bash
# Deploy
kubectl apply -k kustomize/overlays/staging

# Status
kubectl get pods -n savvy-staging -w

# Delete
kubectl delete -k kustomize/overlays/staging
```

### Production

```bash
# Review FIRST!
kustomize build kustomize/overlays/production > prod.yaml
less prod.yaml

# Deploy
kubectl apply -k kustomize/overlays/production

# Rollout Status
kubectl rollout status deployment/savvy -n savvy

# Delete (careful!)
kubectl delete -k kustomize/overlays/production
```

## 🔧 Common Tasks

### Update Image Tag

**Development** (always latest):
```bash
# Already set to "dev" tag with imagePullPolicy: Always
kubectl rollout restart deployment/dev-savvy -n savvy-dev
```

**Production**:
```bash
# Edit kustomization.yaml
vim kustomize/overlays/production/kustomization.yaml

# Change:
images:
  - name: ghcr.io/sbaerlocher/savvy
    newTag: 1.2.0  # Update this

# Apply
kubectl apply -k kustomize/overlays/production

# Watch
kubectl rollout status deployment/savvy -n savvy
```

### Scale Manually

```bash
# Development
kubectl scale deployment/dev-savvy -n savvy-dev --replicas=2

# Production (temporary, HPA will override)
kubectl scale deployment/savvy -n savvy --replicas=5
```

### Feature Toggle

```bash
# Edit ConfigMap
kubectl edit configmap savvy -n savvy

# Or rebuild and apply
vim kustomize/overlays/production/configmap-patch.yaml
kubectl apply -k kustomize/overlays/production

# Restart pods to pick up changes
kubectl rollout restart deployment/savvy -n savvy
```

### View Diff

```bash
# Compare current vs new
kustomize build kustomize/overlays/production | kubectl diff -f -
```

### Dry Run

```bash
# Client-side validation
kustomize build kustomize/overlays/production | \
  kubectl apply --dry-run=client -f -

# Server-side validation (checks cluster state)
kustomize build kustomize/overlays/production | \
  kubectl apply --dry-run=server -f -
```

## 🐛 Debugging

### Build Failed

```bash
# Check syntax
kustomize build kustomize/overlays/production

# Verbose
kustomize build kustomize/overlays/production -v 10

# Validate each file
kubectl apply --dry-run=client -f kustomize/base/deployment.yaml
```

### Pods CrashLoopBackOff

```bash
# Check logs
kubectl logs -n savvy -l app.kubernetes.io/name=savvy --tail=50

# Describe pod
kubectl describe pod -n savvy <pod-name>

# Check secrets
kubectl get secret -n savvy
kubectl get externalsecret -n savvy
```

### ConfigMap not updated

```bash
# Check ConfigMap
kubectl get configmap savvy -n savvy -o yaml

# Recreate
kubectl delete configmap savvy -n savvy
kubectl apply -k kustomize/overlays/production

# Restart pods
kubectl rollout restart deployment/savvy -n savvy
```

### Database Connection Failed

```bash
# Check secret
kubectl get secret savvy-db -n savvy -o yaml

# Test connection
kubectl run -it --rm debug --image=postgres:16 -n savvy -- \
  psql -h postgres.database.svc.cluster.local -U savvy -d savvy

# Check External Secret
kubectl describe externalsecret savvy-db -n savvy
```

## 📋 Cheat Sheet

```bash
# Build
kustomize build kustomize/overlays/<env>

# Apply
kubectl apply -k kustomize/overlays/<env>

# Delete
kubectl delete -k kustomize/overlays/<env>

# Diff
kustomize build kustomize/overlays/<env> | kubectl diff -f -

# Dry Run
kubectl apply -k kustomize/overlays/<env> --dry-run=server

# Watch Pods
kubectl get pods -n <namespace> -w

# Logs
kubectl logs -n <namespace> -l app.kubernetes.io/name=savvy -f

# Rollout Status
kubectl rollout status deployment/<name> -n <namespace>

# Restart
kubectl rollout restart deployment/<name> -n <namespace>

# Port Forward
kubectl port-forward -n <namespace> svc/<service> 8080:80
```

## 🔗 Links

- **Full README**: [README.md](README.md)
- **Helm Charts**: [../helm/README.md](../helm/README.md)
- **Architecture**: [../ARCHITECTURE.md](../ARCHITECTURE.md)
- **Operations**: [../OPERATIONS.md](../OPERATIONS.md)
