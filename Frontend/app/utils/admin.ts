import type { RoleName } from '~/types/auth'
import type { SecurityStatus } from '~/types/link'

export type Tone = 'neutral' | 'brand' | 'ok' | 'warn' | 'danger' | 'mark'

export function roleMeta(roles?: RoleName[]): { primary: RoleName; label: string; tone: Tone } {
  if (roles?.includes('owner')) return { primary: 'owner', label: 'Sahip', tone: 'mark' }
  if (roles?.includes('admin')) return { primary: 'admin', label: 'Yönetici', tone: 'brand' }
  return { primary: 'user', label: 'Kullanıcı', tone: 'neutral' }
}

export interface SecurityOption {
  value: SecurityStatus
  label: string
  tone: Tone
  icon: string
  /** Yöneticiye kararın sonucunu anlatır. */
  effect: string
  /** Karar bu durumdayken router yönlendirmeyi reddeder mi? */
  blocks: boolean
}

export const SECURITY_OPTIONS: SecurityOption[] = [
  {
    value: 'unchecked',
    label: 'İncelenmedi',
    tone: 'neutral',
    icon: 'circle',
    effect: 'Karar verilmemiş sayılır. Yönlendirme sürer.',
    blocks: false
  },
  {
    value: 'safe',
    label: 'Güvenli',
    tone: 'ok',
    icon: 'shield-check',
    effect: 'İncelendi, sorun yok. Yönlendirme sürer.',
    blocks: false
  },
  {
    value: 'suspicious',
    label: 'Şüpheli',
    tone: 'warn',
    icon: 'alert',
    effect: 'Yalnızca işaretlenir. Yönlendirme sürer.',
    blocks: false
  },
  {
    value: 'malicious',
    label: 'Zararlı',
    tone: 'danger',
    icon: 'shield-alert',
    effect: 'Zararlı içerik. Ziyaretçi güvenlik uyarısı görür.',
    blocks: true
  },
  {
    value: 'blocked',
    label: 'Engelli',
    tone: 'danger',
    icon: 'ban',
    effect: 'Kural ihlali. Ziyaretçi güvenlik uyarısı görür.',
    blocks: true
  }
]

export function securityMeta(status: SecurityStatus): SecurityOption {
  return SECURITY_OPTIONS.find((o) => o.value === status) ?? SECURITY_OPTIONS[0]!
}

export function isBlockedStatus(status: SecurityStatus): boolean {
  return securityMeta(status).blocks
}
