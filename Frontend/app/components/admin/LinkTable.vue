<script setup lang="ts">
import type { MenuItem } from '~/components/ui/Menu.vue'
import type { AdminLinkItem } from '~/types/link'

const props = defineProps<{ links: AdminLinkItem[]; loading?: boolean; hideOwner?: boolean; actions?: boolean; caption: string }>()
const emit = defineEmits<{ changed: [] }>()

const { prefix, full } = useShortUrl()
const { setActive } = useLinkModeration()

function ownerLabel(l: AdminLinkItem): string {
  return l.isAnonymous ? 'Üyeliksiz' : l.userName || l.userEmail || '—'
}

async function toggle(l: AdminLinkItem) {
  if (await setActive(l, !l.isActive)) emit('changed')
}

function menuFor(l: AdminLinkItem): MenuItem[] {
  return [
    { label: 'Linki yeni sekmede aç', icon: 'external', href: full(l.alias) },
    ...(l.userId && l.userId !== '0' && !props.hideOwner
      ? [{ label: 'Sahibini aç', icon: 'user', to: `/dashboard/admin/users/${l.userId}` }]
      : []),
    {
      label: l.isActive ? 'Yönlendirmeyi kapat' : 'Yönlendirmeyi aç',
      icon: l.isActive ? 'ban' : 'check',
      tone: l.isActive ? 'danger' : 'normal',
      divider: true,
      onSelect: () => toggle(l)
    }
  ]
}
</script>

<template>
  <div class="overflow-x-auto border border-line bg-surface transition-opacity" :class="loading ? 'pointer-events-none opacity-55' : ''" :aria-busy="loading">
    <table class="w-full min-w-[56rem] text-left text-sm">
      <caption class="sr-only">{{ caption }}</caption>
      <thead class="bg-sunken text-ink-2">
        <tr>
          <th class="px-4 py-2.5 font-medium">Link ve hedef</th>
          <th v-if="!hideOwner" class="px-4 py-2.5 font-medium">Sahibi</th>
          <th class="px-4 py-2.5 font-medium">Yönlendirme</th>
          <th class="px-4 py-2.5 font-medium">İnceleme</th>
          <th class="px-4 py-2.5 text-right font-medium">Tıklama</th>
          <th class="px-4 py-2.5 font-medium">Oluşturulma</th>
          <th v-if="actions" class="w-12"><span class="sr-only">İşlemler</span></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-line">
        <tr v-for="l in links" :key="l.id" class="relative hover:bg-brand-soft/40">
          <td class="max-w-md px-4 py-3">
            <NuxtLink
              :to="`/dashboard/admin/links/${l.id}`"
              class="font-mono text-sm font-semibold text-ink after:absolute after:inset-0 after:content-[''] hover:text-brand"
            >
              <span class="font-normal text-ink-3">{{ prefix }}</span>{{ l.alias }}
            </NuxtLink>
            <p class="mt-0.5 truncate text-ink-3" :title="l.target">{{ l.title ? `${l.title}, ` : '' }}{{ l.target }}</p>
          </td>
          <td v-if="!hideOwner" class="max-w-52 px-4 py-3">
            <p class="truncate">{{ ownerLabel(l) }}</p>
            <p v-if="l.userEmail && l.userName" class="truncate text-ink-3">{{ l.userEmail }}</p>
          </td>
          <td class="px-4 py-3"><UiStatus :tone="l.isActive ? 'ok' : 'neutral'" :label="l.isActive ? 'Açık' : 'Kapalı'" /></td>
          <td class="px-4 py-3">
            <UiTag :tone="securityMeta(l.securityStatus).tone" :icon="securityMeta(l.securityStatus).icon">
              {{ securityMeta(l.securityStatus).label }}
            </UiTag>
          </td>
          <td class="tnum px-4 py-3 text-right font-semibold">
            {{ formatNumber(l.clicks) }}<span v-if="l.maxClicks" class="font-normal text-ink-3"> / {{ formatNumber(l.maxClicks) }}</span>
          </td>
          <td class="tnum whitespace-nowrap px-4 py-3 text-ink-2">
            {{ formatDateTr(l.createdAt) }}
            <p v-if="l.expiresAt" class="text-ink-3">{{ formatDateTr(l.expiresAt) }} biter</p>
          </td>
          <td v-if="actions" class="relative z-10 px-2">
            <UiMenu :label="`${l.alias} işlemleri`" :items="menuFor(l)" />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
