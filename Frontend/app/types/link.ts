export interface LinkItem {
  id: string
  alias: string
  target: string
  clicks: number
  locked: boolean
  waitEnabled: boolean
  waitSeconds: number
  createdAt: string
}

export interface LinkLogItem {
  id: string
  linkId: string
  ipAddress: string
  userAgent: string
  deviceType: string
  browser: string
  os: string
  referer: string
  acceptLanguage: string
  clickedAt: string
}

/** POST /links gövdesi. */
export interface CreateLinkPayload {
  target: string
  alias?: string
  password?: string
  waitEnabled?: boolean
  waitSeconds?: number
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
  waitEnabled: boolean
  waitSeconds: number
}
