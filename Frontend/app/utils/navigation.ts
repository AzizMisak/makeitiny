import type { RoleName } from '~/types/auth'

export interface NavItem {
  label: string
  to: string
  icon: string
  exact?: boolean
  roles?: RoleName[]
}

export interface NavGroup {
  title: string
  roles?: RoleName[]
  items: NavItem[]
}

export const dashboardNav: NavGroup[] = [
  {
    title: 'Bağlantılar',
    items: [
      { label: 'Genel bakış', to: '/dashboard', icon: 'grid', exact: true },
      { label: 'Linklerim', to: '/dashboard/links', icon: 'link' },
      { label: 'Yeni link', to: '/dashboard/new', icon: 'plus' }
    ]
  },
  {
    title: 'Yönetici',
    roles: ['admin', 'owner'],
    items: [
      { label: 'Yönetim özeti', to: '/dashboard/admin', icon: 'shield', exact: true },
      { label: 'Kullanıcılar', to: '/dashboard/admin/users', icon: 'users' },
      { label: 'Tüm linkler', to: '/dashboard/admin/links', icon: 'globe' }
    ]
  },
  {
    title: 'Özel',
    roles: ['owner'],
    items: [
      { label: 'Roller & yetkiler', to: '/dashboard/owner/roles', icon: 'crown' },
      { label: 'Paket yönetimi', to: '/dashboard/owner/plans', icon: 'layers' }
    ]
  },
  {
    title: 'Hesap',
    items: [
      { label: 'Hesap ayarları', to: '/dashboard/settings', icon: 'sliders' },
      { label: 'Abonelik & paket', to: '/dashboard/subscription', icon: 'sparkle' }
    ]
  }
]

export const landingNav: { label: string; href: string; sectionId: string }[] = [
  { label: 'Özellikler', href: '/#ozellikler', sectionId: 'ozellikler' },
  { label: 'Analitik & Loglar', href: '/#analitik', sectionId: 'analitik' },
  { label: 'Nasıl çalışır', href: '/#nasil', sectionId: 'nasil' }
]
