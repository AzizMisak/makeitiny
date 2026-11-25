import type { RoleName } from '~/types/auth'

export interface NavItem {
  label: string
  to: string
  icon: string
  exact?: boolean
  roles?: RoleName[]
}

export const userNav: NavItem[] = [
  { label: 'Genel bakış', to: '/dashboard', icon: 'home', exact: true },
  { label: 'Linklerim', to: '/dashboard/links', icon: 'link' },
  { label: 'Paketim', to: '/dashboard/subscription', icon: 'gauge' },
  { label: 'Hesap ayarları', to: '/dashboard/settings', icon: 'settings' }
]

export const adminNav: NavItem[] = [
  { label: 'Özet', to: '/dashboard/admin', icon: 'activity', exact: true, roles: ['admin', 'owner'] },
  { label: 'Linkler', to: '/dashboard/admin/links', icon: 'globe', roles: ['admin', 'owner'] },
  { label: 'Kullanıcılar', to: '/dashboard/admin/users', icon: 'users', roles: ['admin', 'owner'] }
]

/** Yalnızca sistem sahibinin gördüğü yönetim sayfaları. */
export const ownerNav: NavItem[] = [
  { label: 'Roller', to: '/dashboard/owner/roles', icon: 'key', roles: ['owner'] },
  { label: 'Paketler', to: '/dashboard/owner/plans', icon: 'layers', roles: ['owner'] }
]

export const landingSections: { id: string; label: string }[] = [
  { id: 'kisalt', label: 'Kısalt' },
  { id: 'ziyaretci', label: 'Ziyaretçi ne görür' },
  { id: 'kayit', label: 'Tıklama kaydı' },
  { id: 'fark', label: 'Üyelik farkı' }
]

export function isNavActive(item: NavItem, path: string): boolean {
  return item.exact ? path === item.to : path === item.to || path.startsWith(`${item.to}/`)
}
