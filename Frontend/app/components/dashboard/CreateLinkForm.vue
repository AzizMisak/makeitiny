<script setup lang="ts">
const { addLink } = useLinks()

const target = ref('')
const alias = ref('')
const password = ref('')

async function submit() {
  if (!target.value) return
  addLink({ alias: alias.value, target: target.value, locked: !!password.value })
  await navigateTo('/dashboard/links')
}
</script>

<template>
  <form class="flex max-w-xl flex-col gap-5 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8" @submit.prevent="submit">
    <UiField id="new-target" v-model="target" label="Hedef adres" type="url" placeholder="https://uzun-bir-adres.com/sayfa" />
    <UiField id="new-alias" v-model="alias" label="Takma ad (isteğe bağlı)" prefix="mkit.ee/" placeholder="kampanya" hint="Boş bırakırsan rastgele bir kod üretilir." />
    <UiField id="new-password" v-model="password" label="Parola (isteğe bağlı)" type="password" hint="Doluysa link şifreli olur." />
    <div class="flex gap-3">
      <UiButton type="submit">Linki oluştur</UiButton>
      <UiButton to="/dashboard/links" variant="ghost">Vazgeç</UiButton>
    </div>
  </form>
</template>
