<script setup lang="ts">
definePageMeta({ layout: 'marketing' })

const route = useRoute()
const alias = computed(() => String(route.query.alias ?? route.query.link ?? ''))
const reason = computed(() => String(route.query.reason ?? 'not_found'))

const isTargetUnreachable = computed(() => reason.value === 'unreachable')

useHead({ title: 'Yönlendirme Hatası — makeiTiny' })

onMounted(async () => {
  if (alias.value) {
    await navigateTo(
      {
        path: `/error/${alias.value}`,
        query: route.query.reason ? { reason: String(route.query.reason) } : undefined
      },
      { replace: true }
    )
  }
})
</script>

<template>
  <AuthCard
    :title="isTargetUnreachable ? 'Hedef adrese ulaşılamıyor.' : 'Kısa link bulunamadı.'"
    :subtitle="
      isTargetUnreachable
        ? 'Yönlendirilen hedef siteye şu anda bağlantı kurulamıyor.'
        : 'Aradığınız kısa bağlantı sistemimizde kayıtlı değil veya geçersiz.'
    "
  >
    <div class="flex flex-col gap-4">
      <p class="text-sm text-ink-2">
        <template v-if="isTargetUnreachable">
          Kısa bağlantı geçerli ancak hedef web sitesi yanıt vermiyor. Lütfen daha sonra tekrar deneyin.
        </template>
        <template v-else>
          Yönlendirme için geçerli ve aktif bir kısa bağlantı kullanmalısınız.
        </template>
      </p>
      <UiButton to="/" variant="primary" block>
        <UiIcon name="back" :size="16" /> Ana sayfaya dön
      </UiButton>
    </div>
  </AuthCard>
</template>
