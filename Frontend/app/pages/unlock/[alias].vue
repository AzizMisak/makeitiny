<script setup lang="ts">
// NOT (TODO): Şifreli linklerde şifre çözülmese bile ilk yönlendirme (/r/:alias) anında
// tıklama sayısı artırılmakta ve cihaz logu tutulmaktadır. İleride şifre doğrulama
// durumuna göre tıklama/log davranışı detaylı olarak ele alınacaktır.
definePageMeta({ layout: 'marketing' })

const route = useRoute()
const { public: config } = useRuntimeConfig()
const alias = computed(() => String(route.params.alias ?? ''))

useHead({ title: () => `Şifreli Link (${alias.value}) — makeiTiny` })

const password = ref('')
const errorMsg = ref('')
const loading = ref(false)

async function submit() {
  if (!password.value) return
  errorMsg.value = ''
  loading.value = true
  try {
    const res = await $fetch<{ target: string }>(`${config.routerBase}/${alias.value}/unlock`, {
      method: 'POST',
      body: { password: password.value }
    })
    if (res?.target) {
      window.location.href = res.target
    }
  } catch (e: any) {
    const code = e?.data?.code
    if (code === 'link_not_found') {
      await navigateTo(`/error/${alias.value}?reason=not_found`, { replace: true })
      return
    }
    if (code === 'target_unreachable') {
      await navigateTo(`/error/${alias.value}?reason=unreachable`, { replace: true })
      return
    }
    errorMsg.value = e?.data?.message ?? 'Parola doğrulanamadı.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthCard
    title="Bu link şifreli."
    :subtitle="`/${alias} kısa bağlantısına devam etmek için parolayı girin.`"
  >
    <form class="flex flex-col gap-5" @submit.prevent="submit">
      <p v-if="errorMsg" class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
        {{ errorMsg }}
      </p>
      <UiField
        id="unlock-password"
        v-model="password"
        label="Link Parolası"
        type="password"
        placeholder="••••••••"
        autocomplete="off"
      />
      <UiButton type="submit" block>
        <UiIcon name="lock" :size="16" /> {{ loading ? 'Doğrulanıyor...' : 'Kilidi Aç ve Git' }}
      </UiButton>
    </form>
    <template #footer>
      <NuxtLink to="/" class="font-semibold text-accent-dark underline underline-offset-2">Ana sayfaya dön</NuxtLink>
    </template>
  </AuthCard>
</template>
