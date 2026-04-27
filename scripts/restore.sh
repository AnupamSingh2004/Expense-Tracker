#!/usr/bin/env sh
# Restores a backup file into the target database.
# Usage: restore.sh <path-to-backup.sql.gz>
# Required env: PGHOST, PGPORT, PGDATABASE, PGUSER, PGPASSWORD

set -eu

FILE="${1:-}"
if [ -z "$FILE" ]; then
  echo "Usage: $0 <backup-file.sql.gz>" >&2
  exit 1
fi
if [ ! -f "$FILE" ]; then
  echo "File not found: $FILE" >&2
  exit 1
fi

echo "[restore] dropping and recreating schema in $PGDATABASE …"
psql -h "$PGHOST" -p "${PGPORT:-5432}" -U "$PGUSER" -d "$PGDATABASE" \
  -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" \
  --no-password

echo "[restore] loading $FILE …"
gunzip -c "$FILE" | psql \
  -h "$PGHOST" \
  -p "${PGPORT:-5432}" \
  -U "$PGUSER" \
  -d "$PGDATABASE" \
  --no-password \
  --quiet

echo "[restore] complete — $FILE restored into $PGDATABASE"
