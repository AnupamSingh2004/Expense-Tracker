# Lightweight backup sidecar: pg_dump + cron
FROM postgres:16-alpine

RUN apk add --no-cache dcron

COPY backup/entrypoint.sh /entrypoint.sh
COPY scripts/backup.sh /usr/local/bin/backup.sh
RUN chmod +x /entrypoint.sh /usr/local/bin/backup.sh

# Default: run at 02:00 every day. Override BACKUP_SCHEDULE to change.
ENV BACKUP_SCHEDULE="0 2 * * *"
ENV BACKUP_DIR="/backups"
ENV BACKUP_RETAIN_DAYS="7"

VOLUME ["/backups"]

ENTRYPOINT ["/entrypoint.sh"]
