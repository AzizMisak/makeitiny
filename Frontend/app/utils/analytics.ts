import type { LinkLogItem } from '~/types/link'

export interface ShareRow {
  label: string
  code?: string
  count: number
  percent: number
}

export interface CountryRow extends ShareRow {
  region: string
  langs: string[]
  lastSeen: string
}

export interface TimeBucket {
  key: string
  label: string
  full: string
  count: number
}

export type TimeRange = '7d' | '14d' | '30d' | 'hourly'

function percentOf(count: number, total: number): number {
  return total === 0 ? 0 : Math.round((count / total) * 1000) / 10
}

/** Değerleri sayıp çoktan aza sıralar. */
export function shareRows(values: string[]): ShareRow[] {
  const counts = new Map<string, number>()
  for (const raw of values) {
    const key = raw || 'Bilinmiyor'
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }
  return [...counts.entries()]
    .sort((a, b) => b[1] - a[1])
    .map(([label, count]) => ({ label, count, percent: percentOf(count, values.length) }))
}

export function countryRows(logs: LinkLogItem[]): CountryRow[] {
  const map = new Map<string, { code: string; region: string; count: number; langs: Set<string>; lastSeen: string }>()
  for (const log of logs) {
    const c = resolveCountry(log)
    const entry = map.get(c.name) ?? {
      code: c.code,
      region: c.region,
      count: 0,
      langs: new Set<string>(),
      lastSeen: log.clickedAt
    }
    entry.count += 1
    if (c.lang) entry.langs.add(c.lang)
    if (new Date(log.clickedAt) > new Date(entry.lastSeen)) entry.lastSeen = log.clickedAt
    map.set(c.name, entry)
  }
  return [...map.entries()]
    .sort((a, b) => b[1].count - a[1].count)
    .map(([label, e]) => ({
      label,
      code: e.code,
      region: e.region,
      count: e.count,
      percent: percentOf(e.count, logs.length),
      langs: [...e.langs],
      lastSeen: e.lastSeen
    }))
}

function dayKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

/** Tıklama kayıtlarını bugünden geriye günlere ya da günün saatlerine böler. */
export function timeBuckets(logs: LinkLogItem[], range: TimeRange): TimeBucket[] {
  if (range === 'hourly') {
    const counts = Array.from({ length: 24 }, () => 0)
    for (const log of logs) {
      const at = new Date(log.clickedAt)
      if (!Number.isNaN(at.getTime())) counts[at.getHours()]! += 1
    }
    return counts.map((count, h) => {
      const hh = String(h).padStart(2, '0')
      return { key: hh, label: hh, full: `${hh}:00 – ${hh}:59`, count }
    })
  }

  const days = range === '7d' ? 7 : range === '14d' ? 14 : 30
  const perDay = new Map<string, number>()
  for (const log of logs) {
    const at = new Date(log.clickedAt)
    if (Number.isNaN(at.getTime())) continue
    const day = dayKey(at)
    perDay.set(day, (perDay.get(day) ?? 0) + 1)
  }

  const today = new Date()
  return Array.from({ length: days }, (_, i) => {
    const d = new Date(today)
    d.setDate(today.getDate() - (days - 1 - i))
    const key = dayKey(d)
    return {
      key,
      label: d.toLocaleDateString('tr-TR', { day: 'numeric', month: 'short' }),
      full: d.toLocaleDateString('tr-TR', { weekday: 'long', day: 'numeric', month: 'long' }),
      count: perDay.get(key) ?? 0
    }
  })
}

/** Grafik ekseni için yuvarlak üst sınır. */
export function niceMax(max: number): number {
  if (max <= 4) return 4
  const pow = 10 ** Math.floor(Math.log10(max))
  for (const step of [1, 2, 4, 5, 10]) {
    if (step * pow >= max) return step * pow
  }
  return 10 * pow
}
