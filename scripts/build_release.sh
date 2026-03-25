!/usr/bin/env bash
set -euo pipefail
APP_NAME="adnx_dns"
VERSION="1.0.0"
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
DIST_DIR="$ROOT_DIR/dist"
PKG_DIR="$DIST_DIR/${APP_NAME}_${VERSION}_linux_amd64"
mkdir -p "$PKG_DIR"
rm -rf "$PKG_DIR"/*
cd "$ROOT_DIR"
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$PKG_DIR/${APP_NAME}" ./cmd/server
cp README.md .env.example schema.sql install.sh RELEASE.md "$PKG_DIR/"
mkdir -p "$PKG_DIR/scripts"
cp scripts/build_release.sh "$PKG_DIR/scripts/"
cd "$DIST_DIR"
tar -czf "${APP_NAME}_${VERSION}_linux_amd64.tar.gz" "${APP_NAME}_${VERSION}_linux_amd64"
echo "Built: $DIST_DIR/${APP_NAME}_${VERSION}_linux_amd64.tar.gz"
