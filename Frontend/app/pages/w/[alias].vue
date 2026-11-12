<script setup lang="ts">
import type { LinkPreview } from '~/types/link'

definePageMeta({ layout: 'marketing' })

const route = useRoute()
const { public: config } = useRuntimeConfig()
const alias = computed(() => String(route.params.alias ?? ''))

useHead({ title: () => `Yönlendiriliyor (${alias.value}) — makeiTiny` })

const loading = ref(true)
const errorMsg = ref('')
const targetUrl = ref('')
const totalSeconds = ref(0)
const remainingSeconds = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

function clearTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

function goToTarget() {
  if (!targetUrl.value || remainingSeconds.value > 0) return
  window.location.href = targetUrl.value
}

async function initWaitPage() {
  loading.value = true
  errorMsg.value = ''
  try {
    const preview = await $fetch<LinkPreview>(`${config.routerBase}/${alias.value}/preview`)
    if (preview.locked) {
      await navigateTo(`/unlock/${alias.value}`, { replace: true })
      return
    }

    targetUrl.value = preview.target
    const secs = preview.waitEnabled && preview.waitSeconds > 0 ? preview.waitSeconds : 0
    totalSeconds.value = secs
    remainingSeconds.value = secs

    if (secs <= 0 && preview.target) {
      window.location.href = preview.target
      return
    }

    timer = setInterval(() => {
      if (remainingSeconds.value > 1) {
        remainingSeconds.value -= 1
      } else {
        remainingSeconds.value = 0
        clearTimer()
        if (targetUrl.value) {
          window.location.href = targetUrl.value
        }
      }
    }, 1000)
  } catch (e: any) {
    errorMsg.value = e?.data?.message ?? 'Aradığınız kısa link bulunamadı veya önizleme bilgisi alınamadı.'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  initWaitPage()
})

onBeforeUnmount(() => {
  clearTimer()
})
</script>

<template>
  <AuthCard
    title="Hedef adrese yönlendiriliyorsunuz"
    :subtitle="`/${alias} bağlantısı için lütfen bekleyin.`"
  >
    <div v-if="loading" class="flex items-center justify-center gap-3 py-6 text-ink-2">
      <span class="h-5 w-5 animate-spin rounded-full border-2 border-ink border-t-transparent" />
      <span class="text-sm font-medium">Bağlantı hazırlanıyor...</span>
    </div>

    <div v-else-if="errorMsg" class="flex flex-col gap-4">
      <p class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
        {{ errorMsg }}
      </p>
      <UiButton to="/" variant="secondary" block>Ana sayfaya dön</UiButton>
    </div>

    <div v-else class="flex flex-col gap-5">
      <!-- Gelecekte reklam konulacak alan -->
      <div class="flex min-h-[140px] flex-col items-center justify-center rounded-md border-[1.5px] border-dashed border-ink bg-paper p-5 text-center">
        <UiBadge tone="neutral">SPONSORLU ALAN</UiBadge>
        <p class="mt-2 text-sm font-semibold text-ink-2">Reklam Alanı</p>
        <p class="mt-1 text-xs text-ink-3">
          Bu alan ileride reklam ve tanıtım içerikleri için kullanılacaktır.
        </p>
      </div>

      <!-- Geri sayım göstergesi -->
      <div class="flex flex-col items-center justify-center rounded-md border border-line bg-paper-2 p-4 text-center">
        <template v-if="remainingSeconds > 0">
          <span class="font-display text-4xl font-extrabold text-accent">{{ remainingSeconds }}</span>
          <span class="mt-1 text-xs font-medium text-ink-2">saniye sonra otomatik yönlendirileceksiniz...</span>
        </template>
        <template v-else>
          <span class="text-sm font-bold text-ok">Süre doldu! Yönlendiriliyorsunuz...</span>
        </template>
      </div>

      <UiButton
        type="button"
        block
        :disabled="remainingSeconds > 0"
        :class="{ 'opacity-50 cursor-not-allowed': remainingSeconds > 0 }"
        @click="goToTarget"
      >
        <template v-if="remainingSeconds > 0">
          <UiIcon name="clock" :size="16" /> Lütfen Bekleyin ({{ remainingSeconds }} sn)
        </template>
        <template v-else>
          Hedef Adrese Git <UiIcon name="arrow" :size="16" />
        </template>
      </UiButton>
    </div>

    <template #footer>
      <NuxtLink to="/" class="font-semibold text-accent-dark underline underline-offset-2">Ana sayfaya dön</NuxtLink>
    </template>
  </AuthCard>
</template>
