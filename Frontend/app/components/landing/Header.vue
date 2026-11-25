<script setup lang="ts">
const { isLoggedIn } = useAuth()
</script>

<template>
  <header class="on-mat bg-mat text-on-mat">
    <div class="container-x flex h-18 items-center justify-between gap-4">
      <BrandLogo tone="mat" />
      <div class="flex items-center gap-1">
        <UiThemeToggle tone="mat" />
        <UiButton v-if="isLoggedIn" to="/dashboard" variant="mat" size="sm">Panele git</UiButton>
        <template v-else>
          <NuxtLink to="/auth/login" class="whitespace-nowrap px-3 py-2 text-[15px] font-medium underline-offset-4 hover:underline">Giriş yap</NuxtLink>
          <UiButton to="/auth/register" variant="mat" size="sm" class="max-sm:hidden">Hesap aç</UiButton>
        </template>
      </div>
    </div>
  </header>
</template>
