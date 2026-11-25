<script setup lang="ts">
definePageMeta({ layout: 'visitor' })

const route = useRoute()
const { public: config } = useRuntimeConfig()
const { display } = useShortUrl()
const alias = computed(() => String(route.params.alias ?? ''))

useHead({ title: () => `Parolalı link: ${alias.value} | makeiTiny` })

const password = ref('')
const error = ref('')
const loading = ref(false)
const leaving = ref(false)

watch(password, () => (error.value = ''))

async function submit() {
  if (!password.value) {
    error.value = 'Parolayı yaz.'
    return
  }
  error.value = ''
  loading.value = true
  try {
    const res = await $fetch<{ target: string }>(`${config.routerBase}/${alias.value}/unlock`, {
      method: 'POST',
      body: { password: password.value }
    })
    if (res?.target) {
      leaving.value = true
      window.location.href = res.target
      return
    }
    error.value = 'Hedef adres alınamadı. Tekrar dene.'
  } catch (e: any) {
    const reason = reasonFromCode(e?.data?.code)
    if (reason) {
      await navigateTo(`/error/${alias.value}?reason=${reason}`, { replace: true })
      return
    }
    error.value = e?.data?.message ?? 'Parola doğrulanamadı. Tekrar dene.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mat-grid flex flex-1 items-center justify-center px-5 py-16">
    <div class="w-full max-w-md bg-surface px-7 pb-7 pt-9 sm:px-10">
      <UiIcon name="lock" :size="36" class="text-brand" />
      <h1 class="display mt-5 text-[2.25rem]">Parola gerekiyor</h1>
      <p class="mt-2 text-ink-2"><span class="break-all font-mono text-[15px]">{{ display(alias) }}</span> linki parolayla korunuyor.</p>

      <form class="mt-7" novalidate @submit.prevent="submit">
        <UiField id="unlock-password" v-model="password" label="Link parolası" type="password" autocomplete="off" large :error="error" :disabled="loading || leaving" />
        <UiButton type="submit" block size="lg" class="mt-2" icon="unlock" :loading="loading || leaving" :loading-label="leaving ? 'Hedefe gidiliyor' : 'Denetleniyor'">
          Kilidi aç
        </UiButton>
      </form>
    </div>
  </div>
</template>
