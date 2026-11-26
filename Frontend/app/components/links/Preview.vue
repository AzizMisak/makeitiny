<script setup lang="ts">
import type { LinkDraft } from '~/utils/linkDraft'

/** Yazılan linkin gerçek ölçekli hâli: iki şeridin boyu karakter sayılarıyla orantılıdır. */
const props = defineProps<{ draft: LinkDraft }>()
const { display } = useShortUrl()

const target = computed(() => normalizeTargetUrl(props.draft.target))
const alias = computed(() => props.draft.alias.trim() || '······')
const short = computed(() => display(alias.value))
const scale = computed(() => Math.max(target.value.length, short.value.length, 32))
const shortWidth = computed(() => (short.value.length / scale.value) * 100)

const rules = computed(() => {
  const list: string[] = []
  if (props.draft.expires && props.draft.expiresAt) list.push(`${formatDateTr(props.draft.expiresAt)} sonunda kapanır.`)
  if (props.draft.limited && Number(props.draft.maxClicks) > 0) list.push(`${formatNumber(Number(props.draft.maxClicks))} tıklamada kapanır.`)
  if (props.draft.adEnabled && props.draft.access === 'wait') list.push('Bekleme sayfasında reklam alanı açık.')
  return list
})
</script>

<template>
  <div class="mat-grid on-mat p-6 text-on-mat">
    <p class="tnum flex justify-between text-sm text-on-mat-2">
      <span>Uzun hâli</span>
      <span>{{ target.length }} karakter</span>
    </p>
    <div class="mt-2 h-10 overflow-hidden bg-white/90 px-3 font-mono text-[13px] leading-10 text-[#3f5249]">
      <span class="block truncate">{{ target || 'Hedef adresi yazınca burada görünür' }}</span>
    </div>

    <p class="tnum mt-5 flex justify-between text-sm text-on-mat-2">
      <span>Kısa hâli</span>
      <span>{{ short.length }} karakter</span>
    </p>
    <div
      class="strip-cut mt-2 h-12 min-w-fit max-w-full bg-mark pl-3 pr-7 font-mono text-sm font-semibold leading-[3rem] text-on-mark transition-[width] duration-300"
      :style="{ width: `${shortWidth}%` }"
    >
      <span class="block truncate">{{ short }}</span>
    </div>

    <p class="mt-7 leading-relaxed">{{ visitorStory(draft) }}</p>
    <!-- En fazla üç kural satırı; yeri baştan ayrılmış. -->
    <ul class="mt-2 min-h-[4.5rem] text-sm leading-6 text-on-mat-2">
      <li v-for="r in rules" :key="r">{{ r }}</li>
    </ul>

    <slot />
  </div>
</template>
