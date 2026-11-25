/** Kısa adresi tam (kopyalanacak) ve sade (gösterilecek) biçimde üretir. */
export function useShortUrl() {
  const { public: config } = useRuntimeConfig()
  const prefix = `${config.routerBase}/`.replace(/^https?:\/\//, '')

  return {
    prefix,
    full: (alias: string) => `${config.routerBase}/${alias}`,
    display: (alias: string) => `${prefix}${alias}`
  }
}
