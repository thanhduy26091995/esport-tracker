#!/usr/bin/env bash
# restore-db.sh — Restore esport_tracker from a local or Firebase Storage backup
#
# Usage:
#   bash restore-db.sh                                       ← restore from local latest
#   bash restore-db.sh backups/esport_tracker_20260514.sql.gz← restore from local file
#   bash restore-db.sh --from-firebase                       ← restore from latest in Firebase
#   bash restore-db.sh --from-firebase esport_20260514.sql.gz← restore specific file from Firebase
#
# WARNING: This drops and recreates all tables. All current data will be lost.

set -eu

# ══════════════════════════════════════════════════════════════
# CONFIG — all values read from backend/.env
# ══════════════════════════════════════════════════════════════
BACKUP_DIR="$(cd "$(dirname "$0")" && pwd)/backups"

ENV_FILE="$(cd "$(dirname "$0")" && pwd)/backend/.env"
# ══════════════════════════════════════════════════════════════

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date '+%Y-%m-%d %T')]${NC} $*"; }
warn() { echo -e "${YELLOW}[$(date '+%Y-%m-%d %T')] WARN:${NC} $*"; }
fail() { echo -e "${RED}[$(date '+%Y-%m-%d %T')] ERROR:${NC} $*"; exit 1; }

[[ -f "$ENV_FILE" ]] || fail ".env not found at $ENV_FILE"

_env() { grep "^${1}=" "$ENV_FILE" | cut -d= -f2-; }

DB_HOST="$(_env DB_HOST)"
DB_PORT="$(_env DB_PORT)"
DB_USER="$(_env DB_USER)"
DB_NAME="$(_env DB_NAME)"
FIREBASE_BUCKET="$(_env FIREBASE_BUCKET)"
FIREBASE_SA_KEY="$(_env FIREBASE_SA_KEY)"
export PGPASSWORD="$(_env DB_PASSWORD)"

command -v gunzip >/dev/null 2>&1 || fail "gunzip not found."

# Detect whether to use host psql or Docker exec
DOCKER_CONTAINER="esport-postgres"
if command -v psql >/dev/null 2>&1; then
  PSQL_CMD="psql --host=$DB_HOST --port=$DB_PORT --username=$DB_USER --dbname=$DB_NAME --no-password --quiet"
elif docker inspect "$DOCKER_CONTAINER" >/dev/null 2>&1; then
  warn "psql not on host — using Docker container '$DOCKER_CONTAINER'."
  PSQL_CMD="docker exec -i -e PGPASSWORD=$PGPASSWORD $DOCKER_CONTAINER psql --username=$DB_USER --dbname=$DB_NAME --quiet"
else
  fail "psql not found and no Docker container '$DOCKER_CONTAINER' running. Install postgresql-client or start Docker."
fi

# ── Resolve backup file ────────────────────────────────────────
BACKUP_FILE=""
FROM_FIREBASE=false

if [[ "${1:-}" == "--from-firebase" ]]; then
  FROM_FIREBASE=true
  [[ -z "$FIREBASE_BUCKET" ]] && fail "FIREBASE_BUCKET not configured in restore-db.sh."
  [[ -z "$FIREBASE_SA_KEY" ]]  && fail "FIREBASE_SA_KEY not configured in restore-db.sh."
  command -v gsutil >/dev/null 2>&1 || fail "gsutil not found — run: curl https://sdk.cloud.google.com | bash"

  if [[ -n "${2:-}" ]]; then
    REMOTE_FILE="gs://${FIREBASE_BUCKET}/backups/${2}"
  else
    # Pick the most recent file in Firebase
    log "Finding latest backup in Firebase Storage..."
    REMOTE_FILE="$(GOOGLE_APPLICATION_CREDENTIALS="$FIREBASE_SA_KEY" \
      gsutil ls "gs://${FIREBASE_BUCKET}/backups/*.sql.gz" 2>/dev/null \
      | sort | tail -1)"
    [[ -z "$REMOTE_FILE" ]] && fail "No backups found in gs://${FIREBASE_BUCKET}/backups/"
  fi

  mkdir -p "$BACKUP_DIR"
  LOCAL_COPY="$BACKUP_DIR/$(basename "$REMOTE_FILE")"

  log "Downloading $REMOTE_FILE..."
  GOOGLE_APPLICATION_CREDENTIALS="$FIREBASE_SA_KEY" \
    gsutil -q cp "$REMOTE_FILE" "$LOCAL_COPY"
  log "Downloaded to $LOCAL_COPY"
  BACKUP_FILE="$LOCAL_COPY"

elif [[ -n "${1:-}" ]]; then
  BACKUP_FILE="$1"

else
  BACKUP_FILE="$BACKUP_DIR/latest.sql.gz"
fi

[[ -f "$BACKUP_FILE" ]] || fail "Backup file not found: $BACKUP_FILE"

# ── Confirm ────────────────────────────────────────────────────
SIZE="$(du -sh "$BACKUP_FILE" | cut -f1)"
MTIME="$(date -r "$BACKUP_FILE" '+%Y-%m-%d %T' 2>/dev/null \
  || stat -c '%y' "$BACKUP_FILE" 2>/dev/null | cut -d. -f1)"

echo ""
echo -e "${YELLOW}══════════════════════════════════════════${NC}"
echo -e "${YELLOW}  WARNING: Destructive operation!${NC}"
echo -e "${YELLOW}  Database : $DB_NAME${NC}"
if [[ "$FROM_FIREBASE" == "true" ]]; then
  echo -e "${YELLOW}  Source   : Firebase → $REMOTE_FILE${NC}"
fi
echo -e "${YELLOW}  File     : $BACKUP_FILE${NC}"
echo -e "${YELLOW}  Size     : $SIZE  (created $MTIME)${NC}"
echo -e "${YELLOW}  All current data will be overwritten.${NC}"
echo -e "${YELLOW}══════════════════════════════════════════${NC}"
echo ""
read -r -p "Type YES to confirm restore: " CONFIRM
[[ "$CONFIRM" == "YES" ]] || { echo "Aborted."; exit 0; }

# Stop backend during restore to prevent dirty writes
if systemctl is-active --quiet esport-backend 2>/dev/null; then
  log "Stopping esport-backend service..."
  sudo systemctl stop esport-backend
  RESTART_BACKEND=true
else
  RESTART_BACKEND=false
fi

log "Restoring from $BACKUP_FILE..."
gunzip -c "$BACKUP_FILE" | $PSQL_CMD

log "Restore complete."

if [[ "$RESTART_BACKEND" == "true" ]]; then
  log "Restarting esport-backend service..."
  sudo systemctl start esport-backend
  sleep 2
  if systemctl is-active --quiet esport-backend; then
    log "Backend restarted successfully."
  else
    warn "Backend failed to restart. Check: sudo journalctl -u esport-backend -n 50"
  fi
fi

log "Done. Database '$DB_NAME' restored."
