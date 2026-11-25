<script setup lang="ts" generic="T extends string | number">
defineProps<{
  id: string
  label?: string
  options: { value: T; label: string }[]
  disabled?: boolean
}>()

const model = defineModel<T>({ required: true })
</script>

<template>
  <div class="flex min-w-0 flex-col">
    <label v-if="label" :for="id" class="mb-1.5 text-sm font-semibold text-ink">{{ label }}</label>
    <div class="relative">
      <select
        :id="id"
        v-model="model"
        :disabled="disabled"
        class="min-h-11 w-full appearance-none rounded-md border border-line-2 bg-surface py-2 pl-3.5 pr-10 text-[15px] outline-none focus:border-brand focus:ring-1 focus:ring-brand disabled:cursor-not-allowed disabled:border-dashed disabled:bg-transparent disabled:text-ink-3"
      >
        <option v-for="o in options" :key="o.value" :value="o.value">{{ o.label }}</option>
      </select>
      <UiIcon name="down" :size="16" class="pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 text-ink-3" />
    </div>
  </div>
</template>
