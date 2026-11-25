export type Space = 'user' | 'admin'

/**
 * Kullanıcı ve yönetim panelleri iki ayrı çalışma alanıdır.
 * Her alan en son açık olan sayfasını hatırlar; geçişte oraya dönülür.
 */
export function useSpace() {
  const route = useRoute()
  const router = useRouter()
  const lastPath = useState<Record<Space, string>>('mit_space_last', () => ({
    user: '/dashboard',
    admin: '/dashboard/admin'
  }))
  const switchingTo = useState<Space | null>('mit_space_switching', () => null)

  const current = computed<Space>(() =>
    route.path.startsWith('/dashboard/admin') || route.path.startsWith('/dashboard/owner') ? 'admin' : 'user'
  )

  function remember() {
    if (!route.path.startsWith('/dashboard') || route.path === '/dashboard/forbidden') return
    lastPath.value = { ...lastPath.value, [current.value]: route.fullPath }
  }

  async function switchTo(space: Space) {
    if (switchingTo.value) return
    // Perde, kaydedilmemiş değişiklik sorusu geçildikten sonra iner.
    const stop = router.beforeResolve(async () => {
      switchingTo.value = space
      await new Promise((r) => setTimeout(r, 280))
    })
    try {
      await navigateTo(lastPath.value[space])
    } finally {
      stop()
      switchingTo.value = null
    }
  }

  return { current, switchingTo, remember, switchTo }
}
