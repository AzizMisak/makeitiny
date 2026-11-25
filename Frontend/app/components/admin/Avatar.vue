<script setup lang="ts">
withDefaults(defineProps<{ name?: string; email?: string; large?: boolean }>(), { name: '', email: '', large: false })
</script>

<template>
  <span
    class="cut-corner inline-flex shrink-0 items-center justify-center font-semibold text-brand-ink [--cut-bg:var(--color-brand-soft)]"
    :class="large ? 'h-14 w-14 text-xl' : 'h-9 w-9 text-[13px]'"
    aria-hidden="true"
  >{{ initialsOf(name, email) }}</span>
</template>
