# Operations Guide

**Letzte Aktualisierung**: 2026-01-25
**Projekt**: Savvy (Savvy System)

---

## 📋 Übersicht

Dieses Dokument enthält alle operationalen Aspekte des Systems:

- Audit Logging & Compliance
- Observability & Monitoring (OpenTelemetry)
- Security Best Practices

---

## 🔍 Audit Logging

### Übersicht

Das System implementiert umfassendes Audit Logging für:

- **Transaction Creator Tracking** - Wer hat welche Gift Card Transaction erstellt
- **Deletion Auditing** - Alle Löschoperationen (soft-delete & hard-delete)
- **User Context** - IP-Adresse und User-Agent bei jeder Löschung
- **Data Snapshots** - JSON-Snapshot der gelöschten Ressourcen

### Architektur

```
User Action (Delete)
     ↓
Handler → audit.LogDeletion()
     ↓
GORM Delete Operation
     ↓
AfterDelete Callback
     ↓
AuditLog Entry Created
```

### Database Schema

**audit_logs Table**:

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR NOT NULL,              -- "delete", "hard_delete", "restore"
    resource_type VARCHAR NOT NULL,       -- "cards", "vouchers", etc.
    resource_id UUID NOT NULL,
    resource_data JSONB,                  -- Full snapshot
    ip_address VARCHAR(45),               -- IPv4/IPv6
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes für Performance
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

**gift_card_transactions.created_by_user_id**:

```sql
ALTER TABLE gift_card_transactions
ADD COLUMN created_by_user_id UUID REFERENCES users(id);
```

### Verwendung

#### Automatisches Logging

GORM Hooks erstellen automatisch Audit Logs bei Deletions:

```go
// Einfach löschen - Audit Log wird automatisch erstellt
database.DB.Delete(&card)
```

#### Manuelles Logging

```go
import "savvy/internal/audit"

// Explizites Audit Logging
err := audit.LogDeletionFromContext(
    c,              // Echo context
    database.DB,
    "custom_resource",
    resourceID,
    resourceData,
)
```

### Queries

```sql
-- Alle Deletions eines Users
SELECT * FROM audit_logs
WHERE user_id = '<user_uuid>'
ORDER BY created_at DESC;

-- Alle Deletions eines Resource-Typs
SELECT * FROM audit_logs
WHERE resource_type = 'gift_card_transactions'
ORDER BY created_at DESC;

-- Deletions der letzten 7 Tage
SELECT * FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '7 days';

-- Gelöschte Ressource anzeigen
SELECT
    resource_type,
    resource_id,
    resource_data::jsonb->>'merchant_name' AS merchant,
    created_at
FROM audit_logs
WHERE resource_type = 'cards';
```

### GDPR Compliance

**Right to be Forgotten**:

```sql
DELETE FROM audit_logs WHERE user_id = '<user_uuid>';
DELETE FROM users WHERE id = '<user_uuid>';
```

**Data Retention** (Beispiel: 7 Jahre):

```sql
DELETE FROM audit_logs
WHERE created_at < NOW() - INTERVAL '7 years';
```

---

## 📊 Observability & OpenTelemetry

### Übersicht

Das System nutzt **OpenTelemetry (OTel)** für:

- **Distributed Tracing** - Request-Tracking durch den gesamten Stack
- **Database Query Tracing** - GORM-Operationen monitoren
- **Error Tracking** - Automatische Fehler-Erfassung in Spans
- **Request Context** - Trace IDs in allen Logs

### Konfiguration

Environment Variables:

```bash
# OpenTelemetry Exporter
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=savvy
OTEL_SERVICE_VERSION=1.1.0

# Optional: Sampling
OTEL_TRACES_SAMPLER=always_on
```

### Integration

**Automatisches Tracing**:

- Alle HTTP-Requests via Echo Middleware
- Alle GORM Queries via otelgorm Plugin
- Automatic Context Propagation

**Custom Spans**:

```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("savvy")
ctx, span := tracer.Start(ctx, "custom-operation")
defer span.End()

// Your code here
```

### Monitoring Queries

**Trace ID in Logs**:

```bash
# Logs mit Trace ID filtern
docker compose logs app | grep "trace_id=abc123"
```

**Performance Monitoring**:

- Jaeger UI: http://localhost:16686 (if running)
- Grafana Cloud: https://sbaerlo.grafana.net

---

## 🔐 Security

### Implementierte Security Features

✅ **Authentication & Authorization**:

- Session-based Auth (Gorilla Sessions)
- Bcrypt Password Hashing
- Admin Role Check Middleware
- OAuth/OIDC Support

✅ **Input Validation**:

- GORM SQL Injection Protection
- Templ XSS Auto-Escaping
- CSRF Protection (Echo Middleware)

✅ **Data Protection**:

- UUID statt Integer IDs
- Soft Deletes (GORM)
- Granulare Sharing-Berechtigungen

✅ **Infrastructure**:

- HTTPS-Ready (TLS/SSL)
- Secure Session Cookies
- Rate Limiting (Echo Middleware)

### Security Best Practices

**1. Secrets Management**:

```bash
# Niemals in .env committen
SESSION_SECRET=change-me-in-production

# Verwende starke Secrets (32+ Zeichen)
SESSION_SECRET=$(openssl rand -base64 32)
```

**2. CSRF Tokens**:

```html
<!-- In Formularen immer CSRF Token -->
<form method="POST">
  <input type="hidden" name="csrf" value="{{.csrf_token}}" />
</form>
```

**3. Input Sanitization**:

```go
// GORM escaped automatisch
database.DB.Where("email = ?", userInput).First(&user)

// Templ escaped automatisch
{ userInput }
```

### Security Monitoring

**Verdächtige Aktivität** (Viele Deletions in kurzer Zeit):

```sql
SELECT
    user_id,
    COUNT(*) AS deletions,
    MIN(created_at) AS first_delete,
    MAX(created_at) AS last_delete
FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '1 hour'
GROUP BY user_id
HAVING COUNT(*) > 10
ORDER BY deletions DESC;
```

---

## 🚨 Incident Response

### Common Issues

#### 1. Hohe Deletion-Rate

**Detection**:

```sql
SELECT COUNT(*) FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '1 hour';
```

**Action**:

1. User Account sperren
2. Audit Logs reviewen
3. Soft-deleted Items wiederherstellen falls nötig

#### 2. Unauthorized Access

**Detection**:

- Fehlerhafte Login-Versuche in Logs
- Zugriffe ohne Session

**Action**:

1. IP-Adresse blockieren (Firewall)
2. Passwörter zurücksetzen
3. Sessions invalidieren

#### 3. Data Loss

**Recovery**:

```sql
-- Soft-deleted Items wiederherstellen
UPDATE cards
SET deleted_at = NULL
WHERE id = '<card_uuid>';

-- Aus Audit Logs rekonstruieren
SELECT resource_data FROM audit_logs
WHERE resource_type = 'cards'
AND resource_id = '<card_uuid>';
```

---

## 📈 Performance Monitoring

### Key Metrics

**Database Performance**:

```sql
-- Slow Queries (via GORM Logs)
-- Check application logs for queries > 100ms
```

**Audit Log Growth**:

```sql
SELECT
    DATE_TRUNC('day', created_at) AS day,
    COUNT(*) AS entries
FROM audit_logs
GROUP BY day
ORDER BY day DESC
LIMIT 30;
```

### Optimization

**Index Usage**:

```sql
-- Verify index usage
EXPLAIN ANALYZE
SELECT * FROM audit_logs
WHERE user_id = '<uuid>'
ORDER BY created_at DESC;
```

**JSONB Performance**:

```sql
-- Optional: GIN Index für JSONB Queries
CREATE INDEX idx_audit_logs_resource_data_gin
ON audit_logs USING GIN (resource_data jsonb_path_ops);
```

---

## 🔄 Backup & Recovery

### Database Backups

**Daily Backup**:

```bash
# Full database backup
pg_dump -h localhost -U savvy -d savvy -F c -f savvy_$(date +%Y%m%d).dump

# Audit logs only
pg_dump -t audit_logs -h localhost -U savvy -d savvy > audit_logs_$(date +%Y%m%d).sql
```

**Restore**:

```bash
# Full restore
pg_restore -h localhost -U savvy -d savvy savvy_20260125.dump

# Audit logs restore
psql -h localhost -U savvy -d savvy < audit_logs_20260125.sql
```

### Data Retention Policy

**Empfohlene Retention**:

- Audit Logs: 7 Jahre (Compliance)
- Soft-deleted Items: 30 Tage
- Application Logs: 90 Tage

**Cleanup Script**:

```sql
-- Alte Audit Logs (> 7 Jahre)
DELETE FROM audit_logs
WHERE created_at < NOW() - INTERVAL '7 years';

-- Alte Soft-Deletes (> 30 Tage)
DELETE FROM cards
WHERE deleted_at < NOW() - INTERVAL '30 days'
AND deleted_at IS NOT NULL;
```

---

**Ende Operations Guide**

Für weitere technische Details siehe:

- `AGENTS.md` - Entwickler-Dokumentation
- `README.md` - Benutzer-Dokumentation
- `migrations/README.md` - Datenbank-Schema
