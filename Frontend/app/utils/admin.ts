import type { RoleName } from '~/types/auth'
import type { SecurityStatus } from '~/types/link'

export type BadgeTone = 'neutral' | 'accent' | 'ok'

export function formatDateTr(raw?: string): string {
  if (!raw) return '—'
  const dt = new Date(raw)
  if (Number.isNaN(dt.getTime())) return raw
  return dt.toLocaleDateString('tr-TR')
}

export function formatDateTimeTr(raw?: string): string {
  if (!raw) return '—'
  const dt = new Date(raw)
  if (Number.isNaN(dt.getTime())) return raw
  return dt.toLocaleString('tr-TR')
}

export function formatPriceMonthly(price: number): string {
  if (!price || price <= 0) return 'Ücretsiz'
  return `${price.toLocaleString('tr-TR')} TL / ay`
}

export function roleBadgeMeta(roles?: RoleName[]): { primary: RoleName; label: string; tone: BadgeTone } {
  if (roles?.includes('owner')) return { primary: 'owner', label: 'OWNER', tone: 'accent' }
  if (roles?.includes('admin')) return { primary: 'admin', label: 'ADMİN', tone: 'ok' }
  return { primary: 'user', label: 'KULLANICI', tone: 'neutral' }
}

export const SECURITY_STATUS_OPTIONS: { value: SecurityStatus; label: string }[] = [
  { value: 'unchecked', label: 'İncelenmedi' },
  { value: 'safe', label: 'Güvenli' },
  { value: 'suspicious', label: 'Şüpheli' },
  { value: 'malicious', label: 'Zararlı' },
  { value: 'blocked', label: 'Engelli' }
]

export function securityStatusMeta(status: SecurityStatus): { label: string; tone: BadgeTone } {
  switch (status) {
    case 'safe':
      return { label: 'Güvenli', tone: 'ok' }
    case 'suspicious':
      return { label: 'Şüpheli', tone: 'accent' }
    case 'malicious':
      return { label: 'Zararlı', tone: 'accent' }
    case 'blocked':
      return { label: 'Engelli', tone: 'accent' }
    default:
      return { label: 'İncelenmedi', tone: 'neutral' }
  }
}
