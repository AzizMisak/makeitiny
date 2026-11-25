/** Panoya kopyalama; kopyalanan metin kısa süre `copied` içinde kalır. */
export function useCopy(resetMs = 1600) {
  const copied = ref('')
  let timer: ReturnType<typeof setTimeout> | null = null

  async function copy(text: string): Promise<boolean> {
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      useToast().error('Panoya kopyalanamadı. Adresi elle seçip kopyala.')
      return false
    }
    copied.value = text
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => (copied.value = ''), resetMs)
    return true
  }

  onBeforeUnmount(() => {
    if (timer) clearTimeout(timer)
  })

  return { copied, copy }
}
