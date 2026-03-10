/**
 * RunSync Pro - API 型定義
 * バックエンドのリクエスト/レスポンスと同期させる
 */

// --- 認証 ---
export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
}

export interface SignUpRequest {
  name: string
  email: string
  password: string
}

export interface SignUpResponse {
  message: string
}

export interface AuthMeResponse {
  user_id: number
  message: string
}

// --- エラー（BE の gin.H{"error": "..."} に合わせる）---
export interface ApiErrorBody {
  error: string
}
