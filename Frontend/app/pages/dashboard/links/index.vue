<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Linklerim — makeiTiny' })

const { user, isAdmin, remainingLinks } = useAuth()
const {
  links,
  currentPage,
  pageSize,
  totalLinks,
  totalClicks,
  lockedCount,
  totalPages,
  loading,
  initialized,
  fetchLinks,
  removeLink
} = useLinks()

const isLoading = computed(() => loading.value || !initialized.value)

const quotaValue = computed(() => {
  if (isAdmin.value || (user.value?.plan?.maxLinks ?? 50) < 0) return 'Sınırsız'
  return `${remainingLinks.value ?? 0} kalan`
})

const rangeStart = computed(() => (totalLinks.value === 0 ? 0 : (currentPage.value - 1) * pageSize.value + 1))
const rangeEnd = computed(() => Math.min(currentPage.value * pageSize.value, totalLinks.value))

const visiblePages = computed<(number | '...')[]>(() => {
  const total = totalPages.value
  const current = currentPage.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  const pages: (number | '...')[] = [1]
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let p = start; p <= end; p++) {
    pages.push(p)
  }
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

async function goToPage(page: number) {
  if (page < 1 || page > totalPages.value || page === currentPage.value) return
  await fetchLinks(page, pageSize.value)
}

onMounted(() => {
  fetchLinks(currentPage.value, pageSize.value)
})
</script>

<template>
  <DashboardPageHeader title="Linklerim" description="Oluşturduğun tüm kısa linkler. Detay ve düzenleme için bir linke tıkla.">
    <template #actions>
      <UiButton to="/dashboard/new"><UiIcon name="plus" :size="16" /> Yeni link</UiButton>
    </template>
  </DashboardPageHeader>

  <div class="mb-8 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
    <DashboardStatTile label="Toplam link" :value="isLoading ? '...' : totalLinks.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Toplam tıklama" :value="isLoading ? '...' : totalClicks.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Şifreli link" :value="isLoading ? '...' : lockedCount.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Kalan Kota" :value="quotaValue" :note="user?.plan?.name ?? 'Ücretsiz Paket'" />
  </div>

  <DashboardLinkList :links="links" :loading="isLoading" @remove="removeLink" />

  <!-- Backend Pagination Kontrolleri -->
  <div
    v-if="totalLinks > 0"
    class="mt-6 flex flex-wrap items-center justify-between gap-4 rounded-lg border-[1.5px] border-ink bg-card px-5 py-3.5 shadow-hard"
  >
    <p class="text-xs text-ink-2">
      Toplam <span class="font-mono font-bold text-ink">{{ totalLinks }}</span> linkten
      <span class="font-mono font-bold text-ink">{{ rangeStart }}–{{ rangeEnd }}</span> arası gösteriliyor
      <span class="text-ink-3">(Sayfa {{ currentPage }} / {{ totalPages }})</span>
    </p>

    <div class="flex flex-wrap items-center gap-1.5">
      <button
        type="button"
        :disabled="currentPage <= 1 || isLoading"
        class="inline-flex items-center gap-1 rounded-sm border border-ink bg-card px-2.5 py-1 text-xs font-semibold text-ink transition-colors hover:bg-paper-2 disabled:cursor-not-allowed disabled:opacity-40 cursor-pointer"
        @click="goToPage(currentPage - 1)"
      >
        ← Önceki
      </button>

      <template v-for="(pageItem, idx) in visiblePages" :key="`${pageItem}-${idx}`">
        <span
          v-if="pageItem === '...'"
          class="px-1.5 font-mono text-xs text-ink-3 select-none"
        >
          ...
        </span>
        <button
          v-else
          type="button"
          :disabled="isLoading"
          class="inline-flex h-7 min-w-7 items-center justify-center rounded-sm border px-2 font-mono text-xs font-bold transition-colors cursor-pointer"
          :class="
            pageItem === currentPage
              ? 'border-ink bg-accent text-white'
              : 'border-line bg-card text-ink-2 hover:border-ink hover:text-ink'
          "
          @click="goToPage(pageItem)"
        >
          {{ pageItem }}
        </button>
      </template>

      <button
        type="button"
        :disabled="currentPage >= totalPages || isLoading"
        class="inline-flex items-center gap-1 rounded-sm border border-ink bg-card px-2.5 py-1 text-xs font-semibold text-ink transition-colors hover:bg-paper-2 disabled:cursor-not-allowed disabled:opacity-40 cursor-pointer"
        @click="goToPage(currentPage + 1)"
      >
        Sonraki →
      </button>
    </div>
  </div>
</template>
