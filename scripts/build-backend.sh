#!/bin/bash
# 8GB Mac 等で Docker 内ビルドが OOM になる場合に使用。
# Mac 上でバイナリをビルドし、Docker ではそのバイナリだけを実行する。

set -e
if ! command -v go >/dev/null 2>&1; then
  echo "Error: Go がインストールされていません。brew install go などで入れてから再実行してください。"
  exit 1
fi
cd "$(dirname "$0")/../backend"

# Docker コンテナは linux。Mac が arm64 なら linux/arm64、それ以外は amd64
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
  GOARCH=arm64
else
  GOARCH=amd64
fi

echo "Building backend for linux/$GOARCH (low-memory settings)..."
GOOS=linux GOARCH=$GOARCH GOGC=50 GOMAXPROCS=1 go build -p 1 -o main .

echo "Done. Binary: backend/main"
echo "Next: docker-compose -f docker-compose.yml -f docker-compose.binary.yml up -d db backend frontend"
