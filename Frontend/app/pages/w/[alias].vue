<script setup lang="ts">
import type { LinkPreview } from '~/types/link'

definePageMeta({ layout: 'visitor' })

const route = useRoute()
const { public: config } = useRuntimeConfig()
const { display } = useShortUrl()
const alias = computed(() => String(route.params.alias ?? ''))

useHead({ title: () => `Yönlendiriliyor: ${alias.value} | makeiTiny` })

const loading = ref(true)
const error = ref('')
const target = ref('')
const total = ref(0)
const left = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const progress = computed(() => (total.value > 0 ? (left.value / total.value) * 100 : 0))

function stop() {
  if (timer) clearInterval(timer)
  timer = null
}

function go() {
  if (target.value && left.value <= 0) window.location.href = target.value
}

async function start() {
  stop()
  loading.value = true
  error.value = ''
  try {
    const preview = await $fetch<LinkPreview>(`${config.routerBase}/${alias.value}/preview`)
    if (preview.locked) {
      await navigateTo(`/unlock/${alias.value}`, { replace: true })
      return
    }
    target.value = preview.target
    total.value = preview.waitEnabled && preview.waitSeconds > 0 ? preview.waitSeconds : 0
    left.value = total.value
    if (left.value <= 0) {
      go()
      return
    }
    timer = setInterval(() => {
      left.value = Math.max(0, left.value - 1)
      if (left.value === 0) {
        stop()
        go()
      }
    }, 1000)
  } catch (e: any) {
    const reason = reasonFromCode(e?.data?.code)
    if (reason) {
      await navigateTo(`/error/${alias.value}?reason=${reason}`, { replace: true })
      return
    }
    error.value = e?.data?.message ?? 'Link bilgisi alınamadı.'
  } finally {
    loading.value = false
  }
}

onMounted(start)
onBeforeUnmount(stop)
</script>

<template>
  <div class="mx-auto flex w-full max-w-2xl flex-1 flex-col justify-center px-5 pb-16">
    <p class="break-all text-center text-ink-2"><span class="font-mono text-[15px]">{{ display(alias) }}</span></p>

    <!-- Sayı, yükleme ve hata aynı alanı kullanır. -->
    <div class="mt-2 flex min-h-56 flex-col items-center justify-center text-center" aria-live="polite">
      <template v-if="loading">
        <UiSpinner :size="32" class="text-brand" />
        <p class="mt-5 text-lg text-ink-2">Link hazırlanıyor</p>
      </template>
      <template v-else-if="error">
        <UiIcon name="alert" :size="36" class="text-danger" />
        <p class="mt-4 text-lg">{{ error }}</p>
        <UiButton variant="outline" icon="refresh" class="mt-5" @click="start">Tekrar dene</UiButton>
      </template>
      <template v-else>
        <h1 class="sr-only">Hedefe yönlendiriliyorsun</h1>
        <p class="display tnum text-[9rem] leading-none" aria-hidden="true">{{ left }}</p>
        <p class="mt-2 text-lg text-ink-2">{{ left > 0 ? 'saniye sonra hedefe geçiyorsun' : 'Hedefe geçiliyor' }}</p>
      </template>
    </div>

    <div class="mt-6 h-3 bg-sunken" aria-hidden="true">
      <div class="h-full bg-mark transition-[width] duration-1000 ease-linear" :style="{ width: `${progress}%` }" />
    </div>

    <!-- Reklam için ayrılmış alan -->
    <div class="mt-8 flex min-h-36 items-center justify-center border border-dashed border-line-2 text-sm text-ink-3">Reklam alanı</div>

    <UiButton size="lg" block class="mt-8" :disabled="loading || !!error || left > 0" @click="go">
      {{ left > 0 || loading ? 'Bekleniyor' : 'Hedefe git' }}
    </UiButton>
  </div>
</template>
