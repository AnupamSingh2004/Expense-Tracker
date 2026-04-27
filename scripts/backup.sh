#!/usr/bin/env sh
# Dumps the PostgreSQL database to /backups/<timestamp>.sql.gz
# Required env: PGHOST, PGPORT, PGDATABASE, PGUSER, PGPASSWORD
# Optional env: BACKUP_RETAIN_DAYS (default 7)

set -eu

BACKUP_DIR="${BACKUP_DIR:-/backups}"
RETAIN_DAYS="${BACKUP_RETAIN_DAYS:-7}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
FILE="${BACKUP_DIR}/${PGDATABASE}_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

echo "[backup] starting dump → $FILE"
pg_dump \
  -h "$PGHOST" \
  -p "${PGPORT:-5432}" \
  -U "$PGUSER" \
  -d "$PGDATABASE" \
  --no-password \
  --format=plain \
  --no-owner \
  --no-acl \
  | gzip > "$FILE"

SIZE=$(du -sh "$FILE" | cut -f1)
echo "[backup] done — $FILE ($SIZE)"

# Prune old backups
find "$BACKUP_DIR" -name "*.sql.gz" -mtime "+${RETAIN_DAYS}" -delete
echo "[backup] pruned backups older than ${RETAIN_DAYS} days"
