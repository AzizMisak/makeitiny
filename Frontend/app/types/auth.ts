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

/** Backend'deki kullanıcı. GET /auth/user, POST /auth/login ve POST /auth/register yanıtları. */
export interface User {
  id: string
  name: string
  email: string
  isActive: boolean
  roles: RoleName[]
  plan: PlanItem
  planExpiresAt?: string
  linkCount: number
  createdAt: string
}

/** POST /auth/login ve POST /auth/register yanıtı. */
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
  newPassword: string
}
