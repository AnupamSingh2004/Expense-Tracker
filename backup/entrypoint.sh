#!/usr/bin/env sh
set -eu

# Write cron job using environment variables
CRON_FILE="/etc/crontabs/root"
echo "${BACKUP_SCHEDULE} PGHOST=${PGHOST} PGPORT=${PGPORT:-5432} PGDATABASE=${PGDATABASE} PGUSER=${PGUSER} PGPASSWORD=${PGPASSWORD} BACKUP_DIR=${BACKUP_DIR} BACKUP_RETAIN_DAYS=${BACKUP_RETAIN_DAYS} /usr/local/bin/backup.sh >> /var/log/backup.log 2>&1" > "$CRON_FILE"

echo "[backup-sidecar] cron schedule: ${BACKUP_SCHEDULE}"
echo "[backup-sidecar] target db: ${PGDATABASE}@${PGHOST}"

# Run an immediate backup on startup to verify connectivity
/usr/local/bin/backup.sh

# Start cron in foreground
exec crond -f -l 8
