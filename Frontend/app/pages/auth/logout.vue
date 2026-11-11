<script setup lang="ts">
definePageMeta({ layout: 'marketing' })

const { logout } = useAuth()
const { resetLinks } = useLinks()

await logout()
resetLinks()
await navigateTo('/auth/login', { replace: true })
</script>

<template>
  <div />
</template>
