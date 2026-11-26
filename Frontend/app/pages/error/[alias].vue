<script setup lang="ts">
import type { VisitorReason } from '~/utils/visitorErrors'

definePageMeta({ layout: 'visitor' })

const route = useRoute()
const { full } = useShortUrl()

const alias = computed(() => String(route.params.alias ?? ''))
const info = computed(() => VISITOR_ERRORS[String(route.query.reason ?? 'not_found') as VisitorReason] ?? VISITOR_ERRORS.internal)
const retrying = ref(false)

useHead({ title: () => `${info.value.title} | makeiTiny` })

function retry() {
  retrying.value = true
  window.location.href = full(alias.value)
}
</script>

<template>
  <VisitorNotice :title="info.title" :text="info.text" :alias="alias">
    <UiButton v-if="info.canRetry" icon="refresh" :loading="retrying" loading-label="Deneniyor" @click="retry">Tekrar dene</UiButton>
  </VisitorNotice>
</template>
