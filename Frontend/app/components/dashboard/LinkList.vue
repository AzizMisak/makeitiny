<script setup lang="ts">
import type { LinkItem } from '~/types/link'

const props = defineProps<{
  links: LinkItem[]
  loading?: boolean
}>()

const emit = defineEmits<{ remove: [id: string] }>()

const { public: config } = useRuntimeConfig()
const targetToDelete = ref<LinkItem | null>(null)

function requestRemove(id: string) {
  const found = props.links.find((l) => l.id === id) ?? null
  targetToDelete.value = found
}

function confirmRemove() {
  if (!targetToDelete.value) return
  emit('remove', targetToDelete.value.id)
  targetToDelete.value = null
}

const deleteLabel = computed(() => {
  if (!targetToDelete.value) return ''
  return `${config.routerBase}/${targetToDelete.value.alias}`.replace(/^https?:\/\//, '')
})
</script>

<template>
  <div class="overflow-hidden rounded-lg border-[1.5px] border-ink bg-card shadow-hard">
    <div v-if="loading" class="flex items-center justify-center gap-3 p-10 text-ink-2">
      <span class="h-4 w-4 animate-spin rounded-full border-2 border-ink border-t-transparent" />
      <span class="text-sm font-medium">Linkler yükleniyor...</span>
    </div>
    <ul v-else-if="links.length">
      <DashboardLinkRow v-for="l in links" :key="l.id" :link="l" @remove="requestRemove" />
    </ul>
    <p v-else class="p-10 text-center text-ink-3">Henüz link yok. İlk linkini oluştur.</p>

    <UiConfirmModal
      :open="!!targetToDelete"
      title="Linki kalıcı olarak sil"
      :target-label="deleteLabel"
      description="Bu kısa linki sildiğinizde artık yönlendirme yapmaz ve tıklama istatistiklerine erişilemez. Emin misiniz?"
      confirm-label="Evet, linki sil"
      variant="danger"
      @cancel="targetToDelete = null"
      @confirm="confirmRemove"
    />
  </div>
</template>
