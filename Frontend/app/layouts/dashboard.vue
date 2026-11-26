<script setup lang="ts">
const route = useRoute()
const { remember } = useSpace()

watch(() => route.fullPath, remember, { immediate: true })
</script>

<template>
  <div class="min-h-screen lg:grid lg:grid-cols-[17rem_minmax(0,1fr)]">
    <div class="sticky top-0 hidden h-screen lg:block">
      <PanelSidebar />
    </div>
    <div class="min-w-0">
      <PanelMobileBar />
      <main class="mx-auto w-full max-w-5xl px-5 pb-36 pt-8 md:px-10 lg:pb-20 lg:pt-12">
        <slot />
      </main>
    </div>
  </div>
</template>
