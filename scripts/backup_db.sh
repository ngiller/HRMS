#!/bin/bash
# ============================================================
# HRMS Database Backup Script
# Usage: ./scripts/backup_db.sh [--s3] [--retention N]
#
# Environment variables (or defaults):
#   DB_HOST=localhost  DB_PORT=5432  DB_USER=magnum
#   DB_PASSWORD=magnum  DB_NAME=hrms
#   BACKUP_DIR=./backups  RETENTION_DAYS=30
#   S3_BUCKET=""  (optional, set to enable S3 upload)
#   S3_PREFIX="hrms-backups"
# ============================================================

set -euo pipefail

# ─── Configuration ──────────────────────────────────────────

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-magnum}"
DB_PASSWORD="${DB_PASSWORD:-magnum}"
DB_NAME="${DB_NAME:-hrms}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
S3_BUCKET="${S3_BUCKET:-}"
S3_PREFIX="${S3_PREFIX:-hrms-backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/hrms_${TIMESTAMP}.dump"
LOG_FILE="${BACKUP_DIR}/backup.log"

# ─── Help ───────────────────────────────────────────────────

if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
    echo "Usage: $0 [--s3] [--retention N]"
    echo ""
    echo "Options:"
    echo "  --s3           Upload backup to S3 after creation"
    echo "  --retention N  Override retention days (default: 30)"
    echo "  --help, -h     Show this help"
    exit 0
fi

# Parse arguments
for arg in "$@"; do
    case "$arg" in
        --s3) S3_UPLOAD=true ;;
        --retention=*) RETENTION_DAYS="${arg#*=}" ;;
    esac
done

# ─── Setup ──────────────────────────────────────────────────

mkdir -p "$BACKUP_DIR"

log() {
    local level="$1"
    local message="$2"
    echo "[$(date +%Y-%m-%d\ %H:%M:%S)] [$level] $message" | tee -a "$LOG_FILE"
}

# ─── Pre-flight Checks ──────────────────────────────────────

if ! command -v pg_dump &>/dev/null; then
    log "ERROR" "pg_dump not found. Install PostgreSQL client tools."
    exit 1
fi

# Export password for non-interactive connection
export PGPASSWORD="$DB_PASSWORD"

# Test connection
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" &>/dev/null; then
    log "ERROR" "Cannot connect to PostgreSQL at $DB_HOST:$DB_PORT/$DB_NAME"
    exit 1
fi

log "INFO" "Starting backup of $DB_NAME@$DB_HOST:$DB_PORT"

# ─── Backup ─────────────────────────────────────────────────

# Using custom format (compressed, restorable with pg_restore)
pg_dump \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    --format=custom \
    --compress=9 \
    --verbose \
    --file="$BACKUP_FILE" 2>>"$LOG_FILE"

BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
log "INFO" "Backup created: $BACKUP_FILE ($BACKUP_SIZE)"

# Also generate a plain SQL backup for easy inspection
SQL_FILE="${BACKUP_DIR}/hrms_${TIMESTAMP}.sql"
pg_dump \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    --format=plain \
    --no-owner \
    --file="$SQL_FILE" 2>>"$LOG_FILE"

gzip -f "$SQL_FILE"
log "INFO" "SQL backup created: ${SQL_FILE}.gz"

# ─── S3 Upload (Optional) ──────────────────────────────────

if [ "${S3_UPLOAD:-false}" = true ] && [ -n "$S3_BUCKET" ]; then
    if command -v aws &>/dev/null; then
        log "INFO" "Uploading to S3: s3://$S3_BUCKET/$S3_PREFIX/"
        aws s3 cp "$BACKUP_FILE" "s3://${S3_BUCKET}/${S3_PREFIX}/$(basename $BACKUP_FILE)" --storage-class STANDARD_IA
        aws s3 cp "${SQL_FILE}.gz" "s3://${S3_BUCKET}/${S3_PREFIX}/$(basename ${SQL_FILE}).gz" --storage-class STANDARD_IA
        log "INFO" "S3 upload complete"
    else
        log "WARN" "aws CLI not found, skipping S3 upload"
    fi
fi

# ─── Retention Cleanup ──────────────────────────────────────

log "INFO" "Cleaning up backups older than $RETENTION_DAYS days"
find "$BACKUP_DIR" -name "hrms_*.dump" -mtime "+$RETENTION_DAYS" -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "hrms_*.sql.gz" -mtime "+$RETENTION_DAYS" -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "hrms_*.sql" -mtime "+$RETENTION_DAYS" -delete 2>/dev/null || true
log "INFO" "Cleanup complete"

# ─── Summary ────────────────────────────────────────────────

log "SUCCESS" "Backup complete: $BACKUP_FILE ($BACKUP_SIZE)"
log "INFO" "Retention: $RETENTION_DAYS days"
echo ""
echo "═══════════════════════════════════════════"
echo "  ✅ Backup Selesai"
echo "  📁 $BACKUP_FILE ($BACKUP_SIZE)"
echo "  📁 ${SQL_FILE}.gz"
echo "  📋 Log: $LOG_FILE"
echo "═══════════════════════════════════════════"

# Clean up password from environment
unset PGPASSWORD
