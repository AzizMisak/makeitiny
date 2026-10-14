<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Genel bakış — makeiTiny' })

const { links, totalClicks, removeLink } = useLinks()
const top = computed(() => [...links.value].sort((a, b) => b.clicks - a.clicks).slice(0, 3))
</script>

<template>
  <DashboardPageHeader title="Genel bakış" description="Linklerinin son durumu.">
    <template #actions>
      <UiButton to="/dashboard/new"><UiIcon name="plus" :size="16" /> Yeni link</UiButton>
    </template>
  </DashboardPageHeader>

  <div class="grid gap-5 sm:grid-cols-3">
    <DashboardStatTile label="Toplam link" :value="links.length" />
    <DashboardStatTile label="Toplam tıklama" :value="totalClicks.toLocaleString('tr-TR')" />
    <DashboardStatTile label="Şifreli link" :value="links.filter((l) => l.locked).length" />
  </div>

  <h2 class="mb-4 mt-12 text-2xl font-bold">En çok tıklananlar</h2>
  <DashboardLinkList :links="top" @remove="removeLink" />
</template>
