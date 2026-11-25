import type { RoleName, User } from './auth'

export type SecurityStatus = 'unchecked' | 'safe' | 'suspicious' | 'malicious' | 'blocked' | string

export interface LinkItem {
  id: string
  alias: string
  target: string
  title: string
  clicks: number
  locked: boolean
  waitEnabled: boolean
  waitSeconds: number
  isActive: boolean
  expiresAt?: string
  maxClicks?: number
  adEnabled: boolean
  securityStatus: SecurityStatus
  hasQr: boolean
  qrUrl: string
  isAnonymous: boolean
  createdAt: string
}

export interface AdminLinkItem extends LinkItem {
  userId: string
  userEmail: string
  userName: string
}

export interface CreateAnonymousLinkPayload {
  target: string
  password?: string
  waitEnabled?: boolean
  waitSeconds?: number
  ttlHours?: number
}

export interface AnonymousLinkResponse {
  id: string
  alias: string
  shortUrl: string
  target: string
  locked: boolean
  waitEnabled: boolean
  waitSeconds: number
  expiresAt: string
  hasQr: boolean
  qrUrl: string
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
  countryCode: string
  clickedAt: string
}

export interface PaginatedLinksResponse {
  items: LinkItem[]
  page: number
  limit: number
  total: number
  totalPages: number
  totalClicks: number
  lockedCount: number
  topLinks: LinkItem[]
}

export interface PaginatedLogsResponse {
  items: LinkLogItem[]
  page: number
  limit: number
  total: number
  totalPages: number
}

/** POST /links gövdesi. */
export interface CreateLinkPayload {
  target: string
  alias?: string
  title?: string
  password?: string
  waitEnabled?: boolean
  waitSeconds?: number
  expiresAt?: string
  maxClicks?: number
  adEnabled?: boolean
}

/** PATCH /links/:id gövdesi. Tüm alanlar isteğe bağlı. */
export interface UpdateLinkPayload extends Partial<CreateLinkPayload> {
  isActive?: boolean
}

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

/** Admin / Owner API tipleri */
export interface AdminStatsResponse {
  totalUsers: number
  activeUsers: number
  totalLinks: number
  activeLinks: number
  blockedLinks: number
  totalClicks: number
}

export interface AdminPaginatedLinksResponse {
  items: AdminLinkItem[]
  page: number
  limit: number
  total: number
  totalPages: number
}

export interface AdminUpdateLinkPayload extends UpdateLinkPayload {
  securityStatus?: SecurityStatus
}

export interface AdminPaginatedUsersResponse {
  items: User[]
  page: number
  limit: number
  total: number
  totalPages: number
}

export interface AdminUpdateUserPayload {
  isActive?: boolean
  banReason?: string
  bannedUntil?: string
  planId?: number
  planExpiresAt?: string
}

export interface AdminUpdateRolePayload {
  role: Extract<RoleName, 'user' | 'admin'>
}

export interface AdminUpdatePlanPayload {
  name?: string
  priceMonthly?: number
  maxLinks?: number
  allowCustomAlias?: boolean
  allowPassword?: boolean
  allowQr?: boolean
  allowAds?: boolean
  analyticsDays?: number
}
