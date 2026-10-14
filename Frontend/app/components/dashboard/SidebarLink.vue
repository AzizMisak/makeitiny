<script setup lang="ts">
import type { NavItem } from '~/utils/navigation'

const props = defineProps<{ item: NavItem }>()
const route = useRoute()

const active = computed(() =>
  props.item.exact ? route.path === props.item.to : route.path.startsWith(props.item.to)
)
</script>

<template>
  <NuxtLink
    :to="item.to"
    class="flex items-center gap-3 rounded-md border-[1.5px] px-3 py-2 text-sm font-medium transition-colors"
    :class="
      active
        ? 'border-ink bg-card text-ink shadow-hard'
        : 'border-transparent text-ink-2 hover:bg-paper-2 hover:text-ink'
    "
    :aria-current="active ? 'page' : undefined"
  >
    <UiIcon :name="item.icon" />
    {{ item.label }}
  </NuxtLink>
</template>
