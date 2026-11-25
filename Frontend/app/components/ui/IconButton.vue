<script setup lang="ts">
import { NuxtLink } from '#components'

const props = withDefaults(
  defineProps<{
    icon: string
    label: string
    to?: string
    href?: string
    loading?: boolean
    disabled?: boolean
    tone?: 'normal' | 'mat' | 'danger' | 'ok'
    size?: number
  }>(),
  { tone: 'normal', size: 18 }
)

const tag = computed(() => (props.to ? NuxtLink : props.href ? 'a' : 'button'))
const attrs = computed(() => {
  if (props.to) return { to: props.to }
  if (props.href) return { href: props.href, target: '_blank', rel: 'noopener' }
  return { type: 'button', disabled: props.disabled || props.loading }
})

const tones = {
  normal: 'text-ink-2 hover:bg-sunken hover:text-ink',
  mat: 'text-on-mat hover:bg-white/15',
  danger: 'text-danger hover:bg-danger-soft',
  ok: 'text-ok'
}
</script>

<template>
  <component
    :is="tag"
    v-bind="attrs"
    class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md transition-colors disabled:opacity-45"
    :class="tones[tone]"
    :title="label"
    :aria-label="label"
  >
    <UiSpinner v-if="loading" :size="size - 2" />
    <UiIcon v-else :name="icon" :size="size" />
  </component>
</template>
