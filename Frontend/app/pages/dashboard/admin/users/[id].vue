<script setup lang="ts">
import type { ConfirmChangeItem } from '~/components/ui/ConfirmModal.vue'
import type { PlanItem, User } from '~/types/auth'
import type { AdminLinkItem } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth', roles: ['admin', 'owner'] })

const route = useRoute()
const { isOwner } = useAuth()
const { public: config } = useRuntimeConfig()
const {
  fetchPlans,
  fetchAdminUserById,
  fetchAdminUserLinks,
  adminUpdateUser,
  adminUpdateUserRole
} = useAdmin()
const toast = useToast()

const userId = computed(() => String(route.params.id ?? ''))
const targetUser = ref<User | null>(null)
const plans = ref<PlanItem[]>([])
const userLinks = ref<AdminLinkItem[]>([])
const pageLoading = ref(true)
const saving = ref(false)

const selectedPlanId = ref(1)
const selectedPlanExpiresAt = ref('')

const confirmPlanOpen = ref(false)
const confirmStatusOpen = ref(false)
const confirmRoleOpen = ref(false)

useHead({
  title: computed(() =>
    targetUser.value ? `${targetUser.value.name || targetUser.value.email} — Kullanıcı İnceleme` : 'Kullanıcı İnceleme — makeiTiny'
  )
})

function displayShortUrl(alias: string): string {
  return `${config.routerBase}/${alias}`.replace(/^https?:\/\//, '')
}

function syncUserDraft(u: User) {
  targetUser.value = u
  selectedPlanId.value = u.plan?.id ?? 1
  selectedPlanExpiresAt.value = u.planExpiresAt ?? ''
}

const canManageAccount = computed(() => {
  const u = targetUser.value
  if (!u) return false
  if (isOwner.value) return true
  return !u.roles.includes('owner') && !u.roles.includes('admin')
})

const canToggleActive = computed(() => {
  const u = targetUser.value
  if (!u || u.roles.includes('owner')) return false
  return canManageAccount.value
})

const canChangeRole = computed(() => {
  const u = targetUser.value
  if (!u || !isOwner.value) return false
  return !u.roles.includes('owner')
})

const planChanges = computed<ConfirmChangeItem[]>(() => {
  const u = targetUser.value
  if (!u) return []
  const changes: ConfirmChangeItem[] = []

  if (selectedPlanId.value !== u.plan.id) {
    const nextPlanName = plans.value.find((p) => p.id === selectedPlanId.value)?.name ?? String(selectedPlanId.value)
    changes.push({
      label: 'Paket',
      from: u.plan.name,
      to: nextPlanName
    })
  }

  const curExp = u.planExpiresAt ?? ''
  const nextExp = selectedPlanExpiresAt.value.trim()
  if (nextExp !== curExp) {
    changes.push({
      label: 'Paket bitiş tarihi',
      from: curExp ? formatDateTr(curExp) : 'Süresiz',
      to: nextExp ? formatDateTr(nextExp) : 'Süresiz'
    })
  }

  return changes
})

async function loadUserDetail() {
  pageLoading.value = true
  try {
    const [found, planList] = await Promise.all([
      fetchAdminUserById(userId.value),
      fetchPlans()
    ])
    plans.value = planList
    if (found) {
      syncUserDraft(found)
      userLinks.value = await fetchAdminUserLinks(found.id, found.email).catch(() => [])
    }
  } catch (err: any) {
    toast.error(err?.message ?? 'Kullanıcı bilgileri yüklenemedi.')
  } finally {
    pageLoading.value = false
  }
}

function requestPlanSave() {
  if (!canManageAccount.value) return
  if (planChanges.value.length === 0) {
    toast.info('Kaydedilecek bir paket değişikliği yok.')
    return
  }
  confirmPlanOpen.value = true
}

async function confirmPlanSave() {
  if (!targetUser.value) return
  saving.value = true
  try {
    const updated = await adminUpdateUser(targetUser.value.id, {
      planId: selectedPlanId.value,
      planExpiresAt: selectedPlanExpiresAt.value.trim()
    })
    syncUserDraft(updated)
    confirmPlanOpen.value = false
    toast.success('Kullanıcı paketi güncellendi.')
  } catch (err: any) {
    toast.error(err?.message ?? 'Paket güncellenemedi.')
  } finally {
    saving.value = false
  }
}

async function confirmStatusToggle() {
  if (!targetUser.value || !canToggleActive.value) return
  saving.value = true
  try {
    const updated = await adminUpdateUser(targetUser.value.id, {
      isActive: !targetUser.value.isActive
    })
    syncUserDraft(updated)
    confirmStatusOpen.value = false
    toast.success(
      updated.isActive ? 'Kullanıcı hesabı aktifleştirildi.' : 'Kullanıcı hesabı askıya alındı.'
    )
  } catch (err: any) {
    toast.error(err?.message ?? 'Hesap durumu güncellenemedi.')
  } finally {
    saving.value = false
  }
}

async function confirmRoleToggle() {
  if (!targetUser.value || !canChangeRole.value) return
  const nextRole = targetUser.value.roles.includes('admin') ? 'user' : 'admin'
  saving.value = true
  try {
    const updated = await adminUpdateUserRole(targetUser.value.id, nextRole)
    syncUserDraft(updated)
    confirmRoleOpen.value = false
    toast.success(
      nextRole === 'admin' ? 'Kullanıcıya Admin yetkisi verildi.' : 'Kullanıcının Admin yetkisi kaldırıldı.'
    )
  } catch (err: any) {
    toast.error(err?.message ?? 'Rol güncellenemedi.')
  } finally {
    saving.value = false
  }
}

onMounted(loadUserDetail)
</script>

<template>
  <div>
    <div class="mb-5">
      <NuxtLink
        to="/dashboard/admin/users"
        class="inline-flex items-center gap-2 text-sm font-medium text-ink-2 hover:text-ink"
      >
        <UiIcon name="back" /> Kullanıcılara dön
      </NuxtLink>
    </div>

    <div v-if="pageLoading" class="flex flex-col gap-4">
      <div class="h-16 animate-pulse rounded-lg bg-paper-2" />
      <div class="h-64 animate-pulse rounded-lg bg-paper-2" />
    </div>

    <div
      v-else-if="!targetUser"
      class="rounded-lg border-[1.5px] border-ink bg-card p-8 text-center shadow-hard"
    >
      <h1 class="text-xl font-bold">Kullanıcı bulunamadı</h1>
      <p class="mt-1 text-sm text-ink-3">Aradığın kullanıcı mevcut değil veya silinmiş.</p>
      <div class="mt-5">
        <UiButton to="/dashboard/admin/users" variant="secondary">Listeye dön</UiButton>
      </div>
    </div>

    <template v-else>
      <DashboardPageHeader
        :title="targetUser.name || 'İsimsiz Kullanıcı'"
        :description="targetUser.email"
      />

      <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
        <DashboardStatTile
          label="Hesap durumu"
          :value="targetUser.isActive ? 'Aktif' : 'Askıda'"
        />
        <DashboardStatTile
          label="Rol"
          :value="roleBadgeMeta(targetUser.roles).label"
        />
        <DashboardStatTile
          label="Mevcut paket"
          :value="targetUser.plan.name"
          :note="targetUser.planExpiresAt ? `Bitiş: ${formatDateTr(targetUser.planExpiresAt)}` : 'Süresiz'"
        />
        <DashboardStatTile
          label="Oluşturulan link"
          :value="targetUser.linkCount.toLocaleString('tr-TR')"
          :note="`Kayıt: ${formatDateTr(targetUser.createdAt)}`"
        />
      </div>

      <div class="mt-8 grid gap-6 lg:grid-cols-2">
        <!-- Paket Yönetimi Kartı -->
        <form
          class="flex flex-col justify-between rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard"
          @submit.prevent="requestPlanSave"
        >
          <div>
            <div class="flex items-center justify-between gap-2 border-b border-line pb-4">
              <h2 class="text-xl font-bold">Paket & süre yönetimi</h2>
              <UiBadge v-if="!canManageAccount" tone="neutral">
                <UiIcon name="lock" :size="11" /> Korumalı
              </UiBadge>
            </div>

            <div class="mt-5 flex flex-col gap-4">
              <div>
                <label for="user-plan-select" class="mb-1.5 block text-sm font-semibold text-ink">Tanımlı paket</label>
                <select
                  id="user-plan-select"
                  v-model.number="selectedPlanId"
                  :disabled="!canManageAccount"
                  class="w-full rounded-md border-[1.5px] border-ink bg-card px-3 py-2.5 text-sm font-semibold text-ink focus:outline-none focus:ring-2 focus:ring-accent disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option v-for="p in plans" :key="p.id" :value="p.id">
                    {{ p.name }} ({{ formatPriceMonthly(p.priceMonthly) }})
                  </option>
                </select>
              </div>

              <div>
                <label for="user-plan-exp" class="mb-1.5 block text-sm font-semibold text-ink">
                  Paket bitiş tarihi (boş bırakılırsa süresiz)
                </label>
                <input
                  id="user-plan-exp"
                  v-model="selectedPlanExpiresAt"
                  type="date"
                  :disabled="!canManageAccount"
                  class="w-full rounded-md border-[1.5px] border-ink bg-card px-3 py-2.5 font-mono text-sm text-ink focus:outline-none focus:ring-2 focus:ring-accent disabled:cursor-not-allowed disabled:opacity-50"
                />
              </div>
            </div>
          </div>

          <div class="mt-6 flex items-center justify-between border-t border-line pt-4">
            <span class="text-xs text-ink-3">
              {{ planChanges.length > 0 ? `${planChanges.length} değişiklik seçildi` : 'Değişiklik yok' }}
            </span>
            <UiButton
              type="submit"
              :disabled="!canManageAccount || planChanges.length === 0 || saving"
            >
              Değişiklikleri incele
            </UiButton>
          </div>
        </form>

        <!-- Hesap Durumu ve Yetki Kartı -->
        <div class="flex flex-col gap-6">
          <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
            <div class="flex items-center justify-between gap-2 border-b border-line pb-4">
              <h2 class="text-xl font-bold">Hesap erişim durumu</h2>
              <UiBadge :tone="targetUser.isActive ? 'ok' : 'accent'">
                {{ targetUser.isActive ? 'Aktif' : 'Askıda' }}
              </UiBadge>
            </div>

            <p v-if="!canToggleActive" class="mt-4 text-sm text-ink-2">
              Bu hesabın durumu mevcut yetki seviyenle değiştirilemez.
            </p>
            <template v-else>
              <p class="mt-4 text-sm text-ink-2">
                {{
                  targetUser.isActive
                    ? 'Hesap askıya alındığında kullanıcı sisteme giriş yapamaz.'
                    : 'Hesap şu anda askıda. Yeniden aktifleştirerek giriş izni verebilirsin.'
                }}
              </p>
              <div class="mt-4">
                <UiButton
                  :variant="targetUser.isActive ? 'danger' : 'primary'"
                  @click="confirmStatusOpen = true"
                >
                  {{ targetUser.isActive ? 'Hesabı askıya al...' : 'Hesabı aktifleştir...' }}
                </UiButton>
              </div>
            </template>
          </div>

          <div v-if="isOwner" class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
            <div class="flex items-center justify-between gap-2 border-b border-line pb-4">
              <h2 class="text-xl font-bold">Sistem rolü</h2>
              <UiBadge :tone="roleBadgeMeta(targetUser.roles).tone">
                {{ roleBadgeMeta(targetUser.roles).label }}
              </UiBadge>
            </div>

            <p v-if="targetUser.roles.includes('owner')" class="mt-4 text-sm text-ink-2">
              Sistemde tek Owner bulunur; bu rol değiştirilemez.
            </p>
            <template v-else>
              <p class="mt-4 text-sm text-ink-2">
                {{
                  targetUser.roles.includes('admin')
                    ? 'Bu kullanıcı yönetici yetkilerine sahip. Yetkiyi geri alarak standart kullanıcıya dönüştürebilirsin.'
                    : 'Bu kullanıcıya yönetici (Admin) yetkisi verebilirsin.'
                }}
              </p>
              <div class="mt-4">
                <UiButton
                  :variant="targetUser.roles.includes('admin') ? 'danger' : 'secondary'"
                  @click="confirmRoleOpen = true"
                >
                  {{ targetUser.roles.includes('admin') ? 'Admin yetkisini kaldır...' : 'Admin yetkisi ver...' }}
                </UiButton>
              </div>
            </template>
          </div>
        </div>
      </div>

      <!-- Kullanıcının Oluşturduğu Linkler -->
      <section class="mt-8 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
        <div class="flex items-center justify-between gap-2 border-b border-line pb-4">
          <h2 class="text-xl font-bold">Kullanıcının linkleri</h2>
          <UiBadge tone="neutral">{{ userLinks.length }} link</UiBadge>
        </div>

        <p v-if="userLinks.length === 0" class="py-8 text-center text-sm text-ink-3">
          Bu kullanıcının oluşturduğu link bulunmuyor.
        </p>

        <ul v-else class="mt-2 divide-y divide-line">
          <li v-for="item in userLinks" :key="item.id">
            <NuxtLink
              :to="`/dashboard/admin/links/${item.id}`"
              class="group flex flex-wrap items-center justify-between gap-3 py-3.5 transition-colors hover:bg-paper/60"
            >
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-mono text-sm font-bold text-ink group-hover:text-accent group-hover:underline">
                    {{ displayShortUrl(item.alias) }}
                  </span>
                  <span v-if="item.title" class="truncate text-xs text-ink-2">• {{ item.title }}</span>
                  <UiBadge :tone="item.isActive ? 'ok' : 'accent'">
                    {{ item.isActive ? 'Aktif' : 'Pasif' }}
                  </UiBadge>
                  <UiBadge :tone="securityStatusMeta(item.securityStatus).tone">
                    {{ securityStatusMeta(item.securityStatus).label }}
                  </UiBadge>
                </div>
                <p class="mt-0.5 truncate text-xs text-ink-3">{{ item.target }}</p>
              </div>

              <div class="flex items-center gap-3">
                <span class="font-mono text-xs font-bold text-ink">{{ item.clicks.toLocaleString('tr-TR') }} tıklama</span>
                <span class="text-xs font-semibold text-ink-2 group-hover:text-ink">İncele →</span>
              </div>
            </NuxtLink>
          </li>
        </ul>
      </section>

      <!-- Paket Güncelleme Onay Modalı -->
      <UiConfirmModal
        :open="confirmPlanOpen"
        title="Paket değişikliklerini onayla"
        :target-label="`${targetUser.name || 'Kullanıcı'} (${targetUser.email})`"
        description="Kullanıcının abonelik bilgileri aşağıdaki gibi güncellenecek:"
        :changes="planChanges"
        confirm-label="Paketi güncelle"
        :loading="saving"
        @cancel="confirmPlanOpen = false"
        @confirm="confirmPlanSave"
      />

      <!-- Hesap Askıya Alma / Aktifleştirme Onay Modalı -->
      <UiConfirmModal
        :open="confirmStatusOpen"
        :title="targetUser.isActive ? 'Hesabı askıya al' : 'Hesabı aktifleştir'"
        :target-label="`${targetUser.name || 'Kullanıcı'} (${targetUser.email})`"
        :description="
          targetUser.isActive
            ? 'Bu kullanıcının hesabı askıya alınacak ve mevcut oturumu sonlandırılacak.'
            : 'Bu kullanıcının hesabı yeniden aktifleştirilecek.'
        "
        :confirm-label="targetUser.isActive ? 'Evet, askıya al' : 'Evet, aktifleştir'"
        :variant="targetUser.isActive ? 'danger' : 'primary'"
        :loading="saving"
        @cancel="confirmStatusOpen = false"
        @confirm="confirmStatusToggle"
      />

      <!-- Rol Değiştirme Onay Modalı -->
      <UiConfirmModal
        :open="confirmRoleOpen"
        :title="targetUser.roles.includes('admin') ? 'Admin yetkisini kaldır' : 'Admin yetkisi ver'"
        :target-label="`${targetUser.name || 'Kullanıcı'} (${targetUser.email})`"
        :description="
          targetUser.roles.includes('admin')
            ? 'Bu kullanıcının yönetici yetkileri kaldırılacak ve standart kullanıcı rolüne düşürülecek.'
            : 'Bu kullanıcı tüm linkleri, tıklama loglarını ve standart kullanıcı hesaplarını yönetebilecek Admin yetkisine yükseltilecek.'
        "
        :changes="[
          {
            label: 'Rol',
            from: targetUser.roles.includes('admin') ? 'ADMİN' : 'KULLANICI',
            to: targetUser.roles.includes('admin') ? 'KULLANICI' : 'ADMİN'
          }
        ]"
        :confirm-label="targetUser.roles.includes('admin') ? 'Yetkiyi kaldır' : 'Evet, Admin yap'"
        variant="danger"
        :loading="saving"
        @cancel="confirmRoleOpen = false"
        @confirm="confirmRoleToggle"
      />
    </template>
  </div>
</template>
