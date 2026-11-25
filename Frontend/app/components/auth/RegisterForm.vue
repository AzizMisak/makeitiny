<script setup lang="ts">
const { register } = useAuth()
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

const emailHint = computed(() => {
  const name = email.value.trim().split('@')[0]
  return name && email.value.includes('@') ? `Adın ${name} olarak kaydedilir.` : 'Adın, @ işaretinden önceki kısım olur.'
})

async function submit() {
  errors.value = {
    email: /^\S+@\S+$/.test(email.value.trim()) ? undefined : 'Geçerli bir e-posta yaz.',
    password: password.value.length >= 4 ? undefined : 'En az 4 karakter yaz.'
  }
  if (errors.value.email || errors.value.password) return

  loading.value = true
  try {
    resetLinks()
    await register({ email: email.value, password: password.value })
    await navigateTo('/dashboard')
  } catch (e: any) {
    formError.value = e?.message ?? 'Hesap açılamadı. Tekrar dene.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <form class="flex flex-col gap-1" novalidate @submit.prevent="submit">
    <UiField id="register-email" v-model="email" label="E-posta" type="email" autocomplete="email" placeholder="sen@ornek.com" :hint="emailHint" :error="errors.email" :disabled="loading" />
    <UiField id="register-password" v-model="password" label="Parola" type="password" autocomplete="new-password" hint="En az 4 karakter" :error="errors.password || formError" :disabled="loading" />
    <UiButton type="submit" block size="lg" class="mt-1" :loading="loading" loading-label="Hesap açılıyor">Hesap aç</UiButton>
  </form>
</template>
