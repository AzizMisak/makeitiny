export type RoleName = 'user' | 'admin' | 'owner'

export interface PlanItem {
  id: number
  slug: string
  name: string
  priceMonthly: number
  maxLinks: number
  allowCustomAlias: boolean
  allowPassword: boolean
  allowQr: boolean
  allowAds: boolean
  analyticsDays: number
}

/** Backend'deki kullanıcı. GET /auth/user, POST /auth/login, POST /auth/register ve POST /auth/refresh yanıtları. */
export interface User {
  id: string
  name: string
  email: string
  isActive: boolean
  banReason: string
  bannedUntil?: string
  roles: RoleName[]
  plan: PlanItem
  planExpiresAt?: string
  linkCount: number
  createdAt: string
}

/** POST /auth/login, POST /auth/register ve POST /auth/refresh yanıtı. */
export interface AuthResponse {
  token: string
  user: User
}

/** POST /auth/login gövdesi. */
export interface LoginPayload {
  email: string
  password: string
}

/** POST /auth/register gövdesi. */
export interface RegisterPayload {
  name?: string
  email: string
  password: string
}

/** POST /auth/resetpass gövdesi. */
export interface ResetPasswordPayload {
  email: string
  currentPassword?: string
  newPassword: string
}
