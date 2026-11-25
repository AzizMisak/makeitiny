<script setup lang="ts">
import type { AdminLinkStatusFilter } from '~/composables/useAdmin'
import type { AdminLinkItem, AdminStatsResponse } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth', roles: ['admin', 'owner'] })
useHead({ title: 'Tüm Linkler — Yönetici — makeiTiny' })

const { public: config } = useRuntimeConfig()
const { fetchAdminStats, fetchAdminLinks, adminUpdateLink } = useAdmin()
const toast = useToast()

const PAGE_SIZE = 15

const stats = ref<AdminStatsResponse | null>(null)
const links = ref<AdminLinkItem[]>([])
const page = ref(1)
const total = ref(0)
const totalPages = ref(1)
const loading = ref(true)

const searchInput = ref('')
const activeSearch = ref('')
const statusFilter = ref<AdminLinkStatusFilter>('all')

const pendingAction = ref<{
  link: AdminLinkItem
  kind: 'toggleActive' | 'toggleBlock'
} | null>(null)
const actionSaving = ref(false)

const statusTabs: { value: AdminLinkStatusFilter; label: string }[] = [
  { value: 'all', label: 'Tümü' },
  { value: 'active', label: 'Aktif' },
  { value: 'inactive', label: 'Pasif' },
  { value: 'blocked', label: 'Engelli / Zararlı' },
  { value: 'anonymous', label: 'Anonim' }
]

function displayShortUrl(alias: string): string {
  return `${config.routerBase}/${alias}`.replace(/^https?:\/\//, '')
}

function isLinkBlocked(item: AdminLinkItem): boolean {
  return item.securityStatus === 'blocked' || item.securityStatus === 'malicious'
}

async function loadLinks(targetPage = page.value) {
  loading.value = true
  try {
    const [linksRes, statsRes] = await Promise.all([
      fetchAdminLinks(targetPage, PAGE_SIZE, activeSearch.value, statusFilter.value),
      fetchAdminStats().catch(() => null)
    ])
    links.value = linksRes.items
    page.value = linksRes.page
    total.value = linksRes.total
    totalPages.value = Math.max(1, linksRes.totalPages)
    if (statsRes) stats.value = statsRes
  } catch (err: any) {
    toast.error(err?.message ?? 'Platform linkleri yüklenemedi.')
  } finally {
    loading.value = false
  }
}

function submitSearch() {
  activeSearch.value = searchInput.value.trim()
  loadLinks(1)
}

function clearSearch() {
  searchInput.value = ''
  activeSearch.value = ''
  loadLinks(1)
}

function selectStatus(status: AdminLinkStatusFilter) {
  if (statusFilter.value === status) return
  statusFilter.value = status
  loadLinks(1)
}

function openQuickAction(link: AdminLinkItem, kind: 'toggleActive' | 'toggleBlock') {
  pendingAction.value = { link, kind }
}

async function confirmQuickAction() {
  const act = pendingAction.value
  if (!act) return
  actionSaving.value = true
  try {
    if (act.kind === 'toggleActive') {
      const nextActive = !act.link.isActive
      await adminUpdateLink(act.link.id, { isActive: nextActive })
      toast.success(nextActive ? 'Link aktifleştirildi.' : 'Link erişime kapatıldı (pasif).')
    } else {
      const blocked = isLinkBlocked(act.link)
      await adminUpdateLink(act.link.id, {
        securityStatus: blocked ? 'safe' : 'blocked',
        isActive: blocked ? true : false
      })
      toast.success(blocked ? 'Linkin engeli kaldırıldı.' : 'Link güvenlik nedeniyle yasaklandı ve kapatıldı.')
    }
    pendingAction.value = null
    await loadLinks(page.value)
  } catch (err: any) {
    toast.error(err?.message ?? 'İşlem başarısız.')
  } finally {
    actionSaving.value = false
  }
}

const modalMeta = computed(() => {
  const act = pendingAction.value
  if (!act) return { title: '', desc: '', confirm: '', variant: 'primary' as const }
  if (act.kind === 'toggleActive') {
    return act.link.isActive
      ? {
          title: 'Linki erişime kapat',
          desc: 'Bu kısa link pasife alınacak ve ziyaretçiler hata ekranına yönlendirilecek.',
          confirm: 'Evet, kapat',
          variant: 'danger' as const
        }
      : {
          title: 'Linki aktifleştir',
          desc: 'Bu kısa link tekrar yönlendirmeye açılacak.',
          confirm: 'Evet, aktifleştir',
          variant: 'primary' as const
        }
  }
  const blocked = isLinkBlocked(act.link)
  return blocked
    ? {
        title: 'Link yasağını kaldır',
        desc: 'Güvenlik durumu "Güvenli" olarak işaretlenecek ve link tekrar aktifleştirilecek.',
        confirm: 'Yasağı kaldır',
        variant: 'primary' as const
      }
    : {
        title: 'Linki yasakla ve kapat',
        desc: 'Bu kısa link güvenlik nedeniyle "Engelli" durumuna alınacak ve erişimi kapatılacak.',
        confirm: 'Evet, yasakla',
        variant: 'danger' as const
      }
})

onMounted(() => loadLinks(1))
</script>

<template>
  <DashboardPageHeader
    title="Tüm linkler"
    description="Platformdaki tüm linkleri incele, kapat/aç veya güvenlik nedeniyle yasakla."
  >
    <template #actions>
      <UiButton variant="secondary" :disabled="loading" @click="loadLinks(page)">
        <UiIcon name="refresh" :size="15" /> Yenile
      </UiButton>
    </template>
  </DashboardPageHeader>

  <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
    <DashboardStatTile
      label="Toplam link"
      :value="stats ? stats.totalLinks.toLocaleString('tr-TR') : '...'"
    />
    <DashboardStatTile
      label="Aktif link"
      :value="stats ? stats.activeLinks.toLocaleString('tr-TR') : '...'"
    />
    <DashboardStatTile
      label="Engelli / Zararlı"
      :value="stats ? stats.blockedLinks.toLocaleString('tr-TR') : '...'"
    />
    <DashboardStatTile
      label="Toplam tıklama"
      :value="stats ? stats.totalClicks.toLocaleString('tr-TR') : '...'"
    />
  </div>

  <div class="mt-8 rounded-lg border-[1.5px] border-ink bg-card p-5 shadow-hard">
    <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
      <form class="flex flex-1 flex-wrap items-center gap-2" @submit.prevent="submitSearch">
        <div class="relative min-w-56 flex-1">
          <UiIcon name="search" :size="15" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-3" />
          <input
            v-model="searchInput"
            type="search"
            placeholder="Alias, hedef URL, başlık veya sahip e-posta ara..."
            class="w-full rounded-sm border-[1.5px] border-ink bg-paper py-2 pl-9 pr-3 text-sm text-ink placeholder:text-ink-3 focus:outline-none focus:ring-2 focus:ring-accent"
          />
        </div>
        <UiButton type="submit" :disabled="loading">Ara</UiButton>
        <UiButton v-if="activeSearch" variant="ghost" @click="clearSearch">Temizle</UiButton>
      </form>

      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="tab in statusTabs"
          :key="tab.value"
          type="button"
          class="cursor-pointer rounded-sm border px-3 py-1.5 text-xs font-semibold transition-colors"
          :class="
            statusFilter === tab.value
              ? 'border-ink bg-ink text-paper'
              : 'border-line bg-paper text-ink-2 hover:border-ink hover:text-ink'
          "
          @click="selectStatus(tab.value)"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>
  </div>

  <div class="mt-6 rounded-lg border-[1.5px] border-ink bg-card shadow-hard">
    <div v-if="loading" class="flex flex-col gap-2.5 p-5">
      <div v-for="i in 5" :key="i" class="h-14 animate-pulse rounded-md bg-paper-2" />
    </div>

    <div v-else-if="links.length === 0" class="p-10 text-center">
      <p class="text-base font-semibold text-ink">Link bulunamadı.</p>
      <p class="mt-1 text-sm text-ink-3">Arama metnini veya durum filtresini değiştirip tekrar dene.</p>
    </div>

    <ul v-else class="divide-y divide-line">
      <li
        v-for="item in links"
        :key="item.id"
        class="flex flex-col gap-3 p-4 transition-colors hover:bg-paper/60 lg:flex-row lg:items-center lg:justify-between sm:p-5"
      >
        <NuxtLink :to="`/dashboard/admin/links/${item.id}`" class="group min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-mono text-sm font-bold text-ink group-hover:text-accent group-hover:underline">
              {{ displayShortUrl(item.alias) }}
            </span>
            <span v-if="item.title" class="truncate text-xs font-semibold text-ink-2">• {{ item.title }}</span>
            <UiBadge v-if="item.isAnonymous" tone="neutral">Anonim</UiBadge>
            <UiBadge :tone="item.isActive ? 'ok' : 'accent'">
              {{ item.isActive ? 'Aktif' : 'Pasif' }}
            </UiBadge>
            <UiBadge :tone="securityStatusMeta(item.securityStatus).tone">
              {{ securityStatusMeta(item.securityStatus).label }}
            </UiBadge>
            <UiBadge v-if="item.locked" tone="neutral"><UiIcon name="lock" :size="11" /> şifreli</UiBadge>
            <UiBadge v-if="item.hasQr" tone="ok"><UiIcon name="qr" :size="11" /> QR</UiBadge>
          </div>

          <p class="mt-1 truncate text-sm text-ink-3" :title="item.target">{{ item.target }}</p>

          <div class="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 font-mono text-[11px] text-ink-3">
            <span>
              Sahip:
              <strong class="text-ink-2">{{ item.isAnonymous ? 'Üyeliksiz (Anonim)' : (item.userName || item.userEmail) }}</strong>
              <template v-if="item.userEmail"> ({{ item.userEmail }})</template>
            </span>
            <span>Tarih: {{ formatDateTr(item.createdAt) }}</span>
            <span v-if="item.expiresAt">Bitiş: {{ formatDateTr(item.expiresAt) }}</span>
            <span v-if="item.maxClicks">Limit: {{ item.clicks }}/{{ item.maxClicks }}</span>
          </div>
        </NuxtLink>

        <div class="flex shrink-0 flex-wrap items-center gap-2">
          <div class="mr-2 text-right">
            <span class="block font-display text-lg font-bold leading-none text-ink">
              {{ item.clicks.toLocaleString('tr-TR') }}
            </span>
            <span class="text-xs text-ink-3">tıklama</span>
          </div>

          <button
            type="button"
            class="cursor-pointer rounded-sm border border-line bg-paper px-2.5 py-1.5 text-xs font-semibold text-ink-2 transition-colors hover:border-ink hover:text-ink"
            @click="openQuickAction(item, 'toggleActive')"
          >
            {{ item.isActive ? 'Kapat' : 'Aç' }}
          </button>

          <button
            type="button"
            class="cursor-pointer rounded-sm border px-2.5 py-1.5 text-xs font-semibold transition-colors"
            :class="
              isLinkBlocked(item)
                ? 'border-ok bg-ok-soft text-ok hover:bg-card'
                : 'border-accent-dark bg-accent-soft text-accent-dark hover:bg-accent hover:text-white'
            "
            @click="openQuickAction(item, 'toggleBlock')"
          >
            {{ isLinkBlocked(item) ? 'Yasağı Kaldır' : 'Yasakla' }}
          </button>

          <NuxtLink
            :to="`/dashboard/admin/links/${item.id}`"
            class="inline-flex items-center gap-1 rounded-sm border border-ink bg-card px-2.5 py-1.5 text-xs font-semibold text-ink hover:bg-paper-2"
          >
            İncele <UiIcon name="arrow" :size="13" />
          </NuxtLink>
        </div>
      </li>
    </ul>
  </div>

  <DashboardPaginationBar
    :page="page"
    :total-pages="totalPages"
    :total="total"
    :page-size="PAGE_SIZE"
    :loading="loading"
    item-label="link"
    @change="loadLinks"
  />

  <UiConfirmModal
    :open="!!pendingAction"
    :title="modalMeta.title"
    :target-label="pendingAction ? `${displayShortUrl(pendingAction.link.alias)} → ${pendingAction.link.target}` : ''"
    :description="modalMeta.desc"
    :confirm-label="modalMeta.confirm"
    :variant="modalMeta.variant"
    :loading="actionSaving"
    @cancel="pendingAction = null"
    @confirm="confirmQuickAction"
  />
</template>
