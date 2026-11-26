<script setup lang="ts">
const { login } = useAuth()
const { resetLinks } = useLinks()

const email = ref('')
const password = ref('')
const errors = ref<{ email?: string; password?: string }>({})
const formError = ref('')
const loading = ref(false)

watch([email, password], () => {
  errors.value = {}
  formError.value = ''
})

async function submit() {
  errors.value = {
    email: !email.value.trim() ? 'E-postanı yaz.' : /^\S+@\S+\.\S+$/.test(email.value.trim()) ? undefined : 'Geçerli bir e-posta yaz.',
    password: password.value ? undefined : 'Parolanı yaz.'
  }
  if (errors.value.email || errors.value.password) return

  loading.value = true
  try {
    resetLinks()
    await login({ email: email.value, password: password.value })
    await navigateTo('/dashboard')
  } catch (e: any) {
    formError.value = e?.message ?? 'Giriş yapılamadı. Tekrar dene.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <form class="flex flex-col gap-1" novalidate @submit.prevent="submit">
    <UiField id="login-email" v-model="email" label="E-posta" type="email" autocomplete="email" placeholder="sen@ornek.com" :error="errors.email" :disabled="loading" />
    <UiField id="login-password" v-model="password" label="Parola" type="password" autocomplete="current-password" :error="errors.password || formError" :disabled="loading" />
    <UiButton type="submit" block size="lg" class="mt-1" :loading="loading" loading-label="Giriş yapılıyor">Giriş yap</UiButton>
  </form>
</template>
