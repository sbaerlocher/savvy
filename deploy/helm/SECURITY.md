# Helm Chart Security Guide

**Last Updated:** 2026-02-09
**Chart Version:** 1.1.0
**Security Posture:** Production-Ready ✅

---

## Overview

This document describes the security features and best practices implemented in the Savvy Helm chart. The chart follows industry standards including:

- **CIS Kubernetes Benchmark** (Center for Internet Security)
- **Pod Security Standards** (PSS) - Restricted Level
- **OWASP Kubernetes Top 10**
- **GitOps Security Best Practices**

---

## Security Features

### 1. Pod Security Standards (PSS) - Restricted ✅

The chart implements **PSS Restricted** level, the most restrictive security profile:

```yaml
# Pod Security Context
podSecurityContext:
  runAsNonRoot: true # ✅ PSS Restricted
  runAsUser: 65532 # ✅ Non-privileged UID
  runAsGroup: 65532 # ✅ Non-privileged GID
  fsGroup: 65532 # ✅ File system group
  seccompProfile:
    type: RuntimeDefault # ✅ PSS Restricted

# Container Security Context
securityContext:
  allowPrivilegeEscalation: false # ✅ PSS Restricted
  readOnlyRootFilesystem: true # ✅ PSS Restricted
  capabilities:
    drop:
      - ALL # ✅ PSS Restricted
```

**Benefits:**

- Prevents privilege escalation attacks
- Minimizes attack surface
- Enforces least privilege principle
- Complies with security frameworks

---

### 2. ServiceAccount Token Security (CIS 5.1.6) ✅

**Default:** `automountServiceAccountToken: false`

```yaml
serviceAccount:
  create: true
  automount: false # ✅ Don't mount token unless needed
```

**Why this matters:**

- Prevents container escape → Kubernetes API access
- Reduces privilege escalation risk
- Follows principle of least privilege

**When to enable:**
Only set `automount: true` if your application needs to call the Kubernetes API (rare). If enabled, create RBAC rules with minimal permissions.

---

### 3. Read-Only Root Filesystem (CIS 5.2.6) ✅

**Default:** `readOnlyRootFilesystem: true`

```yaml
securityContext:
  readOnlyRootFilesystem: true # ✅ Immutable infrastructure

# Writable volumes for runtime data
volumes:
  - name: tmp
    emptyDir: {}

volumeMounts:
  - name: tmp
    mountPath: /tmp
```

**Benefits:**

- Prevents malware/backdoor persistence
- Stops tampering with binaries
- Makes forensics easier
- Implements immutable infrastructure

---

### 4. Network Policies (Defense in Depth) 🔒

**Production Recommendation:** Always enable Network Policies

```yaml
# values-production.yaml
networkPolicy:
  enabled: true # ✅ Required for production
  ingress:
    # Allow only from Ingress Controller
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 3000
  egress:
    # Allow only DNS, Database, and HTTPS
    - to: [...]
```

**Benefits:**

- Network-level least privilege
- Prevents lateral movement
- Compliance requirement (PCI-DSS, SOC2)
- Default-deny + explicit-allow pattern

---

### 5. Image Digest Support (Supply Chain Security) 🔐

**Production Recommendation:** Use SHA256 digests instead of tags

```yaml
image:
  repository: ghcr.io/sbaerlocher/container/savvy
  tag: "1.1.0"
  digest: "sha256:abc123..." # ✅ Immutable, tamper-proof
```

**Why this matters:**

- Tags are mutable (can be re-pushed)
- Digests are immutable (cryptographically secure)
- Prevents supply chain attacks
- Ensures reproducibility

**CI/CD Integration:**

```bash
# Extract digest after build
DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' $IMAGE | cut -d'@' -f2)

# Deploy with digest
helm upgrade savvy . --set image.digest=$DIGEST
```

---

### 6. Secret Management (GitOps-Safe) 🔑

**Never commit secrets to Git!** Use one of these approaches:

#### Option 1: External Secrets Operator (Recommended)

```yaml
externalSecrets:
  enabled: true
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  sessionSecretId: "prod/savvy/session-secret"
  oauthClientIdSecret: "prod/savvy/oauth-client-id"
  # ... more secrets
```

#### Option 2: Kubernetes Secrets (Manual)

```bash
# Create secrets before deployment
kubectl create secret generic savvy-db-secret \
  --from-literal=DATABASE_URL="postgres://..."

kubectl create secret generic savvy-oauth-secret \
  --from-literal=OAUTH_CLIENT_ID="..." \
  --from-literal=OAUTH_CLIENT_SECRET="..."

# Deploy with references
helm upgrade savvy . \
  --set database.existingSecret=savvy-db-secret \
  --set secrets.oauth.existingSecret=savvy-oauth-secret
```

---

### 7. Required Field Validation (Fail-Fast) ⚡

**Feature:** Helm install fails early if required values are missing

```yaml
# Templates use 'required' function
name: { { required "database.existingSecret is required!" .Values.database.existingSecret } }
```

**Benefits:**

- Prevents runtime failures
- Clear error messages
- Better user experience

**Validation:**

```bash
# Test validation
helm template savvy . --values /dev/null
# ERROR: database.existingSecret is required!
```

---

### 8. JSON Schema Validation (Type Safety) 📋

**File:** `values.schema.json`

Helm 3.4+ automatically validates values against the schema:

```json
{
  "properties": {
    "database": {
      "required": ["existingSecret"],
      "properties": {
        "existingSecret": {
          "type": "string",
          "minLength": 1
        }
      }
    }
  }
}
```

**Benefits:**

- Type safety at install time
- IDE auto-complete
- Prevents configuration errors

---

## Production Deployment Checklist

Use this checklist before deploying to production:

### 🔒 Security

- [ ] Set `serviceAccount.automount: false` (unless K8s API access needed)
- [ ] Set `securityContext.readOnlyRootFilesystem: true`
- [ ] Enable Network Policies (`networkPolicy.enabled: true`)
- [ ] Use image digests instead of tags (`image.digest: sha256:...`)
- [ ] Store secrets in Kubernetes Secrets or External Secrets Operator
- [ ] Validate secrets are not in Git history
- [ ] Enable OAuth/OIDC (`oauth.enabled: true`, disable local login)

### 📊 Reliability

- [ ] Set `replicaCount: 2` or enable HPA
- [ ] Enable PodDisruptionBudget (`podDisruptionBudget.enabled: true`)
- [ ] Configure resource limits and requests
- [ ] Enable Pod Anti-Affinity for HA
- [ ] Configure health probes (startup, liveness, readiness)
- [ ] Set up monitoring (`monitoring.enabled: true`)

### 🌐 Networking

- [ ] Enable Ingress with TLS (`ingress.enabled: true`, configure TLS)
- [ ] Configure Network Policies (ingress + egress rules)
- [ ] Set up cert-manager for TLS certificates
- [ ] Configure proper CORS/CSP headers

### 📝 Compliance

- [ ] Document namespace PSS labels
- [ ] Configure audit logging
- [ ] Set up backup strategy
- [ ] Define retention policies

---

## Security Testing

### 1. Lint Check

```bash
helm lint deploy/helm/
```

### 2. Schema Validation

```bash
helm lint --strict deploy/helm/
```

### 3. Security Scanning

```bash
# Using kubesec
helm template savvy deploy/helm/ | kubesec scan -

# Using trivy
helm template savvy deploy/helm/ | trivy config --severity HIGH,CRITICAL -
```

### 4. Dry-Run Installation

```bash
helm install savvy deploy/helm/ \
  --dry-run --debug \
  --set database.existingSecret=test-secret \
  --set oauth.enabled=true \
  --set secrets.oauth.existingSecret=test-oauth-secret \
  --set oauth.issuer=https://auth.example.com \
  --set oauth.redirectUrl=https://savvy.example.com/callback
```

---

## CIS Kubernetes Benchmark Compliance

| CIS Control | Description                                                         | Status      | Implementation                     |
| ----------- | ------------------------------------------------------------------- | ----------- | ---------------------------------- |
| 5.1.6       | ServiceAccount tokens only where necessary                          | ✅ Pass     | `automount: false`                 |
| 5.2.6       | Read-only root filesystem                                           | ✅ Pass     | `readOnlyRootFilesystem: true`     |
| 5.2.1       | Minimize admission of privileged containers                         | ✅ Pass     | No privileged containers           |
| 5.2.2       | Minimize admission of containers with capabilities                  | ✅ Pass     | `drop: [ALL]`                      |
| 5.2.3       | Minimize admission of containers with added capabilities            | ✅ Pass     | No added capabilities              |
| 5.2.4       | Minimize admission of Windows HostProcess containers                | ✅ N/A      | Linux only                         |
| 5.2.5       | Minimize admission of containers with allowPrivilegeEscalation      | ✅ Pass     | `allowPrivilegeEscalation: false`  |
| 5.3.2       | Ensure that all Namespaces have NetworkPolicies defined             | ⚠️ Optional | Available, disabled by default     |
| 5.4.1       | Prefer using secrets as files over secrets as environment variables | ⚠️ Partial  | Secrets via envFrom (K8s standard) |
| 5.7.3       | Apply SecurityContext to Your Pods and Containers                   | ✅ Pass     | Full SecurityContext               |

---

## OWASP Kubernetes Top 10 Mitigation

| Risk | Description                                   | Mitigation                               |
| ---- | --------------------------------------------- | ---------------------------------------- |
| K01  | Insecure Workload Configurations              | ✅ PSS Restricted + SecurityContext      |
| K02  | Supply Chain Vulnerabilities                  | ✅ Image digests, container scanning     |
| K03  | Overly Permissive RBAC                        | ✅ No RBAC by default, automount=false   |
| K04  | Lack of Centralized Policy Enforcement        | ⚠️ User responsibility (OPA/Kyverno)     |
| K05  | Inadequate Logging and Monitoring             | ✅ Structured logs, Prometheus metrics   |
| K06  | Broken Authentication Mechanisms              | ✅ OAuth/OIDC enforced in production     |
| K07  | Missing Network Segmentation Controls         | ✅ NetworkPolicy support                 |
| K08  | Secrets Management Failures                   | ✅ ExternalSecrets, no plaintext secrets |
| K09  | Misconfigured Cluster Components              | ✅ Secure defaults, validation           |
| K10  | Outdated and Vulnerable Kubernetes Components | ⚠️ User responsibility (cluster updates) |

---

## Additional Security Recommendations

### 1. Pod Security Admission

Apply namespace labels for Pod Security Admission:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: savvy-prod
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: latest
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### 2. OPA Gatekeeper / Kyverno

Consider deploying policy enforcement:

```bash
# Example Kyverno policy: require read-only root filesystem
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-ro-rootfs
spec:
  validationFailureAction: enforce
  rules:
  - name: check-readOnlyRootFilesystem
    match:
      resources:
        kinds:
        - Pod
    validate:
      message: "Root filesystem must be read-only"
      pattern:
        spec:
          containers:
          - securityContext:
              readOnlyRootFilesystem: true
```

### 3. Image Scanning

Integrate image scanning in CI/CD:

```bash
# Scan before deployment
trivy image ghcr.io/sbaerlocher/container/savvy:1.1.0

# Or use admission controller (e.g., Falco, Trivy Operator)
```

### 4. Audit Logging

Enable Kubernetes audit logging to track all API calls.

---

## Support & Questions

- **Security Issues:** Report privately to security contact
- **General Questions:** Open GitHub issue
- **Documentation:** See [README.md](README.md), [DEPLOYMENT.md](DEPLOYMENT.md)

---

**Security is a journey, not a destination.** This chart provides strong defaults, but always adapt to your specific threat model and compliance requirements.
