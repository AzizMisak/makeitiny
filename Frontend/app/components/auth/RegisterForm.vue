<script setup lang="ts">
const { register } = useAuth()
const { resetLinks } = useLinks()

const email = ref('')
const password = ref('')
const errorMsg = ref('')
const loading = ref(false)

async function submit() {
  errorMsg.value = ''
  if (!email.value.trim().includes('@')) {
    errorMsg.value = 'Geçerli bir e-posta adresi giriniz.'
    return
  }
  if (password.value.length < 4) {
    errorMsg.value = 'Parola en az 4 karakter olmalıdır.'
    return
  }
  loading.value = true
  try {
    resetLinks()
    await register({ email: email.value, password: password.value })
    await navigateTo('/dashboard')
  } catch (e: any) {
    errorMsg.value = e?.message ?? 'Kayıt oluşturulamadı.'
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
    <UiField id="register-email" v-model="email" label="E-posta" type="email" placeholder="sen@ornek.com" autocomplete="email" />
    <UiField
      id="register-password"
      v-model="password"
      label="Parola"
      type="password"
      placeholder="En az 4 karakter"
      hint="En az 4 karakter. Kullanıcı adın otomatik olarak e-postanın @ öncesi olarak atanır."
      autocomplete="new-password"
    />
    <UiButton type="submit" block>{{ loading ? 'Hesap oluşturuluyor...' : 'Hesap oluştur' }}</UiButton>
  </form>
</template>
