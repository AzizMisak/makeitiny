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
    return await useApi().get<AdminPaginatedLinksResponse>(`/admin/links?${params.toString()}`)
  }

  /** Kayıt yoksa null döner; diğer hatalar (ağ, 500) çağırana iletilir. */
  async function fetchAdminLinkById(id: string): Promise<AdminLinkItem | null> {
    try {
      return await useApi().get<AdminLinkItem>(`/admin/links/${id}`)
    } catch (e: any) {
      if (e?.code === 'link_not_found') return null
      throw e
    }
  }

  async function adminUpdateLink(id: string, payload: AdminUpdateLinkPayload): Promise<LinkItem> {
    return await useApi().patch<LinkItem, AdminUpdateLinkPayload>(`/admin/links/${id}`, payload)
  }

  async function adminDeleteLink(id: string): Promise<void> {
    await useApi().del(`/admin/links/${id}`)
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
    return await useApi().get<AdminPaginatedUsersResponse>(`/admin/users?${params.toString()}`)
  }

  /** Kayıt yoksa null döner; diğer hatalar (ağ, 500) çağırana iletilir. */
  async function fetchAdminUserById(id: string): Promise<User | null> {
    try {
      return await useApi().get<User>(`/admin/users/${id}`)
    } catch (e: any) {
      if (e?.code === 'user_not_found') return null
      throw e
    }
  }

  async function fetchAdminUserLinks(userId: string, userEmail: string): Promise<AdminLinkItem[]> {
    const res = await fetchAdminLinks(1, 100, userEmail, 'all')
    return res.items.filter((item) => item.userId === userId || item.userEmail === userEmail)
  }

  async function adminUpdateUser(id: string, payload: AdminUpdateUserPayload): Promise<User> {
    return await useApi().patch<User, AdminUpdateUserPayload>(`/admin/users/${id}`, payload)
  }

  async function adminUpdateUserRole(
    id: string,
    role: Extract<RoleName, 'user' | 'admin'>
  ): Promise<User> {
    return await useApi().put<User, AdminUpdateRolePayload>(`/admin/user-role/${id}`, { role })
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
