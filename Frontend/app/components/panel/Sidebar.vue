<script setup lang="ts">
const route = useRoute()
const { user, isAdmin } = useAuth()
const { switchTo, switchingTo } = useSpace()
const signOut = useSignOut()
</script>

<template>
  <aside class="flex h-full w-full flex-col border-r border-line bg-surface">
    <div class="px-6 pb-5 pt-7">
      <BrandLogo />
    </div>

    <div class="px-4">
      <UiButton to="/dashboard/new" icon="plus" block :aria-current="route.path === '/dashboard/new' ? 'page' : undefined">
        Yeni link
      </UiButton>
    </div>

    <nav class="mt-6 flex-1 overflow-y-auto px-4" aria-label="Panel menüsü">
      <ul class="flex flex-col gap-1">
        <li v-for="item in userNav" :key="item.to">
          <NuxtLink
            :to="item.to"
            :aria-current="isNavActive(item, route.path) ? 'page' : undefined"
            class="flex min-h-11 items-center gap-3 px-3 text-[15px] transition-colors"
            :class="
              isNavActive(item, route.path)
                ? 'cut-corner font-semibold text-brand-ink [--cut-bg:var(--color-brand-soft)]'
                : 'rounded-md text-ink-2 hover:bg-sunken hover:text-ink'
            "
          >
            <UiIcon :name="item.icon" :size="19" />
            {{ item.label }}
          </NuxtLink>
        </li>
      </ul>
    </nav>

    <!-- Yönetim paneline geçiş: menü öğesi değil, başka bir yere açılan kapı. -->
    <div v-if="isAdmin" class="px-4 pb-4">
      <button
        type="button"
        class="mat-grid on-mat flex w-full items-center justify-between gap-3 px-4 py-3.5 text-left text-white transition-[filter] hover:brightness-110"
        style="background-color: #173f80"
        :disabled="!!switchingTo"
        @click="switchTo('admin')"
      >
        <span>
          <span class="display block text-lg">Yönetim paneli</span>
          <span class="block text-[13px] text-white/80">Tüm kullanıcılar ve linkler</span>
        </span>
        <UiSpinner v-if="switchingTo === 'admin'" />
        <UiIcon v-else name="swap" :size="20" />
      </button>
    </div>

    <div class="flex items-center gap-1 border-t border-line py-3 pl-6 pr-3">
      <div class="min-w-0 flex-1">
        <p class="truncate text-sm font-semibold">{{ user?.name || 'Hesabım' }}</p>
        <p class="truncate text-[13px] text-ink-3">{{ user?.email }}</p>
      </div>
      <UiThemeToggle />
      <UiIconButton icon="logout" label="Çıkış yap" @click="signOut" />
    </div>
  </aside>
</template>
