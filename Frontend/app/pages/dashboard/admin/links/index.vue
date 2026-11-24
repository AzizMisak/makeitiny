<script setup lang="ts">
import type { AdminLinkItem, AdminStatsResponse } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth', roles: ['admin', 'owner'] })
useHead({ title: 'Tüm Linkler — Yönetici — makeiTiny' })

type StatusFilter = 'all' | 'active' | 'inactive' | 'blocked'

const { public: config } = useRuntimeConfig()
const { fetchAdminStats, fetchAdminLinks } = useAdmin()
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
const statusFilter = ref<StatusFilter>('all')

const statusTabs: { value: StatusFilter; label: string }[] = [
  { value: 'all', label: 'Tümü' },
  { value: 'active', label: 'Aktif' },
  { value: 'inactive', label: 'Pasif' },
  { value: 'blocked', label: 'Engelli / Zararlı' }
]

function displayShortUrl(alias: string): string {
  return `${config.routerBase}/${alias}`.replace(/^https?:\/\//, '')
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

function selectStatus(status: StatusFilter) {
  if (statusFilter.value === status) return
  statusFilter.value = status
  loadLinks(1)
}

onMounted(() => loadLinks(1))
</script>

<template>
  <DashboardPageHeader
    title="Tüm linkler"
    description="İncelemek veya müdahale etmek istediğin linkin üzerine tıkla."
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
      <li v-for="item in links" :key="item.id">
        <NuxtLink
          :to="`/dashboard/admin/links/${item.id}`"
          class="group flex flex-col gap-3 p-4 transition-colors hover:bg-paper/60 sm:flex-row sm:items-center sm:justify-between sm:p-5"
        >
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <span class="font-mono text-sm font-bold text-ink group-hover:text-accent group-hover:underline">
                {{ displayShortUrl(item.alias) }}
              </span>
              <span v-if="item.title" class="truncate text-xs font-semibold text-ink-2">• {{ item.title }}</span>
              <UiBadge :tone="item.isActive ? 'ok' : 'accent'">
                {{ item.isActive ? 'Aktif' : 'Pasif' }}
              </UiBadge>
              <UiBadge :tone="securityStatusMeta(item.securityStatus).tone">
                {{ securityStatusMeta(item.securityStatus).label }}
              </UiBadge>
              <UiBadge v-if="item.locked" tone="neutral"><UiIcon name="lock" :size="11" /> şifreli</UiBadge>
              <UiBadge v-if="item.adEnabled" tone="neutral">reklamlı</UiBadge>
            </div>

            <p class="mt-1 truncate text-sm text-ink-3" :title="item.target">{{ item.target }}</p>

            <div class="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 font-mono text-[11px] text-ink-3">
              <span>Sahip: <strong class="text-ink-2">{{ item.userName || item.userEmail }}</strong> ({{ item.userEmail }})</span>
              <span>Tarih: {{ formatDateTr(item.createdAt) }}</span>
              <span v-if="item.expiresAt">Bitiş: {{ formatDateTr(item.expiresAt) }}</span>
              <span v-if="item.maxClicks">Limit: {{ item.clicks }}/{{ item.maxClicks }}</span>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-4">
            <div class="text-right">
              <span class="block font-display text-lg font-bold leading-none text-ink">
                {{ item.clicks.toLocaleString('tr-TR') }}
              </span>
              <span class="text-xs text-ink-3">tıklama</span>
            </div>
            <span class="inline-flex items-center gap-1 rounded-sm border border-line bg-paper px-2.5 py-1 text-xs font-semibold text-ink-2 transition-colors group-hover:border-ink group-hover:text-ink">
              İncele <UiIcon name="arrow" :size="13" />
            </span>
          </div>
        </NuxtLink>
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
</template>
