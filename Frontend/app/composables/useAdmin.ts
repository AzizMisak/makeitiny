import type { PlanItem, RoleName, User } from '~/types/auth'
import type {
  AdminLinkItem,
  AdminPaginatedLinksResponse,
  AdminPaginatedUsersResponse,
  AdminStatsResponse,
  AdminUpdateLinkPayload,
  AdminUpdatePlanPayload,
  AdminUpdateRolePayload,
  AdminUpdateUserPayload,
  LinkItem,
  PaginatedLogsResponse
} from '~/types/link'

export type AdminLinkStatusFilter = 'all' | 'active' | 'inactive' | 'blocked' | 'anonymous'

export function useAdmin() {
  const usersCache = useState<Record<string, User>>('mit_admin_users_cache', () => ({}))
  const linksCache = useState<Record<string, AdminLinkItem>>('mit_admin_links_cache', () => ({}))

  async function fetchPlans(): Promise<PlanItem[]> {
    return await useApi().get<PlanItem[]>('/plans')
  }

  async function fetchAdminStats(): Promise<AdminStatsResponse> {
    return await useApi().get<AdminStatsResponse>('/admin/stats')
  }

  async function fetchAdminLinks(
    page = 1,
    limit = 20,
    search = '',
    status: AdminLinkStatusFilter = 'all'
  ): Promise<AdminPaginatedLinksResponse> {
    const params = new URLSearchParams({
      page: String(page),
      limit: String(limit),
      status
    })
    if (search.trim()) {
      params.set('search', search.trim())
    }
    const res = await useApi().get<AdminPaginatedLinksResponse>(`/admin/links?${params.toString()}`)
    for (const item of res.items) {
      linksCache.value[item.id] = item
    }
    return res
  }

  async function fetchAdminLinkById(id: string): Promise<AdminLinkItem | null> {
    try {
      const item = await useApi().get<AdminLinkItem>(`/admin/links/${id}`)
      linksCache.value[item.id] = item
      return item
    } catch {
      return linksCache.value[id] ?? null
    }
  }

  async function adminUpdateLink(id: string, payload: AdminUpdateLinkPayload): Promise<LinkItem> {
    const updated = await useApi().patch<LinkItem, AdminUpdateLinkPayload>(`/admin/links/${id}`, payload)
    if (linksCache.value[id]) {
      linksCache.value[id] = { ...linksCache.value[id]!, ...updated }
    }
    return updated
  }

  async function adminDeleteLink(id: string): Promise<void> {
    await useApi().del(`/admin/links/${id}`)
    delete linksCache.value[id]
  }

  async function fetchAdminLinkLogs(id: string, page = 1, limit = 20): Promise<PaginatedLogsResponse> {
    return await useApi().get<PaginatedLogsResponse>(`/admin/link-logs/${id}?page=${page}&limit=${limit}`)
  }

  async function fetchAdminUsers(
    page = 1,
    limit = 20,
    search = ''
  ): Promise<AdminPaginatedUsersResponse> {
    const params = new URLSearchParams({
      page: String(page),
      limit: String(limit)
    })
    if (search.trim()) {
      params.set('search', search.trim())
    }
    const res = await useApi().get<AdminPaginatedUsersResponse>(`/admin/users?${params.toString()}`)
    for (const u of res.items) {
      usersCache.value[u.id] = u
    }
    return res
  }

  async function fetchAdminUserById(id: string): Promise<User | null> {
    try {
      const u = await useApi().get<User>(`/admin/users/${id}`)
      usersCache.value[u.id] = u
      return u
    } catch {
      return usersCache.value[id] ?? null
    }
  }

  async function fetchAdminUserLinks(userId: string, userEmail: string): Promise<AdminLinkItem[]> {
    const res = await fetchAdminLinks(1, 100, userEmail, 'all')
    return res.items.filter((item) => item.userId === userId || item.userEmail === userEmail)
  }

  async function adminUpdateUser(id: string, payload: AdminUpdateUserPayload): Promise<User> {
    const updated = await useApi().patch<User, AdminUpdateUserPayload>(`/admin/users/${id}`, payload)
    usersCache.value[id] = updated
    return updated
  }

  async function adminUpdateUserRole(
    id: string,
    role: Extract<RoleName, 'user' | 'admin'>
  ): Promise<User> {
    const updated = await useApi().put<User, AdminUpdateRolePayload>(`/admin/user-role/${id}`, { role })
    usersCache.value[id] = updated
    return updated
  }

  async function adminUpdatePlan(id: number, payload: AdminUpdatePlanPayload): Promise<PlanItem> {
    return await useApi().patch<PlanItem, AdminUpdatePlanPayload>(`/admin/plans/${id}`, payload)
  }

  return {
    fetchPlans,
    fetchAdminStats,
    fetchAdminLinks,
    fetchAdminLinkById,
    adminUpdateLink,
    adminDeleteLink,
    fetchAdminLinkLogs,
    fetchAdminUsers,
    fetchAdminUserById,
    fetchAdminUserLinks,
    adminUpdateUser,
    adminUpdateUserRole,
    adminUpdatePlan
  }
}
