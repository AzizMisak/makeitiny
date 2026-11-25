<script setup lang="ts">
import type { PlanItem } from '~/types/auth'

definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Paketim | makeiTiny' })

const { user, isAdmin, remainingLinks, fetchCurrentUser } = useAuth()
const { fetchPlans } = useAdmin()

const plans = ref<PlanItem[]>([])
const loading = ref(true)
const error = ref('')

const used = computed(() => user.value?.linkCount ?? 0)
const max = computed(() => (isAdmin.value ? -1 : (user.value?.plan?.maxLinks ?? 50)))

const usageText = computed(() => {
  if (isAdmin.value) return 'Yönetici hesapları paket sınırlarından muaftır.'
  if (max.value < 0) return 'Bu pakette link sınırı yok.'
  return `${formatNumber(remainingLinks.value ?? 0)} link hakkın kaldı.`
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [list] = await Promise.all([fetchPlans(), fetchCurrentUser()])
    plans.value = list
  } catch (e: any) {
    error.value = e?.message ?? 'Paket bilgileri yüklenemedi.'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PanelHeader :title="user?.plan?.name ?? 'Paketim'">
      <p class="text-ink-2">
        {{ formatPriceMonthly(user?.plan?.priceMonthly ?? 0) }}<template v-if="user?.planExpiresAt">, {{ formatDateTr(user.planExpiresAt) }} tarihine kadar</template>
      </p>
    </PanelHeader>

    <section>
      <p class="flex flex-wrap items-baseline gap-x-3">
        <span class="display tnum text-6xl">{{ formatNumber(used) }}</span>
        <span class="text-lg text-ink-2">{{ max < 0 ? 'link kullandın' : `/ ${formatNumber(max)} link` }}</span>
      </p>
      <UiTape class="mt-5" :used="used" :max="max" label="Link kullanımı" />
      <p class="mt-3 text-ink-2">{{ usageText }}</p>
    </section>

    <section class="mt-16">
      <h2 class="display text-2xl">Paketleri karşılaştır</h2>
      <p class="mt-2 text-ink-2">Paket değişikliği şimdilik yönetim tarafından yapılıyor.</p>

      <UiLoadError v-if="error" :message="error" :loading="loading" class="mt-6" @retry="load" />
      <div v-else-if="loading" class="mt-6 flex flex-col gap-2" aria-busy="true" aria-label="Paketler yükleniyor">
        <div class="skeleton h-20" />
        <div v-for="i in 6" :key="i" class="skeleton h-11" />
      </div>

      <div v-else class="mt-6 overflow-x-auto">
        <table class="w-full min-w-[38rem] table-fixed border-collapse text-left">
          <thead>
            <tr>
              <td class="w-1/4" />
              <th
                v-for="p in plans"
                :key="p.id"
                scope="col"
                class="px-4 pb-4 pt-5 align-bottom"
                :class="p.id === user?.plan?.id ? 'bg-brand-soft' : ''"
              >
                <span class="block min-h-6 text-[13px] font-medium text-brand-ink">{{ p.id === user?.plan?.id ? 'Şu anki paketin' : '' }}</span>
                <span class="display block text-xl">{{ p.name }}</span>
                <span class="tnum mt-1 block text-sm font-normal text-ink-2">{{ formatPriceMonthly(p.priceMonthly) }}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in PLAN_ROWS" :key="row.key" class="border-t border-line">
              <th scope="row" class="py-3.5 pr-4 text-[15px] font-normal text-ink-2">{{ row.label }}</th>
              <td v-for="p in plans" :key="p.id" class="px-4 py-3.5" :class="p.id === user?.plan?.id ? 'bg-brand-soft' : ''">
                <span v-if="row.kind === 'text'" class="tnum font-semibold">{{ row.value(p) }}</span>
                <span v-else-if="row.value(p)" class="inline-flex items-center gap-2 text-ok"><UiIcon name="check" /> Var</span>
                <span v-else class="text-ink-3">Yok</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
