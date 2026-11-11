<script setup lang="ts">
import type { LinkItem } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const route = useRoute()
const { public: config } = useRuntimeConfig()
const { getLink, updateLink, removeLink } = useLinks()

const id = computed(() => String(route.params.id ?? ''))
const link = ref<LinkItem | null>(null)
const pageLoading = ref(true)
const pageError = ref('')

const editTarget = ref('')
const editAlias = ref('')
const newPassword = ref('')
const removePasswordFlag = ref(false)

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

  saving.value = true
  try {
    const updated = await updateLink(link.value.id, {
      target: editTarget.value,
      alias: editAlias.value.trim(),
      password: passwordPayload
    })
    link.value = updated
    editTarget.value = updated.target
    editAlias.value = updated.alias
    newPassword.value = ''
    removePasswordFlag.value = false
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
        <DashboardStatTile label="Erişim durumu" :value="link.locked ? 'Şifreli' : 'Herkese Açık'" />
        <DashboardStatTile label="Oluşturulma tarihi" :value="link.createdAt" />
      </div>

      <div class="mt-8 grid gap-8 lg:grid-cols-[1.4fr_1fr]">
        <form class="flex flex-col gap-5 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8" @submit.prevent="handleSave">
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-bold">Linki Düzenle</h2>
            <UiBadge v-if="link.locked" tone="accent"><UiIcon name="lock" :size="11" /> şifreli</UiBadge>
            <UiBadge v-else tone="ok"><UiIcon name="check" :size="11" /> şifresiz</UiBadge>
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
    </template>
  </div>
</template>
