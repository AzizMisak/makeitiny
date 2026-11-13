<script setup lang="ts">
definePageMeta({ layout: 'marketing' })

const route = useRoute()
const { public: config } = useRuntimeConfig()

const alias = computed(() => String(route.params.alias ?? ''))
const reason = computed(() => String(route.query.reason ?? 'not_found'))

const isTargetUnreachable = computed(() => reason.value === 'unreachable')
const isNotFound = computed(() => reason.value === 'not_found')

const pageTitle = computed(() => {
  if (isTargetUnreachable.value) {
    return `Hedefe Ulaşılamıyor (${alias.value}) — makeiTiny`
  }
  if (isNotFound.value) {
    return `Kısa Link Bulunamadı (${alias.value}) — makeiTiny`
  }
  return `Yönlendirme Hatası (${alias.value}) — makeiTiny`
})

useHead({ title: pageTitle })

const cardTitle = computed(() => {
  if (isTargetUnreachable.value) {
    return 'Hedef adrese ulaşılamıyor.'
  }
  if (isNotFound.value) {
    return 'Kısa link bulunamadı.'
  }
  return 'Yönlendirme tamamlanamadı.'
})

const cardSubtitle = computed(() => {
  if (isTargetUnreachable.value) {
    return `/${alias.value} kısa bağlantısı sistemimizde kayıtlı ancak yönlendirilen hedef siteye şu anda erişilemiyor.`
  }
  if (isNotFound.value) {
    return `/${alias.value} kısa bağlantısı veritabanımızda mevcut değil veya silinmiş.`
  }
  return `/${alias.value} bağlantısı yönlendirilirken beklenmeyen bir sorun oluştu.`
})

function retryRedirect() {
  if (!alias.value) return
  window.location.href = `${config.routerBase}/${encodeURIComponent(alias.value)}`
}
</script>

<template>
  <AuthCard :title="cardTitle" :subtitle="cardSubtitle">
    <div class="flex flex-col gap-5">
      <div class="flex items-center justify-between rounded-md border border-line bg-paper-2 px-3.5 py-2.5">
        <span class="font-mono text-xs text-ink-2">/{{ alias }}</span>
        <UiBadge tone="accent">
          <template v-if="isTargetUnreachable">502 • HEDEF KAYNAKLI HATA</template>
          <template v-else-if="isNotFound">404 • KAYIT BULUNAMADI</template>
          <template v-else>HATA • YÖNLENDİRME BAŞARISIZ</template>
        </UiBadge>
      </div>

      <!-- Hedef URL erişilemiyor (Link DB'de var ama hedef site yanıt vermiyor) -->
      <div
        v-if="isTargetUnreachable"
        class="flex flex-col gap-2 rounded-md border border-accent-dark bg-accent-soft p-4 text-sm text-accent-dark"
      >
        <div class="flex items-center gap-2 font-semibold">
          <UiIcon name="alert" :size="18" />
          <span>Hedef web sitesi yanıt vermiyor</span>
        </div>
        <p class="text-xs leading-relaxed text-ink-2">
          Kısa link sistemimizde doğrulandı fakat yönlendirilmek istenen hedef sunucuya bağlantı kurulamadı.
          Hedef site geçici olarak kapalı, alan adı süresi dolmuş veya sayfa kaldırılmış olabilir.
        </p>
      </div>

      <!-- Kısa link DB'de yok -->
      <div
        v-else-if="isNotFound"
        class="flex flex-col gap-2 rounded-md border border-accent-dark bg-accent-soft p-4 text-sm text-accent-dark"
      >
        <div class="flex items-center gap-2 font-semibold">
          <UiIcon name="alert" :size="18" />
          <span>Veritabanımızda böyle bir link yok</span>
        </div>
        <p class="text-xs leading-relaxed text-ink-2">
          Aradığınız kısa bağlantı kodu sistemimizde kayıtlı değil. Bağlantı adresini yanlış yazmış olabilirsiniz
          veya bu kısa link sahibi tarafından silinmiş olabilir.
        </p>
      </div>

      <!-- Genel yönlendirme hatası -->
      <div
        v-else
        class="flex flex-col gap-2 rounded-md border border-accent-dark bg-accent-soft p-4 text-sm text-accent-dark"
      >
        <div class="flex items-center gap-2 font-semibold">
          <UiIcon name="alert" :size="18" />
          <span>Beklenmeyen bir yönlendirme hatası oluştu</span>
        </div>
        <p class="text-xs leading-relaxed text-ink-2">
          Kısa bağlantı çözümlenirken geçici bir sorun meydana geldi. Lütfen biraz sonra tekrar deneyin.
        </p>
      </div>

      <div class="flex flex-col gap-2.5">
        <UiButton
          v-if="isTargetUnreachable || !isNotFound"
          type="button"
          block
          @click="retryRedirect"
        >
          <UiIcon name="refresh" :size="16" /> Tekrar Dene
        </UiButton>

        <UiButton to="/" :variant="isNotFound ? 'primary' : 'secondary'" block>
          <UiIcon name="back" :size="16" /> Ana sayfaya dön
        </UiButton>
      </div>
    </div>

    <template #footer>
      Kendi kısa bağlantınızı oluşturmak için
      <NuxtLink to="/dashboard/new" class="font-semibold text-accent-dark underline underline-offset-2">
        yeni link kısaltın
      </NuxtLink>
    </template>
  </AuthCard>
</template>
