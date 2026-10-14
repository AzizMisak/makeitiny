<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Linklerim — makeiTiny' })

const { links, removeLink } = useLinks()
</script>

<template>
  <DashboardPageHeader title="Linklerim" description="Oluşturduğun tüm kısa linkler.">
    <template #actions>
      <UiButton to="/dashboard/new"><UiIcon name="plus" :size="16" /> Yeni link</UiButton>
    </template>
  </DashboardPageHeader>

  <DashboardLinkList :links="links" @remove="removeLink" />
</template>
