<script setup lang="ts">
const route = useRoute()
const menuOpen = ref(false)

watch(() => route.path, () => (menuOpen.value = false))
</script>

<template>
  <div class="min-h-screen lg:grid lg:grid-cols-[16rem_1fr]">
    <!-- Masaüstü menü -->
    <div class="sticky top-0 hidden h-screen lg:block">
      <DashboardSidebar />
    </div>

    <!-- Mobil menü -->
    <div v-if="menuOpen" class="fixed inset-0 z-30 flex lg:hidden">
      <DashboardSidebar class="shadow-hard-lg" />
      <button type="button" class="flex-1 bg-ink/40" aria-label="Menüyü kapat" @click="menuOpen = false" />
    </div>

    <div class="min-w-0">
      <DashboardTopbar @toggle-menu="menuOpen = !menuOpen" />
      <main class="mx-auto w-full max-w-5xl px-5 py-8 md:px-10 md:py-12">
        <slot />
      </main>
    </div>
  </div>
</template>
