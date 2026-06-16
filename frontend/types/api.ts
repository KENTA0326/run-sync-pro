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

export interface PasswordResetRequestBody {
  email: string
}

export interface PasswordResetRequestResponse {
  message: string
  reset_url?: string
}

export interface PasswordResetConfirmBody {
  token: string
  new_password: string
}

export interface PasswordResetConfirmResponse {
  message: string
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

export interface PaginationMeta {
  limit: number
  offset: number
  total: number
}

export interface PaginatedTrainingLogsResponse {
  items: TrainingLog[]
  pagination: PaginationMeta
}

export interface ListShoesResponse {
  shoes: Shoe[]
  brands: string[]
  pagination: PaginationMeta
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

export interface ImportTrainingLogsResponse {
  message: string
  created_count: number
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

// --- 解析（月別レポート・Goroutine並列集計）---
export interface MonthlyReport {
  year_month: string
  total_distance: number
  total_duration: number
  run_count: number
  avg_pace_sec_per_km: number
  avg_vdot: number
  max_vdot: number
}

export interface AnalysisResponse {
  monthly_reports: MonthlyReport[]
  total_distance: number
  total_duration: number
  total_run_count: number
}

// --- エラー（BE の { error: { code, message } } に合わせる）---
export interface ApiErrorDetail {
  code: string
  message: string
}

export interface ApiErrorBody {
  error: ApiErrorDetail | string
}
