<script setup lang="ts">
withDefaults(defineProps<{ tone?: 'neutral' | 'accent' | 'ok' }>(), { tone: 'neutral' })
</script>

<template>
  <span
    class="inline-flex items-center gap-1 rounded-sm px-2 py-0.5 font-mono text-[11px] font-medium"
    :class="{
      'bg-paper-2 text-ink-2': tone === 'neutral',
      'bg-accent-soft text-accent-dark': tone === 'accent',
      'bg-ok-soft text-ok': tone === 'ok'
    }"
  >
    <slot />
  </span>
</template>
