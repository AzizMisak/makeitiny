<script setup lang="ts">
import type { ConfirmChange } from '~/composables/useConfirm'
import type { PlanItem } from '~/types/auth'

definePageMeta({ layout: 'admin', middleware: 'auth', roles: ['owner'] })
useHead({ title: 'Paketler | Yönetim | makeiTiny' })

const { fetchPlans, adminUpdatePlan } = useAdmin()
const { ask } = useConfirm()
const toast = useToast()

const plans = ref<PlanItem[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    plans.value = await fetchPlans()
  } catch (e: any) {
    error.value = e?.message ?? 'Paketler yüklenemedi.'
  } finally {
    loading.value = false
  }
}

onMounted(load)

// --- Düzenleme ---
const editing = ref<PlanItem | null>(null)

function draftOf(p: PlanItem) {
  return {
    name: p.name,
    price: String(p.priceMonthly),
    unlimited: p.maxLinks < 0,
    maxLinks: p.maxLinks < 0 ? '' : String(p.maxLinks),
    days: String(p.analyticsDays),
    allowPassword: p.allowPassword,
    allowCustomAlias: p.allowCustomAlias,
    allowQr: p.allowQr,
    allowAds: p.allowAds
  }
}

type Draft = ReturnType<typeof draftOf>
const draft = ref<Draft | null>(null)
const errors = ref<{ name?: string; price?: string; maxLinks?: string; days?: string }>({})

const FLAGS = [
  { key: 'allowPassword', label: 'Parolalı link' },
  { key: 'allowCustomAlias', label: 'Kendi kısa adını seçme' },
  { key: 'allowQr', label: 'QR kodu' },
  { key: 'allowAds', label: 'Reklam gösterimi' }
] as const

function open(p: PlanItem) {
  draft.value = draftOf(p)
  errors.value = {}
  editing.value = p
}

function close() {
  editing.value = null
}

const changes = computed<ConfirmChange[]>(() => {
  const p = editing.value
  const d = draft.value
  if (!p || !d) return []
  const onOff = (v: boolean) => (v ? 'Var' : 'Yok')
  const limit = (n: number) => (n < 0 ? 'Sınırsız' : formatNumber(n))
  const max = d.unlimited ? -1 : Number(d.maxLinks)
  const list: ConfirmChange[] = []
  if (d.name.trim() !== p.name) list.push({ label: 'Ad', from: p.name, to: d.name.trim() })
  if (Number(d.price) !== p.priceMonthly) list.push({ label: 'Fiyat', from: formatPriceMonthly(p.priceMonthly), to: formatPriceMonthly(Number(d.price)) })
  if (max !== p.maxLinks) list.push({ label: 'Link hakkı', from: limit(p.maxLinks), to: limit(max) })
  if (Number(d.days) !== p.analyticsDays) list.push({ label: 'Analiz geçmişi', from: `${p.analyticsDays} gün`, to: `${d.days} gün` })
  for (const f of FLAGS) {
    if (d[f.key] !== p[f.key]) list.push({ label: f.label, from: onOff(p[f.key]), to: onOff(d[f.key]) })
  }
  return list
})

function validate(d: Draft): boolean {
  const price = Number(d.price)
  const max = Number(d.maxLinks)
  const days = Number(d.days)
  errors.value = {
    name: d.name.trim() ? undefined : 'Paket adını yaz.',
    price: d.price.trim() !== '' && Number.isFinite(price) && price >= 0 ? undefined : '0 veya üstü yaz.',
    maxLinks: d.unlimited || (d.maxLinks.trim() !== '' && Number.isInteger(max) && max >= 0) ? undefined : '0 veya üstü tam sayı yaz.',
    days: Number.isInteger(days) && days >= 1 ? undefined : 'En az 1 gün yaz.'
  }
  return !Object.values(errors.value).some(Boolean)
}

async function save() {
  const p = editing.value
  const d = draft.value
  if (!p || !d || !validate(d) || changes.value.length === 0) return

  const summary = [...changes.value]
  const payload = {
    name: d.name.trim(),
    priceMonthly: Number(d.price),
    maxLinks: d.unlimited ? -1 : Number(d.maxLinks),
    analyticsDays: Number(d.days),
    allowCustomAlias: d.allowCustomAlias,
    allowPassword: d.allowPassword,
    allowQr: d.allowQr,
    allowAds: d.allowAds
  }

  editing.value = null
  const ok = await ask({
    title: 'Paket güncellensin mi?',
    subject: p.name,
    message: 'Bu paketteki bütün kullanıcılar için hemen geçerli olur.',
    changes: summary,
    confirmLabel: 'Paketi güncelle',
    cancelLabel: 'Düzenlemeye dön',
    action: async () => {
      const updated = await adminUpdatePlan(p.id, payload)
      plans.value = plans.value.map((x) => (x.id === p.id ? updated : x))
    }
  })
  if (ok) toast.success(`${payload.name} güncellendi.`)
  else editing.value = p
}
</script>

<template>
  <div>
    <AdminHeader title="Paketler" note="Yalnızca sistem sahibi görür" refreshable :loading="loading" @refresh="load" />

    <UiLoadError v-if="error" :message="error" :loading="loading" @retry="load" />
    <AdminTableSkeleton v-else-if="loading && plans.length === 0" :rows="6" label="Paketler yükleniyor" />

    <div v-else class="overflow-x-auto border border-line bg-surface transition-opacity" :class="loading ? 'opacity-55' : ''">
      <table class="w-full min-w-[46rem] table-fixed text-left">
        <thead class="bg-sunken">
          <tr>
            <th class="w-1/4 px-5 py-4 text-sm font-medium text-ink-2">Paket</th>
            <th v-for="p in plans" :key="p.id" scope="col" class="px-5 py-4 align-top">
              <div class="flex items-start justify-between gap-2">
                <div>
                  <p class="display text-xl">{{ p.name }}</p>
                  <p class="tnum mt-1 text-sm font-normal text-ink-2">{{ formatPriceMonthly(p.priceMonthly) }}</p>
                </div>
                <UiIconButton icon="edit" :label="`${p.name} paketini düzenle`" class="-mr-2 -mt-1" @click="open(p)" />
              </div>
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-line">
          <tr v-for="row in PLAN_ROWS" :key="row.key">
            <th scope="row" class="px-5 py-3.5 text-[15px] font-normal text-ink-2">{{ row.label }}</th>
            <td v-for="p in plans" :key="p.id" class="px-5 py-3.5">
              <span v-if="row.kind === 'text'" class="tnum font-semibold">{{ row.value(p) }}</span>
              <span v-else-if="row.value(p)" class="inline-flex items-center gap-2 text-ok"><UiIcon name="check" /> Var</span>
              <span v-else class="text-ink-3">Yok</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <UiModal :open="!!editing" :title="editing ? `Düzenle: ${editing.name}` : ''" wide @close="close">
      <form v-if="draft" id="plan-form" class="pt-3" novalidate @submit.prevent="save">
        <div class="grid grid-cols-1 gap-x-5 sm:grid-cols-2">
          <UiField id="plan-name" v-model="draft.name" label="Ad" :error="errors.name" />
          <UiField id="plan-price" v-model="draft.price" label="Aylık fiyat" type="number" min="0" step="0.01" inputmode="decimal" suffix="TL" :error="errors.price" />
          <UiField
            id="plan-max"
            v-model="draft.maxLinks"
            label="Link hakkı"
            type="number"
            min="0"
            inputmode="numeric"
            suffix="link"
            :disabled="draft.unlimited"
            :placeholder="draft.unlimited ? 'Sınırsız' : ''"
            :error="errors.maxLinks"
          />
          <UiField id="plan-days" v-model="draft.days" label="Analiz geçmişi" type="number" min="1" inputmode="numeric" suffix="gün" :error="errors.days" />
        </div>
        <ul class="divide-y divide-line border-y border-line">
          <li class="flex items-center justify-between gap-4 py-3">
            <span class="font-medium">Sınırsız link</span>
            <UiSwitch v-model="draft.unlimited" label="Sınırsız link" />
          </li>
          <li v-for="f in FLAGS" :key="f.key" class="flex items-center justify-between gap-4 py-3">
            <span class="font-medium">{{ f.label }}</span>
            <UiSwitch v-model="draft[f.key]" :label="f.label" />
          </li>
        </ul>
      </form>
      <template #footer>
        <UiButton variant="quiet" @click="close">Vazgeç</UiButton>
        <UiButton type="submit" form="plan-form" :disabled="changes.length === 0">
          {{ changes.length === 0 ? 'Değişiklik yok' : `${changes.length} değişikliği gözden geçir` }}
        </UiButton>
      </template>
    </UiModal>
  </div>
</template>
