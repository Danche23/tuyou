# TuYou on the existing 2 GiB ECS

This stack is isolated from the Evenstar Compose project. It uses its own
MySQL, Redis, volumes, and Docker network. The API test port is bound on
`127.0.0.1:9301`, and the two H5 sites on `127.0.0.1:9080` and `9081`.
Do not expose them publicly before DNS, HTTPS, SMS rate limits, backups,
and production SMS configuration are ready.

Create a private `.env` beside `compose.yml` on the server, with
`TUYOU_DB_ROOT_PASSWORD`, `TUYOU_DB_PASSWORD`, `TUYOU_REDIS_PASSWORD`,
`TUYOU_JWT_SECRET`, and `TUYOU_IMAGE`. Generate each password independently
with `openssl rand -hex 32`, keep the file mode at `0600`, and pin
`TUYOU_IMAGE` to a tested full commit SHA tag from GHCR.

Run `docker compose --env-file .env -f compose.yml up -d` from this directory.
Check `docker compose ps`, `curl -f http://127.0.0.1:9301/api/v1/health`,
`free -h`, and `docker stats --no-stream`. Do not run `down -v`: that removes
the TuYou database and uploaded files. The application runs GORM AutoMigrate
at startup; back up its MySQL volume before updating the API image.

The current `sms.driver` field does not implement a local/log sender despite
the older example config. Without real Aliyun credentials, SMS workflows will
fail. This is intentional while the API is only reachable from the server.

Install `backup.sh` in `/opt/tuyou` and the two `tuyou-backup.*` units in
`/etc/systemd/system`, then run `systemctl enable --now tuyou-backup.timer`.
Backups are written under `/opt/tuyou/backups` with mode `0700`. They are
local copies only; arrange off-server retention before treating them as
disaster recovery. This script does not delete older backups automatically.

If GHCR is unreachable during the initial deployment, cross-compile the
unmodified backend commit for `linux/amd64` with `CGO_ENABLED=0`, copy the
binary into a separate `runtime/` directory alongside `Dockerfile.runtime`,
and build an image there. Never use `/opt/tuyou` as a Docker build context:
it contains `.env` secrets. Pin `TUYOU_IMAGE` to that local image until GHCR
pulls work reliably.
