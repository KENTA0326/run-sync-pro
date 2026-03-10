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

// --- VDOT ---
export interface VDOTCalculateRequest {
  distance_meters: number
  time_seconds: number
  riegel_exponent?: number
}

export interface VDOTPaces {
  easy_min_sec_per_km: number
  easy_max_sec_per_km: number
  marathon_sec_per_km: number
  threshold_sec_per_km: number
  interval_sec_per_km: number
  repetition_sec_per_km: number
}

export interface VDOTCalculateResponse {
  vdot: number
  paces: VDOTPaces
  riegel_exponent: number
  riegel_predictions: {
    full_seconds: number
    half_seconds: number
    ten_k_seconds: number
    five_k_seconds: number
  }
}

// --- シューズ ---
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

// --- 走行ログ ---
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

// --- Splits ---
export interface FullMarathonSplitsRequest {
  pace_sec_per_km: number
  page?: number
}

export interface SplitRow {
  km: number
  label: string
  cumulative_seconds: number
}

export interface FullMarathonSplitsResponse {
  page: number
  total_pages: number
  rows: SplitRow[]
}

// --- エラー（BE の gin.H{"error": "..."} に合わせる）---
export interface ApiErrorBody {
  error: string
}
