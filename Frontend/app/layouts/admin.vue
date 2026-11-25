<script setup lang="ts">
import type { MenuItem } from '~/components/ui/Menu.vue'

useHead({ htmlAttrs: { 'data-space': 'admin' } })

const route = useRoute()
const { user, isOwner } = useAuth()
const { remember, switchTo, switchingTo } = useSpace()
const signOut = useSignOut()

watch(() => route.fullPath, remember, { immediate: true })

const menu = computed<MenuItem[]>(() => [{ label: 'Çıkış yap', icon: 'logout', onSelect: signOut }])
</script>

<template>
  <div class="min-h-screen">
    <header class="mat-grid on-mat sticky top-0 z-30 text-on-mat">
      <div class="mx-auto flex w-full max-w-[92rem] flex-wrap items-center gap-x-4 px-5 md:gap-x-8 md:px-8">
        <div class="flex h-16 items-baseline gap-3 pt-[1.35rem]">
          <BrandLogo to="/dashboard/admin" tone="mat" class="max-sm:hidden" />
          <span class="display text-xl text-on-mat-2 sm:text-lg">Yönetim</span>
        </div>

        <nav
          class="order-last -mx-5 flex w-[calc(100%+2.5rem)] items-center gap-1 overflow-x-auto px-5 [scrollbar-width:none] md:order-none md:mx-0 md:w-auto md:px-0"
          aria-label="Yönetim menüsü"
        >
          <NuxtLink
            v-for="item in adminNav"
            :key="item.to"
            :to="item.to"
            class="relative flex h-12 shrink-0 items-center gap-2 px-3.5 text-[15px] transition-colors md:h-16"
            :class="isNavActive(item, route.path) ? 'font-semibold text-on-mat' : 'text-on-mat-2 hover:text-on-mat'"
            :aria-current="isNavActive(item, route.path) ? 'page' : undefined"
          >
            <UiIcon :name="item.icon" :size="17" />
            {{ item.label }}
            <span v-if="isNavActive(item, route.path)" class="absolute inset-x-3.5 bottom-0 h-1 bg-mark" />
          </NuxtLink>

          <template v-if="isOwner">
            <span class="mx-2 h-6 w-px shrink-0 bg-white/30" aria-hidden="true" />
            <NuxtLink
              v-for="item in ownerNav"
              :key="item.to"
              :to="item.to"
              title="Yalnızca sistem sahibi görür"
              class="relative flex h-12 shrink-0 items-center gap-2 px-3.5 text-[15px] transition-colors md:h-16"
              :class="isNavActive(item, route.path) ? 'font-semibold text-mark' : 'text-mark/85 hover:text-mark'"
              :aria-current="isNavActive(item, route.path) ? 'page' : undefined"
            >
              <UiIcon :name="item.icon" :size="17" />
              {{ item.label }}
              <span v-if="isNavActive(item, route.path)" class="absolute inset-x-3.5 bottom-0 h-1 bg-mark" />
            </NuxtLink>
          </template>
        </nav>

        <div class="ml-auto flex items-center gap-1">
          <UiButton
            variant="mat"
            size="sm"
            icon="swap"
            :loading="switchingTo === 'user'"
            aria-label="Kullanıcı paneline dön"
            @click="switchTo('user')"
          >
            <span class="hidden sm:inline">Kullanıcı paneline dön</span>
          </UiButton>
          <UiThemeToggle tone="mat" />
          <UiMenu :label="user?.email || 'Hesap menüsü'" icon="user" tone="mat" :items="menu" />
        </div>
      </div>
    </header>

    <main class="mx-auto w-full max-w-[92rem] px-5 pb-24 pt-8 md:px-8">
      <slot />
    </main>
  </div>
</template>
