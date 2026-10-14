import type { ApiError } from '~/types/api'

type Method = 'GET' | 'POST' | 'PATCH' | 'DELETE'

interface RequestOptions<TBody> {
  body?: TBody
  query?: Record<string, string | number | boolean | undefined>
}

/**
 * Backend'e tip güvenli istek atar. Dönüş tipi çağıran tarafından verilir:
 *   const links = await useApi().get<Paginated<LinkItem>>('/links')
 * Hatalar ApiError şekline çevrilip fırlatılır.
 */
export function useApi() {
  const { public: config } = useRuntimeConfig()

  async function request<TResponse, TBody = unknown>(
    method: Method,
    path: string,
    options: RequestOptions<TBody> = {}
  ): Promise<TResponse> {
    try {
      return await $fetch<TResponse>(path, {
        baseURL: config.apiBase,
        method,
        body: options.body as Record<string, unknown> | undefined,
        query: options.query,
        credentials: 'include'
      })
    } catch (e: any) {
      const error: ApiError = e?.data ?? { code: 'network_error', message: 'Sunucuya ulaşılamadı.' }
      throw error
    }
  }

  return {
    get: <T>(path: string, query?: RequestOptions<never>['query']) => request<T>('GET', path, { query }),
    post: <T, B = unknown>(path: string, body?: B) => request<T, B>('POST', path, { body }),
    patch: <T, B = unknown>(path: string, body?: B) => request<T, B>('PATCH', path, { body }),
    del: <T = void>(path: string) => request<T>('DELETE', path)
  }
}
