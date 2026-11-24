<script setup lang="ts">
import type { User } from '~/types/auth'
import type { AdminStatsResponse } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth', roles: ['admin', 'owner'] })
useHead({ title: 'Kullanıcılar — Yönetici — makeiTiny' })

const { fetchAdminStats, fetchAdminUsers } = useAdmin()
const toast = useToast()

const PAGE_SIZE = 15

const stats = ref<AdminStatsResponse | null>(null)
const users = ref<User[]>([])
const page = ref(1)
const total = ref(0)
const totalPages = ref(1)
const loading = ref(true)

const searchInput = ref('')
const activeSearch = ref('')

async function loadUsers(targetPage = page.value) {
  loading.value = true
  try {
    const [usersRes, statsRes] = await Promise.all([
      fetchAdminUsers(targetPage, PAGE_SIZE, activeSearch.value),
      fetchAdminStats().catch(() => null)
    ])
    users.value = usersRes.items
    page.value = usersRes.page
    total.value = usersRes.total
    totalPages.value = Math.max(1, usersRes.totalPages)
    if (statsRes) stats.value = statsRes
  } catch (err: any) {
    toast.error(err?.message ?? 'Kullanıcı listesi alınamadı.')
  } finally {
    loading.value = false
  }
}

function submitSearch() {
  activeSearch.value = searchInput.value.trim()
  loadUsers(1)
}

function clearSearch() {
  searchInput.value = ''
  activeSearch.value = ''
  loadUsers(1)
}

onMounted(() => loadUsers(1))
</script>

<template>
  <DashboardPageHeader
    title="Kullanıcılar"
    description="Profil detaylarını, linklerini ve paket ayarlarını görüntülemek için bir kullanıcıya tıkla."
  >
    <template #actions>
      <UiButton variant="secondary" :disabled="loading" @click="loadUsers(page)">
        <UiIcon name="refresh" :size="15" /> Yenile
      </UiButton>
    </template>
  </DashboardPageHeader>

  <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
    <DashboardStatTile
      label="Toplam kullanıcı"
      :value="stats ? stats.totalUsers.toLocaleString('tr-TR') : total.toLocaleString('tr-TR')"
    />
    <DashboardStatTile
      label="Aktif hesap"
      :value="stats ? stats.activeUsers.toLocaleString('tr-TR') : '...'"
    />
    <DashboardStatTile
      label="Askıdaki hesap"
      :value="stats ? Math.max(0, stats.totalUsers - stats.activeUsers).toLocaleString('tr-TR') : '...'"
    />
    <DashboardStatTile
      label="Listelenen"
      :value="total.toLocaleString('tr-TR')"
      :note="activeSearch ? `'${activeSearch}'` : undefined"
    />
  </div>

  <form
    class="mt-8 flex flex-wrap items-center gap-2 rounded-lg border-[1.5px] border-ink bg-card p-5 shadow-hard"
    @submit.prevent="submitSearch"
  >
    <div class="relative min-w-60 flex-1">
      <UiIcon name="search" :size="15" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-3" />
      <input
        v-model="searchInput"
        type="search"
        placeholder="E-posta veya ad soyad ile ara..."
        class="w-full rounded-sm border-[1.5px] border-ink bg-paper py-2 pl-9 pr-3 text-sm text-ink placeholder:text-ink-3 focus:outline-none focus:ring-2 focus:ring-accent"
      />
    </div>
    <UiButton type="submit" :disabled="loading">Ara</UiButton>
    <UiButton v-if="activeSearch" variant="ghost" @click="clearSearch">Temizle</UiButton>
  </form>

  <div class="mt-6 rounded-lg border-[1.5px] border-ink bg-card shadow-hard">
    <div v-if="loading" class="flex flex-col gap-2.5 p-5">
      <div v-for="i in 5" :key="i" class="h-14 animate-pulse rounded-md bg-paper-2" />
    </div>

    <div v-else-if="users.length === 0" class="p-10 text-center">
      <p class="text-base font-semibold text-ink">Kullanıcı bulunamadı.</p>
      <p class="mt-1 text-sm text-ink-3">Farklı bir isim veya e-posta adresi ile ara.</p>
    </div>

    <ul v-else class="divide-y divide-line">
      <li v-for="u in users" :key="u.id">
        <NuxtLink
          :to="`/dashboard/admin/users/${u.id}`"
          class="group flex flex-col gap-3 p-4 transition-colors hover:bg-paper/60 sm:flex-row sm:items-center sm:justify-between sm:p-5"
        >
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-base font-bold text-ink group-hover:text-accent group-hover:underline">
                {{ u.name || 'İsimsiz Kullanıcı' }}
              </span>
              <span class="font-mono text-xs text-ink-2">{{ u.email }}</span>
              <UiBadge :tone="roleBadgeMeta(u.roles).tone">{{ roleBadgeMeta(u.roles).label }}</UiBadge>
              <UiBadge :tone="u.isActive ? 'ok' : 'accent'">
                {{ u.isActive ? 'Aktif' : 'Askıda' }}
              </UiBadge>
            </div>

            <div class="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 font-mono text-[11px] text-ink-3">
              <span>Paket: <strong class="text-ink-2">{{ u.plan.name }}</strong></span>
              <span>Bitiş: {{ u.planExpiresAt ? formatDateTr(u.planExpiresAt) : 'Süresiz' }}</span>
              <span>Kayıt: {{ formatDateTr(u.createdAt) }}</span>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-4">
            <div class="text-right">
              <span class="block font-display text-lg font-bold leading-none text-ink">
                {{ u.linkCount.toLocaleString('tr-TR') }}
              </span>
              <span class="text-xs text-ink-3">link</span>
            </div>
            <span class="inline-flex items-center gap-1 rounded-sm border border-line bg-paper px-2.5 py-1 text-xs font-semibold text-ink-2 transition-colors group-hover:border-ink group-hover:text-ink">
              Profili aç <UiIcon name="arrow" :size="13" />
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
    item-label="kullanıcı"
    @change="loadUsers"
  />
</template>
