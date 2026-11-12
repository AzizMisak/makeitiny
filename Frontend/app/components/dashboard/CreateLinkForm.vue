<script setup lang="ts">
const { addLink } = useLinks()

const target = ref('')
const alias = ref('')
const password = ref('')
const waitEnabled = ref(false)
const waitSeconds = ref('5')
const errorMsg = ref('')
const loading = ref(false)

const hasPassword = computed(() => password.value.trim().length > 0)

watch(hasPassword, (active) => {
  if (active) {
    waitEnabled.value = false
  }
})

async function submit() {
  if (!target.value.trim()) {
    errorMsg.value = 'Lütfen kısaltmak istediğiniz hedef adresi girin.'
    return
  }

  const parsedSeconds = Number.parseInt(String(waitSeconds.value), 10)
  const finalWaitEnabled = !hasPassword.value && waitEnabled.value

  if (finalWaitEnabled && (!Number.isFinite(parsedSeconds) || parsedSeconds < 1)) {
    errorMsg.value = 'Bekletme açıkken bekleme süresi en az 1 saniye olmalıdır.'
    return
  }

  errorMsg.value = ''
  loading.value = true
  try {
    const created = await addLink({
      target: target.value,
      alias: alias.value.trim() || undefined,
      password: password.value || undefined,
      waitEnabled: finalWaitEnabled,
      waitSeconds: finalWaitEnabled ? parsedSeconds : 0
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
  <form class="flex max-w-xl flex-col gap-5 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8" @submit.prevent="submit">
    <p v-if="errorMsg" class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
      {{ errorMsg }}
    </p>
    <UiField
      id="new-target"
      v-model="target"
      label="Hedef adres"
      type="text"
      placeholder="ornek.com/sayfa veya https://ornek.com/sayfa"
      hint="Başında http:// veya https:// yoksa varsayılan olarak güvenli https:// eklenir."
    />
    <UiField
      id="new-alias"
      v-model="alias"
      label="Takma ad (isteğe bağlı)"
      prefix="localhost/r/"
      placeholder="kampanya"
      hint="Boş bırakırsan rastgele 6 karakterli Base62 kod üretilir."
    />
    <UiField
      id="new-password"
      v-model="password"
      label="Parola (isteğe bağlı)"
      type="password"
      placeholder="İsteğe bağlı parola"
      hint="Doluysa link Argon2id ile şifreli olur. Şifre açıkken ara bekletme sayfası kapatılır."
    />

    <div class="rounded-md border border-line bg-paper p-4">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm font-semibold">Ara Bekletme Sayfası (/w)</p>
          <p class="mt-0.5 text-xs text-ink-3">
            <template v-if="hasPassword">
              Şifre koruması girildiği için ara bekletme sayfası devre dışıdır (şifre ve bekletme aynı anda açık olamaz).
            </template>
            <template v-else>
              Açık olduğunda ziyaretçiler hedef adrese gitmeden önce <span class="font-mono">/w</span> ara sayfasında belirttiğiniz süre kadar bekletilir (reklam gösterim alanı).
            </template>
          </p>
        </div>
        <label class="inline-flex shrink-0 cursor-pointer items-center gap-2 select-none" :class="{ 'cursor-not-allowed opacity-50': hasPassword }">
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
          label="Bekleme Süresi (saniye)"
          type="number"
          placeholder="5"
          hint="Kullanıcının ara sayfada kaç saniye bekletileceğini girin."
        />
      </div>
    </div>

    <div class="flex gap-3">
      <UiButton type="submit">{{ loading ? 'Oluşturuluyor...' : 'Linki oluştur' }}</UiButton>
      <UiButton to="/dashboard/links" variant="ghost">Vazgeç</UiButton>
    </div>
  </form>
</template>
