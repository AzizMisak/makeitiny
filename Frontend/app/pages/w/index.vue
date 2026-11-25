<script setup lang="ts">
definePageMeta({ layout: 'visitor' })
useHead({ title: 'Eksik adres | makeiTiny' })

const route = useRoute()
const alias = computed(() => String(route.query.alias ?? route.query.link ?? ''))

onMounted(async () => {
  if (alias.value) await navigateTo(`/w/${alias.value}`, { replace: true })
})
</script>

<template>
  <VisitorNotice title="Adres eksik" text="Kısa linkin kodu adreste yok. Sana gönderilen linkin tamamını aç." />
</template>
