<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Genel bakış — makeiTiny' })

const { user, isAdmin, remainingLinks, fetchCurrentUser } = useAuth()
const { topLinks, totalLinks, totalClicks, lockedCount, loading, initialized, fetchLinks, removeLink } = useLinks()
const isLoading = computed(() => loading.value || !initialized.value)

const quotaValue = computed(() => {
  if (isAdmin.value || (user.value?.plan?.maxLinks ?? 50) < 0) return 'Sınırsız'
  return `${remainingLinks.value ?? 0} kalan`
})

const quotaNote = computed(() => {
  const pkg = user.value?.plan?.name ?? 'Ücretsiz Paket'
  if (isAdmin.value) return `${pkg} (Kota muaf)`
  return pkg
})

onMounted(() => {
  fetchLinks(1, 10)
  fetchCurrentUser().catch(() => {})
})
</script>

<template>
  <DashboardPageHeader title="Genel bakış" description="Linklerinin son durumu ve hesap kotan.">
    <template #actions>
      <UiButton to="/dashboard/new"><UiIcon name="plus" :size="16" /> Yeni link</UiButton>
    </template>
  </DashboardPageHeader>

  <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
    <DashboardStatTile label="Toplam link" :value="isLoading ? '...' : totalLinks.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Toplam tıklama" :value="isLoading ? '...' : totalClicks.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Şifreli link" :value="isLoading ? '...' : lockedCount.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Kalan Kota / Paket" :value="quotaValue" :note="quotaNote" />
  </div>

  <h2 class="mb-4 mt-10 text-2xl font-bold">En çok tıklananlar</h2>
  <DashboardLinkList :links="topLinks" :loading="isLoading" @remove="removeLink" />
</template>
