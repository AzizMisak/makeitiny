<script setup lang="ts">
import type { User } from '~/types/auth'

definePageMeta({ layout: 'admin', middleware: 'auth', roles: ['admin', 'owner'] })
useHead({ title: 'Kullanıcılar | Yönetim | makeiTiny' })

const PAGE_SIZE = 20
const route = useRoute()
const router = useRouter()
const { fetchAdminUsers } = useAdmin()

const users = ref<User[]>([])
const page = ref(1)
const total = ref(0)
const totalPages = ref(1)
const loading = ref(true)
const loadedOnce = ref(false)
const error = ref('')
const search = ref(String(route.query.ara ?? ''))

async function load(next = page.value) {
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
  router.replace({ query: value ? { ara: value } : {} })
  load(1)
}

onMounted(() => load(1))
</script>

<template>
  <div>
    <AdminHeader title="Kullanıcılar" :note="loadedOnce ? `${formatNumber(total)} kayıt` : ''" refreshable :loading="loading" @refresh="load()">
      <UiSearchBox
        id="admin-user-search"
        class="sm:w-96"
        placeholder="Ad, e-posta veya kimlik"
        :active="search"
        :loading="loading"
        @search="onSearch"
      />
    </AdminHeader>

    <UiLoadError v-if="error" :message="error" :loading="loading" @retry="load()" />
    <AdminTableSkeleton v-else-if="!loadedOnce" :rows="10" label="Kullanıcılar yükleniyor" />
    <UiEmpty v-else-if="users.length === 0" icon="search" title="Eşleşen kullanıcı yok" text="Başka bir ad veya e-posta dene." class="border border-line bg-surface" />

    <template v-else>
      <div class="overflow-x-auto border border-line bg-surface transition-opacity" :class="loading ? 'pointer-events-none opacity-55' : ''" :aria-busy="loading">
        <table class="w-full min-w-[54rem] text-left text-sm">
          <caption class="sr-only">Platformdaki kullanıcılar</caption>
          <thead class="bg-sunken text-ink-2">
            <tr>
              <th class="px-4 py-2.5 font-medium">Kullanıcı</th>
              <th class="px-4 py-2.5 font-medium">Rol</th>
              <th class="px-4 py-2.5 font-medium">Durum</th>
              <th class="px-4 py-2.5 font-medium">Paket</th>
              <th class="px-4 py-2.5 text-right font-medium">Link</th>
              <th class="px-4 py-2.5 font-medium">Kayıt</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-line">
            <tr v-for="u in users" :key="u.id" class="relative hover:bg-brand-soft/40">
              <td class="max-w-sm px-4 py-3">
                <div class="flex items-center gap-3">
                  <AdminAvatar :name="u.name" :email="u.email" />
                  <div class="min-w-0">
                    <NuxtLink
                      :to="`/dashboard/admin/users/${u.id}`"
                      class="block truncate font-semibold after:absolute after:inset-0 after:content-[''] hover:text-brand"
                    >{{ u.name || 'İsimsiz' }}</NuxtLink>
                    <p class="truncate text-ink-3">{{ u.email }}</p>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3">
                <UiTag v-if="roleMeta(u.roles).primary !== 'user'" :tone="roleMeta(u.roles).tone">{{ roleMeta(u.roles).label }}</UiTag>
                <span v-else class="text-ink-2">Kullanıcı</span>
              </td>
              <td class="max-w-56 px-4 py-3">
                <UiStatus :tone="u.isActive ? 'ok' : 'danger'" :label="u.isActive ? 'Aktif' : 'Askıda'" />
                <p v-if="!u.isActive && u.banReason" class="truncate text-ink-3" :title="u.banReason">{{ u.banReason }}</p>
              </td>
              <td class="px-4 py-3">
                {{ u.plan.name }}
                <p v-if="u.planExpiresAt" class="tnum text-ink-3">{{ formatDateTr(u.planExpiresAt) }} biter</p>
              </td>
              <td class="tnum px-4 py-3 text-right font-semibold">{{ formatNumber(u.linkCount) }}</td>
              <td class="tnum whitespace-nowrap px-4 py-3 text-ink-2">{{ formatDateTr(u.createdAt) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <UiPagination :page="page" :total-pages="totalPages" :total="total" :page-size="PAGE_SIZE" :loading="loading" item-label="kullanıcı" @change="load" />
    </template>
  </div>
</template>
