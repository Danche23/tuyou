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
Create a separate, mode `0600` `sms.env` beside the Compose file. Run
`configure-sms-env.sh` from an interactive server console to enter the Aliyun
AccessKey without shell history or Git; `sms.env` is required by Compose.
Set the approved `scheme_name`, `sign_name`, `template_code`, and
`template_param` in the private server `config.yaml` before restarting API.

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

The `web` service also joins the existing `evenstar_net` network as
`tuyou-h5`, allowing the blog's public Nginx to proxy the two H5 ports without
publishing them directly. The shared network must exist before starting this
Compose stack. The blog's Nginx must serve ACME HTTP-01 files from its
read-only `/etc/nginx/ssl` mount and proxy the two HTTPS hostnames to
`http://tuyou-h5:8080` and `http://tuyou-h5:8081`.
Append `blog-vhosts.conf` to the blog's existing Nginx site config only when
production SMS and upload flows have been validated. Test with `nginx -t`
before reloading; preserve the blog's existing server blocks.

For push-triggered H5 delivery, `deploy-h5.sh` is the forced SSH command for
separate `tuyou-web-deploy` and `tuyou-admin-deploy` users. Each key is bound
to one target in `authorized_keys`; the workflow streams its tested `dist/`
archive and commit SHA. The script checks the workflow run number and switches
the `current` symlink only after a complete extraction. Keep the keys out of
the repository and store each private key only in its corresponding GitHub
Actions secret.
The H5 Nginx `/editorial/` location serves photos and credits from
`/opt/tuyou/web/editorial/`, outside release directories. Back up this directory
separately before replacing the host; H5 deployments do not copy it from Git.
For the initial files already in each site directory, `current` may point to
`.`. A later release switches it to `releases/<sha>` without copying the
initial tree into itself. Restore that initial site with a new `current`
symlink to `.` if the first delivered release must be rolled back.

`deploy-api.sh` is the forced SSH command for a separate `tuyou-api-deploy`
key. It accepts a single Linux binary archive from the tested main-branch
workflow, builds a local image from the root-owned runtime Dockerfile, and
backs up MySQL and uploads before changing `TUYOU_IMAGE`. A failed API health
check restores the prior image. Keep the prior image and local backups for
manual recovery if a schema migration itself is incompatible.

After the first certificate is issued into `/opt/tuyou/letsencrypt`, install
`renew-cert.sh` in `/opt/tuyou` and `tuyou-cert-renew.*` in
`/etc/systemd/system`, then enable the timer. The renewal job copies the
current certificate into the blog's existing SSL mount and reloads Nginx.

If GHCR is unreachable during the initial deployment, cross-compile the
unmodified backend commit for `linux/amd64` with `CGO_ENABLED=0`, copy the
binary into a separate `runtime/` directory alongside `Dockerfile.runtime`,
and build an image there. Never use `/opt/tuyou` as a Docker build context:
it contains `.env` secrets. Pin `TUYOU_IMAGE` to that local image until GHCR
pulls work reliably.
