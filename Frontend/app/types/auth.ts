/** Backend'deki kullanıcı. GET /auth/user yanıtı. */
export interface User {
  id: string
  name: string
  email: string
}

/** POST /auth/login gövdesi. */
export interface LoginPayload {
  email: string
  password: string
}

/** POST /auth/register gövdesi. */
export interface RegisterPayload {
  email: string
  password: string
}

/** POST /auth/resetpass gövdesi. */
export interface ResetPasswordPayload {
  email: string
}
