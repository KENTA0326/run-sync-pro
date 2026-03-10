# MR: 走行解析（Goroutine 並列集計）・VDOT 推移グラフ・8GB Mac 向け Docker 対応

## 概要

- **走行ログの解析 API** を実装しました。月ごとの走行距離・平均ペース・**VDOT（平均/最高）** を **Goroutine と Channel** で並列集計し、将来的な大量データにも耐えうる設計にしています。
- フロントに **解析ページ**（`/analysis`）を追加し、月別レポート・**VDOT 推移グラフ**（Chart.js）・VDOT 推移テーブルを表示します。
- **8GB Mac** 等で Docker 内の Go ビルドが OOM になる問題に対し、**Mac 上でバイナリをビルドし Docker ではそのバイナリだけを実行する**方式（`docker-compose.binary.yml`）と、メモリ節約用の環境変数（Dockerfile）を追加しました。

---

## バックエンド（Go）

### 解析サービス（Goroutine + Channel）

**ファイル**: `backend/service/analysis.go`

- **AnalyzeByMonth(logs []model.TrainingLog) AnalysisResponse**
  - 走行ログを **月（year-month）ごとにグループ化** し、各月の集計を **Goroutine** で並列実行。
  - 各 Goroutine で計算: 合計距離・合計時間・走行回数・平均ペース（秒/km）、および **VDOT**（`service.CalculateVDOT(距離m, 時間秒)` で各走行から算出し、月内の **平均 VDOT** と **最高 VDOT** を算出）。
  - 結果を **Channel** で回収し、月順にソートして `AnalysisResponse` を構築。
- **返却構造体**
  - `MonthlyReport`: `year_month`, `total_distance`, `total_duration`, `run_count`, `avg_pace_sec_per_km`, `avg_vdot`, `max_vdot`
  - `AnalysisResponse`: `monthly_reports`, `total_distance`, `total_duration`, `total_run_count`

### 解析 API

**ファイル**: `backend/handler/analysis_handler.go`

- **GET `/auth/analysis` → `MonthlyReport`**
  - 認証ミドルウェアから `userID` を取得。
  - `database.DB.Where("user_id = ?", userID).Order("training_date ASC").Find(&logs)` で走行ログを取得。
  - `service.AnalyzeByMonth(logs)` を呼び出し、結果を JSON で返却。

### ルーティング

**`backend/main.go`**

- 認証グループ内に追加: `authGroup.GET("/analysis", handler.MonthlyReport)`

### メモリ節約（Docker 内ビルド用）

**`backend/Dockerfile`**

- `ENV GOGC=50` / `ENV GOMAXPROCS=1` を追加（8GB Mac 等でビルド OOM を軽減するため）。

**`backend/.air.toml`**

- 既存: `cmd = "go build -p 1 -o ./tmp/main ."`（パッケージを 1 つずつビルド）。

---

## フロントエンド（Nuxt 3）

### 型定義

**`frontend/types/api.ts`**

- **解析**
  - `MonthlyReport`: `year_month`, `total_distance`, `total_duration`, `run_count`, `avg_pace_sec_per_km`, `avg_vdot`, `max_vdot`
  - `AnalysisResponse`: `monthly_reports`, `total_distance`, `total_duration`, `total_run_count`

### 解析ページ `/analysis`

**ファイル**: `frontend/pages/analysis.vue`

- **ガード**: `definePageMeta({ middleware: 'auth' })`
- **取得**: `GET /auth/analysis` で解析結果を取得。
- **表示**
  - サマリーカード: 全期間の総走行距離・走行回数・平均ペース。
  - **VDOT 推移**: Chart.js（`vue-chartjs`）による折れ線グラフ（最高 VDOT / 平均 VDOT）。`ClientOnly` でラップ。
  - **VDOT 推移データ**: VDOT が有効な月のみのテーブル（月・距離・走行回数・平均ペース・平均 VDOT・最高 VDOT）。
  - **月別レポート**: 全月のテーブル（上記に加え平均 VDOT・最高 VDOT 列あり）。VDOT が無い月は「-」表示。

### VDOT 推移グラフコンポーネント

**ファイル**: `frontend/components/VdotTrendChart.vue`

- Chart.js の `Line` を使用。X 軸: 月（`y/m`）、Y 軸: VDOT。
- データセット: 最高 VDOT（青）、平均 VDOT（緑）。
- `chart.js` / `vue-chartjs` を dependencies に追加。`nuxt.config.ts` の `vite.optimizeDeps.include` に `chart.js`, `vue-chartjs` を指定。

### ナビゲーション

**`frontend/layouts/default.vue`**

- メニューに **解析**（`/analysis`）リンクを追加。

---

## 8GB Mac 向け Docker 対応（バイナリビルド方式）

Docker 内で `go build` すると `github.com/ugorji/go/codec` のコンパイルで OOM になるため、**ビルドは Mac 上で行い、コンテナにはバイナリだけを渡す**オプションを追加しました。

### 追加ファイル

| ファイル | 説明 |
|----------|------|
| **`scripts/build-backend.sh`** | Mac 上で `GOOS=linux GOARCH=arm64`（または amd64）のバイナリをビルド。`GOGC=50` `GOMAXPROCS=1` `-p 1` でメモリ節約。 |
| **`backend/Dockerfile.binary`** | Alpine ベース。事前にビルドした `main` を `COPY` して実行するだけ（コンテナ内で `go build` しない）。 |
| **`docker-compose.binary.yml`** | `backend` の `build.dockerfile` を `Dockerfile.binary` にし、`volumes` を空にしてバイナリのみで動作させるオーバーライド。 |
| **`docs/docker-memory.md`** | OOM 時の対処（Mac でビルドする手順・Docker のメモリ設定の目安）を記載。 |

### 使い方（8GB Mac 等）

1. Mac に Go をインストール（`go version` で確認）。
2. `./scripts/build-backend.sh` で `backend/main` を生成。
3. `docker-compose -f docker-compose.yml -f docker-compose.binary.yml build backend`
4. `docker-compose -f docker-compose.yml -f docker-compose.binary.yml up -d db backend frontend`

コード変更時は 2 → 3 → 4 をやり直す（ホットリロードはなし）。再起動のみの場合は 4 の `up -d` だけでよい。

### その他

- **`.gitignore`**: `backend/main` を追加（ビルド成果物をコミットしない）。

---

## 対応 API 一覧（本 MR 分）

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/auth/analysis` | 月別走行レポート・VDOT（平均/最高）を Goroutine 並列集計で返す |

---

## チェックポイント

- 解析処理が **Goroutine** と **Channel** で月ごとに並列集計されている。
- 各月の **VDOT** は走行ログの距離・時間から `service.CalculateVDOT` で算出し、平均・最高を返している。
- 解析 API は認証必須（`/auth` グループ）で、`userID` でフィルタした走行ログのみを集計している。
- フロントの VDOT 推移グラフは Chart.js で表示し、VDOT が有効な月のみグラフ・VDOT テーブルに表示している。
- 8GB Mac では `docker-compose.binary.yml` と `scripts/build-backend.sh` で、Docker 内ビルドなしでバックエンドを起動できる。

---

## 動作確認のポイント

1. ログイン後、メニューから「解析」を開き、月別レポート・VDOT 推移グラフ・VDOT 推移データが表示されること。
2. 走行ログが無い場合は「走行ログがまだありません」等のメッセージが出ること。
3. VDOT が算出できる走行がある月は、グラフとテーブルに平均 VDOT・最高 VDOT が表示されること。
4. 8GB Mac で `./scripts/build-backend.sh` 実行後、`docker-compose -f docker-compose.yml -f docker-compose.binary.yml up -d ...` で DB / backend / frontend が起動すること。
