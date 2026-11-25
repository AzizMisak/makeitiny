<script setup lang="ts">
import type { ConfirmChangeItem } from '~/components/ui/ConfirmModal.vue'
import type { AdminLinkItem, LinkLogItem, SecurityStatus } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth', roles: ['admin', 'owner'] })

const route = useRoute()
const { public: config } = useRuntimeConfig()
const {
  fetchAdminLinkById,
  adminUpdateLink,
  adminDeleteLink,
  fetchAdminLinkLogs
} = useAdmin()
const toast = useToast()

const linkId = computed(() => String(route.params.id ?? ''))
const link = ref<AdminLinkItem | null>(null)
const pageLoading = ref(true)
const saving = ref(false)
const deleting = ref(false)

const editTarget = ref('')
const editTitle = ref('')
const editIsActive = ref(true)
const editSecurityStatus = ref<SecurityStatus>('unchecked')
const editExpiresAt = ref('')
const editMaxClicks = ref('')
const editAdEnabled = ref(false)

const confirmSaveOpen = ref(false)
const confirmDeleteOpen = ref(false)
const confirmBanOpen = ref(false)

const LOGS_PER_PAGE = 15
const logs = ref<LinkLogItem[]>([])
const logsLoading = ref(false)
const logsPage = ref(1)
const logsTotal = ref(0)
const logsTotalPages = ref(1)

useHead({
  title: computed(() => (link.value ? `Link İnceleme (${link.value.alias}) — makeiTiny` : 'Link İnceleme — makeiTiny'))
})

const fullShortUrl = computed(() => (link.value ? `${config.routerBase}/${link.value.alias}` : ''))
const displayShortUrl = computed(() => fullShortUrl.value.replace(/^https?:\/\//, ''))

function syncForm(data: AdminLinkItem) {
  link.value = data
  editTarget.value = data.target
  editTitle.value = data.title ?? ''
  editIsActive.value = data.isActive
  editSecurityStatus.value = data.securityStatus
  editExpiresAt.value = data.expiresAt ?? ''
  editMaxClicks.value = data.maxClicks ? String(data.maxClicks) : ''
  editAdEnabled.value = data.adEnabled
}

const pendingChanges = computed<ConfirmChangeItem[]>(() => {
  const cur = link.value
  if (!cur) return []
  const list: ConfirmChangeItem[] = []

  if (editIsActive.value !== cur.isActive) {
    list.push({
      label: 'Durum',
      from: cur.isActive ? 'Aktif' : 'Pasif',
      to: editIsActive.value ? 'Aktif' : 'Pasif'
    })
  }
  if (editSecurityStatus.value !== cur.securityStatus) {
    list.push({
      label: 'Güvenlik',
      from: securityStatusMeta(cur.securityStatus).label,
      to: securityStatusMeta(editSecurityStatus.value).label
    })
  }
  if (editTarget.value.trim() !== cur.target) {
    list.push({
      label: 'Hedef adres',
      from: cur.target,
      to: editTarget.value.trim()
    })
  }
  if (editTitle.value.trim() !== (cur.title ?? '')) {
    list.push({
      label: 'Başlık',
      from: cur.title || '—',
      to: editTitle.value.trim() || '—'
    })
  }
  const curExp = cur.expiresAt ?? ''
  const nextExp = editExpiresAt.value.trim()
  if (nextExp !== curExp) {
    list.push({
      label: 'Son kullanma',
      from: curExp || 'Süresiz',
      to: nextExp || 'Süresiz'
    })
  }
  const curMax = cur.maxClicks ? String(cur.maxClicks) : ''
  const nextMax = String(editMaxClicks.value ?? '').trim()
  if (nextMax !== curMax) {
    list.push({
      label: 'Tıklama limiti',
      from: curMax || 'Sınırsız',
      to: nextMax || 'Sınırsız'
    })
  }
  if (editAdEnabled.value !== cur.adEnabled) {
    list.push({
      label: 'Reklam modu',
      from: cur.adEnabled ? 'Açık' : 'Kapalı',
      to: editAdEnabled.value ? 'Açık' : 'Kapalı'
    })
  }

  return list
})

async function loadLogs(targetPage = 1) {
  if (!linkId.value) return
  logsLoading.value = true
  try {
    const res = await fetchAdminLinkLogs(linkId.value, targetPage, LOGS_PER_PAGE)
    logs.value = res.items
    logsPage.value = res.page
    logsTotal.value = res.total
    logsTotalPages.value = Math.max(1, res.totalPages)
  } catch (err: any) {
    toast.error(err?.message ?? 'Loglar yüklenemedi.')
  } finally {
    logsLoading.value = false
  }
}

async function loadPage() {
  pageLoading.value = true
  try {
    const found = await fetchAdminLinkById(linkId.value)
    if (found) {
      syncForm(found)
      await loadLogs(1)
    }
  } catch (err: any) {
    toast.error(err?.message ?? 'Link bilgileri alınamadı.')
  } finally {
    pageLoading.value = false
  }
}

function requestSave() {
  if (!editTarget.value.trim()) {
    toast.error('Hedef adres boş bırakılamaz.')
    return
  }
  const rawMax = String(editMaxClicks.value ?? '').trim()
  if (rawMax !== '') {
    const n = Number.parseInt(rawMax, 10)
    if (!Number.isFinite(n) || n < 1) {
      toast.error('Maksimum tıklama pozitif bir sayı olmalı.')
      return
    }
  }
  if (pendingChanges.value.length === 0) {
    toast.info('Kaydedilecek bir değişiklik yok.')
    return
  }
  confirmSaveOpen.value = true
}

async function confirmSave() {
  if (!link.value) return
  const rawMax = String(editMaxClicks.value ?? '').trim()
  const parsedMax = rawMax !== '' ? Number.parseInt(rawMax, 10) : 0

  saving.value = true
  try {
    const updated = await adminUpdateLink(link.value.id, {
      target: editTarget.value.trim(),
      title: editTitle.value.trim(),
      isActive: editIsActive.value,
      securityStatus: editSecurityStatus.value,
      expiresAt: editExpiresAt.value.trim(),
      maxClicks: parsedMax,
      adEnabled: editAdEnabled.value
    })
    syncForm({
      ...link.value,
      ...updated
    })
    confirmSaveOpen.value = false
    toast.success('Link ayarları güncellendi.')
  } catch (err: any) {
    toast.error(err?.message ?? 'Link güncellenemedi.')
  } finally {
    saving.value = false
  }
}

async function confirmBanToggle() {
  if (!link.value) return
  const currentlyBlocked = link.value.securityStatus === 'blocked' || link.value.securityStatus === 'malicious'
  saving.value = true
  try {
    const updated = await adminUpdateLink(link.value.id, {
      securityStatus: currentlyBlocked ? 'safe' : 'blocked',
      isActive: currentlyBlocked ? true : false
    })
    syncForm({ ...link.value, ...updated })
    confirmBanOpen.value = false
    toast.success(currentlyBlocked ? 'Link yasağı kaldırıldı ve erişime açıldı.' : 'Link yasaklandı ve erişime kapatıldı.')
  } catch (err: any) {
    toast.error(err?.message ?? 'İşlem başarısız.')
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!link.value) return
  deleting.value = true
  try {
    await adminDeleteLink(link.value.id)
    confirmDeleteOpen.value = false
    toast.success(`${link.value.alias} kalıcı olarak silindi.`)
    await navigateTo('/dashboard/admin/links')
  } catch (err: any) {
    toast.error(err?.message ?? 'Link silinemedi.')
  } finally {
    deleting.value = false
  }
}

onMounted(loadPage)
</script>

<template>
  <div>
    <div class="mb-5">
      <NuxtLink
        to="/dashboard/admin/links"
        class="inline-flex items-center gap-2 text-sm font-medium text-ink-2 hover:text-ink"
      >
        <UiIcon name="back" /> Tüm linklere dön
      </NuxtLink>
    </div>

    <div v-if="pageLoading" class="flex flex-col gap-4">
      <div class="h-16 animate-pulse rounded-lg bg-paper-2" />
      <div class="h-64 animate-pulse rounded-lg bg-paper-2" />
    </div>

    <div
      v-else-if="!link"
      class="rounded-lg border-[1.5px] border-ink bg-card p-8 text-center shadow-hard"
    >
      <h1 class="text-xl font-bold">Link bulunamadı</h1>
      <p class="mt-1 text-sm text-ink-3">Bu link silinmiş veya geçersiz bir ID ile açılmış olabilir.</p>
      <div class="mt-5">
        <UiButton to="/dashboard/admin/links" variant="secondary">Listeye dön</UiButton>
      </div>
    </div>

    <template v-else>
      <DashboardPageHeader :title="displayShortUrl" :description="link.target">
        <template #actions>
          <UiButton
            v-if="link.userId && link.userId !== '0'"
            :to="`/dashboard/admin/users/${link.userId}`"
            variant="secondary"
          >
            <UiIcon name="users" :size="15" /> Sahibini incele
          </UiButton>
          <a
            :href="fullShortUrl"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center justify-center gap-2 rounded-md border-[1.5px] border-ink bg-card px-4 py-2.5 text-sm font-semibold text-ink shadow-hard transition-all hover:bg-paper-2"
          >
            Ziyaret et <UiIcon name="arrow" :size="15" />
          </a>
        </template>
      </DashboardPageHeader>

      <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
        <DashboardStatTile
          label="Durum"
          :value="link.isActive ? 'Aktif' : 'Pasif'"
          :note="`Güvenlik: ${securityStatusMeta(link.securityStatus).label}`"
        />
        <DashboardStatTile
          label="Toplam tıklama"
          :value="link.clicks.toLocaleString('tr-TR')"
          :note="link.maxClicks ? `Limit: ${link.maxClicks}` : 'Limitsiz'"
        />
        <DashboardStatTile
          label="Sahip"
          :value="link.userName || '—'"
          :note="link.userEmail"
        />
        <DashboardStatTile
          label="Oluşturulma"
          :value="formatDateTr(link.createdAt)"
          :note="link.expiresAt ? `Bitiş: ${formatDateTr(link.expiresAt)}` : 'Süresiz'"
        />
      </div>

      <div class="mt-8 grid gap-6 lg:grid-cols-[1.45fr_1fr]">
        <form
          class="flex flex-col gap-5 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8"
          @submit.prevent="requestSave"
        >
          <div class="flex flex-wrap items-center justify-between gap-2 border-b border-line pb-4">
            <h2 class="text-xl font-bold">Moderasyon & link ayarları</h2>
            <span class="font-mono text-xs text-ink-3">ID: {{ link.id }}</span>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label for="mod-status" class="mb-1.5 block text-sm font-semibold text-ink">Yönlendirme durumu</label>
              <select
                id="mod-status"
                v-model="editIsActive"
                class="w-full rounded-md border-[1.5px] border-ink bg-card px-3 py-2.5 text-sm font-semibold text-ink focus:outline-none focus:ring-2 focus:ring-accent"
              >
                <option :value="true">Aktif (Yönlendirmeye açık)</option>
                <option :value="false">Pasif (Erişime kapalı)</option>
              </select>
            </div>

            <div>
              <label for="mod-sec" class="mb-1.5 block text-sm font-semibold text-ink">Güvenlik durumu</label>
              <select
                id="mod-sec"
                v-model="editSecurityStatus"
                class="w-full rounded-md border-[1.5px] border-ink bg-card px-3 py-2.5 text-sm font-semibold text-ink focus:outline-none focus:ring-2 focus:ring-accent"
              >
                <option v-for="opt in SECURITY_STATUS_OPTIONS" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </div>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <UiField
              id="mod-target"
              v-model="editTarget"
              label="Hedef adres"
              type="text"
            />
            <UiField
              id="mod-title"
              v-model="editTitle"
              label="Başlık"
              type="text"
              placeholder="İsteğe bağlı"
            />
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <UiField
              id="mod-expires"
              v-model="editExpiresAt"
              label="Son kullanma tarihi"
              type="date"
            />
            <UiField
              id="mod-maxclicks"
              v-model="editMaxClicks"
              label="Maksimum tıklama"
              type="number"
              placeholder="Sınırsız"
            />
          </div>

          <div class="flex items-center justify-between rounded-md border border-line bg-paper px-4 py-3">
            <span class="text-sm font-semibold text-ink">Reklam gösterimi</span>
            <label class="inline-flex cursor-pointer items-center gap-2 select-none">
              <input v-model="editAdEnabled" type="checkbox" class="h-4 w-4 accent-accent" />
              <span class="text-xs font-semibold">{{ editAdEnabled ? 'Açık' : 'Kapalı' }}</span>
            </label>
          </div>

          <div class="flex flex-wrap items-center justify-between gap-3 border-t border-line pt-4">
            <span class="text-xs text-ink-3">
              {{ pendingChanges.length > 0 ? `${pendingChanges.length} alan değiştirildi` : 'Değişiklik yok' }}
            </span>
            <UiButton type="submit" :disabled="pendingChanges.length === 0 || saving">
              Değişiklikleri kaydet
            </UiButton>
          </div>
        </form>

        <div class="flex flex-col gap-6">
          <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
            <h2 class="text-lg font-bold">Bağlantı özeti</h2>
            <dl class="mt-4 divide-y divide-line text-sm">
              <div class="flex items-center justify-between py-2.5">
                <dt class="text-ink-3">Kısa kod (Alias)</dt>
                <dd class="font-mono font-bold text-ink">{{ link.alias }}</dd>
              </div>
              <div class="flex items-center justify-between py-2.5">
                <dt class="text-ink-3">Şifre kilidi</dt>
                <dd class="font-semibold">{{ link.locked ? 'Var' : 'Yok' }}</dd>
              </div>
              <div class="flex items-center justify-between py-2.5">
                <dt class="text-ink-3">Ara bekletme</dt>
                <dd class="font-semibold">{{ link.waitEnabled ? `${link.waitSeconds} sn` : 'Kapalı' }}</dd>
              </div>
              <div class="flex items-center justify-between py-2.5">
                <dt class="text-ink-3">Sahip e-posta</dt>
                <dd class="truncate font-mono text-xs font-semibold">{{ link.userEmail }}</dd>
              </div>
            </dl>
          </div>

          <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
            <h2 class="text-lg font-bold">Hızlı Moderasyon</h2>
            <p class="mt-1.5 text-sm text-ink-2">
              {{
                link.securityStatus === 'blocked' || link.securityStatus === 'malicious'
                  ? 'Bu link şu an güvenlik nedeniyle engelli ve kapalı.'
                  : 'Zararlı veya kural dışı bir linki tek tıkla yasaklayıp kapatabilirsin.'
              }}
            </p>
            <div class="mt-4">
              <UiButton
                :variant="link.securityStatus === 'blocked' || link.securityStatus === 'malicious' ? 'secondary' : 'primary'"
                @click="confirmBanOpen = true"
              >
                <UiIcon name="ban" :size="15" />
                {{
                  link.securityStatus === 'blocked' || link.securityStatus === 'malicious'
                    ? 'Yasağı Kaldır ve Aç'
                    : 'Linki Yasakla ve Kapat'
                }}
              </UiButton>
            </div>
          </div>

          <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
            <h2 class="text-lg font-bold text-accent-dark">Linki sil</h2>
            <p class="mt-1.5 text-sm text-ink-2">
              Bu linki sildiğinde yönlendirme durur ve tüm tıklama kayıtları kalıcı olarak kaldırılır.
            </p>
            <div class="mt-4">
              <UiButton variant="danger" @click="confirmDeleteOpen = true">
                <UiIcon name="trash" :size="15" /> Linki kalıcı sil...
              </UiButton>
            </div>
          </div>
        </div>
      </div>

      <!-- Tıklama Logları -->
      <section class="mt-8 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-line pb-4">
          <div class="flex items-center gap-2">
            <h2 class="text-xl font-bold">Tıklama logları</h2>
            <UiBadge tone="neutral">{{ logsTotal.toLocaleString('tr-TR') }} kayıt</UiBadge>
          </div>
          <UiButton variant="secondary" :disabled="logsLoading" @click="loadLogs(logsPage)">
            <UiIcon name="refresh" :size="14" /> Yenile
          </UiButton>
        </div>

        <p v-if="logsLoading" class="py-8 text-center text-sm text-ink-3">Loglar yükleniyor...</p>
        <p v-else-if="logs.length === 0" class="py-8 text-center text-sm text-ink-3">Bu link için tıklama kaydı yok.</p>

        <div v-else class="mt-4 overflow-x-auto rounded-md border border-line">
          <table class="w-full text-left text-xs">
            <thead class="border-b border-line bg-paper text-[11px] uppercase tracking-wide text-ink-3">
              <tr>
                <th class="px-3.5 py-2.5">Tarih</th>
                <th class="px-3.5 py-2.5">IP Adresi</th>
                <th class="px-3.5 py-2.5">Ülke</th>
                <th class="px-3.5 py-2.5">Cihaz / Tarayıcı / İS</th>
                <th class="px-3.5 py-2.5">Yönlendiren (Referer)</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-line">
              <tr v-for="log in logs" :key="log.id" class="hover:bg-paper/60">
                <td class="whitespace-nowrap px-3.5 py-2.5 font-mono">{{ formatDateTimeTr(log.clickedAt) }}</td>
                <td class="whitespace-nowrap px-3.5 py-2.5 font-mono">{{ log.ipAddress || '—' }}</td>
                <td class="whitespace-nowrap px-3.5 py-2.5 font-mono font-bold">{{ log.countryCode || '—' }}</td>
                <td class="px-3.5 py-2.5">{{ [log.deviceType, log.browser, log.os].filter(Boolean).join(' • ') || '—' }}</td>
                <td class="max-w-xs truncate px-3.5 py-2.5 text-ink-3" :title="log.referer">{{ log.referer || 'Doğrudan' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <DashboardPaginationBar
          v-if="logsTotal > 0"
          :page="logsPage"
          :total-pages="logsTotalPages"
          :total="logsTotal"
          :page-size="LOGS_PER_PAGE"
          :loading="logsLoading"
          item-label="log kaydı"
          @change="loadLogs"
        />
      </section>

      <!-- Kaydetme Onay Modalı -->
      <UiConfirmModal
        :open="confirmSaveOpen"
        title="Link ayarlarını güncelle"
        :target-label="`${displayShortUrl} (${link.userEmail})`"
        description="Aşağıdaki moderasyon ve yapılandırma değişiklikleri uygulanacak:"
        :changes="pendingChanges"
        confirm-label="Onayla ve kaydet"
        :loading="saving"
        @cancel="confirmSaveOpen = false"
        @confirm="confirmSave"
      />

      <!-- Hızlı Yasaklama Onay Modalı -->
      <UiConfirmModal
        :open="confirmBanOpen"
        :title="link.securityStatus === 'blocked' || link.securityStatus === 'malicious' ? 'Link yasağını kaldır' : 'Linki yasakla ve kapat'"
        :target-label="`${displayShortUrl} → ${link.target}`"
        :description="
          link.securityStatus === 'blocked' || link.securityStatus === 'malicious'
            ? 'Link güvenlik durumu güvenli yapılacak ve tekrar yönlendirmeye açılacak.'
            : 'Link güvenlik nedeniyle engellenecek ve erişime kapatılacak.'
        "
        :confirm-label="link.securityStatus === 'blocked' || link.securityStatus === 'malicious' ? 'Evet, yasağı kaldır' : 'Evet, yasakla'"
        :variant="link.securityStatus === 'blocked' || link.securityStatus === 'malicious' ? 'primary' : 'danger'"
        :loading="saving"
        @cancel="confirmBanOpen = false"
        @confirm="confirmBanToggle"
      />

      <!-- Silme Onay Modalı -->
      <UiConfirmModal
        :open="confirmDeleteOpen"
        title="Linki kalıcı olarak sil"
        :target-label="`${displayShortUrl} → ${link.target}`"
        description="Bu link ve bağlı tüm tıklama geçmişi kalıcı olarak silinecek. Bu işlem geri alınamaz."
        confirm-label="Evet, kalıcı sil"
        variant="danger"
        :loading="deleting"
        @cancel="confirmDeleteOpen = false"
        @confirm="confirmDelete"
      />
    </template>
  </div>
</template>
