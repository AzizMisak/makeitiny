<script setup lang="ts">
import type { NuxtError } from '#app'

// Uygulama genelindeki özel hata sayfası: var olmayan adresler (404) ve beklenmeyen hatalar.
const props = defineProps<{ error: NuxtError }>()

const { isLoggedIn } = useAuth()
const notFound = computed(() => props.error?.statusCode === 404)
const path = computed(() => {
  const raw = props.error?.url || (props.error?.data as { path?: string } | undefined)?.path
  if (raw) return raw
  return import.meta.client ? window.location.pathname : '/'
})

useHead({
  htmlAttrs: { lang: 'tr' },
  title: () => (notFound.value ? 'Sayfa bulunamadı | makeiTiny' : 'Bir hata oluştu | makeiTiny')
})

const home = computed(() => (isLoggedIn.value ? '/dashboard' : '/'))

function goHome() {
  clearError({ redirect: home.value })
}

function back() {
  if (import.meta.client && window.history.length > 1) {
    clearError()
    window.history.back()
    return
  }
  goHome()
}
</script>

<template>
  <div class="flex min-h-dvh flex-col bg-ground text-ink">
    <header class="flex items-center justify-between px-5 py-4 md:px-8">
      <BrandLogo />
      <UiThemeToggle />
    </header>
    <main class="flex flex-1 flex-col">
      <VisitorNotice
        :label="notFound ? path : String(error?.statusCode ?? 500)"
        :title="notFound ? 'Bu sayfa yok' : 'Bir şeyler ters gitti'"
        :text="
          notFound
            ? 'Adres yanlış yazılmış ya da sayfa kaldırılmış olabilir. Kısa bir link açmaya çalışıyorsan adresin /r/ ile başladığından emin ol.'
            : 'Sayfa yüklenirken beklenmeyen bir sorun oldu. Biraz sonra tekrar dene.'
        "
        :home="false"
      >
        <UiButton icon="home" @click="goHome">{{ isLoggedIn ? 'Panele dön' : 'Ana sayfa' }}</UiButton>
        <UiButton variant="outline" icon="back" @click="back">Geri dön</UiButton>
      </VisitorNotice>
    </main>
    <UiToastHost />
  </div>
</template>
