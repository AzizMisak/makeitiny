<script setup lang="ts">
const { login } = useAuth()

const email = ref('')
const password = ref('')

async function submit() {
  // Sahte giriş: API bağlanınca burada POST /auth/login çağrılacak.
  login(email.value || 'demo@makeitiny.com')
  await navigateTo('/dashboard')
}
</script>

<template>
  <form class="flex flex-col gap-5" @submit.prevent="submit">
    <UiField id="login-email" v-model="email" label="E-posta" type="email" placeholder="sen@ornek.com" autocomplete="email" />
    <UiField id="login-password" v-model="password" label="Parola" type="password" placeholder="••••••••" autocomplete="current-password" />
    <UiButton type="submit" block>Giriş yap</UiButton>
  </form>
</template>
