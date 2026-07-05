#!/usr/bin/env bash
# scripts/smoke-test.sh — end-to-end validation of the deployed config-service
set -euo pipefail

NAMESPACE="${NAMESPACE:-config-service}"
PORT="${PORT:-8080}"
BASE_URL="http://localhost:${PORT}"
PASS=0
FAIL=0

# ─── Helpers ──────────────────────────────────────────────────────────────────
green() { printf '\033[32m✅ %s\033[0m\n' "$*"; }
red()   { printf '\033[31m❌ %s\033[0m\n' "$*"; }
info()  { printf '\033[34m→  %s\033[0m\n' "$*"; }

check() {
  local label="$1" expected_status="$2" actual_status="$3" body="$4"
  if [[ "$actual_status" == "$expected_status" ]]; then
    green "$label (HTTP $actual_status)"
    PASS=$((PASS + 1))
  else
    red "$label — expected HTTP $expected_status, got $actual_status"
    echo "   body: $body"
    FAIL=$((FAIL + 1))
  fi
}

# ─── Port-forward ─────────────────────────────────────────────────────────────
info "Starting port-forward for svc/config-service on :${PORT}"
kubectl port-forward -n "${NAMESPACE}" svc/config-service "${PORT}:${PORT}" &>/dev/null &
PF_PID=$!
trap 'kill ${PF_PID} 2>/dev/null || true' EXIT

# Wait until the port is accepting connections (up to 15s)
for i in $(seq 1 15); do
  if curl -sf "${BASE_URL}/ping" &>/dev/null; then
    break
  fi
  sleep 1
done

echo ""
info "Running smoke tests against ${BASE_URL}"
echo ""

# ─── 1. Health check ──────────────────────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" "${BASE_URL}/ping")
body=$(cat /tmp/body)
check "GET /ping returns pong" "200" "$resp" "$body"
if [[ "$body" != "pong" ]]; then
  red "GET /ping body mismatch — expected 'pong', got '${body}'"
  FAIL=$((FAIL + 1))
fi

# ─── 2. Create config ─────────────────────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" -X POST "${BASE_URL}/configs" \
  -H "Content-Type: application/json" \
  -d '{"id":"smoke_1","host":"localhost","port":9090,"app_name":"smoke-test","log_level":"DEBUG"}')
body=$(cat /tmp/body)
check "POST /configs creates config" "200" "$resp" "$body"

# ─── 3. Read config back ──────────────────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" "${BASE_URL}/configs/smoke_1")
body=$(cat /tmp/body)
check "GET /configs/smoke_1 returns config" "200" "$resp" "$body"

# ─── 4. Upsert (update) ───────────────────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" -X POST "${BASE_URL}/configs" \
  -H "Content-Type: application/json" \
  -d '{"id":"smoke_1","host":"updated-host","port":9090,"app_name":"smoke-test","log_level":"INFO"}')
body=$(cat /tmp/body)
check "POST /configs upserts existing config" "200" "$resp" "$body"

# ─── 5. Verify update persisted ───────────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" "${BASE_URL}/configs/smoke_1")
body=$(cat /tmp/body)
check "GET /configs/smoke_1 reflects update" "200" "$resp" "$body"
if echo "$body" | grep -q '"updated-host"'; then
  green "Update persisted correctly"
  PASS=$((PASS + 1))
else
  red "Update not persisted — host field missing 'updated-host'"
  FAIL=$((FAIL + 1))
fi

# ─── 6. Not found ─────────────────────────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" "${BASE_URL}/configs/nonexistent")
body=$(cat /tmp/body)
check "GET /configs/nonexistent returns 404" "404" "$resp" "$body"

# ─── 7. Validation — missing host ─────────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" -X POST "${BASE_URL}/configs" \
  -H "Content-Type: application/json" \
  -d '{"id":"bad_1","host":"","port":8080,"app_name":"test","log_level":"INFO"}')
body=$(cat /tmp/body)
check "POST /configs with empty host returns 400" "400" "$resp" "$body"

# ─── 8. Validation — invalid port ─────────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" -X POST "${BASE_URL}/configs" \
  -H "Content-Type: application/json" \
  -d '{"id":"bad_2","host":"host","port":99999,"app_name":"test","log_level":"INFO"}')
body=$(cat /tmp/body)
check "POST /configs with invalid port returns 400" "400" "$resp" "$body"

# ─── 9. Validation — invalid log level ────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" -X POST "${BASE_URL}/configs" \
  -H "Content-Type: application/json" \
  -d '{"id":"bad_3","host":"host","port":8080,"app_name":"test","log_level":"VERBOSE"}')
body=$(cat /tmp/body)
check "POST /configs with invalid log_level returns 400" "400" "$resp" "$body"

# ─── 10. Malformed JSON ───────────────────────────────────────────────────────
resp=$(curl -s -o /tmp/body -w "%{http_code}" -X POST "${BASE_URL}/configs" \
  -H "Content-Type: application/json" \
  -d 'not-json')
body=$(cat /tmp/body)
check "POST /configs with malformed JSON returns 400" "400" "$resp" "$body"

# ─── Summary ──────────────────────────────────────────────────────────────────
echo ""
echo "────────────────────────────────"
printf "Results: \033[32m%d passed\033[0m, \033[31m%d failed\033[0m\n" "$PASS" "$FAIL"
echo "────────────────────────────────"

[[ "$FAIL" -eq 0 ]] && exit 0 || exit 1
