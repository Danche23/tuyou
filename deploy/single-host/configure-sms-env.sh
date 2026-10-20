#!/usr/bin/env bash
set -euo pipefail

umask 077
read -r -s -p 'Aliyun AccessKey ID: ' access_key_id
printf '\n'
read -r -s -p 'Aliyun AccessKey Secret: ' access_key_secret
printf '\n'

if [[ ! "$access_key_id" =~ ^[A-Za-z0-9]+$ || ! "$access_key_secret" =~ ^[A-Za-z0-9]+$ ]]; then
  echo 'AccessKey values must contain only letters and digits.' >&2
  exit 1
fi

tmp_file=$(mktemp /opt/tuyou/sms.env.XXXXXX)
trap 'rm -f "$tmp_file"' EXIT
printf 'SMS_ACCESS_KEY_ID=%s\nSMS_ACCESS_KEY_SECRET=%s\n' \
  "$access_key_id" "$access_key_secret" > "$tmp_file"
chmod 600 "$tmp_file"
mv -f "$tmp_file" /opt/tuyou/sms.env
echo 'SMS credentials saved to /opt/tuyou/sms.env.'
