#!/usr/bin/env bash
# マイグレーション up → 全 down → up の可逆性を検証する。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export DATABASE_URL="${DATABASE_URL:-postgres://user:password@localhost:5432/runsync_db?sslmode=disable&TimeZone=Asia/Tokyo}"

cd "$ROOT/backend"
go test -count=1 -run TestMigrationsReversible ./database
