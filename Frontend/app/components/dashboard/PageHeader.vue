<script setup lang="ts">
defineProps<{ title: string; description?: string }>()
</script>

<template>
  <div class="mb-8 flex flex-wrap items-end justify-between gap-4">
    <div>
      <h1 class="text-3xl font-extrabold md:text-4xl">{{ title }}</h1>
      <p v-if="description" class="mt-2 text-ink-2">{{ description }}</p>
    </div>
    <slot name="actions" />
  </div>
</template>
