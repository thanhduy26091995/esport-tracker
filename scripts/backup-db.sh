#!/usr/bin/env bash
# backup-db.sh — PostgreSQL backup for esport_tracker with Firebase Storage upload
#
# Usage:
#   bash backup-db.sh              ← backup + upload to Firebase
#   bash backup-db.sh --local      ← backup only, skip Firebase upload
#   bash backup-db.sh --list       ← list local backups
#   bash backup-db.sh --list-remote← list backups in Firebase Storage
#   bash backup-db.sh --prune      ← delete local backups older than KEEP_DAYS
#
# Crontab example (daily at 2 AM):
#   0 2 * * * /path/to/esport/backup-db.sh >> /var/log/esport-backup.log 2>&1

set -eu

# ══════════════════════════════════════════════════════════════
# CONFIG — all values read from backend/.env
# ══════════════════════════════════════════════════════════════
BACKUP_DIR="$(cd "$(dirname "$0")/.." && pwd)/backups"
KEEP_DAYS=30

ENV_FILE="$(cd "$(dirname "$0")/.." && pwd)/backend/.env"
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

# ── Firebase upload helper ─────────────────────────────────────
firebase_upload() {
  local file="$1"
  local filename
  filename="$(basename "$file")"

  if [[ -z "$FIREBASE_BUCKET" ]] || [[ -z "$FIREBASE_SA_KEY" ]]; then
    warn "Firebase not configured (FIREBASE_BUCKET / FIREBASE_SA_KEY empty). Skipping upload."
    return 0
  fi

  if ! command -v gsutil >/dev/null 2>&1; then
    warn "gsutil not found — run: curl https://sdk.cloud.google.com | bash"
    warn "Then re-run this script to upload."
    return 0
  fi

  if [[ ! -f "$FIREBASE_SA_KEY" ]]; then
    warn "Service account key not found: $FIREBASE_SA_KEY"
    return 0
  fi

  log "Uploading to Firebase Storage: gs://${FIREBASE_BUCKET}/backups/${filename}"
  GOOGLE_APPLICATION_CREDENTIALS="$FIREBASE_SA_KEY" \
    gsutil -q cp "$file" "gs://${FIREBASE_BUCKET}/backups/${filename}"
  log "Upload complete."
}

# ── List local backups ─────────────────────────────────────────
if [[ "${1:-}" == "--list" ]]; then
  echo -e "\n${BLUE}── Local backups in $BACKUP_DIR ──${NC}"
  if [[ ! -d "$BACKUP_DIR" ]] || [[ -z "$(ls "$BACKUP_DIR"/*.sql.gz 2>/dev/null || true)" ]]; then
    echo "  (none)"
  else
    ls -lh "$BACKUP_DIR"/*.sql.gz 2>/dev/null | awk '{print "  " $5 "\t" $9}' || echo "  (none)"
  fi
  echo ""
  exit 0
fi

# ── List remote backups ────────────────────────────────────────
if [[ "${1:-}" == "--list-remote" ]]; then
  [[ -z "$FIREBASE_BUCKET" ]] && fail "FIREBASE_BUCKET not configured."
  command -v gsutil >/dev/null 2>&1 || fail "gsutil not found."
  echo -e "\n${BLUE}── Remote backups in gs://${FIREBASE_BUCKET}/backups/ ──${NC}"
  GOOGLE_APPLICATION_CREDENTIALS="$FIREBASE_SA_KEY" \
    gsutil ls -lh "gs://${FIREBASE_BUCKET}/backups/*.sql.gz" 2>/dev/null || echo "  (none)"
  echo ""
  exit 0
fi

# ── Prune local backups ────────────────────────────────────────
if [[ "${1:-}" == "--prune" ]]; then
  log "Pruning local backups older than ${KEEP_DAYS} days..."
  find "$BACKUP_DIR" -name "*.sql.gz" ! -name "latest.sql.gz" -mtime +"$KEEP_DAYS" -delete
  log "Prune done."
  exit 0
fi

# ── Backup ────────────────────────────────────────────────────
# Detect whether to use host pg_dump or Docker exec
DOCKER_CONTAINER="esport-postgres"
if command -v pg_dump >/dev/null 2>&1; then
  PG_DUMP_CMD="pg_dump --host=$DB_HOST --port=$DB_PORT --username=$DB_USER --no-password --format=plain --clean --if-exists $DB_NAME"
elif docker inspect "$DOCKER_CONTAINER" >/dev/null 2>&1; then
  log "pg_dump not on host — using Docker container '$DOCKER_CONTAINER'."
  PG_DUMP_CMD="docker exec -e PGPASSWORD=$PGPASSWORD $DOCKER_CONTAINER pg_dump --username=$DB_USER --format=plain --clean --if-exists $DB_NAME"
else
  fail "pg_dump not found and no Docker container '$DOCKER_CONTAINER' running. Install postgresql-client or start Docker."
fi

mkdir -p "$BACKUP_DIR"

TIMESTAMP="$(date '+%Y%m%d_%H%M%S')"
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_${TIMESTAMP}.sql.gz"
LATEST_LINK="$BACKUP_DIR/latest.sql.gz"

log "Starting backup of '$DB_NAME'..."

$PG_DUMP_CMD | gzip > "$BACKUP_FILE"

SIZE="$(du -sh "$BACKUP_FILE" | cut -f1)"
log "Backup saved: $BACKUP_FILE ($SIZE)"

# Update symlink to latest
ln -sf "$BACKUP_FILE" "$LATEST_LINK"

# Auto-prune old local backups
PRUNED="$(find "$BACKUP_DIR" -name "*.sql.gz" ! -name "latest.sql.gz" -mtime +"$KEEP_DAYS" | wc -l)"
if [[ "$PRUNED" -gt 0 ]]; then
  find "$BACKUP_DIR" -name "*.sql.gz" ! -name "latest.sql.gz" -mtime +"$KEEP_DAYS" -delete
  warn "Pruned $PRUNED local backup(s) older than ${KEEP_DAYS} days."
fi

# Upload to Firebase (skipped if --local flag passed)
if [[ "${1:-}" != "--local" ]]; then
  firebase_upload "$BACKUP_FILE"
fi

log "Backup complete."
