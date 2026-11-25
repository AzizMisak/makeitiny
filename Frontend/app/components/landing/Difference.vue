<script setup lang="ts">
import type { PlanItem } from '~/types/auth'

const { isLoggedIn } = useAuth()
const { fetchPlans } = useAdmin()

// Paket sayıları yönetimden değişebildiği için API'den okunur; gelene kadar varsayılanlar görünür.
const free = ref<PlanItem | null>(null)

onMounted(async () => {
  const plans = await fetchPlans().catch(() => [])
  free.value = plans.find((p) => p.slug === 'free') ?? null
})

const rows = computed(() => {
  const max = free.value?.maxLinks ?? 50
  return [
    { label: 'Link sayısı', guest: 'Günde 1', member: max < 0 ? 'Sınırsız' : `${formatNumber(max)} link` },
    { label: 'Ömrü', guest: 'En fazla 48 saat', member: 'Sen silene kadar' },
    { label: 'Sonradan düzenleme', guest: 'Yok', member: 'Hedef, parola, sınırlar' },
    { label: 'Ziyaret kayıtları', guest: 'Yok', member: `Son ${formatNumber(free.value?.analyticsDays ?? 7)} gün` },
    { label: 'Parola ve bekletme', guest: 'Var', member: (free.value?.allowPassword ?? true) ? 'Var' : 'Yalnızca bekletme' }
  ]
})
</script>

<template>
  <section id="fark" class="overflow-hidden py-20 md:py-28">
    <div class="container-x lg:pr-20">
      <h2 class="display max-w-3xl text-[clamp(2.25rem,5.5vw,4rem)]">Geçici olan kısa, hesaplı olan uzun ömürlü</h2>

      <div class="mt-14 flex flex-col gap-8">
        <div>
          <p class="mb-2 font-semibold">Üyeliksiz</p>
          <div class="tape-ticks relative h-14 w-40 bg-mark text-on-mark"><span class="absolute bottom-1 right-2 text-sm font-semibold">48 saat</span></div>
        </div>
        <div>
          <p class="mb-2 font-semibold">Ücretsiz hesap</p>
          <!-- Şerit sayfadan taşar: bitiş tarihi yok. -->
          <div class="tape-ticks relative h-14 w-[200vw] bg-mark text-on-mark" />
        </div>
      </div>

      <div class="mt-14 overflow-x-auto">
        <table class="w-full min-w-[32rem] text-left">
          <thead>
            <tr class="border-b-2 border-ink">
              <td class="w-1/3" />
              <th class="py-3 pr-4 font-semibold">Üyeliksiz</th>
              <th class="py-3 font-semibold">Ücretsiz hesap</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-line">
            <tr v-for="r in rows" :key="r.label">
              <th scope="row" class="py-3.5 pr-4 font-normal text-ink-2">{{ r.label }}</th>
              <td class="tnum py-3.5 pr-4" :class="r.guest === 'Yok' ? 'text-ink-3' : ''">{{ r.guest }}</td>
              <td class="tnum py-3.5 font-semibold">{{ r.member }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="mt-4 text-ink-2">Kendi kısa adını seçmek ve hesabındaki linklere QR kodu üretmek ücretli paketlerde açılır.</p>

      <UiButton v-if="isLoggedIn" to="/dashboard/new" size="lg" icon="plus" class="mt-10">Yeni link</UiButton>
      <UiButton v-else to="/auth/register" size="lg" class="mt-10">Ücretsiz hesap aç</UiButton>
    </div>
  </section>
</template>

<style scoped>
.tape-ticks {
  background-image:
    linear-gradient(90deg, rgb(42 31 0 / 0.75) 1px, transparent 1px),
    linear-gradient(90deg, rgb(42 31 0 / 0.6) 1px, transparent 1px);
  background-size:
    80px 26px,
    16px 14px;
  background-repeat: repeat-x;
}
</style>
