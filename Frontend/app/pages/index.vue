<script setup lang="ts">
definePageMeta({ layout: 'marketing' })

const { isLoggedIn } = useAuth()
</script>

<template>
  <LandingHero />
  <LandingAnonymousShortener v-if="!isLoggedIn" />
  <LandingFeatures />
  <LandingSteps />
  <LandingCta />
</template>
