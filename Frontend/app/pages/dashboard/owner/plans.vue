<script setup lang="ts">
import type { ConfirmChangeItem } from '~/components/ui/ConfirmModal.vue'
import type { PlanItem } from '~/types/auth'

definePageMeta({ layout: 'dashboard', middleware: 'auth', roles: ['owner'] })
useHead({ title: 'Paket Yönetimi — Owner — makeiTiny' })

interface PlanDraft {
  name: string
  priceMonthly: number
  maxLinks: number
  analyticsDays: number
  allowCustomAlias: boolean
  allowPassword: boolean
  allowQr: boolean
  allowAds: boolean
}

const { fetchPlans, adminUpdatePlan } = useAdmin()
const toast = useToast()

const plans = ref<PlanItem[]>([])
const loading = ref(true)
const saving = ref(false)

const editingPlan = ref<PlanItem | null>(null)
const reviewStep = ref(false)
const draft = reactive<PlanDraft>({
  name: '',
  priceMonthly: 0,
  maxLinks: 50,
  analyticsDays: 7,
  allowCustomAlias: false,
  allowPassword: false,
  allowQr: false,
  allowAds: false
})

const unlimitedPlansCount = computed(() => plans.value.filter((p) => p.maxLinks < 0).length)
const maxAnalyticsDays = computed(() =>
  plans.value.length === 0 ? 0 : Math.max(...plans.value.map((p) => p.analyticsDays))
)

function openEditModal(plan: PlanItem) {
  editingPlan.value = plan
  reviewStep.value = false
  draft.name = plan.name
  draft.priceMonthly = plan.priceMonthly
  draft.maxLinks = plan.maxLinks
  draft.analyticsDays = plan.analyticsDays
  draft.allowCustomAlias = plan.allowCustomAlias
  draft.allowPassword = plan.allowPassword
  draft.allowQr = plan.allowQr
  draft.allowAds = plan.allowAds
}

function closeEditModal() {
  if (saving.value) return
  editingPlan.value = null
  reviewStep.value = false
}

const pendingChanges = computed<ConfirmChangeItem[]>(() => {
  const orig = editingPlan.value
  if (!orig) return []
  const list: ConfirmChangeItem[] = []

  if (draft.name.trim() !== orig.name) {
    list.push({ label: 'Paket adı', from: orig.name, to: draft.name.trim() })
  }
  if (Number(draft.priceMonthly) !== orig.priceMonthly) {
    list.push({
      label: 'Aylık fiyat',
      from: formatPriceMonthly(orig.priceMonthly),
      to: formatPriceMonthly(Number(draft.priceMonthly))
    })
  }
  if (Number(draft.maxLinks) !== orig.maxLinks) {
    list.push({
      label: 'Link kotası',
      from: orig.maxLinks < 0 ? 'Sınırsız' : `${orig.maxLinks}`,
      to: Number(draft.maxLinks) < 0 ? 'Sınırsız' : `${Number(draft.maxLinks)}`
    })
  }
  if (Number(draft.analyticsDays) !== orig.analyticsDays) {
    list.push({
      label: 'Analitik süresi',
      from: `${orig.analyticsDays} gün`,
      to: `${Number(draft.analyticsDays)} gün`
    })
  }
  if (draft.allowCustomAlias !== orig.allowCustomAlias) {
    list.push({
      label: 'Özel takma ad',
      from: orig.allowCustomAlias ? 'Açık' : 'Kapalı',
      to: draft.allowCustomAlias ? 'Açık' : 'Kapalı'
    })
  }
  if (draft.allowPassword !== orig.allowPassword) {
    list.push({
      label: 'Şifreli link',
      from: orig.allowPassword ? 'Açık' : 'Kapalı',
      to: draft.allowPassword ? 'Açık' : 'Kapalı'
    })
  }
  if (draft.allowQr !== orig.allowQr) {
    list.push({
      label: 'QR kod',
      from: orig.allowQr ? 'Açık' : 'Kapalı',
      to: draft.allowQr ? 'Açık' : 'Kapalı'
    })
  }
  if (draft.allowAds !== orig.allowAds) {
    list.push({
      label: 'Reklam modu',
      from: orig.allowAds ? 'Açık' : 'Kapalı',
      to: draft.allowAds ? 'Açık' : 'Kapalı'
    })
  }

  return list
})

async function loadPlans() {
  loading.value = true
  try {
    plans.value = await fetchPlans()
  } catch (err: any) {
    toast.error(err?.message ?? 'Paketler yüklenemedi.')
  } finally {
    loading.value = false
  }
}

function proceedToReview() {
  if (!draft.name.trim()) {
    toast.error('Paket adı boş bırakılamaz.')
    return
  }
  if (!Number.isFinite(draft.priceMonthly) || draft.priceMonthly < 0) {
    toast.error('Aylık fiyat 0 veya pozitif olmalı.')
    return
  }
  if (!Number.isInteger(draft.maxLinks) || draft.maxLinks < -1) {
    toast.error('Maksimum link -1 (sınırsız) veya >= 0 tam sayı olmalı.')
    return
  }
  if (!Number.isInteger(draft.analyticsDays) || draft.analyticsDays < 1) {
    toast.error('Analitik süresi en az 1 gün olmalı.')
    return
  }
  if (pendingChanges.value.length === 0) {
    toast.info('Kaydedilecek bir değişiklik yapmadın.')
    return
  }
  reviewStep.value = true
}

async function confirmSavePlan() {
  const target = editingPlan.value
  if (!target) return
  saving.value = true
  try {
    const updated = await adminUpdatePlan(target.id, {
      name: draft.name.trim(),
      priceMonthly: Number(draft.priceMonthly),
      maxLinks: Number(draft.maxLinks),
      analyticsDays: Number(draft.analyticsDays),
      allowCustomAlias: draft.allowCustomAlias,
      allowPassword: draft.allowPassword,
      allowQr: draft.allowQr,
      allowAds: draft.allowAds
    })
    const idx = plans.value.findIndex((p) => p.id === target.id)
    if (idx !== -1) plans.value[idx] = updated
    editingPlan.value = null
    reviewStep.value = false
    toast.success(`'${updated.name}' paketi güncellendi.`)
  } catch (err: any) {
    toast.error(err?.message ?? 'Paket güncellenemedi.')
  } finally {
    saving.value = false
  }
}

onMounted(loadPlans)
</script>

<template>
  <DashboardPageHeader
    title="Paket yönetimi"
    description="Mevcut paket yapılandırmalarını incele; düzenlemek istediğin paketi tekil olarak aç."
  >
    <template #actions>
      <UiButton variant="secondary" :disabled="loading" @click="loadPlans">
        <UiIcon name="refresh" :size="15" /> Yenile
      </UiButton>
    </template>
  </DashboardPageHeader>

  <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
    <DashboardStatTile
      label="Tanımlı paket"
      :value="loading ? '...' : plans.length"
    />
    <DashboardStatTile
      label="Sınırsız plan"
      :value="loading ? '...' : unlimitedPlansCount"
    />
    <DashboardStatTile
      label="Maks. analitik"
      :value="loading ? '...' : `${maxAnalyticsDays} gün`"
    />
    <DashboardStatTile
      label="Yetki"
      value="Owner"
    />
  </div>

  <div v-if="loading" class="mt-8 grid gap-6 lg:grid-cols-3">
    <div v-for="i in 3" :key="i" class="h-64 animate-pulse rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard" />
  </div>

  <!-- Salt Okunur Paket Kartları -->
  <div v-else class="mt-8 grid gap-6 lg:grid-cols-3">
    <article
      v-for="plan in plans"
      :key="plan.id"
      class="flex flex-col justify-between rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard"
    >
      <div>
        <div class="flex items-start justify-between gap-2 border-b border-line pb-4">
          <div>
            <p class="font-mono text-xs uppercase tracking-wide text-ink-3">#{{ plan.id }} • {{ plan.slug }}</p>
            <h2 class="mt-0.5 text-xl font-bold text-ink">{{ plan.name }}</h2>
          </div>
          <UiBadge :tone="plan.maxLinks < 0 ? 'accent' : 'neutral'">
            {{ plan.maxLinks < 0 ? 'Sınırsız' : `${plan.maxLinks} link` }}
          </UiBadge>
        </div>

        <p class="mt-4 font-display text-2xl font-extrabold text-ink">
          {{ formatPriceMonthly(plan.priceMonthly) }}
        </p>

        <dl class="mt-4 divide-y divide-line text-xs">
          <div class="flex items-center justify-between py-2">
            <dt class="text-ink-3">Link kotası</dt>
            <dd class="font-mono font-bold text-ink">{{ plan.maxLinks < 0 ? 'Sınırsız (-1)' : plan.maxLinks }}</dd>
          </div>
          <div class="flex items-center justify-between py-2">
            <dt class="text-ink-3">Analitik geçmişi</dt>
            <dd class="font-mono font-bold text-ink">{{ plan.analyticsDays }} gün</dd>
          </div>
          <div class="flex items-center justify-between py-2">
            <dt class="text-ink-3">Özel takma ad</dt>
            <dd class="font-semibold" :class="plan.allowCustomAlias ? 'text-ok' : 'text-ink-3'">
              {{ plan.allowCustomAlias ? 'Açık' : 'Kapalı' }}
            </dd>
          </div>
          <div class="flex items-center justify-between py-2">
            <dt class="text-ink-3">Şifre koruması</dt>
            <dd class="font-semibold" :class="plan.allowPassword ? 'text-ok' : 'text-ink-3'">
              {{ plan.allowPassword ? 'Açık' : 'Kapalı' }}
            </dd>
          </div>
          <div class="flex items-center justify-between py-2">
            <dt class="text-ink-3">QR kod</dt>
            <dd class="font-semibold" :class="plan.allowQr ? 'text-ok' : 'text-ink-3'">
              {{ plan.allowQr ? 'Açık' : 'Kapalı' }}
            </dd>
          </div>
          <div class="flex items-center justify-between py-2">
            <dt class="text-ink-3">Reklam modu</dt>
            <dd class="font-semibold" :class="plan.allowAds ? 'text-ok' : 'text-ink-3'">
              {{ plan.allowAds ? 'Açık' : 'Kapalı' }}
            </dd>
          </div>
        </dl>
      </div>

      <div class="mt-6 border-t border-line pt-4">
        <UiButton variant="secondary" block @click="openEditModal(plan)">
          <UiIcon name="sliders" :size="15" /> Paketi düzenle...
        </UiButton>
      </div>
    </article>
  </div>

  <!-- Tekil Paket Düzenleme ve Onay Modalı -->
  <Teleport to="body">
    <div
      v-if="editingPlan"
      class="fixed inset-0 z-50 flex items-center justify-center bg-ink/50 p-4"
      role="dialog"
      aria-modal="true"
      @keydown.esc="closeEditModal"
    >
      <div class="w-full max-w-lg rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard-lg">
        <div class="flex items-start justify-between gap-3 border-b border-line pb-4">
          <div>
            <p class="font-mono text-xs uppercase text-ink-3">Paket #{{ editingPlan.id }} • {{ editingPlan.slug }}</p>
            <h2 class="mt-0.5 text-xl font-bold text-ink">
              {{ reviewStep ? 'Değişiklikleri onayla' : `${editingPlan.name} paketini düzenle` }}
            </h2>
          </div>
          <button
            type="button"
            class="cursor-pointer rounded px-2 py-1 font-mono text-xs text-ink-3 hover:bg-paper-2 hover:text-ink"
            @click="closeEditModal"
          >
            ✕
          </button>
        </div>

        <!-- Adım 1: Tekil Düzenleme Formu -->
        <form v-if="!reviewStep" class="mt-5 flex flex-col gap-4" @submit.prevent="proceedToReview">
          <div>
            <label for="edit-plan-name" class="mb-1.5 block text-xs font-semibold text-ink-2">Paket adı</label>
            <input
              id="edit-plan-name"
              v-model="draft.name"
              type="text"
              class="w-full rounded-sm border-[1.5px] border-ink bg-paper px-3 py-2 text-sm text-ink focus:outline-none focus:ring-2 focus:ring-accent"
            />
          </div>

          <div class="grid gap-3 sm:grid-cols-3">
            <div>
              <label for="edit-plan-price" class="mb-1.5 block text-xs font-semibold text-ink-2">Aylık fiyat (TL)</label>
              <input
                id="edit-plan-price"
                v-model.number="draft.priceMonthly"
                type="number"
                min="0"
                step="0.01"
                class="w-full rounded-sm border-[1.5px] border-ink bg-paper px-3 py-2 font-mono text-sm text-ink focus:outline-none focus:ring-2 focus:ring-accent"
              />
            </div>

            <div>
              <label for="edit-plan-max" class="mb-1.5 block text-xs font-semibold text-ink-2">Max link (-1=Sınırsız)</label>
              <input
                id="edit-plan-max"
                v-model.number="draft.maxLinks"
                type="number"
                min="-1"
                step="1"
                class="w-full rounded-sm border-[1.5px] border-ink bg-paper px-3 py-2 font-mono text-sm text-ink focus:outline-none focus:ring-2 focus:ring-accent"
              />
            </div>

            <div>
              <label for="edit-plan-days" class="mb-1.5 block text-xs font-semibold text-ink-2">Analitik (Gün)</label>
              <input
                id="edit-plan-days"
                v-model.number="draft.analyticsDays"
                type="number"
                min="1"
                step="1"
                class="w-full rounded-sm border-[1.5px] border-ink bg-paper px-3 py-2 font-mono text-sm text-ink focus:outline-none focus:ring-2 focus:ring-accent"
              />
            </div>
          </div>

          <div class="rounded-md border border-line bg-paper p-3.5">
            <p class="mb-2.5 text-xs font-semibold uppercase tracking-wide text-ink-3">Özellik İzinleri</p>
            <div class="grid gap-2.5 sm:grid-cols-2">
              <label class="flex cursor-pointer items-center gap-2 text-xs font-medium text-ink select-none">
                <input v-model="draft.allowCustomAlias" type="checkbox" class="h-4 w-4 accent-accent" />
                <span>Özel takma ad</span>
              </label>
              <label class="flex cursor-pointer items-center gap-2 text-xs font-medium text-ink select-none">
                <input v-model="draft.allowPassword" type="checkbox" class="h-4 w-4 accent-accent" />
                <span>Şifreli link</span>
              </label>
              <label class="flex cursor-pointer items-center gap-2 text-xs font-medium text-ink select-none">
                <input v-model="draft.allowQr" type="checkbox" class="h-4 w-4 accent-accent" />
                <span>QR kod</span>
              </label>
              <label class="flex cursor-pointer items-center gap-2 text-xs font-medium text-ink select-none">
                <input v-model="draft.allowAds" type="checkbox" class="h-4 w-4 accent-accent" />
                <span>Reklam modu</span>
              </label>
            </div>
          </div>

          <div class="mt-2 flex items-center justify-end gap-2.5 border-t border-line pt-4">
            <UiButton variant="ghost" @click="closeEditModal">Vazgeç</UiButton>
            <UiButton type="submit" :disabled="pendingChanges.length === 0">
              Değişiklikleri incele ({{ pendingChanges.length }})
            </UiButton>
          </div>
        </form>

        <!-- Adım 2: Önce / Sonra Fark Onayı -->
        <div v-else class="mt-5">
          <p class="text-sm text-ink-2">
            <strong>{{ editingPlan.name }}</strong> paketinde aşağıdaki değişiklikler bu pakete sahip tüm kullanıcılar için geçerli olacak:
          </p>

          <dl class="mt-4 divide-y divide-line rounded-md border border-line bg-paper text-xs">
            <div v-for="ch in pendingChanges" :key="ch.label" class="flex items-center justify-between gap-2 px-3.5 py-2.5">
              <dt class="font-medium text-ink-3">{{ ch.label }}</dt>
              <dd class="font-mono font-semibold text-ink">
                <span class="text-ink-3 line-through">{{ ch.from }}</span>
                <span class="mx-1.5 text-ink-3">→</span>
                <span class="text-ink">{{ ch.to }}</span>
              </dd>
            </div>
          </dl>

          <div class="mt-6 flex items-center justify-end gap-2.5 border-t border-line pt-4">
            <UiButton variant="secondary" :disabled="saving" @click="reviewStep = false">
              ← Geri dön
            </UiButton>
            <UiButton :disabled="saving" @click="confirmSavePlan">
              {{ saving ? 'Kaydediliyor...' : 'Onayla ve kaydet' }}
            </UiButton>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
