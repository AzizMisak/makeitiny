<script setup lang="ts">
definePageMeta({ layout: 'auth' })
useHead({ title: 'Giriş yap | makeiTiny' })

const { isLoggedIn } = useAuth()
if (isLoggedIn.value) {
  await navigateTo('/dashboard', { replace: true })
}

// Oturum askı nedeniyle kapatıldıysa sebep bir kez gösterilir.
const notice = useAuthNotice()
const message = ref(notice.value)
onMounted(() => {
  notice.value = ''
})
</script>

<template>
  <div>
    <h1 class="display mb-8 text-[2.5rem]">Giriş yap</h1>
    <p v-if="message" role="alert" class="mb-6 border-l-4 border-danger bg-danger-soft px-4 py-3 text-sm text-danger">
      Oturumun kapatıldı: {{ message }}
    </p>
    <AuthLoginForm />
    <p class="mt-8 text-ink-2">
      Hesabın yok mu?
      <NuxtLink to="/auth/register" class="font-semibold text-brand-ink underline underline-offset-4">Hesap aç</NuxtLink>
    </p>
  </div>
</template>
