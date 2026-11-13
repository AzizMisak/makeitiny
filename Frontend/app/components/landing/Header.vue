<script setup lang="ts">
const { isLoggedIn } = useAuth()
const route = useRoute()
const router = useRouter()

function scrollToElementById(sectionId: string) {
  if (!import.meta.client) return
  const el = document.getElementById(sectionId)
  if (!el) return
  const headerOffset = 72
  const elementPosition = el.getBoundingClientRect().top + window.scrollY
  const offsetPosition = Math.max(0, elementPosition - headerOffset)

  window.scrollTo({
    top: offsetPosition,
    behavior: 'smooth'
  })
}

async function handleSectionClick(event: MouseEvent, item: { href: string; sectionId: string }) {
  event.preventDefault()

  if (route.path !== '/') {
    await router.push('/')
    setTimeout(() => {
      scrollToElementById(item.sectionId)
      window.history.replaceState(null, '', `/#${item.sectionId}`)
    }, 120)
    return
  }

  scrollToElementById(item.sectionId)
  window.history.replaceState(null, '', `/#${item.sectionId}`)
}

function handleLogoClick(event: MouseEvent) {
  if (route.path === '/' && import.meta.client) {
    event.preventDefault()
    window.scrollTo({ top: 0, behavior: 'smooth' })
    window.history.replaceState(null, '', '/')
  }
}
</script>

<template>
  <header class="sticky top-0 z-30 border-b-[1.5px] border-ink bg-paper/95 backdrop-blur-md">
    <div class="container-x flex h-16 items-center justify-between gap-4">
      <div @click.capture="handleLogoClick">
        <BrandLogo />
      </div>

      <nav class="hidden items-center gap-2 md:flex" aria-label="Ana menü">
        <a
          v-for="item in landingNav"
          :key="item.href"
          :href="item.href"
          class="rounded-md border border-transparent px-3.5 py-1.5 text-sm font-semibold text-ink-2 transition-all duration-150 hover:border-ink hover:bg-card hover:text-ink hover:shadow-[2px_2px_0_var(--color-ink)] active:translate-x-px active:translate-y-px active:shadow-none"
          @click="handleSectionClick($event, item)"
        >
          {{ item.label }}
        </a>
      </nav>

      <div class="flex items-center gap-2.5">
        <template v-if="isLoggedIn">
          <UiButton to="/dashboard" variant="secondary">Panele git</UiButton>
          <UiButton to="/dashboard/new" variant="primary" class="hidden sm:inline-flex">
            <UiIcon name="plus" :size="15" /> Yeni link
          </UiButton>
        </template>
        <template v-else>
          <UiButton to="/auth/login" variant="ghost">Giriş yap</UiButton>
          <UiButton to="/auth/register" variant="primary" class="hidden sm:inline-flex">
            Ücretsiz başla <UiIcon name="arrow" :size="15" />
          </UiButton>
        </template>
      </div>
    </div>

    <!-- Mobil Hızlı Bölüm Geçiş Çubuğu -->
    <div class="border-t border-line bg-card/80 md:hidden">
      <nav class="container-x flex items-center justify-center gap-4 py-2" aria-label="Sayfa içi menü">
        <a
          v-for="item in landingNav"
          :key="item.href"
          :href="item.href"
          class="text-xs font-semibold text-ink-2 transition-colors hover:text-accent-dark"
          @click="handleSectionClick($event, item)"
        >
          {{ item.label }}
        </a>
      </nav>
    </div>
  </header>
</template>
