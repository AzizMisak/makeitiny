<script setup lang="ts">
const { user, resetPassword, deleteAccount } = useAuth()
const { resetLinks } = useLinks()

const newPassword = ref('')
const confirmPassword = ref('')
const saved = ref(false)
const errorMsg = ref('')
const loading = ref(false)
const confirmDelete = ref(false)

async function submitPassword() {
  errorMsg.value = ''
  saved.value = false

  if (!newPassword.value || !confirmPassword.value) {
    errorMsg.value = 'Lütfen yeni parolanızı her iki alana da girin.'
    return
  }

  if (newPassword.value.length < 4) {
    errorMsg.value = 'Parola en az 4 karakter olmalıdır.'
    return
  }

  if (newPassword.value !== confirmPassword.value) {
    errorMsg.value = 'Girdiğiniz parolalar birbiriyle eşleşmiyor.'
    return
  }

  if (!user.value?.email) {
    errorMsg.value = 'Oturum bilgisi bulunamadı.'
    return
  }

  loading.value = true
  try {
    await resetPassword({ email: user.value.email, newPassword: newPassword.value })
    newPassword.value = ''
    confirmPassword.value = ''
    saved.value = true
    setTimeout(() => (saved.value = false), 2500)
  } catch (e: any) {
    errorMsg.value = e?.message ?? 'Parola güncellenemedi.'
  } finally {
    loading.value = false
  }
}

async function handleDeleteAccount() {
  errorMsg.value = ''
  try {
    await deleteAccount()
    resetLinks()
    await navigateTo('/auth/register')
  } catch (e: any) {
    errorMsg.value = e?.message ?? 'Hesap silinemedi.'
  }
}
</script>

<template>
  <div class="flex max-w-xl flex-col gap-8">
    <form class="flex flex-col gap-5 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8" @submit.prevent="submitPassword">
      <div>
        <h2 class="text-lg font-bold">Hesap Bilgileri &amp; Şifre Değiştir</h2>
        <p class="mt-1 text-sm text-ink-3">
          Kullanıcı: <strong class="text-ink">{{ user?.name }}</strong> ({{ user?.email }})
        </p>
      </div>

      <p v-if="errorMsg" class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
        {{ errorMsg }}
      </p>

      <UiField
        id="set-password"
        v-model="newPassword"
        label="Yeni Parola"
        type="password"
        placeholder="En az 4 karakter"
        hint="Parolanız en az 4 karakter uzunluğunda olmalıdır."
        autocomplete="new-password"
      />

      <UiField
        id="set-password-confirm"
        v-model="confirmPassword"
        label="Yeni Parola (Tekrar)"
        type="password"
        placeholder="Yeni parolanızı tekrar girin"
        autocomplete="new-password"
      />

      <div class="flex items-center gap-3">
        <UiButton type="submit">{{ loading ? 'Güncelleniyor...' : 'Şifreyi Güncelle' }}</UiButton>
        <UiBadge v-if="saved" tone="ok"><UiIcon name="check" :size="12" /> güncellendi</UiBadge>
      </div>
    </form>

    <div class="flex flex-col gap-4 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8">
      <div>
        <h2 class="text-lg font-bold text-accent-dark">Tehlikeli Bölge</h2>
        <p class="mt-1 text-sm text-ink-3">Hesabınızı sildiğinizde oluşturduğunuz tüm kısa linkler kalıcı olarak silinir.</p>
      </div>

      <div v-if="!confirmDelete">
        <UiButton variant="danger" @click="confirmDelete = true">
          <UiIcon name="trash" :size="16" /> Hesabımı Sil
        </UiButton>
      </div>
      <div v-else class="flex flex-wrap items-center gap-3">
        <UiButton variant="primary" @click="handleDeleteAccount">Evet, Kalıcı Olarak Sil</UiButton>
        <UiButton variant="ghost" @click="confirmDelete = false">Vazgeç</UiButton>
      </div>
    </div>
  </div>
</template>
