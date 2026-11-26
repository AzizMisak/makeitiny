import type { AdminLinkItem } from '~/types/link'

/** Yönetici işlemlerinin ortak onay akışı. Her işlem onaylatılır, sonucu bildirilir. */
export function useLinkModeration() {
  const { adminUpdateLink, adminDeleteLink } = useAdmin()
  const { ask } = useConfirm()
  const { display } = useShortUrl()
  const toast = useToast()

  async function setActive(link: AdminLinkItem, active: boolean): Promise<boolean> {
    const ok = await ask({
      title: active ? 'Yönlendirme açılsın mı?' : 'Yönlendirme kapatılsın mı?',
      subject: display(link.alias),
      message: active
        ? 'Kısa adres yeniden hedefe yönlendirir.'
        : 'Ziyaretçiler hedef yerine "link devre dışı" sayfasını görür.',
      confirmLabel: active ? 'Yönlendirmeyi aç' : 'Yönlendirmeyi kapat',
      tone: active ? 'normal' : 'danger',
      action: () => adminUpdateLink(link.id, { isActive: active })
    })
    if (ok) toast.success(active ? `${link.alias} açıldı.` : `${link.alias} kapatıldı.`)
    return ok
  }

  async function remove(link: AdminLinkItem): Promise<boolean> {
    const ok = await ask({
      title: 'Link silinsin mi?',
      subject: `${display(link.alias)}${link.userEmail ? ` (${link.userEmail})` : ''}`,
      message: 'Bu işlem geri alınamaz. Linki yalnızca durdurmak istiyorsan yönlendirmeyi kapat.',
      effects: ['Kısa adres hemen çalışmaz olur.', 'Sahibi linki ve analizini göremez.'],
      confirmLabel: 'Linki sil',
      tone: 'danger',
      typeToConfirm: link.alias,
      action: () => adminDeleteLink(link.id)
    })
    if (ok) toast.success(`${link.alias} silindi.`)
    return ok
  }

  return { setActive, remove }
}
