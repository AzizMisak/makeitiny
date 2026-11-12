<script setup lang="ts">
definePageMeta({ layout: 'marketing' })

const route = useRoute()
const alias = computed(() => String(route.query.alias ?? route.query.link ?? ''))

onMounted(async () => {
  if (alias.value) {
    await navigateTo(`/w/${alias.value}`, { replace: true })
  }
})
</script>

<template>
  <AuthCard
    title="Ara Bekletme Sayfası"
    subtitle="Geçerli bir kısa link kodu belirtilmedi."
  >
    <div class="flex flex-col gap-4">
      <p class="text-sm text-ink-2">
        Yönlendirme için geçerli bir kısa bağlantı kullanmalısınız.
      </p>
      <UiButton to="/" variant="secondary" block>Ana sayfaya dön</UiButton>
    </div>
  </AuthCard>
</template>
