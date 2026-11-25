<script setup lang="ts">
withDefaults(defineProps<{ tone?: 'normal' | 'mat' }>(), { tone: 'normal' })

const { isDark, toggleTheme, syncFromDom } = useTheme()
onMounted(syncFromDom)

const label = computed(() => (isDark.value ? 'Aydınlık temaya geç' : 'Karanlık temaya geç'))
</script>

<template>
  <UiIconButton :icon="isDark ? 'sun' : 'moon'" :label="label" :tone="tone" @click="toggleTheme" />
</template>
