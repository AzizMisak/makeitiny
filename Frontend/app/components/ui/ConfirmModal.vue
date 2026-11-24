<script setup lang="ts">
export interface ConfirmChangeItem {
  label: string
  from: string
  to: string
}

const props = withDefaults(
  defineProps<{
    open: boolean
    title: string
    description: string
    targetLabel?: string
    changes?: ConfirmChangeItem[]
    confirmLabel?: string
    cancelLabel?: string
    variant?: 'primary' | 'danger'
    loading?: boolean
  }>(),
  {
    targetLabel: '',
    changes: () => [],
    confirmLabel: 'Onayla',
    cancelLabel: 'Vazgeç',
    variant: 'primary',
    loading: false
  }
)

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const cancelBtnRef = ref<HTMLButtonElement | null>(null)

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen && import.meta.client) {
      nextTick(() => cancelBtnRef.value?.focus())
    }
  }
)
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center bg-ink/50 p-4"
      role="dialog"
      aria-modal="true"
      :aria-label="title"
      @keydown.esc="emit('cancel')"
    >
      <div class="w-full max-w-md rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard-lg">
        <div class="flex items-start justify-between gap-3 border-b border-line pb-4">
          <h3 class="text-lg font-bold text-ink">{{ title }}</h3>
          <UiBadge :tone="variant === 'danger' ? 'accent' : 'neutral'">
            {{ variant === 'danger' ? 'Kritik İşlem' : 'Onay Gerekli' }}
          </UiBadge>
        </div>

        <p v-if="targetLabel" class="mt-4 rounded-md border border-line bg-paper px-3 py-2 font-mono text-xs font-semibold text-ink">
          {{ targetLabel }}
        </p>

        <p class="mt-3 text-sm leading-relaxed text-ink-2">
          {{ description }}
        </p>

        <dl v-if="changes.length > 0" class="mt-4 divide-y divide-line rounded-md border border-line bg-paper text-xs">
          <div v-for="ch in changes" :key="ch.label" class="flex items-center justify-between gap-2 px-3 py-2">
            <dt class="font-medium text-ink-3">{{ ch.label }}</dt>
            <dd class="font-mono font-semibold text-ink">
              <span class="text-ink-3 line-through">{{ ch.from }}</span>
              <span class="mx-1.5 text-ink-3">→</span>
              <span class="text-ink">{{ ch.to }}</span>
            </dd>
          </div>
        </dl>

        <div class="mt-6 flex flex-wrap items-center justify-end gap-2.5 border-t border-line pt-4">
          <button
            ref="cancelBtnRef"
            type="button"
            :disabled="loading"
            class="inline-flex cursor-pointer items-center justify-center rounded-md border-[1.5px] border-ink bg-card px-4 py-2 text-sm font-semibold text-ink transition-colors hover:bg-paper-2 disabled:cursor-not-allowed disabled:opacity-50"
            @click="emit('cancel')"
          >
            {{ cancelLabel }}
          </button>

          <UiButton
            :variant="variant === 'danger' ? 'danger' : 'primary'"
            :disabled="loading"
            @click="emit('confirm')"
          >
            {{ loading ? 'İşleniyor...' : confirmLabel }}
          </UiButton>
        </div>
      </div>
    </div>
  </Teleport>
</template>
