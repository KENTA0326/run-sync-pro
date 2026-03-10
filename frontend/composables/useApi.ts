/**
 * API 共通処理 + インターセプター
 * - ベースURL は runtimeConfig.public.apiBase
 * - リクエスト: 認証トークンを自動付与
 * - レスポンス: 401 時にトークン削除＆ログインへリダイレクト
 */
import type { ApiErrorBody } from '~/types/api'

export function useApi() {
  const config = useRuntimeConfig()
  const auth = useAuthStore()
  const baseURL = config.public.apiBase as string

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
