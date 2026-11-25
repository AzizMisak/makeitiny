<script setup lang="ts">
/** Kaydırma çubuğunun yerini alan şerit metre. Tıklanan yere kaydırır, bölümlere atlatır. */
useHead({ htmlAttrs: { 'data-scroll': 'ruler' } })

const progress = ref(0)
const stops = ref<{ id: string; label: string; at: number }[]>([])
const active = ref('')

function measure() {
  const room = document.documentElement.scrollHeight - window.innerHeight
  stops.value = landingSections.flatMap((s) => {
    const el = document.getElementById(s.id)
    if (!el) return []
    const top = el.getBoundingClientRect().top + window.scrollY
    return [{ ...s, at: room > 0 ? Math.min(1, top / room) : 0 }]
  })
  onScroll()
}

function onScroll() {
  const room = document.documentElement.scrollHeight - window.innerHeight
  progress.value = room > 0 ? Math.min(1, Math.max(0, window.scrollY / room)) : 0
  let current = ''
  for (const s of stops.value) if (progress.value >= s.at - 0.02) current = s.id
  active.value = current
}

function jump(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' })
}

function jumpTo(e: MouseEvent) {
  const box = (e.currentTarget as HTMLElement).getBoundingClientRect()
  const ratio = (e.clientY - box.top) / box.height
  window.scrollTo({ top: ratio * (document.documentElement.scrollHeight - window.innerHeight), behavior: 'smooth' })
}

let observer: ResizeObserver | null = null

onMounted(() => {
  measure()
  window.addEventListener('scroll', onScroll, { passive: true })
  observer = new ResizeObserver(measure)
  observer.observe(document.body)
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
  observer?.disconnect()
})
</script>

<template>
  <nav
    class="fixed inset-y-0 right-0 z-40 hidden w-9 cursor-pointer bg-mark shadow-[-2px_0_8px_rgb(0_0_0/0.18)] lg:block"
    aria-label="Sayfa bölümleri"
    @click.self="jumpTo"
  >
    <svg class="pointer-events-none absolute inset-0 h-full w-full" aria-hidden="true">
      <defs>
        <pattern id="tape-ticks" width="36" height="60" patternUnits="userSpaceOnUse">
          <path
            d="M0 .5h20M0 6.5h7M0 12.5h7M0 18.5h7M0 24.5h7M0 30.5h13M0 36.5h7M0 42.5h7M0 48.5h7M0 54.5h7"
            stroke="#2a1f00"
            stroke-width="1"
          />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#tape-ticks)" />
    </svg>

    <button
      v-for="s in stops"
      :key="s.id"
      type="button"
      class="group absolute right-0 flex h-7 w-9 -translate-y-1/2 items-center justify-end"
      :style="{ top: `calc(1.75rem + ${s.at} * (100% - 3.5rem))` }"
      :aria-current="active === s.id ? 'true' : undefined"
      :aria-label="s.label"
      @click="jump(s.id)"
    >
      <span
        class="pointer-events-none absolute right-11 whitespace-nowrap bg-[#10211b] px-3 py-1.5 text-sm font-medium text-white opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100"
      >{{ s.label }}</span>
      <span class="mr-1.5 h-3 w-3 border-2 border-[#2a1f00]" :class="active === s.id ? 'bg-[#2a1f00]' : 'bg-mark'" />
    </button>

    <span
      class="pointer-events-none absolute inset-x-0 h-[3px] bg-[#b3261e]"
      :style="{ top: `calc(1.75rem + ${progress} * (100% - 3.5rem))` }"
      aria-hidden="true"
    />
  </nav>
</template>
