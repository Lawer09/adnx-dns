#!/usr/bin/env bash
set -euo pipefail

APP_NAME="adnx_dns"
APP_DIR="/opt/${APP_NAME}"
BIN_PATH="/usr/local/bin/${APP_NAME}"
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"

MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_DB="${MYSQL_DB:-adnx_dns}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-}"
API_TOKEN="${API_TOKEN:-replace_with_your_api_token}"
HTTP_ADDR="${HTTP_ADDR:-:8080}"
GODADDY_BASE_URL="${GODADDY_BASE_URL:-https://api.godaddy.com}"
GODADDY_API_KEY="${GODADDY_API_KEY:-}"
GODADDY_API_SECRET="${GODADDY_API_SECRET:-}"
DOMAIN_SYNC_INTERVAL_SECONDS="${DOMAIN_SYNC_INTERVAL_SECONDS:-300}"
GODADDY_REQUEST_TIMEOUT_SECONDS="${GODADDY_REQUEST_TIMEOUT_SECONDS:-15}"
GODADDY_RATE_LIMIT_PER_MINUTE="${GODADDY_RATE_LIMIT_PER_MINUTE:-60}"
RANDOM_SUBDOMAIN_LENGTH="${RANDOM_SUBDOMAIN_LENGTH:-8}"

if [[ $EUID -ne 0 ]]; then
  echo "Please run as root"
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y curl git ca-certificates mysql-client rsync

if ! command -v go >/dev/null 2>&1; then
  GO_VERSION="1.22.12"
  ARCH=$(dpkg --print-architecture)
  case "$ARCH" in
    amd64) GO_ARCH="amd64" ;;
    arm64) GO_ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
  esac
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o /tmp/go.tgz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tgz
  export PATH="/usr/local/go/bin:$PATH"
  if ! grep -q '/usr/local/go/bin' /etc/profile; then
    echo 'export PATH=/usr/local/go/bin:$PATH' >/etc/profile.d/go.sh
    chmod +x /etc/profile.d/go.sh
  fi
else
  export PATH="$(dirname $(command -v go)):$PATH"
fi

mkdir -p "$APP_DIR"
rsync -a --delete ./ "$APP_DIR/" --exclude .git --exclude dist || cp -a . "$APP_DIR"
cd "$APP_DIR"

cat > "$APP_DIR/.env" <<EOF
APP_ENV=prod
HTTP_ADDR=${HTTP_ADDR}
API_TOKEN=${API_TOKEN}
MYSQL_DSN=${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DB}?parseTime=true&charset=utf8mb4&loc=Local
GODADDY_BASE_URL=${GODADDY_BASE_URL}
GODADDY_API_KEY=${GODADDY_API_KEY}
GODADDY_API_SECRET=${GODADDY_API_SECRET}
DOMAIN_SYNC_INTERVAL_SECONDS=${DOMAIN_SYNC_INTERVAL_SECONDS}
GODADDY_REQUEST_TIMEOUT_SECONDS=${GODADDY_REQUEST_TIMEOUT_SECONDS}
GODADDY_RATE_LIMIT_PER_MINUTE=${GODADDY_RATE_LIMIT_PER_MINUTE}
RANDOM_SUBDOMAIN_LENGTH=${RANDOM_SUBDOMAIN_LENGTH}
EOF

MYSQL_AUTH_ARGS=()
if [[ -n "$MYSQL_ROOT_PASSWORD" ]]; then
  MYSQL_AUTH_ARGS=(-uroot "-p${MYSQL_ROOT_PASSWORD}" -h "$MYSQL_HOST" -P "$MYSQL_PORT")
else
  MYSQL_AUTH_ARGS=(-u"${MYSQL_USER}" "-p${MYSQL_PASSWORD}" -h "$MYSQL_HOST" -P "$MYSQL_PORT")
fi

mysql "${MYSQL_AUTH_ARGS[@]}" -e "CREATE DATABASE IF NOT EXISTS ${MYSQL_DB} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql "${MYSQL_AUTH_ARGS[@]}" "$MYSQL_DB" < "$APP_DIR/schema.sql"

/usr/local/go/bin/go mod tidy
/usr/local/go/bin/go build -o "$BIN_PATH" ./cmd/server
chmod +x "$BIN_PATH"

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=adnx_dns service
After=network.target

[Service]
Type=simple
WorkingDirectory=${APP_DIR}
ExecStart=${BIN_PATH}
Restart=always
RestartSec=3
EnvironmentFile=${APP_DIR}/.env

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable ${APP_NAME}
systemctl restart ${APP_NAME}

echo "Installed successfully"
echo "Service: systemctl status ${APP_NAME}"
echo "Health: curl http://127.0.0.1${HTTP_ADDR}/healthz"
