<script setup lang="ts">
import type { LinkItem, LinkLogItem } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const route = useRoute()
const { public: config } = useRuntimeConfig()
const { getLink, getLinkLogs, updateLink, removeLink } = useLinks()

const id = computed(() => String(route.params.id ?? ''))
const link = ref<LinkItem | null>(null)
const logs = ref<LinkLogItem[]>([])
const logsLoading = ref(false)
const logsError = ref('')
const pageLoading = ref(true)
const pageError = ref('')

const editTarget = ref('')
const editAlias = ref('')
const newPassword = ref('')
const removePasswordFlag = ref(false)
const editWaitEnabled = ref(false)
const editWaitSeconds = ref('5')

const saving = ref(false)
const saveSuccess = ref(false)
const saveError = ref('')
const copied = ref(false)
const confirmDelete = ref(false)

useHead({
  title: () => (link.value ? `${link.value.alias} — Link Detayı — makeiTiny` : 'Link Detayı — makeiTiny')
})

const fullShortUrl = computed(() => (link.value ? `${config.routerBase}/${link.value.alias}` : ''))
const displayShortUrl = computed(() => fullShortUrl.value.replace(/^https?:\/\//, ''))

// Şifre aktifse (mevcut şifre korunuyorsa veya yeni şifre yazılıyorsa) bekletme açılamaz.
const isPasswordActive = computed(() => {
  const hasExistingPassword = Boolean(link.value?.locked && !removePasswordFlag.value)
  const hasNewPassword = newPassword.value.trim().length > 0
  return hasExistingPassword || hasNewPassword
})

watch(isPasswordActive, (active) => {
  if (active) {
    editWaitEnabled.value = false
  }
})

const accessModeLabel = computed(() => {
  if (!link.value) return ''
  if (link.value.locked) return 'Şifreli'
  if (link.value.waitEnabled) return `Bekletmeli (${link.value.waitSeconds} sn)`
  return 'Herkese Açık'
})

async function loadLogs() {
  logsLoading.value = true
  logsError.value = ''
  try {
    logs.value = await getLinkLogs(id.value)
  } catch (e: any) {
    logsError.value = e?.message ?? 'Tıklama logları yüklenemedi.'
  } finally {
    logsLoading.value = false
  }
}

async function loadLink() {
  pageLoading.value = true
  pageError.value = ''
  try {
    const data = await getLink(id.value)
    link.value = data
    editTarget.value = data.target
    editAlias.value = data.alias
    newPassword.value = ''
    removePasswordFlag.value = false
    editWaitEnabled.value = !data.locked && data.waitEnabled
    editWaitSeconds.value = String(data.waitSeconds > 0 ? data.waitSeconds : 5)
    await loadLogs()
  } catch (e: any) {
    pageError.value = e?.message ?? 'Link bilgileri yüklenemedi.'
  } finally {
    pageLoading.value = false
  }
}

onMounted(() => {
  loadLink()
})

async function copyShortUrl() {
  if (!fullShortUrl.value) return
  await navigator.clipboard?.writeText(fullShortUrl.value)
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}

async function handleSave() {
  if (!link.value) return
  saveError.value = ''
  saveSuccess.value = false

  if (!editTarget.value.trim()) {
    saveError.value = 'Hedef adres boş bırakılamaz.'
    return
  }

  let passwordPayload: string | undefined = undefined
  if (removePasswordFlag.value) {
    passwordPayload = ''
  } else if (newPassword.value !== '') {
    passwordPayload = newPassword.value
  }

  const finalWaitEnabled = !isPasswordActive.value && editWaitEnabled.value
  const parsedSeconds = Number.parseInt(String(editWaitSeconds.value), 10)

  if (finalWaitEnabled && (!Number.isFinite(parsedSeconds) || parsedSeconds < 1)) {
    saveError.value = 'Bekletme açıkken bekleme süresi en az 1 saniye olmalıdır.'
    return
  }

  saving.value = true
  try {
    const updated = await updateLink(link.value.id, {
      target: editTarget.value,
      alias: editAlias.value.trim(),
      password: passwordPayload,
      waitEnabled: finalWaitEnabled,
      waitSeconds: finalWaitEnabled ? parsedSeconds : 0
    })
    link.value = updated
    editTarget.value = updated.target
    editAlias.value = updated.alias
    newPassword.value = ''
    removePasswordFlag.value = false
    editWaitEnabled.value = !updated.locked && updated.waitEnabled
    editWaitSeconds.value = String(updated.waitSeconds > 0 ? updated.waitSeconds : 5)
    saveSuccess.value = true
    setTimeout(() => (saveSuccess.value = false), 2500)
  } catch (e: any) {
    saveError.value = e?.message ?? 'Link güncellenirken bir hata oluştu.'
  } finally {
    saving.value = false
  }
}

async function handleRemovePasswordNow() {
  if (!link.value) return
  saveError.value = ''
  saveSuccess.value = false
  saving.value = true
  try {
    const updated = await updateLink(link.value.id, {
      password: ''
    })
    link.value = updated
    newPassword.value = ''
    removePasswordFlag.value = false
    saveSuccess.value = true
    setTimeout(() => (saveSuccess.value = false), 2500)
  } catch (e: any) {
    saveError.value = e?.message ?? 'Şifre kaldırılamadı.'
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  if (!link.value) return
  try {
    await removeLink(link.value.id)
    await navigateTo('/dashboard/links')
  } catch (e: any) {
    saveError.value = e?.message ?? 'Link silinemedi.'
  }
}
</script>

<template>
  <div>
    <div class="mb-6">
      <NuxtLink to="/dashboard/links" class="inline-flex items-center gap-2 text-sm font-semibold text-ink-2 hover:text-ink">
        <UiIcon name="back" :size="16" /> Linklerime dön
      </NuxtLink>
    </div>

    <div v-if="pageLoading" class="flex items-center justify-center gap-3 rounded-lg border-[1.5px] border-ink bg-card p-12 shadow-hard text-ink-2">
      <span class="h-5 w-5 animate-spin rounded-full border-2 border-ink border-t-transparent" />
      <span class="font-medium">Link bilgileri yükleniyor...</span>
    </div>

    <div v-else-if="pageError" class="rounded-lg border-[1.5px] border-ink bg-card p-8 shadow-hard">
      <p class="text-accent-dark font-semibold">{{ pageError }}</p>
      <div class="mt-4">
        <UiButton to="/dashboard/links" variant="secondary">Linklerime dön</UiButton>
      </div>
    </div>

    <template v-else-if="link">
      <DashboardPageHeader
        :title="`/${link.alias}`"
        :description="link.target"
      >
        <template #actions>
          <UiButton variant="secondary" @click="copyShortUrl">
            <UiIcon :name="copied ? 'check' : 'copy'" :size="16" />
            {{ copied ? 'Kopyalandı' : 'Kısa Linki Kopyala' }}
          </UiButton>
          <a
            :href="fullShortUrl"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center justify-center gap-2 rounded-md border-[1.5px] border-ink bg-accent px-4 py-2.5 text-sm font-semibold text-white shadow-hard transition-all duration-100 hover:-translate-x-px hover:-translate-y-px hover:shadow-hard-lg"
          >
            Ziyaret Et <UiIcon name="arrow" :size="16" />
          </a>
        </template>
      </DashboardPageHeader>

      <div class="grid gap-5 sm:grid-cols-3">
        <DashboardStatTile label="Toplam tıklama" :value="link.clicks.toLocaleString('tr-TR')" />
        <DashboardStatTile label="Erişim durumu" :value="accessModeLabel" />
        <DashboardStatTile label="Oluşturulma tarihi" :value="link.createdAt" />
      </div>

      <div class="mt-8 grid gap-8 lg:grid-cols-[1.4fr_1fr]">
        <form class="flex flex-col gap-5 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8" @submit.prevent="handleSave">
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-bold">Linki Düzenle</h2>
            <UiBadge v-if="link.locked" tone="accent"><UiIcon name="lock" :size="11" /> şifreli</UiBadge>
            <UiBadge v-else-if="link.waitEnabled" tone="neutral"><UiIcon name="clock" :size="11" /> {{ link.waitSeconds }} sn bekletmeli</UiBadge>
            <UiBadge v-else tone="ok"><UiIcon name="check" :size="11" /> doğrudan</UiBadge>
          </div>

          <p v-if="saveError" class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
            {{ saveError }}
          </p>

          <UiField
            id="detail-target"
            v-model="editTarget"
            label="Hedef adres"
            type="text"
            placeholder="ornek.com/sayfa veya https://..."
            hint="Başında http:// veya https:// yoksa otomatik olarak https:// eklenir."
          />

          <UiField
            id="detail-alias"
            v-model="editAlias"
            label="Takma ad (Alias)"
            prefix="localhost/r/"
            placeholder="kampanya"
          />

          <div class="rounded-md border border-line bg-paper p-4">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <p class="text-sm font-semibold">Parola Koruması</p>
                <p class="text-xs text-ink-3">
                  {{ link.locked ? 'Bu link şu an parola ile korunuyor.' : 'Bu linkte şu an parola koruması yok.' }}
                </p>
              </div>
              <UiButton
                v-if="link.locked"
                type="button"
                variant="danger"
                @click="handleRemovePasswordNow"
              >
                <UiIcon name="unlock" :size="15" /> Şifreyi Kaldır
              </UiButton>
            </div>

            <div class="mt-3">
              <UiField
                id="detail-password"
                v-model="newPassword"
                :label="link.locked ? 'Parolayı Değiştir (İsteğe bağlı)' : 'Parola Ekle (İsteğe bağlı)'"
                type="password"
                :placeholder="link.locked ? 'Mevcut parolayı korumak için boş bırakın' : 'Şifre koymak için yazın'"
                autocomplete="new-password"
              />
            </div>
            <!-- TODO / Not: Şifreli linklerde şifre çözülmese bile ilk tıklamada tıklama sayılır ve log kaydı oluşturulur. İleride detaylı incelenecek. -->
            <p class="mt-2 text-[11px] text-ink-3">
              Not: Şifreli linklerde parola girilmese dahi linke tıklandığı anda tıklama sayısı ve cihaz logu kaydedilir (detaylı ayrım ileride eklenecektir).
            </p>
          </div>

          <div class="rounded-md border border-line bg-paper p-4">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-sm font-semibold">Ara Bekletme Sayfası (/w)</p>
                <p class="mt-0.5 text-xs text-ink-3">
                  <template v-if="isPasswordActive">
                    Şifre koruması açık olduğu için ara bekletme sayfası kullanılamaz. Bekletmeyi açmak için önce şifreyi kaldırın.
                  </template>
                  <template v-else>
                    Açık olduğunda ziyaretçiler hedef adrese gitmeden önce <span class="font-mono">/w/{{ link.alias }}</span> ara sayfasında seçtiğiniz süre kadar bekletilir (reklam alanı).
                  </template>
                </p>
              </div>
              <label class="inline-flex shrink-0 cursor-pointer items-center gap-2 select-none" :class="{ 'cursor-not-allowed opacity-50': isPasswordActive }">
                <input
                  v-model="editWaitEnabled"
                  type="checkbox"
                  :disabled="isPasswordActive"
                  class="h-4 w-4 accent-accent"
                />
                <span class="text-xs font-semibold">{{ editWaitEnabled && !isPasswordActive ? 'Açık' : 'Kapalı' }}</span>
              </label>
            </div>

            <div v-if="editWaitEnabled && !isPasswordActive" class="mt-3">
              <UiField
                id="detail-wait-seconds"
                v-model="editWaitSeconds"
                label="Bekleme Süresi (saniye)"
                type="number"
                placeholder="5"
                hint="Ziyaretçinin hedefe yönlendirilmeden önce bekletileceği saniye süresi."
              />
            </div>
          </div>

          <div class="flex items-center gap-3">
            <UiButton type="submit">{{ saving ? 'Kaydediliyor...' : 'Değişiklikleri Kaydet' }}</UiButton>
            <UiBadge v-if="saveSuccess" tone="ok"><UiIcon name="check" :size="12" /> kaydedildi</UiBadge>
          </div>
        </form>

        <div class="flex flex-col gap-6">
          <div class="flex flex-col gap-3 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
            <h3 class="text-lg font-bold">Bağlantı Özeti</h3>
            <div class="rounded-md border border-line bg-paper p-3">
              <p class="text-xs text-ink-3">Kısa Adres</p>
              <p class="mt-1 font-mono text-sm font-semibold break-all">{{ fullShortUrl }}</p>
            </div>
            <div class="rounded-md border border-line bg-paper p-3">
              <p class="text-xs text-ink-3">Hedef Adres</p>
              <p class="mt-1 text-sm break-all text-ink-2">{{ link.target }}</p>
            </div>
            <div class="rounded-md border border-line bg-paper p-3">
              <p class="text-xs text-ink-3">Yönlendirme Modu</p>
              <p class="mt-1 text-sm font-medium text-ink-2">
                <template v-if="link.locked">Parola Korumalı (/unlock/{{ link.alias }})</template>
                <template v-else-if="link.waitEnabled">Ara Bekletme Sayfası (/w/{{ link.alias }} — {{ link.waitSeconds }} sn)</template>
                <template v-else>Doğrudan Yönlendirme (302 Found)</template>
              </p>
            </div>
          </div>

          <div class="flex flex-col gap-4 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
            <div>
              <h3 class="text-lg font-bold text-accent-dark">Linki Sil</h3>
              <p class="mt-1 text-sm text-ink-3">
                Bu kısa linki sildiğinizde <span class="font-mono">{{ displayShortUrl }}</span> adresi artık çalışmaz.
              </p>
            </div>

            <div v-if="!confirmDelete">
              <UiButton variant="danger" @click="confirmDelete = true">
                <UiIcon name="trash" :size="16" /> Bu Linki Sil
              </UiButton>
            </div>
            <div v-else class="flex flex-wrap items-center gap-3">
              <UiButton variant="primary" @click="handleDelete">Evet, Kalıcı Olarak Sil</UiButton>
              <UiButton variant="ghost" @click="confirmDelete = false">Vazgeç</UiButton>
            </div>
          </div>
        </div>
      </div>

      <!-- Tıklama Logları Bölümü -->
      <div class="mt-8 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-xl font-bold">Tıklama Logları</h2>
            <p class="mt-1 text-xs text-ink-3">
              Bu kısa linke tıklayan cihazların IP, tarayıcı, işletim sistemi ve yönlendiren kaynak kayıtları.
            </p>
          </div>
          <UiButton variant="secondary" @click="loadLogs">
            <UiIcon name="refresh" :size="15" /> {{ logsLoading ? 'Yükleniyor...' : 'Logları Yenile' }}
          </UiButton>
        </div>

        <p v-if="logsError" class="mt-4 rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
          {{ logsError }}
        </p>

        <div v-if="logsLoading && logs.length === 0" class="mt-6 flex items-center justify-center gap-3 rounded-md border border-line bg-paper p-8 text-ink-2">
          <span class="h-4 w-4 animate-spin rounded-full border-2 border-ink border-t-transparent" />
          <span class="text-sm font-medium">Tıklama logları yükleniyor...</span>
        </div>

        <div v-else-if="logs.length === 0" class="mt-6 rounded-md border border-line bg-paper p-8 text-center">
          <p class="text-sm font-semibold text-ink-2">Henüz tıklama logu bulunmuyor.</p>
          <p class="mt-1 text-xs text-ink-3">Kısa link ziyaret edildiğinde cihaz bilgileri burada listelenecektir.</p>
        </div>

        <ul v-else class="mt-6 flex flex-col gap-3">
          <li
            v-for="item in logs"
            :key="item.id"
            class="flex flex-col gap-2.5 rounded-md border border-line bg-paper p-4 transition-colors hover:bg-paper-2/60"
          >
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div class="flex flex-wrap items-center gap-2">
                <span class="font-mono text-sm font-bold text-ink">{{ item.ipAddress || 'Bilinmiyor' }}</span>
                <UiBadge tone="accent">{{ item.deviceType || 'Cihaz' }}</UiBadge>
                <UiBadge tone="neutral">{{ item.browser || 'Tarayıcı' }}</UiBadge>
                <UiBadge tone="ok">{{ item.os || 'İS' }}</UiBadge>
              </div>
              <span class="font-mono text-xs text-ink-3">{{ item.clickedAt }}</span>
            </div>

            <div class="grid gap-2 text-xs text-ink-2 sm:grid-cols-2">
              <div>
                <span class="font-semibold text-ink-3">Kaynak (Referer): </span>
                <span class="break-all">{{ item.referer || 'Doğrudan' }}</span>
              </div>
              <div>
                <span class="font-semibold text-ink-3">Tarayıcı Dili: </span>
                <span class="font-mono">{{ item.acceptLanguage || '-' }}</span>
              </div>
            </div>

            <p v-if="item.userAgent" class="truncate font-mono text-[11px] text-ink-3" :title="item.userAgent">
              UA: {{ item.userAgent }}
            </p>
          </li>
        </ul>
      </div>
    </template>
  </div>
</template>
