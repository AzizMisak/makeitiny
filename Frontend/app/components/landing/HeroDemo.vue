<script setup lang="ts">
type DemoMode = 'direct' | 'locked' | 'wait'

const activeMode = ref<DemoMode>('direct')

const modeDetails = computed(() => {
  switch (activeMode.value) {
    case 'locked':
      return {
        badgeTone: 'accent' as const,
        badgeIcon: 'lock',
        badgeText: 'şifreli (/unlock)',
        routeText: '/unlock/kampanya-26',
        statusTitle: 'Argon2id Parola Koruması',
        statusDesc: 'Ziyaretçi doğru parolayı girmeden hedef URL gizli kalır ve yönlendirme yapılmaz.'
      }
    case 'wait':
      return {
        badgeTone: 'neutral' as const,
        badgeIcon: 'clock',
        badgeText: '5 sn bekletmeli (/w)',
        routeText: '/w/kampanya-26',
        statusTitle: 'Ara Bekletme & Reklam Alanı',
        statusDesc: 'Ziyaretçi 5 saniyelik geri sayım ve sponsorlu duyuru ekranından sonra otomatik yönlendirilir.'
      }
    default:
      return {
        badgeTone: 'ok' as const,
        badgeIcon: 'check',
        badgeText: 'doğrudan (302)',
        routeText: '302 Found → Hedef',
        statusTitle: 'Redis Önbellekli Anında Yönlendirme',
        statusDesc: 'Link Redis Hash üzerinden okunur, tıklama logu kaydedilir ve anında hedefe yönlendirilir.'
      }
  }
})
</script>

<template>
  <div
    class="relative rounded-lg border-[1.5px] border-ink bg-card p-5 shadow-hard-lg sm:p-6"
    role="region"
    aria-label="makeiTiny kısa link ve analitik önizlemesi"
  >
    <!-- Üst Başlık Çubuğu -->
    <div class="mb-4 flex items-center justify-between border-b border-line pb-3">
      <div class="flex items-center gap-1.5">
        <span class="h-2.5 w-2.5 rounded-full border border-ink bg-accent" />
        <span class="h-2.5 w-2.5 rounded-full border border-ink bg-paper-2" />
        <span class="h-2.5 w-2.5 rounded-full border border-ink bg-ok" />
        <span class="ml-1.5 font-mono text-xs font-semibold text-ink-2">makeiTiny · Canlı Link Kartı</span>
      </div>
      <span class="rounded-sm bg-paper-2 px-2 py-0.5 font-mono text-[11px] font-medium text-ink-2">
        ID #42
      </span>
    </div>

    <!-- Uzun Hedef Adres -->
    <div class="rounded-md border border-line bg-paper p-3.5">
      <div class="mb-1.5 flex items-center justify-between">
        <span class="eyebrow">Hedef Adres (Uzun URL)</span>
        <span class="font-mono text-[11px] text-ink-3">114 karakter</span>
      </div>
      <p class="break-all font-mono text-xs leading-relaxed text-ink-3 line-through decoration-accent decoration-2">
        https://magaza.ornek.com/urunler/kampanya-2026?utm_source=bulten&amp;utm_medium=paylasim&amp;ref=mit94
      </p>
    </div>

    <!-- Kısa Link Çıktısı -->
    <div class="mt-3 rounded-md border-[1.5px] border-ink bg-paper px-4 py-3 shadow-[2px_2px_0_var(--color-ink)]">
      <div class="mb-1 flex items-center justify-between">
        <span class="eyebrow">Kısaltılmış Adres</span>
        <UiBadge :tone="modeDetails.badgeTone">
          <UiIcon :name="modeDetails.badgeIcon" :size="11" /> {{ modeDetails.badgeText }}
        </UiBadge>
      </div>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="font-mono text-base font-bold sm:text-lg">
          localhost/r/<span class="text-accent">kampanya-26</span>
        </span>
        <span class="font-mono text-xs font-semibold text-ink-2">{{ modeDetails.routeText }}</span>
      </div>
    </div>

    <!-- Yönlendirme Modu Seçici (İnteraktif Önizleme) -->
    <div class="mt-4">
      <p class="mb-2 font-mono text-[11px] font-semibold uppercase tracking-wider text-ink-3">
        Yönlendirme Modunu Dene:
      </p>
      <div class="grid grid-cols-3 gap-1.5">
        <button
          type="button"
          class="rounded-sm border px-2 py-1.5 font-mono text-xs font-semibold transition-all cursor-pointer"
          :class="
            activeMode === 'direct'
              ? 'border-ink bg-ink text-white'
              : 'border-line bg-paper text-ink-2 hover:border-ink hover:text-ink'
          "
          @click="activeMode = 'direct'"
        >
          Doğrudan
        </button>
        <button
          type="button"
          class="rounded-sm border px-2 py-1.5 font-mono text-xs font-semibold transition-all cursor-pointer"
          :class="
            activeMode === 'locked'
              ? 'border-ink bg-accent text-white'
              : 'border-line bg-paper text-ink-2 hover:border-ink hover:text-ink'
          "
          @click="activeMode = 'locked'"
        >
          Şifreli
        </button>
        <button
          type="button"
          class="rounded-sm border px-2 py-1.5 font-mono text-xs font-semibold transition-all cursor-pointer"
          :class="
            activeMode === 'wait'
              ? 'border-ink bg-ok text-white'
              : 'border-line bg-paper text-ink-2 hover:border-ink hover:text-ink'
          "
          @click="activeMode = 'wait'"
        >
          Bekletmeli (/w)
        </button>
      </div>

      <div class="mt-2.5 rounded-md border border-line bg-paper/70 p-3">
        <p class="text-xs font-bold text-ink">{{ modeDetails.statusTitle }}</p>
        <p class="mt-0.5 text-xs text-ink-2">{{ modeDetails.statusDesc }}</p>
      </div>
    </div>

    <!-- Alt Mini Log & İstatistik Özeti -->
    <div class="mt-4 border-t border-line pt-4">
      <div class="grid grid-cols-3 gap-2.5 text-center">
        <div class="rounded-md border border-line bg-paper p-2.5">
          <p class="text-[11px] text-ink-3">Toplam Tıklama</p>
          <p class="mt-0.5 font-display text-lg font-extrabold text-ink">1.284</p>
        </div>
        <div class="rounded-md border border-line bg-paper p-2.5">
          <p class="text-[11px] text-ink-3">En Yoğun Ülke</p>
          <p class="mt-0.5 font-display text-lg font-extrabold text-ink">TR %74</p>
        </div>
        <div class="rounded-md border border-line bg-paper p-2.5">
          <p class="text-[11px] text-ink-3">Lider Cihaz</p>
          <p class="mt-0.5 font-display text-lg font-extrabold text-ink">Mobil</p>
        </div>
      </div>

      <div class="mt-2.5 flex items-center justify-between rounded-sm border border-line bg-paper px-3 py-1.5 font-mono text-[11px] text-ink-2">
        <span>Son log: Chrome · Windows · tr-TR</span>
        <span class="text-ok font-semibold">● kaydedildi</span>
      </div>
    </div>
  </div>
</template>
