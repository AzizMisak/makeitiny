import type { LinkLogItem } from '~/types/link'

export interface CountryResolution {
  code: string
  name: string
  region: string
  lang: string
}

const LANG_TO_COUNTRY: Record<string, string> = {
  tr: 'TR', en: 'US', de: 'DE', fr: 'FR', es: 'ES', it: 'IT', nl: 'NL', ru: 'RU', az: 'AZ', ar: 'SA',
  pt: 'PT', zh: 'CN', ja: 'JP', ko: 'KR', pl: 'PL', uk: 'UA', el: 'GR', bg: 'BG', ro: 'RO', sv: 'SE',
  no: 'NO', nb: 'NO', nn: 'NO', da: 'DK', fi: 'FI', cs: 'CZ', hu: 'HU', sk: 'SK', hr: 'HR', sr: 'RS',
  fa: 'IR', hi: 'IN', id: 'ID', ms: 'MY', vi: 'VN', th: 'TH', he: 'IL', ca: 'ES', ka: 'GE', kk: 'KZ', uz: 'UZ'
}

const REGIONS: Record<string, string[]> = {
  'Türkiye ve yakın çevre': ['TR', 'AZ', 'CY', 'GE'],
  'Kuzey Amerika': ['US', 'CA', 'MX'],
  'Güney Amerika': ['BR', 'AR', 'CO', 'CL'],
  Avrupa: [
    'GB', 'DE', 'FR', 'NL', 'ES', 'IT', 'BE', 'AT', 'CH', 'SE', 'NO', 'DK', 'FI', 'PL', 'CZ', 'HU', 'RO', 'BG',
    'GR', 'PT', 'IE', 'UA', 'RU', 'SK', 'HR', 'RS'
  ],
  Asya: ['KZ', 'UZ', 'CN', 'JP', 'KR', 'IN', 'ID', 'MY', 'VN', 'TH'],
  'Orta Doğu ve Afrika': ['SA', 'AE', 'IR', 'IL', 'EG', 'ZA'],
  Okyanusya: ['AU', 'NZ']
}

const UNKNOWN = 'Bilinmiyor'

let regionNames: Intl.DisplayNames | null = null

export function countryName(code: string): string {
  try {
    regionNames ??= new Intl.DisplayNames(['tr'], { type: 'region' })
    return regionNames.of(code) ?? code
  } catch {
    return code
  }
}

function regionOf(code: string): string {
  for (const [region, codes] of Object.entries(REGIONS)) {
    if (codes.includes(code)) return region
  }
  return 'Diğer bölgeler'
}

/** Ziyaretin ülkesi: kayıtta ülke kodu varsa o, yoksa tarayıcı dilinden tahmin. */
export function resolveCountry(log: Pick<LinkLogItem, 'countryCode' | 'acceptLanguage'>): CountryResolution {
  const lang = (log.acceptLanguage ?? '').split(',')[0]?.split(';')[0]?.trim() ?? ''
  const known = /^[A-Za-z]{2}$/.test(log.countryCode ?? '') ? log.countryCode.toUpperCase() : ''

  let code = known
  if (!code && lang && lang !== '*' && lang !== '-') {
    const parts = lang.replace(/_/g, '-').split('-')
    const region = parts.slice(1).find((p) => /^[A-Za-z]{2}$/.test(p))
    code = region?.toUpperCase() ?? LANG_TO_COUNTRY[(parts[0] ?? '').toLowerCase()] ?? ''
  }

  if (!code) return { code: '', name: UNKNOWN, region: UNKNOWN, lang }
  return { code, name: countryName(code), region: regionOf(code), lang }
}
