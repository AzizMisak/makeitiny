<script setup lang="ts">
import type { ConfirmChange } from '~/composables/useConfirm'
import type { PlanItem, User } from '~/types/auth'
import type { AdminLinkItem } from '~/types/link'

definePageMeta({ layout: 'admin', middleware: 'auth', roles: ['admin', 'owner'] })

type BanLength = '1d' | '7d' | '30d' | 'permanent' | 'custom'

const route = useRoute()
const { isOwner, isSelf } = useAuth()
const { fetchPlans, fetchAdminUserById, fetchAdminUserLinks, adminUpdateUser, adminUpdateUserRole } = useAdmin()
const { ask } = useConfirm()
const toast = useToast()
const { busy: exporting, exportAdminUserLinks } = useExport()

const id = computed(() => String(route.params.id ?? ''))
const person = ref<User | null>(null)
const plans = ref<PlanItem[]>([])
const links = ref<AdminLinkItem[]>([])
const loading = ref(true)
const linksLoading = ref(true)
const linksError = ref('')
const error = ref('')

const planId = ref(1)
const planExpires = ref('')
const planError = ref('')

const banReason = ref('')
const banLength = ref<BanLength>('7d')
const banDate = ref('')
const banError = ref('')

watch([planExpires, planId], () => (planError.value = ''))
watch([banLength, banDate], () => (banError.value = ''))

const BAN_OPTIONS: { value: BanLength; label: string }[] = [
  { value: '1d', label: '1 gün' },
  { value: '7d', label: '7 gün' },
  { value: '30d', label: '30 gün' },
  { value: 'permanent', label: 'Süresiz' },
  { value: 'custom', label: 'Tarih seç' }
]

useHead({ title: () => `${person.value?.name || person.value?.email || 'Kullanıcı'} | Yönetim | makeiTiny` })

function sync(u: User) {
  person.value = u
  planId.value = u.plan?.id ?? 1
  planExpires.value = u.planExpiresAt ?? ''
}

async function loadLinks() {
  const u = person.value
  if (!u) return
  linksLoading.value = true
  linksError.value = ''
  try {
    links.value = await fetchAdminUserLinks(u.id, u.email)
  } catch (e: any) {
    linksError.value = e?.message ?? 'Linkler yüklenemedi.'
  } finally {
    linksLoading.value = false
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [found, list] = await Promise.all([fetchAdminUserById(id.value), fetchPlans()])
    plans.value = list
    if (found) {
      sync(found)
      loadLinks()
    }
  } catch (e: any) {
    error.value = e?.message ?? 'Kullanıcı bilgileri yüklenemedi.'
  } finally {
    loading.value = false
  }
}

onMounted(load)

const isTargetOwner = computed(() => !!person.value?.roles.includes('owner'))
const isTargetAdmin = computed(() => !!person.value?.roles.includes('admin'))
const self = computed(() => isSelf(person.value?.id))
const planless = computed(() => isPrivileged(person.value?.roles))
const canManage = computed(() => !self.value && (isOwner.value || (!isTargetOwner.value && !isTargetAdmin.value)))
const canBan = computed(() => canManage.value && !isTargetOwner.value)
const role = computed(() => roleMeta(person.value?.roles))
const who = computed(() => (person.value ? `${person.value.name || 'İsimsiz'} (${person.value.email})` : ''))

const facts = computed(() => {
  const u = person.value
  if (!u) return []
  return [
    { label: 'Link', value: formatNumber(u.linkCount) },
    ...(isPrivileged(u.roles)
      ? [{ label: 'Paket', value: 'Pakete tabi değil' }]
      : [
          { label: 'Paket', value: u.plan.name },
          { label: 'Paket bitişi', value: u.planExpiresAt ? formatDateTr(u.planExpiresAt) : 'Süresiz' }
        ]),
    { label: 'Kayıt', value: formatDateTr(u.createdAt) },
    { label: 'Kimlik', value: u.id }
  ]
})

const planOptions = computed(() => plans.value.map((p) => ({ value: p.id, label: `${p.name}, ${formatPriceMonthly(p.priceMonthly)}` })))

const planChanges = computed<ConfirmChange[]>(() => {
  const u = person.value
  if (!u) return []
  const list: ConfirmChange[] = []
  if (planId.value !== u.plan.id) {
    list.push({ label: 'Paket', from: u.plan.name, to: plans.value.find((p) => p.id === planId.value)?.name ?? '—' })
  }
  const was = u.planExpiresAt ?? ''
  if (planExpires.value !== was) {
    list.push({ label: 'Bitiş', from: was ? formatDateTr(was) : 'Süresiz', to: planExpires.value ? formatDateTr(planExpires.value) : 'Süresiz' })
  }
  return list
})

useUnsavedGuard(() => !planless.value && planChanges.value.length > 0)

async function savePlan() {
  const u = person.value
  if (!u || planless.value || planChanges.value.length === 0) return
  planError.value = ''
  if (planExpires.value && planExpires.value < todayIso()) {
    planError.value = 'Bitiş tarihi bugünden önce olamaz.'
    return
  }
  const ok = await ask({
    title: 'Paket değiştirilsin mi?',
    subject: who.value,
    message: 'Kullanıcının hakları hemen değişir.',
    changes: planChanges.value,
    confirmLabel: 'Paketi değiştir',
    action: async () => sync(await adminUpdateUser(u.id, { planId: planId.value, planExpiresAt: planExpires.value }))
  })
  if (ok) toast.success('Paket değiştirildi.')
}

function computeBanUntil(): string {
  if (banLength.value === 'permanent') return ''
  if (banLength.value === 'custom') return banDate.value ? new Date(`${banDate.value}T23:59:59`).toISOString() : ''
  const days = { '1d': 1, '7d': 7, '30d': 30 }[banLength.value]
  return new Date(Date.now() + days * 86_400_000).toISOString()
}

async function ban() {
  const u = person.value
  if (!u || !canBan.value) return
  banError.value = ''
  if (banLength.value === 'custom' && !banDate.value) {
    banError.value = 'Bir bitiş tarihi seç.'
    return
  }
  const until = computeBanUntil()
  const ok = await ask({
    title: 'Hesap askıya alınsın mı?',
    subject: who.value,
    message: `${until ? `${formatDateTimeTr(until)} tarihine kadar` : 'Süresiz olarak'} askıda kalır.`,
    effects: [
      'Açık oturumları hemen kapanır, giriş yapamaz.',
      'Linkleri yönlendirmeyi bırakır.',
      banReason.value.trim() ? `Gösterilecek sebep: ${banReason.value.trim()}` : 'Sebep yazılmadı.'
    ],
    confirmLabel: 'Hesabı askıya al',
    tone: 'danger',
    action: async () => sync(await adminUpdateUser(u.id, { isActive: false, banReason: banReason.value.trim(), bannedUntil: until }))
  })
  if (ok) {
    banReason.value = ''
    toast.success('Hesap askıya alındı.')
  }
}

async function unban() {
  const u = person.value
  if (!u) return
  const ok = await ask({
    title: 'Askı kaldırılsın mı?',
    subject: who.value,
    message: 'Kullanıcı yeniden giriş yapabilir, linkleri yönlendirmeye başlar.',
    confirmLabel: 'Askıyı kaldır',
    action: async () => sync(await adminUpdateUser(u.id, { isActive: true, banReason: '', bannedUntil: '' }))
  })
  if (ok) toast.success('Askı kaldırıldı.')
}

async function toggleRole() {
  const u = person.value
  if (!u || !isOwner.value || isTargetOwner.value || self.value) return
  const next = isTargetAdmin.value ? 'user' : 'admin'
  const ok = await ask(roleChangePrompt(u, next, async () => sync(await adminUpdateUserRole(u.id, next))))
  if (ok) toast.success(next === 'admin' ? 'Yönetici yapıldı.' : 'Yönetici yetkisi kaldırıldı.')
}
const exportLinksMenu = computed(() => [
  {
    label: 'CSV olarak indir',
    icon: 'download',
    disabled: !!exporting.value || !person.value,
    onSelect: () => person.value && exportAdminUserLinks(person.value.id, 'csv')
  },
  {
    label: 'JSON olarak indir',
    icon: 'download',
    disabled: !!exporting.value || !person.value,
    onSelect: () => person.value && exportAdminUserLinks(person.value.id, 'json')
  }
])
</script>

<template>
  <div>
    <NuxtLink to="/dashboard/admin/users" class="mb-5 inline-flex items-center gap-2 text-sm font-medium text-ink-2 hover:text-ink">
      <UiIcon name="back" :size="16" /> Kullanıcılar
    </NuxtLink>

    <div v-if="loading" class="grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,1fr)_24rem]" aria-busy="true" aria-label="Kullanıcı yükleniyor">
      <div><div class="skeleton h-14 w-80 max-w-full" /><div class="skeleton mt-8 h-72" /></div>
      <div class="skeleton h-96" />
    </div>
    <UiLoadError v-else-if="error" :message="error" @retry="load" />
    <UiEmpty v-else-if="!person" icon="user" title="Kullanıcı bulunamadı" text="Hesap silinmiş olabilir." class="border border-line bg-surface">
      <UiButton to="/dashboard/admin/users" variant="outline">Kullanıcılara dön</UiButton>
    </UiEmpty>

    <div v-else class="grid grid-cols-1 items-start gap-x-8 gap-y-10 lg:grid-cols-[minmax(0,1fr)_24rem]">
      <div class="min-w-0">
        <header class="flex items-center gap-4">
          <AdminAvatar :name="person.name" :email="person.email" large />
          <div class="min-w-0">
            <h1 class="display truncate text-[1.75rem] leading-[1.3]">{{ person.name || 'İsimsiz' }}</h1>
            <p class="flex flex-wrap items-center gap-x-4 gap-y-1 text-ink-2">
              <span class="truncate">{{ person.email }}</span>
              <UiTag :tone="role.tone">{{ role.label }}</UiTag>
              <UiTag v-if="self" tone="mark" icon="user">Bu sensin</UiTag>
              <UiStatus :tone="person.isActive ? 'ok' : 'danger'" :label="person.isActive ? 'Aktif' : 'Askıda'" />
            </p>
          </div>
        </header>

        <dl class="mt-7 flex flex-wrap gap-x-10 gap-y-4">
          <div v-for="f in facts" :key="f.label">
            <dt class="text-[13px] text-ink-3">{{ f.label }}</dt>
            <dd class="tnum mt-0.5 font-semibold">{{ f.value }}</dd>
          </div>
        </dl>

        <div class="mb-4 mt-10 flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-lg font-semibold">Linkleri</h2>
          <UiMenu v-if="person.linkCount > 0" label="Kullanıcının linklerini dışa aktar" :items="exportLinksMenu">
            <template #trigger="{ toggle, open }">
              <UiButton
                variant="outline"
                size="sm"
                icon="download"
                :loading="exporting.startsWith('admin-user-links-')"
                loading-label="Hazırlanıyor"
                :aria-expanded="open"
                aria-haspopup="menu"
                @click="toggle"
              >
                Dışa aktar
              </UiButton>
            </template>
          </UiMenu>
        </div>
        <UiLoadError v-if="linksError" :message="linksError" @retry="loadLinks" />
        <AdminTableSkeleton v-else-if="linksLoading && links.length === 0" :rows="4" label="Kullanıcının linkleri yükleniyor" />
        <UiEmpty v-else-if="links.length === 0" title="Link oluşturmamış" class="border border-line bg-surface" />
        <template v-else>
          <AdminLinkTable :links="links" :loading="linksLoading" hide-owner actions caption="Kullanıcının linkleri" @changed="loadLinks" />
          <p v-if="(person.linkCount ?? 0) > links.length" class="mt-3 text-sm text-ink-2">
            En yeni 100 link gösteriliyor; tümünü görmek için
            <NuxtLink :to="`/dashboard/admin/links?search=${encodeURIComponent(person.email)}`" class="font-medium text-brand underline underline-offset-4 hover:text-ink">
              Linkler sayfasına git
            </NuxtLink>.
          </p>
        </template>
      </div>

      <div class="flex flex-col divide-y divide-line border border-line bg-surface">
        <div v-if="self" class="bg-mark/15 px-5 py-4 text-sm">
          <p class="font-semibold">Kendi hesabına bakıyorsun.</p>
          <p class="mt-1 text-ink-2">Kendi rolünü, durumunu ve paketini buradan değiştiremezsin. Adını ve parolanı Hesap ayarlarından değiştirebilirsin.</p>
        </div>
        <p v-else-if="!canManage" class="flex items-center gap-3 px-5 py-4 text-sm text-ink-2">
          <UiIcon name="lock" :size="17" /> Yönetici hesaplarını yalnızca sistem sahibi değiştirebilir.
        </p>

        <section v-if="planless" class="p-5">
          <h2 class="text-lg font-semibold">Paket</h2>
          <p class="mt-2 text-sm text-ink-2">{{ PRIVILEGED_PLAN_NOTE }} Bu hesaba paket atanamaz.</p>
          <p v-if="!isTargetOwner" class="mt-2 text-sm text-ink-3">Kayıtlı paketi ({{ person.plan.name }}) yönetici yetkisi kaldırılırsa yeniden geçerli olur.</p>
        </section>

        <form v-else class="p-5" novalidate @submit.prevent="savePlan">
          <h2 class="text-lg font-semibold">Paket</h2>
          <UiSelect id="user-plan" v-model="planId" class="mt-3" :options="planOptions" :disabled="!canManage" />
          <UiField id="user-plan-exp" v-model="planExpires" class="mt-3" label="Bitiş tarihi" type="date" :min="todayIso()" :error="planError" hint="Boşsa süresiz" :disabled="!canManage" />
          <UiButton type="submit" variant="outline" size="sm" class="mt-2" :disabled="!canManage || planChanges.length === 0">
            {{ planChanges.length === 0 ? 'Değişiklik yok' : 'Paketi değiştir' }}
          </UiButton>
        </form>

        <section class="p-5">
          <h2 class="text-lg font-semibold">Hesap durumu</h2>
          <p v-if="!canBan" class="mt-2 text-sm text-ink-2">Bu hesap askıya alınamaz.</p>

          <form v-else-if="person.isActive" class="mt-3" novalidate @submit.prevent="ban">
            <UiField id="ban-reason" v-model="banReason" label="Askı sebebi" placeholder="Spam veya zararlı link" hint="Kullanıcı giriş yaparken görür." />
            <p class="mb-1.5 mt-2 text-sm font-semibold">Süre</p>
            <UiSelect id="ban-length" v-model="banLength" :options="BAN_OPTIONS" />
            <!-- Tarih alanı hep yerinde; yalnızca "Tarih seç" ile açılır. -->
            <UiField
              id="ban-date"
              v-model="banDate"
              class="mt-3"
              type="date"
              :min="todayIso()"
              :disabled="banLength !== 'custom'"
              :error="banError"
              :hint="banLength === 'custom' ? 'Bu günün sonunda kalkar.' : 'Tarih seçmek için süreyi değiştir.'"
            />
            <UiButton type="submit" variant="danger" size="sm" icon="ban" class="-ml-3 mt-1">Hesabı askıya al</UiButton>
          </form>

          <div v-else class="mt-3">
            <p class="text-sm text-ink-2">
              {{ person.bannedUntil ? `${formatDateTimeTr(person.bannedUntil)} tarihine kadar askıda.` : 'Süresiz askıda.' }}
            </p>
            <p class="mt-1 text-sm text-ink-2">Sebep: {{ person.banReason || 'yazılmamış' }}</p>
            <UiButton variant="outline" size="sm" icon="unlock" class="mt-4" @click="unban">Askıyı kaldır</UiButton>
          </div>
        </section>

        <section v-if="isOwner" class="p-5">
          <h2 class="text-lg font-semibold">Rol</h2>
          <p class="mt-2 text-sm text-ink-2">
            <template v-if="isTargetOwner">Sistemin tek sahibi. Bu rol değiştirilemez.</template>
            <template v-else-if="isTargetAdmin">Bütün linkleri ve standart kullanıcıları yönetebilir.</template>
            <template v-else>Yalnızca kendi linklerini yönetebilir.</template>
          </p>
          <p v-if="self && !isTargetOwner" class="mt-2 text-sm text-ink-3">Kendi rolünü değiştiremezsin.</p>
          <UiButton v-if="!isTargetOwner && !self" :variant="isTargetAdmin ? 'danger' : 'outline'" size="sm" icon="key" class="mt-4" :class="isTargetAdmin ? '-ml-3' : ''" @click="toggleRole">
            {{ isTargetAdmin ? 'Yönetici yetkisini kaldır' : 'Yönetici yap' }}
          </UiButton>
        </section>
      </div>
    </div>
  </div>
</template>
