export type VisitorReason =
  | 'not_found'
  | 'unreachable'
  | 'disabled'
  | 'blocked'
  | 'expired'
  | 'limit_reached'
  | 'internal'

export interface VisitorError {
  title: string
  text: string
  canRetry: boolean
}

export const VISITOR_ERRORS: Record<VisitorReason, VisitorError> = {
  not_found: {
    title: 'Böyle bir link yok',
    text: 'Adres yanlış yazılmış ya da sahibi linki silmiş olabilir.',
    canRetry: false
  },
  unreachable: {
    title: 'Hedef site yanıt vermiyor',
    text: 'Kısa link doğru, ama gittiği siteye şu an ulaşılamıyor. Biraz sonra tekrar dene.',
    canRetry: true
  },
  disabled: {
    title: 'Bu link kapatılmış',
    text: 'Sahibi ya da yönetim linki durdurmuş. Yeniden açılırsa aynı adres çalışır.',
    canRetry: false
  },
  blocked: {
    title: 'Bu link engellendi',
    text: 'Hedef adres güvenlik ya da kullanım kuralları nedeniyle engellendi. Seni oraya göndermiyoruz.',
    canRetry: false
  },
  expired: {
    title: 'Bu linkin süresi dolmuş',
    text: 'Sahibinin belirlediği son kullanma tarihi geçti.',
    canRetry: false
  },
  limit_reached: {
    title: 'Bu linkin hakkı bitmiş',
    text: 'Sahibinin belirlediği tıklama sayısına ulaşıldı.',
    canRetry: false
  },
  internal: {
    title: 'Link açılamadı',
    text: 'Beklenmeyen bir sorun oldu. Tekrar dene.',
    canRetry: true
  }
}

/** Router hata kodunu hata sayfası nedenine çevirir. */
export function reasonFromCode(code?: string): VisitorReason | null {
  switch (code) {
    case 'link_not_found':
      return 'not_found'
    case 'target_unreachable':
      return 'unreachable'
    case 'link_disabled':
      return 'disabled'
    case 'link_blocked':
      return 'blocked'
    case 'link_expired':
      return 'expired'
    case 'link_limit_reached':
      return 'limit_reached'
    default:
      return null
  }
}
