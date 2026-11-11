<script setup lang="ts">
import type { LinkItem } from '~/types/link'

const props = defineProps<{ link: LinkItem }>()
defineEmits<{ remove: [id: string] }>()

const { public: config } = useRuntimeConfig()

const copied = ref(false)
const fullShortUrl = computed(() => `${config.routerBase}/${props.link.alias}`)
const displayShortUrl = computed(() => fullShortUrl.value.replace(/^https?:\/\//, ''))

async function copy() {
  await navigator.clipboard?.writeText(fullShortUrl.value)
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}
</script>

<template>
  <li class="border-b border-line transition-colors last:border-b-0 hover:bg-paper/60">
    <div class="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5">
      <NuxtLink :to="`/dashboard/links/${link.id}`" class="group min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="font-mono text-sm font-semibold group-hover:text-accent group-hover:underline">
            {{ displayShortUrl }}
          </span>
          <UiBadge v-if="link.locked" tone="accent"><UiIcon name="lock" :size="11" /> şifreli</UiBadge>
        </div>
        <p class="mt-1 truncate text-sm text-ink-3" :title="link.target">{{ link.target }}</p>
      </NuxtLink>

      <div class="flex shrink-0 items-center gap-1">
        <span class="mr-3 text-right">
          <span class="block font-display text-lg font-bold leading-none">{{ link.clicks.toLocaleString('tr-TR') }}</span>
          <span class="text-xs text-ink-3">tıklama</span>
        </span>
        <button
          type="button"
          class="rounded-md p-2 text-ink-2 hover:bg-paper-2 hover:text-ink"
          :title="`${displayShortUrl} adresini kopyala`"
          :aria-label="`${displayShortUrl} kopyala`"
          @click.stop="copy"
        >
          <UiIcon :name="copied ? 'check' : 'copy'" />
        </button>
        <NuxtLink
          :to="`/dashboard/links/${link.id}`"
          class="rounded-md p-2 text-ink-2 hover:bg-paper-2 hover:text-ink"
          title="Link detayı ve düzenleme"
          :aria-label="`${displayShortUrl} detayı`"
        >
          <UiIcon name="sliders" />
        </NuxtLink>
        <button
          type="button"
          class="rounded-md p-2 text-ink-2 hover:bg-accent-soft hover:text-accent-dark"
          title="Linki sil"
          :aria-label="`${displayShortUrl} sil`"
          @click.stop="$emit('remove', link.id)"
        >
          <UiIcon name="trash" />
        </button>
      </div>
    </div>
  </li>
</template>
