import type { AdminLinkStatusFilter } from '~/composables/useAdmin'

export type ExportFormat = 'csv' | 'json'

/** Linkleri ve ziyaret kayıtlarını oturumla indirip tarayıcıya kaydettirir. */
export function useExport() {
  const toast = useToast()
  const busy = ref('')

  function save(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    a.remove()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  }

  async function download(
    key: string,
    path: string,
    format: ExportFormat,
    filename: string,
    extraQuery?: Record<string, string | number | boolean | undefined>
  ) {
    if (busy.value) return
    busy.value = key
    try {
      const blob = await useApi().blob(path, { format, ...extraQuery })
      save(blob, `${filename}.${format}`)
      toast.success(`${format.toUpperCase()} dosyası indirildi.`)
    } catch (e: any) {
      toast.error(e?.message ?? 'Dosya hazırlanamadı. Tekrar dene.')
    } finally {
      busy.value = ''
    }
  }

  const stamp = () => todayIso()

  return {
    busy,
    exportLinks: (format: ExportFormat) => download(`links-${format}`, '/links/export', format, `makeitiny-linkler-${stamp()}`),
    exportVisits: (id: string, alias: string, format: ExportFormat) =>
      download(`visits-${format}`, `/links/logs/${id}/export`, format, `makeitiny-${alias}-ziyaretler-${stamp()}`),
    exportAdminLinks: (format: ExportFormat, search = '', status: AdminLinkStatusFilter = 'all') =>
      download(`admin-links-${format}`, '/admin/links/export', format, `makeitiny-yonetim-linkler-${stamp()}`, {
        status,
        ...(search.trim() ? { search: search.trim() } : {})
      }),
    exportAdminUsers: (format: ExportFormat, search = '') =>
      download(
        `admin-users-${format}`,
        '/admin/users/export',
        format,
        `makeitiny-kullanicilar-${stamp()}`,
        search.trim() ? { search: search.trim() } : undefined
      ),
    exportAdminUserLinks: (userId: string, format: ExportFormat) =>
      download(`admin-user-links-${format}`, `/admin/users/${userId}/links/export`, format, `makeitiny-${userId}-linkler-${stamp()}`)
  }
}
