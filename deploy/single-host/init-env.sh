#!/usr/bin/env bash
set -euo pipefail

cd /opt/tuyou
if [[ -e .env ]]; then
  echo '/opt/tuyou/.env already exists; leaving it unchanged'
  exit 0
fi

umask 077
{
  printf 'TUYOU_DB_ROOT_PASSWORD=%s\n' "$(openssl rand -hex 32)"
  printf 'TUYOU_DB_PASSWORD=%s\n' "$(openssl rand -hex 32)"
  printf 'TUYOU_REDIS_PASSWORD=%s\n' "$(openssl rand -hex 32)"
  printf 'TUYOU_JWT_SECRET=%s\n' "$(openssl rand -hex 48)"
  printf 'TUYOU_IMAGE=%s\n' 'ghcr.io/danche23/tuyou:c7d06621ddd4bb8a6a732ec78f63ff93300cdd6c'
} > .env
chmod 600 .env
echo 'Created /opt/tuyou/.env with private credentials'
