<script setup lang="ts">
definePageMeta({ layout: 'marketing' })

const route = useRoute()
const { public: config } = useRuntimeConfig()

const alias = computed(() => String(route.params.alias ?? ''))
const reason = computed(() => String(route.query.reason ?? 'not_found'))

const reasonConfig = computed(() => {
  switch (reason.value) {
    case 'unreachable':
      return {
        pageTitle: `Hedefe Ulaşılamıyor (${alias.value}) — makeiTiny`,
        cardTitle: 'Hedef adrese ulaşılamıyor.',
        cardSubtitle: `/${alias.value} kısa bağlantısı sistemimizde kayıtlı ancak yönlendirilen hedef siteye şu anda erişilemiyor.`,
        badge: '502 • HEDEF KAYNAKLI HATA',
        boxTitle: 'Hedef web sitesi yanıt vermiyor',
        boxDesc: 'Kısa link sistemimizde doğrulandı fakat yönlendirilmek istenen hedef sunucuya bağlantı kurulamadı. Hedef site geçici olarak kapalı olabilir.',
        canRetry: true
      }
    case 'disabled':
      return {
        pageTitle: `Link Devre Dışı (${alias.value}) — makeiTiny`,
        cardTitle: 'Bu kısa link şu an devre dışı.',
        cardSubtitle: `/${alias.value} kısa bağlantısı sahibi veya sistem yöneticisi tarafından geçici olarak pasife alınmış.`,
        badge: '410 • LİNK PASİF',
        boxTitle: 'Bağlantı erişime kapatılmış',
        boxDesc: 'Bu kısa link şu anda yönlendirme yapmamaktadır. Bağlantı sahibi panelden linki yeniden aktifleştirdiğinde çalışmaya devam edecektir.',
        canRetry: false
      }
    case 'blocked':
      return {
        pageTitle: `Güvenlik Engeli (${alias.value}) — makeiTiny`,
        cardTitle: 'Bağlantı güvenlik nedeniyle engellendi.',
        cardSubtitle: `/${alias.value} kısa bağlantısı güvenlik politikalarımız gereği yönlendirmeye kapatılmıştır.`,
        badge: '403 • GÜVENLİK ENGELİ',
        boxTitle: 'Güvenlik ve moderasyon engeli',
        boxDesc: 'Bu bağlantının hedef adresi zararlı içerik, oltalama (phishing) veya kullanım şartları ihlali nedeniyle moderasyon ekibi tarafından engellenmiştir.',
        canRetry: false
      }
    case 'expired':
      return {
        pageTitle: `Süresi Doldu (${alias.value}) — makeiTiny`,
        cardTitle: 'Bu kısa linkin süresi dolmuş.',
        cardSubtitle: `/${alias.value} bağlantısı için tanımlanan geçerlilik tarihi sona erdi.`,
        badge: '410 • SÜRESİ DOLDU',
        boxTitle: 'Zaman sınırı sona erdi',
        boxDesc: 'Link sahibi tarafından belirlenen son kullanma tarihi geçtiği için artık hedef adrese yönlendirme yapılmamaktadır.',
        canRetry: false
      }
    case 'limit_reached':
      return {
        pageTitle: `Tıklama Limiti Doldu (${alias.value}) — makeiTiny`,
        cardTitle: 'Maksimum tıklama sınırına ulaşıldı.',
        cardSubtitle: `/${alias.value} bağlantısı izin verilen azami tıklama sayısını doldurdu.`,
        badge: '410 • LİMİT DOLDU',
        boxTitle: 'Tıklama kotası tamamlandı',
        boxDesc: 'Bu bağlantı için belirlenen maksimum tıklama limiti dolduğu için yeni yönlendirmeler durdurulmuştur.',
        canRetry: false
      }
    case 'not_found':
      return {
        pageTitle: `Kısa Link Bulunamadı (${alias.value}) — makeiTiny`,
        cardTitle: 'Kısa link bulunamadı.',
        cardSubtitle: `/${alias.value} kısa bağlantısı veritabanımızda mevcut değil veya silinmiş.`,
        badge: '404 • KAYIT BULUNAMADI',
        boxTitle: 'Veritabanımızda böyle bir link yok',
        boxDesc: 'Aradığınız kısa bağlantı kodu sistemimizde kayıtlı değil. Adresi yanlış yazmış olabilirsiniz veya sahibi tarafından silinmiş olabilir.',
        canRetry: false
      }
    default:
      return {
        pageTitle: `Yönlendirme Hatası (${alias.value}) — makeiTiny`,
        cardTitle: 'Yönlendirme tamamlanamadı.',
        cardSubtitle: `/${alias.value} bağlantısı yönlendirilirken beklenmeyen bir sorun oluştu.`,
        badge: 'HATA • YÖNLENDİRME BAŞARISIZ',
        boxTitle: 'Beklenmeyen bir yönlendirme hatası oluştu',
        boxDesc: 'Kısa bağlantı çözümlenirken geçici bir sorun meydana geldi. Lütfen biraz sonra tekrar deneyin.',
        canRetry: true
      }
  }
})

useHead({ title: () => reasonConfig.value.pageTitle })

function retryRedirect() {
  if (!alias.value) return
  window.location.href = `${config.routerBase}/${encodeURIComponent(alias.value)}`
}
</script>

<template>
  <AuthCard :title="reasonConfig.cardTitle" :subtitle="reasonConfig.cardSubtitle">
    <div class="flex flex-col gap-5">
      <div class="flex items-center justify-between rounded-md border border-line bg-paper-2 px-3.5 py-2.5">
        <span class="font-mono text-xs text-ink-2">/{{ alias }}</span>
        <UiBadge tone="accent">{{ reasonConfig.badge }}</UiBadge>
      </div>

      <div class="flex flex-col gap-2 rounded-md border border-accent-dark bg-accent-soft p-4 text-sm text-accent-dark">
        <div class="flex items-center gap-2 font-semibold">
          <UiIcon name="alert" :size="18" />
          <span>{{ reasonConfig.boxTitle }}</span>
        </div>
        <p class="text-xs leading-relaxed text-ink-2">
          {{ reasonConfig.boxDesc }}
        </p>
      </div>

      <div class="flex flex-col gap-2.5">
        <UiButton
          v-if="reasonConfig.canRetry"
          type="button"
          block
          @click="retryRedirect"
        >
          <UiIcon name="refresh" :size="16" /> Tekrar Dene
        </UiButton>

        <UiButton to="/" :variant="reasonConfig.canRetry ? 'secondary' : 'primary'" block>
          <UiIcon name="back" :size="16" /> Ana sayfaya dön
        </UiButton>
      </div>
    </div>

    <template #footer>
      Kendi kısa bağlantınızı oluşturmak için
      <NuxtLink to="/dashboard/new" class="font-semibold text-accent-dark underline underline-offset-2">
        yeni link kısaltın
      </NuxtLink>
    </template>
  </AuthCard>
</template>
