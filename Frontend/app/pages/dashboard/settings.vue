<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Ayarlar — makeiTiny' })
</script>

<template>
  <DashboardPageHeader title="Ayarlar" description="Hesap bilgilerini güncelle." />
  <DashboardSettingsForm />
</template>
