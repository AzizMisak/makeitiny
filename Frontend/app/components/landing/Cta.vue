<template>
  <section class="container-x pb-20">
    <div class="flex flex-col items-start justify-between gap-8 rounded-lg bg-ink p-8 text-paper md:flex-row md:items-center md:p-12">
      <div>
        <h2 class="text-3xl font-bold md:text-4xl">İlk linkini şimdi kısalt.</h2>
        <p class="mt-3 max-w-md text-paper/70">Hesap açmak otuz saniye sürer. Sonrası zaten kısa.</p>
      </div>
      <UiButton to="/auth/register" variant="primary">Ücretsiz başla</UiButton>
    </div>
  </section>
</template>
