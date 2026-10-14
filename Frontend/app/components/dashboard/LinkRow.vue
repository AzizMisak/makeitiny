<script setup lang="ts">
import type { LinkItem } from '~/types/link'

const props = defineProps<{ link: LinkItem }>()
defineEmits<{ remove: [id: string] }>()

const copied = ref(false)
const shortUrl = computed(() => `mkit.ee/${props.link.alias}`)

async function copy() {
  await navigator.clipboard?.writeText(`https://${shortUrl.value}`)
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}
</script>

<template>
  <li class="flex flex-col gap-3 border-b border-line p-4 last:border-b-0 sm:flex-row sm:items-center sm:justify-between sm:p-5">
    <div class="min-w-0">
      <div class="flex items-center gap-2">
        <span class="font-mono text-sm font-medium">{{ shortUrl }}</span>
        <UiBadge v-if="link.locked" tone="accent"><UiIcon name="lock" :size="11" /> şifreli</UiBadge>
      </div>
      <p class="mt-1 truncate text-sm text-ink-3" :title="link.target">{{ link.target }}</p>
    </div>

    <div class="flex shrink-0 items-center gap-1">
      <span class="mr-3 text-right">
        <span class="block font-display text-lg font-bold leading-none">{{ link.clicks.toLocaleString('tr-TR') }}</span>
        <span class="text-xs text-ink-3">tıklama</span>
      </span>
      <button type="button" class="rounded-md p-2 text-ink-2 hover:bg-paper-2 hover:text-ink" :aria-label="`${shortUrl} kopyala`" @click="copy">
        <UiIcon :name="copied ? 'check' : 'copy'" />
      </button>
      <button type="button" class="rounded-md p-2 text-ink-2 hover:bg-accent-soft hover:text-accent-dark" :aria-label="`${shortUrl} sil`" @click="$emit('remove', link.id)">
        <UiIcon name="trash" />
      </button>
    </div>
  </li>
</template>
