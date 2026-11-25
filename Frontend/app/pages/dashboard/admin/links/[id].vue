<script setup lang="ts">
import type { AdminLinkItem } from '~/types/link'

definePageMeta({ layout: 'admin', middleware: 'auth', roles: ['admin', 'owner'] })

const route = useRoute()
const { fetchAdminLinkById, fetchAdminLinkLogs } = useAdmin()
const { prefix, full } = useShortUrl()

const id = computed(() => String(route.params.id ?? ''))
const link = ref<AdminLinkItem | null>(null)
const loading = ref(true)
const tab = ref<'ziyaretler' | 'duzenle'>('ziyaretler')
const dirty = ref(false)

useUnsavedGuard(() => dirty.value)
useHead({ title: () => `${link.value?.alias ?? 'Link'} | Yönetim | makeiTiny` })

async function load() {
  loading.value = true
  link.value = await fetchAdminLinkById(id.value)
  loading.value = false
}

onMounted(load)

const parts = computed(() => splitUrl(link.value?.target ?? ''))

const facts = computed(() => {
  const l = link.value
  if (!l) return []
  return [
    { label: 'Tıklama', value: l.maxClicks ? `${formatNumber(l.clicks)} / ${formatNumber(l.maxClicks)}` : formatNumber(l.clicks) },
    { label: 'Erişim', value: l.locked ? 'Parolalı' : l.waitEnabled ? `${l.waitSeconds} sn bekletmeli` : 'Herkese açık' },
    { label: 'Oluşturulma', value: formatDateTr(l.createdAt) },
    { label: 'Bitiş', value: l.expiresAt ? formatDateTr(l.expiresAt) : 'Süresiz' },
    { label: 'Kimlik', value: l.id }
  ]
})

const tabs = computed(() => [
  { value: 'ziyaretler' as const, label: 'Ziyaretler', icon: 'eye' },
  { value: 'duzenle' as const, label: 'Linki düzenle', icon: 'edit', dot: dirty.value }
])
</script>

<template>
  <div>
    <NuxtLink to="/dashboard/admin/links" class="mb-5 inline-flex items-center gap-2 text-sm font-medium text-ink-2 hover:text-ink">
      <UiIcon name="back" :size="16" /> Linkler
    </NuxtLink>

    <div v-if="loading" class="grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]" aria-busy="true" aria-label="Link yükleniyor">
      <div><div class="skeleton h-8 w-72" /><div class="skeleton mt-5 h-28" /><div class="skeleton mt-8 h-72" /></div>
      <div class="skeleton h-[30rem]" />
    </div>

    <UiEmpty v-else-if="!link" icon="search" title="Link bulunamadı" text="Silinmiş olabilir ya da adres yanlış." class="border border-line bg-surface">
      <UiButton variant="outline" icon="refresh" @click="load">Tekrar dene</UiButton>
      <UiButton to="/dashboard/admin/links" variant="quiet">Linklere dön</UiButton>
    </UiEmpty>

    <div v-else class="grid grid-cols-1 items-start gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
      <div class="min-w-0">
        <header class="flex flex-wrap items-center justify-between gap-3">
          <h1 class="min-w-0 break-all font-mono text-2xl font-semibold">
            <span class="font-normal text-ink-3">{{ prefix }}</span>{{ link.alias }}
          </h1>
          <UiButton :href="full(link.alias)" target="_blank" rel="noopener" variant="outline" size="sm" icon="external">Kısa linki aç</UiButton>
        </header>

        <!-- Kararın kanıtı: hedef adres, alan adı öne çıkarılmış. -->
        <section class="mt-5 border-l-4 border-mark bg-surface px-5 py-4">
          <p class="text-[13px] text-ink-3">Hedef</p>
          <p class="mt-1 break-all font-mono text-sm text-ink-3">
            {{ parts.protocol }}<strong class="text-lg text-ink">{{ parts.host }}</strong>{{ parts.rest }}
          </p>
          <p v-if="link.title" class="mt-2 text-sm text-ink-2">Başlık: {{ link.title }}</p>
        </section>

        <dl class="mt-6 flex flex-wrap gap-x-10 gap-y-4">
          <div>
            <dt class="text-[13px] text-ink-3">Sahibi</dt>
            <dd class="mt-0.5 font-semibold">
              <NuxtLink
                v-if="link.userId && link.userId !== '0'"
                :to="`/dashboard/admin/users/${link.userId}`"
                class="text-brand-ink underline underline-offset-4"
                :title="link.userEmail"
              >{{ link.userName || link.userEmail }}</NuxtLink>
              <template v-else>Üyeliksiz</template>
            </dd>
          </div>
          <div v-for="f in facts" :key="f.label">
            <dt class="text-[13px] text-ink-3">{{ f.label }}</dt>
            <dd class="tnum mt-0.5 font-semibold">{{ f.value }}</dd>
          </div>
        </dl>

        <UiTabs v-model="tab" :tabs="tabs" label="Link bölümleri" class="mt-8" />
        <div :id="`panel-${tab}`" role="tabpanel" :aria-labelledby="`tab-${tab}`" class="pt-6">
          <LinksVisits v-show="tab === 'ziyaretler'" :fetch-page="(p: number, l: number) => fetchAdminLinkLogs(id, p, l)" />
          <AdminLinkEditor
            v-show="tab === 'duzenle'"
            v-model="link"
            @dirty="dirty = $event"
            @removed="navigateTo('/dashboard/admin/links')"
          />
        </div>
      </div>

      <AdminDecisionPanel v-model="link" class="lg:sticky lg:top-24" />
    </div>
  </div>
</template>
