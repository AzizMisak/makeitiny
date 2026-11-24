import type {
  AliasCheckResult,
  CreateLinkPayload,
  LinkItem,
  PaginatedLinksResponse,
  PaginatedLogsResponse,
  UpdateLinkPayload
} from '~/types/link'

export function normalizeTargetUrl(raw: string): string {
  const trimmed = raw.trim()
  if (!trimmed) return ''
  if (!/^https?:\/\//i.test(trimmed)) {
    return `https://${trimmed}`
  }
  return trimmed
}

export function useLinks() {
  const links = useState<LinkItem[]>('links', () => [])
  const topLinks = useState<LinkItem[]>('links_top', () => [])
  const currentPage = useState<number>('links_page', () => 1)
  const pageSize = useState<number>('links_limit', () => 10)
  const totalLinks = useState<number>('links_total', () => 0)
  const totalPages = useState<number>('links_total_pages', () => 1)
  const totalClicks = useState<number>('links_total_clicks', () => 0)
  const lockedCount = useState<number>('links_locked_count', () => 0)
  const loading = useState<boolean>('links_loading', () => false)
  const initialized = useState<boolean>('links_initialized', () => false)

  async function fetchLinks(page = currentPage.value, limit = pageSize.value) {
    loading.value = true
    try {
      const res = await useApi().get<PaginatedLinksResponse>(`/links?page=${page}&limit=${limit}`)
      links.value = res.items ?? []
      topLinks.value = res.topLinks ?? []
      currentPage.value = res.page ?? 1
      pageSize.value = res.limit ?? limit
      totalLinks.value = res.total ?? 0
      totalPages.value = res.totalPages ?? 1
      totalClicks.value = res.totalClicks ?? 0
      lockedCount.value = res.lockedCount ?? 0
    } finally {
      loading.value = false
      initialized.value = true
    }
  }

  async function getLink(id: string): Promise<LinkItem> {
    const item = await useApi().get<LinkItem>(`/links/${id}`)
    const exists = links.value.some((l) => l.id === id)
    if (exists) {
      links.value = links.value.map((l) => (l.id === id ? item : l))
    }
    return item
  }

  async function getLinkLogs(id: string, page = 1, limit = 20): Promise<PaginatedLogsResponse> {
    return await useApi().get<PaginatedLogsResponse>(`/links/logs/${id}?page=${page}&limit=${limit}`)
  }

  async function checkAlias(alias: string): Promise<AliasCheckResult> {
    return await useApi().get<AliasCheckResult>(`/links/check-alias/${encodeURIComponent(alias.trim())}`)
  }

  async function addLink(input: CreateLinkPayload) {
    const payload: CreateLinkPayload = {
      ...input,
      target: normalizeTargetUrl(input.target)
    }
    const created = await useApi().post<LinkItem, CreateLinkPayload>('/links', payload)
    await Promise.all([
      fetchLinks(1, pageSize.value),
      useAuth().fetchCurrentUser().catch(() => {})
    ])
    return created
  }

  async function updateLink(id: string, input: UpdateLinkPayload) {
    const payload: UpdateLinkPayload = {
      ...input,
      target: input.target !== undefined ? normalizeTargetUrl(input.target) : undefined
    }
    const updated = await useApi().patch<LinkItem, UpdateLinkPayload>(`/links/${id}`, payload)
    links.value = links.value.map((l) => (l.id === id ? updated : l))
    topLinks.value = topLinks.value.map((l) => (l.id === id ? updated : l))
    return updated
  }

  async function removeLink(id: string) {
    await useApi().del(`/links/${id}`)
    await Promise.all([
      fetchLinks(currentPage.value, pageSize.value),
      useAuth().fetchCurrentUser().catch(() => {})
    ])
  }

  function resetLinks() {
    links.value = []
    topLinks.value = []
    currentPage.value = 1
    totalLinks.value = 0
    totalPages.value = 1
    totalClicks.value = 0
    lockedCount.value = 0
    initialized.value = false
    loading.value = false
  }

  return {
    links,
    topLinks,
    currentPage,
    pageSize,
    totalLinks,
    totalPages,
    totalClicks,
    lockedCount,
    loading,
    initialized,
    fetchLinks,
    getLink,
    getLinkLogs,
    checkAlias,
    addLink,
    updateLink,
    removeLink,
    resetLinks
  }
}
