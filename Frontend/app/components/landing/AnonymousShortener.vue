<script setup lang="ts">
import type { AnonymousLinkResponse } from '~/types/link'

const { public: config } = useRuntimeConfig()
const { createAnonymousLink } = useLinks()

const target = ref('')
const password = ref('')
const waitEnabled = ref(false)
const waitSeconds = ref('5')
const ttlHours = ref(48)

const loading = ref(false)
const errorMsg = ref('')
const created = ref<AnonymousLinkResponse | null>(null)
const copied = ref(false)

const hasPassword = computed(() => password.value.trim().length > 0)

watch(hasPassword, (active) => {
  if (active) waitEnabled.value = false
})

const qrImageUrl = computed(() => {
  if (!created.value?.hasQr) return ''
  return `${config.apiBase}/links/${created.value.id}/qr`
})

const qrDownloadUrl = computed(() => {
  if (!created.value?.hasQr) return ''
  return `${config.apiBase}/links/${created.value.id}/qr?download=1`
})

function handleBeforeUnload(e: BeforeUnloadEvent) {
  if (created.value && !copied.value) {
    e.preventDefault()
    e.returnValue = ''
  }
}

onMounted(() => {
  window.addEventListener('beforeunload', handleBeforeUnload)
})

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', handleBeforeUnload)
})

async function submit() {
  errorMsg.value = ''
  if (!target.value.trim()) {
    errorMsg.value = 'Lütfen kısaltmak istediğiniz hedef adresi girin.'
    return
  }

  const parsedWait = Number.parseInt(String(waitSeconds.value), 10)
  const useWait = !hasPassword.value && waitEnabled.value
  if (useWait && (!Number.isFinite(parsedWait) || parsedWait < 1 || parsedWait > 30)) {
    errorMsg.value = 'Bekleme süresi 1 ile 30 saniye arasında olmalıdır.'
    return
  }

  loading.value = true
  try {
    created.value = await createAnonymousLink({
      target: target.value,
      password: password.value.trim() || undefined,
      waitEnabled: useWait,
      waitSeconds: useWait ? parsedWait : 0,
      ttlHours: ttlHours.value
    })
    copied.value = false
  } catch (e: any) {
    errorMsg.value = e?.message ?? 'Anonim link oluşturulamadı.'
  } finally {
    loading.value = false
  }
}

async function copyCreatedUrl() {
  if (!created.value) return
  await navigator.clipboard?.writeText(created.value.shortUrl)
  copied.value = true
}

function formatExpiry(iso: string): string {
  if (!iso) return '48 saat'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString('tr-TR', { dateStyle: 'medium', timeStyle: 'short' })
}
</script>

<template>
  <section id="hizli-kisalt" class="container-x py-14 md:py-20">
    <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard-lg sm:p-10">
      <div class="flex flex-wrap items-start justify-between gap-4 border-b border-line pb-5">
        <div>
          <span class="eyebrow">Üyeliksiz Hızlı Kısaltma</span>
          <h2 class="mt-1.5 text-2xl font-extrabold sm:text-3xl">Kayıt olmadan 2 günlük kısa link oluştur.</h2>
          <p class="mt-2 max-w-2xl text-sm text-ink-2">
            Üye olmadan geçici bağlantı kısaltabilir, parola ve bekleme süresi ekleyebilirsin.
          </p>
        </div>
        <UiBadge tone="neutral">IP başına günde 1 adet • Maks. 48 saat</UiBadge>
      </div>

      <!-- Kurallar Bilgilendirmesi -->
      <div class="mt-5 flex items-start gap-3 rounded-md border border-line bg-paper p-4 text-xs text-ink-2">
        <UiIcon name="alert" :size="18" class="mt-0.5 shrink-0 text-accent-dark" />
        <div class="space-y-1">
          <p class="font-semibold text-ink">Anonim Link Kuralları:</p>
          <p>
            Anonim linkler için bir yönetim paneli bulunmaz. Linki oluşturduktan sonra
            <strong>bağlantıyı mutlaka kopyalayıp kaydetmelisiniz</strong> (sayfadan ayrılırsanız bir daha göremezsiniz).
            Süresi (en fazla 2 gün) dolduğunda link otomatik olarak silinir.
          </p>
        </div>
      </div>

      <!-- Sonuç Kartı -->
      <div v-if="created" class="mt-6 rounded-lg border-[1.5px] border-ink bg-paper p-6 shadow-hard">
        <div
          v-if="!copied"
          class="mb-5 flex flex-wrap items-center justify-between gap-3 rounded-md border-[1.5px] border-accent-dark bg-accent-soft px-4 py-3 text-sm font-semibold text-accent-dark"
        >
          <span>Dikkat: Yönetim paneli olmadığı için bu sayfadan çıkmadan önce kısa linkinizi mutlaka kopyalayın!</span>
          <UiButton variant="primary" @click="copyCreatedUrl">
            <UiIcon name="copy" :size="15" /> Şimdi Kopyala
          </UiButton>
        </div>

        <div
          v-else
          class="mb-5 flex items-center gap-2 rounded-md border border-ok bg-ok-soft px-4 py-2.5 text-sm font-semibold text-ok"
        >
          <UiIcon name="check" :size="16" /> Kısa link panoya kopyalandı!
        </div>

        <div class="grid gap-6 md:grid-cols-[1fr_auto]">
          <div class="flex flex-col justify-between gap-4">
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-ink-3">Oluşturulan Kısa Adres</p>
              <div class="mt-2 flex flex-wrap items-center gap-3">
                <a
                  :href="created.shortUrl"
                  target="_blank"
                  rel="noopener"
                  class="font-mono text-lg font-bold text-accent-dark underline underline-offset-4 break-all"
                >
                  {{ created.shortUrl }}
                </a>
                <UiButton variant="secondary" @click="copyCreatedUrl">
                  <UiIcon :name="copied ? 'check' : 'copy'" :size="15" />
                  {{ copied ? 'Kopyalandı' : 'Bağlantıyı Kopyala' }}
                </UiButton>
              </div>
              <p class="mt-2 text-xs text-ink-3 break-all">Hedef: {{ created.target }}</p>
            </div>

            <div class="flex flex-wrap gap-2">
              <UiBadge tone="neutral">Bitiş: {{ formatExpiry(created.expiresAt) }}</UiBadge>
              <UiBadge v-if="created.locked" tone="accent"><UiIcon name="lock" :size="11" /> Parola Korumalı</UiBadge>
              <UiBadge v-else-if="created.waitEnabled" tone="neutral">
                <UiIcon name="clock" :size="11" /> {{ created.waitSeconds }} sn bekletmeli
              </UiBadge>
              <UiBadge v-else tone="ok">Doğrudan 302</UiBadge>
            </div>
          </div>

          <div v-if="created.hasQr" class="flex flex-col items-center gap-2 rounded-md border border-line bg-card p-3">
            <img
              :src="qrImageUrl"
              :alt="`${created.alias} QR Kodu`"
              class="h-28 w-28 rounded-sm border border-line bg-white p-1"
            />
            <a
              :href="qrDownloadUrl"
              class="inline-flex items-center gap-1 text-xs font-semibold text-ink-2 hover:text-ink"
            >
              <UiIcon name="download" :size="13" /> QR İndir (PNG)
            </a>
          </div>
        </div>
      </div>

      <!-- Oluşturma Formu -->
      <form v-else class="mt-6 flex flex-col gap-5" @submit.prevent="submit">
        <p v-if="errorMsg" class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
          {{ errorMsg }}
        </p>

        <div class="grid gap-4 md:grid-cols-[1.5fr_1fr]">
          <UiField
            id="anon-target"
            v-model="target"
            label="Hedef Web Adresi *"
            type="text"
            placeholder="ornek.com/paylasilacak-sayfa"
            hint="Başında https:// yoksa otomatik olarak güvenli https:// eklenir."
          />

          <div>
            <label for="anon-ttl" class="mb-1.5 block text-sm font-semibold">Geçerlilik Süresi</label>
            <select
              id="anon-ttl"
              v-model.number="ttlHours"
              class="w-full rounded-md border-[1.5px] border-ink bg-card px-3 py-2.5 text-sm font-semibold text-ink focus:outline-none focus:ring-2 focus:ring-accent"
            >
              <option :value="6">6 Saat</option>
              <option :value="24">24 Saat (1 Gün)</option>
              <option :value="48">48 Saat (2 Gün — Maksimum)</option>
            </select>
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <UiField
            id="anon-password"
            v-model="password"
            label="Parola (İsteğe bağlı)"
            type="password"
            placeholder="Şifre koymak için yazın"
            hint="Parola girilirse ara bekletme kapanır."
          />

          <div class="rounded-md border border-line bg-paper p-3.5">
            <div class="flex items-center justify-between gap-2">
              <div>
                <p class="text-sm font-semibold">Ara Bekletme Süresi</p>
                <p class="text-xs text-ink-3">Yönlendirmeden önce sayaç ekranı açar.</p>
              </div>
              <label
                class="inline-flex cursor-pointer items-center gap-2 select-none"
                :class="{ 'cursor-not-allowed opacity-50': hasPassword }"
              >
                <input
                  v-model="waitEnabled"
                  type="checkbox"
                  :disabled="hasPassword"
                  class="h-4 w-4 accent-accent"
                />
                <span class="text-xs font-semibold">{{ waitEnabled && !hasPassword ? 'Açık' : 'Kapalı' }}</span>
              </label>
            </div>
            <div v-if="waitEnabled && !hasPassword" class="mt-2.5">
              <UiField
                id="anon-wait-sec"
                v-model="waitSeconds"
                label="Bekleme Süresi (1-30 sn)"
                type="number"
                placeholder="5"
              />
            </div>
          </div>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-4 border-t border-line pt-4">
          <p class="text-xs text-ink-3">
            Daha uzun süreli linkler, özel alias ve detaylı tıklama grafikleri için
            <NuxtLink to="/auth/register" class="font-semibold text-accent-dark underline underline-offset-2">
              ücretsiz hesap açın
            </NuxtLink>.
          </p>
          <UiButton type="submit" :disabled="loading">
            <UiIcon name="sparkle" :size="15" /> {{ loading ? 'Kısaltılıyor...' : 'Geçici Link Oluştur' }}
          </UiButton>
        </div>
      </form>
    </div>
  </section>
</template>
