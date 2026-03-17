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

## 環境変数の外部化

接続先やパスワードは `.env` で管理しています（`.env` は Git に含めません）。

- **db**: `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` を `.env` から読み込み
- **backend**: `DATABASE_URL` を `.env` の POSTGRES_* から自動組み立て（または .env で `DATABASE_URL` を直接指定可能）

詳細は [.env.example](.env.example) を参照してください。
