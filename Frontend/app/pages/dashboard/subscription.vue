<script setup lang="ts">
import type { PlanItem } from '~/types/auth'

definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Paketim | makeiTiny' })

const { user, isAdmin, remainingLinks, fetchCurrentUser, changePlan } = useAuth()
const { fetchPlans } = useAdmin()
const { ask } = useConfirm()
const toast = useToast()

const plans = ref<PlanItem[]>([])
const loading = ref(true)
const error = ref('')

const used = computed(() => user.value?.linkCount ?? 0)
const max = computed(() => (isAdmin.value ? -1 : (user.value?.plan?.maxLinks ?? 50)))

const usageText = computed(() => {
  if (isAdmin.value) return PRIVILEGED_PLAN_NOTE
  if (max.value < 0) return 'Bu pakette link sınırı yok.'
  return `${formatNumber(remainingLinks.value ?? 0)} link hakkın kaldı.`
})

function isCurrent(p: PlanItem): boolean {
  return !isAdmin.value && p.id === user.value?.plan?.id
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [list] = await Promise.all([fetchPlans(), fetchCurrentUser()])
    plans.value = list
  } catch (e: any) {
    error.value = e?.message ?? 'Paket bilgileri yüklenemedi.'
  } finally {
    loading.value = false
  }
}

function limitText(n: number): string {
  return n < 0 ? 'Sınırsız' : `${formatNumber(n)} link`
}

/** Geçiş doğrudan yapılmaz; önce sonuçlarını gösteren onay penceresi açılır. */
async function switchTo(p: PlanItem) {
  const current = user.value?.plan
  if (!current || isCurrent(p)) return

  const downgrade = p.maxLinks >= 0 && (current.maxLinks < 0 || p.maxLinks < current.maxLinks)
  const over = p.maxLinks >= 0 ? Math.max(0, used.value - p.maxLinks) : 0
  const effects: string[] = []
  if (over > 0) {
    effects.push(`Sınırı aşan en eski ${formatNumber(over)} linkin askıya alınır.`)
    effects.push('Askıdaki linkler çalışmaya devam eder ve 2 ay sonra kalıcı olarak silinir; bu sürede istediğini kendin silebilirsin.')
    effects.push('2 ay dolmadan yeterli bir pakete dönersen askı kalkar.')
    effects.push('Toplam link sayın sınırın altına inene kadar yeni link oluşturamazsın.')
  } else if (!downgrade) {
    effects.push('Askıdaki linklerin varsa yeni sınıra sığanların askısı kalkar.')
  }

  const ok = await ask({
    title: downgrade ? 'Daha düşük pakete geçilsin mi?' : 'Paket değiştirilsin mi?',
    subject: p.name,
    message: 'Değişiklik hemen geçerli olur. Ödeme alınmaz; paket sen değiştirene kadar süresiz geçerlidir.',
    changes: [
      { label: 'Paket', from: current.name, to: p.name },
      { label: 'Aylık ücret', from: formatPriceMonthly(current.priceMonthly), to: formatPriceMonthly(p.priceMonthly) },
      { label: 'Link hakkı', from: limitText(current.maxLinks), to: limitText(p.maxLinks) }
    ],
    effects,
    tone: over > 0 ? 'danger' : 'normal',
    confirmLabel: `${p.name} paketine geç`,
    action: () => changePlan(p.id)
  })
  if (ok) toast.success(`${p.name} paketine geçtin.`)
}

onMounted(load)
</script>

<template>
  <div>
    <PanelHeader :title="isAdmin ? `${roleMeta(user?.roles).label} hesabı` : (user?.plan?.name ?? 'Paketim')">
      <UiTag v-if="isAdmin" tone="mark" icon="key">Pakete tabi değil</UiTag>
      <p v-else class="text-ink-2">
        {{ formatPriceMonthly(user?.plan?.priceMonthly ?? 0) }}<template v-if="user?.planExpiresAt">, {{ formatDateTr(user.planExpiresAt) }} tarihine kadar</template>
      </p>
    </PanelHeader>

    <section>
      <p class="flex flex-wrap items-baseline gap-x-3">
        <span class="display tnum text-6xl">{{ formatNumber(used) }}</span>
        <span class="text-lg text-ink-2">{{ max < 0 ? 'link kullandın' : `/ ${formatNumber(max)} link` }}</span>
      </p>
      <UiTape class="mt-5" :used="used" :max="max" label="Link kullanımı" />
      <p class="mt-3 text-ink-2">{{ usageText }}</p>
    </section>

    <section class="mt-16">
      <h2 class="display text-2xl">Paketleri karşılaştır</h2>
      <p class="mt-2 text-ink-2">
        {{ isAdmin ? 'Bu paketler standart kullanıcılar içindir. Yönetici yetkin kaldırılırsa kayıtlı paketin yeniden geçerli olur.' : 'İstediğin pakete geçebilirsin; geçmeden önce ne değişeceğini gösterip onayını isteriz.' }}
      </p>

      <UiLoadError v-if="error" :message="error" :loading="loading" class="mt-6" @retry="load" />
      <div v-else-if="loading" class="mt-6 flex flex-col gap-2" aria-busy="true" aria-label="Paketler yükleniyor">
        <div class="skeleton h-20" />
        <div v-for="i in 6" :key="i" class="skeleton h-11" />
      </div>

      <div v-else class="mt-6 overflow-x-auto">
        <table class="w-full min-w-[38rem] table-fixed border-collapse text-left">
          <thead>
            <tr>
              <td class="w-1/4" />
              <th
                v-for="p in plans"
                :key="p.id"
                scope="col"
                class="px-4 pb-4 pt-5 align-bottom"
                :class="isCurrent(p) ? 'bg-brand-soft' : ''"
              >
                <span class="block min-h-6 text-[13px] font-medium text-brand-ink">{{ isCurrent(p) ? 'Şu anki paketin' : '' }}</span>
                <span class="display block text-xl">{{ p.name }}</span>
                <span class="tnum mt-1 block text-sm font-normal text-ink-2">{{ formatPriceMonthly(p.priceMonthly) }}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in PLAN_ROWS" :key="row.key" class="border-t border-line">
              <th scope="row" class="py-3.5 pr-4 text-[15px] font-normal text-ink-2">{{ row.label }}</th>
              <td v-for="p in plans" :key="p.id" class="px-4 py-3.5" :class="isCurrent(p) ? 'bg-brand-soft' : ''">
                <span v-if="row.kind === 'text'" class="tnum font-semibold">{{ row.value(p) }}</span>
                <span v-else-if="row.value(p)" class="inline-flex items-center gap-2 text-ok"><UiIcon name="check" /> Var</span>
                <span v-else class="text-ink-3">Yok</span>
              </td>
            </tr>
          </tbody>
          <tfoot v-if="!isAdmin">
            <tr class="border-t border-line">
              <td />
              <td v-for="p in plans" :key="p.id" class="px-4 py-4" :class="isCurrent(p) ? 'bg-brand-soft' : ''">
                <span v-if="isCurrent(p)" class="text-sm text-ink-3">Kullanıyorsun</span>
                <UiButton v-else size="sm" variant="outline" @click="switchTo(p)">Bu pakete geç</UiButton>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </section>
  </div>
</template>
