<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    page: number
    totalPages: number
    total: number
    pageSize: number
    loading?: boolean
    itemLabel?: string
  }>(),
  {
    loading: false,
    itemLabel: 'kayıt'
  }
)

const emit = defineEmits<{ change: [page: number] }>()

const rangeStart = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1))
const rangeEnd = computed(() => Math.min(props.page * props.pageSize, props.total))

const visiblePages = computed<(number | '...')[]>(() => {
  const total = props.totalPages
  const current = props.page
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  const pages: (number | '...')[] = [1]
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let p = start; p <= end; p++) {
    pages.push(p)
  }
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

function goTo(targetPage: number) {
  if (props.loading || targetPage < 1 || targetPage > props.totalPages || targetPage === props.page) return
  emit('change', targetPage)
}
</script>

<template>
  <div
    v-if="total > 0"
    class="mt-6 flex flex-wrap items-center justify-between gap-4 rounded-lg border-[1.5px] border-ink bg-card px-5 py-3.5 shadow-hard"
  >
    <p class="text-xs text-ink-2">
      Toplam <span class="font-mono font-bold text-ink">{{ total.toLocaleString('tr-TR') }}</span> {{ itemLabel }} içinden
      <span class="font-mono font-bold text-ink">{{ rangeStart }}–{{ rangeEnd }}</span> arası gösteriliyor
      <span class="text-ink-3">(Sayfa {{ page }} / {{ totalPages }})</span>
    </p>

    <div class="flex flex-wrap items-center gap-1.5">
      <button
        type="button"
        :disabled="page <= 1 || loading"
        class="inline-flex cursor-pointer items-center gap-1 rounded-sm border border-ink bg-card px-2.5 py-1 text-xs font-semibold text-ink transition-colors hover:bg-paper-2 disabled:cursor-not-allowed disabled:opacity-40"
        @click="goTo(page - 1)"
      >
        ← Önceki
      </button>

      <template v-for="(item, idx) in visiblePages" :key="`${item}-${idx}`">
        <span v-if="item === '...'" class="select-none px-1.5 font-mono text-xs text-ink-3">...</span>
        <button
          v-else
          type="button"
          :disabled="loading"
          class="cursor-pointer rounded-sm border px-2.5 py-1 font-mono text-xs font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-50"
          :class="
            item === page
              ? 'border-ink bg-ink text-paper'
              : 'border-line bg-card text-ink-2 hover:border-ink hover:bg-paper-2 hover:text-ink'
          "
          @click="goTo(item)"
        >
          {{ item }}
        </button>
      </template>

      <button
        type="button"
        :disabled="page >= totalPages || loading"
        class="inline-flex cursor-pointer items-center gap-1 rounded-sm border border-ink bg-card px-2.5 py-1 text-xs font-semibold text-ink transition-colors hover:bg-paper-2 disabled:cursor-not-allowed disabled:opacity-40"
        @click="goTo(page + 1)"
      >
        Sonraki →
      </button>
    </div>
  </div>
</template>
