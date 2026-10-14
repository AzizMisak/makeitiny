// Backend'in ortak yanıt/hata zarfları. Endpoint'ler bunları generic olarak kullanır.

/** Backend'in döndüğü hata gövdesi. */
export interface ApiError {
  code: string
  message: string
  /** Alan bazlı doğrulama hataları: { email: 'Geçersiz e-posta' } */
  fields?: Record<string, string>
}

/** Sayfalı liste yanıtı. */
export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

/** Asenkron verinin ekranda üç hali. */
export type Loadable<T> =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'error'; error: ApiError }
  | { status: 'success'; data: T }
