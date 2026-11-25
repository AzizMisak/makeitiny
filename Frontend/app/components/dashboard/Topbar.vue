<script setup lang="ts">
defineProps<{
  workspace?: 'personal' | 'admin' | 'owner'
  workspaceLabel?: string
}>()
defineEmits<{ toggleMenu: [] }>()

const { isDark, toggleTheme } = useTheme()
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
    <div class="flex items-center gap-2">
      <BrandLogo />
      <span
        v-if="workspaceLabel"
        class="hidden rounded-xs border border-line bg-card px-1.5 py-0.5 font-mono text-[10px] font-bold sm:inline-block"
      >
        {{ workspace === 'admin' ? 'ADMİN' : workspace === 'owner' ? 'OWNER' : 'KİŞİSEL' }}
      </span>
    </div>
    <button
      type="button"
      class="inline-flex h-9 w-9 cursor-pointer items-center justify-center rounded-md border border-line bg-card text-ink hover:bg-paper-2"
      :aria-label="isDark ? 'Gündüz moduna geç' : 'Gece moduna geç'"
      @click="toggleTheme"
    >
      <UiIcon :name="isDark ? 'sun' : 'moon'" :size="16" />
    </button>
  </header>
</template>
