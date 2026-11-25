<script setup lang="ts">
const { toasts, remove } = useToast()
</script>

<template>
  <div
    aria-live="polite"
    class="pointer-events-none fixed inset-x-0 bottom-20 z-[90] flex flex-col items-center gap-2 px-4 lg:bottom-6"
  >
    <TransitionGroup name="pop">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="pointer-events-auto flex w-full max-w-md items-center gap-3 rounded-md bg-ink py-2 pl-4 pr-1.5 text-ground shadow-pop"
      >
        <UiIcon
          :name="t.tone === 'ok' ? 'check' : t.tone === 'error' ? 'alert' : 'info'"
          :size="18"
          :class="t.tone === 'ok' ? 'text-mark' : ''"
        />
        <p class="min-w-0 flex-1 text-sm font-medium">{{ t.message }}</p>
        <button
          type="button"
          class="inline-flex h-9 w-9 items-center justify-center rounded-md opacity-70 hover:opacity-100"
          aria-label="Bildirimi kapat"
          @click="remove(t.id)"
        >
          <UiIcon name="x" :size="16" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>
