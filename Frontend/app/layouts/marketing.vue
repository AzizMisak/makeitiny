<template>
  <div class="flex min-h-screen flex-col">
    <LandingHeader />
    <main class="flex-1">
      <slot />
    </main>
    <LandingFooter />
  </div>
</template>
