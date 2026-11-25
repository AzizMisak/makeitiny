export interface ConfirmChange {
  label: string
  from: string
  to: string
}

export interface ConfirmOptions {
  title: string
  /** Ne olacağını anlatan tek paragraf. */
  message: string
  /** İşlemin yapılacağı nesne (kısa adres, e-posta vb.). */
  subject?: string
  /** Önce / sonra listesi. */
  changes?: ConfirmChange[]
  /** Eylemin sonuçları; madde madde gösterilir. */
  effects?: string[]
  confirmLabel: string
  cancelLabel?: string
  tone?: 'normal' | 'danger'
  /** Doluysa kullanıcı onaylamak için bu metni yazmak zorundadır. */
  typeToConfirm?: string
  /**
   * Onaylanınca çalışır; bitene kadar pencere yükleniyor durumunda kalır.
   * Hata fırlatırsa pencere açık kalır ve hata gösterilir.
   */
  action?: () => Promise<unknown>
}

interface ConfirmState {
  open: boolean
  busy: boolean
  error: string
}

// Seçenekler fonksiyon taşıdığı için SSR'a serileştirilen useState yerine modül düzeyinde tutulur.
const options = shallowRef<ConfirmOptions | null>(null)
let resolver: ((value: boolean) => void) | null = null

/** Önemli işlemlerden önce tek tip onay penceresi açar. `await ask(...)` onaylandıysa true döner. */
export function useConfirm() {
  const state = useState<ConfirmState>('mit_confirm', () => ({ open: false, busy: false, error: '' }))

  function settle(value: boolean) {
    state.value = { open: false, busy: false, error: '' }
    resolver?.(value)
    resolver = null
  }

  function ask(next: ConfirmOptions): Promise<boolean> {
    resolver?.(false)
    options.value = next
    state.value = { open: true, busy: false, error: '' }
    return new Promise((resolve) => {
      resolver = resolve
    })
  }

  async function accept() {
    const current = options.value
    if (!current || state.value.busy) return
    if (!current.action) {
      settle(true)
      return
    }
    state.value = { open: true, busy: true, error: '' }
    try {
      await current.action()
      settle(true)
    } catch (e: any) {
      state.value = { open: true, busy: false, error: e?.message ?? 'İşlem tamamlanamadı. Tekrar dene.' }
    }
  }

  function cancel() {
    if (state.value.busy) return
    settle(false)
  }

  return { state, options, ask, accept, cancel }
}
