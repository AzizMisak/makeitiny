<script setup lang="ts">
defineProps<{
  id: string
  label: string
  type?: string
  placeholder?: string
  hint?: string
  prefix?: string
  autocomplete?: string
  disabled?: boolean
}>()

const model = defineModel<string>({ default: '' })
</script>

<template>
  <div class="flex flex-col gap-1.5">
    <label :for="id" class="text-sm font-semibold">{{ label }}</label>
    <div
      class="flex items-center overflow-hidden rounded-md border-[1.5px] border-ink bg-card focus-within:ring-2 focus-within:ring-accent"
      :class="disabled ? 'cursor-not-allowed opacity-60 bg-paper-2' : ''"
    >
      <span v-if="prefix" class="border-r border-line bg-paper-2 px-3 py-2.5 font-mono text-sm text-ink-2">
        {{ prefix }}
      </span>
      <input
        :id="id"
        :value="model"
        :type="type ?? 'text'"
        :placeholder="placeholder"
        :autocomplete="autocomplete"
        :disabled="disabled"
        class="w-full bg-transparent px-3 py-2.5 text-sm outline-none placeholder:text-ink-3 disabled:cursor-not-allowed"
        @input="model = ($event.target as HTMLInputElement).value"
      />
    </div>
    <p v-if="hint" class="text-xs text-ink-3">{{ hint }}</p>
  </div>
</template>
