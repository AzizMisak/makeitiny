import type {
  AuthResponse,
  LoginPayload,
  RegisterPayload,
  ResetPasswordPayload,
  RoleName,
  User
} from '~/types/auth'

export type PlanFeatureKey = 'allowCustomAlias' | 'allowPassword' | 'allowQr' | 'allowAds'

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

  const roles = computed<RoleName[]>(() => {
    const r = user.value?.roles
    return r && r.length > 0 ? r : ['user']
  })

  const isOwner = computed(() => roles.value.includes('owner'))
  const isAdmin = computed(() => roles.value.includes('admin') || roles.value.includes('owner'))

  const primaryRole = computed<RoleName>(() => {
    if (isOwner.value) return 'owner'
    if (roles.value.includes('admin')) return 'admin'
    return 'user'
  })

  const primaryRoleLabel = computed(() => {
    switch (primaryRole.value) {
      case 'owner':
        return 'Owner (Sistem Sahibi)'
      case 'admin':
        return 'Admin (Yönetici)'
      default:
        return 'Kullanıcı'
    }
  })

  function hasAnyRole(allowed?: string[]): boolean {
    if (!allowed || allowed.length === 0) return true
    return roles.value.some((r) => allowed.includes(r))
  }

  function canUseFeature(feature: PlanFeatureKey): boolean {
    if (isAdmin.value) return true
    return Boolean(user.value?.plan?.[feature])
  }

  const remainingLinks = computed<number | null>(() => {
    if (isAdmin.value) return null
    const maxLinks = user.value?.plan?.maxLinks ?? 50
    if (maxLinks < 0) return null
    const used = user.value?.linkCount ?? 0
    return Math.max(0, maxLinks - used)
  })

  async function fetchCurrentUser(): Promise<User> {
    const fresh = await useApi().get<User>('/auth/user')
    user.value = fresh
    return fresh
  }

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

  return {
    token,
    user,
    isLoggedIn,
    roles,
    isOwner,
    isAdmin,
    primaryRole,
    primaryRoleLabel,
    remainingLinks,
    hasAnyRole,
    canUseFeature,
    fetchCurrentUser,
    login,
    register,
    logout,
    resetPassword,
    deleteAccount
  }
}
