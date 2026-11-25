<script setup lang="ts">
import type { LinkDraft, LinkDraftErrors } from '~/utils/linkDraft'
import type { LinkItem } from '~/types/link'

const props = defineProps<{ errors: LinkDraftErrors; original?: LinkItem | null; disabled?: boolean }>()
const draft = defineModel<LinkDraft>({ required: true })

const { canUseFeature } = useAuth()
const { checkAlias } = useLinks()
const { prefix } = useShortUrl()

const canAlias = computed(() => canUseFeature('allowCustomAlias'))
const canPassword = computed(() => canUseFeature('allowPassword') || !!props.original?.locked)
const canAds = computed(() => canUseFeature('allowAds'))

// Kısa ad yazıldıkça müsaitlik kendiliğinden denetlenir; ayrı bir "kontrol et" düğmesi yok.
const aliasState = ref<'idle' | 'checking' | 'free' | 'taken' | 'failed'>('idle')
let timer: ReturnType<typeof setTimeout> | null = null
let ticket = 0

watch(
  () => draft.value.alias,
  (raw) => {
    if (timer) clearTimeout(timer)
    const alias = raw.trim()
    const mine = ++ticket
    aliasState.value = 'idle'
    if (!canAlias.value || alias === props.original?.alias || !ALIAS_PATTERN.test(alias)) return
    aliasState.value = 'checking'
    timer = setTimeout(async () => {
      try {
        const res = await checkAlias(alias)
        if (mine === ticket) aliasState.value = res.available ? 'free' : 'taken'
      } catch {
        if (mine === ticket) aliasState.value = 'failed'
      }
    }, 450)
  }
)

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
})

const aliasError = computed(() => {
  if (props.errors.alias) return props.errors.alias
  if (aliasState.value === 'taken') return 'Bu kısa ad alınmış, başka bir ad dene.'
  if (aliasState.value === 'failed') return 'Müsaitlik denetlenemedi.'
  const alias = draft.value.alias.trim()
  if (alias && !ALIAS_PATTERN.test(alias) && alias.length >= 3) return '3–32 karakter: harf, rakam, tire veya alt çizgi.'
  return ''
})

const aliasHint = computed(() => {
  if (!canAlias.value) return props.original ? 'Kısa adı değiştirmek paketinde kapalı.' : 'Paketinde kapalı, rastgele 6 karakter verilir.'
  if (aliasState.value === 'checking') return 'Müsait mi diye bakılıyor…'
  return props.original ? 'Değişirse eski adres çalışmaz.' : 'Boş bırakırsan rastgele 6 karakter verilir.'
})

const accessOptions = computed(() => [
  { value: 'open' as const, label: 'Herkese açık', icon: 'globe' },
  { value: 'password' as const, label: 'Parolalı', icon: 'lock', disabled: !canPassword.value },
  { value: 'wait' as const, label: 'Bekletmeli', icon: 'clock' }
])

defineExpose({ aliasState })
</script>

<template>
  <fieldset :disabled="disabled" class="contents">
    <section class="flex flex-col gap-3">
      <UiField
        id="link-target"
        v-model="draft.target"
        label="Hedef adres"
        placeholder="ornek.com/uzun/bir/sayfa"
        inputmode="url"
        large
        hint="https:// yazmasan da eklenir."
        :error="errors.target"
      />
      <div class="grid grid-cols-1 gap-x-5 gap-y-3 sm:grid-cols-2">
        <UiField
          id="link-alias"
          v-model="draft.alias"
          label="Kısa ad"
          mono
          :prefix="prefix"
          :placeholder="canAlias ? 'yaz-kampanyasi' : ''"
          :maxlength="32"
          :disabled="!canAlias"
          :hint="aliasHint"
          :error="aliasError"
          :ok="aliasState === 'free' ? 'Müsait' : ''"
          :busy="aliasState === 'checking'"
        >
          <template #label-end><UiTag v-if="!canAlias" tone="mark" icon="lock">Pro</UiTag></template>
        </UiField>
        <UiField
          id="link-title"
          v-model="draft.title"
          label="Başlık"
          placeholder="Yaz kampanyası"
          hint="Yalnızca sen görürsün. Boş kalabilir."
        />
      </div>
    </section>

    <section class="border-t border-line pt-6">
      <h3 class="font-semibold">Ziyaretçi nasıl geçsin?</h3>
      <div class="mt-3">
        <UiSegmented v-model="draft.access" label="Erişim türü" :options="accessOptions" />
      </div>
      <!-- Üç seçenek de aynı yüksekliği kullanır; geçişte sayfa oynamaz. -->
      <div class="mt-4 min-h-[4.75rem]">
        <p v-if="draft.access === 'open'" class="max-w-lg pt-2 text-ink-2">
          Linke tıklayan herkes beklemeden hedef adrese gider.
        </p>
        <UiField
          v-else-if="draft.access === 'password'"
          id="link-password"
          v-model="draft.password"
          type="password"
          class="max-w-sm"
          autocomplete="new-password"
          :placeholder="original?.locked ? 'Mevcut parola geçerli' : 'Link parolası'"
          :hint="original?.locked ? 'Değiştirmek istersen yeni parolayı yaz.' : 'Hedef, doğru parola girilene kadar gizli kalır.'"
          :error="errors.password"
        />
        <UiField
          v-else
          id="link-wait"
          v-model="draft.waitSeconds"
          type="number"
          class="max-w-40"
          min="1"
          suffix="saniye"
          inputmode="numeric"
          :error="errors.waitSeconds"
          hint="Geri sayım süresi"
        />
      </div>
    </section>

    <section class="border-t border-line pt-6">
      <h3 class="font-semibold">Sınırlar</h3>
      <ul class="mt-1 divide-y divide-line">
        <li class="flex flex-wrap items-center gap-x-4 gap-y-3 py-4">
          <UiSwitch v-model="draft.expires" label="Son kullanma tarihi" />
          <div class="min-w-48 flex-1">
            <p class="font-medium">Son kullanma tarihi</p>
            <p class="text-sm" :class="errors.expiresAt ? 'text-danger' : 'text-ink-3'">
              {{ errors.expiresAt || 'Bu günün sonunda link kapanır.' }}
            </p>
          </div>
          <UiField
            id="link-expires"
            v-model="draft.expiresAt"
            bare
            type="date"
            class="w-44"
            :min="original ? undefined : todayIso()"
            :disabled="!draft.expires"
            :error="errors.expiresAt"
          />
        </li>
        <li class="flex flex-wrap items-center gap-x-4 gap-y-3 py-4">
          <UiSwitch v-model="draft.limited" label="Tıklama sınırı" />
          <div class="min-w-48 flex-1">
            <p class="font-medium">Tıklama sınırı</p>
            <p class="text-sm" :class="errors.maxClicks ? 'text-danger' : 'text-ink-3'">
              {{ errors.maxClicks || 'Bu sayıya ulaşınca link kapanır.' }}
            </p>
          </div>
          <UiField
            id="link-max"
            v-model="draft.maxClicks"
            bare
            type="number"
            min="1"
            inputmode="numeric"
            class="w-44"
            suffix="tıklama"
            :disabled="!draft.limited"
            :error="errors.maxClicks"
          />
        </li>
        <li class="flex flex-wrap items-center gap-x-4 gap-y-3 py-4">
          <UiSwitch v-model="draft.adEnabled" label="Reklam gösterimi" :disabled="!canAds" />
          <div class="min-w-48 flex-1">
            <p class="flex items-center gap-2 font-medium">
              Reklam gösterimi
              <UiTag v-if="!canAds" tone="mark" icon="lock">Pro</UiTag>
            </p>
            <p class="text-sm text-ink-3">Ara sayfada reklam alanı açılır.</p>
          </div>
        </li>
        <li v-if="original" class="flex flex-wrap items-center gap-x-4 gap-y-3 py-4">
          <UiSwitch v-model="draft.isActive" label="Link yayında" />
          <div class="min-w-48 flex-1">
            <p class="font-medium">Yayında</p>
            <p class="text-sm text-ink-3">Kapalıyken ziyaretçi "link devre dışı" sayfasını görür.</p>
          </div>
        </li>
      </ul>
    </section>
  </fieldset>
</template>
