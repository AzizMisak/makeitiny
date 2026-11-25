<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Yetkin yok | makeiTiny' })

const route = useRoute()
const { roles } = useAuth()

const names: Record<string, string> = { admin: 'Yönetici', owner: 'Sistem sahibi', user: 'Kullanıcı' }
const required = computed(() =>
  String(route.query.required ?? '')
    .split(',')
    .map((r) => names[r.trim()] ?? r.trim())
    .filter(Boolean)
    .join(' veya ')
)
const from = computed(() => String(route.query.from ?? ''))
</script>

<template>
  <div class="max-w-2xl py-10">
    <UiIcon name="lock" :size="40" class="text-line-2" />
    <h1 class="display mt-5 text-[2.5rem]">Bu sayfa sana kapalı</h1>
    <p class="mt-3 text-lg text-ink-2">
      <template v-if="from"><span class="font-mono text-base">{{ from }}</span> sayfasını </template><template v-else>Sayfayı </template>
      açmak için {{ required || 'daha yüksek bir' }} rolü gerekiyor. Senin rolün: {{ roleMeta(roles).label }}.
    </p>
    <UiButton to="/dashboard" class="mt-8" icon="back">Genel bakışa dön</UiButton>
  </div>
</template>
