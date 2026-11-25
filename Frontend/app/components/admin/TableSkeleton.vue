<script setup lang="ts">
withDefaults(defineProps<{ rows?: number; label: string }>(), { rows: 8 })
</script>

<template>
  <div class="border border-line bg-surface" aria-busy="true" :aria-label="label">
    <div class="h-10 bg-sunken" />
    <div v-for="i in rows" :key="i" class="flex items-center gap-8 border-t border-line px-4 py-4">
      <div class="flex-1"><div class="skeleton h-4 w-48" /><div class="skeleton mt-2 h-3 w-80 max-w-full" /></div>
      <div class="skeleton hidden h-4 w-28 md:block" />
      <div class="skeleton hidden h-4 w-20 md:block" />
      <div class="skeleton h-4 w-14" />
    </div>
  </div>
</template>
