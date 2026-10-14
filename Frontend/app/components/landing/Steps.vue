<script setup lang="ts">
const steps = [
  { title: 'Linki yapıştır', text: 'Uzun adresi kutuya bırak. İstersen takma ad ve parola ekle.' },
  { title: 'Kısa adresi al', text: 'Saniyeler içinde mkit.ee/senin-adin gibi paylaşılabilir bir link oluşur.' },
  { title: 'Sonuçları izle', text: 'Panelde her linkin kaç kez açıldığını gör, istediğin an düzenle ya da sil.' }
]
</script>

<template>
  <section id="nasil" class="container-x py-16 md:py-24">
    <p class="eyebrow mb-3">Nasıl çalışır</p>
    <h2 class="max-w-xl text-4xl font-bold md:text-5xl">Üç adım, hepsi bu.</h2>

    <ol class="mt-12 grid gap-10 md:grid-cols-3">
      <li v-for="(s, i) in steps" :key="s.title" class="relative">
        <span class="font-display text-7xl font-extrabold leading-none text-paper-2 [-webkit-text-stroke:1.5px_var(--color-ink)]">
          {{ i + 1 }}
        </span>
        <h3 class="mt-4 text-xl font-bold">{{ s.title }}</h3>
        <p class="mt-2 text-ink-2">{{ s.text }}</p>
      </li>
    </ol>
  </section>
</template>
