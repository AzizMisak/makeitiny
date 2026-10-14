<script setup lang="ts">
defineProps<{ label: string; value: string | number; note?: string }>()
</script>

<template>
  <div class="rounded-lg border-[1.5px] border-ink bg-card p-5 shadow-hard">
    <p class="text-sm text-ink-2">{{ label }}</p>
    <p class="mt-2 font-display text-4xl font-extrabold">{{ value }}</p>
    <p v-if="note" class="mt-1 text-xs text-ink-3">{{ note }}</p>
  </div>
</template>
