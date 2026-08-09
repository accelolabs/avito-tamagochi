#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE_URL="${BASE_URL:-http://localhost:${BACKEND_PORT:-8080}}"
WS_URL="${WS_URL:-ws://localhost:${BACKEND_PORT:-8080}/ws}"
PROJECT="${LIVE_COMPOSE_PROJECT:-avito-tamagotchi-live}"
START_STACK=1
STOP_STACK=0
MODE="all"

for arg in "$@"; do
  case "$arg" in
    --no-start) START_STACK=0 ;;
    --stop) STOP_STACK=1 ;;
    --smoke) MODE="smoke" ;;
    --all) MODE="all" ;;
    *) echo "unknown argument: $arg" >&2; exit 2 ;;
  esac
done

TMP_DIR="$(mktemp -d)"
COOKIE_JAR="$TMP_DIR/cookies.txt"
WS_BIN="$TMP_DIR/live-ws-listener"
trap 'rm -rf "$TMP_DIR"' EXIT

command -v curl >/dev/null || { echo "curl is required" >&2; exit 1; }
command -v python3 >/dev/null || { echo "python3 is required" >&2; exit 1; }
command -v docker >/dev/null || { echo "docker is required" >&2; exit 1; }

COMPOSE=(docker compose -p "$PROJECT" -f "$ROOT_DIR/compose.yaml" -f "$ROOT_DIR/compose.live.yaml")
GO_CACHE_DIR="${GOCACHE:-/tmp/avito-go-cache}"

if [[ "$START_STACK" == 1 ]]; then
  echo "== start backend-only stack: $PROJECT =="
  "${COMPOSE[@]}" up --build -d postgres migrate backend
fi

echo "== build WebSocket listener =="
(cd "$ROOT_DIR/backend" && GOCACHE="$GO_CACHE_DIR" go build -o "$WS_BIN" ./cmd/live-ws-listener)

for attempt in $(seq 1 30); do
  if status="$(curl -sS -o /dev/null -w '%{http_code}' "$BASE_URL/api/v1/auth/me" 2>/dev/null)" && [[ "$status" == 401 ]]; then
    break
  fi
  [[ "$attempt" == 30 ]] && { echo "backend did not become ready" >&2; exit 1; }
  sleep 1
done

EMAIL="live-$(date +%s)-$RANDOM@example.com"
PASSWORD="password123"
DISPLAY_NAME="LiveTester"

die() { echo "FAIL: $*" >&2; exit 1; }

request() {
  local method="$1" path="$2" body="${3:-}" expected="$4"
  local response="$TMP_DIR/response.json"
  local status
  if [[ -n "$body" ]]; then
    status="$(curl -sS -o "$response" -w '%{http_code}' -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
      -H 'Content-Type: application/json' -X "$method" "$BASE_URL$path" --data "$body")"
  else
    status="$(curl -sS -o "$response" -w '%{http_code}' -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
      -X "$method" "$BASE_URL$path")"
  fi
  [[ "$status" == "$expected" ]] || { cat "$response" >&2; die "$method $path returned $status, expected $expected"; }
  cp "$response" "$TMP_DIR/last.json"
}

json_value() {
  local file="$1"; shift
  python3 - "$file" "$@" <<'PY'
import json, sys
value = json.load(open(sys.argv[1]))
for part in sys.argv[2:]:
    value = value[int(part)] if isinstance(value, list) else value[part]
print(json.dumps(value, separators=(',', ':')) if isinstance(value, (dict, list)) else value)
PY
}

assert_json() {
  python3 - "$TMP_DIR/last.json" "$@" <<'PY'
import json, sys
payload = json.load(open(sys.argv[1]))
for expression in sys.argv[2:]:
    path, expected = expression.split('=', 1)
    value = payload
    for part in path.split('.'):
        value = value[int(part)] if isinstance(value, list) else value[part]
    actual = str(value).lower() if isinstance(value, bool) else str(value)
    if actual != expected:
        raise SystemExit(f"expected {path}={expected}, got {actual}")
PY
}

wait_events() {
  local events="$1" cookie="$2" log="$3" pid
  "$WS_BIN" -url "$WS_URL" -cookie "$cookie" -event "$events" -timeout 15s >"$log" 2>&1 &
  pid=$!
  WS_PID="$pid"
}

finish_events() {
  wait "$WS_PID" || { cat "$1" >&2; die "WebSocket events not received"; }
}

echo "== auth =="
request POST /api/v1/auth/register "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"displayName\":\"$DISPLAY_NAME\"}" 201
USER_ID="$(json_value "$TMP_DIR/last.json" id)"
SESSION_ID="$(awk '$6 == "session_id" {print $7}' "$COOKIE_JAR")"
[[ -n "$SESSION_ID" ]] || die "session_id cookie was not returned"
request GET /api/v1/auth/me '' 200
assert_json "id=$USER_ID" "displayName=$DISPLAY_NAME"
request POST /api/v1/auth/login "{\"email\":\"$EMAIL\",\"password\":\"wrong-password\"}" 401
request POST /api/v1/auth/register "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"displayName\":\"$DISPLAY_NAME\"}" 409

echo "== unauthenticated access =="
EMPTY_COOKIE="$TMP_DIR/empty-cookies.txt"
status="$(curl -sS -o /dev/null -w '%{http_code}' -b "$EMPTY_COOKIE" "$BASE_URL/api/v1/pet")"
[[ "$status" == 401 ]] || die "unauthenticated pet request returned $status"

echo "== initial state =="
request GET /api/v1/pet '' 200
assert_json xp=0 level=1
INITIAL_ENERGY="$(json_value "$TMP_DIR/last.json" energy)"
[[ "$INITIAL_ENERGY" =~ ^(49|50|51)$ ]] || die "unexpected initial energy: $INITIAL_ENERGY"
request GET /api/v1/tasks/today '' 200
cp "$TMP_DIR/last.json" "$TMP_DIR/tasks.json"
TASK_COUNT="$(python3 - "$TMP_DIR/tasks.json" <<'PY'
import json, sys
print(len(json.load(open(sys.argv[1]))['tasks']))
PY
)"
[[ "$TASK_COUNT" == 3 ]] || die "expected 3 tasks, got $TASK_COUNT"
request GET /api/v1/rewards '' 200
assert_json 'rewards=[]'
request GET /api/v1/summary/today '' 200
assert_json xpEarned=0 completedTasks=0 charges=0 currentXP=0 level=1
request GET /api/v1/leaderboard '' 200
python3 - "$TMP_DIR/last.json" "$USER_ID" <<'PY'
import json, sys
data = json.load(open(sys.argv[1]))
if len(data['entries']) > 10:
    raise SystemExit('leaderboard returned more than top-10')
if data['currentUser']['userId'] != sys.argv[2]:
    raise SystemExit('current user is missing from leaderboard')
print('leaderboard shape: OK')
PY

echo "== charge + WebSocket =="
WS_LOG="$TMP_DIR/ws-charge.log"
wait_events pet_updated "$SESSION_ID" "$WS_LOG"
sleep 0.3
request POST /api/v1/pet/actions '"charge"' 200
finish_events "$WS_LOG"
grep -q 'event=pet_updated' "$WS_LOG" || die "pet_updated was not received"
request POST /api/v1/pet/actions '"charge"' 200
assert_json xp=10

echo "== mock Avito actions =="
while IFS=$'\t' read -r task_type required_count; do
  for ((step=1; step<=required_count; step++)); do
    if [[ "$step" == "$required_count" ]]; then
      TASK_WS_LOG="$TMP_DIR/ws-task-$task_type.log"
      wait_events tasks_updated,pet_updated "$SESSION_ID" "$TASK_WS_LOG"
      sleep 0.3
    fi
    request POST /api/v1/mock-avito/actions "\"$task_type\"" 200
    if [[ "$step" == "$required_count" ]]; then
      finish_events "$TASK_WS_LOG"
      grep -q 'event=tasks_updated' "$TASK_WS_LOG" || die "tasks_updated was not received"
      grep -q 'event=pet_updated' "$TASK_WS_LOG" || die "pet_updated was not received for completed task"
    fi
  done
done < <(python3 - "$TMP_DIR/tasks.json" <<'PY'
import json, sys
for task in json.load(open(sys.argv[1]))['tasks']:
    print(task['type'], task['requiredCount'], sep='\t')
PY
)
request GET /api/v1/tasks/today '' 200
python3 - "$TMP_DIR/last.json" <<'PY'
import json, sys
for task in json.load(open(sys.argv[1]))['tasks']:
    if task['progress'] > task['requiredCount']:
        raise SystemExit(f"progress overflow for {task['type']}")
    if task['progress'] != task['requiredCount'] or task['status'] != 'completed':
        raise SystemExit(f"task did not complete: {task['type']}")
print('progress bounds: OK')
PY

echo "== invalid and idempotent actions =="
request POST /api/v1/mock-avito/actions '"not-an-action"' 400
TASK_TYPE="$(json_value "$TMP_DIR/tasks.json" tasks 0 type)"
request POST /api/v1/mock-avito/actions "\"$TASK_TYPE\"" 200
request GET /api/v1/summary/today '' 200

if [[ "$MODE" == all ]]; then
  echo "== controlled reward flow =="
  REWARD_EMAIL="reward-$(date +%s)-$RANDOM@example.com"
  REWARD_COOKIE="$TMP_DIR/reward-cookies.txt"
  REWARD_STATUS="$(curl -sS -o "$TMP_DIR/reward-user.json" -w '%{http_code}' -c "$REWARD_COOKIE" -H 'Content-Type: application/json' \
    -X POST "$BASE_URL/api/v1/auth/register" \
    --data "{\"email\":\"$REWARD_EMAIL\",\"password\":\"$PASSWORD\",\"displayName\":\"RewardTester\"}")"
  [[ "$REWARD_STATUS" == 201 ]] || { cat "$TMP_DIR/reward-user.json" >&2; die "reward user registration failed"; }
  REWARD_USER_ID="$(json_value "$TMP_DIR/reward-user.json" id)"
  curl -sS -o /dev/null -b "$REWARD_COOKIE" "$BASE_URL/api/v1/pet"
  "${COMPOSE[@]}" exec -T postgres psql -U tamagotchi -d tamagotchi -v ON_ERROR_STOP=1 \
    -c "UPDATE pets SET xp = 95 WHERE user_id = '$REWARD_USER_ID';" >/dev/null
  REWARD_SESSION="$(awk '$6 == "session_id" {print $7}' "$REWARD_COOKIE")"
  REWARD_WS_LOG="$TMP_DIR/ws-reward.log"
  wait_events pet_updated,rewards_updated "$REWARD_SESSION" "$REWARD_WS_LOG"
  sleep 0.3
  curl -sS -o "$TMP_DIR/reward-pet.json" -b "$REWARD_COOKIE" -c "$REWARD_COOKIE" \
    -H 'Content-Type: application/json' -X POST "$BASE_URL/api/v1/pet/actions" --data '"charge"' >/dev/null
  finish_events "$REWARD_WS_LOG"
  grep -q 'event=rewards_updated' "$REWARD_WS_LOG" || die "rewards_updated was not received"
  curl -sS -o "$TMP_DIR/rewards.json" -b "$REWARD_COOKIE" "$BASE_URL/api/v1/rewards"
  REWARD_ID="$(json_value "$TMP_DIR/rewards.json" rewards 0 id)"
  REWARD_LEVEL="$(json_value "$TMP_DIR/rewards.json" rewards 0 level)"
  [[ "$REWARD_LEVEL" == 2 ]] || die "expected reward level 2, got $REWARD_LEVEL"
  curl -sS -o "$TMP_DIR/used-reward.json" -w '%{http_code}' -b "$REWARD_COOKIE" \
    -X POST "$BASE_URL/api/v1/rewards/$REWARD_ID/use" >"$TMP_DIR/use-status"
  [[ "$(cat "$TMP_DIR/use-status")" == 200 ]] || die "reward use failed"
  curl -sS -o /dev/null -w '%{http_code}' -b "$REWARD_COOKIE" \
    -X POST "$BASE_URL/api/v1/rewards/$REWARD_ID/use" >"$TMP_DIR/reuse-status"
  [[ "$(cat "$TMP_DIR/reuse-status")" == 409 ]] || die "reward reuse did not return 409"

  LEVEL3_EMAIL="level3-$(date +%s)-$RANDOM@example.com"
  LEVEL3_COOKIE="$TMP_DIR/level3-cookies.txt"
  LEVEL3_STATUS="$(curl -sS -o "$TMP_DIR/level3-user.json" -w '%{http_code}' -c "$LEVEL3_COOKIE" -H 'Content-Type: application/json' \
    -X POST "$BASE_URL/api/v1/auth/register" \
    --data "{\"email\":\"$LEVEL3_EMAIL\",\"password\":\"$PASSWORD\",\"displayName\":\"LevelTester\"}")"
  [[ "$LEVEL3_STATUS" == 201 ]] || { cat "$TMP_DIR/level3-user.json" >&2; die "level3 user registration failed"; }
  LEVEL3_USER_ID="$(json_value "$TMP_DIR/level3-user.json" id)"
  curl -sS -o /dev/null -b "$LEVEL3_COOKIE" "$BASE_URL/api/v1/pet"
  "${COMPOSE[@]}" exec -T postgres psql -U tamagotchi -d tamagotchi -v ON_ERROR_STOP=1 \
    -c "UPDATE pets SET xp = 195 WHERE user_id = '$LEVEL3_USER_ID';" >/dev/null
  LEVEL3_SESSION="$(awk '$6 == "session_id" {print $7}' "$LEVEL3_COOKIE")"
  LEVEL3_WS_LOG="$TMP_DIR/ws-level3.log"
  wait_events pet_updated,rewards_updated "$LEVEL3_SESSION" "$LEVEL3_WS_LOG"
  sleep 0.3
  curl -sS -o /dev/null -b "$LEVEL3_COOKIE" -c "$LEVEL3_COOKIE" \
    -H 'Content-Type: application/json' -X POST "$BASE_URL/api/v1/pet/actions" --data '"charge"'
  finish_events "$LEVEL3_WS_LOG"
  curl -sS -o "$TMP_DIR/level3-rewards.json" -b "$LEVEL3_COOKIE" "$BASE_URL/api/v1/rewards"
  [[ "$(json_value "$TMP_DIR/level3-rewards.json" rewards 0 level)" == 3 ]] || die "level 3 reward was not created"
  [[ "$(json_value "$TMP_DIR/level3-rewards.json" rewards 0 type)" == free_delivery ]] || die "level 3 reward type is wrong"

  echo "== concurrent charge =="
  CONCURRENT_EMAIL="concurrent-$(date +%s)-$RANDOM@example.com"
  CONCURRENT_COOKIE="$TMP_DIR/concurrent-cookies.txt"
  CONCURRENT_STATUS="$(curl -sS -o "$TMP_DIR/concurrent-user.json" -w '%{http_code}' -c "$CONCURRENT_COOKIE" -H 'Content-Type: application/json' \
    -X POST "$BASE_URL/api/v1/auth/register" \
    --data "{\"email\":\"$CONCURRENT_EMAIL\",\"password\":\"$PASSWORD\",\"displayName\":\"Concurrent\"}")"
  [[ "$CONCURRENT_STATUS" == 201 ]] || { cat "$TMP_DIR/concurrent-user.json" >&2; die "concurrent user registration failed"; }
  for n in 1 2; do
    curl -sS -o "$TMP_DIR/concurrent-$n.json" -w '%{http_code}' -b "$CONCURRENT_COOKIE" \
      -H 'Content-Type: application/json' -X POST "$BASE_URL/api/v1/pet/actions" \
      --data '"charge"' >"$TMP_DIR/concurrent-$n.status" &
  done
  wait
  [[ "$(cat "$TMP_DIR/concurrent-1.status")" == 200 && "$(cat "$TMP_DIR/concurrent-2.status")" == 200 ]] || die "concurrent charge failed"
  curl -sS -o "$TMP_DIR/concurrent-pet.json" -b "$CONCURRENT_COOKIE" "$BASE_URL/api/v1/pet"
  [[ "$(json_value "$TMP_DIR/concurrent-pet.json" xp)" == 10 ]] || die "concurrent charge duplicated XP"
fi

echo "== logout =="
request POST /api/v1/auth/logout '' 204
request GET /api/v1/auth/me '' 401

if [[ "$STOP_STACK" == 1 ]]; then
  "${COMPOSE[@]}" down -v
fi

echo "LIVE BACKEND $MODE: PASS"
