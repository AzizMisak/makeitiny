<script setup lang="ts">
const route = useRoute()
const { isAdmin, isOwner } = useAuth()
const { isDark, toggleTheme, syncFromDom } = useTheme()
const menuOpen = ref(false)

onMounted(syncFromDom)

watch(() => route.path, () => (menuOpen.value = false))

type WorkspaceMode = 'personal' | 'admin' | 'owner'

const workspace = computed<WorkspaceMode>(() => {
  if (route.path.startsWith('/dashboard/owner')) return 'owner'
  if (route.path.startsWith('/dashboard/admin')) return 'admin'
  return 'personal'
})

const workspaceMeta = computed(() => {
  switch (workspace.value) {
    case 'owner':
      return {
        label: 'SİSTEM SAHİBİ (OWNER)',
        stripClass: 'border-t-4 border-t-violet-500',
        pillClass: 'border-violet-500/60 bg-violet-500/15 text-violet-700 dark:text-violet-300',
        barBgClass: 'bg-violet-500/5'
      }
    case 'admin':
      return {
        label: 'YÖNETİM PANELİ (ADMİN)',
        stripClass: 'border-t-4 border-t-amber-500',
        pillClass: 'border-amber-500/60 bg-amber-500/15 text-amber-800 dark:text-amber-300',
        barBgClass: 'bg-amber-500/5'
      }
    default:
      return {
        label: 'KİŞİSEL ALAN',
        stripClass: 'border-t-4 border-t-accent',
        pillClass: 'border-accent/50 bg-accent-soft text-accent-dark',
        barBgClass: 'bg-card/70'
      }
  }
})

interface BreadcrumbItem {
  label: string
  to?: string
}

const breadcrumbs = computed<BreadcrumbItem[]>(() => {
  const path = route.path
  if (path === '/dashboard') return [{ label: 'Kişisel Alan', to: '/dashboard' }, { label: 'Genel Bakış' }]
  if (path === '/dashboard/links') return [{ label: 'Kişisel Alan', to: '/dashboard' }, { label: 'Linklerim' }]
  if (path.startsWith('/dashboard/links/')) {
    return [
      { label: 'Kişisel Alan', to: '/dashboard' },
      { label: 'Linklerim', to: '/dashboard/links' },
      { label: 'Link Detayı' }
    ]
  }
  if (path === '/dashboard/new') return [{ label: 'Kişisel Alan', to: '/dashboard' }, { label: 'Yeni Link' }]
  if (path === '/dashboard/settings') return [{ label: 'Kişisel Alan', to: '/dashboard' }, { label: 'Hesap Ayarları' }]
  if (path === '/dashboard/subscription') return [{ label: 'Kişisel Alan', to: '/dashboard' }, { label: 'Abonelik & Paket' }]

  if (path === '/dashboard/admin') return [{ label: 'Yönetim Paneli', to: '/dashboard/admin' }, { label: 'Yönetim Özeti' }]
  if (path === '/dashboard/admin/users') return [{ label: 'Yönetim Paneli', to: '/dashboard/admin' }, { label: 'Kullanıcılar' }]
  if (path.startsWith('/dashboard/admin/users/')) {
    return [
      { label: 'Yönetim Paneli', to: '/dashboard/admin' },
      { label: 'Kullanıcılar', to: '/dashboard/admin/users' },
      { label: 'Kullanıcı İnceleme' }
    ]
  }
  if (path === '/dashboard/admin/links') return [{ label: 'Yönetim Paneli', to: '/dashboard/admin' }, { label: 'Tüm Linkler' }]
  if (path.startsWith('/dashboard/admin/links/')) {
    return [
      { label: 'Yönetim Paneli', to: '/dashboard/admin' },
      { label: 'Tüm Linkler', to: '/dashboard/admin/links' },
      { label: 'Moderasyon Detayı' }
    ]
  }

  if (path === '/dashboard/owner/roles') return [{ label: 'Owner', to: '/dashboard/owner/roles' }, { label: 'Roller & Yetkiler' }]
  if (path === '/dashboard/owner/plans') return [{ label: 'Owner', to: '/dashboard/owner/plans' }, { label: 'Paket Yönetimi' }]
  if (path === '/dashboard/forbidden') return [{ label: 'Panel', to: '/dashboard' }, { label: 'Yetki Sınırı' }]

  return [{ label: 'Panel', to: '/dashboard' }]
})
</script>

<template>
  <div class="min-h-screen lg:grid lg:grid-cols-[16.5rem_1fr]" :class="workspaceMeta.stripClass">
    <!-- Masaüstü menü -->
    <div class="sticky top-0 hidden h-screen lg:block">
      <DashboardSidebar />
    </div>

    <!-- Mobil menü -->
    <div v-if="menuOpen" class="fixed inset-0 z-30 flex lg:hidden">
      <DashboardSidebar class="shadow-hard-lg" />
      <button type="button" class="flex-1 bg-ink/40" aria-label="Menüyü kapat" @click="menuOpen = false" />
    </div>

    <div class="min-w-0">
      <DashboardTopbar :workspace="workspace" :workspace-label="workspaceMeta.label" @toggle-menu="menuOpen = !menuOpen" />

      <!-- Kalıcı Çalışma Alanı (Workspace) & Konum Çubuğu -->
      <div class="border-b border-line px-5 py-2.5 md:px-10" :class="workspaceMeta.barBgClass">
        <div class="mx-auto flex w-full max-w-5xl flex-wrap items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2.5 text-xs">
            <span
              class="inline-flex items-center gap-1.5 rounded-sm border px-2 py-0.5 font-mono text-[11px] font-bold tracking-wide"
              :class="workspaceMeta.pillClass"
            >
              <span class="h-1.5 w-1.5 rounded-full bg-current" />
              {{ workspaceMeta.label }}
            </span>

            <nav class="flex flex-wrap items-center gap-1 text-ink-3" aria-label="Konum yolu">
              <template v-for="(crumb, idx) in breadcrumbs" :key="`${crumb.label}-${idx}`">
                <span v-if="idx > 0" class="px-1 select-none">›</span>
                <NuxtLink
                  v-if="crumb.to"
                  :to="crumb.to"
                  class="font-medium text-ink-2 transition-colors hover:text-ink hover:underline"
                >
                  {{ crumb.label }}
                </NuxtLink>
                <span v-else class="font-semibold text-ink">{{ crumb.label }}</span>
              </template>
            </nav>
          </div>

          <div class="flex items-center gap-2">
            <!-- Hızlı Mod Değiştirici (Rolüne Göre) -->
            <div v-if="isAdmin" class="hidden items-center gap-1 rounded-md border border-line bg-paper p-0.5 sm:inline-flex">
              <NuxtLink
                to="/dashboard"
                class="rounded-xs px-2.5 py-1 font-mono text-[11px] font-bold transition-colors"
                :class="workspace === 'personal' ? 'bg-accent text-white' : 'text-ink-2 hover:text-ink'"
              >
                Kişisel
              </NuxtLink>
              <NuxtLink
                to="/dashboard/admin"
                class="rounded-xs px-2.5 py-1 font-mono text-[11px] font-bold transition-colors"
                :class="workspace === 'admin' ? 'bg-amber-500 text-white' : 'text-ink-2 hover:text-ink'"
              >
                Yönetim
              </NuxtLink>
              <NuxtLink
                v-if="isOwner"
                to="/dashboard/owner/roles"
                class="rounded-xs px-2.5 py-1 font-mono text-[11px] font-bold transition-colors"
                :class="workspace === 'owner' ? 'bg-violet-600 text-white' : 'text-ink-2 hover:text-ink'"
              >
                Owner
              </NuxtLink>
            </div>

            <button
              type="button"
              class="hidden h-8 w-8 cursor-pointer items-center justify-center rounded-sm border border-line bg-card text-ink-2 transition-colors hover:border-ink hover:text-ink lg:inline-flex"
              :title="isDark ? 'Gündüz moduna geç' : 'Gece moduna geç'"
              :aria-label="isDark ? 'Gündüz moduna geç' : 'Gece moduna geç'"
              @click="toggleTheme"
            >
              <UiIcon :name="isDark ? 'sun' : 'moon'" :size="15" />
            </button>
          </div>
        </div>
      </div>

      <main class="mx-auto w-full max-w-5xl px-5 py-8 md:px-10 md:py-12">
        <slot />
      </main>
    </div>

    <UiToastContainer />
  </div>
</template>
