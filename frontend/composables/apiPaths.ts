/**
 * バックエンドの `internal/httpserver`（/api/v1）とパスを揃える。
 * 変更時は backend とここをセットで更新する。
 */
export const API_V1 = '/api/v1' as const

export const apiPath = {
  authSignup: `${API_V1}/auth/signup`,
  authLogin: `${API_V1}/auth/login`,
  shoes: `${API_V1}/shoes`,
  shoe: (id: number | string) => `${API_V1}/shoes/${id}`,
  trainingLogs: `${API_V1}/training-logs`,
  trainingLogsFormatted: `${API_V1}/training-logs/formatted`,
  trainingLogsImportStream: `${API_V1}/training-logs/import/stream`,
  analysisMonthly: `${API_V1}/analysis/monthly`,
  usersMe: `${API_V1}/users/me`,
  vdotCalculate: `${API_V1}/vdot/calculate`,
  splitsFullMarathon: `${API_V1}/splits/full-marathon`,
} as const
