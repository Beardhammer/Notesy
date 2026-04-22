# Notesy

A self-hosted notes & collaboration stack: FileBrowser + CodiMD + OnlyOffice +
CyberChef behind Caddy. Two deployment modes:

- **Legacy** — FileBrowser handles its own login. Single password per user.
  Zero external dependencies.
- **Authentik** — Single sign-on via OIDC (oauth2-proxy → Authentik), plus a
  small `session-bridge` service that issues short-lived transfer codes so
  off-network users can claim a 7-day session without reaching the IdP.

```
                 ┌─────────────┐
        :80/:443 │    Caddy    │
                 └──────┬──────┘
       ┌───────────┬────┴────┬────────────┐
       ▼           ▼         ▼            ▼
 filebrowser    codimd   onlyoffice   cyberchef
       │
       └─ (authentik mode adds: oauth2-proxy + session-bridge)
```


## Prerequisites

- Docker 24+ with the Compose plugin (`docker compose ...`)
- A DNS name pointing at this host
- (Authentik mode only) An Authentik instance with an OIDC provider configured
  for this app

You do **not** need Go or Node installed locally — both `filebrowser` and
`codimd` build from source inside multi-stage Docker images.


## Quick start

### 1. Configure

```bash
cp .env.example .env
$EDITOR .env
```

Fill in at minimum:

- `DNS_NAME` — what users type in the browser
- `DATA_DIR` — host path for persistent data (will be created)
- `FB_USERS` — space-separated list of usernames to provision
- `ONLYOFFICE_JWT_SECRET` — `openssl rand -hex 32`
- `CMD_SESSION_SECRET` — `openssl rand -hex 24`

For **Authentik** mode also set:

- `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_ISSUER_URL` (from your
  Authentik provider)
- `OAUTH2_PROXY_COOKIE_SECRET` — `openssl rand -base64 32 | head -c 32`
- `SESSION_BRIDGE_SECRET` — `openssl rand -hex 32`

For **Legacy** mode also set:

- `FB_DEFAULT_USER_PASSWORD` — initial password for every user in `FB_USERS`

### 2. Create the data directory

```bash
sudo mkdir -p "${DATA_DIR}"/{share,database,session-bridge}
sudo chown -R 1000:1000 "${DATA_DIR}"
```

`1000:1000` is the UID/GID that runs inside the FileBrowser and CodiMD
containers.

### 3. Bring it up

**Legacy (FileBrowser auth):**

```bash
docker compose -f docker-compose.legacy.yml up -d --build
```

**Authentik (SSO):**

```bash
docker compose -f docker-compose.authentik.yml up -d --build
```

The first build downloads Go and Node base images and compiles the
FileBrowser frontend + backend; expect 3–5 minutes on a cold cache.

### 4. Log in

Browse to `http://${DNS_NAME}/`.

- **Legacy:** sign in as one of the names in `FB_USERS` with
  `FB_DEFAULT_USER_PASSWORD`. Change it under Settings.
- **Authentik:** you'll be redirected to Authentik to sign in.


## Authentik provider setup

In Authentik, create an **OAuth2/OpenID Provider** with:

- **Redirect URI:** `http://<DNS_NAME>/oauth2/callback`
  (use `https://` if you front Caddy with TLS — set
  `OAUTH2_PROXY_COOKIE_SECURE: "true"` in the compose file too)
- **Signing key:** any RSA key
- **Scopes:** `openid`, `profile`, `email`

Then create an **Application** that uses this provider. The application slug
goes into `OIDC_ISSUER_URL`:

```
https://<your-authentik-host>/application/o/<application-slug>/
```

The usernames listed in `FB_USERS` must match the `username` claim Authentik
emits.


## Off-network access (Authentik mode)

`session-bridge` lets a signed-in user generate an 8-digit code on the
internal network, then redeem it from anywhere to get a 7-day cookie that
bypasses the IdP.

- **Generate (auth required):** `http://<DNS_NAME>/transfer/new`
- **Redeem (public):** `http://<DNS_NAME>/transfer/claim`

Codes are single-use and expire in 10 minutes. Anyone holding a valid code
becomes the issuing user for 7 days, so treat them like a password.


## Routing reference

Caddy routes everything under one hostname:

| Path | Service |
|------|---------|
| `/`              | FileBrowser |
| `/md/*`          | CodiMD |
| `/oo/*`          | OnlyOffice |
| `/cyberchef/*`   | CyberChef |
| `/oauth2/*`      | oauth2-proxy *(authentik only)* |
| `/transfer/*`    | session-bridge *(authentik only)* |


## Operations

**Logs**

```bash
docker compose -f docker-compose.<mode>.yml logs -f
```

**Stop**

```bash
docker compose -f docker-compose.<mode>.yml down
```

**Backups** — FileBrowser runs a daily cron at 02:00 inside the container
that snapshots Kanban tasks + events to `/backups` (a named volume). Override
the schedule with `FB_BACKUP_CRON` in the compose env.

**Switching modes** — Stop the running stack first; the two compose files
share container names, so they cannot run simultaneously.


## Layout

```
.
├── Caddyfile.authentik         # routes for SSO mode
├── Caddyfile.legacy            # routes for FileBrowser-auth mode
├── docker-compose.authentik.yml
├── docker-compose.legacy.yml
├── .env.example
├── filebrowser/                # FileBrowser (Go + Vue), built in Docker
├── codimd/                     # CodiMD (Node), built in Docker
└── session-bridge/             # tiny Go service for OIDC-bridging + transfer codes
```
