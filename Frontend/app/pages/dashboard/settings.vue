<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Hesap ayarları | makeiTiny' })

const { user, roles, isOwner, fetchCurrentUser, resetPassword, deleteAccount, token } = useAuth()
const { resetLinks } = useLinks()
const { ask } = useConfirm()
const toast = useToast()

const loading = ref(true)

onMounted(async () => {
  await fetchCurrentUser().catch(() => {})
  loading.value = false
})

const role = computed(() => roleMeta(roles.value))

const facts = computed(() => [
  { label: 'Ad', value: user.value?.name || '—' },
  { label: 'E-posta', value: user.value?.email || '—' },
  { label: 'Rol', value: role.value.label },
  { label: 'Hesap durumu', value: user.value?.isActive === false ? 'Askıda' : 'Aktif' },
  { label: 'Kayıt tarihi', value: formatDateTr(user.value?.createdAt) }
])

const newPassword = ref('')
const repeat = ref('')
const touched = ref(false)

const lengthError = computed(() => (touched.value && newPassword.value.length < 4 ? 'En az 4 karakter yaz.' : ''))
const repeatError = computed(() => (touched.value && repeat.value !== newPassword.value ? 'İki parola aynı değil.' : ''))
const repeatOk = computed(() => (newPassword.value.length >= 4 && repeat.value === newPassword.value ? 'Parolalar aynı' : ''))

async function changePassword() {
  touched.value = true
  if (lengthError.value || repeatError.value || !user.value?.email) return
  const email = user.value.email

  const ok = await ask({
    title: 'Parola değiştirilsin mi?',
    subject: email,
    message: 'Parolan hemen değişir.',
    effects: ['Bütün cihazlardaki oturumların kapanır.', 'Bu cihazda da yeni parolanla tekrar giriş yaparsın.'],
    confirmLabel: 'Parolayı değiştir',
    action: () => resetPassword({ email, newPassword: newPassword.value })
  })
  if (!ok) return

  // Sunucu bütün oturumları iptal eder; kullanıcıyı açıkça giriş ekranına götür.
  toast.success('Parolan değişti. Yeni parolanla giriş yap.')
  token.value = null
  user.value = null
  resetLinks()
  await navigateTo('/auth/login', { replace: true })
}

async function removeAccount() {
  const email = user.value?.email
  if (isOwner.value || !email) return
  const ok = await ask({
    title: 'Hesap silinsin mi?',
    subject: email,
    message: 'Bu işlem geri alınamaz.',
    effects: [
      `${formatNumber(user.value?.linkCount)} linkin silinir, kısa adresler çalışmaz olur.`,
      'Bütün analizlerin ve ziyaret kayıtların gider.'
    ],
    confirmLabel: 'Hesabımı sil',
    tone: 'danger',
    typeToConfirm: email,
    action: async () => {
      await deleteAccount()
      resetLinks()
    }
  })
  if (ok) await navigateTo('/', { replace: true })
}
</script>

<template>
  <div class="max-w-4xl">
    <PanelHeader title="Hesap ayarları" />

    <section class="grid grid-cols-1 gap-x-10 gap-y-4 border-t border-line py-8 md:grid-cols-[14rem_minmax(0,1fr)]">
      <div>
        <h2 class="text-lg font-semibold">Profil</h2>
        <p class="mt-1 text-sm text-ink-3">Hesabının kimlik bilgileri.</p>
      </div>
      <dl class="divide-y divide-line">
        <div v-for="f in facts" :key="f.label" class="flex min-h-12 items-center justify-between gap-6 py-2">
          <dt class="text-ink-2">{{ f.label }}</dt>
          <dd class="min-w-0 truncate font-medium">
            <span v-if="loading && !user" class="skeleton inline-block h-4 w-32" />
            <template v-else>{{ f.value }}</template>
          </dd>
        </div>
      </dl>
    </section>

    <section class="grid grid-cols-1 gap-x-10 gap-y-4 border-t border-line py-8 md:grid-cols-[14rem_minmax(0,1fr)]">
      <div>
        <h2 class="text-lg font-semibold">Parola</h2>
        <p class="mt-1 text-sm text-ink-3">Değişince bütün cihazlardaki oturumların kapanır.</p>
      </div>
      <form class="flex max-w-sm flex-col gap-2" novalidate @submit.prevent="changePassword">
        <UiField
          id="set-new"
          v-model="newPassword"
          label="Yeni parola"
          type="password"
          autocomplete="new-password"
          hint="En az 4 karakter"
          :error="lengthError"
        />
        <UiField
          id="set-repeat"
          v-model="repeat"
          label="Yeni parola, tekrar"
          type="password"
          autocomplete="new-password"
          :error="repeatError"
          :ok="repeatOk"
        />
        <UiButton type="submit" variant="outline" class="self-start">Parolayı değiştir</UiButton>
      </form>
    </section>

    <section class="grid grid-cols-1 gap-x-10 gap-y-4 border-t-2 border-danger py-8 md:grid-cols-[14rem_minmax(0,1fr)]">
      <h2 class="text-lg font-semibold text-danger">Hesabı sil</h2>
      <div>
        <p v-if="isOwner" class="text-ink-2">Sistem sahibinin hesabı silinemez.</p>
        <template v-else>
          <p class="max-w-lg text-ink-2">Linklerin, analizlerin ve hesabın kalıcı olarak silinir.</p>
          <UiButton variant="danger" icon="trash" class="-ml-3 mt-3" @click="removeAccount">Hesabımı sil</UiButton>
        </template>
      </div>
    </section>
  </div>
</template>
