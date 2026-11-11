<script setup lang="ts">
const { addLink } = useLinks()

const target = ref('')
const alias = ref('')
const password = ref('')
const errorMsg = ref('')
const loading = ref(false)

async function submit() {
  if (!target.value.trim()) {
    errorMsg.value = 'Lütfen kısaltmak istediğiniz hedef adresi girin.'
    return
  }
  errorMsg.value = ''
  loading.value = true
  try {
    const created = await addLink({
      target: target.value,
      alias: alias.value.trim() || undefined,
      password: password.value || undefined
    })
    await navigateTo(`/dashboard/links/${created.id}`)
  } catch (e: any) {
    errorMsg.value = e?.message ?? 'Link oluşturulamadı.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <form class="flex max-w-xl flex-col gap-5 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8" @submit.prevent="submit">
    <p v-if="errorMsg" class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
      {{ errorMsg }}
    </p>
    <UiField
      id="new-target"
      v-model="target"
      label="Hedef adres"
      type="text"
      placeholder="ornek.com/sayfa veya https://ornek.com/sayfa"
      hint="Başında http:// veya https:// yoksa varsayılan olarak güvenli https:// eklenir."
    />
    <UiField
      id="new-alias"
      v-model="alias"
      label="Takma ad (isteğe bağlı)"
      prefix="localhost/r/"
      placeholder="kampanya"
      hint="Boş bırakırsan rastgele 6 karakterli Base62 kod üretilir."
    />
    <UiField
      id="new-password"
      v-model="password"
      label="Parola (isteğe bağlı)"
      type="password"
      placeholder="İsteğe bağlı parola"
      hint="Doluysa link Argon2id ile şifreli olur."
    />
    <div class="flex gap-3">
      <UiButton type="submit">{{ loading ? 'Oluşturuluyor...' : 'Linki oluştur' }}</UiButton>
      <UiButton to="/dashboard/links" variant="ghost">Vazgeç</UiButton>
    </div>
  </form>
</template>
