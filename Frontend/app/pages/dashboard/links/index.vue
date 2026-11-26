<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Linklerim | makeiTiny' })

const { links, currentPage, pageSize, totalLinks, totalPages, loading, initialized, fetchLinks } = useLinks()
const remove = useLinkRemoval()
const { busy: exporting, exportLinks } = useExport()

const exportMenu = computed(() => [
  { label: 'CSV olarak indir', icon: 'download', disabled: !!exporting.value, onSelect: () => exportLinks('csv') },
  { label: 'JSON olarak indir', icon: 'download', disabled: !!exporting.value, onSelect: () => exportLinks('json') }
])

const error = ref('')
const firstLoad = computed(() => !initialized.value && !error.value)

async function load(page = currentPage.value) {
  error.value = ''
  try {
    await fetchLinks(page, pageSize.value)
  } catch (e: any) {
    error.value = e?.message ?? 'Linkler yüklenemedi.'
  }
}

onMounted(() => load())
</script>

<template>
  <div>
    <PanelHeader title="Linklerim">
      <p v-if="initialized" class="tnum text-ink-3">{{ formatNumber(totalLinks) }} link</p>
      <UiMenu v-if="totalLinks > 0" label="Linkleri dışa aktar" :items="exportMenu">
        <template #trigger="{ toggle, open }">
          <UiButton variant="outline" size="sm" icon="download" :loading="!!exporting" loading-label="Hazırlanıyor" :aria-expanded="open" aria-haspopup="menu" @click="toggle">
            Dışa aktar
          </UiButton>
        </template>
      </UiMenu>
    </PanelHeader>

    <UiLoadError v-if="error" :message="error" :loading="loading" class="mb-6" @retry="load()" />

    <ul v-if="firstLoad" class="divide-y divide-line border-y border-line" aria-label="Linkler yükleniyor" aria-busy="true">
      <li v-for="i in 6" :key="i" class="flex items-center justify-between gap-6 py-5">
        <div class="flex-1"><div class="skeleton h-5 w-56" /><div class="skeleton mt-2 h-4 w-3/4" /></div>
        <div class="skeleton h-8 w-20" />
      </li>
    </ul>

    <template v-else-if="links.length">
      <ul
        class="divide-y divide-line border-y border-line transition-opacity"
        :class="loading ? 'pointer-events-none opacity-50' : ''"
        :aria-busy="loading"
      >
        <LinksRow v-for="l in links" :key="l.id" :link="l" @remove="remove" />
      </ul>
      <UiPagination
        :page="currentPage"
        :total-pages="totalPages"
        :total="totalLinks"
        :page-size="pageSize"
        :loading="loading"
        item-label="link"
        @change="load"
      />
    </template>

    <UiEmpty v-else-if="!error" title="Henüz linkin yok" text="İlk linkini kestiğinde burada sıralanır.">
      <UiButton to="/dashboard/new" icon="scissors">İlk linki oluştur</UiButton>
    </UiEmpty>
  </div>
</template>
