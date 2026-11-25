<script setup lang="ts">
import type { Tone } from '~/utils/admin'

/** Nokta ve yazıyla durum. Satır içinde, kutusuz. */
withDefaults(defineProps<{ tone?: Tone; label: string }>(), { tone: 'neutral' })

const dots: Record<Tone, string> = {
  neutral: 'border-2 border-line-2',
  brand: 'bg-brand',
  ok: 'bg-ok',
  warn: 'bg-warn',
  danger: 'bg-danger',
  mark: 'bg-mark'
}
</script>

<template>
  <span class="inline-flex items-center gap-2 whitespace-nowrap text-sm" :class="tone === 'danger' ? 'text-danger' : 'text-ink-2'">
    <span class="h-2.5 w-2.5 shrink-0" :class="[dots[tone], tone === 'danger' ? 'rounded-[2px]' : 'rounded-full']" />
    {{ label }}
  </span>
</template>
