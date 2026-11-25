<script setup lang="ts">
import type { MenuItem } from '~/components/ui/Menu.vue'
import type { LinkItem, LinkLogItem } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth' })

type Tab = 'analiz' | 'ziyaretler' | 'qr' | 'ayarlar'

const route = useRoute()
const router = useRouter()
const { user, isAdmin } = useAuth()
const { getLink, getLinkLogs } = useLinks()
const { prefix, full } = useShortUrl()
const { copied, copy } = useCopy()

const id = computed(() => String(route.params.id ?? ''))
const link = ref<LinkItem | null>(null)
const loading = ref(true)
const error = ref('')

const TABS: Tab[] = ['analiz', 'ziyaretler', 'qr', 'ayarlar']
const tab = computed<Tab>({
  get: () => (TABS.includes(route.query.tab as Tab) ? (route.query.tab as Tab) : 'analiz'),
  set: (value) => router.replace({ query: value === 'analiz' ? {} : { tab: value } })
})

const dirty = ref(false)
useUnsavedGuard(() => dirty.value)

const windowLabel = computed(() => (isAdmin.value ? 'Tüm geçmiş' : `Son ${user.value?.plan?.analyticsDays ?? 7} gün`))

// Analiz, tek sayfadaki kayıtlardan değil en yeni 500 ziyaretten hesaplanır.
const MAX_ANALYSIS_PAGES = 5
const logs = ref<LinkLogItem[]>([])
const logsTotal = ref(0)
const logsLoading = ref(true)
const logsError = ref('')

async function loadAnalysis() {
  logsLoading.value = true
  logsError.value = ''
  try {
    const first = await getLinkLogs(id.value, 1, 100)
    const pages = Math.min(first.totalPages ?? 1, MAX_ANALYSIS_PAGES)
    const rest = await Promise.all(Array.from({ length: pages - 1 }, (_, i) => getLinkLogs(id.value, i + 2, 100)))
    logs.value = [first, ...rest].flatMap((r) => r.items ?? [])
    logsTotal.value = first.total ?? logs.value.length
  } catch (e: any) {
    logsError.value = e?.message ?? 'Analiz verisi yüklenemedi.'
  } finally {
    logsLoading.value = false
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    link.value = await getLink(id.value)
    loadAnalysis()
  } catch (e: any) {
    error.value = e?.message ?? 'Link bilgileri yüklenemedi.'
  } finally {
    loading.value = false
  }
}

onMounted(load)

useHead({ title: () => `${link.value?.alias ?? 'Link'} | makeiTiny` })

const url = computed(() => (link.value ? full(link.value.alias) : ''))
const state = computed(() => (link.value ? linkState(link.value) : null))

const facts = computed(() => {
  const l = link.value
  if (!l) return []
  return [
    { label: 'Tıklama', value: l.maxClicks ? `${formatNumber(l.clicks)} / ${formatNumber(l.maxClicks)}` : formatNumber(l.clicks) },
    { label: 'Erişim', value: l.locked ? 'Parolalı' : l.waitEnabled ? `${l.waitSeconds} sn bekletmeli` : 'Herkese açık' },
    { label: 'Oluşturulma', value: formatDateTr(l.createdAt) },
    { label: 'Bitiş', value: l.expiresAt ? formatDateTr(l.expiresAt) : 'Süresiz' },
    ...(l.adEnabled ? [{ label: 'Reklam', value: 'Açık' }] : [])
  ]
})

const menu = computed<MenuItem[]>(() => [
  { label: 'Yeni sekmede aç', icon: 'external', href: url.value },
  ...(isAdmin.value ? [{ label: 'Yönetim panelinde incele', icon: 'shield', to: `/dashboard/admin/links/${id.value}` }] : [])
])

const tabs = computed(() => [
  { value: 'analiz' as const, label: 'Analiz', icon: 'activity' },
  { value: 'ziyaretler' as const, label: 'Ziyaretler', icon: 'eye' },
  { value: 'qr' as const, label: 'QR kodu', icon: 'qr' },
  { value: 'ayarlar' as const, label: 'Ayarlar', icon: 'settings', dot: dirty.value }
])
</script>

<template>
  <div>
    <NuxtLink to="/dashboard/links" class="mb-6 inline-flex items-center gap-2 text-sm font-medium text-ink-2 hover:text-ink">
      <UiIcon name="back" :size="16" /> Linklerim
    </NuxtLink>

    <div v-if="loading" aria-busy="true" aria-label="Link yükleniyor">
      <div class="skeleton h-14 w-3/4" />
      <div class="skeleton mt-4 h-5 w-1/2" />
      <div class="skeleton mt-10 h-12" />
      <div class="skeleton mt-8 h-64" />
    </div>

    <UiLoadError v-else-if="error" :message="error" @retry="load" />

    <template v-else-if="link && state">
      <header>
        <p v-if="link.title" class="mb-1 font-medium text-ink-2">{{ link.title }}</p>
        <div class="flex flex-wrap items-center justify-between gap-4">
          <h1 class="min-w-0 break-all font-mono text-[clamp(1.4rem,4.5vw,2.25rem)] font-semibold leading-tight">
            <span class="font-normal text-ink-3">{{ prefix }}</span>{{ link.alias }}
          </h1>
          <div class="flex items-center gap-1">
            <UiButton :icon="copied ? 'check' : 'copy'" class="min-w-36" @click="copy(url)">{{ copied ? 'Kopyalandı' : 'Kopyala' }}</UiButton>
            <UiMenu label="Diğer işlemler" :items="menu" />
          </div>
        </div>
        <p class="mt-2 flex min-w-0 items-center gap-2 text-ink-2">
          <UiIcon name="right" :size="15" class="text-ink-3" />
          <span class="truncate" :title="link.target">{{ link.target }}</span>
        </p>

        <dl class="mt-6 flex flex-wrap gap-x-10 gap-y-4">
          <div>
            <dt class="text-[13px] text-ink-3">Durum</dt>
            <dd class="mt-0.5" :title="state.note"><UiStatus :tone="state.tone" :label="state.label" class="!text-base !font-semibold" /></dd>
          </div>
          <div v-for="f in facts" :key="f.label">
            <dt class="text-[13px] text-ink-3">{{ f.label }}</dt>
            <dd class="tnum mt-0.5 font-semibold">{{ f.value }}</dd>
          </div>
        </dl>
      </header>

      <UiTabs v-model="tab" :tabs="tabs" label="Link bölümleri" class="mt-8" />

      <div :id="`panel-${tab}`" role="tabpanel" :aria-labelledby="`tab-${tab}`" class="pb-16 pt-8">
        <LinksAnalysis
          v-if="tab === 'analiz'"
          :logs="logs"
          :total="logsTotal"
          :loading="logsLoading"
          :error="logsError"
          :window-label="windowLabel"
          @retry="loadAnalysis"
        />
        <LinksVisits
          v-else-if="tab === 'ziyaretler'"
          :fetch-page="(p: number, l: number) => getLinkLogs(id, p, l)"
          :window-label="windowLabel"
        />
        <LinksQr v-else-if="tab === 'qr'" v-model="link" />
        <!-- Sekme değişince taslak kaybolmasın diye form gizlenir, kaldırılmaz. -->
        <div v-show="tab === 'ayarlar'">
          <LinksSettings v-model="link" @dirty="dirty = $event" @removed="navigateTo('/dashboard/links')" />
        </div>
      </div>
    </template>
  </div>
</template>
