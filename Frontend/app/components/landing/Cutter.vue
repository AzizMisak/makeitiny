<script setup lang="ts">
import type { AnonymousLinkResponse } from '~/types/link'

type Access = 'open' | 'password' | 'wait'

const { public: config } = useRuntimeConfig()
const { isLoggedIn } = useAuth()
const { createAnonymousLink } = useLinks()
const { prefix } = useShortUrl()
const { copied, copy } = useCopy(4000)
const handoff = useState<string>('mit_draft_target', () => '')

const target = ref('')
const ttl = ref(48)
const access = ref<Access>('open')
const password = ref('')
const seconds = ref('5')
const loading = ref(false)
const error = ref('')
const created = ref<AnonymousLinkResponse | null>(null)
const everCopied = ref(false)
const qrReady = ref(false)
const qrFailed = ref(false)

const ttlOptions = [
  { value: 6, label: '6 saat' },
  { value: 24, label: '24 saat' },
  { value: 48, label: '48 saat' }
]
const accessOptions = [
  { value: 'open' as const, label: 'Herkese açık' },
  { value: 'password' as const, label: 'Parolalı' },
  { value: 'wait' as const, label: 'Bekletmeli' }
]

watch([target, access, password, seconds, ttl], () => (error.value = ''))

// --- Cetvel: yazı tipi sabit genişlikte olduğu için her karakter tam bir çizgiye denk gelir. ---
const input = ref<HTMLInputElement | null>(null)
const offset = ref(0)

function syncRuler() {
  offset.value = input.value?.scrollLeft ?? 0
}

const full = computed(() => normalizeTargetUrl(target.value))
const shortText = computed(() => created.value?.shortUrl.replace(/^https?:\/\//, '') ?? '')
const saved = computed(() => {
  const before = created.value?.target.length ?? 0
  return before > 0 ? Math.max(0, Math.round((1 - shortText.value.length / before) * 100)) : 0
})

async function submit() {
  error.value = ''
  if (!target.value.trim()) {
    error.value = 'Kısaltmak istediğin adresi yapıştır.'
    input.value?.focus()
    return
  }

  // Oturum açıksa iş panelde biter; adres oraya taşınır.
  if (isLoggedIn.value) {
    handoff.value = target.value.trim()
    loading.value = true
    await navigateTo('/dashboard/new')
    return
  }

  const wait = Number(seconds.value)
  if (access.value === 'password' && !password.value.trim()) {
    error.value = 'Bir parola yaz ya da "Herkese açık" seç.'
    return
  }
  if (access.value === 'wait' && !(Number.isInteger(wait) && wait >= 1 && wait <= 30)) {
    error.value = 'Bekletme süresi 1 ile 30 saniye arasında olmalı.'
    return
  }

  loading.value = true
  try {
    created.value = await createAnonymousLink({
      target: target.value,
      password: access.value === 'password' ? password.value.trim() : undefined,
      waitEnabled: access.value === 'wait',
      waitSeconds: access.value === 'wait' ? wait : 0,
      ttlHours: ttl.value
    })
    everCopied.value = false
    qrReady.value = false
    qrFailed.value = false
  } catch (e: any) {
    error.value = e?.message ?? 'Link oluşturulamadı. Tekrar dene.'
  } finally {
    loading.value = false
  }
}

async function copyShort() {
  if (created.value && (await copy(created.value.shortUrl))) everCopied.value = true
}

function warn(e: BeforeUnloadEvent) {
  if (!created.value || everCopied.value) return
  e.preventDefault()
  e.returnValue = ''
}

onMounted(() => window.addEventListener('beforeunload', warn))
onBeforeUnmount(() => window.removeEventListener('beforeunload', warn))

const expiry = computed(() => (created.value ? formatDateTimeTr(created.value.expiresAt) : ''))
</script>

<template>
  <div>
    <!-- Sonuç -->
    <div v-if="created" class="min-h-[16rem]" aria-live="polite">
      <div class="grid grid-cols-1 items-start gap-8 md:grid-cols-[minmax(0,1fr)_auto]">
        <div class="min-w-0">
          <p class="tnum mb-3 text-on-mat-2">
            {{ created.target.length }} karakterden {{ shortText.length }} karaktere indi<template v-if="saved > 0">, %{{ saved }} kısaldı</template>.
          </p>
          <div class="flex flex-wrap items-stretch gap-3">
            <div class="cut-in strip-cut flex min-h-16 min-w-0 items-center bg-mark pl-6 pr-12 font-mono text-[clamp(1.1rem,3.4vw,1.75rem)] font-semibold text-on-mark">
              <a :href="created.shortUrl" target="_blank" rel="noopener" class="truncate underline-offset-4 hover:underline">{{ shortText }}</a>
            </div>
            <UiButton variant="mat" size="lg" :icon="copied ? 'check' : 'copy'" class="min-w-44 !min-h-16" @click="copyShort">
              {{ copied ? 'Kopyalandı' : 'Kopyala' }}
            </UiButton>
          </div>
          <p class="mt-4 max-w-xl font-medium" :class="everCopied ? 'text-on-mat' : 'text-mark'">
            {{ everCopied ? 'Kopyalandı. Artık paylaşabilirsin.' : 'Sayfadan çıkmadan kopyala. Bu link hiçbir panelde görünmez.' }}
          </p>
          <p class="mt-1 text-on-mat-2">
            {{ expiry }} tarihinde silinir.
            <template v-if="created.locked">Açmak için parola istenir.</template>
            <template v-else-if="created.waitEnabled">{{ created.waitSeconds }} saniye bekletir.</template>
          </p>
        </div>

        <div v-if="created.hasQr" class="flex items-end gap-4 md:flex-col md:items-stretch">
          <div class="relative flex h-40 w-40 items-center justify-center bg-white p-3">
            <img v-show="qrReady" :src="`${config.apiBase}/links/${created.id}/qr`" alt="Kısa linkin QR kodu" class="h-full w-full" @load="qrReady = true" @error="qrFailed = true" />
            <UiSpinner v-if="!qrReady && !qrFailed" :size="22" class="absolute text-[#0c5c47]" />
            <span v-if="qrFailed" class="px-2 text-center text-xs text-[#5a6d64]">QR yüklenemedi</span>
          </div>
          <UiButton variant="mat" size="sm" icon="download" :href="`${config.apiBase}/links/${created.id}/qr?download=1`">QR indir</UiButton>
        </div>
      </div>
    </div>

    <!-- Kesim tezgâhı -->
    <form v-else :class="isLoggedIn ? '' : 'min-h-[16rem]'" novalidate @submit.prevent="submit">
      <div class="flex flex-col gap-3 md:flex-row md:items-end">
        <div class="min-w-0 flex-1">
          <!-- Cetvel -->
          <div class="relative h-8 overflow-hidden pl-5 font-mono text-base text-on-mat-2 md:text-lg" aria-hidden="true">
            <div class="ruler-ticks absolute inset-y-0 left-5 w-[300ch]" :style="{ transform: `translateX(${-offset}px)` }">
              <span v-for="n in 29" :key="n" class="absolute top-0" :style="{ left: `${n * 10}ch` }">
                <span class="tnum absolute -translate-x-1/2 text-[12px]" :class="n * 10 <= target.length ? 'text-mark' : ''">{{ n * 10 }}</span>
              </span>
            </div>
          </div>
          <label for="cutter-target" class="sr-only">Kısaltılacak adres</label>
          <input
            id="cutter-target"
            ref="input"
            v-model="target"
            type="text"
            inputmode="url"
            autocomplete="off"
            spellcheck="false"
            placeholder="Uzun adresi buraya yapıştır"
            class="h-16 w-full bg-white pl-5 pr-4 font-mono text-base text-[#10211b] placeholder:text-[#5a6d64] focus:outline-3 focus:outline-offset-0 focus:outline-mark md:text-lg"
            :aria-invalid="!!error || undefined"
            aria-describedby="cutter-msg"
            @scroll="syncRuler"
            @input="syncRuler"
            @keyup="syncRuler"
          />
        </div>
        <UiButton type="submit" variant="mark" size="lg" icon="scissors" class="!min-h-16 md:min-w-40" :loading="loading" loading-label="Kesiliyor">
          {{ isLoggedIn ? 'Panelde kes' : 'Kes' }}
        </UiButton>
      </div>

      <p id="cutter-msg" class="tnum mt-3 flex min-h-6 items-center gap-2" :class="error ? 'font-medium text-mark' : 'text-on-mat-2'" aria-live="polite">
        <template v-if="error"><UiIcon name="alert" :size="17" /> {{ error }}</template>
        <template v-else-if="target">{{ full.length }} karakter, kesince yaklaşık {{ prefix.length + 6 }} karakter.</template>
        <template v-else>Her çizgi bir karakter.</template>
      </p>

      <div v-if="!isLoggedIn" class="mt-5 grid grid-cols-1 gap-x-8 gap-y-5 lg:grid-cols-[auto_auto_minmax(0,1fr)]">
        <div>
          <p class="mb-2 text-sm text-on-mat-2">Ne kadar yaşasın?</p>
          <UiSegmented v-model="ttl" tone="mat" label="Linkin ömrü" :options="ttlOptions" />
        </div>
        <div class="min-w-0">
          <p class="mb-2 text-sm text-on-mat-2">Tıklayan nasıl geçsin?</p>
          <UiSegmented v-model="access" tone="mat" label="Erişim türü" :options="accessOptions" />
        </div>
        <!-- Seçime göre değişen tek alan; yeri her seçimde aynı. -->
        <div class="flex h-[4.5rem] flex-col justify-end">
          <p v-if="access === 'open'" class="flex min-h-12 items-center text-on-mat-2">Beklemeden hedefe gider.</p>
          <template v-else-if="access === 'password'">
            <label for="cutter-password" class="mb-2 text-sm text-on-mat-2">Link parolası</label>
            <UiField id="cutter-password" v-model="password" bare type="password" autocomplete="new-password" class="max-w-64" />
          </template>
          <template v-else>
            <label for="cutter-wait" class="mb-2 text-sm text-on-mat-2">Geri sayım, 1–30 saniye</label>
            <UiField id="cutter-wait" v-model="seconds" bare type="number" min="1" max="30" inputmode="numeric" suffix="saniye" class="max-w-40" />
          </template>
        </div>
      </div>
    </form>

    <p v-if="!isLoggedIn" class="mt-6 text-sm text-on-mat-2">Üyeliksiz günde bir link. En geç 48 saat sonra kendiliğinden silinir.</p>
  </div>
</template>

<style scoped>
.ruler-ticks {
  background-image:
    linear-gradient(90deg, currentcolor 1px, transparent 1px),
    linear-gradient(90deg, currentcolor 1px, transparent 1px),
    linear-gradient(90deg, currentcolor 1px, transparent 1px);
  background-size:
    10ch 14px,
    5ch 10px,
    1ch 6px;
  background-position: left bottom;
  background-repeat: repeat-x;
}

.cut-in {
  animation: cut-in 620ms cubic-bezier(0.7, 0, 0.2, 1) both;
}

@keyframes cut-in {
  from {
    padding-right: min(60vw, 40rem);
    background-color: #fff;
  }
}
</style>
