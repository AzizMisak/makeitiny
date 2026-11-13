<script setup lang="ts">
const { isLoggedIn } = useAuth()
</script>

<template>
  <section class="container-x pb-20">
    <div class="flex flex-col items-start justify-between gap-8 rounded-lg border-[1.5px] border-ink bg-ink p-8 text-paper shadow-hard-lg md:flex-row md:items-center md:p-12">
      <div>
        <span class="inline-block rounded-sm bg-accent px-2.5 py-0.5 font-mono text-xs font-bold uppercase tracking-wider text-white">
          Hemen Başla
        </span>
        <h2 class="mt-3 text-3xl font-extrabold md:text-4xl">
          {{ isLoggedIn ? 'Yeni bir bağlantıyı hemen kısalt.' : 'İlk kısa linkini şimdi oluştur.' }}
        </h2>
        <p class="mt-2.5 max-w-lg text-sm leading-relaxed text-paper/80 sm:text-base">
          {{
            isLoggedIn
              ? 'Yönetim panelinden tüm kısa linklerini düzenleyebilir, parola/bekletme ayarlarını değiştirebilir ve detaylı tıklama grafiklerini inceleyebilirsin.'
              : 'Ücretsiz hesabını saniyeler içinde aç; özel takma ad, şifreli yönlendirme ve detaylı log analitiğini hemen kullanmaya başla.'
          }}
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <UiButton v-if="isLoggedIn" to="/dashboard/new" variant="primary">
          <UiIcon name="plus" :size="16" /> Yeni link oluştur
        </UiButton>
        <UiButton v-else to="/auth/register" variant="primary">
          Ücretsiz hesap aç <UiIcon name="arrow" :size="16" />
        </UiButton>
      </div>
    </div>
  </section>
</template>
