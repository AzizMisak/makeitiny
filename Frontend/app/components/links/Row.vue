<script setup lang="ts">
import type { MenuItem } from '~/components/ui/Menu.vue'
import type { LinkItem } from '~/types/link'

const props = defineProps<{ link: LinkItem }>()
const emit = defineEmits<{ remove: [link: LinkItem] }>()

const { prefix, full } = useShortUrl()
const { copied, copy } = useCopy()

const url = computed(() => full(props.link.alias))
const state = computed(() => linkState(props.link))

const traits = computed(() => {
  const l = props.link
  const list: { icon: string; label: string }[] = []
  if (l.locked) list.push({ icon: 'lock', label: 'Parolalı' })
  else if (l.waitEnabled) list.push({ icon: 'clock', label: `${l.waitSeconds} sn bekletmeli` })
  if (l.hasQr) list.push({ icon: 'qr', label: 'QR kodu var' })
  if (l.adEnabled) list.push({ icon: 'ad', label: 'Reklam açık' })
  if (l.expiresAt) list.push({ icon: 'calendar', label: `${formatDateTr(l.expiresAt)} tarihinde biter` })
  if (l.maxClicks) list.push({ icon: 'target', label: `${formatNumber(l.maxClicks)} tıklama sınırı` })
  return list
})

const menu = computed<MenuItem[]>(() => [
  { label: 'Linki yeni sekmede aç', icon: 'external', href: url.value },
  { label: 'Ayarları düzenle', icon: 'edit', to: `/dashboard/links/${props.link.id}?tab=ayarlar` },
  { label: 'Linki sil', icon: 'trash', tone: 'danger', divider: true, onSelect: () => emit('remove', props.link) }
])
</script>

<template>
  <li class="group relative grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-1 py-4 md:grid-cols-[minmax(0,1fr)_7rem_6.5rem_auto]">
    <div class="col-start-1 row-start-1 min-w-0">
      <NuxtLink
        :to="`/dashboard/links/${link.id}`"
        class="font-mono text-[15px] font-semibold after:absolute after:inset-0 after:content-[''] group-hover:text-brand"
      >
        <span class="font-normal text-ink-3">{{ prefix }}</span>{{ link.alias }}
      </NuxtLink>
      <p class="truncate text-sm text-ink-3" :title="link.target">
        <span v-if="link.title" class="font-medium text-ink-2">{{ link.title }}, </span>{{ link.target }}
      </p>
    </div>

    <div class="hidden md:block">
      <UiStatus :tone="state.tone" :label="state.label" />
      <p class="mt-1 flex h-4 gap-2 text-ink-3">
        <span v-for="t in traits" :key="t.icon" class="relative z-10" :title="t.label">
          <UiIcon :name="t.icon" :size="15" /><span class="sr-only">{{ t.label }}</span>
        </span>
      </p>
    </div>

    <p class="hidden text-right md:block">
      <span class="tnum block text-lg font-semibold leading-tight">{{ formatNumber(link.clicks) }}</span>
      <span class="text-[13px] text-ink-3">tıklama</span>
    </p>

    <div class="relative z-10 col-start-2 row-start-1 flex items-center md:col-start-4">
      <UiIconButton
        :icon="copied === url ? 'check' : 'copy'"
        :tone="copied === url ? 'ok' : 'normal'"
        :label="copied === url ? 'Kopyalandı' : 'Kısa adresi kopyala'"
        @click="copy(url)"
      />
      <UiMenu :label="`${link.alias} işlemleri`" :items="menu" />
    </div>

    <div class="col-span-2 flex items-center gap-4 text-sm md:hidden">
      <UiStatus :tone="state.tone" :label="state.label" />
      <span class="tnum text-ink-2">{{ formatNumber(link.clicks) }} tıklama</span>
    </div>
  </li>
</template>
