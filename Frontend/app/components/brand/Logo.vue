<script setup lang="ts">
withDefaults(defineProps<{ light?: boolean }>(), { light: false })
</script>

<template>
  <NuxtLink to="/" class="inline-flex items-center gap-2.5" aria-label="makeiTiny ana sayfa">
    <svg width="30" height="30" viewBox="0 0 32 32" fill="none" aria-hidden="true">
      <rect x="1" y="1" width="30" height="30" rx="7" :fill="light ? '#f5f1e8' : '#1b1a17'" />
      <path d="M6 11h20" :stroke="light ? '#1b1a17' : '#f5f1e8'" stroke-width="2.4" stroke-linecap="round" />
      <path d="M6 16h13" :stroke="light ? '#1b1a17' : '#f5f1e8'" stroke-width="2.4" stroke-linecap="round" />
      <path d="M6 21h5" stroke="#e5481f" stroke-width="2.4" stroke-linecap="round" />
    </svg>
    <span class="font-display text-[22px] font-extrabold tracking-tight" :class="light ? 'text-paper' : 'text-ink'">
      make<span class="text-accent">i</span>Tiny
    </span>
  </NuxtLink>
</template>
