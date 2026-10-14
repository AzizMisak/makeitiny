<template>
  <div
    class="relative rounded-lg border-[1.5px] border-ink bg-card p-5 shadow-hard-lg sm:p-6"
    role="img"
    aria-label="Uzun bir bağlantının kısa adrese dönüşümünü gösteren örnek"
  >
    <div class="mb-5 flex items-center justify-between">
      <span class="eyebrow">Önce</span>
      <span class="font-mono text-xs text-ink-3">127 karakter</span>
    </div>
    <p class="break-all font-mono text-[13px] leading-relaxed text-ink-3 line-through decoration-accent decoration-2">
      https://magaza.example.com/kampanyalar/yaz-indirimi-2026?utm_source=mail&amp;utm_medium=bulten&amp;utm_campaign=yaz&amp;ref=a81f
    </p>

    <div class="my-6 flex items-center gap-3 text-ink-3" aria-hidden="true">
      <span class="h-px flex-1 bg-line" />
      <UiIcon name="arrow" />
      <span class="h-px flex-1 bg-line" />
    </div>

    <div class="mb-3 flex items-center justify-between">
      <span class="eyebrow">Sonra</span>
      <span class="font-mono text-xs text-ink-3">17 karakter</span>
    </div>
    <div class="flex items-center justify-between gap-3 rounded-md border-[1.5px] border-ink bg-paper px-4 py-3">
      <span class="font-mono text-lg font-medium">mkit.ee/<span class="text-accent">yaz26</span></span>
      <UiBadge tone="ok"><UiIcon name="check" :size="12" /> aktif</UiBadge>
    </div>

    <dl class="mt-5 grid grid-cols-3 gap-3 border-t border-line pt-4 text-sm">
      <div>
        <dt class="text-xs text-ink-3">Tıklama</dt>
        <dd class="font-display text-xl font-bold">1.284</dd>
      </div>
      <div>
        <dt class="text-xs text-ink-3">Yönlendirme</dt>
        <dd class="font-display text-xl font-bold">302</dd>
      </div>
      <div>
        <dt class="text-xs text-ink-3">Koruma</dt>
        <dd class="font-display text-xl font-bold">Şifreli</dd>
      </div>
    </dl>
  </div>
</template>
