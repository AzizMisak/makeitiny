<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Yeni link — makeiTiny' })
</script>

<template>
  <DashboardPageHeader title="Yeni link" description="Uzun bir adresi kısalt." />
  <DashboardCreateLinkForm />
</template>
