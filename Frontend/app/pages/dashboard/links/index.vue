<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Linklerim — makeiTiny' })

const { links, loading, initialized, fetchLinks, removeLink } = useLinks()
const isLoading = computed(() => loading.value || !initialized.value)

onMounted(() => {
  fetchLinks()
})
</script>

<template>
  <DashboardPageHeader title="Linklerim" description="Oluşturduğun tüm kısa linkler. Detay ve düzenleme için bir linke tıkla.">
    <template #actions>
      <UiButton to="/dashboard/new"><UiIcon name="plus" :size="16" /> Yeni link</UiButton>
    </template>
  </DashboardPageHeader>

  <DashboardLinkList :links="links" :loading="isLoading" @remove="removeLink" />
</template>
