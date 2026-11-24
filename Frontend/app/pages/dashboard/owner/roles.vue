<script setup lang="ts">
import type { User } from '~/types/auth'
import type { AdminStatsResponse } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth', roles: ['owner'] })
useHead({ title: 'Roller & Yetkiler — Owner — makeiTiny' })

type RoleFilter = 'all' | 'owner' | 'admin' | 'user'

const { fetchAdminStats, fetchAdminUsers, adminUpdateUserRole } = useAdmin()
const toast = useToast()

const PAGE_SIZE = 20

const stats = ref<AdminStatsResponse | null>(null)
const users = ref<User[]>([])
const page = ref(1)
const total = ref(0)
const totalPages = ref(1)
const loading = ref(true)

const searchInput = ref('')
const activeSearch = ref('')
const roleFilter = ref<RoleFilter>('all')

const pendingRoleUser = ref<User | null>(null)
const savingRole = ref(false)

const roleTabs: { value: RoleFilter; label: string }[] = [
  { value: 'all', label: 'Tümü' },
  { value: 'owner', label: 'Owner' },
  { value: 'admin', label: 'Admin' },
  { value: 'user', label: 'Kullanıcı' }
]

const filteredUsers = computed(() => {
  if (roleFilter.value === 'all') return users.value
  return users.value.filter((u) => roleBadgeMeta(u.roles).primary === roleFilter.value)
})

const adminCountOnPage = computed(() =>
  users.value.filter((u) => u.roles.includes('admin')).length
)

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
    toast.error(err?.message ?? 'Kullanıcı rolleri yüklenemedi.')
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

function openRoleModal(u: User) {
  if (u.roles.includes('owner')) return
  pendingRoleUser.value = u
}

async function confirmRoleChange() {
  const target = pendingRoleUser.value
  if (!target || target.roles.includes('owner')) return
  const nextRole = target.roles.includes('admin') ? 'user' : 'admin'

  savingRole.value = true
  try {
    const updated = await adminUpdateUserRole(target.id, nextRole)
    const idx = users.value.findIndex((u) => u.id === target.id)
    if (idx !== -1) users.value[idx] = updated
    pendingRoleUser.value = null
    toast.success(
      `${updated.email} rolü '${nextRole === 'admin' ? 'Admin' : 'Kullanıcı'}' olarak güncellendi.`
    )
  } catch (err: any) {
    toast.error(err?.message ?? 'Rol güncellenemedi.')
  } finally {
    savingRole.value = false
  }
}

onMounted(() => loadUsers(1))
</script>

<template>
  <DashboardPageHeader
    title="Roller & yetkiler"
    description="Kullanıcı profillerini incele veya onay adımıyla yönetici yetkisi ata."
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
      label="Sayfadaki admin"
      :value="adminCountOnPage"
    />
    <DashboardStatTile
      label="Owner"
      value="1"
      note="Sabit tek hesap"
    />
    <DashboardStatTile
      label="Roller"
      value="owner / admin / user"
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
            placeholder="E-posta veya ad soyad ile ara..."
            class="w-full rounded-sm border-[1.5px] border-ink bg-paper py-2 pl-9 pr-3 text-sm text-ink placeholder:text-ink-3 focus:outline-none focus:ring-2 focus:ring-accent"
          />
        </div>
        <UiButton type="submit" :disabled="loading">Ara</UiButton>
        <UiButton v-if="activeSearch" variant="ghost" @click="clearSearch">Temizle</UiButton>
      </form>

      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="tab in roleTabs"
          :key="tab.value"
          type="button"
          class="cursor-pointer rounded-sm border px-3 py-1.5 text-xs font-semibold transition-colors"
          :class="
            roleFilter === tab.value
              ? 'border-ink bg-ink text-paper'
              : 'border-line bg-paper text-ink-2 hover:border-ink hover:text-ink'
          "
          @click="roleFilter = tab.value"
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

    <div v-else-if="filteredUsers.length === 0" class="p-10 text-center">
      <p class="text-base font-semibold text-ink">Kullanıcı bulunamadı.</p>
      <p class="mt-1 text-sm text-ink-3">Rol filtresini 'Tümü' yap veya başka bir arama dene.</p>
    </div>

    <ul v-else class="divide-y divide-line">
      <li v-for="u in filteredUsers" :key="u.id" class="p-4 transition-colors hover:bg-paper/60 sm:p-5">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <NuxtLink :to="`/dashboard/admin/users/${u.id}`" class="group min-w-0 flex-1">
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
            <p class="mt-1 font-mono text-[11px] text-ink-3">
              Paket: {{ u.plan.name }} • {{ u.linkCount }} link • Kayıt: {{ formatDateTr(u.createdAt) }}
            </p>
          </NuxtLink>

          <div class="flex shrink-0 flex-wrap items-center gap-2">
            <UiButton :to="`/dashboard/admin/users/${u.id}`" variant="ghost">
              Profili aç
            </UiButton>

            <UiBadge v-if="u.roles.includes('owner')" tone="accent">
              <UiIcon name="lock" :size="11" /> Sabit Owner
            </UiBadge>
            <UiButton
              v-else
              variant="secondary"
              @click="openRoleModal(u)"
            >
              Rolü değiştir...
            </UiButton>
          </div>
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
    item-label="kullanıcı"
    @change="loadUsers"
  />

  <UiConfirmModal
    v-if="pendingRoleUser"
    :open="Boolean(pendingRoleUser)"
    :title="pendingRoleUser.roles.includes('admin') ? 'Admin yetkisini kaldır' : 'Admin yetkisi ver'"
    :target-label="`${pendingRoleUser.name || 'Kullanıcı'} (${pendingRoleUser.email})`"
    :description="
      pendingRoleUser.roles.includes('admin')
        ? 'Bu kullanıcının yönetici yetkileri kaldırılacak ve standart kullanıcı rolüne düşürülecek.'
        : 'Bu kullanıcı tüm linkleri, tıklama loglarını ve standart kullanıcı hesaplarını yönetebilecek Admin yetkisine yükseltilecek.'
    "
    :changes="[
      {
        label: 'Rol',
        from: pendingRoleUser.roles.includes('admin') ? 'ADMİN' : 'KULLANICI',
        to: pendingRoleUser.roles.includes('admin') ? 'KULLANICI' : 'ADMİN'
      }
    ]"
    :confirm-label="pendingRoleUser.roles.includes('admin') ? 'Yetkiyi kaldır' : 'Evet, Admin yap'"
    variant="danger"
    :loading="savingRole"
    @cancel="pendingRoleUser = null"
    @confirm="confirmRoleChange"
  />
</template>
