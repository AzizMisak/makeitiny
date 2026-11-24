<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Yeni link — makeiTiny' })
</script>

<template>
  <DashboardPageHeader
    title="Yeni link"
    description="Kısaltmak istediğin adresi ve isteğe bağlı link ayarlarını gir."
  />
  <DashboardCreateLinkForm />
</template>
