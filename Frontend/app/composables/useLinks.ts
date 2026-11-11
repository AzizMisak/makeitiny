import type { CreateLinkPayload, LinkItem, UpdateLinkPayload } from '~/types/link'

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
  const loading = useState<boolean>('links_loading', () => false)
  const initialized = useState<boolean>('links_initialized', () => false)

  const totalClicks = computed(() => links.value.reduce((sum, l) => sum + l.clicks, 0))

  async function fetchLinks() {
    loading.value = true
    try {
      links.value = await useApi().get<LinkItem[]>('/links')
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

  async function addLink(input: CreateLinkPayload) {
    const payload: CreateLinkPayload = {
      ...input,
      target: normalizeTargetUrl(input.target)
    }
    const created = await useApi().post<LinkItem, CreateLinkPayload>('/links', payload)
    links.value = [created, ...links.value]
    return created
  }

  async function updateLink(id: string, input: UpdateLinkPayload) {
    const payload: UpdateLinkPayload = {
      ...input,
      target: input.target !== undefined ? normalizeTargetUrl(input.target) : undefined
    }
    const updated = await useApi().patch<LinkItem, UpdateLinkPayload>(`/links/${id}`, payload)
    links.value = links.value.map((l) => (l.id === id ? updated : l))
    return updated
  }

  async function removeLink(id: string) {
    await useApi().del(`/links/${id}`)
    links.value = links.value.filter((l) => l.id !== id)
  }

  function resetLinks() {
    links.value = []
    initialized.value = false
    loading.value = false
  }

  return {
    links,
    loading,
    initialized,
    totalClicks,
    fetchLinks,
    getLink,
    addLink,
    updateLink,
    removeLink,
    resetLinks
  }
}
