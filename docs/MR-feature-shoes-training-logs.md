# MR: シューズ管理・走行ログ管理の実装

## 概要

ユーザーごとの **シューズ管理（CRUD）** と **走行ログ管理** を実装しました。走行ログ登録時には、使用したシューズの累計距離を **トランザクションで更新** します。フロントには `/shoes` と `/training-logs` の画面を追加し、登録と一覧表示ができるようにしています。

---

## バックエンド（Go）

### モデル

- **`model/Shoe`**
  - `TotalDistance float64`：累計走行距離
  - `IsActive bool`：論理削除用フラグ（まだ履いているか）
- **`model/TrainingLog`**
  - `ShoeID uint`：使用シューズID
  - `Shoe Shoe`：`Preload("Shoe")` 用の関連フィールド（シューズ情報を一緒に返すため）

### シューズ管理 API

**ファイル**: `backend/handler/shoe_training_handler.go`

- **POST `/auth/shoes` → `CreateShoe`**
  - リクエスト: `{ brand, model, purchase_date(YYYY-MM-DD) }`
  - 認証ミドルウェアから `userID` を取得し、`Shoe{UserID, Brand, Model, PurchaseDate}` を作成。
- **GET `/auth/shoes` → `ListShoes`**
  - `user_id = userID` かつ `is_active = true` のシューズ一覧を取得。
  - 購入日の新しい順（`purchase_date DESC, id DESC`）で返却。
- **DELETE `/auth/shoes/:id` → `DeleteShoe`**
  - 対象ユーザーのシューズを1件取得し、`is_active = false` に更新（**論理削除**）。
  - すでに非アクティブな場合は 200 で「すでに削除済み」と返却。

### 走行ログ管理 API

**入力構造体**

```go
type createTrainingLogInput struct {
  TrainingDate string  `json:"training_date" binding:"required"` // YYYY-MM-DD
  Distance     float64 `json:"distance" binding:"required,gt=0"`
  Duration     int     `json:"duration" binding:"required,gt=0"` // 秒
  Pace         string  `json:"pace" binding:"required"`
  Memo         string  `json:"memo"`
  Kind         int     `json:"kind" binding:"gte=0,lte=3"`       // 0:ジョグ,1:LSD,2:ペース走,3:インターバル
  ShoeID       uint    `json:"shoe_id" binding:"required"`       // 使用シューズ
}
```

- **POST `/auth/training-logs` → `CreateTrainingLog`**
  - `db.Transaction(func(tx *gorm.DB) error { ... })` で実装。
  - トランザクション内の処理:
    1. `Shoe(id, user_id)` を取得し、本人のシューズかチェック。見つからない場合は `gin.Error` で「シューズが見つかりません」を返す。
    2. `TrainingLog` を作成。
    3. `shoe.total_distance` に `distance` を加算  
       `Update("total_distance", gorm.Expr("total_distance + ?", input.Distance))`
  - バリデーションエラー時は英語の生メッセージではなく、**日本語**で返却:
    - `{"error": "入力内容が正しくありません。日付・距離・時間・種類・シューズを確認してください。"}`
- **GET `/auth/training-logs` → `ListTrainingLogs`**
  - `Where("user_id = ?", userID)` ＋ `Preload("Shoe")` ＋ `Order("training_date DESC, id DESC")`
  - ログと、そのとき使ったシューズ情報をまとめて返却。

### ルーティング

**`backend/main.go`**

```go
authGroup := r.Group("/auth")
authGroup.Use(middleware.AuthMiddleware()) {
  authGroup.GET("/me", ...)

  // シューズ管理
  authGroup.POST("/shoes", handler.CreateShoe)
  authGroup.GET("/shoes", handler.ListShoes)
  authGroup.DELETE("/shoes/:id", handler.DeleteShoe)

  // 走行ログ管理
  authGroup.POST("/training-logs", handler.CreateTrainingLog)
  authGroup.GET("/training-logs", handler.ListTrainingLogs)
}
```

---

## フロントエンド（Nuxt 3）

### 型定義

**`frontend/types/api.ts`**

- **シューズ**

```ts
export interface Shoe {
  id: number
  user_id: number
  brand: string
  model: string
  purchase_date: string
  total_distance: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateShoeRequest {
  brand: string
  model: string
  purchase_date: string
}
```

- **走行ログ**

```ts
export interface TrainingLog {
  id: number
  user_id: number
  training_date: string
  distance: number
  duration: number
  pace: string
  memo: string
  kind: number
  shoe_id: number
  shoe: Shoe
  created_at: string
  updated_at: string
}

export interface CreateTrainingLogRequest {
  training_date: string
  distance: number
  duration: number
  pace: string
  memo: string
  kind: number
  shoe_id: number
}
```

### シューズ管理ページ `/shoes`

**ファイル**: `frontend/pages/shoes.vue`

- **機能**
  - フォームで「ブランド / モデル / 購入日」を入力 → `POST /auth/shoes` で登録。
  - 登録済みシューズを `GET /auth/shoes` で取得し、一覧表示。
  - 削除ボタンから `DELETE /auth/shoes/:id` を呼び、一覧を再読込。
- **UI**
  - 上部: 登録フォーム（入力エラー時は「ブランド・モデル・購入日を入力してください。」を表示）。
  - 下部: テーブル（ブランド/モデル・購入日・累計距離・削除ボタン）。
- **ガード**
  - `definePageMeta({ middleware: 'auth' })` でログイン必須。

### 走行ログページ `/training-logs`

**ファイル**: `frontend/pages/training-logs.vue`

- **登録フォーム**
  - 日付（`type="date"`）
  - 距離(km)
  - 時間（分・秒入力 → 合計秒に変換）
  - ペース文字列（例: `5:15`）
  - 種類（セレクト: ジョグ / LSD / ペース走 / インターバル → `kind: 0〜3`）
  - メモ
  - シューズ選択（`GET /auth/shoes` で取得した一覧から選ぶ）
- **保存処理**
  - `CreateTrainingLogRequest` を組み立てて `POST /auth/training-logs`
  - 成功時:
    - 距離・時間・ペース・メモ・種類をリセット（シューズ選択・日付は維持）
    - `GET /auth/training-logs` でログ一覧を再取得
    - `GET /auth/shoes` も再取得（シューズの `total_distance` 更新を反映）
  - バリデーションエラーなどは `api.getErrorMessage(err)` で表示（バックエンドの日本語メッセージがそのまま出る）。
- **一覧表示**
  - テーブル形式で 日付 / 距離 / 時間 / ペース / 種類 / シューズ を表示。
  - 時間は `h:mm:ss` / `m:ss` 形式に整形。
  - 種類は `0〜3` を `ジョグ / LSD / ペース走 / インターバル` に変換。

### ナビゲーション

**`frontend/layouts/default.vue`**

- ハンバーガーメニュー内のリンクを追加:
  - **トップ**: `/`
  - **シューズ**: `/shoes`
  - **走行ログ**: `/training-logs`
  - **VDOT**: `/vdot`

---

## チェックポイント

- `CreateTrainingLog` が **`db.Transaction`** で実装されている。
- `userID` は認証ミドルウェアから `c.Get("userID")` 経由で取得している。
- 走行ログ保存時に **`total_distance = total_distance + distance`** が正しく実行される。
- 走行ログ取得時に **`Preload("Shoe")`** でシューズ情報が一緒に返却される。
- バリデーションエラー時に、フロントには **日本語メッセージ** が表示される。
