<script setup lang="ts">
const features = [
  {
    title: 'Kendi takma adın',
    text: 'Rastgele kod yerine akılda kalan bir ad yaz. Yazarken adın boş olup olmadığını anında görürsün.'
  },
  {
    title: 'Gerçek tıklama sayısı',
    text: 'Yönlendirme önbelleğe alınmaz, bu yüzden her tıklama sayılır. Rakamlar tahmin değil, ölçümdür.'
  },
  {
    title: 'Şifreli linkler',
    text: 'Linke parola koy. Ziyaretçi parolayı girmeden hedefi göremez, yönlendirilmez.'
  },
  {
    title: 'Önizleme sayfası',
    text: 'Paylaştığın kısa linkin nereye gittiğini tıklamadan görmek mümkün. Güven veren bir detay.'
  }
]
</script>

<template>
  <section id="ozellikler" class="border-y border-ink bg-card py-16 md:py-24">
    <div class="container-x">
      <p class="eyebrow mb-3">Özellikler</p>
      <h2 class="max-w-xl text-4xl font-bold md:text-5xl">Az şey yapar, hepsini düzgün yapar.</h2>

      <ul class="mt-12 grid border-l border-t border-line sm:grid-cols-2">
        <li v-for="(f, i) in features" :key="f.title" class="border-b border-r border-line p-7 md:p-9">
          <span class="font-mono text-sm text-accent-dark">0{{ i + 1 }}</span>
          <h3 class="mt-3 text-2xl font-bold">{{ f.title }}</h3>
          <p class="mt-3 max-w-sm text-ink-2">{{ f.text }}</p>
        </li>
      </ul>
    </div>
  </section>
</template>
