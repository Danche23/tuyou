#!/usr/bin/env bash
set -euo pipefail

cd /opt/tuyou
umask 077
backup_dir=/opt/tuyou/backups
install -d -m 700 "$backup_dir"
stamp=$(date -u +%Y%m%dT%H%M%SZ)
database_tmp="$backup_dir/$stamp-mysql.sql.gz.tmp"
uploads_tmp="$backup_dir/$stamp-uploads.tar.gz.tmp"
trap 'rm -f "$database_tmp" "$uploads_tmp"' EXIT

docker compose -f compose.yml exec -T mysql sh -c \
  'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysqldump --single-transaction --routines --triggers -u root tuyou' \
  | gzip -c > "$database_tmp"
test -s "$database_tmp"

docker run --rm \
  -v tuyou_uploads:/data:ro \
  -v "$backup_dir":/backup \
  docker.m.daocloud.io/library/alpine:3.20 \
  tar -czf "/backup/$stamp-uploads.tar.gz.tmp" -C /data .
test -s "$uploads_tmp"

mv "$database_tmp" "$backup_dir/$stamp-mysql.sql.gz"
mv "$uploads_tmp" "$backup_dir/$stamp-uploads.tar.gz"
chmod 600 "$backup_dir/$stamp-mysql.sql.gz" "$backup_dir/$stamp-uploads.tar.gz"
echo "TuYou backup created: $stamp"
