<script setup lang="ts">
/** Ziyaretçinin gördüğünü gerçekten çalışan küçük bir pencerede gösterir. Sunucuya istek atmaz. */
type Mode = 'direct' | 'password' | 'wait' | 'limit'

const DEMO_ALIAS = 'yaz26'
const DEMO_TARGET = 'magaza.ornek.com/yaz-koleksiyonu'
const DEMO_PASSWORD = '1234'

const { display } = useShortUrl()

const modes: { value: Mode; title: string; text: string }[] = [
  { value: 'direct', title: 'Doğrudan', text: 'Tıklar, beklemeden hedefe geçer.' },
  { value: 'password', title: 'Parolalı', text: `Parolayı bilen geçer. Dene: ${DEMO_PASSWORD}` },
  { value: 'wait', title: 'Bekletmeli', text: 'Geri sayım biter, hedef açılır.' },
  { value: 'limit', title: 'Sınırlı', text: 'Son kullanma tarihi ya da tıklama hakkı bitince link kapanır.' }
]

const mode = ref<Mode>('direct')
const arrived = ref(false)
const typed = ref('')
const wrong = ref(false)
const left = ref(5)
const tickets = ref(3)
let timer: ReturnType<typeof setInterval> | null = null

function stop() {
  if (timer) clearInterval(timer)
  timer = null
}

function reset() {
  stop()
  arrived.value = false
  typed.value = ''
  wrong.value = false
  left.value = 5
  tickets.value = 3
}

function select(m: Mode) {
  mode.value = m
  reset()
}

function startWait() {
  stop()
  timer = setInterval(() => {
    left.value -= 1
    if (left.value <= 0) {
      stop()
      arrived.value = true
    }
  }, 1000)
}

function open() {
  if (mode.value === 'direct') arrived.value = true
  if (mode.value === 'wait') startWait()
  if (mode.value === 'limit' && tickets.value > 0) tickets.value -= 1
}

function unlock() {
  if (typed.value === DEMO_PASSWORD) arrived.value = true
  else wrong.value = true
}

const counting = computed(() => mode.value === 'wait' && left.value < 5 && !arrived.value)
const address = computed(() => {
  if (arrived.value) return DEMO_TARGET
  return display(DEMO_ALIAS)
})

onBeforeUnmount(stop)
</script>

<template>
  <section id="ziyaretci" class="py-20 md:py-28">
    <div class="container-x lg:pr-20">
      <h2 class="display max-w-3xl text-[clamp(2.25rem,5.5vw,4rem)]">Tıklayan ne görsün, sen seç</h2>

      <div class="mt-12 grid grid-cols-1 gap-10 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)] lg:items-center">
        <ul class="flex flex-col">
          <li v-for="m in modes" :key="m.value" class="border-t border-line last:border-b">
            <button
              type="button"
              class="block w-full py-4 text-left transition-colors"
              :aria-pressed="mode === m.value"
              @click="select(m.value)"
            >
              <span
                class="display block text-[clamp(1.75rem,4vw,2.75rem)] transition-[font-stretch,color] duration-300"
                :class="mode === m.value ? 'text-ink' : 'text-ink-3 hover:text-ink-2'"
                :style="{ fontStretch: mode === m.value ? '125%' : '62%' }"
              >{{ m.title }}</span>
              <span class="mt-1 block text-ink-2">{{ m.text }}</span>
            </button>
          </li>
        </ul>

        <div class="overflow-hidden rounded-lg border border-line-2 bg-surface">
          <div class="flex items-center gap-3 border-b border-line bg-sunken px-4 py-3">
            <UiIconButton icon="refresh" label="Örneği baştan al" :size="16" class="!h-8 !w-8" @click="reset" />
            <p class="min-w-0 flex-1 truncate rounded-full bg-surface px-4 py-1.5 font-mono text-[13px] text-ink-2" aria-live="polite">{{ address }}</p>
          </div>

          <div class="relative h-80">
            <!-- Hedef -->
            <div v-if="arrived" class="flex h-full flex-col items-center justify-center gap-3 px-6 text-center">
              <UiIcon name="check" :size="40" class="text-ok" />
              <p class="text-xl font-semibold">Hedefe ulaştı</p>
              <p class="text-ink-2">Tıklama, cihaz bilgisiyle birlikte kaydedildi.</p>
            </div>

            <div v-else-if="mode === 'direct'" class="flex h-full flex-col items-center justify-center gap-5 px-6 text-center">
              <p class="text-ink-2">Arkadaşın sana bir link gönderdi:</p>
              <button type="button" class="strip-cut bg-mark py-3.5 pl-5 pr-9 font-mono text-lg font-semibold text-on-mark hover:bg-mark-hover" @click="open">
                {{ display(DEMO_ALIAS) }}
              </button>
              <p class="text-sm text-ink-3">Üstüne tıkla</p>
            </div>

            <form v-else-if="mode === 'password'" class="mx-auto flex h-full max-w-xs flex-col justify-center" novalidate @submit.prevent="unlock">
              <UiIcon name="lock" :size="30" class="text-brand" />
              <p class="display mt-3 text-2xl">Parola gerekiyor</p>
              <UiField
                id="demo-password"
                v-model="typed"
                class="mt-4"
                type="password"
                autocomplete="off"
                placeholder="Örnek parola: 1234"
                :error="wrong ? 'Parola hatalı.' : ''"
                @update:model-value="wrong = false"
              />
              <UiButton type="submit" icon="unlock" block>Kilidi aç</UiButton>
            </form>

            <div v-else-if="mode === 'wait'" class="flex h-full flex-col items-center justify-center px-6 text-center">
              <p class="display tnum text-[7rem] leading-none">{{ left }}</p>
              <p class="mt-1 text-ink-2">saniye sonra hedefe geçiyorsun</p>
              <div class="mt-6 h-3 w-full max-w-sm bg-sunken">
                <div class="h-full bg-mark transition-[width] duration-1000 ease-linear" :style="{ width: `${(left / 5) * 100}%` }" />
              </div>
              <div class="mt-6 h-11">
                <UiButton v-if="!counting" variant="outline" @click="open">Sayacı başlat</UiButton>
              </div>
            </div>

            <div v-else class="flex h-full flex-col items-center justify-center px-6 text-center">
              <p class="display text-[2.25rem]">{{ tickets > 0 ? `${tickets} tıklama hakkı var` : 'Bu linkin hakkı bitmiş' }}</p>
              <div class="mt-6 flex gap-2" aria-hidden="true">
                <span v-for="i in 3" :key="i" class="h-4 w-16 transition-colors" :class="i <= tickets ? 'bg-mark' : 'bg-sunken'" />
              </div>
              <div class="mt-8 flex h-11 items-center">
                <UiButton v-if="tickets > 0" variant="outline" @click="open">Linke tıkla</UiButton>
                <p v-else class="text-ink-2">Ziyaretçi artık hedefe gidemez.</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
