<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Yetkisiz Erişim — makeiTiny' })

const route = useRoute()
const { primaryRoleLabel, user } = useAuth()

const targetPath = computed(() => String(route.query.from ?? ''))
const requiredRoles = computed(() => String(route.query.required ?? ''))
</script>

<template>
  <DashboardPageHeader
    title="Yetkisiz erişim"
    description="Bu sayfayı görüntülemek için gerekli yetkiye sahip değilsin."
  />

  <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8">
    <div class="flex flex-wrap items-start justify-between gap-4 border-b border-line pb-5">
      <div class="flex items-start gap-3">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md border-[1.5px] border-ink bg-accent-soft text-accent-dark">
          <UiIcon name="ban" :size="20" />
        </div>
        <div>
          <h2 class="text-lg font-bold text-ink">Bu alana erişim iznin yok</h2>
          <p class="mt-0.5 text-sm text-ink-2">
            İstediğin sayfa mevcut rolünün yetkisi dışında.
          </p>
        </div>
      </div>
      <UiBadge tone="accent">403</UiBadge>
    </div>

    <div class="mt-5 grid gap-4 sm:grid-cols-3">
      <div class="rounded-md border border-line bg-paper p-4">
        <p class="text-xs text-ink-3">Mevcut rol</p>
        <p class="mt-1 font-mono text-sm font-bold text-ink">{{ primaryRoleLabel }}</p>
      </div>
      <div class="rounded-md border border-line bg-paper p-4">
        <p class="text-xs text-ink-3">Gereken rol</p>
        <p class="mt-1 font-mono text-sm font-bold text-accent-dark">{{ requiredRoles || 'Yetkili rol' }}</p>
      </div>
      <div class="rounded-md border border-line bg-paper p-4">
        <p class="text-xs text-ink-3">Hesap</p>
        <p class="mt-1 truncate font-mono text-sm font-semibold text-ink">{{ user?.email }}</p>
      </div>
    </div>

    <p v-if="targetPath" class="mt-4 font-mono text-xs text-ink-3">
      Hedef: <span class="text-ink">{{ targetPath }}</span>
    </p>

    <div class="mt-6 flex flex-wrap gap-3">
      <UiButton to="/dashboard">
        <UiIcon name="grid" :size="16" /> Genel bakış
      </UiButton>
      <UiButton to="/dashboard/settings" variant="secondary">
        <UiIcon name="sliders" :size="16" /> Hesap ayarları
      </UiButton>
    </div>
  </div>
</template>
