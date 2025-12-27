# Monitoring

This document describes the observability stack for Landly: logs, metrics, and dashboards.

## Overview

Landly uses:
- **Structured Logging**: Zap logger with JSON output and trace correlation
- **Metrics**: Prometheus metrics exposed at `/metrics` endpoint
- **Dashboards**: Grafana dashboards for visualization

## Logs

### Format

All logs are structured JSON with consistent fields:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "caller": "services/auth.go:85",
  "msg": "user logged in",
  "service": "landly-backend",
  "version": "1.0.0",
  "env": "production",
  "trace_id": "abc123",
  "user_id": "uuid-here"
}
```

### Standard Fields

| Field | Description |
|-------|-------------|
| `level` | Log level: debug, info, warn, error, fatal |
| `timestamp` | ISO8601 timestamp |
| `caller` | Source file and line |
| `msg` | Log message |
| `service` | Service name (landly-backend) |
| `version` | Application version |
| `env` | Environment (development/staging/production) |
| `trace_id` | Request trace ID for correlation |
| `user_id` | Authenticated user ID (when available) |

### Log Levels

- **DEBUG**: Detailed debugging information (development only)
- **INFO**: General operational events
- **WARN**: Unexpected but non-critical events
- **ERROR**: Errors that need attention
- **FATAL**: Critical errors causing shutdown

## Metrics

### Endpoint

Prometheus metrics are available at:
```
GET /metrics
```

### Available Metrics

#### HTTP Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `landly_http_requests_total` | Counter | method, path, status | Total HTTP requests |
| `landly_http_request_duration_seconds` | Histogram | method, path | Request latency |
| `landly_http_requests_in_flight` | Gauge | - | Current active requests |

#### Database Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `landly_db_pool_open_connections` | Gauge | Open DB connections |
| `landly_db_pool_in_use` | Gauge | Connections in use |
| `landly_db_pool_idle` | Gauge | Idle connections |
| `landly_db_pool_wait_count_total` | Gauge | Total waits for connection |

#### Business Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `landly_user_registrations_total` | Counter | - | User signups |
| `landly_user_logins_total` | Counter | - | Successful logins |
| `landly_user_login_failures_total` | Counter | - | Failed login attempts |
| `landly_projects_created_total` | Counter | - | Projects created |
| `landly_projects_deleted_total` | Counter | - | Projects deleted |
| `landly_generations_total` | Counter | type, status | Site generations |
| `landly_generation_duration_seconds` | Histogram | type | Generation time |
| `landly_publish_operations_total` | Counter | action | Publish/unpublish ops |

## Local Development

### Start Monitoring Stack

```bash
# Start all services including monitoring
docker-compose --profile monitoring up -d

# Or start only monitoring on top of existing services
docker-compose --profile monitoring up -d prometheus grafana
```

### Access

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)

### Pre-configured Dashboards

Grafana comes with the "Landly Overview" dashboard showing:
- HTTP request rate and latency
- Error rates (4xx, 5xx)
- Database connection pool
- Business metrics (registrations, logins, generations, publishes)

## Production Setup

### Option 1: Yandex Cloud Monitoring

Yandex Cloud provides native monitoring. Configure the backend to export metrics:

1. Install Yandex Unified Agent on your container/VM
2. Configure the agent to scrape `/metrics` endpoint
3. View metrics in Yandex Monitoring console

### Option 2: External Prometheus (Grafana Cloud)

1. Sign up for Grafana Cloud (free tier available)
2. Get your Prometheus remote write endpoint and credentials
3. Deploy Prometheus with remote write configuration:

```yaml
remote_write:
  - url: https://prometheus-prod-xx.grafana.net/api/prom/push
    basic_auth:
      username: <your-username>
      password: <your-api-key>
```

### Option 3: Self-hosted Prometheus

Deploy Prometheus alongside your backend:

1. Create a VM/container for Prometheus
2. Configure scraping of backend `/metrics` endpoint
3. Set up persistent storage for metrics
4. Configure retention period (default 15 days)

### Alerting

Example alert rules for Prometheus:

```yaml
groups:
  - name: landly-alerts
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(landly_http_requests_total{status=~"5.."}[5m]))
          / sum(rate(landly_http_requests_total[5m])) > 0.01
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected
          
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, sum(rate(landly_http_request_duration_seconds_bucket[5m])) by (le)) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High request latency (p95 > 2s)
          
      - alert: DBConnectionPoolExhausted
        expr: landly_db_pool_in_use / landly_db_pool_open_connections > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: Database connection pool nearly exhausted
```

## Troubleshooting

### Metrics Not Appearing

1. Check if `/metrics` endpoint returns data:
   ```bash
   curl http://localhost:8080/metrics
   ```

2. Verify Prometheus is scraping:
   - Go to http://localhost:9090/targets
   - Check if `landly-backend` target is UP

3. Check Prometheus logs:
   ```bash
   docker-compose logs prometheus
   ```

### Grafana Dashboard Empty

1. Verify Prometheus datasource is configured correctly
2. Check time range in Grafana (default: Last 1 hour)
3. Generate some traffic to produce metrics

### Log Correlation

To trace a request through logs:
1. Get the `trace_id` from response header `X-Request-ID`
2. Search logs: `grep "trace_id\":\"<your-trace-id>\"" logs.json`

