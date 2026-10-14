<script setup lang="ts">
const { login } = useAuth()

const email = ref('')
const password = ref('')

async function submit() {
  // Sahte kayıt: API bağlanınca burada POST /auth/register çağrılacak.
  login(email.value || 'demo@makeitiny.com')
  await navigateTo('/dashboard')
}
</script>

<template>
  <form class="flex flex-col gap-5" @submit.prevent="submit">
    <UiField id="register-email" v-model="email" label="E-posta" type="email" placeholder="sen@ornek.com" autocomplete="email" />
    <UiField
      id="register-password"
      v-model="password"
      label="Parola"
      type="password"
      placeholder="En az 8 karakter"
      hint="Harf ve rakam karıştırmak iyi fikir."
      autocomplete="new-password"
    />
    <UiButton type="submit" block>Hesap oluştur</UiButton>
  </form>
</template>
