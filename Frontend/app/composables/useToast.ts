export interface ToastItem {
  id: number
  tone: 'ok' | 'error' | 'neutral'
  message: string
}

let nextToastId = 1

export function useToast() {
  const toasts = useState<ToastItem[]>('mit_toasts', () => [])

  function remove(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function push(message: string, tone: ToastItem['tone'] = 'neutral', durationMs = 4000) {
    const id = nextToastId++
    toasts.value = [...toasts.value, { id, tone, message }]
    if (import.meta.client && durationMs > 0) {
      setTimeout(() => remove(id), durationMs)
    }
  }

  return {
    toasts,
    remove,
    success: (message: string) => push(message, 'ok', 3500),
    error: (message: string) => push(message, 'error', 6000),
    info: (message: string) => push(message, 'neutral', 4000)
  }
}
