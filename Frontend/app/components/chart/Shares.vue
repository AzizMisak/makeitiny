<script setup lang="ts">
import type { ShareRow } from '~/utils/analytics'

const props = withDefaults(defineProps<{ title: string; rows: ShareRow[]; limit?: number }>(), { limit: 5 })

const shown = computed(() => {
  if (props.rows.length <= props.limit) return props.rows
  const head = props.rows.slice(0, props.limit - 1)
  const tail = props.rows.slice(props.limit - 1)
  return [
    ...head,
    {
      label: `Diğer (${tail.length})`,
      count: tail.reduce((s, r) => s + r.count, 0),
      percent: Math.round(tail.reduce((s, r) => s + r.percent, 0) * 10) / 10
    }
  ]
})
</script>

<template>
  <section>
    <h3 class="font-semibold">{{ title }}</h3>
    <ul class="mt-3 flex flex-col gap-3">
      <li v-for="r in shown" :key="r.label">
        <p class="flex items-baseline justify-between gap-3 text-sm">
          <span class="min-w-0 truncate">{{ r.label }}</span>
          <span class="tnum shrink-0 text-ink-3"><span class="font-semibold text-ink">%{{ r.percent.toLocaleString('tr-TR') }}</span>&ensp;{{ formatNumber(r.count) }}</span>
        </p>
        <span class="mt-1 block h-2 bg-sunken"><span class="block h-full rounded-r-[2px] bg-brand" :style="{ width: `${r.percent}%` }" /></span>
      </li>
    </ul>
  </section>
</template>
