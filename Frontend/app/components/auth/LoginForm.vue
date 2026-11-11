<script setup lang="ts">
const { login } = useAuth()
const { resetLinks } = useLinks()

const email = ref('')
const password = ref('')
const errorMsg = ref('')
const loading = ref(false)

async function submit() {
  errorMsg.value = ''
  if (!email.value.trim() || !password.value) {
    errorMsg.value = 'Lütfen e-posta ve parolanızı girin.'
    return
  }
  loading.value = true
  try {
    resetLinks()
    await login({ email: email.value, password: password.value })
    await navigateTo('/dashboard')
  } catch (e: any) {
    errorMsg.value = e?.message ?? 'Giriş yapılamadı.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <form class="flex flex-col gap-5" @submit.prevent="submit">
    <p v-if="errorMsg" class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
      {{ errorMsg }}
    </p>
    <UiField id="login-email" v-model="email" label="E-posta" type="email" placeholder="sen@ornek.com" autocomplete="email" />
    <UiField id="login-password" v-model="password" label="Parola" type="password" placeholder="••••••••" autocomplete="current-password" />
    <UiButton type="submit" block>{{ loading ? 'Giriş yapılıyor...' : 'Giriş yap' }}</UiButton>
  </form>
</template>
