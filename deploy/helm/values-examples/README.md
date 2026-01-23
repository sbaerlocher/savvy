# External Secrets Values Examples

Provider-spezifische Konfigurationen für External Secrets Operator mit verschiedenen Secret Managern.

## 📁 Available Examples

| Provider | File | Notes |
|----------|------|-------|
| **Bitwarden** | [bitwarden.yaml](bitwarden.yaml) | Verwendet `key` (Item Name) + `property` (Field) |
| **HashiCorp Vault** | [vault.yaml](vault.yaml) | KV v2 path format: `secret/data/path` |
| **AWS Secrets Manager** | [aws-secrets-manager.yaml](aws-secrets-manager.yaml) | Verwendet ARN als `key` |
| **Google Secret Manager** | [gcp-secret-manager.yaml](gcp-secret-manager.yaml) | Full resource path mit `versions/latest` |
| **Azure Key Vault** | [azure-key-vault.yaml](azure-key-vault.yaml) | Secret Name ohne `property` |

## 🚀 Usage

### Install with Specific Provider

```bash
# Bitwarden
helm install savvy ./helm/savvy \
  -f helm/savvy/values-production.yaml \
  -f helm/savvy/values-examples/bitwarden.yaml

# Vault
helm install savvy ./helm/savvy \
  -f helm/savvy/values-production.yaml \
  -f helm/savvy/values-examples/vault.yaml

# AWS Secrets Manager
helm install savvy ./helm/savvy \
  -f helm/savvy/values-production.yaml \
  -f helm/savvy/values-examples/aws-secrets-manager.yaml
```

### Override Multiple Values

```bash
helm install savvy ./helm/savvy \
  -f helm/savvy/values-production.yaml \
  -f helm/savvy/values-examples/bitwarden.yaml \
  -f my-custom-values.yaml \
  --set image.tag=1.2.0
```

## 🔐 Provider-Specific Details

### Bitwarden

**remoteRef Structure**:
```yaml
remoteRef:
  key: "item-name"      # Bitwarden Item Name
  property: password    # Field: password, username, notes, etc.
```

**SecretStore Example**:
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

### HashiCorp Vault

**remoteRef Structure**:
```yaml
remoteRef:
  key: "secret/data/path/to/secret"  # KV v2 path
  property: "fieldname"                # JSON field
```

**SecretStore Example**:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-secret-store
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "savvy"
```

### AWS Secrets Manager

**remoteRef Structure**:
```yaml
remoteRef:
  key: "arn:aws:secretsmanager:region:account:secret:name"  # Full ARN
  property: "fieldname"  # Optional, for JSON secrets
```

**SecretStore Example**:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: savvy
```

### Google Secret Manager

**remoteRef Structure**:
```yaml
remoteRef:
  key: "projects/PROJECT_ID/secrets/SECRET_NAME/versions/VERSION"
```

**SecretStore Example**:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: gcp-secret-manager
spec:
  provider:
    gcpsm:
      projectID: "my-project-123"
      auth:
        workloadIdentity:
          clusterLocation: us-central1
          clusterName: my-cluster
          serviceAccountRef:
            name: savvy
```

### Azure Key Vault

**remoteRef Structure**:
```yaml
remoteRef:
  key: "secret-name"  # Key Vault Secret Name (no property needed)
```

**SecretStore Example**:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: azure-key-vault
spec:
  provider:
    azurekv:
      vaultUrl: "https://my-vault.vault.azure.net"
      authType: WorkloadIdentity
      serviceAccountRef:
        name: savvy
```

## 🔧 Creating Secrets in Each Provider

### Bitwarden

```bash
# Via Bitwarden CLI
bw login
bw unlock
bw create item --organizationid ORG_ID \
  --name "savvy-session-secret" \
  --username "session" \
  --password "$(openssl rand -base64 32)"
```

### Vault

```bash
# Via Vault CLI
vault login
vault kv put secret/savvy/session secret="$(openssl rand -base64 32)"
vault kv put secret/savvy/database password="db-password"
```

### AWS Secrets Manager

```bash
# Via AWS CLI
aws secretsmanager create-secret \
  --name savvy/session \
  --secret-string "$(openssl rand -base64 32)" \
  --region us-east-1

aws secretsmanager create-secret \
  --name savvy/database \
  --secret-string '{"password":"db-password"}' \
  --region us-east-1
```

### Google Secret Manager

```bash
# Via gcloud CLI
echo -n "$(openssl rand -base64 32)" | \
  gcloud secrets create savvy-session \
  --data-file=- \
  --project=my-project-123

echo -n "db-password" | \
  gcloud secrets create savvy-db-password \
  --data-file=- \
  --project=my-project-123
```

### Azure Key Vault

```bash
# Via Azure CLI
az keyvault secret set \
  --vault-name my-vault \
  --name savvy-session-secret \
  --value "$(openssl rand -base64 32)"

az keyvault secret set \
  --vault-name my-vault \
  --name savvy-db-password \
  --value "db-password"
```

## 🔍 Testing External Secrets

### Check ExternalSecret Status

```bash
# Get ExternalSecrets
kubectl get externalsecret -n savvy

# Describe to see sync status
kubectl describe externalsecret savvy -n savvy
```

### Check Created Secrets

```bash
# List secrets
kubectl get secret -n savvy

# View secret (base64 decoded)
kubectl get secret savvy-secrets -n savvy -o jsonpath='{.data.SESSION_SECRET}' | base64 -d
```

### Debug Sync Issues

```bash
# Check External Secrets Operator logs
kubectl logs -n external-secrets-system -l app.kubernetes.io/name=external-secrets

# Check SecretStore status
kubectl get secretstore -n savvy
kubectl describe secretstore aws-secrets-manager -n savvy
```

## 📚 Documentation

- **External Secrets Operator**: https://external-secrets.io
- **Bitwarden Provider**: https://external-secrets.io/latest/provider/bitwarden/
- **Vault Provider**: https://external-secrets.io/latest/provider/hashicorp-vault/
- **AWS Provider**: https://external-secrets.io/latest/provider/aws-secrets-manager/
- **GCP Provider**: https://external-secrets.io/latest/provider/google-secrets-manager/
- **Azure Provider**: https://external-secrets.io/latest/provider/azure-key-vault/

## 🔗 Related Files

- **Main values.yaml**: [../values.yaml](../values.yaml)
- **Production values**: [../values-production.yaml](../values-production.yaml)
- **External Secrets Template**: [../templates/externalsecret.yaml](../templates/externalsecret.yaml)
