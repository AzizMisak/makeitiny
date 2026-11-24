<script setup lang="ts">
import type { PlanItem } from '~/types/auth'

definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Abonelik & Paket — makeiTiny' })

const { user, isAdmin, remainingLinks, fetchCurrentUser } = useAuth()
const { fetchPlans } = useAdmin()
const toast = useToast()

const plans = ref<PlanItem[]>([])
const loading = ref(true)

const usedLinks = computed(() => user.value?.linkCount ?? 0)

const quotaValue = computed(() => {
  const max = user.value?.plan?.maxLinks ?? 50
  if (isAdmin.value || max < 0) {
    return `${usedLinks.value.toLocaleString('tr-TR')} / Sınırsız`
  }
  return `${usedLinks.value.toLocaleString('tr-TR')} / ${max.toLocaleString('tr-TR')}`
})

const quotaNote = computed(() => {
  if (isAdmin.value) return 'Kota muafiyeti'
  const max = user.value?.plan?.maxLinks ?? 50
  if (max < 0) return 'Sınırsız link'
  return `${remainingLinks.value ?? 0} kalan hak`
})

function planFeatures(plan: PlanItem) {
  return [
    {
      label: plan.maxLinks < 0 ? 'Sınırsız link' : `${plan.maxLinks.toLocaleString('tr-TR')} link kotası`,
      allowed: true
    },
    { label: `Son ${plan.analyticsDays} gün analitik`, allowed: true },
    { label: 'Özel takma ad', allowed: plan.allowCustomAlias },
    { label: 'Şifre koruması', allowed: plan.allowPassword },
    { label: 'QR kod', allowed: plan.allowQr },
    { label: 'Reklam modu', allowed: plan.allowAds }
  ]
}

async function loadSubscriptionData() {
  loading.value = true
  try {
    const [planList] = await Promise.all([
      fetchPlans(),
      fetchCurrentUser().catch(() => null)
    ])
    plans.value = planList
  } catch (err: any) {
    toast.error(err?.message ?? 'Paket bilgileri yüklenemedi.')
  } finally {
    loading.value = false
  }
}

onMounted(loadSubscriptionData)
</script>

<template>
  <DashboardPageHeader
    title="Abonelik & paket"
    description="Mevcut paketin, link kotan ve paket seçenekleri."
  >
    <template #actions>
      <UiButton variant="secondary" :disabled="loading" @click="loadSubscriptionData">
        <UiIcon name="refresh" :size="15" /> Yenile
      </UiButton>
    </template>
  </DashboardPageHeader>

  <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
    <DashboardStatTile
      label="Mevcut paket"
      :value="user?.plan?.name ?? 'Ücretsiz Paket'"
      :note="formatPriceMonthly(user?.plan?.priceMonthly ?? 0)"
    />
    <DashboardStatTile
      label="Kullanım"
      :value="quotaValue"
      :note="quotaNote"
    />
    <DashboardStatTile
      label="Analitik"
      :value="isAdmin ? 'Sınırsız' : `${user?.plan?.analyticsDays ?? 7} gün`"
    />
    <DashboardStatTile
      label="Paket bitişi"
      :value="user?.planExpiresAt ? formatDateTr(user.planExpiresAt) : 'Süresiz'"
    />
  </div>

  <section class="mt-8">
    <h2 class="mb-4 text-2xl font-bold">Paketler</h2>

    <div v-if="loading" class="grid gap-5 md:grid-cols-3">
      <div v-for="i in 3" :key="i" class="h-64 animate-pulse rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard" />
    </div>

    <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <article
        v-for="plan in plans"
        :key="plan.id"
        class="flex flex-col justify-between rounded-lg border-[1.5px] border-ink p-6 shadow-hard"
        :class="plan.id === user?.plan?.id ? 'bg-paper-2' : 'bg-card'"
      >
        <div>
          <div class="flex items-start justify-between gap-2 border-b border-line pb-4">
            <div>
              <p class="font-mono text-xs uppercase tracking-wide text-ink-3">{{ plan.slug }}</p>
              <h3 class="mt-0.5 text-xl font-bold text-ink">{{ plan.name }}</h3>
            </div>
            <UiBadge v-if="plan.id === user?.plan?.id" tone="ok">Aktif</UiBadge>
          </div>

          <p class="mt-4 font-display text-3xl font-extrabold text-ink">
            {{ formatPriceMonthly(plan.priceMonthly) }}
          </p>

          <ul class="mt-5 flex flex-col gap-2.5 text-sm">
            <li
              v-for="feat in planFeatures(plan)"
              :key="feat.label"
              class="flex items-start gap-2"
              :class="feat.allowed ? 'text-ink' : 'text-ink-3 line-through'"
            >
              <UiIcon
                :name="feat.allowed ? 'check' : 'lock'"
                :size="15"
                class="mt-0.5 shrink-0"
                :class="feat.allowed ? 'text-ok' : 'text-ink-3'"
              />
              <span>{{ feat.label }}</span>
            </li>
          </ul>
        </div>
      </article>
    </div>
  </section>
</template>
