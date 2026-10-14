<script setup lang="ts">
const { user } = useAuth()
const name = ref(user.value?.name ?? '')
const email = ref(user.value?.email ?? '')
const saved = ref(false)

function submit() {
  if (user.value) user.value = { ...user.value, name: name.value, email: email.value }
  saved.value = true
  setTimeout(() => (saved.value = false), 1800)
}
</script>

<template>
  <form class="flex max-w-xl flex-col gap-5 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8" @submit.prevent="submit">
    <UiField id="set-name" v-model="name" label="Ad" />
    <UiField id="set-email" v-model="email" label="E-posta" type="email" />
    <div class="flex items-center gap-3">
      <UiButton type="submit">Kaydet</UiButton>
      <UiBadge v-if="saved" tone="ok"><UiIcon name="check" :size="12" /> kaydedildi</UiBadge>
    </div>
  </form>
</template>
