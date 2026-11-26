<script setup lang="ts">
import type { WaitInfo } from '~/types/link'

definePageMeta({ layout: 'visitor' })

const route = useRoute()
const { public: config } = useRuntimeConfig()
const { display, full } = useShortUrl()
const alias = computed(() => String(route.params.alias ?? ''))
const ticket = computed(() => String(route.query.t ?? ''))

useHead({ title: () => `Yönlendiriliyor: ${alias.value} | makeiTiny` })

const loading = ref(true)
const going = ref(false)
const error = ref('')
const expired = ref(false)
const total = ref(0)
const left = ref(0)
const adEnabled = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

const progress = computed(() => (total.value > 0 ? (left.value / total.value) * 100 : 0))

function stop() {
  if (timer) clearInterval(timer)
  timer = null
}

async function fail(e: any) {
  const code = e?.data?.code
  if (code === 'ticket_invalid') {
    expired.value = true
    return
  }
  if (code === 'link_locked') {
    await navigateTo(`/unlock/${encodeURIComponent(alias.value)}`, { replace: true })
    return
  }
  const reason = reasonFromCode(code)
  if (reason) {
    await navigateTo(`/error/${encodeURIComponent(alias.value)}?reason=${reason}`, { replace: true })
    return
  }
  error.value = e?.data?.message ?? 'Link bilgisi alınamadı.'
}

// Hedef adres yalnızca bilet süresi dolunca, tek kullanımlık olarak verilir.
async function go() {
  if (going.value || left.value > 0) return
  going.value = true
  try {
    const res = await $fetch<{ target: string }>(`${config.routerBase}/${encodeURIComponent(alias.value)}/continue`, {
      method: 'POST',
      body: { ticket: ticket.value }
    })
    window.location.href = res.target
  } catch (e: any) {
    going.value = false
    if (e?.data?.code === 'wait_not_finished') {
      await start()
      return
    }
    await fail(e)
  }
}

async function start() {
  stop()
  loading.value = true
  error.value = ''
  expired.value = false
  if (!ticket.value) {
    expired.value = true
    loading.value = false
    return
  }
  try {
    const info = await $fetch<WaitInfo>(`${config.routerBase}/${encodeURIComponent(alias.value)}/wait`, { query: { t: ticket.value } })
    total.value = Math.max(info.waitSeconds, info.remainingSeconds)
    left.value = info.remainingSeconds
    adEnabled.value = info.adEnabled
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
    await fail(e)
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
      <template v-else-if="expired">
        <UiIcon name="clock" :size="36" class="text-ink-3" />
        <p class="mt-4 text-lg">Bu bekleme sayfasının süresi geçmiş.</p>
        <p class="mt-1 text-ink-2">Linki baştan açarsan geri sayım yeniden başlar.</p>
        <UiButton :href="full(alias)" icon="refresh" class="mt-5">Linki baştan aç</UiButton>
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

    <!-- Reklam alanı yalnızca link sahibi reklamı açtıysa gösterilir. -->
    <div v-if="adEnabled" class="mt-8 flex min-h-36 items-center justify-center border border-dashed border-line-2 text-sm text-ink-3">Reklam alanı</div>

    <UiButton size="lg" block class="mt-8" :loading="going" loading-label="Hedefe gidiliyor" :disabled="loading || !!error || expired || left > 0" @click="go">
      {{ left > 0 || loading ? 'Bekleniyor' : 'Hedefe git' }}
    </UiButton>
  </div>
</template>
