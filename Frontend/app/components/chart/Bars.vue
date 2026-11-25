<script setup lang="ts">
import type { TimeBucket } from '~/utils/analytics'

/** Tek seri sütun grafiği. Sütuna gelince ya da odaklanınca değer üstte okunur. */
const props = defineProps<{ buckets: TimeBucket[]; label: string }>()

const H = 220
const active = ref<number | null>(null)

const max = computed(() => niceMax(Math.max(0, ...props.buckets.map((b) => b.count))))
const gridLines = computed(() => [0, 0.5, 1].map((f) => ({ y: f, value: max.value * (1 - f) })))
const total = computed(() => props.buckets.reduce((s, b) => s + b.count, 0))
const peak = computed(() => props.buckets.reduce<TimeBucket | null>((p, b) => (b.count > 0 && (!p || b.count > p.count) ? b : p), null))

// Eksen etiketleri sıkışmasın: en fazla 8 etiket.
const stride = computed(() => Math.max(1, Math.ceil(props.buckets.length / 8)))

const readout = computed(() => {
  const i = active.value
  const b = i === null ? null : props.buckets[i]
  if (b) return { title: b.full, value: `${formatNumber(b.count)} tıklama` }
  return {
    title: peak.value ? `En yoğun: ${peak.value.full}` : 'Bu aralıkta tıklama yok',
    value: `Toplam ${formatNumber(total.value)} tıklama`
  }
})
</script>

<template>
  <figure :aria-label="label">
    <figcaption class="flex h-12 flex-col justify-center" aria-live="polite">
      <span class="tnum text-lg font-semibold leading-tight">{{ readout.value }}</span>
      <span class="text-sm text-ink-3">{{ readout.title }}</span>
    </figcaption>

    <div class="relative mt-4 pl-9" :style="{ height: `${H}px` }" @mouseleave="active = null">
      <div
        v-for="g in gridLines"
        :key="g.y"
        class="absolute inset-x-0 flex items-center"
        :style="{ top: `${g.y * 100}%` }"
        aria-hidden="true"
      >
        <span class="tnum w-7 -translate-y-1/2 text-right text-[13px] text-ink-3">{{ formatNumber(g.value) }}</span>
        <span class="ml-2 h-px flex-1" :class="g.y === 1 ? 'bg-line-2' : 'bg-line'" />
      </div>

      <div class="absolute inset-y-0 left-9 right-0 flex">
        <button
          v-for="(b, i) in buckets"
          :key="b.key"
          type="button"
          class="group relative flex h-full flex-1 items-end justify-center px-px outline-none"
          :aria-label="`${b.full}: ${b.count} tıklama`"
          @mouseenter="active = i"
          @focus="active = i"
          @blur="active = null"
        >
          <span
            class="w-full max-w-9 rounded-t-[3px] transition-colors"
            :class="active === null || active === i ? 'bg-brand' : 'bg-brand/45'"
            :style="{ height: `${(b.count / max) * 100}%`, minHeight: b.count > 0 ? '3px' : '0' }"
          />
        </button>
      </div>
    </div>

    <div class="mt-2 flex pl-9 text-[13px] text-ink-3" aria-hidden="true">
      <span v-for="(b, i) in buckets" :key="b.key" class="tnum relative h-5 flex-1">
        <span v-if="i % stride === 0" class="absolute left-1/2 -translate-x-1/2 whitespace-nowrap">{{ b.label }}</span>
      </span>
    </div>
  </figure>
</template>
