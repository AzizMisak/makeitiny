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

  async function refreshSession(): Promise<boolean> {
    try {
      const res = await useApi().post<AuthResponse>('/auth/refresh')
      token.value = res.token
      user.value = res.user
      return true
    } catch {
      token.value = null
      user.value = null
      clearNuxtState(['mit_space_last'])
      return false
    }
  }

  async function login(payload: LoginPayload) {
    const res = await useApi().post<AuthResponse, LoginPayload>('/auth/login', payload)
    token.value = res.token
    user.value = res.user
    clearNuxtState(['mit_space_last'])
    return res.user
  }

  async function register(payload: RegisterPayload) {
    const res = await useApi().post<AuthResponse, RegisterPayload>('/auth/register', payload)
    token.value = res.token
    user.value = res.user
    clearNuxtState(['mit_space_last'])
    return res.user
  }

  /** Görünen adı değiştirir; sunucu boşlukları kırpar, 2–40 karakter ister. */
  async function updateName(name: string) {
    const fresh = await useApi().patch<User, { name: string }>('/auth/user', { name })
    user.value = fresh
    return fresh
  }

  /** Standart kullanıcının paketini değiştirir; sınırı aşan en eski linkler 2 ay askıya alınır. */
  async function changePlan(planId: number) {
    const fresh = await useApi().put<User, { planId: number }>('/auth/user/plan', { planId })
    user.value = fresh
    return fresh
  }

  /** Verilen kullanıcı kimliği oturumdaki kişiye mi ait? */
  function isSelf(id?: string): boolean {
    return !!id && id === user.value?.id
  }

  async function logout() {
    if (token.value) {
      await useApi().post('/auth/logout').catch(() => {})
    }
    token.value = null
    user.value = null
    // Aynı sekmede başka bir hesapla girilirse önceki hesabın panel sayfaları hatırlanmasın.
    clearNuxtState(['mit_space_last'])
  }

  async function resetPassword(payload: ResetPasswordPayload) {
    await useApi().post('/auth/resetpass', payload)
  }

  async function deleteAccount() {
    await useApi().del('/auth/user')
    token.value = null
    user.value = null
    clearNuxtState(['mit_space_last'])
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
    refreshSession,
    login,
    register,
    logout,
    resetPassword,
    updateName,
    changePlan,
    isSelf,
    deleteAccount
  }
}
