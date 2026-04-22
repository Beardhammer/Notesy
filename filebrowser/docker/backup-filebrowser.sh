#!/bin/bash
# =============================================================================
# FileBrowser Kanban & Events Backup Script
# Runs inside the container via crond. Edit the variables below as needed.
# =============================================================================

# --- EDIT THESE ---
FILEBROWSER_URL="http://localhost"
USERNAME="admin"
PASSWORD="${FB_DEFAULT_USER_PASSWORD:-abc123}"
BACKUP_DIR="/backups"
KEEP_DAYS=30
# ------------------

DATE=$(date +%Y-%m-%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Login and get JWT token
TOKEN=$(curl -sf -X POST "$FILEBROWSER_URL/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | tr -d '"')

if [ -z "$TOKEN" ]; then
  echo "[backup] $(date) - ERROR: Failed to authenticate" >&2
  exit 1
fi

# Fetch tasks and events
TASKS=$(curl -sf "$FILEBROWSER_URL/api/kanban" -H "X-Auth: $TOKEN")
EVENTS=$(curl -sf "$FILEBROWSER_URL/api/events" -H "X-Auth: $TOKEN")

# Default to empty arrays if endpoints return nothing
TASKS=${TASKS:-"[]"}
EVENTS=${EVENTS:-"[]"}

# Combine into backup JSON
echo "{\"exportedAt\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"tasks\":${TASKS},\"events\":${EVENTS}}" \
  | jq '.' > "$BACKUP_DIR/filebrowser-backup-$DATE.json"

echo "[backup] $(date) - Backup saved to $BACKUP_DIR/filebrowser-backup-$DATE.json"

# Remove backups older than KEEP_DAYS
find "$BACKUP_DIR" -name "filebrowser-backup-*.json" -mtime +$KEEP_DAYS -delete
