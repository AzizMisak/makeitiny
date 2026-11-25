<script setup lang="ts">
/** Sayfayı açan kişinin kendi cihazıyla, bir tıklamanın ne kaydettiğini gösterir. Hiçbir yere gönderilmez. */
interface DemoClick {
  id: number
  time: string
  device: string
  browser: string
  os: string
  country: string
  lang: string
}

const { display } = useShortUrl()
const me = ref<Omit<DemoClick, 'id' | 'time'> | null>(null)
const clicks = ref<DemoClick[]>([])

function readDevice() {
  const ua = navigator.userAgent
  const browser = /Edg\//.test(ua) ? 'Edge' : /OPR\//.test(ua) ? 'Opera' : /Firefox\//.test(ua) ? 'Firefox' : /Chrome\//.test(ua) ? 'Chrome' : /Safari\//.test(ua) ? 'Safari' : 'Diğer'
  const os = /Windows/.test(ua) ? 'Windows' : /Android/.test(ua) ? 'Android' : /iPhone|iPad/.test(ua) ? 'iOS' : /Mac OS/.test(ua) ? 'macOS' : /Linux/.test(ua) ? 'Linux' : 'Diğer'
  const device = /iPad|Tablet/.test(ua) ? 'Tablet' : /Mobi|iPhone|Android/.test(ua) ? 'Telefon' : 'Masaüstü'
  const lang = navigator.language || ''
  me.value = { device, browser, os, lang, country: resolveCountry({ countryCode: '', acceptLanguage: lang }).name }
}

function add() {
  if (!me.value) readDevice()
  if (!me.value) return
  const now = new Date()
  clicks.value = [
    { id: now.getTime(), time: now.toLocaleTimeString('tr-TR'), ...me.value },
    ...clicks.value
  ]
}

onMounted(readDevice)
</script>

<template>
  <section id="kayit" class="bg-sunken py-20 md:py-28">
    <div class="container-x lg:pr-20">
      <div class="grid grid-cols-1 gap-12 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.15fr)] lg:items-center">
        <div>
          <h2 class="display text-[clamp(2.25rem,5.5vw,4rem)]">Her tıklama bir satır</h2>
          <p class="mt-5 max-w-md text-lg text-ink-2">
            Linkine kim, ne zaman, hangi cihazla geldi; panelinde satır satır durur. Şu linke bas, senin tıklaman nasıl görünürdü bak.
          </p>
          <button
            type="button"
            class="strip-cut mt-8 bg-mark py-4 pl-6 pr-12 font-mono text-[clamp(1.1rem,3vw,1.5rem)] font-semibold text-on-mark transition-transform hover:bg-mark-hover active:translate-y-0.5"
            @click="add"
          >{{ display('yaz26') }}</button>
          <p class="mt-4 text-sm text-ink-3">Bu bir gösterim. Tıklaman kaydedilmez, hiçbir yere gönderilmez.</p>
        </div>

        <div class="bg-surface p-6 md:p-8">
          <p class="flex items-baseline gap-3" aria-live="polite">
            <span class="display tnum text-7xl">{{ clicks.length }}</span>
            <span class="text-lg text-ink-2">tıklama</span>
          </p>

          <div class="mt-6 h-64 overflow-hidden">
            <table class="w-full text-left text-sm">
              <thead class="text-ink-3">
                <tr class="border-b border-line">
                  <th class="py-2 pr-4 font-medium">Zaman</th>
                  <th class="py-2 pr-4 font-medium">Cihaz</th>
                  <th class="py-2 pr-4 font-medium max-sm:hidden">Tarayıcı, sistem</th>
                  <th class="py-2 font-medium">Ülke</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-line">
                <tr v-for="c in clicks.slice(0, 5)" :key="c.id">
                  <td class="tnum py-3 pr-4 font-medium">{{ c.time }}</td>
                  <td class="py-3 pr-4">{{ c.device }}</td>
                  <td class="py-3 pr-4 max-sm:hidden">{{ c.browser }}, {{ c.os }}</td>
                  <td class="py-3">{{ c.country }}<span class="text-ink-3">, {{ c.lang }}</span></td>
                </tr>
                <tr v-if="clicks.length === 0">
                  <td colspan="4" class="py-10 text-center text-ink-3">Linke bastığında ilk satır buraya düşer.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
