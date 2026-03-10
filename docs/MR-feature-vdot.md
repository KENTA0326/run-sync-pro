# MR: VDOT計算・リーゲル予想・スプリット・強度解説

## 概要

プロジェクト計画の「ペース計算」まわりに対応する機能です。  
レースの距離・タイムから **VDOT** と **トレーニングペース（E/M/T/I/R）** を算出し、**リーゲル公式**でフル/ハーフ/10km/5kmの予想タイムを表示。あわせて **フルマラソンの1kmスプリット**（10km単位ページング）と **強度ごとの解説**（クリックで表示）を実装しています。

---

## 変更内容

### 1. バックエンド（Go）

#### VDOT計算・トレーニングペース
- **追加**: `backend/service/vdot.go`
  - **CalculateVDOT(distanceMeters, timeSeconds)**  
    Jack Daniels の公式で VDOT を算出（距離・タイム → VDOT）
  - **CalculateTrainingPaces(vdot)**  
    VDOT から各強度の推奨ペース（秒/km）を算出（E は範囲、M/T/I/R は単一ペース）
  - **CalculateRiegelPredictions(distanceMeters, timeSeconds, exponent)**  
    リーゲル公式 T2 = T1 × (D2/D1)^exponent でフル/ハーフ/10km/5km の予想タイム（秒）を返す
  - 指数は 1.06（超スタミナ型）/ 1.08（標準）/ 1.12（初心者・スピード型）を想定

#### VDOT API
- **追加**: `backend/handler/vdot_handler.go`
  - **POST /vdot/calculate**
  - リクエスト: `distance_meters`, `time_seconds`, `riegel_exponent`（任意、未指定時 1.08）
  - レスポンス: `vdot`, `paces`, `riegel_exponent`, `riegel_predictions`（full_seconds, half_seconds, ten_k_seconds, five_k_seconds）
  - ※ 予想タイムは「VDOT換算」は廃止し、**リーゲル換算のみ**返却

#### スプリット計算（スライス利用）
- **追加**: `backend/service/splits.go`
  - **GenerateFullMarathonSplits(paceSecPerKm)**  
    フルマラソン 42.195km について、1km〜42km と Finish の通過タイム（累積秒）を **スライス** で生成
- **追加**: `backend/handler/splits_handler.go`
  - **POST /splits/fullmarathon**
  - リクエスト: `pace_sec_per_km`, `page`（任意）
  - レスポンス: `page`, `total_pages`, `rows`（1ページあたり最大10件。10km単位のページネーション）

#### ルーティング
- **変更**: `backend/main.go`
  - `r.POST("/vdot/calculate", handler.VDOTCalculate)`
  - `r.POST("/splits/fullmarathon", handler.FullMarathonSplits)`

---

### 2. フロントエンド（Nuxt 3）

#### 型定義
- **変更**: `frontend/types/api.ts`
  - VDOT: `VDOTCalculateRequest`（`riegel_exponent` 追加）, `VDOTCalculateResponse`（`predictions` 削除、`riegel_exponent` / `riegel_predictions` 追加）
  - スプリット: `FullMarathonSplitsRequest`, `FullMarathonSplitsResponse`, `SplitRow`

#### VDOTページ
- **変更**: `frontend/pages/vdot.vue`
  - **入力**: 距離(km)、タイム(分:秒)、リーゲル公式の指数（ラジオ: 1.06 / 1.08 / 1.12）
  - **結果表示**:
    - あなたのVDOTは ○○ です
    - 予想タイム（リーゲル換算）テーブル（フル/ハーフ/10km/5km）
    - 推奨ペース（強度 E/M/T/I/R）テーブル
  - **強度クリック**:
    - 強度名をボタン風に表示（青リンクではなく、ホバーで薄い背景）
    - クリックで「強度の解説」エリアにスクロールし、該当強度の説明を表示
    - 「上の表の『強度』をクリックすると〜」の案内文は、**いずれかの強度を1回でもクリックしたら非表示**
  - **マラソン(M)選択時**: フルマラソンのスプリット（1kmごとの通過タイム）を表示。10km単位で前へ/次へページネーション（API: `/splits/fullmarathon`）

#### 強度解説データ
- **追加**: `frontend/data/training-intensities.ts`
  - E / M / T / I / R 各強度の「意味・目的」「推奨練習量の目安」「メニュー例」を日本語で定義（Jack Daniels のVDOT理論に基づく解説）

---

### 3. レイアウト・ヘッダー

- **変更**: `frontend/layouts/default.vue`
  - ロゴ「RunSync Pro」を **ヘッダー中央** に配置（absolute + translate で中央寄せ）
  - 右端の「メニュー」を **ハンバーガーアイコン（横線3本）** に変更（`aria-label="メニューを開く"` 付与）
  - ヘッダーは全幅（`w-full`）で、左右端にスペーサーとボタンを配置

---

## 対応API一覧

| メソッド | パス | 説明 |
|----------|------|------|
| POST | `/vdot/calculate` | VDOT・ペース・リーゲル予想タイムを返す |
| POST | `/splits/fullmarathon` | フルマラソン1kmスプリット（10件/ページ）を返す |

---

## 技術要素（スケジュールとの対応）

- **関数・型変換・スライス（Go）**: VDOT計算式、秒⇔分の変換、スプリット配列の生成とスライスによるページ切り出し
- **TypeScript**: リクエスト/レスポンスの型定義、フォームとAPIの連携
- **UI**: ラジオボタン、テーブル、ページネーション、クリックで解説表示・スムーズスクロール

---

## 動作確認のポイント

1. `/vdot` で距離・タイムを入力し「算出」→ VDOT・リーゲル予想・ペース表が表示されること
2. リーゲル指数を 1.06 / 1.08 / 1.12 で切り替えて再算出し、予想タイムが変わること
3. 強度（例: マラソン(M)）をクリック → 解説エリアにスクロールし、該当強度の説明が表示されること
4. マラソン(M)選択時、フルマラソンスプリットが表示され、「前へ」「次へ」で10km単位のページが切り替わること
5. 強度を1回でもクリックしたあと、「上の表の『強度』をクリックすると〜」の文章が表示されないこと
6. ヘッダーでロゴが中央、右にハンバーガーアイコンであること

---

## 今後の拡張例

- 他強度（E/T/I/R）でもスプリット表示（例: 5km の 1km スプリット）
- スプリットのエクスポート（CSV 等）
