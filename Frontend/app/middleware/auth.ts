export default defineNuxtRouteMiddleware(async (to) => {
  const { token, user, isLoggedIn, fetchCurrentUser, hasAnyRole } = useAuth()

  if (!isLoggedIn.value) {
    return navigateTo('/auth/login')
  }

  // Eski oturum çerezinde rol veya paket alanı yoksa sunucudan taze kullanıcı bilgisini çek.
  if (!user.value?.roles || !user.value?.plan) {
    try {
      await fetchCurrentUser()
    } catch {
      token.value = null
      user.value = null
      return navigateTo('/auth/login')
    }
  }

  const requiredRoles = to.meta.roles as string[] | undefined
  if (requiredRoles && requiredRoles.length > 0 && !hasAnyRole(requiredRoles)) {
    return navigateTo({
      path: '/dashboard/forbidden',
      query: {
        from: to.fullPath,
        required: requiredRoles.join(', ')
      }
    })
  }
})
