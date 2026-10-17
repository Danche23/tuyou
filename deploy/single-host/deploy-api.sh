#!/usr/bin/env bash
set -euo pipefail

if (( EUID != 0 )); then
  echo 'API deployment requires the restricted sudo command' >&2
  exit 1
fi

if [[ ! ${SSH_ORIGINAL_COMMAND:-} =~ ^([0-9a-f]{40})[[:space:]]([0-9]+)$ ]]; then
  echo 'Expected commit SHA and workflow run number' >&2
  exit 1
fi
sha=${BASH_REMATCH[1]}
run_number=${BASH_REMATCH[2]}
root=/opt/tuyou
cd "$root"

exec 9>"$root/.api-deploy.lock"
flock 9
if [[ -f "$root/.api-run-number" ]]; then
  read -r previous_run < "$root/.api-run-number"
  if (( 10#$run_number < 10#$previous_run )); then
    echo 'A newer API deployment is already active'
    exit 0
  fi
fi

archive=$(mktemp "$root/.api-upload.XXXXXXXX")
staging=$(mktemp -d "$root/.api-build.XXXXXXXX")
trap 'rm -f "$archive"; rm -rf "$staging"' EXIT
timeout 600 head -c 104857601 > "$archive"
if (( $(stat -c '%s' "$archive") > 104857600 )); then
  echo 'API archive exceeds 100 MiB' >&2
  exit 1
fi

python3 - "$archive" "$staging" <<'PY'
import sys
import tarfile

with tarfile.open(sys.argv[1], "r:gz") as archive:
    members = archive.getmembers()
    files = [m for m in members if m.isfile()]
    if len(files) != 1 or files[0].name != "tuyou-api" or files[0].size > 104857600:
        raise SystemExit("Expected one tuyou-api binary under 100 MiB")
    if any(m.name not in (".", "tuyou-api") for m in members):
        raise SystemExit("Unexpected API archive entry")
    archive.extract(files[0], sys.argv[2])
PY
chmod 755 "$staging/tuyou-api"
docker build -q -f "$root/runtime/Dockerfile.runtime" -t "tuyou-api:sha-$sha" "$staging" > /dev/null

previous_image=$(sed -n 's/^TUYOU_IMAGE=//p' "$root/.env")
if [[ -z "$previous_image" ]]; then
  echo 'Current image is not configured' >&2
  exit 1
fi

bash "$root/backup.sh"
cp -p "$root/.env" "$root/.env.bak-api-$sha"
sed -i "s|^TUYOU_IMAGE=.*$|TUYOU_IMAGE=tuyou-api:sha-$sha|" "$root/.env"

healthy=false
if docker compose --env-file .env -f compose.yml up -d --no-deps api; then
  for _ in $(seq 1 30); do
    if [[ $(docker inspect tuyou-api-1 --format '{{.State.Health.Status}}' 2>/dev/null || true) == healthy ]]; then
      healthy=true
      break
    fi
    sleep 5
  done
fi

if [[ "$healthy" == true ]]; then
  if ! docker exec tuyou-web-1 nginx -s reload || \
    ! curl -fsS -o /dev/null http://127.0.0.1:9080/api/v1/health || \
    ! curl -fsS -o /dev/null http://127.0.0.1:9081/api/v1/health; then
    healthy=false
  fi
fi

if [[ "$healthy" != true ]]; then
  sed -i "s|^TUYOU_IMAGE=.*$|TUYOU_IMAGE=$previous_image|" "$root/.env"
  docker compose --env-file .env -f compose.yml up -d --no-deps api
  docker exec tuyou-web-1 nginx -s reload
  echo 'API health check failed; previous image restored' >&2
  exit 1
fi

printf '%s\n' "$run_number" > "$root/.api-run-number"
echo "API deployed at $sha"
