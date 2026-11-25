<script setup lang="ts">
defineProps<{ message: string; loading?: boolean }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <div role="alert" class="flex flex-wrap items-center justify-between gap-4 bg-danger-soft px-5 py-4 text-danger">
    <p class="flex items-center gap-3 font-medium"><UiIcon name="alert" /> {{ message }}</p>
    <UiButton variant="outline" size="sm" icon="refresh" :loading="loading" @click="$emit('retry')">Tekrar dene</UiButton>
  </div>
</template>
