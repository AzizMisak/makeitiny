<template>
  <section class="container-x grid items-center gap-14 py-16 md:py-24 lg:grid-cols-[1.05fr_0.95fr]">
    <div>
      <p class="eyebrow mb-5">Link kısaltıcı · sade ve hızlı</p>
      <h1 class="text-[clamp(2.75rem,7vw,5.25rem)] font-extrabold">
        Uzun linkler.<br />
        <span class="relative inline-block">
          Küçük adresler.
          <span class="absolute inset-x-0 -bottom-1 h-[0.18em] bg-accent" aria-hidden="true" />
        </span>
      </h1>
      <p class="mt-7 max-w-lg text-lg text-ink-2">
        makeiTiny, uzun bağlantıları paylaşması kolay kısa adreslere çevirir. Kendi takma adını seç,
        şifre koy, kaç kişinin tıkladığını gör.
      </p>
      <div class="mt-9 flex flex-wrap items-center gap-3">
        <UiButton to="/auth/register">Ücretsiz hesap aç</UiButton>
        <UiButton to="/auth/login" variant="secondary">Giriş yap</UiButton>
      </div>
      <p class="mt-5 text-sm text-ink-3">Kredi kartı gerekmez. Dakikalar içinde ilk linkini kısalt.</p>
    </div>

    <LandingHeroDemo />
  </section>
</template>
