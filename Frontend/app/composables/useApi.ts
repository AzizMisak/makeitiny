import type { ApiError } from '~/types/api'
import type { AuthResponse, User } from '~/types/auth'

type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

interface RequestOptions<TBody> {
  body?: TBody
  query?: Record<string, string | number | boolean | undefined>
  skipAuthRefresh?: boolean
}

let refreshPromise: Promise<string | null> | null = null

/**
 * Backend'e tip güvenli istek atar. 401 dönen isteklerde otomatik olarak
 * HttpOnly refresh cookie üzerinden /auth/refresh ile token yeniler ve isteği tekrarlar.
 */
export function useApi() {
  const { public: config } = useRuntimeConfig()
  const token = useCookie<string | null>('mit_token', { maxAge: 60 * 60 * 24 * 7 })
  const user = useCookie<User | null>('mit_session', { maxAge: 60 * 60 * 24 * 7 })

  const baseURL = import.meta.server ? 'http://api:8080' : config.apiBase

  async function refreshAccessToken(): Promise<string | null> {
    if (import.meta.server) return null
    if (refreshPromise) return refreshPromise

    refreshPromise = (async () => {
      try {
        const res = await $fetch<AuthResponse>('/auth/refresh', {
          baseURL,
          method: 'POST',
          credentials: 'include'
        })
        token.value = res.token
        user.value = res.user
        return res.token
      } catch {
        token.value = null
        user.value = null
        return null
      } finally {
        refreshPromise = null
      }
    })()

    return refreshPromise
  }

  async function executeFetch<TResponse, TBody = unknown>(
    method: Method,
    path: string,
    options: RequestOptions<TBody>,
    bearerToken: string | null
  ): Promise<TResponse> {
    const headers: Record<string, string> = {}
    if (bearerToken) {
      headers.Authorization = `Bearer ${bearerToken}`
    }

    return await $fetch<TResponse>(path, {
      baseURL,
      method,
      headers,
      body: options.body as Record<string, unknown> | undefined,
      query: options.query,
      credentials: 'include'
    })
  }

  async function request<TResponse, TBody = unknown>(
    method: Method,
    path: string,
    options: RequestOptions<TBody> = {}
  ): Promise<TResponse> {
    try {
      return await executeFetch<TResponse, TBody>(method, path, options, token.value)
    } catch (e: any) {
      const statusCode = e?.statusCode ?? e?.response?.status
      const isAuthEndpoint =
        path.startsWith('/auth/login') ||
        path.startsWith('/auth/register') ||
        path.startsWith('/auth/refresh') ||
        path.startsWith('/links/anonymous')

      if (statusCode === 401 && !options.skipAuthRefresh && !isAuthEndpoint && import.meta.client) {
        const newToken = await refreshAccessToken()
        if (newToken) {
          try {
            return await executeFetch<TResponse, TBody>(method, path, options, newToken)
          } catch (retryErr: any) {
            const error: ApiError = retryErr?.data ?? { code: 'network_error', message: 'Sunucuya ulaşılamadı.' }
            throw error
          }
        }
        await navigateTo('/auth/login')
      }

      const error: ApiError = e?.data ?? { code: 'network_error', message: 'Sunucuya ulaşılamadı.' }
      throw error
    }
  }

  return {
    get: <T>(path: string, query?: RequestOptions<never>['query']) => request<T>('GET', path, { query }),
    post: <T, B = unknown>(path: string, body?: B) => request<T, B>('POST', path, { body }),
    put: <T, B = unknown>(path: string, body?: B) => request<T, B>('PUT', path, { body }),
    patch: <T, B = unknown>(path: string, body?: B) => request<T, B>('PATCH', path, { body }),
    del: <T = void>(path: string) => request<T>('DELETE', path)
  }
}
