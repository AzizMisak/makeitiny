import type { LinkItem } from '~/types/link'

/** Link silmeyi onaylatır. Silindiyse true döner. */
export function useLinkRemoval() {
  const { ask } = useConfirm()
  const { removeLink } = useLinks()
  const { display } = useShortUrl()
  const toast = useToast()

  return async function remove(link: LinkItem): Promise<boolean> {
    const ok = await ask({
      title: 'Link silinsin mi?',
      subject: display(link.alias),
      message: 'Bu işlem geri alınamaz.',
      effects: [
        'Kısa adres hemen çalışmaz olur.',
        `${formatNumber(link.clicks)} tıklamanın analizine bir daha ulaşamazsın.`,
        ...(link.hasQr ? ['Basılı ya da paylaşılmış QR kodu da boşa düşer.'] : [])
      ],
      confirmLabel: 'Linki sil',
      cancelLabel: 'Vazgeç',
      tone: 'danger',
      action: () => removeLink(link.id)
    })
    if (ok) toast.success(`${link.alias} silindi.`)
    return ok
  }
}
