# Savvy Deployment Guide

Vollständige Anleitung zum Deployment des Savvys auf Kubernetes mit Helm.

## Voraussetzungen

### 1. Kubernetes Cluster

- Kubernetes 1.24+
- Helm 3.8+
- kubectl konfiguriert

### 2. Erforderliche Operator/Controller

```bash
# Cert-Manager (für TLS Zertifikate)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml

# External Secrets Operator (für Secret Management)
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets \
  external-secrets/external-secrets \
  -n external-secrets-system \
  --create-namespace

# Ingress-Nginx (für Ingress)
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx \
  --create-namespace
```

## Production Deployment

### Schritt 1: Secrets in Bitwarden erstellen

Erstelle folgende Items in Bitwarden:

1. **savvy-session-secret**
   - Type: Login
   - Password: Generiere mit `openssl rand -base64 32`

2. **savvy-db-password**
   - Type: Login
   - Password: PostgreSQL User Password

3. **savvy-oauth-secret** (optional)
   - Type: Login
   - Password: OAuth Client Secret

### Schritt 2: External Secrets konfigurieren

Erstelle ClusterSecretStore für Bitwarden:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: bitwarden-secret-store
spec:
  provider:
    bitwarden:
      apiUrl: "https://vault.bitwarden.com"
      identityUrl: "https://identity.bitwarden.com"
      organizationId: "your-org-id"
      auth:
        secretRef:
          credentials:
            name: bitwarden-credentials
            namespace: external-secrets-system
            key: token
```

### Schritt 3: PostgreSQL Database erstellen

```bash
# Option A: Externe PostgreSQL (empfohlen)
# Erstelle Database und User manuell

# Option B: PostgreSQL im Cluster
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install postgres bitnami/postgresql \
  -n database \
  --create-namespace \
  --set auth.username=savvy \
  --set auth.password=changeme \
  --set auth.database=savvy \
  --set primary.persistence.size=20Gi
```

### Schritt 4: Helm Values anpassen

Erstelle `values-custom.yaml`:

```yaml
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
  enableLocalLogin: false  # OAuth only
  enableRegistration: false  # Invite only
```

### Schritt 5: Deployment

```bash
# Namespace erstellen
kubectl create namespace savvy

# Helm Chart deployen
helm install savvy ./helm/savvy \
  -f helm/savvy/values-production.yaml \
  -f values-custom.yaml \
  --namespace savvy

# Deployment Status prüfen
kubectl get pods -n savvy -w
```

### Schritt 6: DNS konfigurieren

```bash
# Ingress IP ermitteln
kubectl get ingress -n savvy savvy

# DNS A-Record erstellen:
# savvy.example.com -> <INGRESS-IP>
```

### Schritt 7: Verifizierung

```bash
# Health Check
curl https://savvy.example.com/health

# Readiness Check
curl https://savvy.example.com/ready

# Metrics
curl https://savvy.example.com/metrics

# Web UI
open https://savvy.example.com
```

## Development Deployment

Schnelles Setup für Development/Testing:

```bash
# Mit interner PostgreSQL
helm install savvy-dev ./helm/savvy \
  -f helm/savvy/values-development.yaml \
  --namespace savvy-dev \
  --create-namespace

# Port-Forward für lokalen Zugriff
kubectl port-forward -n savvy-dev svc/savvy-dev 8080:80

# Browser öffnen
open http://localhost:8080
```

## Upgrade

```bash
# Helm Chart upgraden
helm upgrade savvy ./helm/savvy \
  -f helm/savvy/values-production.yaml \
  -f values-custom.yaml \
  --namespace savvy

# Rollback bei Problemen
helm rollback savvy --namespace savvy
```

## Monitoring

### Prometheus Integration

```yaml
# values-custom.yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    labels:
      prometheus: kube-prometheus
```

### Grafana Dashboard

Importiere Dashboard aus `helm/grafana-dashboard.json` (TODO)

### Logs

```bash
# Alle Pods
kubectl logs -n savvy -l app.kubernetes.io/name=savvy -f

# Spezifischer Pod
kubectl logs -n savvy savvy-<pod-id> -f

# Letzte 100 Zeilen
kubectl logs -n savvy -l app.kubernetes.io/name=savvy --tail=100
```

## Backup & Recovery

### Database Backup

```bash
# PostgreSQL Backup
kubectl exec -n database postgres-postgresql-0 -- \
  pg_dump -U savvy savvy > backup-$(date +%Y%m%d).sql

# Restore
kubectl exec -i -n database postgres-postgresql-0 -- \
  psql -U savvy savvy < backup-20260127.sql
```

### Helm Release Backup

```bash
# Values exportieren
helm get values savvy -n savvy > values-backup.yaml

# Release-Metadaten
helm get all savvy -n savvy > release-backup.yaml
```

## Troubleshooting

### Pod startet nicht

```bash
# Events prüfen
kubectl describe pod -n savvy savvy-<pod-id>

# Init Container Logs
kubectl logs -n savvy savvy-<pod-id> -c init-container

# Main Container Logs
kubectl logs -n savvy savvy-<pod-id>
```

### Database Connection Fehler

```bash
# External Secret prüfen
kubectl get externalsecret -n savvy
kubectl describe externalsecret savvy -n savvy

# Secret Inhalt prüfen (vorsichtig!)
kubectl get secret savvy-db -n savvy -o yaml

# Database Connection testen
kubectl run -it --rm debug --image=postgres:16 --restart=Never -n savvy -- \
  psql -h postgres.database.svc.cluster.local -U savvy -d savvy
```

### Ingress/TLS Probleme

```bash
# Ingress prüfen
kubectl describe ingress -n savvy savvy

# Certificate prüfen
kubectl get certificate -n savvy
kubectl describe certificate savvy-tls -n savvy

# Cert-Manager Logs
kubectl logs -n cert-manager -l app=cert-manager -f
```

### Performance Issues

```bash
# Resource Usage
kubectl top pods -n savvy

# HPA Status
kubectl get hpa -n savvy
kubectl describe hpa savvy -n savvy

# Metrics
kubectl port-forward -n savvy svc/savvy 8080:80
curl http://localhost:8080/metrics
```

## Uninstall

```bash
# Helm Release löschen
helm uninstall savvy --namespace savvy

# Namespace löschen (vorsichtig!)
kubectl delete namespace savvy

# Secrets löschen
kubectl delete secret -n savvy savvy-db
kubectl delete secret -n savvy savvy-session
```

## Security Checklist

- [ ] Secrets via External Secrets Operator (nicht plain YAML)
- [ ] PostgreSQL mit SSL/TLS (`sslMode: require`)
- [ ] Ingress mit TLS (cert-manager + letsencrypt)
- [ ] Non-root User (UID 65532)
- [ ] Resource Limits gesetzt
- [ ] Network Policies definiert
- [ ] OAuth/OIDC aktiviert (`enableLocalLogin: false`)
- [ ] Registration deaktiviert (`enableRegistration: false`)
- [ ] Prometheus Metrics scraping (keine sensitiven Daten)
- [ ] Audit Logs aktiviert (PostgreSQL)

## Next Steps

Nach erfolgreichem Deployment:

1. **OAuth Provider konfigurieren**
   - Client erstellen
   - Redirect URLs setzen
   - Scopes definieren

2. **Monitoring einrichten**
   - Grafana Dashboards
   - Alerting Rules
   - Log Aggregation

3. **Backup Strategy**
   - PostgreSQL Backup Cron Job
   - Off-site Backup Storage
   - Recovery Testing

4. **Network Policies**
   - Ingress nur von Ingress Controller
   - Egress nur zu Database + OAuth Provider
   - Default Deny

## Support

- **Documentation**: [README.md](../README.md)
- **Architecture**: [ARCHITECTURE.md](../ARCHITECTURE.md)
- **Operations**: [OPERATIONS.md](../OPERATIONS.md)
- **GitHub Issues**: https://github.com/sbaerlocher/loyalty-system/issues
