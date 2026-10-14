export interface LinkItem {
  id: string
  alias: string
  target: string
  clicks: number
  locked: boolean
  createdAt: string
}

/** POST /links gövdesi. */
export interface CreateLinkPayload {
  target: string
  alias?: string
  password?: string
}

/** PATCH /links/:id gövdesi. Tüm alanlar isteğe bağlı. */
export type UpdateLinkPayload = Partial<CreateLinkPayload>

/** GET /links/check-alias/:alias yanıtı. */
export interface AliasCheckResult {
  alias: string
  available: boolean
}

/** Router: GET /:link/preview yanıtı. */
export interface LinkPreview {
  alias: string
  target: string
  locked: boolean
}
