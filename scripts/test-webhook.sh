#!/bin/bash
# Test AlertManager webhook endpoint
# Usage: bash scripts/test-webhook.sh [host:port]

HOST="${1:-localhost:8080}"

echo "=== Sending test alert to $HOST ==="

curl -s -X POST "http://$HOST/webhooks/alertmanager" \
  -H "Content-Type: application/json" \
  -d '{
  "version": "4",
  "groupKey": "test-group",
  "status": "firing",
  "receiver": "sreagent",
  "groupLabels": {
    "alertname": "HighCPUUsage"
  },
  "commonLabels": {
    "alertname": "HighCPUUsage",
    "severity": "critical",
    "business_line": "payment",
    "tenant": "prod",
    "instance": "payment-api-01:9090"
  },
  "commonAnnotations": {
    "summary": "High CPU usage on payment-api-01",
    "description": "CPU usage has been above 90% for more than 5 minutes"
  },
  "externalURL": "http://alertmanager:9093",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "HighCPUUsage",
        "severity": "critical",
        "business_line": "payment",
        "tenant": "prod",
        "instance": "payment-api-01:9090",
        "job": "payment-api"
      },
      "annotations": {
        "summary": "High CPU usage on payment-api-01",
        "description": "CPU usage is at 95.2% on payment-api-01"
      },
      "startsAt": "2026-04-01T10:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus:9090/graph?g0.expr=cpu_usage",
      "fingerprint": "abc123def456"
    },
    {
      "status": "firing",
      "labels": {
        "alertname": "HighMemoryUsage",
        "severity": "warning",
        "business_line": "order",
        "tenant": "prod",
        "instance": "order-svc-03:9090",
        "job": "order-service"
      },
      "annotations": {
        "summary": "High memory usage on order-svc-03",
        "description": "Memory usage is at 88.5% on order-svc-03"
      },
      "startsAt": "2026-04-01T10:05:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus:9090/graph?g0.expr=mem_usage",
      "fingerprint": "xyz789ghi012"
    }
  ]
}' | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== Test resolve ==="

curl -s -X POST "http://$HOST/webhooks/alertmanager" \
  -H "Content-Type: application/json" \
  -d '{
  "version": "4",
  "groupKey": "test-group",
  "status": "resolved",
  "receiver": "sreagent",
  "alerts": [
    {
      "status": "resolved",
      "labels": {
        "alertname": "HighCPUUsage",
        "severity": "critical",
        "business_line": "payment"
      },
      "annotations": {
        "summary": "High CPU usage resolved"
      },
      "startsAt": "2026-04-01T10:00:00Z",
      "endsAt": "2026-04-01T10:30:00Z",
      "fingerprint": "abc123def456"
    }
  ]
}' | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== Done ==="
