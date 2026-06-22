/**
 * API パス（/api/v1）。
 * - リソース操作: 名詞 + パスパラメータ（例: GET /shoes/:id）
 * - 認証・計算など: 動詞・操作名をパスに含める
 * - 走行ログ CSV: GET/POST /training-logs + Accept / Content-Type で切替
 */
export const apiPath = {
  authSignup: '/api/v1/auth/signup',
  authLogin: '/api/v1/auth/login',
  authPasswordResetRequest: '/api/v1/auth/password-reset/request',
  authPasswordResetConfirm: '/api/v1/auth/password-reset/confirm',
  shoes: '/api/v1/shoes',
  shoe: (id: number | string) => `/api/v1/shoes/${id}`,
  trainingLogs: '/api/v1/training-logs',
  trainingLog: (id: number | string) => `/api/v1/training-logs/${id}`,
  trainingLogKind: (id: number | string) => `/api/v1/training-logs/${id}/kind`,
  analysisMonthly: '/api/v1/analysis/monthly',
  usersMe: '/api/v1/users/me',
  vdotCalculate: '/api/v1/vdot/calculate',
  splitsFullMarathon: '/api/v1/splits/full-marathon',
} as const
