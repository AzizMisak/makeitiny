<script setup lang="ts">
const props = withDefaults(defineProps<{ name: string; size?: number }>(), { size: 18 })

const icons: Record<string, string[]> = {
  grid: ['M4 4h6v6H4z', 'M14 4h6v6h-6z', 'M4 14h6v6H4z', 'M14 14h6v6h-6z'],
  link: ['M10 13a5 5 0 0 0 7.07 0l3-3a5 5 0 0 0-7.07-7.07l-1 1', 'M14 11a5 5 0 0 0-7.07 0l-3 3a5 5 0 0 0 7.07 7.07l1-1'],
  plus: ['M12 5v14', 'M5 12h14'],
  sliders: ['M4 6h9', 'M17 6h3', 'M4 12h3', 'M11 12h9', 'M4 18h11', 'M19 18h1', 'M15 4a2 2 0 1 0 0 4a2 2 0 1 0 0-4', 'M9 10a2 2 0 1 0 0 4a2 2 0 1 0 0-4', 'M17 16a2 2 0 1 0 0 4a2 2 0 1 0 0-4'],
  logout: ['M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4', 'M16 17l5-5-5-5', 'M21 12H9'],
  chart: ['M4 20v-6', 'M10 20V5', 'M16 20v-9', 'M21 20H3'],
  lock: ['M6 11h12v9H6z', 'M8 11V8a4 4 0 0 1 8 0v3'],
  arrow: ['M5 12h14', 'M13 6l6 6-6 6'],
  copy: ['M9 9h11v11H9z', 'M5 15V5a1 1 0 0 1 1-1h10'],
  trash: ['M4 7h16', 'M9 7V4h6v3', 'M6 7l1 13h10l1-13'],
  check: ['M5 12l5 5 9-10'],
  eye: ['M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12z', 'M12 9a3 3 0 1 0 0 6a3 3 0 0 0 0-6'],
  menu: ['M4 7h16', 'M4 12h16', 'M4 17h16']
}

const paths = computed(() => icons[props.name] ?? [])
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.8"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path v-for="d in paths" :key="d" :d="d" />
  </svg>
</template>
