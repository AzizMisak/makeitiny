<script setup lang="ts">
defineProps<{ title: string; subtitle: string }>()
</script>

<template>
  <section class="container-x flex justify-center py-14 md:py-20">
    <div class="w-full max-w-md">
      <h1 class="text-4xl font-extrabold">{{ title }}</h1>
      <p class="mt-3 text-ink-2">{{ subtitle }}</p>

      <div class="mt-8 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard-lg sm:p-8">
        <slot />
      </div>

      <p class="mt-6 text-center text-sm text-ink-2">
        <slot name="footer" />
      </p>
    </div>
  </section>
</template>
