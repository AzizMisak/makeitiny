<script setup lang="ts">
import type { AdminLinkItem, AdminStatsResponse } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth', roles: ['admin', 'owner'] })
useHead({ title: 'Yönetim Özeti — makeiTiny' })

const { public: config } = useRuntimeConfig()
const { fetchAdminStats, fetchAdminLinks } = useAdmin()
const toast = useToast()

const stats = ref<AdminStatsResponse | null>(null)
const recentLinks = ref<AdminLinkItem[]>([])
const loading = ref(true)

function displayShortUrl(alias: string): string {
  return `${config.routerBase}/${alias}`.replace(/^https?:\/\//, '')
}

async function loadOverview() {
  loading.value = true
  try {
    const [statsRes, linksRes] = await Promise.all([
      fetchAdminStats(),
      fetchAdminLinks(1, 8, '', 'all')
    ])
    stats.value = statsRes
    recentLinks.value = linksRes.items
  } catch (err: any) {
    toast.error(err?.message ?? 'Yönetim verileri alınamadı.')
  } finally {
    loading.value = false
  }
}

onMounted(loadOverview)
</script>

<template>
  <DashboardPageHeader
    title="Yönetim özeti"
    description="Platform genelindeki kullanıcı, link ve yönlendirme metrikleri."
  >
    <template #actions>
      <UiButton variant="secondary" :disabled="loading" @click="loadOverview">
        <UiIcon name="refresh" :size="15" /> Yenile
      </UiButton>
    </template>
  </DashboardPageHeader>

  <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
    <DashboardStatTile
      label="Toplam kullanıcı"
      :value="loading ? '...' : (stats?.totalUsers ?? 0).toLocaleString('tr-TR')"
    />
    <DashboardStatTile
      label="Aktif kullanıcı"
      :value="loading ? '...' : (stats?.activeUsers ?? 0).toLocaleString('tr-TR')"
    />
    <DashboardStatTile
      label="Toplam link"
      :value="loading ? '...' : (stats?.totalLinks ?? 0).toLocaleString('tr-TR')"
    />
    <DashboardStatTile
      label="Aktif link"
      :value="loading ? '...' : (stats?.activeLinks ?? 0).toLocaleString('tr-TR')"
    />
    <DashboardStatTile
      label="Engelli / Zararlı"
      :value="loading ? '...' : (stats?.blockedLinks ?? 0).toLocaleString('tr-TR')"
    />
    <DashboardStatTile
      label="Toplam tıklama"
      :value="loading ? '...' : (stats?.totalClicks ?? 0).toLocaleString('tr-TR')"
    />
  </div>

  <section class="mt-8 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
    <div class="border-b border-line pb-4">
      <h2 class="text-xl font-bold">Son oluşturulan linkler</h2>
    </div>

    <div v-if="loading" class="mt-4 flex flex-col gap-2.5">
      <div v-for="i in 4" :key="i" class="h-14 animate-pulse rounded-md bg-paper-2" />
    </div>

    <p v-else-if="recentLinks.length === 0" class="py-8 text-center text-sm text-ink-3">
      Henüz link yok.
    </p>

    <ul v-else class="mt-2 divide-y divide-line">
      <li v-for="item in recentLinks" :key="item.id">
        <NuxtLink
          :to="`/dashboard/admin/links/${item.id}`"
          class="group flex flex-wrap items-center justify-between gap-3 py-3.5 transition-colors hover:bg-paper/60"
        >
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <span class="font-mono text-sm font-bold text-ink group-hover:text-accent group-hover:underline">
                {{ displayShortUrl(item.alias) }}
              </span>
              <span v-if="item.title" class="truncate text-xs font-medium text-ink-2">• {{ item.title }}</span>
              <UiBadge v-if="!item.isActive" tone="accent">Pasif</UiBadge>
              <UiBadge :tone="securityStatusMeta(item.securityStatus).tone">
                {{ securityStatusMeta(item.securityStatus).label }}
              </UiBadge>
            </div>
            <p class="mt-0.5 truncate text-xs text-ink-3">{{ item.target }}</p>
            <p class="mt-1 font-mono text-[11px] text-ink-3">
              {{ item.userName || item.userEmail }} ({{ item.userEmail }})
            </p>
          </div>
          <div class="flex shrink-0 items-center gap-4">
            <div class="text-right">
              <p class="font-display text-lg font-bold leading-none">{{ item.clicks.toLocaleString('tr-TR') }}</p>
              <p class="mt-1 text-[11px] text-ink-3">{{ formatDateTr(item.createdAt) }}</p>
            </div>
            <span class="text-xs font-semibold text-ink-2 group-hover:text-ink">İncele →</span>
          </div>
        </NuxtLink>
      </li>
    </ul>
  </section>
</template>
