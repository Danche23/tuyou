#!/usr/bin/env bash
set -euo pipefail

app=${1:?web or admin required}
case "$app:$(id -un)" in
  web:tuyou-web-deploy|admin:tuyou-admin-deploy) ;;
  *) echo 'Unauthorized deployment target' >&2; exit 1 ;;
esac

if [[ ! ${SSH_ORIGINAL_COMMAND:-} =~ ^([0-9a-f]{40})[[:space:]]([0-9]+)$ ]]; then
  echo 'Expected commit SHA and workflow run number' >&2
  exit 1
fi
sha=${BASH_REMATCH[1]}
run_number=${BASH_REMATCH[2]}
root=/opt/tuyou/$app
umask 022

exec 9>"$root/.deploy.lock"
flock 9
if [[ -f "$root/.run-number" ]]; then
  read -r previous_run < "$root/.run-number"
  if (( 10#$run_number < 10#$previous_run )); then
    echo 'A newer deployment is already active'
    exit 0
  fi
fi

archive=$(mktemp "$root/.upload.XXXXXXXX")
staging="$root/releases/$sha.tmp.$$"
trap 'rm -f "$archive"; rm -rf "$staging"' EXIT
timeout 600 head -c 104857601 > "$archive"
if (( $(stat -c '%s' "$archive") > 104857600 )); then
  echo 'H5 archive exceeds 100 MiB' >&2
  exit 1
fi

mkdir -p "$staging"
python3 - "$archive" "$staging" <<'PY'
import pathlib
import sys
import tarfile

with tarfile.open(sys.argv[1], "r:gz") as archive:
    members = archive.getmembers()
    if sum(member.size for member in members) > 314572800:
        raise SystemExit("H5 archive exceeds 300 MiB unpacked")
    for member in members:
        path = pathlib.PurePosixPath(member.name)
        if member.name == "." and member.isdir():
            continue
        if path.is_absolute() or not path.parts or ".." in path.parts or not (member.isfile() or member.isdir()):
            raise SystemExit("Unsafe H5 archive entry")
    archive.extractall(sys.argv[2], members=members)
PY
if [[ ! -s "$staging/index.html" ]]; then
  echo 'H5 archive must contain index.html' >&2
  exit 1
fi
chmod -R u=rwX,go=rX "$staging"
if [[ -d "$root/releases/$sha" ]]; then
  rm -rf "$staging"
else
  mv "$staging" "$root/releases/$sha"
fi

ln -s "releases/$sha" "$root/.current.next.$$"
mv -Tf "$root/.current.next.$$" "$root/current"
printf '%s\n' "$run_number" > "$root/.run-number"
echo "$app deployed at $sha"
