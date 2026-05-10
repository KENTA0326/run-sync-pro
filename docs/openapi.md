# OpenAPI（Swagger）仕様書の見方

API の仕様は [openapi.yaml](openapi.yaml) にあります。以下のいずれかで閲覧・試行できます。

## 1. Swagger Editor（オンライン）

1. [Swagger Editor](https://editor.swagger.io/) を開く
2. `File` → `Import file` で `docs/openapi.yaml` をアップロード  
   または YAML の内容をコピー＆ペースト

## 2. VS Code 拡張

- **OpenAPI (Swagger) Editor** や **Swagger Viewer** を入れると、YAML を開いた状態でプレビューや「Try it out」が使えます。

## 3. Docker で Swagger UI を立てる

```bash
# プロジェクトルートで
docker run -p 8081:8080 -e SWAGGER_JSON=/spec/openapi.yaml -v $(pwd)/docs:/spec swaggerapi/swagger-ui
```

ブラウザで http://localhost:8081 を開くと、`docs/openapi.yaml` を読み込んだ Swagger UI が表示されます。  
認証が必要な API は「Authorize」で `Bearer <ログインで取得したtoken>` を指定してください。

## 4. npx で Redoc や Swagger UI を一度だけ起動

```bash
# Redoc（読みやすいドキュメント表示）
npx @redocly/cli preview-docs docs/openapi.yaml

# Swagger UI（試行可能）
npx swagger-ui-wizard
# または
npx -y swagger-ui-widget
```

---

認証付きエンドポイント（`/auth/*`）を試す場合は、先に `POST /login` で `token` を取得し、Swagger UI の「Authorize」に `Bearer <token>` を入力してください。
