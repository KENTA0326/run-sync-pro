# マイグレーション（golang-migrate）の使い方

## 概要

スキーマ変更を **バージョン管理された SQL マイグレーション** で行うようにしています。  
`backend/database/Connect()` 実行時に、未適用のマイグレーションが自動で `Up` されます。

- **ツール**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **場所**: `backend/database/migrations/`（`000001_名前.up.sql` / `000001_名前.down.sql`）
- **適用タイミング**: アプリ起動時（`main.go` → `database.Connect()`）

---

## 日常の使い方

### アプリ起動で自動適用

```bash
cd backend && go run .
# または Docker で起動
docker compose up -d
```

起動時に未適用のマイグレーションがあれば自動で `migrate up` が実行されます。  
「マイグレーション忘れ」で本番とスキーマがずれることを防げます。

### 接続先の切り替え（ローカルなど）

デフォルトは Docker Compose 用の接続先（`host=db`）です。  
ローカルで PostgreSQL を動かしている場合は、**URL 形式**で指定します。

```bash
export DATABASE_URL="postgres://user:password@localhost:5432/runsync_db?sslmode=disable&TimeZone=Asia/Tokyo"
cd backend && go run .
```

---

## 新規マイグレーションの追加手順

新しいテーブル追加・カラム追加などは、**新しい番号の up/down ファイル**を追加します。

### 1. migrate CLI のインストール（推奨）

```bash
# macOS (Homebrew)
brew install golang-migrate
```

### 2. マイグレーションファイルの作成

`backend/database` にいる想定で:

```bash
cd backend/database
migrate create -ext sql -dir migrations -seq add_workouts_table
```

すると次のファイルができます。

- `migrations/000002_add_workouts_table.up.sql`
- `migrations/000002_add_workouts_table.down.sql`

### 3. up / down の SQL を書く

**up（適用時）** 例:

```sql
-- 000002_add_workouts_table.up.sql
CREATE TABLE IF NOT EXISTS workouts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON workouts (user_id);
```

**down（ロールバック時）** 例:

```sql
-- 000002_add_workouts_table.down.sql
DROP TABLE IF EXISTS workouts;
```

### 4. 動作確認

- アプリを起動して `Up` が通ることを確認
- 必要ならローカルで `migrate -path migrations -database "postgres://..." down 1` で 1 段下げてから再度 `go run .` で `Up` を確認

---

## ファイル構成（イメージ）

```
backend/database/
├── database.go          # 起動時に embed で migrations を読み込み migrate.Up() を実行
└── migrations/
    ├── 000001_init_schema.up.sql
    ├── 000001_init_schema.down.sql
    ├── 000002_xxx.up.sql   # 今後の追加分
    └── 000002_xxx.down.sql
```

---

## 面接でどう説明するか（ポイント）

採用面接で「マイグレーションを整備した」と話すときの**説明の流れ**と**キーワード**です。

### 1. 何をしたか（一言）

- 「本番の DB スキーマを、**バージョン管理された SQL マイグレーション**（golang-migrate）で管理するようにしました。以前は GORM の AutoMigrate だけに頼っていました。」

### 2. なぜやったか（背景・課題）

- 「**本番で意図しないスキーマ変更**が入るリスクを避けたかったです。AutoMigrate は『コードのモデル定義＝本番の形』になるので、削除や型変更がそのまま本番に反映されてしまいます。」
- 「**誰が・いつ・どんな変更を入れたか**を Git の履歴で追えるようにしたかったです。複数人開発で『本番と自分の DB がずれた』という事態を減らすためです。」

### 3. どういう仕組みにしたか

- 「**up/down の 2 種類の SQL ファイル**で、適用とロールバックの両方を定義しています。番号で順序が決まるので、CI やデプロイで `migrate up` を実行するだけで一貫した状態にできます。」
- 「このプロジェクトでは**アプリ起動時に自動で migrate up** するようにして、『マイグレーションを忘れてデプロイする』というミスを防いでいます。マイグレーションは `embed` でバイナリに含めているので、別途 CLI を本番サーバーに置く必要もありません。」

### 4. 効果・学び

- 「**リリース手順が明確**になりました。『コードをデプロイ → アプリ起動でマイグレーションが走る』と説明できます。」
- 「**ロールバック手順**（down を実行する）も定義されているので、障害時に『ひとつ前のスキーマに戻す』という選択肢が持てます。」

### 5. 突っ込まれたとき用

- **「AutoMigrate と併用しないの？」**  
  「このプロジェクトではマイグレーションに一本化しています。スキーマの真実のソースを『SQL ファイル』に置くことで、GORM の解釈差や将来の ORM 変更に振り回されにくくするためです。」
- **「down は本当に使う？」**  
  「本番で down を実行することは少ないですが、**ローカルやステージングで『ひとつ前の状態に戻して再現する』**ときに使います。また、down を書くことで『この変更は戻せる』と設計レベルで意識できます。」

---

以上を押さえておくと、「チーム開発を支える基盤」として**なぜマイグレーションを導入したか・どう運用しているか**を簡潔に説明しやすくなります。
