/**
 * API 共通処理 + インターセプター
 * - ベースURL は runtimeConfig.public.apiBase
 * - リクエスト: 認証トークンを自動付与
 * - レスポンス: 401 時にトークン削除＆ログインへリダイレクト
 */
import type { ApiErrorBody } from '~/types/api'

function parseApiErrorMessage(body: ApiErrorBody | undefined): string | undefined {
  if (!body?.error) return undefined
  if (typeof body.error === 'string') return body.error
  return body.error.message
}

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

  async function postFormData<T>(path: string, formData: FormData): Promise<T> {
    const token = auth.tokenOrStorage
    const headers: Record<string, string> = {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
    const url = path.startsWith('http') ? path : `${baseURL}${path}`
    return $fetch<T>(url, {
      method: 'POST',
      body: formData,
      headers,
      onResponseError({ response }) {
        if (response.status === 401) {
          auth.clearToken()
          if (import.meta.client) navigateTo('/login')
        }
      },
    })
  }

  /** GET で CSV 等のバイナリを取得し、ブラウザから保存する */
  async function downloadGet(
    path: string,
    fallbackFilename: string,
    extraHeaders?: Record<string, string>
  ): Promise<void> {
    const token = auth.tokenOrStorage
    const url = path.startsWith('http') ? path : `${baseURL}${path}`
    const res = await fetch(url, {
      method: 'GET',
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...extraHeaders,
      },
    })
    if (res.status === 401) {
      auth.clearToken()
      if (import.meta.client) navigateTo('/login')
      throw new Error('認証が必要です')
    }
    if (!res.ok) {
      let msg = 'ダウンロードに失敗しました'
      try {
        const data = (await res.json()) as ApiErrorBody
        msg = parseApiErrorMessage(data) ?? msg
      } catch {
        /* 本文が JSON でない */
      }
      throw new Error(msg)
    }
    const blob = await res.blob()
    let filename = fallbackFilename
    const cd = res.headers.get('Content-Disposition')
    const match = cd?.match(/filename="?([^";\n]+)"?/)
    if (match?.[1]) filename = match[1]
    const objectUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = objectUrl
    a.download = filename
    a.click()
    URL.revokeObjectURL(objectUrl)
  }

  return {
    get: <T>(path: string, query?: RequestOptions['query']) =>
      request<T>(path, { method: 'GET', query }),
    downloadGet,
    post: <T>(path: string, body?: Record<string, any> | string | null) =>
      request<T>(path, { method: 'POST', body }),
    postFormData: <T>(path: string, formData: FormData) => postFormData<T>(path, formData),
    put: <T>(path: string, body?: Record<string, any> | string | null) =>
      request<T>(path, { method: 'PUT', body }),
    patch: <T>(path: string, body?: Record<string, any> | string | null) =>
      request<T>(path, { method: 'PATCH', body }),
    delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
    /** エラー時の message を BE の { error: { code, message } } から取得 */
    getErrorMessage(err: unknown): string {
      const body = (err as { data?: ApiErrorBody })?.data
      return parseApiErrorMessage(body) ?? 'リクエストに失敗しました'
    },
  }
}
