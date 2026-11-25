<template>
  <NuxtLoadingIndicator color="var(--color-mark)" :height="3" :throttle="80" />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <UiConfirmHost />
  <UiToastHost />
  <PanelSpaceCurtain />
</template>
