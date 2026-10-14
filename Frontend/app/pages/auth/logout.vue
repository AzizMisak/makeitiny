<script setup lang="ts">
definePageMeta({ layout: 'marketing' })

// Sahte çıkış: API bağlanınca burada POST /auth/logout çağrılacak.
const { logout } = useAuth()
logout()
await navigateTo('/auth/login', { replace: true })
</script>

<template>
  <div />
</template>
