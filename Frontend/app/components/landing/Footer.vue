<script setup lang="ts">
const route = useRoute()

function scrollToSection(event: MouseEvent, sectionId: string) {
  if (route.path !== '/' || !import.meta.client) return
  event.preventDefault()
  const el = document.getElementById(sectionId)
  if (!el) return
  const top = Math.max(0, el.getBoundingClientRect().top + window.scrollY - 72)
  window.scrollTo({ top, behavior: 'smooth' })
  window.history.replaceState(null, '', `/#${sectionId}`)
}
</script>

<template>
  <footer class="border-t-[1.5px] border-ink bg-card">
    <div class="container-x flex flex-col items-start justify-between gap-6 py-8 text-sm text-ink-3 sm:flex-row sm:items-center">
      <div class="flex flex-col gap-1">
        <BrandLogo />
        <p class="text-xs text-ink-3">Hızlı link kısaltma, şifreli/bekletmeli yönlendirme ve tıklama analitiği.</p>
      </div>

      <div class="flex flex-wrap items-center gap-6 text-xs font-semibold text-ink-2">
        <a
          v-for="item in landingNav"
          :key="item.href"
          :href="item.href"
          class="transition-colors hover:text-accent-dark"
          @click="scrollToSection($event, item.sectionId)"
        >
          {{ item.label }}
        </a>
      </div>

      <p class="font-mono text-xs text-ink-3">© 2026 makeiTiny.</p>
    </div>
  </footer>
</template>
