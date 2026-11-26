<script setup lang="ts">
import type { AdminLinkStatusFilter } from '~/composables/useAdmin'
import type { AdminLinkItem } from '~/types/link'

definePageMeta({ layout: 'admin', middleware: 'auth', roles: ['admin', 'owner'] })
useHead({ title: 'Linkler | Yönetim | makeiTiny' })

const PAGE_SIZE = 20
const FILTERS: { value: AdminLinkStatusFilter; label: string }[] = [
  { value: 'all', label: 'Tümü' },
  { value: 'active', label: 'Açık' },
  { value: 'inactive', label: 'Kapalı' },
  { value: 'blocked', label: 'Engelli' },
  { value: 'anonymous', label: 'Üyeliksiz' }
]

const route = useRoute()
const router = useRouter()
const { fetchAdminLinks } = useAdmin()
const { busy: exporting, exportAdminLinks } = useExport()

const links = ref<AdminLinkItem[]>([])
const page = ref(1)
const total = ref(0)
const totalPages = ref(1)
const loading = ref(true)
const loadedOnce = ref(false)
const error = ref('')
const search = ref(String(route.query.ara ?? ''))
const searchKey = ref(0)
const filter = ref<AdminLinkStatusFilter>(
  FILTERS.some((f) => f.value === route.query.durum) ? (route.query.durum as AdminLinkStatusFilter) : 'all'
)

async function load(next = page.value) {
  loading.value = true
  error.value = ''
  try {
    const res = await fetchAdminLinks(next, PAGE_SIZE, search.value, filter.value)
    links.value = res.items
    page.value = res.page
    total.value = res.total
    totalPages.value = Math.max(1, res.totalPages)
    loadedOnce.value = true
  } catch (e: any) {
    error.value = e?.message ?? 'Linkler yüklenemedi.'
  } finally {
    loading.value = false
  }
}

function syncUrl() {
  router.replace({
    query: {
      ...(filter.value !== 'all' ? { durum: filter.value } : {}),
      ...(search.value ? { ara: search.value } : {})
    }
  })
}

function onSearch(value: string) {
  search.value = value
  syncUrl()
  load(1)
}

watch(filter, () => {
  syncUrl()
  load(1)
})

// Üst menüden yeniden tıklanınca adres sorgusu temizlenir; ekrandaki süzgeç de ona uyar.
watch(
  () => route.query,
  (q) => {
    const nextFilter = FILTERS.some((f) => f.value === q.durum) ? (q.durum as AdminLinkStatusFilter) : 'all'
    const nextSearch = String(q.ara ?? '')
    if (nextFilter === filter.value && nextSearch === search.value) return
    search.value = nextSearch
    searchKey.value++
    if (nextFilter !== filter.value) filter.value = nextFilter
    else load(1)
  }
)

onMounted(() => load(1))

const exportMenu = computed(() => [
  { label: 'CSV olarak indir', icon: 'download', disabled: !!exporting.value, onSelect: () => exportAdminLinks('csv', search.value, filter.value) },
  { label: 'JSON olarak indir', icon: 'download', disabled: !!exporting.value, onSelect: () => exportAdminLinks('json', search.value, filter.value) }
])
</script>

<template>
  <div>
    <AdminHeader title="Linkler" :note="loadedOnce ? `${formatNumber(total)} kayıt` : ''" refreshable :loading="loading" @refresh="load()" />

    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <UiSegmented v-model="filter" label="Duruma göre süz" :options="FILTERS" small />
      <div class="flex w-full min-w-0 items-center justify-end gap-2 sm:w-auto">
        <UiMenu v-if="total > 0" label="Linkleri dışa aktar" :items="exportMenu">
          <template #trigger="{ toggle, open }">
            <UiIconButton
              icon="download"
              label="Linkleri dışa aktar"
              :loading="exporting.startsWith('admin-links-')"
              :aria-expanded="open"
              aria-haspopup="menu"
              @click="toggle"
            />
          </template>
        </UiMenu>
        <UiSearchBox
          id="admin-link-search"
          :key="searchKey"
          placeholder="Kısa ad, hedef, başlık veya sahip e-postası"
          :active="search"
          :loading="loading"
          @search="onSearch"
        />
      </div>
    </div>

    <UiLoadError v-if="error" :message="error" :loading="loading" @retry="load()" />
    <AdminTableSkeleton v-else-if="!loadedOnce" :rows="10" label="Linkler yükleniyor" />
    <UiEmpty
      v-else-if="links.length === 0"
      icon="search"
      title="Eşleşen link yok"
      text="Aramayı veya durum süzgecini değiştir."
      class="border border-line bg-surface"
    />
    <template v-else>
      <AdminLinkTable :links="links" :loading="loading" actions caption="Platformdaki linkler" @changed="load()" />
      <UiPagination
        :page="page"
        :total-pages="totalPages"
        :total="total"
        :page-size="PAGE_SIZE"
        :loading="loading"
        item-label="link"
        @change="load"
      />
    </template>
  </div>
</template>
