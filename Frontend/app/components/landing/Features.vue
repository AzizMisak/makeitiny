<script setup lang="ts">
const features = [
  {
    icon: 'link',
    badge: 'Base62 & Özel Alias',
    title: 'Özel takma ad veya otomatik kod',
    text: 'Kendi belirlediğin 3–32 karakterlik özel takma adı kullan ya da alanı boş bırakarak sistemin benzersiz 6 haneli Base62 kod üretmesini sağla.'
  },
  {
    icon: 'lock',
    badge: 'Argon2id Güvenliği',
    title: 'Parola korumalı linkler (/unlock)',
    text: 'Linkine parola koyduğunda hedef adres gizlenir. Ziyaretçiler /unlock ekranında doğru parolayı girmeden hedef siteye yönlendirilmez.'
  },
  {
    icon: 'clock',
    badge: 'Geri Sayım & Reklam',
    title: 'Ara bekletme sayfası (/w)',
    text: 'Şifresiz linklerde istediğin saniye kadar bekletme süresi tanımla. Ziyaretçiler /w ara sayfasında geri sayımı bekledikten sonra otomatik yönlendirilir.'
  },
  {
    icon: 'chart',
    badge: 'Redis + PostgreSQL',
    title: 'Gerçek zamanlı tıklama sayımı',
    text: 'Yönlendirmeler Redis Hash hızında çalışır ve no-store başlığıyla tarayıcı önbelleğine takılmaz; böylece her tıklama eksiksiz kaydedilir.'
  },
  {
    icon: 'sliders',
    badge: 'Detaylı Loglama',
    title: 'Cihaz, tarayıcı ve kaynak takibi',
    text: 'Tıklayan her ziyaretçinin IP adresi, cihaz türü (Mobil/Masaüstü/Tablet), tarayıcısı, işletim sistemi ve geldiği kaynak (Referer) loglanır.'
  },
  {
    icon: 'alert',
    badge: 'Hedef Doğrulama',
    title: 'Erişilebilirlik ve akıllı hata ekranı',
    text: 'Hedef sitenin yanıt verip vermediği kontrol edilir. Silinmiş bir kısa link veya ulaşılamayan bir hedef varsa kullanıcı anlaşılır hata ekranına yönlendirilir.'
  }
]

const sampleBars = [
  { day: 'Pzt', val: 42 },
  { day: 'Sal', val: 68 },
  { day: 'Çar', val: 55 },
  { day: 'Per', val: 92 },
  { day: 'Cum', val: 100 },
  { day: 'Cmt', val: 76 },
  { day: 'Paz', val: 61 }
]

const sampleGeo = [
  { code: 'TR', name: 'Türkiye', lang: 'tr-TR', pct: 68 },
  { code: 'US', name: 'Amerika Birleşik Devletleri', lang: 'en-US', pct: 18 },
  { code: 'DE', name: 'Almanya', lang: 'de-DE', pct: 9 },
  { code: 'AZ', name: 'Azerbaycan', lang: 'az-AZ', pct: 5 }
]
</script>

<template>
  <!-- Özellikler Bölümü -->
  <section id="ozellikler" class="border-b-[1.5px] border-ink bg-card py-16 md:py-24">
    <div class="container-x">
      <div class="flex flex-col justify-between gap-4 md:flex-row md:items-end">
        <div>
          <p class="eyebrow mb-3">Özellikler</p>
          <h2 class="max-w-xl text-4xl font-extrabold md:text-5xl">
            Sadece kısaltmaz; korur, bekletir ve ölçer.
          </h2>
        </div>
        <p class="max-w-md text-sm leading-relaxed text-ink-2">
          makeiTiny; yüksek hızlı Go yönlendirme servisi, Redis önbellekleme ve zengin yönetim paneliyle linklerin üzerindeki tüm kontrolü sana verir.
        </p>
      </div>

      <ul class="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        <li
          v-for="(f, i) in features"
          :key="f.title"
          class="flex flex-col justify-between rounded-lg border-[1.5px] border-ink bg-paper p-6 shadow-hard transition-transform duration-150 hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-hard-lg"
        >
          <div>
            <div class="flex items-center justify-between gap-2">
              <span class="inline-flex h-9 w-9 items-center justify-center rounded-md border-[1.5px] border-ink bg-card text-ink">
                <UiIcon :name="f.icon" :size="18" />
              </span>
              <span class="font-mono text-xs font-bold text-accent-dark">0{{ i + 1 }} · {{ f.badge }}</span>
            </div>
            <h3 class="mt-5 text-xl font-bold">{{ f.title }}</h3>
            <p class="mt-2.5 text-sm leading-relaxed text-ink-2">{{ f.text }}</p>
          </div>
        </li>
      </ul>
    </div>
  </section>

  <!-- Analitik ve Log Okuma Tanıtım Bölümü -->
  <section id="analitik" class="border-b-[1.5px] border-ink bg-paper py-16 md:py-24">
    <div class="container-x">
      <div class="max-w-2xl">
        <p class="eyebrow mb-3">Analitik &amp; Log Okuma</p>
        <h2 class="text-4xl font-extrabold md:text-5xl">
          Her linkin kendi sayfasında üç katmanlı log analizi.
        </h2>
        <p class="mt-4 text-base text-ink-2">
          Link detay sayfasında ham tıklama kayıtlarını 20'şerli sayfalar halinde inceleyebilir, hemen altındaki iki ayrı analiz kutusunda zaman/cihaz grafiklerini ve tarayıcı dilinden çözümlenen coğrafi dağılımı okuyabilirsin.
        </p>
      </div>

      <div class="mt-10 grid gap-8 lg:grid-cols-2">
        <!-- Örnek Kutu 1: Grafiksel Log Analizi -->
        <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
          <div class="flex items-center justify-between gap-2">
            <div>
              <span class="eyebrow">1. Kutu Önizlemesi</span>
              <h3 class="mt-1 text-xl font-bold">Zaman ve Kategori Grafikleri</h3>
            </div>
            <UiBadge tone="accent"><UiIcon name="chart" :size="12" /> 7 / 14 / 30 Gün &amp; Saatlik</UiBadge>
          </div>
          <p class="mt-2 text-xs text-ink-3">
            Günlük veya saatlik tıklama yoğunluğunu sütun grafiğinde gör; Cihaz, Tarayıcı, İşletim Sistemi ve Ülke kırılımlarını karşılaştır.
          </p>

          <div class="mt-5 rounded-md border border-line bg-paper p-4">
            <div class="flex items-center justify-between text-xs">
              <span class="font-bold text-ink">Son 7 Gün Tıklama Grafiği</span>
              <span class="font-mono text-ink-3">Zirve: Cuma</span>
            </div>
            <div class="mt-4 grid h-28 grid-cols-7 items-end gap-2 pt-2">
              <div v-for="b in sampleBars" :key="b.day" class="flex h-full flex-col items-center justify-end">
                <div class="flex w-full flex-1 items-end justify-center">
                  <div
                    class="w-full max-w-7 rounded-t-sm border border-ink bg-accent"
                    :style="{ height: `${b.val}%` }"
                  />
                </div>
                <span class="mt-1.5 font-mono text-[10px] text-ink-3">{{ b.day }}</span>
              </div>
            </div>
          </div>

          <div class="mt-4 flex flex-wrap gap-2">
            <UiBadge tone="neutral">Masaüstü / Mobil / Tablet</UiBadge>
            <UiBadge tone="neutral">Chrome / Safari / Firefox / Edge</UiBadge>
            <UiBadge tone="ok">Windows / macOS / Android / iOS</UiBadge>
          </div>
        </div>

        <!-- Örnek Kutu 2: Coğrafi Log Okuma -->
        <div class="rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard">
          <div class="flex items-center justify-between gap-2">
            <div>
              <span class="eyebrow">2. Kutu Önizlemesi</span>
              <h3 class="mt-1 text-xl font-bold">Tarayıcı Dili Bazlı Coğrafi Okuma</h3>
            </div>
            <UiBadge tone="ok">Accept-Language Çözümleyici</UiBadge>
          </div>
          <p class="mt-2 text-xs text-ink-3">
            Ziyaretçilerin tarayıcı dili (<span class="font-mono">tr-TR</span>, <span class="font-mono">en-US</span>, <span class="font-mono">de-DE</span> vb.) verisinden ülke ve bölge yoğunluğu doğrudan arayüzde çözümlenir.
          </p>

          <div class="mt-5 flex flex-col gap-3">
            <div
              v-for="row in sampleGeo"
              :key="row.code"
              class="rounded-md border border-line bg-paper p-3"
            >
              <div class="flex items-center justify-between text-xs">
                <div class="flex items-center gap-2">
                  <span class="rounded-xs border border-ink bg-accent-soft px-1.5 py-0.5 font-mono text-[10px] font-bold text-accent-dark">
                    {{ row.code }}
                  </span>
                  <span class="font-bold text-ink">{{ row.name }}</span>
                  <span class="font-mono text-[11px] text-ink-3">({{ row.lang }})</span>
                </div>
                <span class="font-mono font-bold text-ink">%{{ row.pct }}</span>
              </div>
              <div class="mt-2 h-2 w-full overflow-hidden rounded-sm border border-ink/40 bg-card">
                <div class="h-full bg-ok" :style="{ width: `${row.pct}%` }" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
