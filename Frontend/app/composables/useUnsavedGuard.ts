import { onBeforeRouteLeave } from 'vue-router'

/** Kaydedilmemiş değişiklik varken sayfadan çıkışı sorar. */
export function useUnsavedGuard(dirty: () => boolean) {
  const { ask } = useConfirm()

  onBeforeRouteLeave(async () => {
    if (!dirty()) return true
    return await ask({
      title: 'Değişiklikler kaydedilmedi',
      message: 'Bu sayfadan çıkarsan yaptığın değişiklikler kaybolur.',
      confirmLabel: 'Kaydetmeden çık',
      cancelLabel: 'Sayfada kal',
      tone: 'danger'
    })
  })

  function onUnload(e: BeforeUnloadEvent) {
    if (!dirty()) return
    e.preventDefault()
    e.returnValue = ''
  }

  onMounted(() => window.addEventListener('beforeunload', onUnload))
  onBeforeUnmount(() => window.removeEventListener('beforeunload', onUnload))
}
