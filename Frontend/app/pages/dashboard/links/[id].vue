<script setup lang="ts">
import type { LinkItem, LinkLogItem } from '~/types/link'

definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const route = useRoute()
const { public: config } = useRuntimeConfig()
const { getLink, getLinkLogs, updateLink, removeLink } = useLinks()

const id = computed(() => String(route.params.id ?? ''))
const link = ref<LinkItem | null>(null)
const logs = ref<LinkLogItem[]>([])
const logsLoading = ref(false)
const logsError = ref('')
const pageLoading = ref(true)
const pageError = ref('')

const editTarget = ref('')
const editAlias = ref('')
const newPassword = ref('')
const removePasswordFlag = ref(false)
const editWaitEnabled = ref(false)
const editWaitSeconds = ref('5')

const saving = ref(false)
const saveSuccess = ref(false)
const saveError = ref('')
const copied = ref(false)
const confirmDelete = ref(false)

// Pagination (Sayfa başına 20 log kaydı)
const LOGS_PER_PAGE = 20
const currentLogPage = ref(1)

const totalLogPages = computed(() => Math.max(1, Math.ceil(logs.value.length / LOGS_PER_PAGE)))

const paginatedLogs = computed(() => {
  const start = (currentLogPage.value - 1) * LOGS_PER_PAGE
  return logs.value.slice(start, start + LOGS_PER_PAGE)
})

const logRangeStart = computed(() => (logs.value.length === 0 ? 0 : (currentLogPage.value - 1) * LOGS_PER_PAGE + 1))
const logRangeEnd = computed(() => Math.min(currentLogPage.value * LOGS_PER_PAGE, logs.value.length))

const visibleLogPages = computed<(number | '...')[]>(() => {
  const total = totalLogPages.value
  const current = currentLogPage.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  const pages: (number | '...')[] = [1]
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let p = start; p <= end; p++) {
    pages.push(p)
  }
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

function goToLogPage(page: number) {
  if (page < 1 || page > totalLogPages.value) return
  currentLogPage.value = page
}

// Grafik ve Coğrafi Analiz Durumları
type TimeRangeMode = '7d' | '14d' | '30d' | 'hourly'
type DistributionTab = 'all' | 'device' | 'browser' | 'os' | 'country'

const selectedTimeRange = ref<TimeRangeMode>('7d')
const selectedDistTab = ref<DistributionTab>('all')

interface CountryResolution {
  code: string
  name: string
  region: string
  normalizedLang: string
}

const LANG_TO_COUNTRY: Record<string, string> = {
  tr: 'TR',
  en: 'US',
  de: 'DE',
  fr: 'FR',
  es: 'ES',
  it: 'IT',
  nl: 'NL',
  ru: 'RU',
  az: 'AZ',
  ar: 'SA',
  pt: 'PT',
  zh: 'CN',
  ja: 'JP',
  ko: 'KR',
  pl: 'PL',
  uk: 'UA',
  el: 'GR',
  bg: 'BG',
  ro: 'RO',
  sv: 'SE',
  no: 'NO',
  nb: 'NO',
  nn: 'NO',
  da: 'DK',
  fi: 'FI',
  cs: 'CZ',
  hu: 'HU',
  sk: 'SK',
  hr: 'HR',
  sr: 'RS',
  fa: 'IR',
  hi: 'IN',
  id: 'ID',
  ms: 'MY',
  vi: 'VN',
  th: 'TH',
  he: 'IL',
  ca: 'ES',
  ka: 'GE',
  kk: 'KZ',
  uz: 'UZ'
}

const COUNTRY_NAMES_TR: Record<string, { name: string; region: string }> = {
  TR: { name: 'Türkiye', region: 'Türkiye & Yakın Coğrafya' },
  AZ: { name: 'Azerbaycan', region: 'Türkiye & Yakın Coğrafya' },
  CY: { name: 'Kuzey Kıbrıs / Kıbrıs', region: 'Türkiye & Yakın Coğrafya' },
  GE: { name: 'Gürcistan', region: 'Türkiye & Yakın Coğrafya' },
  US: { name: 'Amerika Birleşik Devletleri', region: 'Kuzey Amerika' },
  CA: { name: 'Kanada', region: 'Kuzey Amerika' },
  MX: { name: 'Meksika', region: 'Kuzey Amerika' },
  GB: { name: 'Birleşik Krallık', region: 'Avrupa' },
  DE: { name: 'Almanya', region: 'Avrupa' },
  FR: { name: 'Fransa', region: 'Avrupa' },
  NL: { name: 'Hollanda', region: 'Avrupa' },
  ES: { name: 'İspanya', region: 'Avrupa' },
  IT: { name: 'İtalya', region: 'Avrupa' },
  BE: { name: 'Belçika', region: 'Avrupa' },
  AT: { name: 'Avusturya', region: 'Avrupa' },
  CH: { name: 'İsviçre', region: 'Avrupa' },
  SE: { name: 'İsveç', region: 'Avrupa' },
  NO: { name: 'Norveç', region: 'Avrupa' },
  DK: { name: 'Danimarka', region: 'Avrupa' },
  FI: { name: 'Finlandiya', region: 'Avrupa' },
  PL: { name: 'Polonya', region: 'Avrupa' },
  CZ: { name: 'Çekya', region: 'Avrupa' },
  HU: { name: 'Macaristan', region: 'Avrupa' },
  RO: { name: 'Romanya', region: 'Avrupa' },
  BG: { name: 'Bulgaristan', region: 'Avrupa' },
  GR: { name: 'Yunanistan', region: 'Avrupa' },
  PT: { name: 'Portekiz', region: 'Avrupa' },
  IE: { name: 'İrlanda', region: 'Avrupa' },
  UA: { name: 'Ukrayna', region: 'Avrupa' },
  RU: { name: 'Rusya', region: 'Avrupa & Asya' },
  KZ: { name: 'Kazakistan', region: 'Asya' },
  UZ: { name: 'Özbekistan', region: 'Asya' },
  CN: { name: 'Çin', region: 'Asya' },
  JP: { name: 'Japonya', region: 'Asya' },
  KR: { name: 'Güney Kore', region: 'Asya' },
  IN: { name: 'Hindistan', region: 'Asya' },
  ID: { name: 'Endonezya', region: 'Asya' },
  MY: { name: 'Malezya', region: 'Asya' },
  VN: { name: 'Vietnam', region: 'Asya' },
  TH: { name: 'Tayland', region: 'Asya' },
  SA: { name: 'Suudi Arabistan', region: 'Orta Doğu' },
  AE: { name: 'Birleşik Arap Emirlikleri', region: 'Orta Doğu' },
  IR: { name: 'İran', region: 'Orta Doğu' },
  IL: { name: 'İsrail', region: 'Orta Doğu' },
  EG: { name: 'Mısır', region: 'Orta Doğu & Afrika' },
  BR: { name: 'Brezilya', region: 'Güney Amerika' },
  AR: { name: 'Arjantin', region: 'Güney Amerika' },
  CO: { name: 'Kolombiya', region: 'Güney Amerika' },
  CL: { name: 'Şili', region: 'Güney Amerika' },
  AU: { name: 'Avustralya', region: 'Okyanusya' },
  NZ: { name: 'Yeni Zelanda', region: 'Okyanusya' },
  ZA: { name: 'Güney Afrika', region: 'Afrika' }
}

function resolveCountryFromLanguage(rawLang: string | undefined | null): CountryResolution {
  const cleaned = String(rawLang ?? '')
    .split(',')[0]
    ?.split(';')[0]
    ?.trim() ?? ''

  if (!cleaned || cleaned === '-' || cleaned === '*') {
    return {
      code: '??',
      name: 'Bilinmiyor (Dil Yok)',
      region: 'Belirsiz',
      normalizedLang: '-'
    }
  }

  const parts = cleaned.replace(/_/g, '-').split('-').filter(Boolean)
  const langSubtag = (parts[0] ?? '').toLowerCase()
  let countryCode = ''

  if (parts.length >= 2) {
    const second = parts[1] ?? ''
    const third = parts[2] ?? ''
    if (/^[a-zA-Z]{2}$/.test(second)) {
      countryCode = second.toUpperCase()
    } else if (/^[a-zA-Z]{2}$/.test(third)) {
      countryCode = third.toUpperCase()
    }
  }

  if (!countryCode && langSubtag && LANG_TO_COUNTRY[langSubtag]) {
    countryCode = LANG_TO_COUNTRY[langSubtag]
  }

  if (!countryCode) {
    return {
      code: langSubtag.toUpperCase().slice(0, 2) || '??',
      name: `Diğer (${cleaned})`,
      region: 'Diğer Bölgeler',
      normalizedLang: cleaned
    }
  }

  const known = COUNTRY_NAMES_TR[countryCode]
  if (known) {
    return {
      code: countryCode,
      name: known.name,
      region: known.region,
      normalizedLang: cleaned
    }
  }

  try {
    const displayNames = new Intl.DisplayNames(['tr'], { type: 'region' })
    const intlName = displayNames.of(countryCode)
    if (intlName && intlName !== countryCode) {
      return {
        code: countryCode,
        name: intlName,
        region: 'Diğer Bölgeler',
        normalizedLang: cleaned
      }
    }
  } catch {
    // Intl.DisplayNames desteklenmiyorsa ülke kodunu kullan
  }

  return {
    code: countryCode,
    name: `Ülke (${countryCode})`,
    region: 'Diğer Bölgeler',
    normalizedLang: cleaned
  }
}

interface DistributionItem {
  label: string
  code?: string
  count: number
  percent: number
}

function buildDistribution(values: { label: string; code?: string }[]): DistributionItem[] {
  const total = values.length
  if (total === 0) return []
  const map = new Map<string, { label: string; code?: string; count: number }>()
  for (const item of values) {
    const key = item.label || 'Bilinmiyor'
    const existing = map.get(key)
    if (existing) {
      existing.count += 1
    } else {
      map.set(key, { label: key, code: item.code, count: 1 })
    }
  }
  return Array.from(map.values())
    .sort((a, b) => b.count - a.count)
    .map((entry) => ({
      label: entry.label,
      code: entry.code,
      count: entry.count,
      percent: Math.round((entry.count / total) * 1000) / 10
    }))
}

const deviceDistribution = computed<DistributionItem[]>(() =>
  buildDistribution(logs.value.map((l) => ({ label: l.deviceType || 'Bilinmiyor' })))
)

const browserDistribution = computed<DistributionItem[]>(() =>
  buildDistribution(logs.value.map((l) => ({ label: l.browser || 'Diğer' })))
)

const osDistribution = computed<DistributionItem[]>(() =>
  buildDistribution(logs.value.map((l) => ({ label: l.os || 'Diğer' })))
)

const countryDistribution = computed<DistributionItem[]>(() =>
  buildDistribution(
    logs.value.map((l) => {
      const resolved = resolveCountryFromLanguage(l.acceptLanguage)
      return { label: resolved.name, code: resolved.code }
    })
  )
)

const activeSingleDistribution = computed<{ title: string; subtitle: string; items: DistributionItem[] }>(() => {
  switch (selectedDistTab.value) {
    case 'device':
      return {
        title: 'Cihaz Türü Dağılımı',
        subtitle: 'Tıklamaların masaüstü, mobil, tablet ve bot cihazlara göre oranı.',
        items: deviceDistribution.value
      }
    case 'browser':
      return {
        title: 'Tarayıcı Dağılımı',
        subtitle: 'Ziyaretçilerin kullandığı web tarayıcılarının dağılımı.',
        items: browserDistribution.value
      }
    case 'os':
      return {
        title: 'İşletim Sistemi Dağılımı',
        subtitle: 'Tıklama yapılan cihazların işletim sistemi kırılımı.',
        items: osDistribution.value
      }
    case 'country':
      return {
        title: 'Ülke Dağılımı (Tarayıcı Dilinden)',
        subtitle: 'Tarayıcı dili (Accept-Language) verisinden çözümlenen ülke oranları.',
        items: countryDistribution.value
      }
    default:
      return {
        title: 'Tüm Kategoriler',
        subtitle: '',
        items: []
      }
  }
})

// Zamana Göre Tıklama Grafiği Verisi
interface TimeBucket {
  key: string
  label: string
  fullLabel: string
  count: number
  heightPercent: number
}

function formatDateKey(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function formatShortDayLabel(dateKey: string): string {
  const parts = dateKey.split('-')
  if (parts.length !== 3) return dateKey
  return `${parts[2]}.${parts[1]}`
}

const timeSeriesData = computed<TimeBucket[]>(() => {
  if (selectedTimeRange.value === 'hourly') {
    const counts = Array.from({ length: 24 }, () => 0)
    for (const item of logs.value) {
      // clickedAt formatı: "YYYY-MM-DD HH:mm:ss"
      const timePart = item.clickedAt?.split(' ')[1] ?? ''
      const hourStr = timePart.split(':')[0] ?? ''
      const hour = Number.parseInt(hourStr, 10)
      if (Number.isFinite(hour) && hour >= 0 && hour < 24) {
        counts[hour] = (counts[hour] ?? 0) + 1
      }
    }
    const max = Math.max(1, ...counts)
    return counts.map((count, hour) => {
      const h = String(hour).padStart(2, '0')
      return {
        key: `${h}:00`,
        label: h,
        fullLabel: `${h}:00 – ${h}:59`,
        count,
        heightPercent: count > 0 ? Math.max(8, Math.round((count / max) * 100)) : 0
      }
    })
  }

  const daysCount = selectedTimeRange.value === '7d' ? 7 : selectedTimeRange.value === '14d' ? 14 : 30

  // En güncel log tarihini veya bugünü baz al
  let anchorDate = new Date()
  if (logs.value.length > 0) {
    const firstLogDateStr = logs.value[0]?.clickedAt?.split(' ')[0] ?? ''
    if (/^\d{4}-\d{2}-\d{2}$/.test(firstLogDateStr)) {
      const parsed = new Date(`${firstLogDateStr}T12:00:00`)
      if (!Number.isNaN(parsed.getTime())) {
        anchorDate = parsed
      }
    }
  }

  const countByDay = new Map<string, number>()
  for (const item of logs.value) {
    const datePart = item.clickedAt?.split(' ')[0] ?? ''
    if (datePart) {
      countByDay.set(datePart, (countByDay.get(datePart) ?? 0) + 1)
    }
  }

  const rawBuckets: { key: string; label: string; fullLabel: string; count: number }[] = []
  for (let i = daysCount - 1; i >= 0; i--) {
    const d = new Date(anchorDate)
    d.setDate(anchorDate.getDate() - i)
    const key = formatDateKey(d)
    const count = countByDay.get(key) ?? 0
    rawBuckets.push({
      key,
      label: formatShortDayLabel(key),
      fullLabel: key,
      count
    })
  }

  const max = Math.max(1, ...rawBuckets.map((b) => b.count))
  return rawBuckets.map((b) => ({
    ...b,
    heightPercent: b.count > 0 ? Math.max(8, Math.round((b.count / max) * 100)) : 0
  }))
})

const timeSeriesSummary = computed(() => {
  const buckets = timeSeriesData.value
  const totalInPeriod = buckets.reduce((sum, b) => sum + b.count, 0)
  let peakBucket: TimeBucket | null = null
  for (const b of buckets) {
    if (!peakBucket || b.count > peakBucket.count) {
      peakBucket = b
    }
  }
  return {
    totalInPeriod,
    peakLabel: peakBucket && peakBucket.count > 0 ? `${peakBucket.fullLabel} (${peakBucket.count})` : '-'
  }
})

// Coğrafi Kutu (2. Kutu) Detaylı Verileri (Tarayıcı Dili Bazlı)
interface GeoCountryRow {
  code: string
  name: string
  region: string
  count: number
  percent: number
  languages: string[]
  lastClickedAt: string
}

const geoCountryRows = computed<GeoCountryRow[]>(() => {
  const total = logs.value.length
  if (total === 0) return []

  const map = new Map<
    string,
    {
      code: string
      name: string
      region: string
      count: number
      langSet: Set<string>
      lastClickedAt: string
    }
  >()

  for (const item of logs.value) {
    const resolved = resolveCountryFromLanguage(item.acceptLanguage)
    const existing = map.get(resolved.name)
    if (existing) {
      existing.count += 1
      existing.langSet.add(resolved.normalizedLang)
    } else {
      map.set(resolved.name, {
        code: resolved.code,
        name: resolved.name,
        region: resolved.region,
        count: 1,
        langSet: new Set([resolved.normalizedLang]),
        lastClickedAt: item.clickedAt || '-'
      })
    }
  }

  return Array.from(map.values())
    .sort((a, b) => b.count - a.count)
    .map((entry) => ({
      code: entry.code,
      name: entry.name,
      region: entry.region,
      count: entry.count,
      percent: Math.round((entry.count / total) * 1000) / 10,
      languages: Array.from(entry.langSet),
      lastClickedAt: entry.lastClickedAt
    }))
})

const geoRegionRows = computed<DistributionItem[]>(() =>
  buildDistribution(
    logs.value.map((l) => {
      const resolved = resolveCountryFromLanguage(l.acceptLanguage)
      return { label: resolved.region }
    })
  )
)

interface RawLanguageRow {
  rawLang: string
  countryCode: string
  countryName: string
  count: number
  percent: number
}

const rawLanguageRows = computed<RawLanguageRow[]>(() => {
  const total = logs.value.length
  if (total === 0) return []

  const map = new Map<string, { rawLang: string; countryCode: string; countryName: string; count: number }>()
  for (const item of logs.value) {
    const resolved = resolveCountryFromLanguage(item.acceptLanguage)
    const key = resolved.normalizedLang
    const existing = map.get(key)
    if (existing) {
      existing.count += 1
    } else {
      map.set(key, {
        rawLang: key,
        countryCode: resolved.code,
        countryName: resolved.name,
        count: 1
      })
    }
  }

  return Array.from(map.values())
    .sort((a, b) => b.count - a.count)
    .map((r) => ({
      ...r,
      percent: Math.round((r.count / total) * 1000) / 10
    }))
})

useHead({
  title: () => (link.value ? `${link.value.alias} — Link Detayı — makeiTiny` : 'Link Detayı — makeiTiny')
})

const fullShortUrl = computed(() => (link.value ? `${config.routerBase}/${link.value.alias}` : ''))
const displayShortUrl = computed(() => fullShortUrl.value.replace(/^https?:\/\//, ''))

// Şifre aktifse (mevcut şifre korunuyorsa veya yeni şifre yazılıyorsa) bekletme açılamaz.
const isPasswordActive = computed(() => {
  const hasExistingPassword = Boolean(link.value?.locked && !removePasswordFlag.value)
  const hasNewPassword = newPassword.value.trim().length > 0
  return hasExistingPassword || hasNewPassword
})

watch(isPasswordActive, (active) => {
  if (active) {
    editWaitEnabled.value = false
  }
})

const accessModeLabel = computed(() => {
  if (!link.value) return ''
  if (link.value.locked) return 'Şifreli'
  if (link.value.waitEnabled) return `Bekletmeli (${link.value.waitSeconds} sn)`
  return 'Herkese Açık'
})

async function loadLogs() {
  logsLoading.value = true
  logsError.value = ''
  try {
    logs.value = await getLinkLogs(id.value)
    if (currentLogPage.value > totalLogPages.value) {
      currentLogPage.value = totalLogPages.value
    }
  } catch (e: any) {
    logsError.value = e?.message ?? 'Tıklama logları yüklenemedi.'
  } finally {
    logsLoading.value = false
  }
}

async function loadLink() {
  pageLoading.value = true
  pageError.value = ''
  try {
    const data = await getLink(id.value)
    link.value = data
    editTarget.value = data.target
    editAlias.value = data.alias
    newPassword.value = ''
    removePasswordFlag.value = false
    editWaitEnabled.value = !data.locked && data.waitEnabled
    editWaitSeconds.value = String(data.waitSeconds > 0 ? data.waitSeconds : 5)
    await loadLogs()
  } catch (e: any) {
    pageError.value = e?.message ?? 'Link bilgileri yüklenemedi.'
  } finally {
    pageLoading.value = false
  }
}

onMounted(() => {
  loadLink()
})

async function copyShortUrl() {
  if (!fullShortUrl.value) return
  await navigator.clipboard?.writeText(fullShortUrl.value)
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}

async function handleSave() {
  if (!link.value) return
  saveError.value = ''
  saveSuccess.value = false

  if (!editTarget.value.trim()) {
    saveError.value = 'Hedef adres boş bırakılamaz.'
    return
  }

  let passwordPayload: string | undefined = undefined
  if (removePasswordFlag.value) {
    passwordPayload = ''
  } else if (newPassword.value !== '') {
    passwordPayload = newPassword.value
  }

  const finalWaitEnabled = !isPasswordActive.value && editWaitEnabled.value
  const parsedSeconds = Number.parseInt(String(editWaitSeconds.value), 10)

  if (finalWaitEnabled && (!Number.isFinite(parsedSeconds) || parsedSeconds < 1)) {
    saveError.value = 'Bekletme açıkken bekleme süresi en az 1 saniye olmalıdır.'
    return
  }

  saving.value = true
  try {
    const updated = await updateLink(link.value.id, {
      target: editTarget.value,
      alias: editAlias.value.trim(),
      password: passwordPayload,
      waitEnabled: finalWaitEnabled,
      waitSeconds: finalWaitEnabled ? parsedSeconds : 0
    })
    link.value = updated
    editTarget.value = updated.target
    editAlias.value = updated.alias
    newPassword.value = ''
    removePasswordFlag.value = false
    editWaitEnabled.value = !updated.locked && updated.waitEnabled
    editWaitSeconds.value = String(updated.waitSeconds > 0 ? updated.waitSeconds : 5)
    saveSuccess.value = true
    setTimeout(() => (saveSuccess.value = false), 2500)
  } catch (e: any) {
    saveError.value = e?.message ?? 'Link güncellenirken bir hata oluştu.'
  } finally {
    saving.value = false
  }
}

async function handleRemovePasswordNow() {
  if (!link.value) return
  saveError.value = ''
  saveSuccess.value = false
  saving.value = true
  try {
    const updated = await updateLink(link.value.id, {
      password: ''
    })
    link.value = updated
    newPassword.value = ''
    removePasswordFlag.value = false
    saveSuccess.value = true
    setTimeout(() => (saveSuccess.value = false), 2500)
  } catch (e: any) {
    saveError.value = e?.message ?? 'Şifre kaldırılamadı.'
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  if (!link.value) return
  try {
    await removeLink(link.value.id)
    await navigateTo('/dashboard/links')
  } catch (e: any) {
    saveError.value = e?.message ?? 'Link silinemedi.'
  }
}
</script>

<template>
  <div>
    <div class="mb-6">
      <NuxtLink to="/dashboard/links" class="inline-flex items-center gap-2 text-sm font-semibold text-ink-2 hover:text-ink">
        <UiIcon name="back" :size="16" /> Linklerime dön
      </NuxtLink>
    </div>

    <div v-if="pageLoading" class="flex items-center justify-center gap-3 rounded-lg border-[1.5px] border-ink bg-card p-12 shadow-hard text-ink-2">
      <span class="h-5 w-5 animate-spin rounded-full border-2 border-ink border-t-transparent" />
      <span class="font-medium">Link bilgileri yükleniyor...</span>
    </div>

    <div v-else-if="pageError" class="rounded-lg border-[1.5px] border-ink bg-card p-8 shadow-hard">
      <p class="text-accent-dark font-semibold">{{ pageError }}</p>
      <div class="mt-4">
        <UiButton to="/dashboard/links" variant="secondary">Linklerime dön</UiButton>
      </div>
    </div>

    <template v-else-if="link">
      <DashboardPageHeader
        :title="`/${link.alias}`"
        :description="link.target"
      >
        <template #actions>
          <UiButton variant="secondary" @click="copyShortUrl">
            <UiIcon :name="copied ? 'check' : 'copy'" :size="16" />
            {{ copied ? 'Kopyalandı' : 'Kısa Linki Kopyala' }}
          </UiButton>
          <a
            :href="fullShortUrl"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center justify-center gap-2 rounded-md border-[1.5px] border-ink bg-accent px-4 py-2.5 text-sm font-semibold text-white shadow-hard transition-all duration-100 hover:-translate-x-px hover:-translate-y-px hover:shadow-hard-lg"
          >
            Ziyaret Et <UiIcon name="arrow" :size="16" />
          </a>
        </template>
      </DashboardPageHeader>

      <div class="grid gap-5 sm:grid-cols-3">
        <DashboardStatTile label="Toplam tıklama" :value="link.clicks.toLocaleString('tr-TR')" />
        <DashboardStatTile label="Erişim durumu" :value="accessModeLabel" />
        <DashboardStatTile label="Oluşturulma tarihi" :value="link.createdAt" />
      </div>

      <div class="mt-8 grid gap-8 lg:grid-cols-[1.4fr_1fr]">
        <form class="flex flex-col gap-5 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8" @submit.prevent="handleSave">
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-bold">Linki Düzenle</h2>
            <UiBadge v-if="link.locked" tone="accent"><UiIcon name="lock" :size="11" /> şifreli</UiBadge>
            <UiBadge v-else-if="link.waitEnabled" tone="neutral"><UiIcon name="clock" :size="11" /> {{ link.waitSeconds }} sn bekletmeli</UiBadge>
            <UiBadge v-else tone="ok"><UiIcon name="check" :size="11" /> doğrudan</UiBadge>
          </div>

          <p v-if="saveError" class="rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
            {{ saveError }}
          </p>

          <UiField
            id="detail-target"
            v-model="editTarget"
            label="Hedef adres"
            type="text"
            placeholder="ornek.com/sayfa veya https://..."
            hint="Başında http:// veya https:// yoksa otomatik olarak https:// eklenir."
          />

          <UiField
            id="detail-alias"
            v-model="editAlias"
            label="Takma ad (Alias)"
            prefix="localhost/r/"
            placeholder="kampanya"
          />

          <div class="rounded-md border border-line bg-paper p-4">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <p class="text-sm font-semibold">Parola Koruması</p>
                <p class="text-xs text-ink-3">
                  {{ link.locked ? 'Bu link şu an parola ile korunuyor.' : 'Bu linkte şu an parola koruması yok.' }}
                </p>
              </div>
              <UiButton
                v-if="link.locked"
                type="button"
                variant="danger"
                @click="handleRemovePasswordNow"
              >
                <UiIcon name="unlock" :size="15" /> Şifreyi Kaldır
              </UiButton>
            </div>

            <div class="mt-3">
              <UiField
                id="detail-password"
                v-model="newPassword"
                :label="link.locked ? 'Parolayı Değiştir (İsteğe bağlı)' : 'Parola Ekle (İsteğe bağlı)'"
                type="password"
                :placeholder="link.locked ? 'Mevcut parolayı korumak için boş bırakın' : 'Şifre koymak için yazın'"
                autocomplete="new-password"
              />
            </div>
            <!-- TODO / Not: Şifreli linklerde şifre çözülmese bile ilk tıklamada tıklama sayılır ve log kaydı oluşturulur. İleride detaylı incelenecek. -->
            <p class="mt-2 text-[11px] text-ink-3">
              Not: Şifreli linklerde parola girilmese dahi linke tıklandığı anda tıklama sayısı ve cihaz logu kaydedilir (detaylı ayrım ileride eklenecektir).
            </p>
          </div>

          <div class="rounded-md border border-line bg-paper p-4">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-sm font-semibold">Ara Bekletme Sayfası (/w)</p>
                <p class="mt-0.5 text-xs text-ink-3">
                  <template v-if="isPasswordActive">
                    Şifre koruması açık olduğu için ara bekletme sayfası kullanılamaz. Bekletmeyi açmak için önce şifreyi kaldırın.
                  </template>
                  <template v-else>
                    Açık olduğunda ziyaretçiler hedef adrese gitmeden önce <span class="font-mono">/w/{{ link.alias }}</span> ara sayfasında seçtiğiniz süre kadar bekletilir (reklam alanı).
                  </template>
                </p>
              </div>
              <label class="inline-flex shrink-0 cursor-pointer items-center gap-2 select-none" :class="{ 'cursor-not-allowed opacity-50': isPasswordActive }">
                <input
                  v-model="editWaitEnabled"
                  type="checkbox"
                  :disabled="isPasswordActive"
                  class="h-4 w-4 accent-accent"
                />
                <span class="text-xs font-semibold">{{ editWaitEnabled && !isPasswordActive ? 'Açık' : 'Kapalı' }}</span>
              </label>
            </div>

            <div v-if="editWaitEnabled && !isPasswordActive" class="mt-3">
              <UiField
                id="detail-wait-seconds"
                v-model="editWaitSeconds"
                label="Bekleme Süresi (saniye)"
                type="number"
                placeholder="5"
                hint="Ziyaretçinin hedefe yönlendirilmeden önce bekletileceği saniye süresi."
              />
            </div>
          </div>

          <div class="flex items-center gap-3">
            <UiButton type="submit">{{ saving ? 'Kaydediliyor...' : 'Değişiklikleri Kaydet' }}</UiButton>
            <UiBadge v-if="saveSuccess" tone="ok"><UiIcon name="check" :size="12" /> kaydedildi</UiBadge>
          </div>
        </form>

        <div class="flex flex-col gap-6">
          <div class="flex flex-col gap-3 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
            <h3 class="text-lg font-bold">Bağlantı Özeti</h3>
            <div class="rounded-md border border-line bg-paper p-3">
              <p class="text-xs text-ink-3">Kısa Adres</p>
              <p class="mt-1 font-mono text-sm font-semibold break-all">{{ fullShortUrl }}</p>
            </div>
            <div class="rounded-md border border-line bg-paper p-3">
              <p class="text-xs text-ink-3">Hedef Adres</p>
              <p class="mt-1 text-sm break-all text-ink-2">{{ link.target }}</p>
            </div>
            <div class="rounded-md border border-line bg-paper p-3">
              <p class="text-xs text-ink-3">Yönlendirme Modu</p>
              <p class="mt-1 text-sm font-medium text-ink-2">
                <template v-if="link.locked">Parola Korumalı (/unlock/{{ link.alias }})</template>
                <template v-else-if="link.waitEnabled">Ara Bekletme Sayfası (/w/{{ link.alias }} — {{ link.waitSeconds }} sn)</template>
                <template v-else>Doğrudan Yönlendirme (302 Found)</template>
              </p>
            </div>
          </div>

          <div class="flex flex-col gap-4 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
            <div>
              <h3 class="text-lg font-bold text-accent-dark">Linki Sil</h3>
              <p class="mt-1 text-sm text-ink-3">
                Bu kısa linki sildiğinizde <span class="font-mono">{{ displayShortUrl }}</span> adresi artık çalışmaz.
              </p>
            </div>

            <div v-if="!confirmDelete">
              <UiButton variant="danger" @click="confirmDelete = true">
                <UiIcon name="trash" :size="16" /> Bu Linki Sil
              </UiButton>
            </div>
            <div v-else class="flex flex-wrap items-center gap-3">
              <UiButton variant="primary" @click="handleDelete">Evet, Kalıcı Olarak Sil</UiButton>
              <UiButton variant="ghost" @click="confirmDelete = false">Vazgeç</UiButton>
            </div>
          </div>
        </div>
      </div>

      <!-- Tıklama Logları Bölümü (Pagination: 20 Kayıt / Sayfa) -->
      <div class="mt-8 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div class="flex flex-wrap items-center gap-2.5">
              <h2 class="text-xl font-bold">Tıklama Logları</h2>
              <UiBadge v-if="logs.length > 0" tone="neutral">
                Toplam {{ logs.length }} kayıt • Sayfa başına {{ LOGS_PER_PAGE }}
              </UiBadge>
            </div>
            <p class="mt-1 text-xs text-ink-3">
              Bu kısa linke tıklayan cihazların IP, tarayıcı, işletim sistemi ve yönlendiren kaynak kayıtları.
            </p>
          </div>
          <UiButton variant="secondary" @click="loadLogs">
            <UiIcon name="refresh" :size="15" /> {{ logsLoading ? 'Yükleniyor...' : 'Logları Yenile' }}
          </UiButton>
        </div>

        <p v-if="logsError" class="mt-4 rounded-md border border-accent-dark bg-accent-soft px-3 py-2 text-sm text-accent-dark">
          {{ logsError }}
        </p>

        <div v-if="logsLoading && logs.length === 0" class="mt-6 flex items-center justify-center gap-3 rounded-md border border-line bg-paper p-8 text-ink-2">
          <span class="h-4 w-4 animate-spin rounded-full border-2 border-ink border-t-transparent" />
          <span class="text-sm font-medium">Tıklama logları yükleniyor...</span>
        </div>

        <div v-else-if="logs.length === 0" class="mt-6 rounded-md border border-line bg-paper p-8 text-center">
          <p class="text-sm font-semibold text-ink-2">Henüz tıklama logu bulunmuyor.</p>
          <p class="mt-1 text-xs text-ink-3">Kısa link ziyaret edildiğinde cihaz bilgileri burada listelenecektir.</p>
        </div>

        <template v-else>
          <ul class="mt-6 flex flex-col gap-3">
            <li
              v-for="item in paginatedLogs"
              :key="item.id"
              class="flex flex-col gap-2.5 rounded-md border border-line bg-paper p-4 transition-colors hover:bg-paper-2/60"
            >
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-mono text-sm font-bold text-ink">{{ item.ipAddress || 'Bilinmiyor' }}</span>
                  <UiBadge tone="accent">{{ item.deviceType || 'Cihaz' }}</UiBadge>
                  <UiBadge tone="neutral">{{ item.browser || 'Tarayıcı' }}</UiBadge>
                  <UiBadge tone="ok">{{ item.os || 'İS' }}</UiBadge>
                </div>
                <span class="font-mono text-xs text-ink-3">{{ item.clickedAt }}</span>
              </div>

              <div class="grid gap-2 text-xs text-ink-2 sm:grid-cols-2">
                <div>
                  <span class="font-semibold text-ink-3">Kaynak (Referer): </span>
                  <span class="break-all">{{ item.referer || 'Doğrudan' }}</span>
                </div>
                <div>
                  <span class="font-semibold text-ink-3">Tarayıcı Dili: </span>
                  <span class="font-mono">{{ item.acceptLanguage || '-' }}</span>
                  <span class="ml-1.5 text-ink-3">
                    ({{ resolveCountryFromLanguage(item.acceptLanguage).name }})
                  </span>
                </div>
              </div>

              <p v-if="item.userAgent" class="truncate font-mono text-[11px] text-ink-3" :title="item.userAgent">
                UA: {{ item.userAgent }}
              </p>
            </li>
          </ul>

          <!-- Pagination Kontrolleri -->
          <div class="mt-6 flex flex-wrap items-center justify-between gap-4 rounded-md border border-line bg-paper px-4 py-3">
            <p class="text-xs text-ink-2">
              Toplam <span class="font-mono font-bold text-ink">{{ logs.length }}</span> kayıttan
              <span class="font-mono font-bold text-ink">{{ logRangeStart }}–{{ logRangeEnd }}</span> arası gösteriliyor
              <span class="text-ink-3">(Sayfa {{ currentLogPage }} / {{ totalLogPages }})</span>
            </p>

            <div class="flex flex-wrap items-center gap-1.5">
              <button
                type="button"
                :disabled="currentLogPage <= 1"
                class="inline-flex items-center gap-1 rounded-sm border border-ink bg-card px-2.5 py-1 text-xs font-semibold text-ink transition-colors hover:bg-paper-2 disabled:cursor-not-allowed disabled:opacity-40 cursor-pointer"
                @click="goToLogPage(currentLogPage - 1)"
              >
                ← Önceki
              </button>

              <template v-for="(pageItem, idx) in visibleLogPages" :key="`${pageItem}-${idx}`">
                <span
                  v-if="pageItem === '...'"
                  class="px-1.5 font-mono text-xs text-ink-3 select-none"
                >
                  ...
                </span>
                <button
                  v-else
                  type="button"
                  class="inline-flex h-7 min-w-7 items-center justify-center rounded-sm border px-2 font-mono text-xs font-bold transition-colors cursor-pointer"
                  :class="
                    pageItem === currentLogPage
                      ? 'border-ink bg-accent text-white'
                      : 'border-line bg-card text-ink-2 hover:border-ink hover:text-ink'
                  "
                  @click="goToLogPage(pageItem)"
                >
                  {{ pageItem }}
                </button>
              </template>

              <button
                type="button"
                :disabled="currentLogPage >= totalLogPages"
                class="inline-flex items-center gap-1 rounded-sm border border-ink bg-card px-2.5 py-1 text-xs font-semibold text-ink transition-colors hover:bg-paper-2 disabled:cursor-not-allowed disabled:opacity-40 cursor-pointer"
                @click="goToLogPage(currentLogPage + 1)"
              >
                Sonraki →
              </button>
            </div>
          </div>
        </template>
      </div>

      <!-- 1. KUTU: Grafiksel Log Okuma (Zamana Göre Grafik + Cihaz, Tarayıcı, İS, Ülke Dağılımı) -->
      <div class="mt-8 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <span class="eyebrow">1. Kutu • Grafik Görünümü</span>
            <h2 class="mt-1 text-xl font-bold">Grafiksel Log Analizi</h2>
            <p class="mt-1 text-xs text-ink-3">
              Zamana göre tıklama grafiği ve cihaz, tarayıcı, işletim sistemi ile ülke bazlı dağılım grafikleri.
            </p>
          </div>
          <UiBadge tone="accent">
            <UiIcon name="chart" :size="12" /> {{ logs.length }} log inceleniyor
          </UiBadge>
        </div>

        <div v-if="logs.length === 0" class="mt-6 rounded-md border border-line bg-paper p-8 text-center">
          <p class="text-sm font-semibold text-ink-2">Grafik oluşturmak için henüz tıklama verisi yok.</p>
          <p class="mt-1 text-xs text-ink-3">Link ziyaret edildiğinde zaman ve kategori grafikleri otomatik oluşturulur.</p>
        </div>

        <template v-else>
          <!-- A) Zamana Göre Tıklama Grafiği -->
          <div class="mt-6 rounded-md border border-line bg-paper p-5">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h3 class="text-base font-bold">Zamana Göre Tıklama Grafiği</h3>
                <p class="mt-0.5 text-xs text-ink-3">
                  Seçili periyotta gerçekleşen toplam tıklama:
                  <span class="font-mono font-bold text-ink">{{ timeSeriesSummary.totalInPeriod }}</span>
                  • En yoğun dilim:
                  <span class="font-mono font-semibold text-ink-2">{{ timeSeriesSummary.peakLabel }}</span>
                </p>
              </div>

              <div class="flex flex-wrap items-center gap-1.5">
                <button
                  v-for="opt in [
                    { id: '7d', label: 'Son 7 Gün' },
                    { id: '14d', label: 'Son 14 Gün' },
                    { id: '30d', label: 'Son 30 Gün' },
                    { id: 'hourly', label: 'Saatlik (00-23)' }
                  ]"
                  :key="opt.id"
                  type="button"
                  class="rounded-sm border px-2.5 py-1 text-xs font-semibold transition-colors cursor-pointer"
                  :class="
                    selectedTimeRange === opt.id
                      ? 'border-ink bg-accent text-white'
                      : 'border-line bg-card text-ink-2 hover:border-ink hover:text-ink'
                  "
                  @click="selectedTimeRange = opt.id as any"
                >
                  {{ opt.label }}
                </button>
              </div>
            </div>

            <!-- Sütun (Bar) Grafiği -->
            <div class="mt-5 overflow-x-auto pb-2">
              <div
                class="grid h-48 items-end gap-1.5 rounded-md border border-line bg-card px-3 pt-6 pb-2"
                :style="{
                  gridTemplateColumns: `repeat(${timeSeriesData.length}, minmax(26px, 1fr))`,
                  minWidth: timeSeriesData.length > 14 ? '680px' : '100%'
                }"
              >
                <div
                  v-for="bucket in timeSeriesData"
                  :key="bucket.key"
                  class="group flex h-full flex-col items-center justify-end"
                  :title="`${bucket.fullLabel}: ${bucket.count} tıklama`"
                >
                  <span
                    class="mb-1 font-mono text-[10px] font-bold transition-colors"
                    :class="bucket.count > 0 ? 'text-ink' : 'text-ink-3/50'"
                  >
                    {{ bucket.count }}
                  </span>
                  <div class="flex w-full flex-1 items-end justify-center">
                    <div
                      class="w-full max-w-8 rounded-t-sm border transition-all duration-200"
                      :class="
                        bucket.count > 0
                          ? 'border-ink bg-accent group-hover:bg-accent-dark'
                          : 'border-line bg-paper-2'
                      "
                      :style="{ height: bucket.count > 0 ? `${bucket.heightPercent}%` : '4px' }"
                    />
                  </div>
                  <span class="mt-1.5 font-mono text-[10px] text-ink-3">
                    {{ bucket.label }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- B) Cihaz, Tarayıcı, İS ve Ülke Dağılım Grafikleri -->
          <div class="mt-6">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h3 class="text-base font-bold">Kategoriye Göre Dağılım Grafikleri</h3>
                <p class="mt-0.5 text-xs text-ink-3">
                  Cihaz, tarayıcı, işletim sistemi ve ülke kırılımlarını toplu veya tekil grafik olarak inceleyin.
                </p>
              </div>

              <div class="flex flex-wrap items-center gap-1.5">
                <button
                  v-for="tab in [
                    { id: 'all', label: 'Tümü (4 Kategori)' },
                    { id: 'device', label: 'Cihaz' },
                    { id: 'browser', label: 'Tarayıcı' },
                    { id: 'os', label: 'İşletim Sistemi' },
                    { id: 'country', label: 'Ülke' }
                  ]"
                  :key="tab.id"
                  type="button"
                  class="rounded-sm border px-2.5 py-1 text-xs font-semibold transition-colors cursor-pointer"
                  :class="
                    selectedDistTab === tab.id
                      ? 'border-ink bg-ink text-white'
                      : 'border-line bg-paper text-ink-2 hover:border-ink hover:text-ink'
                  "
                  @click="selectedDistTab = tab.id as any"
                >
                  {{ tab.label }}
                </button>
              </div>
            </div>

            <!-- 4'lü Toplu Görünüm -->
            <div v-if="selectedDistTab === 'all'" class="mt-4 grid gap-4 md:grid-cols-2">
              <!-- Cihaz Dağılımı -->
              <div class="rounded-md border border-line bg-paper p-4">
                <div class="flex items-center justify-between">
                  <h4 class="text-sm font-bold">Cihaz Dağılımı</h4>
                  <UiBadge tone="accent">{{ deviceDistribution.length }} tür</UiBadge>
                </div>
                <div class="mt-3 flex flex-col gap-2.5">
                  <div v-for="item in deviceDistribution" :key="item.label" class="flex flex-col gap-1">
                    <div class="flex items-center justify-between text-xs">
                      <span class="font-semibold text-ink">{{ item.label }}</span>
                      <span class="font-mono text-ink-2">{{ item.count }} (%{{ item.percent }})</span>
                    </div>
                    <div class="h-2.5 w-full overflow-hidden rounded-sm border border-ink/40 bg-card">
                      <div class="h-full bg-accent transition-all duration-300" :style="{ width: `${item.percent}%` }" />
                    </div>
                  </div>
                </div>
              </div>

              <!-- Tarayıcı Dağılımı -->
              <div class="rounded-md border border-line bg-paper p-4">
                <div class="flex items-center justify-between">
                  <h4 class="text-sm font-bold">Tarayıcı Dağılımı</h4>
                  <UiBadge tone="neutral">{{ browserDistribution.length }} tarayıcı</UiBadge>
                </div>
                <div class="mt-3 flex flex-col gap-2.5">
                  <div v-for="item in browserDistribution" :key="item.label" class="flex flex-col gap-1">
                    <div class="flex items-center justify-between text-xs">
                      <span class="font-semibold text-ink">{{ item.label }}</span>
                      <span class="font-mono text-ink-2">{{ item.count }} (%{{ item.percent }})</span>
                    </div>
                    <div class="h-2.5 w-full overflow-hidden rounded-sm border border-ink/40 bg-card">
                      <div class="h-full bg-ink transition-all duration-300" :style="{ width: `${item.percent}%` }" />
                    </div>
                  </div>
                </div>
              </div>

              <!-- İşletim Sistemi Dağılımı -->
              <div class="rounded-md border border-line bg-paper p-4">
                <div class="flex items-center justify-between">
                  <h4 class="text-sm font-bold">İşletim Sistemi Dağılımı</h4>
                  <UiBadge tone="ok">{{ osDistribution.length }} İS</UiBadge>
                </div>
                <div class="mt-3 flex flex-col gap-2.5">
                  <div v-for="item in osDistribution" :key="item.label" class="flex flex-col gap-1">
                    <div class="flex items-center justify-between text-xs">
                      <span class="font-semibold text-ink">{{ item.label }}</span>
                      <span class="font-mono text-ink-2">{{ item.count }} (%{{ item.percent }})</span>
                    </div>
                    <div class="h-2.5 w-full overflow-hidden rounded-sm border border-ink/40 bg-card">
                      <div class="h-full bg-ok transition-all duration-300" :style="{ width: `${item.percent}%` }" />
                    </div>
                  </div>
                </div>
              </div>

              <!-- Ülke Dağılımı -->
              <div class="rounded-md border border-line bg-paper p-4">
                <div class="flex items-center justify-between">
                  <h4 class="text-sm font-bold">Ülke Dağılımı (Dil Bazlı)</h4>
                  <UiBadge tone="accent">{{ countryDistribution.length }} ülke</UiBadge>
                </div>
                <div class="mt-3 flex flex-col gap-2.5">
                  <div v-for="item in countryDistribution.slice(0, 6)" :key="item.label" class="flex flex-col gap-1">
                    <div class="flex items-center justify-between text-xs">
                      <span class="font-semibold text-ink">
                        <span v-if="item.code" class="mr-1 rounded-xs bg-paper-2 px-1 py-0.5 font-mono text-[10px]">{{ item.code }}</span>
                        {{ item.label }}
                      </span>
                      <span class="font-mono text-ink-2">{{ item.count }} (%{{ item.percent }})</span>
                    </div>
                    <div class="h-2.5 w-full overflow-hidden rounded-sm border border-ink/40 bg-card">
                      <div class="h-full bg-accent-dark transition-all duration-300" :style="{ width: `${item.percent}%` }" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Tekil Kategori Detaylı Sütun + Bar Görünümü -->
            <div v-else class="mt-4 rounded-md border border-line bg-paper p-5">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <h4 class="text-sm font-bold">{{ activeSingleDistribution.title }}</h4>
                  <p class="text-xs text-ink-3">{{ activeSingleDistribution.subtitle }}</p>
                </div>
                <UiBadge tone="neutral">{{ activeSingleDistribution.items.length }} grup</UiBadge>
              </div>

              <div class="mt-5 grid gap-6 lg:grid-cols-[1fr_1.2fr]">
                <!-- Dikey Sütun Karşılaştırma Grafiği -->
                <div class="flex flex-col justify-between rounded-md border border-line bg-card p-4">
                  <p class="text-xs font-semibold text-ink-3">Oransal Sütun Grafiği</p>
                  <div class="mt-4 flex h-44 items-end justify-around gap-2 pt-4">
                    <div
                      v-for="item in activeSingleDistribution.items.slice(0, 8)"
                      :key="item.label"
                      class="flex h-full flex-1 flex-col items-center justify-end"
                    >
                      <span class="mb-1 font-mono text-[10px] font-bold text-ink">%{{ item.percent }}</span>
                      <div class="flex w-full flex-1 items-end justify-center">
                        <div
                          class="w-full max-w-10 rounded-t-sm border border-ink bg-accent transition-all duration-300"
                          :style="{ height: `${Math.max(10, item.percent)}%` }"
                        />
                      </div>
                      <span class="mt-1.5 max-w-20 truncate font-mono text-[10px] text-ink-2" :title="item.label">
                        {{ item.code || item.label }}
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Detaylı Liste -->
                <div class="flex flex-col gap-3">
                  <div
                    v-for="item in activeSingleDistribution.items"
                    :key="item.label"
                    class="rounded-md border border-line bg-card p-3"
                  >
                    <div class="flex items-center justify-between text-xs">
                      <span class="font-semibold text-ink">
                        <span v-if="item.code" class="mr-1.5 rounded-xs border border-line bg-paper px-1.5 py-0.5 font-mono text-[10px]">
                          {{ item.code }}
                        </span>
                        {{ item.label }}
                      </span>
                      <span class="font-mono font-bold text-ink">
                        {{ item.count }} tıklama (%{{ item.percent }})
                      </span>
                    </div>
                    <div class="mt-2 h-2.5 w-full overflow-hidden rounded-sm border border-ink/40 bg-paper">
                      <div class="h-full bg-accent transition-all duration-300" :style="{ width: `${item.percent}%` }" />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- 2. KUTU: Coğrafi Log Okuma (Tarayıcı Dili / Accept-Language Bazlı) -->
      <div class="mt-8 rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard sm:p-8">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <span class="eyebrow">2. Kutu • Coğrafi Görünüm</span>
            <h2 class="mt-1 text-xl font-bold">Coğrafi Dağılım (Tarayıcı Dili Bazlı)</h2>
            <p class="mt-1 text-xs text-ink-3">
              Ziyaretçilerin tarayıcı dili (<span class="font-mono">Accept-Language</span>) bilgisinden ülke ve bölge çıkarımı yapılarak oluşturulan coğrafi okuma.
            </p>
          </div>
          <UiBadge tone="ok">
            {{ geoCountryRows.length }} ülke / {{ rawLanguageRows.length }} dil kodu
          </UiBadge>
        </div>

        <div v-if="logs.length === 0" class="mt-6 rounded-md border border-line bg-paper p-8 text-center">
          <p class="text-sm font-semibold text-ink-2">Coğrafi dağılım için henüz tıklama logu bulunmuyor.</p>
          <p class="mt-1 text-xs text-ink-3">Ziyaretçilerin tarayıcı dili kaydedildiğinde ülke ve bölge analizleri burada listelenir.</p>
        </div>

        <template v-else>
          <!-- Coğrafi Özet Kartları -->
          <div class="mt-6 grid gap-4 sm:grid-cols-3">
            <div class="rounded-md border border-line bg-paper p-4">
              <p class="text-xs text-ink-3">Tespit Edilen Ülke Sayısı</p>
              <p class="mt-1 font-display text-2xl font-extrabold text-ink">{{ geoCountryRows.length }}</p>
              <p class="mt-0.5 text-[11px] text-ink-3">{{ geoRegionRows.length }} farklı coğrafi bölge</p>
            </div>
            <div class="rounded-md border border-line bg-paper p-4">
              <p class="text-xs text-ink-3">En Yoğun Ülke</p>
              <p class="mt-1 font-display text-2xl font-extrabold text-ink">
                {{ geoCountryRows[0]?.name ?? '-' }}
              </p>
              <p class="mt-0.5 font-mono text-[11px] text-ink-3">
                <template v-if="geoCountryRows[0]">
                  {{ geoCountryRows[0].count }} tıklama (%{{ geoCountryRows[0].percent }})
                </template>
                <template v-else>-</template>
              </p>
            </div>
            <div class="rounded-md border border-line bg-paper p-4">
              <p class="text-xs text-ink-3">Farklı Tarayıcı Dili Kodu</p>
              <p class="mt-1 font-display text-2xl font-extrabold text-ink">{{ rawLanguageRows.length }}</p>
              <p class="mt-0.5 font-mono text-[11px] text-ink-3">
                Örn: {{ rawLanguageRows.slice(0, 3).map((r) => r.rawLang).join(', ') }}
              </p>
            </div>
          </div>

          <div class="mt-6 grid gap-6 lg:grid-cols-[1.35fr_1fr]">
            <!-- Sol: Ülke Bazlı Coğrafi Okuma Listesi -->
            <div class="rounded-md border border-line bg-paper p-5">
              <div class="flex items-center justify-between">
                <h3 class="text-base font-bold">Ülke Bazlı Coğrafi Dağılım</h3>
                <span class="font-mono text-xs text-ink-3">Toplam {{ logs.length }} tıklama</span>
              </div>

              <ul class="mt-4 flex flex-col gap-3">
                <li
                  v-for="(row, idx) in geoCountryRows"
                  :key="row.name"
                  class="rounded-md border border-line bg-card p-3.5"
                >
                  <div class="flex flex-wrap items-center justify-between gap-2">
                    <div class="flex items-center gap-2.5">
                      <span class="inline-flex h-6 w-6 items-center justify-center rounded-sm border border-ink bg-paper-2 font-mono text-xs font-bold text-ink">
                        {{ idx + 1 }}
                      </span>
                      <span class="rounded-sm border border-ink bg-accent-soft px-2 py-0.5 font-mono text-xs font-bold text-accent-dark">
                        {{ row.code }}
                      </span>
                      <div>
                        <p class="text-sm font-bold text-ink">{{ row.name }}</p>
                        <p class="text-[11px] text-ink-3">{{ row.region }}</p>
                      </div>
                    </div>

                    <div class="text-right">
                      <span class="font-mono text-sm font-bold text-ink">{{ row.count }} tıklama</span>
                      <span class="ml-1.5 rounded-sm bg-paper-2 px-1.5 py-0.5 font-mono text-xs font-semibold text-ink-2">
                        %{{ row.percent }}
                      </span>
                    </div>
                  </div>

                  <div class="mt-2.5 h-2.5 w-full overflow-hidden rounded-sm border border-ink/40 bg-paper">
                    <div
                      class="h-full bg-accent transition-all duration-300"
                      :style="{ width: `${row.percent}%` }"
                    />
                  </div>

                  <div class="mt-2 flex flex-wrap items-center justify-between gap-2 text-[11px] text-ink-3">
                    <div>
                      <span class="font-semibold">Dil Kodları: </span>
                      <span class="font-mono text-ink-2">{{ row.languages.join(', ') }}</span>
                    </div>
                    <div class="font-mono">
                      Son: {{ row.lastClickedAt }}
                    </div>
                  </div>
                </li>
              </ul>
            </div>

            <!-- Sağ: Bölgesel Yoğunluk ve Ham Dil Kodu Eşleşmeleri -->
            <div class="flex flex-col gap-5">
              <!-- Bölgesel Yoğunluk -->
              <div class="rounded-md border border-line bg-paper p-5">
                <h3 class="text-base font-bold">Bölgesel Yoğunluk</h3>
                <p class="mt-0.5 text-xs text-ink-3">Coğrafi bölgelere göre tıklama dağılımı.</p>

                <div class="mt-4 flex flex-col gap-3">
                  <div v-for="reg in geoRegionRows" :key="reg.label" class="flex flex-col gap-1">
                    <div class="flex items-center justify-between text-xs">
                      <span class="font-semibold text-ink">{{ reg.label }}</span>
                      <span class="font-mono text-ink-2">{{ reg.count }} (%{{ reg.percent }})</span>
                    </div>
                    <div class="h-2.5 w-full overflow-hidden rounded-sm border border-ink/40 bg-card">
                      <div class="h-full bg-ok transition-all duration-300" :style="{ width: `${reg.percent}%` }" />
                    </div>
                  </div>
                </div>
              </div>

              <!-- Tarayıcı Dili -> Ülke Çözümleme Tablosu -->
              <div class="rounded-md border border-line bg-paper p-5">
                <h3 class="text-base font-bold">Tarayıcı Dili → Ülke Eşleşmeleri</h3>
                <p class="mt-0.5 text-xs text-ink-3">
                  Loglardaki ham <span class="font-mono">acceptLanguage</span> değerlerinin ülke karşılıkları.
                </p>

                <div class="mt-4 overflow-x-auto">
                  <table class="w-full border-collapse text-left text-xs">
                    <thead>
                      <tr class="border-b border-line text-ink-3">
                        <th class="pb-2 font-semibold">Dil Kodu</th>
                        <th class="pb-2 font-semibold">Çözümlenen Ülke</th>
                        <th class="pb-2 text-right font-semibold">Adet / Oran</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-line">
                      <tr v-for="langRow in rawLanguageRows" :key="langRow.rawLang" class="hover:bg-card/60">
                        <td class="py-2 font-mono font-bold text-ink">{{ langRow.rawLang }}</td>
                        <td class="py-2 text-ink-2">
                          <span class="mr-1 rounded-xs border border-line bg-card px-1 py-0.5 font-mono text-[10px]">
                            {{ langRow.countryCode }}
                          </span>
                          {{ langRow.countryName }}
                        </td>
                        <td class="py-2 text-right font-mono text-ink">
                          {{ langRow.count }} <span class="text-ink-3">(%{{ langRow.percent }})</span>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>
    </template>
  </div>
</template>
