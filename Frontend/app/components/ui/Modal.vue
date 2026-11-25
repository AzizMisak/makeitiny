<script setup lang="ts">
const props = withDefaults(
  defineProps<{ open: boolean; title: string; busy?: boolean; wide?: boolean }>(),
  { busy: false, wide: false }
)
const emit = defineEmits<{ close: [] }>()

const panel = ref<HTMLElement | null>(null)
const titleId = useId()
let lastFocus: HTMLElement | null = null

function focusables(): HTMLElement[] {
  return [...(panel.value?.querySelectorAll<HTMLElement>('button, a[href], input, select, textarea') ?? [])].filter(
    (el) => !el.hasAttribute('disabled') && el.offsetParent !== null
  )
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    e.stopPropagation()
    if (!props.busy) emit('close')
    return
  }
  if (e.key !== 'Tab') return
  const items = focusables()
  const first = items[0]
  const last = items[items.length - 1]
  if (!first || !last) return
  if (e.shiftKey && document.activeElement === first) {
    e.preventDefault()
    last.focus()
  } else if (!e.shiftKey && document.activeElement === last) {
    e.preventDefault()
    first.focus()
  }
}

watch(
  () => props.open,
  async (open) => {
    if (!import.meta.client) return
    if (open) {
      lastFocus = document.activeElement as HTMLElement | null
      document.body.style.overflow = 'hidden'
      await nextTick()
      const preferred = panel.value?.querySelector<HTMLElement>('[data-autofocus]')
      ;(preferred ?? panel.value)?.focus()
    } else {
      document.body.style.overflow = ''
      lastFocus?.focus?.()
    }
  }
)

onBeforeUnmount(() => {
  if (import.meta.client) document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="open"
        class="fixed inset-0 z-[80] flex items-end justify-center bg-black/55 p-0 sm:items-center sm:p-6"
        @mousedown.self="!busy && emit('close')"
        @keydown="onKey"
      >
        <div
          ref="panel"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="titleId"
          tabindex="-1"
          class="flex max-h-[92dvh] w-full flex-col overflow-hidden bg-surface shadow-pop outline-none sm:rounded-lg"
          :class="wide ? 'sm:max-w-2xl' : 'sm:max-w-lg'"
        >
          <header class="flex items-start justify-between gap-4 px-6 pb-1 pt-6">
            <h2 :id="titleId" class="display text-2xl">{{ title }}</h2>
            <UiIconButton icon="x" label="Pencereyi kapat" class="-mr-2 -mt-1.5" :disabled="busy" @click="emit('close')" />
          </header>
          <div class="min-h-0 flex-1 overflow-y-auto px-6 pb-2">
            <slot />
          </div>
          <footer v-if="$slots.footer" class="flex flex-wrap items-center justify-end gap-2 px-6 pb-6 pt-4">
            <slot name="footer" />
          </footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
