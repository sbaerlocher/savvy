# Savvy System - Grafana Dashboards

Grafana Dashboards für Monitoring & Observability des Savvy Systems.

## 📊 Available Dashboards

### savvy-overview.json

**Main Dashboard** mit umfassenden Metrics:

#### Overview Section
- **Total Users** - Gesamt-Anzahl User
- **Active Sessions** - Aktive User-Sessions (mit Thresholds: >50 gelb, >100 rot)
- **Total Cards** - Anzahl Kundenkarten
- **Total Vouchers** - Anzahl Gutscheine
- **Total Gift Cards** - Anzahl Geschenkkarten
- **Active DB Connections** - Aktive Datenbank-Verbindungen (mit Thresholds: >5 gelb, >10 rot)

#### HTTP Performance Section
- **HTTP Request Rate** - Requests/s by method + path
- **HTTP Request Latency** - p50, p95, p99 Latenz (Thresholds: >0.5s gelb, >1s rot)
- **HTTP Status Codes** - Status Code Distribution (stacked)
- **HTTP Errors** - 4xx + 5xx Errors by method + path

#### OpenTelemetry Traces Section
- **Recent Traces** - Letzte 20 Traces via Tempo
  - Query: `{service.name="savvy"}`
  - TraceQL Search Support
  - Click-through zu Trace Details

#### Database & Resources Section
- **Database Connections** - Active + Idle Connections
- **Memory Usage** - RSS + Heap Allocation
- **Garbage Collection Rate** - Go GC Rate (5min)
- **Goroutines** - Active Goroutines

#### Logs (Loki) Section
- **Recent Logs** - Letzte Logs via Loki
  - Query: `{job="savvy"}`
  - Log Details expandable
  - Descending sort (newest first)

## 🔧 Configuration

### Datasource Variables

Das Dashboard nutzt **Template Variables** für flexible Datasource-Konfiguration:

```json
{
  "datasource": "Prometheus",        // Prometheus Datasource
  "tempo_datasource": "Tempo",       // Tempo (Traces) Datasource
  "loki_datasource": "Loki"          // Loki (Logs) Datasource
}
```

**Vorteile**:
- ✅ Keine hardcoded Datasource UIDs
- ✅ Dashboard funktioniert in verschiedenen Grafana Instanzen
- ✅ Einfacher Wechsel zwischen Datasources

### Dashboard Settings

- **Refresh**: 30s Auto-Refresh
- **Time Range**: Last 6 hours (default)
- **Timezone**: Browser Timezone
- **Schema Version**: 39 (Grafana 11+)

## 📥 Import

### Via Grafana UI

1. **Login** to Grafana: `https://your-grafana.com`
2. **Navigate** to Dashboards → Import
3. **Upload JSON** oder **Paste JSON**:
   ```bash
   cat deploy/grafana/savvy-overview.json
   ```
4. **Select Datasources**:
   - Prometheus → Your Prometheus Datasource
   - Tempo → Your Tempo Datasource
   - Loki → Your Loki Datasource
5. **Import**

### Via Grafana API

```bash
# Set your Grafana URL + API Key
GRAFANA_URL="https://your-grafana.com"
GRAFANA_API_KEY="your-api-key"

# Import Dashboard
curl -X POST "${GRAFANA_URL}/api/dashboards/db" \
  -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
  -H "Content-Type: application/json" \
  -d @deploy/grafana/savvy-overview.json
```

### Via Terraform (Grafana Provider)

```hcl
resource "grafana_dashboard" "savvy_system_overview" {
  config_json = file("${path.module}/deploy/grafana/savvy-overview.json")
  folder      = grafana_folder.savvy_system.id
}
```

## 🎯 Required Metrics

Das Dashboard benötigt folgende **Prometheus Metrics**:

### Application Metrics (Custom)
```
# Business Metrics
users_total{job="savvy"}
active_sessions{job="savvy"}
cards_total{job="savvy"}
vouchers_total{job="savvy"}
gift_cards_total{job="savvy"}

# HTTP Metrics
http_requests_total{job="savvy", method, path, status}
http_request_duration_seconds_bucket{job="savvy", le}

# Database Metrics
db_connections_active{job="savvy"}
db_connections_idle{job="savvy"}
```

### Go Runtime Metrics (Standard)
```
# Memory
process_resident_memory_bytes{job="savvy"}
go_memstats_heap_alloc_bytes{job="savvy"}

# Garbage Collection
go_gc_duration_seconds_count{job="savvy"}

# Goroutines
go_goroutines{job="savvy"}
```

## 📊 OpenTelemetry Integration

### Traces (Tempo)

**Service Name**: `savvy`

**Query**: `{service.name="savvy"}`

**OTEL Configuration** (Staging + Production):
```yaml
env:
  - name: OTEL_ENABLED
    value: "true"
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "grafana-alloy.observability.svc.cluster.local:4318"
  - name: OTEL_SERVICE_NAME
    value: "savvy"
```

**Instrumentation** (Go):
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/sdk/trace"
)

// Initialize OTEL Tracer
tp := trace.NewTracerProvider(
    trace.WithBatcher(otlptrace.New(ctx, client)),
    trace.WithResource(resource.NewWithAttributes(
        semconv.ServiceNameKey.String("savvy"),
    )),
)
otel.SetTracerProvider(tp)
```

### Logs (Loki)

**Job Label**: `job="savvy"`

**Log Format**: Structured JSON Logging (Go slog)

**Grafana Alloy Config** (Loki Integration):
```yaml
# Grafana Alloy discovers logs via Pod annotations
loki.source.kubernetes "savvy_system" {
  targets = discovery.kubernetes.pods.targets
  forward_to = [loki.write.default.receiver]
}

loki.write "default" {
  endpoint {
    url = "https://logs.grafana.net/loki/api/v1/push"
  }
}
```

## 🔍 Usage Examples

### Find Slow Requests

1. **Navigate** to "HTTP Request Latency" panel
2. **Check** p95/p99 values
3. **If > 0.5s**: Click panel → Explore
4. **Run Query**:
   ```promql
   topk(10, rate(http_request_duration_seconds_bucket{job="savvy", le="1"}[5m]))
   ```

### Debug HTTP Errors

1. **Navigate** to "HTTP Errors" panel
2. **Identify** failing endpoints
3. **Click** panel → Explore → Switch to Loki
4. **Run Query**:
   ```logql
   {job="savvy"} |= "error" | json | status >= 500
   ```

### Trace Request Flow

1. **Navigate** to "Recent Traces" panel
2. **Click** on a Trace
3. **View** Span Timeline
4. **Identify** slow spans

### Check Database Performance

1. **Navigate** to "Database Connections" panel
2. **Check** Active/Idle ratio
3. **If Active > 5**: Potential connection leak
4. **Investigate** via Logs:
   ```logql
   {job="savvy"} |= "database" | json
   ```

## 🎨 Dashboard Customization

### Add New Panel

```json
{
  "datasource": {
    "type": "prometheus",
    "uid": "${datasource}"
  },
  "targets": [
    {
      "expr": "your_metric{job=\"savvy\"}",
      "legendFormat": "{{label}}",
      "refId": "A"
    }
  ],
  "title": "Your Panel Title",
  "type": "timeseries"
}
```

### Add Alerts

```json
{
  "alert": {
    "name": "High Error Rate",
    "conditions": [
      {
        "evaluator": {
          "params": [0.05],
          "type": "gt"
        },
        "query": {
          "params": ["A", "5m", "now"]
        },
        "reducer": {
          "type": "avg"
        },
        "type": "query"
      }
    ]
  }
}
```

## 📚 Links

- **Grafana Docs**: https://grafana.com/docs/
- **Prometheus Queries**: https://prometheus.io/docs/prometheus/latest/querying/basics/
- **Tempo TraceQL**: https://grafana.com/docs/tempo/latest/traceql/
- **Loki LogQL**: https://grafana.com/docs/loki/latest/logql/

## 🔗 Related Documentation

- **Main Deploy README**: [../README.md](../README.md)
- **Kustomize**: [../kustomize/README.md](../kustomize/README.md)
- **Helm**: [../helm/README.md](../helm/README.md)
- **Operations**: [../../OPERATIONS.md](../../OPERATIONS.md)
