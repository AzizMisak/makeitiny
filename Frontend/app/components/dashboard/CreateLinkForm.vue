<script setup lang="ts">
const { isAdmin, canUseFeature, remainingLinks } = useAuth()
const { addLink, checkAlias } = useLinks()

const target = ref('')
const title = ref('')
const alias = ref('')
const password = ref('')
const waitEnabled = ref(false)
const waitSeconds = ref('5')
const adEnabled = ref(false)
const expiresAt = ref('')
const maxClicks = ref('')

const errorMsg = ref('')
const loading = ref(false)
const aliasChecking = ref(false)
const aliasCheckResult = ref<{ available: boolean; message: string } | null>(null)

const hasPassword = computed(() => password.value.trim().length > 0)
const canCustomAlias = computed(() => canUseFeature('allowCustomAlias'))
const canPassword = computed(() => canUseFeature('allowPassword'))
const canAds = computed(() => canUseFeature('allowAds'))
const quotaLimitReached = computed(() => !isAdmin.value && remainingLinks.value !== null && remainingLinks.value <= 0)

watch(hasPassword, (active) => {
  if (active) {
    waitEnabled.value = false
  }
})

watch(alias, () => {
  aliasCheckResult.value = null
})

async function handleCheckAlias() {
  const val = alias.value.trim()
  if (!val) {
    aliasCheckResult.value = { available: false, message: 'Önce bir takma ad yaz.' }
    return
  }
  aliasChecking.value = true
  try {
    const res = await checkAlias(val)
    aliasCheckResult.value = {
      available: res.available,
      message: res.available ? `'${res.alias}' kullanılabilir.` : `'${res.alias}' kullanımda.`
    }
  } catch (e: any) {
    aliasCheckResult.value = {
      available: false,
      message: e?.message ?? 'Kontrol edilemedi.'
    }
  } finally {
    aliasChecking.value = false
  }
}

async function submit() {
  if (quotaLimitReached.value) {
    errorMsg.value = 'Link oluşturma kotan doldu. Yeni link için eski bir linki sil veya paket yükselt.'
    return
  }
  if (!target.value.trim()) {
    errorMsg.value = 'Hedef adres gerekli.'
    return
  }

  const parsedSeconds = Number.parseInt(String(waitSeconds.value), 10)
  const finalWaitEnabled = !hasPassword.value && waitEnabled.value

  if (finalWaitEnabled && (!Number.isFinite(parsedSeconds) || parsedSeconds < 1)) {
    errorMsg.value = 'Bekleme süresi en az 1 saniye olmalı.'
    return
  }

  let parsedMaxClicks: number | undefined
  const rawMaxClicks = String(maxClicks.value ?? '').trim()
  if (rawMaxClicks !== '') {
    const n = Number.parseInt(rawMaxClicks, 10)
    if (!Number.isFinite(n) || n < 1) {
      errorMsg.value = 'Maksimum tıklama pozitif bir sayı olmalı.'
      return
    }
    parsedMaxClicks = n
  }

  errorMsg.value = ''
  loading.value = true
  try {
    const created = await addLink({
      target: target.value,
      title: title.value.trim() || undefined,
      alias: canCustomAlias.value && alias.value.trim() ? alias.value.trim() : undefined,
      password: canPassword.value && password.value ? password.value : undefined,
      waitEnabled: finalWaitEnabled,
      waitSeconds: finalWaitEnabled ? parsedSeconds : 0,
      expiresAt: expiresAt.value.trim() || undefined,
      maxClicks: parsedMaxClicks,
      adEnabled: canAds.value ? adEnabled.value : false
    })
    await navigateTo(`/dashboard/links/${created.id}`)
  } catch (e: any) {
    errorMsg.value = e?.message ?? 'Link oluşturulamadı.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <div
      v-if="quotaLimitReached"
      class="mb-6 flex flex-wrap items-center justify-between gap-4 rounded-lg border-[1.5px] border-ink bg-accent-soft p-4 shadow-hard"
    >
      <p class="text-sm font-semibold text-accent-dark">
        Link kotan doldu. Yeni link eklemek için paketini yükselt veya eski bir linki sil.
      </p>
      <UiButton to="/dashboard/subscription" variant="secondary">Paketler</UiButton>
    </div>

    <form
      class="flex flex-col gap-6 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8"
      @submit.prevent="submit"
    >
      <p v-if="errorMsg" class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
        {{ errorMsg }}
      </p>

      <div class="grid gap-5 sm:grid-cols-2">
        <UiField
          id="new-target"
          v-model="target"
          label="Hedef adres *"
          type="text"
          placeholder="https://ornek.com/sayfa"
          hint="Protokol yoksa https:// eklenir."
        />

        <UiField
          id="new-title"
          v-model="title"
          label="Başlık (isteğe bağlı)"
          type="text"
          placeholder="Örn. Yaz kampanyası"
        />
      </div>

      <div class="border-t border-line pt-5">
        <div class="mb-2 flex items-center gap-2">
          <span class="text-sm font-semibold text-ink">Özel takma ad</span>
          <UiBadge v-if="!canCustomAlias" tone="accent">Pro</UiBadge>
        </div>

        <div class="flex flex-col gap-2 sm:flex-row sm:items-start">
          <div class="flex-1">
            <UiField
              id="new-alias"
              v-model="alias"
              label=""
              prefix="localhost/r/"
              :placeholder="canCustomAlias ? 'kampanya-2026' : 'Rastgele üretilir'"
              :disabled="!canCustomAlias"
              :hint="canCustomAlias ? 'Boş kalırsa rastgele 6 karakter üretilir.' : 'Bu özellik paketinde kapalı.'"
            />
          </div>
          <UiButton
            v-if="canCustomAlias"
            type="button"
            variant="secondary"
            class="sm:mt-1.5"
            @click="handleCheckAlias"
          >
            <UiIcon name="search" :size="15" /> {{ aliasChecking ? 'Kontrol...' : 'Müsait mi?' }}
          </UiButton>
        </div>

        <p
          v-if="aliasCheckResult"
          class="mt-2 font-mono text-xs font-semibold"
          :class="aliasCheckResult.available ? 'text-ok' : 'text-accent-dark'"
        >
          {{ aliasCheckResult.message }}
        </p>
      </div>

      <div class="grid gap-5 border-t border-line pt-5 sm:grid-cols-2">
        <div>
          <div class="mb-2 flex items-center gap-2">
            <span class="text-sm font-semibold">Şifre koruması</span>
            <UiBadge v-if="!canPassword" tone="accent">Kilitli</UiBadge>
          </div>
          <UiField
            id="new-password"
            v-model="password"
            label=""
            type="password"
            placeholder="İsteğe bağlı"
            :disabled="!canPassword"
            hint="Şifre varsa ara bekletme kapanır."
          />
        </div>

        <div>
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-sm font-semibold">Ara bekletme</p>
              <p class="mt-0.5 text-xs text-ink-3">
                {{ hasPassword ? 'Şifre varken kapalıdır.' : 'Yönlendirmeden önce sayaç gösterir.' }}
              </p>
            </div>
            <label
              class="inline-flex shrink-0 cursor-pointer items-center gap-2 select-none"
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

          <div v-if="waitEnabled && !hasPassword" class="mt-3">
            <UiField
              id="new-wait-seconds"
              v-model="waitSeconds"
              label="Süre (sn)"
              type="number"
              placeholder="5"
            />
          </div>
        </div>
      </div>

      <div class="grid gap-5 border-t border-line pt-5 sm:grid-cols-3">
        <UiField
          id="new-expires-at"
          v-model="expiresAt"
          label="Son kullanma tarihi"
          type="date"
        />

        <UiField
          id="new-max-clicks"
          v-model="maxClicks"
          label="Maksimum tıklama"
          type="number"
          placeholder="Sınırsız"
        />

        <div class="flex flex-col justify-between rounded-md border border-line bg-paper px-4 py-3">
          <div class="flex items-center justify-between gap-2">
            <span class="text-sm font-semibold">Reklam gösterimi</span>
            <UiBadge v-if="!canAds" tone="accent">Pro</UiBadge>
          </div>
          <label
            class="mt-2 inline-flex cursor-pointer items-center gap-2 select-none"
            :class="{ 'cursor-not-allowed opacity-50': !canAds }"
          >
            <input
              v-model="adEnabled"
              type="checkbox"
              :disabled="!canAds"
              class="h-4 w-4 accent-accent"
            />
            <span class="text-xs font-semibold">{{ adEnabled && canAds ? 'Açık' : 'Kapalı' }}</span>
          </label>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-3 border-t border-line pt-5">
        <UiButton type="submit" :disabled="loading || quotaLimitReached">
          <UiIcon name="plus" :size="16" /> {{ loading ? 'Oluşturuluyor...' : 'Linki oluştur' }}
        </UiButton>
        <UiButton to="/dashboard/links" variant="ghost">Vazgeç</UiButton>
      </div>
    </form>
  </div>
</template>
