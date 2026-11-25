/** Çıkışı onaylatıp oturumu kapatır. */
export function useSignOut() {
  const { logout } = useAuth()
  const { resetLinks } = useLinks()
  const { ask } = useConfirm()

  return async function signOut() {
    const ok = await ask({
      title: 'Çıkış yapılsın mı?',
      message: 'Bu cihazdaki oturumun kapanır. Tekrar girmek için e-posta ve parolan gerekir.',
      confirmLabel: 'Çıkış yap',
      cancelLabel: 'Oturumda kal',
      action: async () => {
        await logout()
        resetLinks()
      }
    })
    if (ok) await navigateTo('/auth/login', { replace: true })
  }
}
