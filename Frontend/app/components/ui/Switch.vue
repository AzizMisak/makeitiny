<script setup lang="ts">
defineProps<{ id?: string; label: string; disabled?: boolean; loading?: boolean }>()
const model = defineModel<boolean>({ default: false })
</script>

<template>
  <button
    :id="id"
    type="button"
    role="switch"
    :aria-checked="model"
    :aria-label="label"
    :disabled="disabled || loading"
    class="relative inline-flex h-7 w-12 shrink-0 items-center rounded-full border transition-colors duration-150 disabled:opacity-50"
    :class="model ? 'border-brand bg-brand' : 'border-line-2 bg-sunken'"
    @click="model = !model"
  >
    <span
      class="inline-flex h-5 w-5 items-center justify-center rounded-full transition-transform duration-150"
      :class="model ? 'translate-x-[23px] bg-on-brand text-brand' : 'translate-x-[3px] bg-line-2 text-surface'"
    >
      <UiSpinner v-if="loading" :size="11" />
      <UiIcon v-else :name="model ? 'check' : 'minus'" :size="12" />
    </span>
  </button>
</template>
