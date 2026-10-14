<script setup lang="ts">
definePageMeta({ layout: 'marketing' })
useHead({ title: 'Kayıt ol — makeiTiny' })
</script>

<template>
  <AuthCard title="Hesap aç." subtitle="Ücretsiz. Kredi kartı istemiyoruz.">
    <AuthRegisterForm />
    <template #footer>
      Zaten hesabın var mı?
      <NuxtLink to="/auth/login" class="font-semibold text-accent-dark underline underline-offset-2">Giriş yap</NuxtLink>
    </template>
  </AuthCard>
</template>
