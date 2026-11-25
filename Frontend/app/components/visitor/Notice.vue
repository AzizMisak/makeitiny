<script setup lang="ts">
/** Ziyaretçinin hedefe ulaşamadığı durumlar. Link, ortasından kopmuş bir şerit olarak gösterilir. */
const props = defineProps<{ title: string; text: string; alias?: string }>()
const { display } = useShortUrl()

const label = computed(() => (props.alias ? display(props.alias) : display('?')))
const cut = computed(() => Math.ceil(label.value.length * 0.6))
</script>

<template>
  <div class="mx-auto flex w-full max-w-2xl flex-1 flex-col justify-center px-5 pb-24">
    <div class="mb-12 flex min-w-0 items-start font-mono text-[clamp(1rem,4vw,1.5rem)] font-semibold text-on-mark" :aria-label="label" role="img">
      <span class="truncate bg-mark py-4 pl-6 pr-5" style="clip-path: polygon(0 0, 100% 0, calc(100% - 10px) 28%, 100% 52%, calc(100% - 12px) 76%, calc(100% - 4px) 100%, 0 100%)">{{ label.slice(0, cut) }}</span>
      <span class="ml-3 mt-5 rotate-6 truncate bg-mark py-4 pl-5 pr-6" style="clip-path: polygon(6px 0, 100% 0, 100% 100%, 8px 100%, 0 76%, 12px 52%, 2px 28%)">{{ label.slice(cut) }}</span>
    </div>

    <h1 class="display text-[clamp(2.25rem,7vw,3.5rem)]">{{ title }}</h1>
    <p class="mt-4 max-w-xl text-lg text-ink-2">{{ text }}</p>
    <div class="mt-9 flex flex-wrap gap-2">
      <slot />
      <UiButton to="/" variant="outline">Ana sayfa</UiButton>
    </div>
  </div>
</template>
