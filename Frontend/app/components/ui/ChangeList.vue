<script setup lang="ts">
import type { ConfirmChange } from '~/composables/useConfirm'

defineProps<{ changes: ConfirmChange[] }>()
</script>

<template>
  <dl class="divide-y divide-line border-y border-line">
    <div v-for="c in changes" :key="c.label" class="grid grid-cols-1 gap-x-4 gap-y-1 py-3 text-sm sm:grid-cols-[8.5rem_1fr]">
      <dt class="text-ink-3">{{ c.label }}</dt>
      <dd class="min-w-0">
        <span class="block break-all text-ink-3 line-through">{{ c.from }}</span>
        <span class="block break-all font-semibold text-ink">{{ c.to }}</span>
      </dd>
    </div>
  </dl>
</template>
