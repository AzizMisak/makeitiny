<script setup lang="ts">
import type { ConfirmChange } from '~/composables/useConfirm'
import type { AdminLinkItem } from '~/types/link'

const link = defineModel<AdminLinkItem>({ required: true })
const emit = defineEmits<{ dirty: [value: boolean]; removed: [] }>()

const { adminUpdateLink } = useAdmin()
const { remove } = useLinkModeration()
const { ask } = useConfirm()
const { display } = useShortUrl()
const toast = useToast()

function fresh(source: AdminLinkItem = link.value) {
  return {
    target: source.target,
    title: source.title ?? '',
    expiresAt: source.expiresAt ?? '',
    maxClicks: source.maxClicks ? String(source.maxClicks) : '',
    adEnabled: source.adEnabled
  }
}

const form = ref(fresh())
const errors = ref<{ target?: string; maxClicks?: string }>({})

watch(form, () => (errors.value = {}), { deep: true })

const changes = computed<ConfirmChange[]>(() => {
  const was = fresh()
  const now = form.value
  const list: ConfirmChange[] = []
  if (now.target.trim() !== was.target) list.push({ label: 'Hedef adres', from: was.target, to: now.target.trim() })
  if (now.title.trim() !== was.title) list.push({ label: 'Başlık', from: was.title || 'Yok', to: now.title.trim() || 'Yok' })
  if (now.expiresAt !== was.expiresAt) {
    list.push({
      label: 'Son kullanma',
      from: was.expiresAt ? formatDateTr(was.expiresAt) : 'Süresiz',
      to: now.expiresAt ? formatDateTr(now.expiresAt) : 'Süresiz'
    })
  }
  if (now.maxClicks !== was.maxClicks) {
    list.push({ label: 'Tıklama sınırı', from: was.maxClicks || 'Sınırsız', to: now.maxClicks || 'Sınırsız' })
  }
  if (now.adEnabled !== was.adEnabled) {
    list.push({ label: 'Reklam', from: was.adEnabled ? 'Açık' : 'Kapalı', to: now.adEnabled ? 'Açık' : 'Kapalı' })
  }
  return list
})

watch(() => changes.value.length, (n) => emit('dirty', n > 0), { immediate: true })

async function save() {
  const max = form.value.maxClicks.trim()
  errors.value = {
    target: form.value.target.trim() ? undefined : 'Hedef adresi yaz.',
    maxClicks: max && !(Number.isInteger(Number(max)) && Number(max) >= 1) ? 'En az 1 yaz veya boş bırak.' : undefined
  }
  if (errors.value.target || errors.value.maxClicks || changes.value.length === 0) return

  const ok = await ask({
    title: 'Link değiştirilsin mi?',
    subject: `${display(link.value.alias)}${link.value.userEmail ? ` (${link.value.userEmail})` : ''}`,
    message: 'Başka bir kullanıcının linkini düzenliyorsun. Değişiklik hemen geçerli olur.',
    changes: changes.value,
    confirmLabel: 'Değişiklikleri kaydet',
    action: async () => {
      const updated = await adminUpdateLink(link.value.id, {
        target: form.value.target.trim(),
        title: form.value.title.trim(),
        expiresAt: form.value.expiresAt,
        maxClicks: max ? Number(max) : 0,
        adEnabled: form.value.adEnabled
      })
      const next = { ...link.value, ...updated }
      link.value = next
      form.value = fresh(next)
    }
  })
  if (ok) toast.success('Link güncellendi.')
}

async function onRemove() {
  if (await remove(link.value)) {
    emit('dirty', false)
    emit('removed')
  }
}
</script>

<template>
  <form class="max-w-2xl" novalidate @submit.prevent="save">
    <UiField id="adm-target" v-model="form.target" label="Hedef adres" :error="errors.target" />
    <UiField id="adm-title" v-model="form.title" label="Başlık" placeholder="Yok" class="mt-1" />
    <div class="mt-1 grid grid-cols-1 gap-x-5 sm:grid-cols-2">
      <UiField id="adm-exp" v-model="form.expiresAt" label="Son kullanma tarihi" type="date" hint="Boşsa süresiz" />
      <UiField
        id="adm-max"
        v-model="form.maxClicks"
        label="Tıklama sınırı"
        type="number"
        min="1"
        placeholder="Sınırsız"
        hint="Boşsa sınırsız"
        :error="errors.maxClicks"
      />
    </div>
    <div class="mt-3 flex items-center gap-4 border-t border-line py-4">
      <UiSwitch v-model="form.adEnabled" label="Reklam gösterimi" />
      <p class="font-medium">Reklam gösterimi</p>
    </div>
    <div class="flex min-h-11 flex-wrap items-center gap-3 border-t border-line pt-4">
      <UiButton type="submit" variant="outline" :disabled="changes.length === 0">
        {{ changes.length === 0 ? 'Değişiklik yok' : `${changes.length} değişikliği kaydet` }}
      </UiButton>
      <UiButton v-if="changes.length > 0" variant="quiet" @click="form = fresh()">Geri al</UiButton>
    </div>

    <section class="mt-12 flex flex-wrap items-center justify-between gap-4 border-t-2 border-danger pt-5">
      <div>
        <h3 class="font-semibold text-danger">Linki sil</h3>
        <p class="text-sm text-ink-2">Geri alınamaz. Durdurmak için karar panelini kullan.</p>
      </div>
      <UiButton variant="danger" icon="trash" @click="onRemove">Linki sil</UiButton>
    </section>
  </form>
</template>
