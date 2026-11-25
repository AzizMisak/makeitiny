<script setup lang="ts">
withDefaults(defineProps<{ to?: string; tone?: 'ink' | 'mat' }>(), { to: '/', tone: 'ink' })

// Kelime soldan sağa daralır: logo, ürünün yaptığı işi yapar.
const letters = [
  { c: 'm', w: 132 },
  { c: 'a', w: 124 },
  { c: 'k', w: 116 },
  { c: 'e', w: 108 },
  { c: 'i', w: 100 },
  { c: 'T', w: 92 },
  { c: 'i', w: 84 },
  { c: 'n', w: 77 },
  { c: 'y', w: 70 }
]
</script>

<template>
  <NuxtLink :to="to" class="inline-flex items-baseline" aria-label="makeiTiny">
    <span
      v-for="(l, i) in letters"
      :key="i"
      aria-hidden="true"
      class="font-display text-[23px] font-extrabold leading-none"
      :class="tone === 'mat' ? (i === 4 ? 'text-mark' : 'text-on-mat') : i === 4 ? 'text-brand' : 'text-ink'"
      :style="{ fontStretch: `${l.w}%` }"
    >{{ l.c }}</span>
  </NuxtLink>
</template>
