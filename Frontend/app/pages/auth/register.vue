<script setup lang="ts">
definePageMeta({ layout: 'auth' })
useHead({ title: 'Hesap aç | makeiTiny' })

const { isLoggedIn } = useAuth()
if (isLoggedIn.value) {
  await navigateTo('/dashboard', { replace: true })
}
</script>

<template>
  <div>
    <h1 class="display text-[2.5rem]">Hesap aç</h1>
    <p class="mb-8 mt-2 text-ink-2">Ücretsiz, kart istemez.</p>
    <AuthRegisterForm />
    <p class="mt-8 text-ink-2">
      Hesabın var mı?
      <NuxtLink to="/auth/login" class="font-semibold text-brand-ink underline underline-offset-4">Giriş yap</NuxtLink>
    </p>
  </div>
</template>
