<script setup lang="ts">
import type { NavItem } from '~/utils/navigation'

const props = withDefaults(
  defineProps<{
    item: NavItem
    tone?: 'personal' | 'admin' | 'owner'
  }>(),
  { tone: 'personal' }
)
const route = useRoute()

const active = computed(() =>
  props.item.exact ? route.path === props.item.to : route.path.startsWith(props.item.to)
)

const activeBorderClass = computed(() => {
  if (!active.value) return 'border-transparent text-ink-2 hover:bg-paper-2 hover:text-ink'
  if (props.tone === 'admin') return 'border-ink border-l-4 border-l-amber-500 bg-card text-ink shadow-hard'
  if (props.tone === 'owner') return 'border-ink border-l-4 border-l-violet-500 bg-card text-ink shadow-hard'
  return 'border-ink border-l-4 border-l-accent bg-card text-ink shadow-hard'
})
</script>

<template>
  <NuxtLink
    :to="item.to"
    class="flex items-center gap-3 rounded-md border-[1.5px] px-3 py-2 text-sm font-medium transition-colors"
    :class="activeBorderClass"
    :aria-current="active ? 'page' : undefined"
  >
    <UiIcon :name="item.icon" />
    {{ item.label }}
  </NuxtLink>
</template>
