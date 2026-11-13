export interface NavItem {
  label: string
  to: string
  icon: string
  exact?: boolean
}

export interface NavGroup {
  title: string
  items: NavItem[]
}

export const dashboardNav: NavGroup[] = [
  {
    title: 'Çalışma alanı',
    items: [
      { label: 'Genel bakış', to: '/dashboard', icon: 'grid', exact: true },
      { label: 'Linklerim', to: '/dashboard/links', icon: 'link' },
      { label: 'Yeni link', to: '/dashboard/new', icon: 'plus' }
    ]
  },
  {
    title: 'Hesap',
    items: [{ label: 'Ayarlar', to: '/dashboard/settings', icon: 'sliders' }]
  }
]

export const landingNav: { label: string; href: string; sectionId: string }[] = [
  { label: 'Özellikler', href: '/#ozellikler', sectionId: 'ozellikler' },
  { label: 'Analitik & Loglar', href: '/#analitik', sectionId: 'analitik' },
  { label: 'Nasıl çalışır', href: '/#nasil', sectionId: 'nasil' }
]
