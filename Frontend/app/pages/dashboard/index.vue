<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Genel bakış — makeiTiny' })

const { topLinks, totalLinks, totalClicks, lockedCount, loading, initialized, fetchLinks, removeLink } = useLinks()
const isLoading = computed(() => loading.value || !initialized.value)

onMounted(() => {
  fetchLinks(1, 10)
})
</script>

<template>
  <DashboardPageHeader title="Genel bakış" description="Linklerinin son durumu.">
    <template #actions>
      <UiButton to="/dashboard/new"><UiIcon name="plus" :size="16" /> Yeni link</UiButton>
    </template>
  </DashboardPageHeader>

  <div class="grid gap-5 sm:grid-cols-3">
    <DashboardStatTile label="Toplam link" :value="isLoading ? '...' : totalLinks.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Toplam tıklama" :value="isLoading ? '...' : totalClicks.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Şifreli link" :value="isLoading ? '...' : lockedCount.toLocaleString('tr-TR')" />
  </div>

  <h2 class="mb-4 mt-12 text-2xl font-bold">En çok tıklananlar</h2>
  <DashboardLinkList :links="topLinks" :loading="isLoading" @remove="removeLink" />
</template>
