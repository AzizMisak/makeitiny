<script setup lang="ts">
import type { User } from '~/types/auth'

definePageMeta({ layout: 'admin', middleware: 'auth', roles: ['owner'] })
useHead({ title: 'Roller | Yönetim | makeiTiny' })

const PAGE_SIZE = 10
const TEAM_PAGES = 10

const { fetchAdminUsers, adminUpdateUserRole } = useAdmin()
const { ask } = useConfirm()
const toast = useToast()

// Sol: yönetim ekibi. API role göre süzmediği için kullanıcı listesi sayfa sayfa taranır.
const team = ref<User[]>([])
const teamLoading = ref(true)
const teamError = ref('')

async function loadTeam() {
  teamLoading.value = true
  teamError.value = ''
  try {
    const first = await fetchAdminUsers(1, 100)
    const pages = Math.min(Math.max(1, first.totalPages), TEAM_PAGES)
    const rest = await Promise.all(Array.from({ length: pages - 1 }, (_, i) => fetchAdminUsers(i + 2, 100)))
    const rank = { owner: 0, admin: 1, user: 2 }
    team.value = [first, ...rest]
      .flatMap((r) => r.items)
      .filter((u) => roleMeta(u.roles).primary !== 'user')
      .sort((a, b) => rank[roleMeta(a.roles).primary] - rank[roleMeta(b.roles).primary] || a.email.localeCompare(b.email))
  } catch (e: any) {
    teamError.value = e?.message ?? 'Yönetim ekibi yüklenemedi.'
  } finally {
    teamLoading.value = false
  }
}

// Sağ: kullanıcı arama ve yetki verme.
const users = ref<User[]>([])
const page = ref(1)
const total = ref(0)
const totalPages = ref(1)
const search = ref('')
const loading = ref(true)
const loadedOnce = ref(false)
const error = ref('')

async function loadUsers(next = page.value) {
  loading.value = true
  error.value = ''
  try {
    const res = await fetchAdminUsers(next, PAGE_SIZE, search.value)
    users.value = res.items
    page.value = res.page
    total.value = res.total
    totalPages.value = Math.max(1, res.totalPages)
    loadedOnce.value = true
  } catch (e: any) {
    error.value = e?.message ?? 'Kullanıcılar yüklenemedi.'
  } finally {
    loading.value = false
  }
}

function onSearch(value: string) {
  search.value = value
  loadUsers(1)
}

async function change(u: User) {
  const next = u.roles.includes('admin') ? 'user' : 'admin'
  const ok = await ask(roleChangePrompt(u, next, () => adminUpdateUserRole(u.id, next)))
  if (!ok) return
  toast.success(next === 'admin' ? `${u.email} artık yönetici.` : `${u.email} artık standart kullanıcı.`)
  loadTeam()
  loadUsers()
}

onMounted(() => {
  loadTeam()
  loadUsers(1)
})
</script>

<template>
  <div>
    <AdminHeader title="Roller" note="Yalnızca sistem sahibi görür" refreshable :loading="loading || teamLoading" @refresh="loadTeam(), loadUsers()" />

    <div class="grid grid-cols-1 items-start gap-x-10 gap-y-12 lg:grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)]">
      <section>
        <h2 class="mb-4 flex items-baseline gap-3 text-lg font-semibold">
          Yönetim ekibi
          <span v-if="!teamLoading" class="tnum text-base font-normal text-ink-3">{{ team.length }} kişi</span>
        </h2>
        <UiLoadError v-if="teamError" :message="teamError" :loading="teamLoading" @retry="loadTeam" />
        <AdminTableSkeleton v-else-if="teamLoading && team.length === 0" :rows="3" label="Yönetim ekibi yükleniyor" />
        <ul v-else class="divide-y divide-line border border-line bg-surface transition-opacity" :class="teamLoading ? 'opacity-55' : ''">
          <li v-for="u in team" :key="u.id" class="flex flex-wrap items-center gap-3 px-4 py-3.5">
            <AdminAvatar :name="u.name" :email="u.email" />
            <div class="min-w-0 flex-1">
              <NuxtLink :to="`/dashboard/admin/users/${u.id}`" class="block truncate font-semibold hover:text-brand">{{ u.name || 'İsimsiz' }}</NuxtLink>
              <p class="truncate text-sm text-ink-3">{{ u.email }}</p>
            </div>
            <UiTag v-if="u.roles.includes('owner')" tone="mark" icon="key">Sahip</UiTag>
            <UiButton v-else variant="danger" size="sm" @click="change(u)">Yetkiyi kaldır</UiButton>
          </li>
        </ul>
        <p class="mt-3 text-sm text-ink-3">Sistemde tek sahip olur; bu rol devredilemez.</p>
      </section>

      <section>
        <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-lg font-semibold">Yönetici ata</h2>
          <UiSearchBox id="role-search" placeholder="Ad veya e-posta" :active="search" :loading="loading" @search="onSearch" />
        </div>
        <UiLoadError v-if="error" :message="error" :loading="loading" @retry="loadUsers()" />
        <AdminTableSkeleton v-else-if="!loadedOnce" :rows="6" label="Kullanıcılar yükleniyor" />
        <UiEmpty v-else-if="users.length === 0" icon="search" title="Eşleşen kullanıcı yok" class="border border-line bg-surface" />
        <template v-else>
          <ul class="divide-y divide-line border border-line bg-surface transition-opacity" :class="loading ? 'pointer-events-none opacity-55' : ''" :aria-busy="loading">
            <li v-for="u in users" :key="u.id" class="flex min-h-16 flex-wrap items-center gap-3 px-4 py-2.5">
              <div class="min-w-0 flex-1">
                <NuxtLink :to="`/dashboard/admin/users/${u.id}`" class="block truncate font-medium hover:text-brand">{{ u.name || 'İsimsiz' }}</NuxtLink>
                <p class="truncate text-sm text-ink-3">{{ u.email }}</p>
              </div>
              <UiStatus v-if="!u.isActive" tone="danger" label="Askıda" />
              <UiTag v-if="roleMeta(u.roles).primary !== 'user'" :tone="roleMeta(u.roles).tone">{{ roleMeta(u.roles).label }}</UiTag>
              <UiButton v-else variant="outline" size="sm" @click="change(u)">Yönetici yap</UiButton>
            </li>
          </ul>
          <UiPagination :page="page" :total-pages="totalPages" :total="total" :page-size="PAGE_SIZE" :loading="loading" item-label="kullanıcı" @change="loadUsers" />
        </template>
      </section>
    </div>
  </div>
</template>
