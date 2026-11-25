<script setup lang="ts">
import type { MenuItem } from '~/components/ui/Menu.vue'

const route = useRoute()
const { isAdmin, user } = useAuth()
const { switchTo } = useSpace()
const signOut = useSignOut()

const items = computed<MenuItem[]>(() => [
  ...(isAdmin.value ? [{ label: 'Yönetim paneline geç', icon: 'swap', onSelect: () => switchTo('admin') }] : []),
  { label: 'Çıkış yap', icon: 'logout', onSelect: signOut, divider: isAdmin.value }
])

const short: Record<string, string> = { '/dashboard/settings': 'Hesap', '/dashboard': 'Genel' }
</script>

<template>
  <div class="lg:hidden">
    <header class="sticky top-0 z-30 flex h-15 items-center justify-between border-b border-line bg-surface px-5">
      <BrandLogo />
      <div class="-mr-2 flex items-center">
        <UiThemeToggle />
        <UiMenu :label="user?.name || 'Hesap menüsü'" icon="user" :items="items" />
      </div>
    </header>

    <nav
      class="fixed inset-x-0 bottom-0 z-30 grid grid-cols-5 items-center border-t border-line bg-surface pb-[env(safe-area-inset-bottom)]"
      aria-label="Panel menüsü"
    >
      <template v-for="(item, i) in userNav" :key="item.to">
        <NuxtLink v-if="i === 2" to="/dashboard/new" class="flex h-16 items-center justify-center" aria-label="Yeni link">
          <span class="cut-corner flex h-12 w-12 items-center justify-center text-on-brand"><UiIcon name="plus" :size="22" /></span>
        </NuxtLink>
        <NuxtLink
          :to="item.to"
          class="flex h-16 flex-col items-center justify-center gap-1 text-[12px]"
          :class="isNavActive(item, route.path) ? 'font-semibold text-brand-ink' : 'text-ink-3'"
          :aria-current="isNavActive(item, route.path) ? 'page' : undefined"
        >
          <UiIcon :name="item.icon" :size="20" />
          {{ short[item.to] ?? item.label }}
        </NuxtLink>
      </template>
    </nav>
  </div>
</template>
