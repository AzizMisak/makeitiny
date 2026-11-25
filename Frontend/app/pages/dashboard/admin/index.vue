<script setup lang="ts">
import type { AdminLinkItem, AdminStatsResponse } from '~/types/link'

definePageMeta({ layout: 'admin', middleware: 'auth', roles: ['admin', 'owner'] })
useHead({ title: 'Özet | Yönetim | makeiTiny' })

const { fetchAdminStats, fetchAdminLinks } = useAdmin()

const stats = ref<AdminStatsResponse | null>(null)
const recent = ref<AdminLinkItem[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [s, l] = await Promise.all([fetchAdminStats(), fetchAdminLinks(1, 10, '', 'all')])
    stats.value = s
    recent.value = l.items
  } catch (e: any) {
    error.value = e?.message ?? 'Yönetim verileri yüklenemedi.'
  } finally {
    loading.value = false
  }
}

function share(part: number, whole: number): number {
  return whole > 0 ? (part / whole) * 100 : 0
}

// Bütün / parça ilişkisi olan sayılar tek çubukta gösterilir.
const ratios = computed(() => {
  const s = stats.value
  if (!s) return []
  return [
    {
      key: 'users',
      total: s.totalUsers,
      label: 'kullanıcı',
      a: { label: 'Aktif', value: s.activeUsers },
      b: { label: 'Askıda', value: Math.max(0, s.totalUsers - s.activeUsers) },
      to: '/dashboard/admin/users'
    },
    {
      key: 'links',
      total: s.totalLinks,
      label: 'link',
      a: { label: 'Açık', value: s.activeLinks },
      b: { label: 'Kapalı', value: Math.max(0, s.totalLinks - s.activeLinks) },
      to: '/dashboard/admin/links'
    }
  ]
})

onMounted(load)
</script>

<template>
  <div>
    <AdminHeader title="Özet" refreshable :loading="loading" @refresh="load" />
    <UiLoadError v-if="error" :message="error" :loading="loading" class="mb-6" @retry="load" />

    <section class="grid grid-cols-1 border border-line bg-surface md:grid-cols-[1fr_1fr_0.8fr_0.8fr] md:divide-x md:divide-line">
      <template v-if="!stats">
        <div v-for="i in 4" :key="i" class="p-6"><div class="skeleton h-10 w-28" /><div class="skeleton mt-4 h-4 w-40" /></div>
      </template>
      <template v-else>
        <NuxtLink v-for="r in ratios" :key="r.key" :to="r.to" class="group block p-6 max-md:border-b max-md:border-line">
          <p class="flex items-baseline gap-2">
            <span class="display tnum text-4xl">{{ formatNumber(r.total) }}</span>
            <span class="text-ink-2 group-hover:text-brand">{{ r.label }}</span>
          </p>
          <span class="mt-4 flex h-3 gap-0.5" aria-hidden="true">
            <span class="bg-brand" :style="{ width: `${share(r.a.value, r.total)}%` }" />
            <span class="flex-1 bg-line-2/50" />
          </span>
          <p class="tnum mt-2 flex justify-between text-sm text-ink-2">
            <span>{{ r.a.label }} {{ formatNumber(r.a.value) }}</span>
            <span>{{ r.b.label }} {{ formatNumber(r.b.value) }}</span>
          </p>
        </NuxtLink>
        <div class="p-6 max-md:border-b max-md:border-line">
          <p class="display tnum text-4xl">{{ formatNumber(stats.totalClicks) }}</p>
          <p class="mt-1 text-ink-2">toplam tıklama</p>
        </div>
        <NuxtLink to="/dashboard/admin/links?durum=blocked" class="group block p-6">
          <p class="display tnum text-4xl" :class="stats.blockedLinks > 0 ? 'text-danger' : ''">{{ formatNumber(stats.blockedLinks) }}</p>
          <p class="mt-1 text-ink-2 group-hover:text-brand">engelli veya zararlı link</p>
        </NuxtLink>
      </template>
    </section>

    <section class="mt-10">
      <div class="mb-4 flex items-baseline justify-between">
        <h2 class="text-lg font-semibold">Son oluşturulan linkler</h2>
        <NuxtLink to="/dashboard/admin/links" class="text-sm font-medium text-brand-ink underline-offset-4 hover:underline">Bütün linkler</NuxtLink>
      </div>
      <AdminTableSkeleton v-if="loading && recent.length === 0" label="Linkler yükleniyor" />
      <UiEmpty v-else-if="recent.length === 0" title="Platformda henüz link yok" class="border border-line bg-surface" />
      <AdminLinkTable v-else :links="recent" :loading="loading" caption="Son oluşturulan linkler" />
    </section>
  </div>
</template>
