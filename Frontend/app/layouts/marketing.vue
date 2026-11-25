<template>
  <div class="flex min-h-dvh flex-col">
    <LandingHeader />
    <main class="flex-1">
      <slot />
    </main>
    <LandingFooter />
    <LandingTape />
  </div>
</template>
