<script setup lang="ts">
defineEmits<{ toggleMenu: [] }>()
</script>

<template>
  <header class="flex h-16 items-center justify-between border-b border-line bg-paper px-5 lg:hidden">
    <button
      type="button"
      class="rounded-md p-2 text-ink hover:bg-paper-2"
      aria-label="Menüyü aç"
      @click="$emit('toggleMenu')"
    >
      <UiIcon name="menu" :size="22" />
    </button>
    <BrandLogo />
    <span class="w-10" />
  </header>
</template>
