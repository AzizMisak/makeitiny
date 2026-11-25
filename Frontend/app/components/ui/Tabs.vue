<script setup lang="ts" generic="T extends string">
defineProps<{ label: string; tabs: { value: T; label: string; icon?: string; count?: number | null; dot?: boolean }[] }>()
const model = defineModel<T>({ required: true })
</script>

<template>
  <div role="tablist" :aria-label="label" class="flex gap-6 [scrollbar-width:none] overflow-x-auto overflow-y-hidden border-b border-line">
    <button
      v-for="t in tabs"
      :id="`tab-${t.value}`"
      :key="t.value"
      type="button"
      role="tab"
      :aria-selected="model === t.value"
      :aria-controls="`panel-${t.value}`"
      class="inline-flex min-h-12 shrink-0 items-center gap-2 border-b-[3px] text-[15px] font-medium transition-colors"
      :class="model === t.value ? 'border-brand text-ink' : 'border-transparent text-ink-3 hover:border-line-2 hover:text-ink'"
      @click="model = t.value"
    >
      <UiIcon v-if="t.icon" :name="t.icon" :size="17" />
      {{ t.label }}
      <span v-if="t.count != null" class="tnum text-sm text-ink-3">{{ formatNumber(t.count) }}</span>
      <span v-if="t.dot" class="h-2 w-2 rounded-full bg-mark" title="Kaydedilmemiş değişiklik var" />
    </button>
  </div>
</template>
