<script setup lang="ts">
definePageMeta({ layout: 'auth' })
useHead({ title: 'Çıkış yapılıyor | makeiTiny' })

const { logout } = useAuth()
const { resetLinks } = useLinks()

onMounted(async () => {
  await logout()
  resetLinks()
  await navigateTo('/auth/login', { replace: true })
})
</script>

<template>
  <p class="flex items-center gap-3 text-lg" role="status"><UiSpinner :size="20" /> Çıkış yapılıyor</p>
</template>
