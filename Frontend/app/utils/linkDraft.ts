import type { ConfirmChange } from '~/composables/useConfirm'
import type { CreateLinkPayload, LinkItem, UpdateLinkPayload } from '~/types/link'

export type LinkAccess = 'open' | 'password' | 'wait'

/** Link oluşturma ve düzenleme formunun ortak taslağı. */
export interface LinkDraft {
  target: string
  title: string
  alias: string
  access: LinkAccess
  password: string
  waitSeconds: string
  isActive: boolean
  expires: boolean
  expiresAt: string
  limited: boolean
  maxClicks: string
  adEnabled: boolean
}

export type LinkDraftErrors = Partial<Record<'target' | 'alias' | 'password' | 'waitSeconds' | 'expiresAt' | 'maxClicks', string>>

export function emptyDraft(): LinkDraft {
  return {
    target: '',
    title: '',
    alias: '',
    access: 'open',
    password: '',
    waitSeconds: '5',
    isActive: true,
    expires: false,
    expiresAt: '',
    limited: false,
    maxClicks: '',
    adEnabled: false
  }
}

export function draftFromLink(link: LinkItem): LinkDraft {
  return {
    target: link.target,
    title: link.title ?? '',
    alias: link.alias,
    access: link.locked ? 'password' : link.waitEnabled ? 'wait' : 'open',
    password: '',
    waitSeconds: String(link.waitSeconds > 0 ? link.waitSeconds : 5),
    isActive: link.isActive,
    expires: !!link.expiresAt,
    expiresAt: link.expiresAt ?? '',
    limited: !!link.maxClicks,
    maxClicks: link.maxClicks ? String(link.maxClicks) : '',
    adEnabled: link.adEnabled
  }
}

export const ALIAS_PATTERN = /^[A-Za-z0-9_-]{3,32}$/

function positiveInt(raw: string): number | null {
  const n = Number(String(raw).trim())
  return Number.isInteger(n) && n >= 1 ? n : null
}

/** Boş nesne dönerse taslak geçerlidir. `original` düzenlemede mevcut linktir. */
export function validateDraft(d: LinkDraft, original?: LinkItem | null): LinkDraftErrors {
  const errors: LinkDraftErrors = {}
  if (!d.target.trim()) errors.target = 'Hedef adresi yaz.'

  const alias = d.alias.trim()
  if (original && !alias) errors.alias = 'Kısa ad boş olamaz.'
  else if (alias && !ALIAS_PATTERN.test(alias)) errors.alias = '3–32 karakter: harf, rakam, tire veya alt çizgi.'

  if (d.access === 'password' && !d.password && !original?.locked) errors.password = 'Linke bir parola belirle.'
  if (d.access === 'wait' && positiveInt(d.waitSeconds) === null) errors.waitSeconds = 'En az 1 saniye yaz.'
  if (d.expires && !d.expiresAt) errors.expiresAt = 'Bir tarih seç veya bu ayarı kapat.'
  if (d.limited && positiveInt(d.maxClicks) === null) errors.maxClicks = 'En az 1 tıklama yaz veya bu ayarı kapat.'
  return errors
}

export function toCreatePayload(
  d: LinkDraft,
  can: { alias: boolean; password: boolean; ads: boolean }
): CreateLinkPayload {
  const wait = d.access === 'wait'
  return {
    target: d.target,
    title: d.title.trim() || undefined,
    alias: can.alias && d.alias.trim() ? d.alias.trim() : undefined,
    password: can.password && d.access === 'password' ? d.password : undefined,
    waitEnabled: wait,
    waitSeconds: wait ? Number(d.waitSeconds) : 0,
    expiresAt: d.expires ? d.expiresAt : undefined,
    maxClicks: d.limited ? Number(d.maxClicks) : undefined,
    adEnabled: can.ads ? d.adEnabled : false
  }
}

export function toUpdatePayload(d: LinkDraft, original: LinkItem, can: { alias: boolean; ads: boolean }): UpdateLinkPayload {
  const wait = d.access === 'wait'
  let password: string | undefined
  if (d.access === 'password') password = d.password || undefined
  else if (original.locked) password = ''

  return {
    target: d.target,
    title: d.title.trim(),
    alias: can.alias || d.alias.trim() === original.alias ? d.alias.trim() : undefined,
    password,
    waitEnabled: wait,
    waitSeconds: wait ? Number(d.waitSeconds) : 0,
    isActive: d.isActive,
    expiresAt: d.expires ? d.expiresAt : '',
    maxClicks: d.limited ? Number(d.maxClicks) : 0,
    // Paket reklama izin vermiyorsa yalnızca kapatma gönderilir; açık kalan reklam tekrar gönderilip reddedilmez.
    adEnabled: can.ads ? d.adEnabled : original.adEnabled && !d.adEnabled ? false : undefined
  }
}

function accessLabel(access: LinkAccess, seconds: string | number): string {
  if (access === 'password') return 'Parolalı'
  if (access === 'wait') return `${seconds} saniye bekletmeli`
  return 'Herkese açık'
}

/** Kaydetmeden önce gösterilecek fark listesi ve dikkat edilmesi gereken sonuçlar. */
export function draftChanges(d: LinkDraft, o: LinkItem): { changes: ConfirmChange[]; effects: string[]; critical: boolean } {
  const before = draftFromLink(o)
  const changes: ConfirmChange[] = []
  const effects: string[] = []
  let critical = false

  if (normalizeTargetUrl(d.target) !== o.target) {
    changes.push({ label: 'Hedef adres', from: o.target, to: normalizeTargetUrl(d.target) })
    effects.push('Linki daha önce paylaştığın herkes artık yeni hedefe gider.')
    critical = true
  }
  if (d.alias.trim() !== o.alias) {
    changes.push({ label: 'Kısa ad', from: o.alias, to: d.alias.trim() })
    effects.push(`Eski /${o.alias} adresi çalışmaz olur. Paylaştığın yerleri güncellemen gerekir.`)
    critical = true
  }
  if (d.title.trim() !== (o.title ?? '')) {
    changes.push({ label: 'Başlık', from: o.title || 'Yok', to: d.title.trim() || 'Yok' })
  }

  const beforeAccess = accessLabel(before.access, o.waitSeconds)
  const afterAccess = accessLabel(d.access, d.waitSeconds)
  if (beforeAccess !== afterAccess) {
    changes.push({ label: 'Erişim', from: beforeAccess, to: afterAccess })
    if (before.access === 'password') effects.push('Parola kaldırılır. Adresi bilen herkes hedefe ulaşır.')
    critical = true
  } else if (d.access === 'password' && d.password) {
    changes.push({ label: 'Parola', from: 'Eski parola', to: 'Yeni parola' })
    effects.push('Eski parolayı bilenler artık giremez.')
    critical = true
  }

  if (d.isActive !== o.isActive) {
    changes.push({ label: 'Yayın', from: o.isActive ? 'Yayında' : 'Kapalı', to: d.isActive ? 'Yayında' : 'Kapalı' })
    if (!d.isActive) effects.push('Ziyaretçiler hedef yerine "link devre dışı" sayfasını görür.')
    critical = true
  }

  const beforeExp = before.expires ? before.expiresAt : ''
  const afterExp = d.expires ? d.expiresAt : ''
  if (beforeExp !== afterExp) {
    changes.push({
      label: 'Son kullanma',
      from: beforeExp ? formatDateTr(beforeExp) : 'Süresiz',
      to: afterExp ? formatDateTr(afterExp) : 'Süresiz'
    })
    critical = true
  }

  const beforeMax = before.limited ? before.maxClicks : ''
  const afterMax = d.limited ? d.maxClicks : ''
  if (beforeMax !== afterMax) {
    changes.push({ label: 'Tıklama sınırı', from: beforeMax || 'Sınırsız', to: afterMax || 'Sınırsız' })
    if (afterMax && Number(afterMax) <= o.clicks) {
      effects.push(`Link şimdiden ${formatNumber(o.clicks)} kez tıklandı. Bu sınırla hemen kapanır.`)
    }
    critical = true
  }

  if (d.adEnabled !== o.adEnabled) {
    changes.push({ label: 'Reklam', from: o.adEnabled ? 'Açık' : 'Kapalı', to: d.adEnabled ? 'Açık' : 'Kapalı' })
  }

  return { changes, effects, critical }
}

/** Ziyaretçinin yaşayacağını tek cümleyle anlatır. */
export function visitorStory(d: Pick<LinkDraft, 'access' | 'waitSeconds'>): string {
  if (d.access === 'password') return 'Tıklayan kişi önce parola ekranını görür. Doğru parolayla hedefe geçer.'
  if (d.access === 'wait') return `Tıklayan kişi ${Number(d.waitSeconds) || 0} saniyelik geri sayımı bekler, sonra hedefe geçer.`
  return 'Tıklayan kişi beklemeden hedefe gider.'
}
