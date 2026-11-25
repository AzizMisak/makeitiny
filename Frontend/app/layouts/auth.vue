<script setup lang="ts">
const LONG = 'https://magaza.ornek.com/yaz-koleksiyonu/2026/erkek/ayakkabi?kaynak=bulten&kampanya=agustos&ref=eposta'
</script>

<template>
  <div class="grid grid-cols-1 min-h-dvh lg:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)]">
    <div class="flex flex-col">
      <header class="flex items-center justify-between px-5 py-4 md:px-10">
        <BrandLogo />
        <UiThemeToggle />
      </header>
      <main class="flex flex-1 items-center px-5 py-10 md:px-10">
        <div class="mx-auto w-full max-w-sm">
          <slot />
        </div>
      </main>
    </div>

    <aside class="mat-grid relative hidden overflow-hidden text-on-mat lg:flex lg:flex-col lg:justify-center" aria-hidden="true">
      <div class="flex flex-col gap-10 pl-16">
        <div class="relative w-[140%] bg-white/90 py-4 pl-6 font-mono text-sm text-[#3f5249]">
          <span class="whitespace-nowrap">{{ LONG }}</span>
        </div>
        <UiIcon name="scissors" :size="44" class="text-mark" />
        <div class="strip-cut w-fit bg-mark py-4 pl-6 pr-12 font-mono text-xl font-semibold text-on-mark">localhost/r/yaz26</div>
      </div>
    </aside>
  </div>
</template>
