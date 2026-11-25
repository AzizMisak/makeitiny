<script setup lang="ts">
import type { LinkItem } from '~/types/link'

const link = defineModel<LinkItem>({ required: true })

const { public: config } = useRuntimeConfig()
const { canUseFeature } = useAuth()
const { generateQr, deleteQr } = useLinks()
const { display } = useShortUrl()
const { ask } = useConfirm()
const toast = useToast()

const canQr = computed(() => canUseFeature('allowQr'))
const busy = ref(false)
const error = ref('')
const nonce = ref(0)
const imageReady = ref(false)
const imageFailed = ref(false)

const src = computed(() => `${config.apiBase}/links/${link.value.id}/qr?v=${nonce.value}`)
const downloadUrl = computed(() => `${config.apiBase}/links/${link.value.id}/qr?download=1`)

watch(src, () => {
  imageReady.value = false
  imageFailed.value = false
})

onMounted(() => {
  nonce.value = Date.now()
})

async function generate() {
  busy.value = true
  error.value = ''
  try {
    link.value = await generateQr(link.value.id)
    nonce.value = Date.now()
    toast.success('QR kodu hazır.')
  } catch (e: any) {
    error.value = e?.message ?? 'QR kodu üretilemedi.'
  } finally {
    busy.value = false
  }
}

async function remove() {
  await ask({
    title: 'QR kodu silinsin mi?',
    subject: display(link.value.alias),
    message: 'Görsel sistemden kaldırılır. Kısa adres çalışmaya devam eder, istersen yeni bir QR üretebilirsin.',
    confirmLabel: 'QR kodunu sil',
    tone: 'danger',
    action: async () => {
      await deleteQr(link.value.id)
      link.value = { ...link.value, hasQr: false, qrUrl: '' }
      toast.success('QR kodu silindi.')
    }
  })
}
</script>

<template>
  <div class="grid grid-cols-1 gap-8 md:grid-cols-[18rem_minmax(0,1fr)] md:items-center">
    <div class="mat-grid flex aspect-square w-full max-w-72 items-center justify-center p-8">
      <div class="relative flex aspect-square w-full items-center justify-center bg-white">
        <template v-if="link.hasQr && nonce">
          <img
            v-show="imageReady"
            :src="src"
            :alt="`${link.alias} için QR kodu`"
            class="h-full w-full"
            @load="imageReady = true"
            @error="imageFailed = true"
          />
          <p v-if="imageFailed" class="px-4 text-center text-sm text-[#b3261e]">Görsel yüklenemedi. Yeniden üretmeyi dene.</p>
          <UiSpinner v-else-if="!imageReady" :size="24" class="absolute text-[#0c5c47]" />
        </template>
        <UiIcon v-else name="qr" :size="64" class="text-[#a9b8b0]" />
      </div>
    </div>

    <div>
      <h3 class="display text-2xl">{{ link.hasQr ? 'QR kodu hazır' : 'QR kodu yok' }}</h3>
      <p class="mt-2 max-w-md text-ink-2">
        <template v-if="link.hasQr">512 × 512 piksel PNG. Kısa adı değiştirirsen kod kendiliğinden yenilenir, eski baskılar çalışmaz.</template>
        <template v-else-if="canQr">Kısa adresi açan, 512 × 512 piksel bir PNG üretilir.</template>
        <template v-else>QR kodu paketinde kapalı. Pro ve Kurumsal paketlerde açıktır.</template>
      </p>

      <div class="mt-6 flex flex-wrap items-center gap-2">
        <template v-if="link.hasQr">
          <UiButton :href="downloadUrl" icon="download">PNG indir</UiButton>
          <UiButton v-if="canQr" variant="outline" icon="refresh" :loading="busy" @click="generate">Yeniden üret</UiButton>
          <UiButton v-if="canQr" variant="danger" icon="trash" :disabled="busy" @click="remove">Sil</UiButton>
        </template>
        <UiButton v-else-if="canQr" icon="qr" :loading="busy" loading-label="Üretiliyor" @click="generate">QR kodu üret</UiButton>
        <UiButton v-else to="/dashboard/subscription" variant="outline">Paketleri gör</UiButton>
      </div>
      <p class="mt-3 min-h-5 text-sm text-danger" aria-live="polite">{{ error }}</p>
    </div>
  </div>
</template>
