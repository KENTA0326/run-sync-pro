# OpenAPI（Swagger）仕様書の見方

API の仕様は [openapi.yaml](openapi.yaml) にあります。以下のいずれかで閲覧・試行できます。

## 1. Swagger UI（おすすめ・Try it out 付き）

プロジェクトルートで Swagger UI を起動します。

```bash
docker compose up swagger-ui -d
```

ブラウザで **http://localhost:8081** を開いてください。

- **Try it out** … API をその場で実行
- **Authorize** … `Bearer <JWT>` を設定（先にログインで token を取得）

API 呼び出し先は `openapi.yaml` の `servers`（既定: `http://localhost:8080`）です。Try it out する前にバックエンドを起動しておいてください。

```bash
docker compose up backend -d   # またはローカルで go run
```

停止する場合:

```bash
docker compose stop swagger-ui
```

## 2. Swagger Editor（オンライン）

1. [Swagger Editor](https://editor.swagger.io/) を開く
2. `File` → `Import file` で `docs/openapi.yaml` をアップロード  
   または YAML の内容をコピー＆ペースト

## 3. VS Code 拡張

- **OpenAPI (Swagger) Editor** や **Swagger Viewer** を入れると、YAML を開いた状態でプレビューや「Try it out」が使えます。

## 4. Redocly CLI（静的 HTML を生成）

`preview-docs` は廃止されています。代わりに `build-docs` で HTML を出力し、ブラウザで開きます。

```bash
# プロジェクトルートで
npx @redocly/cli build-docs docs/openapi.yaml -o docs/api-docs.html
open docs/api-docs.html   # macOS
```

読み取り専用のドキュメント表示向けです（Swagger UI のような「Try it out」はありません）。

## 5. Swagger UI を npx で起動（Try it out 付き）

```bash
npx -y swagger-ui-wizard docs/openapi.yaml
```

対話形式でポート等を聞かれる場合があります。確実なのは上記「1. Swagger UI」です。

---

認証付きエンドポイント（`/auth/*`）を試す場合は、先に `POST /login` で `token` を取得し、Swagger UI の「Authorize」に `Bearer <token>` を入力してください。
