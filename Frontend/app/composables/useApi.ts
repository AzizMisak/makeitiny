import type { ApiError } from '~/types/api'
import type { AuthResponse, User } from '~/types/auth'

type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

interface RequestOptions<TBody> {
  body?: TBody
  query?: Record<string, string | number | boolean | undefined>
  skipAuthRefresh?: boolean
  responseType?: 'json' | 'blob'
}

let refreshPromise: Promise<string | null> | null = null

/** Askıya alınan hesabın oturumu kapatılırken giriş ekranında gösterilecek mesaj. */
export function useAuthNotice() {
  return useState<string>('mit_auth_notice', () => '')
}

/**
 * Backend'e tip güvenli istek atar.
 * - 401: HttpOnly refresh çereziyle bir kez yenilenip istek tekrarlanır. Aynı anda düşen istekler tek yenilemeyi paylaşır;
 *   gönderildiği token zaten yenilenmişse yeniden yenilemeye gitmeden yeni token'la tekrarlanır.
 * - 403 account_suspended: oturum kapatılır ve kullanıcı giriş ekranına sebebiyle birlikte gönderilir.
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
      } catch (e: any) {
        const statusCode = e?.statusCode ?? e?.response?.status
        if (statusCode === 403 && e?.data?.code === 'account_suspended' && typeof e?.data?.message === 'string') {
          useAuthNotice().value = e.data.message
        }
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
      credentials: 'include',
      responseType: options.responseType as any
    })
  }

  async function endSuspendedSession(message: string) {
    token.value = null
    user.value = null
    clearNuxtState(['mit_space_last'])
    useLinks().resetLinks()
    useAuthNotice().value = message
    if (import.meta.client) {
      await navigateTo('/auth/login', { replace: true })
    }
  }

  async function parseApiError(e: any): Promise<ApiError> {
    const data = e?.data
    if (typeof Blob !== 'undefined' && data instanceof Blob) {
      try {
        const text = await data.text()
        const parsed = JSON.parse(text)
        if (parsed && typeof parsed.code === 'string' && typeof parsed.message === 'string') {
          return parsed as ApiError
        }
      } catch {
        // Geçersiz JSON blob durumunda genel ağ/sunucu hatasına düşülür.
      }
      return { code: 'network_error', message: 'Dosya hazırlanamadı. Tekrar dene.' }
    }
    return data ?? { code: 'network_error', message: 'Sunucuya ulaşılamadı.' }
  }

  async function request<TResponse, TBody = unknown>(
    method: Method,
    path: string,
    options: RequestOptions<TBody> = {}
  ): Promise<TResponse> {
    const sentToken = token.value
    try {
      return await executeFetch<TResponse, TBody>(method, path, options, sentToken)
    } catch (e: any) {
      const statusCode = e?.statusCode ?? e?.response?.status
      const isAuthEndpoint =
        path.startsWith('/auth/login') ||
        path.startsWith('/auth/register') ||
        path.startsWith('/auth/refresh') ||
        path.startsWith('/links/anonymous')

      const apiErr = await parseApiError(e)

      if (statusCode === 403 && apiErr.code === 'account_suspended' && !isAuthEndpoint) {
        await endSuspendedSession(apiErr.message)
        throw apiErr
      }

      if (statusCode === 401 && !options.skipAuthRefresh && !isAuthEndpoint && import.meta.client) {
        // Başka bir istek token'ı bu arada yenilediyse yeniden yenilemeye gerek yok.
        const newToken = token.value && token.value !== sentToken ? token.value : await refreshAccessToken()
        if (newToken) {
          try {
            return await executeFetch<TResponse, TBody>(method, path, options, newToken)
          } catch (retryErr: any) {
            const retryApiErr = await parseApiError(retryErr)
            if (retryApiErr.code === 'account_suspended') {
              await endSuspendedSession(retryApiErr.message)
            }
            throw retryApiErr
          }
        }
        clearNuxtState(['mit_space_last'])
        useLinks().resetLinks()
        await navigateTo('/auth/login')
      }

      throw apiErr
    }
  }

  return {
    get: <T>(path: string, query?: RequestOptions<never>['query']) => request<T>('GET', path, { query }),
    post: <T, B = unknown>(path: string, body?: B) => request<T, B>('POST', path, { body }),
    put: <T, B = unknown>(path: string, body?: B) => request<T, B>('PUT', path, { body }),
    patch: <T, B = unknown>(path: string, body?: B) => request<T, B>('PATCH', path, { body }),
    del: <T = void>(path: string) => request<T>('DELETE', path),
    /** Dosya indirir (oturum ve token yenileme dahil) ve Blob döner. */
    blob: (path: string, query?: RequestOptions<never>['query']) => request<Blob>('GET', path, { query, responseType: 'blob' })
  }
}
