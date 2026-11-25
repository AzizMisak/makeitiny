<script setup lang="ts">
defineProps<{ title: string; note?: string; loading?: boolean; refreshable?: boolean }>()
defineEmits<{ refresh: [] }>()
</script>

<template>
  <header class="mb-6 flex flex-wrap items-center justify-between gap-3">
    <div class="flex min-w-0 flex-wrap items-baseline gap-x-4 gap-y-1">
      <h1 class="display text-[1.75rem]">{{ title }}</h1>
      <p v-if="note" class="tnum text-ink-3">{{ note }}</p>
    </div>
    <div class="flex items-center gap-2">
      <slot />
      <UiIconButton v-if="refreshable" icon="refresh" label="Verileri yenile" :loading="loading" @click="$emit('refresh')" />
    </div>
  </header>
</template>
