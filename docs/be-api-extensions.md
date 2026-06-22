# REST / 拡張 API 補足（Run Sync Pro）

本ドキュメントは [openapi.yaml](./openapi.yaml) の補足です。新規クライアントは **`/api/v1`** を利用してください。

## URL 設計

### 名詞リソース + パスパラメータ（個体の特定）

コレクションは複数形の名詞、特定の1件は `:id` で表します。

| 操作 | メソッド | パス |
|------|----------|------|
| シューズ一覧 | GET | `/api/v1/shoes` |
| シューズ登録 | POST | `/api/v1/shoes` |
| シューズ1件取得 | GET | `/api/v1/shoes/:id` |
| シューズ削除 | DELETE | `/api/v1/shoes/:id` |
| 走行ログ一覧 | GET | `/api/v1/training-logs`（`?view=formatted` で日付等を整形） |
| 走行ログ登録 | POST | `/api/v1/training-logs` |
| 走行ログ1件取得 | GET | `/api/v1/training-logs/:id` |

### 動詞・操作名をパスに含める（分かりやすさ優先）

認証・計算・取込/export などは、パスに操作名を含めます。

| 操作 | メソッド | パス |
|------|----------|------|
| ユーザー登録 | POST | `/api/v1/auth/signup` |
| ログイン | POST | `/api/v1/auth/login` |
| パスワードリセット申請 | POST | `/api/v1/auth/password-reset/request` |
| パスワード再設定 | POST | `/api/v1/auth/password-reset/confirm` |
| VDOT 計算 | POST | `/api/v1/vdot/calculate` |
| マラソンスプリット | POST | `/api/v1/splits/full-marathon` |
| CSV エクスポート | GET | `/api/v1/training-logs` + `Accept: text/csv` |
| CSV インポート | POST | `/api/v1/training-logs` + `multipart/form-data`（`file`）または `Content-Type: text/csv` |
| JSON 一括インポート | POST | `/api/v1/training-logs` + `Content-Type: application/json`（トップレベル配列） |
| 月別レポート | GET | `/api/v1/analysis/monthly` |

レガシー別名（`/signup`, `/auth/...`, `/auth/training-logs/formatted` など）は後方互換のため残しています。  
`formatted` 専用 URL は廃止し、`GET /training-logs?view=formatted` に統合しました。  
export/import 専用パスも廃止し、同一リソース `/training-logs` の Content negotiation に統合しました（レガシー `/auth/training-logs/export/csv` 等は残存）。

### 一覧の view クエリ

| 値 | 説明 |
|----|------|
| （省略） | `model.TrainingLog` をそのまま JSON 化 |
| `formatted` | 日付 `YYYY-MM-DD`、タイムスタンプ整形、シューズ要約 |

## バージョニング

- 現行: `/api/v1/...`
- レガシー: `/login`, `/auth/...` 等（後方互換）

## ページネーション

一覧 API（`GET /shoes`, `GET /training-logs` など）:

| クエリ | 説明 | デフォルト |
|--------|------|------------|
| `limit` | 1ページ件数（最大 500） | 50 |
| `offset` | 先頭からの件数 | 0 |

レスポンス例（走行ログ）:

```json
{
  "items": [ ... ],
  "pagination": { "limit": 50, "offset": 0, "total": 261 }
}
```

## 統一エラーレスポンス

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "この操作を実行する権限がありません"
  }
}
```

主な `code`: `INVALID_INPUT`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `INTERNAL_ERROR`

## GraphQL

- `POST /api/v1/graphql`（JWT 必須）
- Body: `{ "query": "{ shoes(limit:10) { id brand } }" }`

## WebSocket

- `GET /api/v1/ws/training-logs`（JWT 必須、Upgrade）
- メッセージ: `{ "type": "ping" }` → `{ "type": "pong" }`
- CSV 取込完了時: `{ "type": "training_logs_imported", "created_count": N }`

## gRPC

- 環境変数 `GRPC_PORT=9090` で標準 **Health** サービスを起動
- `grpc_health_v1.Health/Check`（HTTP/2）

```bash
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
```

## gRPC / GraphQL の位置づけ

本番の主経路は **REST + OpenAPI**。GraphQL / WebSocket / gRPC は拡張・学習用サンプルとして同梱しています。
