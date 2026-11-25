<script setup lang="ts">
const { user, primaryRole, isAdmin, isOwner, hasAnyRole } = useAuth()
const { isDark, toggleTheme } = useTheme()
const route = useRoute()

const visibleNavGroups = computed(() =>
  dashboardNav
    .filter((group) => hasAnyRole(group.roles))
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => hasAnyRole(item.roles))
    }))
    .filter((group) => group.items.length > 0)
)

const activeGroupAccent = (title: string): 'personal' | 'admin' | 'owner' => {
  if (title === 'Yönetici') return 'admin'
  if (title === 'Özel') return 'owner'
  return 'personal'
}

const currentWorkspaceLabel = computed(() => {
  if (route.path.startsWith('/dashboard/owner')) return 'OWNER ALANI'
  if (route.path.startsWith('/dashboard/admin')) return 'ADMİN ALANI'
  return 'KİŞİSEL ALAN'
})

const roleBadge = computed(() => {
  if (isOwner.value) return { label: 'OWNER', tone: 'accent' as const }
  if (isAdmin.value) return { label: 'ADMİN', tone: 'ok' as const }
  return { label: 'KULLANICI', tone: 'neutral' as const }
})

const quotaText = computed(() => {
  if (isAdmin.value || (user.value?.plan?.maxLinks ?? 50) < 0) {
    return `${user.value?.linkCount ?? 0} / Sınırsız`
  }
  return `${user.value?.linkCount ?? 0} / ${user.value?.plan?.maxLinks ?? 50}`
})

const quotaPercent = computed(() => {
  const max = user.value?.plan?.maxLinks ?? 50
  if (isAdmin.value || max <= 0) return 0
  return Math.min(100, Math.round(((user.value?.linkCount ?? 0) / max) * 100))
})
</script>

<template>
  <aside class="flex h-full w-66 flex-col border-r border-line bg-paper-2/60 p-4">
    <div class="flex items-center justify-between px-2 py-2">
      <BrandLogo />
      <button
        type="button"
        class="inline-flex h-8 w-8 cursor-pointer items-center justify-center rounded-sm border border-line bg-card text-ink-2 transition-colors hover:border-ink hover:text-ink"
        :title="isDark ? 'Gündüz moduna geç' : 'Gece moduna geç'"
        :aria-label="isDark ? 'Gündüz moduna geç' : 'Gece moduna geç'"
        @click="toggleTheme"
      >
        <UiIcon :name="isDark ? 'sun' : 'moon'" :size="15" />
      </button>
    </div>

    <div class="mx-2 mt-2 rounded-sm border border-line bg-card/80 px-2.5 py-1.5 font-mono text-[10px] font-bold tracking-wider text-ink-2">
      AKTİF MOD: <span class="text-ink">{{ currentWorkspaceLabel }}</span>
    </div>

    <nav class="mt-5 flex flex-1 flex-col gap-5 overflow-y-auto pr-0.5" aria-label="Panel menüsü">
      <div v-for="group in visibleNavGroups" :key="group.title">
        <p
          class="eyebrow mb-2 flex items-center gap-1.5 px-3"
          :class="{
            '!text-amber-600 dark:!text-amber-400': group.title === 'Yönetici',
            '!text-violet-600 dark:!text-violet-400': group.title === 'Özel',
            '!text-ink-3': group.title !== 'Yönetici' && group.title !== 'Özel'
          }"
        >
          <span
            class="h-1.5 w-1.5 rounded-full"
            :class="{
              'bg-amber-500': group.title === 'Yönetici',
              'bg-violet-500': group.title === 'Özel',
              'bg-accent': group.title !== 'Yönetici' && group.title !== 'Özel'
            }"
          />
          {{ group.title }}
        </p>
        <ul class="flex flex-col gap-1">
          <li v-for="item in group.items" :key="item.to">
            <DashboardSidebarLink :item="item" :tone="activeGroupAccent(group.title)" />
          </li>
        </ul>
      </div>
    </nav>

    <div class="mt-4 border-t border-line pt-4">
      <div class="mb-3 rounded-md border border-line bg-card p-3">
        <div class="flex items-center justify-between gap-2">
          <p class="truncate text-sm font-semibold text-ink">{{ user?.name }}</p>
          <UiBadge :tone="roleBadge.tone">{{ roleBadge.label }}</UiBadge>
        </div>
        <p class="mt-0.5 truncate text-xs text-ink-3">{{ user?.email }}</p>

        <div class="mt-2.5 border-t border-line pt-2 text-[11px]">
          <div class="flex items-center justify-between text-ink-2">
            <span class="font-medium">{{ user?.plan?.name ?? 'Ücretsiz Paket' }}</span>
            <span class="font-mono font-bold text-ink">{{ quotaText }}</span>
          </div>
          <div
            v-if="!isAdmin && (user?.plan?.maxLinks ?? 50) > 0"
            class="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-paper-2"
          >
            <div class="h-full bg-accent transition-all" :style="{ width: `${quotaPercent}%` }" />
          </div>
          <p v-else-if="primaryRole !== 'user'" class="mt-1 font-mono text-[10px] text-ink-3">
            Yetkili hesap — kota muafiyeti
          </p>
        </div>
      </div>

      <NuxtLink
        to="/auth/logout"
        class="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-ink-2 transition-colors hover:bg-accent-soft hover:text-accent-dark"
      >
        <UiIcon name="logout" /> Çıkış yap
      </NuxtLink>
    </div>
  </aside>
</template>
