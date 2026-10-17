#!/usr/bin/env bash
set -euo pipefail

docker run --rm --memory 256m \
  -v /opt/tuyou/letsencrypt:/etc/letsencrypt:z \
  -v /opt/tuyou/letsencrypt-var:/var/lib/letsencrypt:z \
  -v /var/www/evenstar/ssl:/webroot:z \
  docker.m.daocloud.io/certbot/certbot:v3.3.0 \
  renew --quiet

certificate=/opt/tuyou/letsencrypt/live/tuyou.evenstar.love
install -m 644 "$certificate/fullchain.pem" /var/www/evenstar/ssl/tuyou.evenstar.love.pem
install -m 600 "$certificate/privkey.pem" /var/www/evenstar/ssl/tuyou.evenstar.love.key
docker exec evenstar-web nginx -t
docker exec evenstar-web nginx -s reload
