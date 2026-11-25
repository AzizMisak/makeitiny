export type ThemeMode = 'light' | 'dark'

export function useTheme() {
  const themeCookie = useCookie<ThemeMode | null>('mit_theme', {
    default: () => null,
    maxAge: 60 * 60 * 24 * 365
  })

  const isDark = useState<boolean>('mit_is_dark', () => themeCookie.value === 'dark')

  function applyTheme(dark: boolean) {
    isDark.value = dark
    const mode: ThemeMode = dark ? 'dark' : 'light'
    themeCookie.value = mode
    if (import.meta.client) {
      document.documentElement.classList.toggle('dark', dark)
      try {
        localStorage.setItem('mit_theme', mode)
      } catch {
        // ignore storage errors
      }
    }
  }

  function syncFromDom() {
    if (!import.meta.client) return
    isDark.value = document.documentElement.classList.contains('dark')
  }

  function toggleTheme() {
    applyTheme(!isDark.value)
  }

  return {
    isDark,
    toggleTheme,
    applyTheme,
    syncFromDom
  }
}
