<template>
  <footer class="on-mat bg-mat-2 text-on-mat">
    <div class="container-x flex flex-wrap items-center justify-between gap-6 py-10 lg:pr-20">
      <BrandLogo tone="mat" />
      <nav class="flex flex-wrap gap-x-7 gap-y-2 text-[15px] text-on-mat-2" aria-label="Alt menü">
        <a v-for="s in landingSections" :key="s.id" :href="`#${s.id}`" class="underline-offset-4 hover:text-on-mat hover:underline">{{ s.label }}</a>
      </nav>
      <p class="text-sm text-on-mat-2">© 2026 makeiTiny</p>
    </div>
  </footer>
</template>
