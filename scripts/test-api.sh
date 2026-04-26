#!/bin/bash
# Test core API endpoints
# Usage: bash scripts/test-api.sh [host:port]

HOST="${1:-localhost:8080}"
BASE="http://$HOST/api/v1"

echo "=== 1. Login ==="
TOKEN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)

if [ -z "$TOKEN" ]; then
  echo "ERROR: Login failed"
  exit 1
fi
echo "Token: ${TOKEN:0:20}..."

AUTH="Authorization: Bearer $TOKEN"

echo ""
echo "=== 2. Get Profile ==="
curl -s -H "$AUTH" "$BASE/auth/profile" | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== 3. Create DataSource ==="
curl -s -X POST "$BASE/datasources" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production VictoriaMetrics",
    "type": "victoriametrics",
    "endpoint": "http://vm.example.com:8428",
    "description": "Main VM cluster",
    "labels": {"env": "production", "cluster": "main"}
  }' | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== 4. List DataSources ==="
curl -s -H "$AUTH" "$BASE/datasources" | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== 5. Create Alert Rule ==="
curl -s -X POST "$BASE/alert-rules" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "high_cpu_usage",
    "display_name": "High CPU Usage",
    "description": "Alert when CPU usage exceeds 90%",
    "datasource_id": 1,
    "expression": "avg(cpu_usage_percent) by (instance) > 90",
    "for_duration": "5m",
    "severity": "critical",
    "labels": {"business_line": "payment", "team": "infra"},
    "annotations": {"summary": "CPU usage above 90%"},
    "group_name": "infrastructure"
  }' | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== 6. List Alert Events ==="
curl -s -H "$AUTH" "$BASE/alert-events" | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== 7. Dashboard Stats ==="
curl -s -H "$AUTH" "$BASE/dashboard/stats" | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== 8. Create Team ==="
curl -s -X POST "$BASE/teams" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Payment Team",
    "description": "Responsible for payment services",
    "labels": {"business_line": "payment"}
  }' | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== 9. Create Notify Media (v2) ==="
curl -s -X POST "$BASE/notify-media" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Payment Alerts Lark Group",
    "type": "lark_webhook",
    "description": "Lark group for payment team alerts",
    "config": {"webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id"},
    "is_enabled": true
  }' | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== 10. Create Notify Rule (v2) ==="
curl -s -X POST "$BASE/notify-rules" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Payment Critical to Lark",
    "description": "Route payment critical alerts to Lark group",
    "match_labels": {"business_line": "payment"},
    "severities": "critical,warning",
    "media_id": 1,
    "is_enabled": true
  }' | python3 -m json.tool 2>/dev/null || cat

echo ""
echo "=== All tests completed ==="
