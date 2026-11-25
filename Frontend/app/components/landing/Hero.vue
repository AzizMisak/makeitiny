<template>
  <section id="kisalt" class="mat-grid on-mat overflow-hidden text-on-mat">
    <div class="container-x pb-16 pt-10 md:pb-24 md:pt-16 lg:pr-20">
      <!-- Aynı punto, iki uç genişlik: cümle söylediğini yapar. -->
      <h1 class="display text-[clamp(3.1rem,11.2vw,8.75rem)] leading-[0.92]">
        <span class="block whitespace-nowrap" style="font-stretch: 130%">Uzun linki</span>
        <span class="headline-tiny block whitespace-nowrap">kısa kes.</span>
      </h1>
      <div class="mt-12 md:mt-16">
        <LandingCutter />
      </div>
    </div>
  </section>
</template>

<style scoped>
.headline-tiny {
  font-stretch: 50%;
  animation: shrink 900ms cubic-bezier(0.7, 0, 0.15, 1) 250ms backwards;
}

@keyframes shrink {
  from {
    font-stretch: 130%;
  }
}
</style>
