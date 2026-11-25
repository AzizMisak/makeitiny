export function formatNumber(n?: number | null): string {
  return (n ?? 0).toLocaleString('tr-TR')
}

export function formatDateTr(raw?: string): string {
  if (!raw) return '—'
  const dt = new Date(/^\d{4}-\d{2}-\d{2}$/.test(raw) ? `${raw}T12:00:00` : raw)
  if (Number.isNaN(dt.getTime())) return raw
  return dt.toLocaleDateString('tr-TR', { day: 'numeric', month: 'short', year: 'numeric' })
}

export function formatDateTimeTr(raw?: string): string {
  if (!raw) return '—'
  const dt = new Date(raw.includes('T') ? raw : raw.replace(' ', 'T'))
  if (Number.isNaN(dt.getTime())) return raw
  return dt.toLocaleString('tr-TR', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

export function formatPriceMonthly(price: number): string {
  if (!price || price <= 0) return 'Ücretsiz'
  return `${price.toLocaleString('tr-TR')} TL / ay`
}

export function todayIso(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export function initialsOf(name?: string, email?: string): string {
  const source = (name || email || '?').trim()
  const parts = source.split(/[\s._-]+/).filter(Boolean)
  const letters = parts.length > 1 ? `${parts[0]![0]}${parts[1]![0]}` : source.slice(0, 2)
  return letters.toLocaleUpperCase('tr-TR')
}

/** Adresi okunabilir parçalara ayırır (moderasyonda alan adını öne çıkarmak için). */
export function splitUrl(raw: string): { protocol: string; host: string; rest: string } {
  try {
    const u = new URL(raw)
    const base = `${u.protocol}//`
    return { protocol: base, host: u.host, rest: raw.slice(base.length + u.host.length) }
  } catch {
    return { protocol: '', host: raw, rest: '' }
  }
}
