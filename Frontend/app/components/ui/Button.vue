<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    to?: string
    type?: 'button' | 'submit'
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
    block?: boolean
  }>(),
  { type: 'button', variant: 'primary' }
)

const base =
  'inline-flex items-center justify-center gap-2 rounded-md border-[1.5px] px-4 py-2.5 text-sm font-semibold transition-all duration-100 cursor-pointer'

const variants = {
  primary:
    'border-ink bg-accent text-white shadow-hard hover:-translate-x-px hover:-translate-y-px hover:shadow-hard-lg active:translate-x-[3px] active:translate-y-[3px] active:shadow-none',
  secondary:
    'border-ink bg-card text-ink shadow-hard hover:-translate-x-px hover:-translate-y-px hover:shadow-hard-lg active:translate-x-[3px] active:translate-y-[3px] active:shadow-none',
  ghost: 'border-transparent text-ink-2 hover:bg-paper-2 hover:text-ink',
  danger: 'border-transparent text-accent-dark hover:bg-accent-soft'
}

const classes = computed(() => [base, variants[props.variant], props.block && 'w-full'])
</script>

<template>
  <NuxtLink v-if="to" :to="to" :class="classes">
    <slot />
  </NuxtLink>
  <button v-else :type="type" :class="classes">
    <slot />
  </button>
</template>
