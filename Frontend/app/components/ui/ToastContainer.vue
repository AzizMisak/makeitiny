<script setup lang="ts">
const { toasts, remove } = useToast()
</script>

<template>
  <div
    aria-live="polite"
    class="pointer-events-none fixed bottom-5 right-5 z-50 flex w-full max-w-sm flex-col gap-2.5 px-4 sm:px-0"
  >
    <div
      v-for="item in toasts"
      :key="item.id"
      class="pointer-events-auto flex items-start justify-between gap-3 rounded-md border-[1.5px] border-ink bg-card p-3.5 shadow-hard-lg"
    >
      <div class="flex items-start gap-2.5">
        <span
          class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-sm border border-ink text-xs font-bold"
          :class="
            item.tone === 'ok'
              ? 'bg-ok-soft text-ok'
              : item.tone === 'error'
                ? 'bg-accent-soft text-accent-dark'
                : 'bg-paper-2 text-ink'
          "
        >
          <UiIcon :name="item.tone === 'ok' ? 'check' : 'alert'" :size="12" />
        </span>
        <p class="text-xs font-semibold leading-relaxed text-ink">{{ item.message }}</p>
      </div>

      <button
        type="button"
        class="shrink-0 cursor-pointer rounded px-1.5 py-0.5 font-mono text-xs text-ink-3 hover:bg-paper-2 hover:text-ink"
        aria-label="Bildirimi kapat"
        @click="remove(item.id)"
      >
        ✕
      </button>
    </div>
  </div>
</template>
