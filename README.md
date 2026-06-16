# Run Sync Pro

## 初回セットアップ

```bash
# 環境変数（DB 接続情報など）を用意
cp .env.example .env
# 必要に応じて .env を編集

# 起動
docker compose up -d
```

- バックエンド: http://localhost:8080
- フロントエンド: http://localhost:3001
- DB: localhost:5432（POSTGRES_* は .env を参照）

## ビルド設定

Docker / Go / Nuxt / CI の設定は [docs/build-configuration.md](docs/build-configuration.md) にまとめています。

- ビルド設定の基本項目と設定ファイルの編集
- エントリーポイント（`main.go` / `app.vue`）と出力先（`tmp/main` / `.output/`）
- 環境変数の注入（`.env` → Compose → アプリ）
- 本番ビルドの最適化（Go ldflags、Nuxt/Vite minify）
- 開発時のホットリロード（Air）と HMR（Nuxt dev）

## 環境変数の外部化

接続先やパスワードは `.env` で管理しています（`.env` は Git に含めません）。

- **db**: `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` を `.env` から読み込み
- **backend**: `DATABASE_URL` を `.env` の POSTGRES_* から自動組み立て（または .env で `DATABASE_URL` を直接指定可能）

詳細は [.env.example](.env.example) を参照してください。

## API 仕様書（OpenAPI / Swagger）

- 仕様ファイル: [docs/openapi.yaml](docs/openapi.yaml)
- 見方: [docs/openapi.md](docs/openapi.md)（Swagger UI の起動方法など）

## プライベートモジュールと SemVer 運用

- 手順: [docs/private-module-semver.md](docs/private-module-semver.md)
- SemVer チェック: `make semver-check`
- タグ作成: `make release-tag`

## Go（フォーマッター/リンター）

- **フォーマット**: `gofmt`
- **リンター**: `golangci-lint`（設定: `backend/.golangci.yml`）
- **ローカルで実行**:

```bash
# golangci-lint を入れていない場合（例）
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# フォーマット / テスト / リント
make go-fmt
make go-test
make go-lint
```

- **CI**: `.github/workflows/go-ci.yml` で `go test` + `golangci-lint` を実行します（backend 変更時のみ）。

## pre-commit フック（コミット時に自動 format/lint）

コミット前に **自動で gofmt と golangci-lint** を実行して、レビュー前の品質を揃えます。設定は `lefthook.yml` にあります。

### セットアップ

```bash
# macOS（Homebrew）
brew install lefthook

# hooks を有効化（初回だけ）
lefthook install
```

以降、`git commit` 時に次が走ります。

- `backend/**/*.go` に対して `gofmt`（フォーマット結果は自動で stage し直し）
- `golangci-lint run`（失敗するとコミットをブロック）

## ルールのカスタマイズ（プロジェクト要件に合わせる）

### 1. golangci-lint のルール

`backend/.golangci.yml` の `linters.enable` を増減させることで、プロジェクトに合わせて強さを調整できます。

- **例**: もっと厳密にしたい → `gosec`, `gocritic` などを追加
- **例**: ノイズが多い → 一部 linters を外す / `issues.exclude-rules` で特定パターンを除外

### 2. pre-commit で走らせる内容

`lefthook.yml` の `pre-commit.commands` を追加・編集するだけで、コミット時のチェックを増やせます。
