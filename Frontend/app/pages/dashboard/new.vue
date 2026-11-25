<script setup lang="ts">
import type { LinkDraftErrors } from '~/utils/linkDraft'

definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Yeni link | makeiTiny' })

const { isAdmin, canUseFeature, remainingLinks } = useAuth()
const { addLink } = useLinks()
const toast = useToast()

const handoff = useState<string>('mit_draft_target', () => '')
const draft = ref(emptyDraft())
const errors = ref<LinkDraftErrors>({})
const submitError = ref('')
const loading = ref(false)
const created = ref(false)
const form = ref<{ aliasState: string } | null>(null)

// Ana sayfadan taşınan adres
if (handoff.value) draft.value.target = handoff.value

onMounted(() => {
  handoff.value = ''
})

const full = computed(() => !isAdmin.value && remainingLinks.value !== null && remainingLinks.value <= 0)
const dirty = computed(() => JSON.stringify(draft.value) !== JSON.stringify(emptyDraft()))

useUnsavedGuard(() => dirty.value && !created.value)

watch(draft, () => (errors.value = {}), { deep: true })

async function submit() {
  submitError.value = ''
  errors.value = validateDraft(draft.value)
  if (form.value?.aliasState === 'taken') errors.value.alias = 'Bu kısa ad alınmış, başka bir ad dene.'
  if (Object.keys(errors.value).length > 0) return

  loading.value = true
  try {
    const link = await addLink(
      toCreatePayload(draft.value, {
        alias: canUseFeature('allowCustomAlias'),
        password: canUseFeature('allowPassword'),
        ads: canUseFeature('allowAds')
      })
    )
    created.value = true
    toast.success('Link oluşturuldu.')
    await navigateTo(`/dashboard/links/${link.id}`)
  } catch (e: any) {
    submitError.value = e?.message ?? 'Link oluşturulamadı. Tekrar dene.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <PanelHeader title="Yeni link" />

    <div v-if="full" class="mb-8 flex flex-wrap items-center justify-between gap-4 bg-warn-soft px-5 py-4 text-warn">
      <p class="font-medium">Paketindeki link hakkı doldu. Yeni link için eski bir linki sil ya da paketlere bak.</p>
      <NuxtLink to="/dashboard/subscription" class="font-semibold underline underline-offset-4">Paketim</NuxtLink>
    </div>

    <form class="grid grid-cols-1 items-start gap-10 lg:grid-cols-[minmax(0,1fr)_20rem]" novalidate @submit.prevent="submit">
      <div class="flex flex-col gap-6">
        <LinksForm ref="form" v-model="draft" :errors="errors" :disabled="loading" />
      </div>

      <LinksPreview :draft="draft" class="lg:sticky lg:top-12">
        <UiButton variant="mark" size="lg" type="submit" icon="scissors" block :loading="loading" loading-label="Oluşturuluyor" :disabled="full">
          Linki oluştur
        </UiButton>
        <p class="mt-3 min-h-10 text-sm leading-5 text-mark" aria-live="polite">{{ submitError }}</p>
      </LinksPreview>
    </form>
  </div>
</template>
