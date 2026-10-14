// Şimdilik sahte oturum: çerezde saklanır. API bağlanınca sadece bu dosya değişir.
import type { User } from '~/types/auth'

export function useAuth() {
  const user = useCookie<User | null>('mit_session', {
    default: () => null,
    maxAge: 60 * 60 * 24 * 7
  })

  const isLoggedIn = computed(() => !!user.value)

  function login(email: string) {
    const name = email.split('@')[0] || 'kullanici'
    user.value = { id: 'demo', name, email }
  }

  function logout() {
    user.value = null
  }

  return { user, isLoggedIn, login, logout }
}
