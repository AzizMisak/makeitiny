import type { PlanItem } from '~/types/auth'

export interface PlanRow {
  key: keyof PlanItem
  label: string
  kind: 'text' | 'flag'
  value: (p: PlanItem) => string | boolean
}

/** Paket karşılaştırma tablolarının ortak satırları. */
export const PLAN_ROWS: PlanRow[] = [
  { key: 'maxLinks', label: 'Link hakkı', kind: 'text', value: (p) => (p.maxLinks < 0 ? 'Sınırsız' : formatNumber(p.maxLinks)) },
  { key: 'analyticsDays', label: 'Analiz geçmişi', kind: 'text', value: (p) => `${formatNumber(p.analyticsDays)} gün` },
  { key: 'allowPassword', label: 'Parolalı link', kind: 'flag', value: (p) => p.allowPassword },
  { key: 'allowCustomAlias', label: 'Kendi kısa adını seçme', kind: 'flag', value: (p) => p.allowCustomAlias },
  { key: 'allowQr', label: 'QR kodu', kind: 'flag', value: (p) => p.allowQr },
  { key: 'allowAds', label: 'Reklam gösterimi', kind: 'flag', value: (p) => p.allowAds }
]
