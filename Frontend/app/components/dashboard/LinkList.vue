<script setup lang="ts">
import type { LinkItem } from '~/types/link'

defineProps<{ links: LinkItem[] }>()
defineEmits<{ remove: [id: string] }>()
</script>

<template>
  <div class="overflow-hidden rounded-lg border-[1.5px] border-ink bg-card shadow-hard">
    <ul v-if="links.length">
      <DashboardLinkRow v-for="l in links" :key="l.id" :link="l" @remove="$emit('remove', $event)" />
    </ul>
    <p v-else class="p-10 text-center text-ink-3">Henüz link yok. İlk linkini oluştur.</p>
  </div>
</template>
