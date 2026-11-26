<script setup lang="ts">
import type { AdminLinkItem, LinkLogItem } from '~/types/link'

definePageMeta({ layout: 'admin', middleware: 'auth', roles: ['admin', 'owner'] })

const route = useRoute()
const { fetchAdminLinkById, fetchAdminLinkLogs } = useAdmin()
const { prefix, full } = useShortUrl()

const id = computed(() => String(route.params.id ?? ''))
const link = ref<AdminLinkItem | null>(null)
const loading = ref(true)
const tab = ref<'ziyaretler' | 'duzenle'>('ziyaretler')
const dirty = ref(false)

useUnsavedGuard(() => dirty.value)
useHead({ title: () => `${link.value?.alias ?? 'Link'} | Yönetim | makeiTiny` })

const error = ref('')

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
    const first = await fetchAdminLinkLogs(id.value, 1, 100)
    const pages = Math.min(first.totalPages ?? 1, MAX_ANALYSIS_PAGES)
    const rest = await Promise.all(Array.from({ length: pages - 1 }, (_, i) => fetchAdminLinkLogs(id.value, i + 2, 100)))
    logs.value = [first, ...rest].flatMap((r) => r.items ?? [])
    logsTotal.value = first.total ?? logs.value.length
  } catch (e: any) {
    logsError.value = e?.message ?? 'Analiz verisi yüklenemedi.'
  } finally {
    logsLoading.value = false
  }
}

async function refreshVisitData() {
  try {
    const fresh = await fetchAdminLinkById(id.value)
    if (fresh) link.value = fresh
  } catch {
    // Sessiz geçilir; analiz kendi hatasını gösterir.
  }
  loadAnalysis()
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    link.value = await fetchAdminLinkById(id.value)
    if (link.value) loadAnalysis()
  } catch (e: any) {
    error.value = e?.message ?? 'Link bilgileri yüklenemedi.'
  } finally {
    loading.value = false
  }
}

onMounted(load)

watch(tab, (next, prev) => {
  if (next === 'ziyaretler' && prev && link.value) {
    refreshVisitData()
  }
})

const parts = computed(() => splitUrl(link.value?.target ?? ''))

const facts = computed(() => {
  const l = link.value
  if (!l) return []
  return [
    { label: 'Tıklama', value: l.maxClicks ? `${formatNumber(l.clicks)} / ${formatNumber(l.maxClicks)}` : formatNumber(l.clicks) },
    { label: 'Erişim', value: l.locked ? 'Parolalı' : l.waitEnabled ? `${l.waitSeconds} sn bekletmeli` : 'Herkese açık' },
    { label: 'Oluşturulma', value: formatDateTr(l.createdAt) },
    { label: 'Bitiş', value: l.expiresAt ? formatDateTr(l.expiresAt) : 'Süresiz' },
    { label: 'Kimlik', value: l.id }
  ]
})

const tabs = computed(() => [
  { value: 'ziyaretler' as const, label: 'Ziyaretler', icon: 'eye' },
  { value: 'duzenle' as const, label: 'Linki düzenle', icon: 'edit', dot: dirty.value }
])
</script>

<template>
  <div>
    <NuxtLink to="/dashboard/admin/links" class="mb-5 inline-flex items-center gap-2 text-sm font-medium text-ink-2 hover:text-ink">
      <UiIcon name="back" :size="16" /> Linkler
    </NuxtLink>

    <div v-if="loading" class="grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]" aria-busy="true" aria-label="Link yükleniyor">
      <div><div class="skeleton h-8 w-72" /><div class="skeleton mt-5 h-28" /><div class="skeleton mt-8 h-72" /></div>
      <div class="skeleton h-[30rem]" />
    </div>

    <UiLoadError v-else-if="error" :message="error" @retry="load" />

    <UiEmpty v-else-if="!link" icon="search" title="Link bulunamadı" text="Silinmiş olabilir ya da adres yanlış." class="border border-line bg-surface">
      <UiButton variant="outline" icon="refresh" @click="load">Tekrar dene</UiButton>
      <UiButton to="/dashboard/admin/links" variant="quiet">Linklere dön</UiButton>
    </UiEmpty>

    <div v-else class="grid grid-cols-1 items-start gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
      <div class="min-w-0">
        <header class="flex flex-wrap items-center justify-between gap-3">
          <h1 class="min-w-0 break-all font-mono text-2xl font-semibold">
            <span class="font-normal text-ink-3">{{ prefix }}</span>{{ link.alias }}
          </h1>
          <UiButton :href="full(link.alias)" target="_blank" rel="noopener" variant="outline" size="sm" icon="external">Kısa linki aç</UiButton>
        </header>

        <!-- Kararın kanıtı: hedef adres, alan adı öne çıkarılmış. -->
        <section class="mt-5 border-l-4 border-mark bg-surface px-5 py-4">
          <p class="text-[13px] text-ink-3">Hedef</p>
          <p class="mt-1 break-all font-mono text-sm text-ink-3">
            {{ parts.protocol }}<strong class="text-lg text-ink">{{ parts.host }}</strong>{{ parts.rest }}
          </p>
          <p v-if="link.title" class="mt-2 text-sm text-ink-2">Başlık: {{ link.title }}</p>
        </section>

        <dl class="mt-6 flex flex-wrap gap-x-10 gap-y-4">
          <div>
            <dt class="text-[13px] text-ink-3">Sahibi</dt>
            <dd class="mt-0.5 font-semibold">
              <NuxtLink
                v-if="link.userId && link.userId !== '0'"
                :to="`/dashboard/admin/users/${link.userId}`"
                class="text-brand-ink underline underline-offset-4"
                :title="link.userEmail"
              >{{ link.userName || link.userEmail }}</NuxtLink>
              <template v-else>Üyeliksiz</template>
            </dd>
          </div>
          <div v-for="f in facts" :key="f.label">
            <dt class="text-[13px] text-ink-3">{{ f.label }}</dt>
            <dd class="tnum mt-0.5 font-semibold">{{ f.value }}</dd>
          </div>
        </dl>

        <UiTabs v-model="tab" :tabs="tabs" label="Link bölümleri" class="mt-8" />
        <div :id="`panel-${tab}`" role="tabpanel" :aria-labelledby="`tab-${tab}`" class="pt-6">
          <div v-show="tab === 'ziyaretler'">
            <LinksVisits :fetch-page="(p: number, l: number) => fetchAdminLinkLogs(id, p, l)" @refreshed="refreshVisitData" />

            <section class="mt-10 border-t border-line pt-8" aria-labelledby="admin-analysis-title">
              <h2 id="admin-analysis-title" class="mb-4 text-lg font-semibold">Analiz</h2>
              <LinksAnalysis
                :logs="logs"
                :total="logsTotal"
                :loading="logsLoading"
                :error="logsError"
                window-label="Tüm geçmiş"
                @retry="loadAnalysis"
              />
            </section>
          </div>
          <AdminLinkEditor
            v-show="tab === 'duzenle'"
            v-model="link"
            @dirty="dirty = $event"
            @removed="navigateTo('/dashboard/admin/links')"
          />
        </div>
      </div>

      <AdminDecisionPanel v-model="link" class="lg:sticky lg:top-24" />
    </div>
  </div>
</template>
