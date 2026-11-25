<script setup lang="ts">
import type { Tone } from '~/utils/admin'

withDefaults(defineProps<{ tone?: Tone; icon?: string }>(), { tone: 'neutral' })

const tones: Record<Tone, string> = {
  neutral: 'bg-sunken text-ink-2',
  brand: 'bg-brand-soft text-brand-ink',
  ok: 'bg-ok-soft text-ok',
  warn: 'bg-warn-soft text-warn',
  danger: 'bg-danger-soft text-danger',
  mark: 'bg-mark text-on-mark'
}
</script>

<template>
  <span
    class="inline-flex h-6 shrink-0 items-center gap-1.5 whitespace-nowrap rounded-[3px] px-2 text-[13px] font-medium"
    :class="tones[tone]"
  >
    <UiIcon v-if="icon" :name="icon" :size="13" />
    <slot />
  </span>
</template>
