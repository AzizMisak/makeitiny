import type { ConfirmOptions } from '~/composables/useConfirm'
import type { User } from '~/types/auth'

/** Rol değişikliği onayı; kullanıcı ve rol sayfalarında aynı metin. */
export function roleChangePrompt(user: User, next: 'admin' | 'user', action: () => Promise<unknown>): ConfirmOptions {
  const promote = next === 'admin'
  return {
    title: promote ? 'Yönetici yapılsın mı?' : 'Yönetici yetkisi kaldırılsın mı?',
    subject: `${user.name || 'İsimsiz'} (${user.email})`,
    message: promote ? 'Yönetim paneline girebilir.' : 'Yönetim paneline giremez, standart kullanıcı olur.',
    effects: promote
      ? [
          'Herkesin linkini ve ziyaret kayıtlarını görür.',
          'Linkleri engelleyebilir ve silebilir.',
          'Kullanıcıları askıya alabilir, paketlerini değiştirebilir.',
          'Kendi paketi geçersiz olur; sınırı kalmaz, bütün özellikler açılır.'
        ]
      : [`Kayıtlı paketi (${user.plan?.name ?? 'Ücretsiz Paket'}) yeniden geçerli olur ve sınırları uygulanır.`, 'Kendi linkleri aynen kalır.'],
    changes: [{ label: 'Rol', from: promote ? 'Kullanıcı' : 'Yönetici', to: promote ? 'Yönetici' : 'Kullanıcı' }],
    confirmLabel: promote ? 'Yönetici yap' : 'Yetkiyi kaldır',
    tone: 'danger',
    action
  }
}
