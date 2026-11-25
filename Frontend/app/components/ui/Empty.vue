<script setup lang="ts">
withDefaults(defineProps<{ title: string; text?: string; icon?: string }>(), { icon: 'scissors' })
</script>

<template>
  <div class="flex flex-col items-center gap-3 px-6 py-14 text-center">
    <UiIcon :name="icon" :size="30" class="text-line-2" />
    <p class="text-lg font-semibold">{{ title }}</p>
    <p v-if="text" class="max-w-md text-ink-2">{{ text }}</p>
    <div v-if="$slots.default" class="mt-2 flex flex-wrap justify-center gap-2"><slot /></div>
  </div>
</template>
