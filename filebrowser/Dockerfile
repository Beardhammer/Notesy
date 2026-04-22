# ---- Frontend build stage ----
FROM node:20-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---- Backend build stage ----
FROM golang:1.22-alpine AS backend
RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -o /out/filebrowser .

# ---- Runtime stage ----
FROM alpine:3.19

RUN apk --update add ca-certificates \
                      mailcap \
                      curl \
                      jq \
                      bash \
                      su-exec

# Health
HEALTHCHECK --start-period=2s --interval=5s --timeout=3s \
  CMD curl -f http://localhost/health || exit 1

# Copy config and entrypoint
COPY docker/root/defaults/settings.json /defaults/settings.json
COPY docker/entrypoint.sh /entrypoint.sh
COPY docker/backup-filebrowser.sh /usr/local/bin/backup-filebrowser.sh
RUN chmod +x /entrypoint.sh /usr/local/bin/backup-filebrowser.sh

# Copy binary from build stage
COPY --from=backend /out/filebrowser /usr/bin/filebrowser
RUN chmod +x /usr/bin/filebrowser

# Create directories and user
RUN adduser -D -u 1000 abc && \
    mkdir -p /config /database /srv /backups && \
    chown -R abc:abc /config /database /srv /backups

VOLUME /config /database /srv /backups
EXPOSE 80

ENTRYPOINT ["/entrypoint.sh"]
