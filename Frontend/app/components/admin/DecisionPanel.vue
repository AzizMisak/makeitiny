<script setup lang="ts">
import type { AdminLinkItem, SecurityStatus } from '~/types/link'

/**
 * İnceleme kararı. Karar (güvenlik durumu) ve yönlendirme ayrı ayrı görünür;
 * engelleyici bir karar seçilince yönlendirme anahtarı da kapanır ve kullanıcı bunu görür.
 */
const link = defineModel<AdminLinkItem>({ required: true })

const { adminUpdateLink } = useAdmin()
const { ask } = useConfirm()
const { display } = useShortUrl()
const toast = useToast()

const decision = ref<SecurityStatus>(link.value.securityStatus)
const active = ref(link.value.isActive)

watch(link, (l) => {
  decision.value = l.securityStatus
  active.value = l.isActive
})

function choose(value: SecurityStatus) {
  const wasBlocking = isBlockedStatus(decision.value)
  decision.value = value
  if (isBlockedStatus(value)) active.value = false
  else if (wasBlocking) active.value = true
}

const changes = computed(() => {
  const list = []
  if (decision.value !== link.value.securityStatus) {
    list.push({
      label: 'Karar',
      from: securityMeta(link.value.securityStatus).label,
      to: securityMeta(decision.value).label
    })
  }
  if (active.value !== link.value.isActive) {
    list.push({
      label: 'Yönlendirme',
      from: link.value.isActive ? 'Açık' : 'Kapalı',
      to: active.value ? 'Açık' : 'Kapalı'
    })
  }
  return list
})

const outcome = computed(() => {
  if (isBlockedStatus(decision.value)) return 'Ziyaretçi güvenlik uyarısı görür, hedefe gidemez.'
  if (!active.value) return 'Ziyaretçi "link devre dışı" sayfasını görür.'
  return 'Ziyaretçi hedefe ulaşır.'
})

const stops = computed(() => !active.value || isBlockedStatus(decision.value))

async function apply() {
  const ok = await ask({
    title: 'Karar uygulansın mı?',
    subject: `${display(link.value.alias)} → ${link.value.target}`,
    message: outcome.value,
    changes: changes.value,
    confirmLabel: 'Kararı uygula',
    tone: stops.value ? 'danger' : 'normal',
    action: async () => {
      const updated = await adminUpdateLink(link.value.id, { securityStatus: decision.value, isActive: active.value })
      link.value = { ...link.value, ...updated }
    }
  })
  if (ok) toast.success('Karar uygulandı.')
}
</script>

<template>
  <section class="border border-line bg-surface" aria-labelledby="decision-title">
    <h2 id="decision-title" class="display px-5 pt-5 text-xl">İnceleme kararı</h2>

    <div role="radiogroup" aria-labelledby="decision-title" class="mt-3">
      <label
        v-for="o in SECURITY_OPTIONS"
        :key="o.value"
        class="flex cursor-pointer gap-3.5 px-5 py-3 has-focus-visible:outline-2 has-focus-visible:-outline-offset-2 has-focus-visible:outline-brand"
        :class="decision === o.value ? 'bg-brand-soft' : 'hover:bg-sunken'"
      >
        <input
          type="radio"
          name="decision"
          class="sr-only"
          :value="o.value"
          :checked="decision === o.value"
          @change="choose(o.value)"
        />
        <span
          class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2"
          :class="decision === o.value ? 'border-brand' : 'border-line-2'"
          aria-hidden="true"
        >
          <span v-if="decision === o.value" class="h-2.5 w-2.5 rounded-full bg-brand" />
        </span>
        <span class="min-w-0">
          <span class="flex flex-wrap items-center gap-2 font-semibold">
            {{ o.label }}
            <span v-if="o.value === link.securityStatus" class="text-[13px] font-normal text-ink-3">şu anki</span>
          </span>
          <span class="block text-sm text-ink-2">{{ o.effect }}</span>
        </span>
      </label>
    </div>

    <div class="flex items-center justify-between gap-4 border-t border-line px-5 py-4">
      <div>
        <p class="font-semibold">Yönlendirme</p>
        <p class="text-sm text-ink-2">
          {{ isBlockedStatus(decision) ? 'Engelli veya zararlı link yayına alınamaz.' : active ? 'Açık' : 'Kapalı' }}
        </p>
      </div>
      <UiSwitch v-model="active" :disabled="isBlockedStatus(decision)" label="Yönlendirme açık" />
    </div>

    <div class="border-t border-line px-5 py-4">
      <p class="min-h-10 text-sm leading-5" :class="stops ? 'text-danger' : 'text-ok'">{{ outcome }}</p>
      <UiButton block class="mt-3" :disabled="changes.length === 0" @click="apply">
        {{ changes.length === 0 ? 'Değişiklik yok' : 'Kararı uygula' }}
      </UiButton>
    </div>
  </section>
</template>
