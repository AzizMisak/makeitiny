<script setup lang="ts">
import type { LinkDraftErrors } from '~/utils/linkDraft'
import type { LinkItem } from '~/types/link'

const link = defineModel<LinkItem>({ required: true })
const emit = defineEmits<{ dirty: [value: boolean]; removed: [] }>()

const { canUseFeature } = useAuth()
const { updateLink } = useLinks()
const { display } = useShortUrl()
const { ask } = useConfirm()
const removeLink = useLinkRemoval()
const toast = useToast()

const draft = ref(draftFromLink(link.value))
const errors = ref<LinkDraftErrors>({})
const saveError = ref('')
const saving = ref(false)
const form = ref<{ aliasState: string } | null>(null)

const diff = computed(() => draftChanges(draft.value, link.value))
const count = computed(() => diff.value.changes.length)

watch(count, (n) => emit('dirty', n > 0), { immediate: true })
watch(
  draft,
  () => {
    errors.value = {}
    saveError.value = ''
  },
  { deep: true }
)

function reset(source: LinkItem = link.value) {
  draft.value = draftFromLink(source)
  errors.value = {}
  saveError.value = ''
}

async function apply() {
  const updated = await updateLink(
    link.value.id,
    toUpdatePayload(draft.value, link.value, {
      alias: canUseFeature('allowCustomAlias'),
      ads: canUseFeature('allowAds')
    })
  )
  link.value = updated
  reset(updated)
  toast.success('Link güncellendi.')
}

async function save() {
  saveError.value = ''
  errors.value = validateDraft(draft.value, link.value)
  if (form.value?.aliasState === 'taken') errors.value.alias = 'Bu kısa ad alınmış, başka bir ad dene.'
  if (Object.keys(errors.value).length > 0) {
    saveError.value = 'İşaretli alanları düzelt.'
    return
  }
  if (count.value === 0) return

  // Ziyaretçiyi etkilemeyen değişiklikler (başlık, reklam) doğrudan kaydedilir.
  if (diff.value.critical) {
    await ask({
      title: 'Değişiklikler yayına alınsın mı?',
      subject: display(link.value.alias),
      message: 'Aşağıdakiler hemen geçerli olur.',
      changes: diff.value.changes,
      effects: diff.value.effects,
      confirmLabel: 'Kaydet ve yayına al',
      cancelLabel: 'Düzenlemeye dön',
      action: apply
    })
    return
  }

  saving.value = true
  try {
    await apply()
  } catch (e: any) {
    saveError.value = e?.message ?? 'Kaydedilemedi. Tekrar dene.'
  } finally {
    saving.value = false
  }
}

async function remove() {
  if (await removeLink(link.value)) {
    emit('dirty', false)
    emit('removed')
  }
}
</script>

<template>
  <form class="flex max-w-3xl flex-col gap-6" novalidate @submit.prevent="save">
    <LinksForm ref="form" v-model="draft" :errors="errors" :original="link" :disabled="saving" />

    <section class="mt-4 flex flex-wrap items-center justify-between gap-4 border-t-2 border-danger pt-5">
      <div>
        <h3 class="font-semibold text-danger">Linki sil</h3>
        <p class="text-sm text-ink-2">Kısa adres ve analizi kalıcı olarak gider. Geçici durdurmak için "Yayında" anahtarını kapat.</p>
      </div>
      <UiButton variant="danger" icon="trash" @click="remove">Linki sil</UiButton>
    </section>

    <!-- Kayıt çubuğu sayfanın üstünde yüzer; açılıp kapanması içeriği itmez. -->
    <Transition name="pop">
      <div v-if="count > 0" class="fixed inset-x-0 bottom-16 z-30 border-t border-line bg-surface shadow-pop lg:bottom-0 lg:left-[17rem]">
        <div class="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3 px-5 py-3 md:px-10">
          <p class="text-sm" :class="saveError ? 'text-danger' : 'text-ink-2'" aria-live="polite">
            {{ saveError || `${count} değişiklik kaydedilmedi` }}
          </p>
          <div class="flex gap-2">
            <UiButton variant="quiet" size="sm" :disabled="saving" @click="reset()">Geri al</UiButton>
            <UiButton type="submit" size="sm" :loading="saving" loading-label="Kaydediliyor">Kaydet</UiButton>
          </div>
        </div>
      </div>
    </Transition>
  </form>
</template>
