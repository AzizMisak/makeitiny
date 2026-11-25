import type { LinkItem } from '~/types/link'
import type { Tone } from './admin'

export interface LinkState {
  tone: Tone
  label: string
  /** Ziyaretçinin hedefe ulaşıp ulaşamadığı */
  reachable: boolean
  note: string
}

/** Linkin ziyaretçi gözünden gerçek durumu. Router'ın denetim sırasıyla aynıdır. */
export function linkState(l: LinkItem): LinkState {
  if (!l.isActive) return { tone: 'neutral', label: 'Kapalı', reachable: false, note: 'Yönlendirme kapalı.' }
  if (isBlockedStatus(l.securityStatus)) {
    return { tone: 'danger', label: 'Engellendi', reachable: false, note: 'Yönetim bu linki güvenlik nedeniyle engelledi.' }
  }
  if (l.expiresAt && new Date(`${l.expiresAt}T23:59:59Z`).getTime() <= Date.now()) {
    return { tone: 'warn', label: 'Süresi doldu', reachable: false, note: 'Son kullanma tarihi geçti.' }
  }
  if (l.maxClicks && l.clicks >= l.maxClicks) {
    return { tone: 'warn', label: 'Sınır doldu', reachable: false, note: 'Tıklama sınırına ulaştı.' }
  }
  return { tone: 'ok', label: 'Yayında', reachable: true, note: 'Ziyaretçiler hedefe ulaşıyor.' }
}
