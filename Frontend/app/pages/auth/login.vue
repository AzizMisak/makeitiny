<script setup lang="ts">
definePageMeta({ layout: 'marketing' })
useHead({ title: 'Giriş yap — makeiTiny' })
</script>

<template>
  <AuthCard title="Tekrar hoş geldin." subtitle="Linklerine ulaşmak için giriş yap.">
    <AuthLoginForm />
    <template #footer>
      Hesabın yok mu?
      <NuxtLink to="/auth/register" class="font-semibold text-accent-dark underline underline-offset-2">Kayıt ol</NuxtLink>
    </template>
  </AuthCard>
</template>
