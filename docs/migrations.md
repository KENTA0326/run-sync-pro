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

---

## 本番運用方針（ロック時間・段階的変更・可逆性）

「動く」だけでなく、**本番で安全に流せる**ことを意識した方針です。スライドや面接でそのまま使えるよう、要点をまとめます。

### 1. 1 マイグレーション = 小さく可逆

- **「カラム追加 + バックフィル + NOT NULL 化」を 1 ファイルでやらない**。複数の番号に分割する。
- **`down` は基本「直前リリースに戻すための逆操作」**だけを書き、**データ破壊的な down は書かない**（戻すときは新しい補完 migration を上げる）。
- すべて **`IF EXISTS` / `IF NOT EXISTS`** を活用し、**冪等**に書く。

### 2. ロック時間を抑える（PostgreSQL）

長時間の `ALTER` は **`AccessExclusiveLock`** を取り、既存接続をブロックする。マイグレーションの先頭で **タイムアウトを明示**しておくと、想定外に長引いたとき安全に失敗できる。

```sql
-- up.sql の先頭で実行（migrate がファイル単位でトランザクションに包む）
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '60s';
```

- **`lock_timeout`**: ロック取得に 5 秒以上かかったらエラーで止める（本番のリクエストを長く待たせない）。
- **`statement_timeout`**: 個別 SQL に上限を付ける（暴走防止）。

### 3. オンラインで安全な DDL を選ぶ

- **インデックス追加**は **`CREATE INDEX CONCURRENTLY`** を使う（書き込みをブロックしない）。  
  ただし **トランザクション外**で実行する必要があるため、**migrate のトランザクション分離オプションが必要**、または **インデックス専用のマイグレーション**として分けて運用する。
- **`ALTER TABLE` を一発でやらない**。型変換・`NOT NULL` 化はテーブルスキャンを伴うため、**バックフィルを別ステップに分ける**。

### 4. 段階的変更パターン

#### 4.1 カラム追加（最も簡単）

1. **`ADD COLUMN xxx NULL`** で追加（DEFAULT 値があるなら **PG 11+ は即時で安全**）。
2. 必要ならアプリを **書き込み対応**にデプロイ。

#### 4.2 カラム名の変更（旧 `a` → 新 `b`）

1. **新カラム `b` を NULL 許可で追加**。
2. アプリを **`a` と `b` 両方に書く**ようデプロイ（二重書き）。
3. バッチで **`b ← a` を埋める**（小さなバッチに分割。例: `WHERE id BETWEEN ? AND ?`）。
4. アプリを **`b` を読む**ようにデプロイ。
5. **次のリリース**で `a` を削除。

#### 4.3 後付けの `NOT NULL`

- いきなり `ALTER COLUMN ... SET NOT NULL` はテーブルスキャンでロック。
- **`ADD COLUMN ... DEFAULT ... NOT NULL`** を使うか、次の段階に分ける:
  1. `NULL` 許可で追加。
  2. バックフィルで全行に値を入れる。
  3. **`SET NOT NULL`**（行が全部埋まっていれば速い）。

#### 4.4 テーブルリネーム・大規模変更

- 一気にやらない。**新テーブル作成 → 二重書き or ビュー → 読み替え → 旧テーブル削除**、と段階を踏む。

### 5. デプロイ前チェック

- ローカル / ステージングで `migrate up` → `migrate down 1` → `migrate up` を回し **可逆性**を確認。
- 重い `ALTER` は `EXPLAIN`・`pg_locks`・`pg_stat_activity` で **影響範囲とロック**を確認。
- 本番では **メンテナンスウィンドウ**または **書き込みが少ない時間帯**を選ぶ。

### 6. このリポジトリの現状

- **`000001`〜`000003`**: 初期スキーマ・RBAC・パスワードリセット。各 `up` 先頭に `SET LOCAL lock_timeout` / `statement_timeout`。
- **`000004_add_created_by`**: 段階1のみ（`ADD COLUMN` + `UPDATE` バックフィル）。索引は含めない。
- **`000005` / `000006`**: `created_by` 索引を **`CREATE INDEX CONCURRENTLY`** で **1 索引 = 1 ファイル**（トランザクション外実行のため）。
- **`down`**: 索引削除（`DROP INDEX CONCURRENTLY`）→ カラム削除の順で可逆。
- **CI**: `TestMigrationsReversible` で **up → 全 down → up** を PostgreSQL 上で検証（`.github/workflows/go-ci.yml`）。
- **ローカル**: `scripts/test-migrations.sh`（要 `DATABASE_URL` と起動中の Postgres）。
- **アプリ起動時に `migrate up` が走る**ので、デプロイ手順は **「コード反映 → 再起動」**でスキーマが追従する。
- 既存テーブル変更が必要になったら、**上の段階的変更パターンに沿って 1 つずつ migration を追加**する運用にする。

