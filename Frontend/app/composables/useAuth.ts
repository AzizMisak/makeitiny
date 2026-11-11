import type { AuthResponse, LoginPayload, RegisterPayload, ResetPasswordPayload, User } from '~/types/auth'

export function useAuth() {
  const token = useCookie<string | null>('mit_token', {
    default: () => null,
    maxAge: 60 * 60 * 24 * 7
  })

  const user = useCookie<User | null>('mit_session', {
    default: () => null,
    maxAge: 60 * 60 * 24 * 7
  })

  const isLoggedIn = computed(() => !!token.value && !!user.value)

  async function login(payload: LoginPayload) {
    const res = await useApi().post<AuthResponse, LoginPayload>('/auth/login', payload)
    token.value = res.token
    user.value = res.user
    return res.user
  }

  async function register(payload: RegisterPayload) {
    const res = await useApi().post<AuthResponse, RegisterPayload>('/auth/register', payload)
    token.value = res.token
    user.value = res.user
    return res.user
  }

  async function logout() {
    if (token.value) {
      await useApi().post('/auth/logout').catch(() => {})
    }
    token.value = null
    user.value = null
  }

  async function resetPassword(payload: ResetPasswordPayload) {
    await useApi().post('/auth/resetpass', payload)
  }

  async function deleteAccount() {
    await useApi().del('/auth/user')
    token.value = null
    user.value = null
  }

  return { token, user, isLoggedIn, login, register, logout, resetPassword, deleteAccount }
}
