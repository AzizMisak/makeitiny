<script setup lang="ts">
import type { LinkLogItem } from '~/types/link'
import type { TimeRange } from '~/utils/analytics'

const props = defineProps<{ logs: LinkLogItem[]; total: number; loading: boolean; error: string; windowLabel: string }>()
defineEmits<{ retry: [] }>()

const range = ref<TimeRange>('7d')
const ranges: { value: TimeRange; label: string }[] = [
  { value: '7d', label: '7 gün' },
  { value: '14d', label: '14 gün' },
  { value: '30d', label: '30 gün' },
  { value: 'hourly', label: 'Günün saatleri' }
]

const buckets = computed(() => timeBuckets(props.logs, range.value))
const devices = computed(() => shareRows(props.logs.map((l) => l.deviceType)))
const browsers = computed(() => shareRows(props.logs.map((l) => l.browser)))
const systems = computed(() => shareRows(props.logs.map((l) => l.os)))
const countries = computed(() => countryRows(props.logs))
const regions = computed(() => shareRows(props.logs.map((l) => resolveCountry(l).region)))

// Analiz en yeni kayıtlardan yapılır; hepsi okunamadıysa bunu söyle.
const partial = computed(() => props.total > props.logs.length)
</script>

<template>
  <div>
    <UiLoadError v-if="error" :message="error" :loading="loading" @retry="$emit('retry')" />

    <div v-else-if="loading" aria-busy="true" aria-label="Analiz yükleniyor">
      <div class="skeleton h-10 w-64" />
      <div class="skeleton mt-6 h-56" />
      <div class="mt-10 grid grid-cols-1 gap-10 md:grid-cols-3"><div v-for="i in 3" :key="i" class="skeleton h-40" /></div>
    </div>

    <UiEmpty
      v-else-if="logs.length === 0"
      icon="activity"
      title="Henüz tıklama yok"
      :text="`${windowLabel} içinde bu linke tıklanmadı. İlk tıklamadan sonra grafikler burada çizilir.`"
    />

    <template v-else>
      <div class="flex flex-wrap items-center justify-between gap-4">
        <UiSegmented v-model="range" label="Zaman aralığı" :options="ranges" small />
        <p class="tnum text-sm text-ink-3">
          {{ windowLabel }}<template v-if="partial">, en yeni {{ formatNumber(logs.length) }} ziyaret</template>
        </p>
      </div>

      <ChartBars class="mt-6" :buckets="buckets" label="Zamana göre tıklama" />

      <!-- Her kırılım kendi ince çerçevesinde; yan yana durunca karışmaz. -->
      <div class="mt-10 grid grid-cols-1 gap-4 md:grid-cols-3">
        <ChartShares title="Cihaz" :rows="devices" class="rounded-md border border-line bg-surface p-5" />
        <ChartShares title="Tarayıcı" :rows="browsers" class="rounded-md border border-line bg-surface p-5" />
        <ChartShares title="İşletim sistemi" :rows="systems" class="rounded-md border border-line bg-surface p-5" />
      </div>

      <section class="mt-4 rounded-md border border-line bg-surface p-5">
        <h3 class="font-semibold">Ülkeler</h3>
        <p class="mt-1 text-sm text-ink-3">Ülke, ziyaretçinin tarayıcı dilinden tahmin edilir; kesin konum değildir.</p>

        <div class="mt-5 grid grid-cols-1 gap-x-10 gap-y-8 lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
          <div class="overflow-x-auto">
            <table class="w-full min-w-[34rem] text-left text-sm">
              <thead class="text-ink-3">
                <tr class="border-b border-line">
                  <th class="py-2 pr-4 font-medium">Ülke</th>
                  <th class="py-2 pr-4 font-medium">Tarayıcı dilleri</th>
                  <th class="py-2 pr-4 font-medium">Son ziyaret</th>
                  <th class="py-2 text-right font-medium">Tıklama</th>
                  <th class="py-2 pl-4 text-right font-medium">Pay</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-line">
                <tr v-for="c in countries" :key="c.label">
                  <td class="py-2.5 pr-4 font-medium">{{ c.label }}</td>
                  <td class="max-w-44 truncate py-2.5 pr-4 text-ink-2" :title="c.langs.join(', ')">{{ c.langs.join(', ') || '—' }}</td>
                  <td class="tnum whitespace-nowrap py-2.5 pr-4 text-ink-2">{{ formatDateTimeTr(c.lastSeen) }}</td>
                  <td class="tnum py-2.5 text-right font-semibold">{{ formatNumber(c.count) }}</td>
                  <td class="tnum py-2.5 pl-4 text-right text-ink-2">%{{ c.percent.toLocaleString('tr-TR') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <ChartShares title="Bölgeler" :rows="regions" :limit="7" class="lg:border-l lg:border-line lg:pl-10" />
        </div>
      </section>
    </template>
  </div>
</template>
