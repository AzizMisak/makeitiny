<script setup lang="ts">
const { user, roles, isOwner, primaryRole, primaryRoleLabel, fetchCurrentUser, resetPassword, deleteAccount } = useAuth()
const { resetLinks } = useLinks()
const toast = useToast()

const newPassword = ref('')
const confirmPassword = ref('')
const saving = ref(false)

const confirmDeleteOpen = ref(false)
const deleting = ref(false)

const roleBadge = computed(() => roleBadgeMeta(roles.value))

async function handlePasswordSave() {
  if (newPassword.value.length < 6) {
    toast.error('Şifre en az 6 karakter olmalı.')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    toast.error('Şifreler eşleşmiyor.')
    return
  }
  if (!user.value?.email) return

  saving.value = true
  try {
    await resetPassword({ email: user.value.email, newPassword: newPassword.value })
    await fetchCurrentUser().catch(() => {})
    newPassword.value = ''
    confirmPassword.value = ''
    toast.success('Şifren güncellendi.')
  } catch (err: any) {
    toast.error(err?.message ?? 'Şifre güncellenemedi.')
  } finally {
    saving.value = false
  }
}

async function handleDeleteAccount() {
  if (isOwner.value) return
  deleting.value = true
  try {
    await deleteAccount()
    resetLinks()
    confirmDeleteOpen.value = false
    await navigateTo('/')
  } catch (err: any) {
    toast.error(err?.message ?? 'Hesap silinemedi.')
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div class="mt-8 grid gap-6 lg:grid-cols-2">
    <div class="flex flex-col gap-6">
      <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
        <div class="flex flex-wrap items-start justify-between gap-3 border-b border-line pb-4">
          <h2 class="text-xl font-bold">Profil bilgileri</h2>
          <UiBadge :tone="roleBadge.tone">{{ roleBadge.label }}</UiBadge>
        </div>

        <dl class="mt-4 grid gap-3 sm:grid-cols-2">
          <div class="rounded-md border border-line bg-paper p-3">
            <dt class="text-xs text-ink-3">Ad soyad</dt>
            <dd class="mt-0.5 font-semibold text-ink">{{ user?.name || '—' }}</dd>
          </div>
          <div class="rounded-md border border-line bg-paper p-3">
            <dt class="text-xs text-ink-3">E-posta</dt>
            <dd class="mt-0.5 truncate font-mono text-sm font-semibold text-ink">{{ user?.email || '—' }}</dd>
          </div>
          <div class="rounded-md border border-line bg-paper p-3">
            <dt class="text-xs text-ink-3">Yetki seviyesi</dt>
            <dd class="mt-0.5 text-sm font-semibold text-ink">{{ primaryRoleLabel }}</dd>
          </div>
          <div class="rounded-md border border-line bg-paper p-3">
            <dt class="text-xs text-ink-3">Roller</dt>
            <dd class="mt-1 flex flex-wrap gap-1.5">
              <UiBadge v-for="r in roles" :key="r" :tone="r === primaryRole ? roleBadge.tone : 'neutral'">
                {{ r }}
              </UiBadge>
            </dd>
          </div>
        </dl>
      </div>

      <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
        <div class="flex items-center justify-between gap-2">
          <h2 class="text-xl font-bold text-accent-dark">Hesabı sil</h2>
          <UiBadge v-if="isOwner" tone="neutral">Korumalı</UiBadge>
        </div>

        <p v-if="isOwner" class="mt-2 text-sm text-ink-2">
          Sistemde tek bir Owner hesabı bulunduğu için bu hesap silinemez.
        </p>
        <template v-else>
          <p class="mt-2 text-sm text-ink-2">
            Hesabını sildiğinde tüm linklerin ve tıklama logların kalıcı olarak kaldırılır.
          </p>
          <div class="mt-4">
            <UiButton variant="danger" @click="confirmDeleteOpen = true">
              <UiIcon name="trash" :size="16" /> Hesabımı sil...
            </UiButton>
          </div>
        </template>
      </div>
    </div>

    <form
      class="flex flex-col justify-between rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard"
      @submit.prevent="handlePasswordSave"
    >
      <div>
        <h2 class="text-xl font-bold">Şifre değiştir</h2>
        <p class="mt-1 text-sm text-ink-3">Hesabın için yeni bir şifre belirle.</p>

        <div class="mt-5 flex flex-col gap-4">
          <UiField
            id="set-new-pw"
            v-model="newPassword"
            label="Yeni şifre"
            type="password"
            placeholder="En az 6 karakter"
            autocomplete="new-password"
          />
          <UiField
            id="set-confirm-pw"
            v-model="confirmPassword"
            label="Yeni şifre tekrar"
            type="password"
            placeholder="••••••••"
            autocomplete="new-password"
          />
        </div>
      </div>

      <div class="mt-6 border-t border-line pt-4">
        <UiButton type="submit" :disabled="saving">
          {{ saving ? 'Kaydediliyor...' : 'Şifreyi güncelle' }}
        </UiButton>
      </div>
    </form>

    <UiConfirmModal
      :open="confirmDeleteOpen"
      title="Hesabı kalıcı olarak sil"
      :target-label="user?.email"
      description="Hesabın, oluşturduğun tüm kısa linkler ve tıklama istatistiklerin kalıcı olarak silinecek. Bu işlem geri alınamaz."
      confirm-label="Evet, hesabımı sil"
      variant="danger"
      :loading="deleting"
      @cancel="confirmDeleteOpen = false"
      @confirm="handleDeleteAccount"
    />
  </div>
</template>
