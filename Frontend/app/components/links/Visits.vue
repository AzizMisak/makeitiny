<script setup lang="ts">
import type { LinkLogItem, PaginatedLogsResponse } from '~/types/link'

/** Ziyaret kayıtları tablosu. Kullanıcı ve yönetim panelinde aynı tablo, farklı kaynak. */
const props = defineProps<{ fetchPage: (page: number, limit: number) => Promise<PaginatedLogsResponse>; windowLabel?: string }>()

const PAGE_SIZE = 20
const rows = ref<LinkLogItem[]>([])
const page = ref(1)
const total = ref(0)
const totalPages = ref(1)
const loading = ref(true)
const loadedOnce = ref(false)
const error = ref('')

async function load(next = page.value) {
  loading.value = true
  error.value = ''
  try {
    const res = await props.fetchPage(next, PAGE_SIZE)
    rows.value = res.items ?? []
    page.value = res.page ?? 1
    total.value = res.total ?? 0
    totalPages.value = Math.max(1, res.totalPages ?? 1)
    loadedOnce.value = true
  } catch (e: any) {
    error.value = e?.message ?? 'Ziyaret kayıtları yüklenemedi.'
  } finally {
    loading.value = false
  }
}

function deviceIcon(type: string): string {
  return /mobil|mobile|phone|tablet/i.test(type) ? 'phone' : 'monitor'
}

onMounted(() => load(1))
</script>

<template>
  <div>
    <div class="flex min-h-10 items-center justify-between gap-4">
      <p class="tnum text-sm text-ink-3">
        <template v-if="loadedOnce">{{ windowLabel ? `${windowLabel}, ` : '' }}{{ formatNumber(total) }} ziyaret</template>
      </p>
      <UiIconButton icon="refresh" label="Kayıtları yenile" :loading="loading" @click="load()" />
    </div>

    <UiLoadError v-if="error" :message="error" :loading="loading" @retry="load()" />

    <div v-else-if="!loadedOnce" class="mt-2 flex flex-col gap-2" aria-busy="true" aria-label="Ziyaretler yükleniyor">
      <div v-for="i in 8" :key="i" class="skeleton h-11" />
    </div>

    <UiEmpty v-else-if="rows.length === 0" icon="eye" title="Ziyaret kaydı yok" text="Linke tıklandığında cihaz ve kaynak bilgisi burada listelenir." />

    <template v-else>
      <div class="mt-2 overflow-x-auto transition-opacity" :class="loading ? 'opacity-50' : ''" :aria-busy="loading">
        <table class="w-full min-w-[46rem] text-left text-sm">
          <thead class="text-ink-3">
            <tr class="border-b border-line">
              <th class="py-2 pr-4 font-medium">Zaman</th>
              <th class="py-2 pr-4 font-medium">Cihaz</th>
              <th class="py-2 pr-4 font-medium">Geldiği yer</th>
              <th class="py-2 pr-4 font-medium">Ülke ve dil</th>
              <th class="py-2 font-medium">IP adresi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-line">
            <tr v-for="r in rows" :key="r.id" :title="r.userAgent">
              <td class="tnum whitespace-nowrap py-3 pr-4">{{ formatDateTimeTr(r.clickedAt) }}</td>
              <td class="py-3 pr-4">
                <span class="flex items-center gap-2.5 whitespace-nowrap">
                  <UiIcon :name="deviceIcon(r.deviceType)" :size="17" class="text-ink-3" />
                  {{ [r.browser, r.os].filter(Boolean).join(', ') || 'Bilinmiyor' }}
                </span>
              </td>
              <td class="max-w-52 truncate py-3 pr-4 text-ink-2" :title="r.referer">{{ r.referer || 'Doğrudan' }}</td>
              <td class="whitespace-nowrap py-3 pr-4 text-ink-2">
                {{ resolveCountry(r).name }}<span v-if="resolveCountry(r).lang" class="text-ink-3">, {{ resolveCountry(r).lang }}</span>
              </td>
              <td class="whitespace-nowrap py-3 font-mono text-[13px] text-ink-2">{{ r.ipAddress || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <UiPagination
        :page="page"
        :total-pages="totalPages"
        :total="total"
        :page-size="PAGE_SIZE"
        :loading="loading"
        item-label="ziyaret"
        @change="load"
      />
    </template>
  </div>
</template>
