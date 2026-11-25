<script setup lang="ts">
import type { VisitorReason } from '~/utils/visitorErrors'

definePageMeta({ layout: 'visitor' })

const route = useRoute()
const alias = computed(() => String(route.query.alias ?? route.query.link ?? ''))
const info = computed(() => VISITOR_ERRORS[String(route.query.reason ?? 'not_found') as VisitorReason] ?? VISITOR_ERRORS.internal)

useHead({ title: () => `${info.value.title} | makeiTiny` })

onMounted(async () => {
  if (!alias.value) return
  await navigateTo(
    { path: `/error/${alias.value}`, query: route.query.reason ? { reason: String(route.query.reason) } : undefined },
    { replace: true }
  )
})
</script>

<template>
  <VisitorNotice :title="info.title" :text="info.text" />
</template>
