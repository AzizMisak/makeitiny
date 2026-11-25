<script setup lang="ts" generic="T extends string | number">
const props = withDefaults(
  defineProps<{
    label: string
    options: { value: T; label: string; icon?: string; disabled?: boolean }[]
    block?: boolean
    small?: boolean
    tone?: 'normal' | 'mat'
  }>(),
  { tone: 'normal' }
)

const model = defineModel<T>({ required: true })
const root = ref<HTMLElement | null>(null)

function move(step: number) {
  const enabled = props.options.filter((o) => !o.disabled)
  const i = enabled.findIndex((o) => o.value === model.value)
  const next = enabled[(i + step + enabled.length) % enabled.length]
  if (!next) return
  model.value = next.value
  nextTick(() => root.value?.querySelector<HTMLElement>('[aria-checked="true"]')?.focus())
}

function stateClass(active: boolean): string {
  if (props.tone === 'mat') return active ? 'bg-white text-[#10211b]' : 'text-on-mat hover:bg-white/10'
  return active ? 'bg-surface text-ink shadow-[0_1px_2px_rgb(0_0_0/0.14)]' : 'text-ink-2 hover:text-ink'
}
</script>

<template>
  <div
    ref="root"
    role="radiogroup"
    :aria-label="label"
    class="max-w-full overflow-x-auto rounded-md p-1 [scrollbar-width:none]"
    :class="[block ? 'flex w-full' : 'inline-flex', tone === 'mat' ? 'bg-black/25' : 'bg-sunken']"
    @keydown.right.prevent="move(1)"
    @keydown.left.prevent="move(-1)"
  >
    <button
      v-for="o in options"
      :key="o.value"
      type="button"
      role="radio"
      :aria-checked="model === o.value"
      :tabindex="model === o.value ? 0 : -1"
      :disabled="o.disabled"
      class="inline-flex shrink-0 items-center justify-center gap-1.5 whitespace-nowrap rounded-[4px] sm:gap-2 font-medium transition-colors disabled:opacity-45"
      :class="[small ? 'min-h-8 px-3 text-[13px]' : 'min-h-10 px-3 text-sm sm:px-4', block ? 'flex-1' : '', stateClass(model === o.value)]"
      @click="model = o.value"
    >
      <UiIcon v-if="o.icon" :name="o.icon" :size="16" />
      {{ o.label }}
    </button>
  </div>
</template>
