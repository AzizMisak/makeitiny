<script setup lang="ts">
const { isLoggedIn } = useAuth()
</script>

<template>
  <header class="sticky top-0 z-20 border-b border-line bg-paper/90 backdrop-blur-sm">
    <div class="container-x flex h-16 items-center justify-between">
      <BrandLogo />

      <nav class="hidden items-center gap-8 md:flex" aria-label="Ana menü">
        <NuxtLink
          v-for="item in landingNav"
          :key="item.href"
          :to="item.href"
          class="text-sm font-medium text-ink-2 transition-colors hover:text-ink"
        >
          {{ item.label }}
        </NuxtLink>
      </nav>

      <div class="flex items-center gap-2">
        <template v-if="isLoggedIn">
          <UiButton to="/dashboard" variant="secondary">Panele git</UiButton>
        </template>
        <template v-else>
          <UiButton to="/auth/login" variant="ghost">Giriş yap</UiButton>
          <UiButton to="/auth/register" variant="primary" class="hidden sm:inline-flex">Ücretsiz başla</UiButton>
        </template>
      </div>
    </div>
  </header>
</template>
