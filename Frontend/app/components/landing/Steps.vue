<script setup lang="ts">
const steps = [
  {
    badge: 'Adım 01',
    title: 'Hedefi ve koruma modunu seç',
    text: 'Uzun web adresini yapıştır. İstersen özel takma ad belirle, Argon2id parola koruması ekle veya saniye bazlı ara bekletme (/w) sayfasını aç.'
  },
  {
    badge: 'Adım 02',
    title: 'Kısa linkini hemen paylaş',
    text: 'Oluşturulan /r/senin-adin kısa bağlantısını tek tıkla panoya kopyala. Linkin anında Redis önbelleğine yazılır ve yönlendirmeye hazır hale gelir.'
  },
  {
    badge: 'Adım 03',
    title: 'Logları, grafikleri ve ülkeleri izle',
    text: 'Link detay sayfasından sayfalanmış tıklama kayıtlarını, günlük/saatlik grafikleri, cihaz kırılımlarını ve tarayıcı diline dayalı coğrafi dağılımı takip et.'
  }
]
</script>

<template>
  <section id="nasil" class="container-x py-16 md:py-24">
    <div class="flex flex-col justify-between gap-4 md:flex-row md:items-end">
      <div>
        <p class="eyebrow mb-3">Nasıl çalışır</p>
        <h2 class="max-w-xl text-4xl font-extrabold md:text-5xl">Üç adımda oluştur, yönet ve analiz et.</h2>
      </div>
      <p class="max-w-md text-sm text-ink-2">
        Karmaşık ayarlar olmadan saniyeler içinde kısa linkini üretir, dilediğin an hedef adresi veya koruma modunu güncelleyebilirsin.
      </p>
    </div>

    <ol class="mt-12 grid gap-6 md:grid-cols-3">
      <li
        v-for="(s, i) in steps"
        :key="s.title"
        class="relative flex flex-col justify-between rounded-lg border-[1.5px] border-ink bg-card p-6 shadow-hard"
      >
        <div>
          <div class="flex items-center justify-between">
            <span class="rounded-sm border border-ink bg-accent-soft px-2.5 py-0.5 font-mono text-xs font-bold text-accent-dark">
              {{ s.badge }}
            </span>
            <span class="font-display text-4xl font-extrabold leading-none text-paper-2 [-webkit-text-stroke:1.5px_var(--color-ink)]">
              {{ i + 1 }}
            </span>
          </div>
          <h3 class="mt-5 text-xl font-bold">{{ s.title }}</h3>
          <p class="mt-2.5 text-sm leading-relaxed text-ink-2">{{ s.text }}</p>
        </div>
      </li>
    </ol>
  </section>
</template>
