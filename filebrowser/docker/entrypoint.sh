#!/bin/bash
set -e

FB_DATABASE=${FB_DATABASE:="/database/filebrowser.db"}

echo "
-------------------------------------
Filebrowser Initializing
-------------------------------------
"
echo "[-] Database File: $FB_DATABASE"

# Copy default config if missing
if [ ! -f "/config/settings.json" ]; then
  cp /defaults/settings.json /config/settings.json
fi

# Template the auth block on every start when FB_AUTH_METHOD is set, so
# switching auth modes takes effect after `docker compose restart filebrowser`.
# When unset, settings.json is left alone (legacy behaviour = FB's own auth).
if [ -n "$FB_AUTH_METHOD" ]; then
  FB_AUTH_HEADER=${FB_AUTH_HEADER:-X-Authentik-Username}
  tmp=$(mktemp)
  jq --arg m "$FB_AUTH_METHOD" --arg h "$FB_AUTH_HEADER" \
     '.auth = {method:$m, header:$h}' /config/settings.json > "$tmp" \
     && mv "$tmp" /config/settings.json
fi

# Ensure ownership
chown -R abc:abc /config /database /srv

# Initialize database
if [ ! -e "$FB_DATABASE" ]; then
  echo "[*] Database doesn't exist.  Initializing..."
  su-exec abc filebrowser -c /config/settings.json -d "$FB_DATABASE" &
  sleep 2
  kill %1 2>/dev/null || true
  wait 2>/dev/null || true
  su-exec abc filebrowser -c /config/settings.json -d "$FB_DATABASE" users update admin -p "$FB_DEFAULT_USER_PASSWORD"
else
  echo "[-] Database ${FB_DATABASE} already exists, skipping initialize..."
fi

# Create additional users
if [ -n "$FB_USERS" ]; then
  for user in $FB_USERS; do
    echo "[*] Creating user $user..."
    su-exec abc filebrowser -c /config/settings.json -d "$FB_DATABASE" users add "$user" "$FB_DEFAULT_USER_PASSWORD" --perm.admin 2>/dev/null || true
  done
fi

# Apply auth method from env on every start. FB stores the auth method in the
# DB (not settings.json) after the initial init, so we have to use `config set`
# to change it on an existing DB.
if [ -n "$FB_AUTH_METHOD" ]; then
  echo "[*] Setting auth method to $FB_AUTH_METHOD (header=${FB_AUTH_HEADER:-X-Authentik-Username})"
  su-exec abc filebrowser -c /config/settings.json -d "$FB_DATABASE" \
    config set --auth.method="$FB_AUTH_METHOD" \
    --auth.header="${FB_AUTH_HEADER:-X-Authentik-Username}" >/dev/null || true
fi

# Set up backup cron job (daily at 2am by default, override with FB_BACKUP_CRON)
FB_BACKUP_CRON=${FB_BACKUP_CRON:-"0 2 * * *"}
mkdir -p /backups
chown -R abc:abc /backups
echo "$FB_BACKUP_CRON /usr/local/bin/backup-filebrowser.sh >> /var/log/backup.log 2>&1" | crontab -u abc -
crond

echo "[*] Backup cron job scheduled: $FB_BACKUP_CRON"

# Start filebrowser
echo '[*] Starting filebrowser'
exec su-exec abc filebrowser -c /config/settings.json -d "$FB_DATABASE"
