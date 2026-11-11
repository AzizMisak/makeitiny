<script setup lang="ts">
import type { LinkItem } from '~/types/link'

defineProps<{
  links: LinkItem[]
  loading?: boolean
}>()

defineEmits<{ remove: [id: string] }>()
</script>

<template>
  <div class="overflow-hidden rounded-lg border-[1.5px] border-ink bg-card shadow-hard">
    <div v-if="loading" class="flex items-center justify-center gap-3 p-10 text-ink-2">
      <span class="h-4 w-4 animate-spin rounded-full border-2 border-ink border-t-transparent" />
      <span class="text-sm font-medium">Linkler yükleniyor...</span>
    </div>
    <ul v-else-if="links.length">
      <DashboardLinkRow v-for="l in links" :key="l.id" :link="l" @remove="$emit('remove', $event)" />
    </ul>
    <p v-else class="p-10 text-center text-ink-3">Henüz link yok. İlk linkini oluştur.</p>
  </div>
</template>
