<script setup lang="ts">
definePageMeta({ layout: 'auth' })
useHead({ title: 'Giriş yap | makeiTiny' })

const { isLoggedIn } = useAuth()
if (isLoggedIn.value) {
  await navigateTo('/dashboard', { replace: true })
}
</script>

<template>
  <div>
    <h1 class="display mb-8 text-[2.5rem]">Giriş yap</h1>
    <AuthLoginForm />
    <p class="mt-8 text-ink-2">
      Hesabın yok mu?
      <NuxtLink to="/auth/register" class="font-semibold text-brand-ink underline underline-offset-4">Hesap aç</NuxtLink>
    </p>
  </div>
</template>
