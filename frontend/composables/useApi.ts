/**
 * API 共通処理 + インターセプター
 * - ベースURL は runtimeConfig.public.apiBase
 * - リクエスト: 認証トークンを自動付与
 * - レスポンス: 401 時にトークン削除＆ログインへリダイレクト
 */
import type { ApiErrorBody } from '~/types/api'

function resolveApiBase(raw: unknown): string {
  const s = typeof raw === 'string' ? raw.trim() : ''
  // 空だと `/api/v1/...` が「今のページのオリジン」向きになり、Nuxt が 404 を返す。
  if (s) return s.replace(/\/+$/, '')
  return 'http://localhost:8080'
}

export function useApi() {
  const config = useRuntimeConfig()
  const auth = useAuthStore()
  const baseURL = resolveApiBase(config.public.apiBase)

  type RequestOptions = {
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
    body?: Record<string, any> | string | null
    query?: Record<string, string | number | boolean>
  }

  async function request<T>(
    path: string,
    options: RequestOptions = {}
  ): Promise<T> {
    const token = auth.tokenOrStorage
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }

    const url = path.startsWith('http') ? path : `${baseURL}${path}`

    return $fetch<T>(url, {
      method: options.method ?? 'GET',
      body: options.body,
      query: options.query,
      headers,
      onResponseError({ response }) {
        if (response.status === 401) {
          auth.clearToken()
          if (import.meta.client) navigateTo('/login')
        }
      },
    })
  }

  return {
    get: <T>(path: string, query?: RequestOptions['query']) =>
      request<T>(path, { method: 'GET', query }),
    post: <T>(path: string, body?: Record<string, any> | string | null) =>
      request<T>(path, { method: 'POST', body }),
    put: <T>(path: string, body?: Record<string, any> | string | null) =>
      request<T>(path, { method: 'PUT', body }),
    delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
    /** エラー時の message を BE の { error: string } から取得 */
    getErrorMessage(err: unknown): string {
      const body = (err as { data?: ApiErrorBody })?.data
      return body?.error ?? 'リクエストに失敗しました'
    },
  }
}
