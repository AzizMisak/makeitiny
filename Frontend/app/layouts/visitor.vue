<template>
  <div class="flex min-h-dvh flex-col">
    <header class="flex items-center justify-between px-5 py-4 md:px-8">
      <BrandLogo />
      <UiThemeToggle />
    </header>
    <main class="flex flex-1 flex-col">
      <slot />
    </main>
  </div>
</template>
