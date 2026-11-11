<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Genel bakış — makeiTiny' })

const { links, loading, initialized, totalClicks, fetchLinks, removeLink } = useLinks()
const isLoading = computed(() => loading.value || !initialized.value)
const top = computed(() => [...links.value].sort((a, b) => b.clicks - a.clicks).slice(0, 3))

onMounted(() => {
  fetchLinks()
})
</script>

<template>
  <DashboardPageHeader title="Genel bakış" description="Linklerinin son durumu.">
    <template #actions>
      <UiButton to="/dashboard/new"><UiIcon name="plus" :size="16" /> Yeni link</UiButton>
    </template>
  </DashboardPageHeader>

  <div class="grid gap-5 sm:grid-cols-3">
    <DashboardStatTile label="Toplam link" :value="isLoading ? '...' : links.length" />
    <DashboardStatTile label="Toplam tıklama" :value="isLoading ? '...' : totalClicks.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Şifreli link" :value="isLoading ? '...' : links.filter((l) => l.locked).length" />
  </div>

  <h2 class="mb-4 mt-12 text-2xl font-bold">En çok tıklananlar</h2>
  <DashboardLinkList :links="top" :loading="isLoading" @remove="removeLink" />
</template>
