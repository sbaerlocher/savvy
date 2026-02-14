# Savvy Observability Guide

Complete guide for monitoring and observing the Savvy application using OpenTelemetry and the Grafana Stack.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Components](#components)
- [Dashboards](#dashboards)
- [Troubleshooting](#troubleshooting)

---

## Overview

Savvy uses a modern observability stack based on OpenTelemetry and Grafana:

- **Logs**: Structured JSON logs exported via OTLP to Grafana Loki
- **Traces**: Distributed tracing for HTTP requests and database queries via Grafana Tempo
- **Metrics**: Prometheus metrics for HTTP, database, and business metrics

### Key Features

✅ Full-Stack Observability (Backend Go + Frontend SvelteKit)  
✅ OpenTelemetry Standard (OTLP export for logs, traces, and metrics)  
✅ Grafana Stack (Loki, Tempo, Prometheus, Grafana)  
✅ Auto-Instrumentation (Automatic HTTP and database tracing)  
✅ Correlation (Link logs ↔ traces ↔ metrics via trace IDs)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Savvy Application                       │
│  ┌──────────────┐              ┌──────────────┐            │
│  │   Backend    │              │   Frontend   │            │
│  │   (Go/Echo)  │              │  (SvelteKit) │            │
│  │  - slog      │              │  - console   │            │
│  │  - otelecho  │              │  - fetch API │            │
│  │  - GORM      │              │  - errors    │            │
│  └──────┬───────┘              └──────┬───────┘            │
│         │ OTLP (4318)                 │ OTLP (proxy)       │
└─────────┼─────────────────────────────┼───────────────────┘
          v                             v
┌─────────────────────────────────────────────────────────────┐
│            OpenTelemetry Collector (4317/4318)              │
│  Receivers:  OTLP (gRPC, HTTP), Prometheus Scrape          │
│  Processors: Batch, Memory Limiter, Resource Detection     │
│  Exporters:  Loki, Tempo, Prometheus                       │
└─────────────────────────────────────────────────────────────┘
          │                    │                    │
          v                    v                    v
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  Grafana Loki   │  │  Grafana Tempo  │  │   Prometheus    │
│     (Logs)      │  │    (Traces)     │  │   (Metrics)     │
│  Retention: 7d  │  │  Retention: 7d  │  │  Retention: 7d  │
└─────────────────┘  └─────────────────┘  └─────────────────┘
          └────────────────────┴────────────────────┘
                               │
                               v
                    ┌─────────────────────┐
                    │      Grafana        │
                    │  http://localhost:3000
                    └─────────────────────┘
```

---

## Quick Start

### 1. Start Observability Stack

```bash
# Start with observability stack (Loki, Tempo, Prometheus, Grafana)
docker compose --profile observability up

# Or set as default
export COMPOSE_PROFILES=observability
docker compose up
```

### 2. Enable OpenTelemetry

```bash
# Edit .env file
OTEL_ENABLED=true
VITE_OTEL_ENABLED=true

# Restart services
docker compose restart api client
```

### 3. Access Grafana

Open: `http://localhost:3000`

Default Credentials: `admin` / `admin`

### 4. Explore Data

Grafana is pre-configured with datasources:

- **Loki** (Logs)
- **Tempo** (Traces)
- **Prometheus** (Metrics)

---

## Components

### Ports

| Service             | Port | Description                |
| ------------------- | ---- | -------------------------- |
| Grafana             | 3000 | Visualization & Dashboards |
| Loki                | 3100 | Log Aggregation            |
| Tempo               | 3200 | Distributed Tracing        |
| Prometheus          | 9090 | Metrics Storage            |
| OTEL Collector gRPC | 4317 | OTLP Receiver (gRPC)       |
| OTEL Collector HTTP | 4318 | OTLP Receiver (HTTP)       |

### Configuration Files

- `otel-collector-config.yaml` - OTEL Collector configuration
- `loki-config.yaml` - Loki configuration
- `tempo-config.yaml` - Tempo configuration
- `prometheus.yml` - Prometheus configuration
- `grafana-datasources.yaml` - Grafana datasources

---

## Dashboards

### Example Queries

**Loki (Logs)**:

```logql
# All logs
{service_name="savvy"}

# Error logs only
{service_name="savvy"} |= "error"

# Logs with trace ID
{service_name="savvy"} | json | trace_id != ""
```

**Prometheus (Metrics)**:

```promql
# HTTP request rate
rate(http_requests_total[5m])

# HTTP request duration (p95)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Resource counts
cards_total
vouchers_total
gift_cards_total

# Resource status breakdown
vouchers_by_status
gift_cards_by_status

# Sharing metrics
shares_total

# Authentication
login_attempts_total

# Notification preferences (per-channel toggles)
push_notifications_enabled_total
email_notifications_enabled_total
push_reminders_enabled_total
push_sharing_enabled_total
email_reminders_enabled_total
email_sharing_enabled_total

# Push & email infrastructure
push_subscriptions_total
push_subscribed_users_total
email_verified_users_total
```

---

## Troubleshooting

### Backend Not Sending Data

Check logs:

```bash
docker compose logs api | grep -i otel
```

Expected output:

```
📊 OpenTelemetry: Enabled
📝 OpenTelemetry Logs: Enabled
📈 OpenTelemetry Metrics: Enabled
```

### Frontend Not Sending Data

Check browser console - should see:

```
📊 OpenTelemetry Tracing: Enabled
📝 OpenTelemetry Logging: Enabled
📈 OpenTelemetry Metrics: Enabled
```

### OTEL Collector Issues

Check collector logs:

```bash
docker compose --profile observability logs otel-collector
```

---

## References

- OpenTelemetry: https://opentelemetry.io/
- Grafana Loki: https://grafana.com/oss/loki/
- Grafana Tempo: https://grafana.com/oss/tempo/
- Prometheus: https://prometheus.io/

---

**Last Updated**: 2026-02-11  
**Version**: 1.8.0
