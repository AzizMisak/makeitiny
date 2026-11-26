<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Genel bakış | makeiTiny' })

const { user, isAdmin, remainingLinks, fetchCurrentUser } = useAuth()
const { links, topLinks, totalLinks, totalClicks, lockedCount, loading, initialized, fetchLinks } = useLinks()
const { display } = useShortUrl()

const error = ref('')
const ready = computed(() => initialized.value && !error.value)

const maxLinks = computed(() => (isAdmin.value ? -1 : (user.value?.plan?.maxLinks ?? 50)))
const topMax = computed(() => Math.max(1, ...topLinks.value.map((l) => l.clicks)))

const quotaText = computed(() => {
  if (isAdmin.value) return 'Pakete tabi değilsin: link sınırı yok, bütün özellikler açık.'
  if (maxLinks.value < 0) return 'Paketinde link sınırı yok.'
  const left = remainingLinks.value ?? 0
  return left > 0 ? `${formatNumber(left)} link hakkın kaldı.` : 'Link hakkın doldu.'
})

async function load() {
  error.value = ''
  try {
    await Promise.all([fetchLinks(1, 10), fetchCurrentUser()])
  } catch (e: any) {
    error.value = e?.message ?? 'Veriler yüklenemedi.'
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PanelHeader title="Genel bakış" />
    <UiLoadError v-if="error" :message="error" :loading="loading" class="mb-8" @retry="load" />

    <!-- Tek büyük sayı: kullanıcının buraya bakma nedeni. -->
    <section class="grid grid-cols-1 gap-10 md:grid-cols-[minmax(0,1.25fr)_minmax(0,1fr)] md:items-end">
      <div>
        <p class="text-lg text-ink-2">{{ user?.name ? user.name + ', linklerin' : 'Linklerin' }} toplam</p>
        <div class="mt-2 flex min-h-[4.5rem] items-end md:min-h-[8.5rem]">
          <div v-if="!ready" class="skeleton h-20 w-64" />
          <p v-else class="display tnum text-[clamp(4.5rem,13vw,8.5rem)] leading-[0.95]" style="font-stretch: 120%">
            {{ formatNumber(totalClicks) }}
          </p>
        </div>
        <p class="mt-2 text-lg text-ink-2">kez tıklandı.</p>
      </div>

      <div>
        <div class="flex items-baseline justify-between gap-4">
          <h2 class="font-semibold">{{ isAdmin ? `${roleMeta(user?.roles).label} hesabı` : (user?.plan?.name ?? 'Paketim') }}</h2>
          <NuxtLink to="/dashboard/subscription" class="text-sm font-medium text-brand-ink underline-offset-4 hover:underline">
            {{ isAdmin ? 'Paketler' : 'Paket ayrıntısı' }}
          </NuxtLink>
        </div>
        <p class="tnum mt-1 text-ink-2">
          <template v-if="ready">{{ formatNumber(totalLinks) }} link oluşturdun, {{ formatNumber(lockedCount) }} tanesi parolalı.</template>
          <span v-else class="skeleton inline-block h-4 w-52 align-middle" />
        </p>
        <UiTape class="mt-4" :used="totalLinks" :max="maxLinks" label="Link kullanımı" />
        <p class="mt-2 text-sm text-ink-3">{{ quotaText }}</p>
      </div>
    </section>

    <section class="mt-16 grid grid-cols-1 gap-12 md:grid-cols-[minmax(0,1.25fr)_minmax(0,1fr)]">
      <div>
        <h2 class="display text-2xl">En çok tıklananlar</h2>
        <div v-if="!ready" class="mt-6 flex flex-col gap-5" aria-busy="true">
          <div v-for="w in ['w-full', 'w-3/4', 'w-1/2']" :key="w" class="skeleton h-12" :class="w" />
        </div>
        <ol v-else-if="topLinks.length" class="mt-6 flex flex-col gap-5">
          <li v-for="l in topLinks" :key="l.id">
            <NuxtLink :to="`/dashboard/links/${l.id}`" class="group block">
              <span class="flex items-baseline justify-between gap-4">
                <span class="min-w-0 truncate font-mono text-[15px] font-semibold group-hover:text-brand">{{ l.alias }}</span>
                <span class="tnum text-lg font-semibold">{{ formatNumber(l.clicks) }}</span>
              </span>
              <span class="mt-1.5 block h-3 bg-sunken">
                <span class="block h-full bg-brand transition-colors group-hover:bg-brand-hover" :style="{ width: `${(l.clicks / topMax) * 100}%` }" />
              </span>
              <span class="mt-1.5 block truncate text-sm text-ink-3">{{ l.title || l.target }}</span>
            </NuxtLink>
          </li>
        </ol>
        <UiEmpty v-else title="Sıralanacak link yok" text="Linklerin tıklandıkça en çok ilgi görenler burada öne çıkar." class="!items-start !px-0 !py-8 !text-left" />
      </div>

      <div>
        <div class="flex items-baseline justify-between gap-4">
          <h2 class="display text-2xl">Son eklenenler</h2>
          <NuxtLink v-if="totalLinks > 5" to="/dashboard/links" class="text-sm font-medium text-brand-ink underline-offset-4 hover:underline">
            Hepsini gör
          </NuxtLink>
        </div>
        <div v-if="!ready" class="mt-6 flex flex-col gap-4" aria-busy="true">
          <div v-for="i in 5" :key="i" class="skeleton h-9" />
        </div>
        <ul v-else-if="links.length" class="mt-4 divide-y divide-line">
          <li v-for="l in links.slice(0, 5)" :key="l.id">
            <NuxtLink :to="`/dashboard/links/${l.id}`" class="group flex items-center justify-between gap-4 py-3" :title="display(l.alias)">
              <span class="min-w-0">
                <span class="block truncate font-mono text-[15px] font-semibold group-hover:text-brand">{{ l.alias }}</span>
                <span class="block text-sm text-ink-3">{{ formatDateTr(l.createdAt) }}</span>
              </span>
              <UiStatus :tone="linkState(l).tone" :label="linkState(l).label" />
            </NuxtLink>
          </li>
        </ul>
        <div v-else class="mt-6">
          <p class="text-ink-2">İlk linkini oluşturduğunda burada görünür.</p>
          <UiButton to="/dashboard/new" icon="scissors" class="mt-4">İlk linki oluştur</UiButton>
        </div>
      </div>
    </section>
  </div>
</template>
