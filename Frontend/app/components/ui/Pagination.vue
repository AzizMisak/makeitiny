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
  { loading: false, itemLabel: 'kayıt' }
)

const emit = defineEmits<{ change: [page: number] }>()

const pending = ref<number | null>(null)
watch(
  () => props.loading,
  (v) => {
    if (!v) pending.value = null
  }
)

const start = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1))
const end = computed(() => Math.min(props.page * props.pageSize, props.total))

const pages = computed<(number | string)[]>(() => {
  const { totalPages: total, page } = props
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const list: (number | string)[] = [1]
  if (page > 3) list.push('a')
  for (let p = Math.max(2, page - 1); p <= Math.min(total - 1, page + 1); p++) list.push(p)
  if (page < total - 2) list.push('b')
  list.push(total)
  return list
})

function go(p: number) {
  if (props.loading || p < 1 || p > props.totalPages || p === props.page) return
  pending.value = p
  emit('change', p)
}
</script>

<template>
  <nav v-if="total > 0" class="flex flex-wrap items-center justify-between gap-4 pt-5" aria-label="Sayfalar">
    <p class="tnum text-sm text-ink-3">
      {{ formatNumber(total) }} {{ itemLabel }} içinden {{ formatNumber(start) }}–{{ formatNumber(end) }}
    </p>
    <div v-if="totalPages > 1" class="flex items-center gap-1">
      <UiIconButton icon="left" label="Önceki sayfa" :disabled="page <= 1 || loading" @click="go(page - 1)" />
      <template v-for="p in pages" :key="p">
        <span v-if="typeof p === 'string'" class="w-6 text-center text-ink-3">…</span>
        <button
          v-else
          type="button"
          class="tnum inline-flex h-10 min-w-10 items-center justify-center rounded-md px-2 text-sm font-medium transition-colors"
          :class="p === page ? 'bg-brand-soft text-brand-ink' : 'text-ink-2 hover:bg-sunken hover:text-ink'"
          :aria-current="p === page ? 'page' : undefined"
          :disabled="loading"
          @click="go(p)"
        >
          <UiSpinner v-if="loading && pending === p" :size="14" />
          <template v-else>{{ p }}</template>
        </button>
      </template>
      <UiIconButton icon="right" label="Sonraki sayfa" :disabled="page >= totalPages || loading" @click="go(page + 1)" />
    </div>
  </nav>
</template>
