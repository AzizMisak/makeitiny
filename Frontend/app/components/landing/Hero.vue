<script setup lang="ts">
const { isLoggedIn } = useAuth()

function scrollToFeatures() {
  if (!import.meta.client) return
  const el = document.getElementById('ozellikler')
  if (!el) return
  const top = Math.max(0, el.getBoundingClientRect().top + window.scrollY - 72)
  window.scrollTo({ top, behavior: 'smooth' })
}

const highlights = [
  { label: 'Özel Alias / Base62', desc: 'Kendi takma adın veya 6 haneli kod' },
  { label: 'Argon2id & /w Modu', desc: 'Parola kilidi veya süreli bekletme' },
  { label: 'Canlı Log & Grafik', desc: 'Cihaz, İS, tarayıcı ve ülke analizi' }
]
</script>

<template>
  <section class="relative overflow-hidden border-b-[1.5px] border-ink py-14 md:py-24">
    <div class="container-x grid items-center gap-12 lg:grid-cols-[1.08fr_0.92fr] lg:gap-14">
      <div>
        <div class="mb-6 inline-flex items-center gap-2 rounded-sm border-[1.5px] border-ink bg-card px-3 py-1 shadow-[2px_2px_0_var(--color-ink)]">
          <span class="h-2 w-2 rounded-full bg-ok" />
          <span class="font-mono text-xs font-semibold uppercase tracking-wider text-ink">
            Redis Hızında 302 Yönlendirme · Detaylı Tıklama Analitiği
          </span>
        </div>

        <h1 class="text-[clamp(2.65rem,6.5vw,5rem)] font-extrabold leading-[1.02]">
          Uzun bağlantıları<br />
          <span class="relative inline-block">
            kontrollü kısa linklere
            <span class="absolute inset-x-0 -bottom-1 -z-10 h-[0.22em] bg-accent/85" aria-hidden="true" />
          </span>
          çevir.
        </h1>

        <p class="mt-6 max-w-xl text-base leading-relaxed text-ink-2 sm:text-lg">
          <strong class="font-semibold text-ink">makeiTiny</strong> ile uzun adresleri saniyeler içinde kısalt.
          İster özel takma ad belirle, ister <strong class="font-semibold text-ink">Argon2id şifre koruması</strong> veya
          <strong class="font-semibold text-ink">geri sayımlı bekletme sayfası (/w)</strong> ekle; her tıklamayı cihaz, tarayıcı, zaman grafiği ve coğrafi dağılımıyla izle.
        </p>

        <div class="mt-8 flex flex-wrap items-center gap-3.5">
          <template v-if="isLoggedIn">
            <UiButton to="/dashboard">
              Panele git <UiIcon name="arrow" :size="16" />
            </UiButton>
            <UiButton to="/dashboard/new" variant="secondary">
              <UiIcon name="plus" :size="16" /> Yeni link kısalt
            </UiButton>
          </template>
          <template v-else>
            <UiButton to="/auth/register">
              Ücretsiz hesap aç <UiIcon name="arrow" :size="16" />
            </UiButton>
            <UiButton to="/auth/login" variant="secondary">Giriş yap</UiButton>
          </template>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3 py-2 text-sm font-semibold text-ink-2 underline decoration-line decoration-2 underline-offset-4 transition-colors hover:text-ink hover:decoration-accent cursor-pointer"
            @click="scrollToFeatures"
          >
            Özellikleri incele ↓
          </button>
        </div>

        <div class="mt-10 grid gap-3 border-t border-line pt-6 sm:grid-cols-3">
          <div
            v-for="item in highlights"
            :key="item.label"
            class="rounded-md border border-line bg-card/80 px-3.5 py-2.5"
          >
            <p class="font-mono text-xs font-bold text-ink">{{ item.label }}</p>
            <p class="mt-0.5 text-xs text-ink-3">{{ item.desc }}</p>
          </div>
        </div>
      </div>

      <LandingHeroDemo />
    </div>
  </section>
</template>
