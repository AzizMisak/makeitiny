<script setup lang="ts">
/** Şerit metre. Kota gibi üst sınırı olan değerleri gerçek ölçekte gösterir. */
const props = defineProps<{ used: number; max: number; label: string }>()

const unlimited = computed(() => props.max < 0)
const percent = computed(() => {
  if (unlimited.value) return 0
  if (props.max === 0) return 100
  return Math.min(100, (props.used / props.max) * 100)
})
const tone = computed(() => (percent.value >= 100 ? 'bg-danger' : percent.value >= 85 ? 'bg-warn' : 'bg-brand'))

const ticks = computed(() => {
  if (unlimited.value) return []
  if (props.max === 0) return [0]
  if (props.max < 4) return [0, props.max]
  return Array.from({ length: 5 }, (_, i) => Math.round((props.max / 4) * i))
})
</script>

<template>
  <div>
    <div
      class="relative h-9 overflow-hidden border-b-2 border-ink bg-sunken"
      role="meter"
      :aria-label="label"
      :aria-valuenow="used"
      aria-valuemin="0"
      :aria-valuemax="unlimited ? undefined : max"
      :aria-valuetext="unlimited ? `${used}, sınırsız` : `${used} / ${max}`"
    >
      <div v-if="!unlimited" class="absolute inset-y-0 left-0 transition-[width] duration-500" :class="tone" :style="{ width: `${percent}%` }" />
      <div v-else class="absolute inset-0 bg-brand-soft" />
      <svg class="absolute inset-0 h-full w-full" aria-hidden="true">
        <line
          v-for="i in 41"
          :key="i"
          :x1="`${((i - 1) / 40) * 100}%`"
          :x2="`${((i - 1) / 40) * 100}%`"
          :y1="(i - 1) % 10 === 0 ? 14 : (i - 1) % 5 === 0 ? 21 : 27"
          y2="36"
          stroke="var(--color-ink)"
          stroke-opacity="0.45"
          stroke-width="1"
        />
      </svg>
    </div>
    <div class="tnum mt-1.5 flex justify-between text-[13px] text-ink-3">
      <template v-if="unlimited">
        <span>0</span>
        <span>sınırsız</span>
      </template>
      <span v-for="(t, i) in ticks" v-else :key="i">{{ formatNumber(t) }}</span>
    </div>
  </div>
</template>
