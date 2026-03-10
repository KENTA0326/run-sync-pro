# MR: フロントエンド「通信」機能の実装（API共通処理・インターセプター・型定義）

## 概要

プロジェクト計画シート「8行目：フロントエンド - 通信」に対応する変更です。  
API共通処理、リクエスト/レスポンスの型定義、認証トークン付与・401時のリダイレクトを行い、あわせて新規登録画面を追加しています。

---

## 変更内容

### 1. 型定義（TypeScript）

- **追加**: `frontend/types/api.ts`
  - 認証: `LoginRequest` / `LoginResponse`、`SignUpRequest` / `SignUpResponse`、`AuthMeResponse`
  - エラー: `ApiErrorBody`（バックエンドの `{"error": "..."}` 用）
  - 今後のエンドポイント追加時もここに型を追加する想定

### 2. API共通処理・インターセプター

- **追加**: `frontend/composables/useApi.ts`
  - **ベースURL**: `runtimeConfig.public.apiBase`（既定値 `http://localhost:8080`）。環境変数 `NUXT_PUBLIC_API_BASE` で上書き可能
  - **メソッド**: `get` / `post` / `put` / `delete` を提供。呼び出し側は `api.get<T>(path)` のように型パラメータでレスポンス型を指定
  - **リクエスト**: 認証トークン（localStorage または store）があれば `Authorization: Bearer <token>` を自動付与
  - **レスポンス**: 401 時にトークン削除 ＋ `/login` へリダイレクト
  - **エラー表示**: `getErrorMessage(err)` でバックエンドの `error` メッセージを取得

### 3. 認証ストアの拡張

- **変更**: `frontend/stores/auth.ts`
  - getter `tokenOrStorage`: ストア ＋ localStorage の両方からトークン取得（リロード後も利用可能）
  - action `clearToken()`: ログアウト・401 時用

### 4. 設定

- **変更**: `frontend/nuxt.config.ts`
  - `runtimeConfig.public.apiBase` を追加（既定値 `http://localhost:8080`）
  - `process.env` は使用せず、Nuxt の環境変数マージに任せる形に変更（型エラー解消）

### 5. 画面

- **変更**: `frontend/pages/login.vue`
  - 生の `$fetch` をやめ、`useApi().post<LoginResponse>('/login', body)` に変更
  - エラー表示を `api.getErrorMessage(err)` に統一
  - 「**新規登録はこちら**」リンクを追加（`/signup` へ）

- **追加**: `frontend/pages/signup.vue`
  - 新規登録画面（名前・メール・パスワード）
  - `useApi().post<SignUpResponse>('/signup', body)` で登録
  - 「**ログインはこちら**」リンクで `/login` へ

- **追加**: `frontend/pages/index.vue`
  - トップページ。「**認証情報を取得（/auth/me）**」ボタンで認証付きAPIの動作確認が可能

### 6. 検証・ドキュメント

- **追加**: `frontend/VERIFICATION.md`
  - 通信機能の検証手順（ログイン成功/失敗、認証付きAPI、401時のリダイレクト）を記載

### 7. 型・ビルド周りの修正

- `useApi.ts`: `body` の型を `unknown` から `Record<string, any> | string | null` に変更し、`$fetch` の型と整合（TS エラー解消）

---

## 対応するAPI（バックエンド）

| メソッド | パス | 用途 |
|----------|------|------|
| POST | `/signup` | 新規登録 |
| POST | `/login` | ログイン（トークン取得） |
| GET | `/auth/me` | ログイン中ユーザー情報取得（要トークン） |

---

## 動作確認のポイント

1. **ログイン** → 成功でトップへ遷移、Local Storage に `auth_token` が保存されること
2. **ログイン失敗** → バックエンドのエラーメッセージが表示されること
3. **新規登録** → 登録後「ログインはこちら」でログインできること
4. **認証付きAPI** → トップの「認証情報を取得」で `user_id` / `message` が表示され、Network で `Authorization: Bearer ...` が付与されていること
5. **401時** → 無効トークンで上記ボタン押下時、トークン削除のうえ `/login` にリダイレクトされること

詳細は `frontend/VERIFICATION.md` を参照。

---

## 今後の利用例

- 認証不要: `api.post<SignUpResponse>('/signup', body)`
- 認証必要: `api.get<AuthMeResponse>('/auth/me')`（トークンは自動付与）
- エラー表示: `catch` 内で `api.getErrorMessage(err)` を使用

新規エンドポイント追加時は `types/api.ts` に型を追加し、画面から `useApi()` で呼び出す流れを推奨。
