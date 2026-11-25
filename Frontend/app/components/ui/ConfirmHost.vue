<script setup lang="ts">
const { state, options, accept, cancel } = useConfirm()

const typed = ref('')

watch(
  () => state.value.open,
  (open) => {
    if (open) typed.value = ''
  }
)

const locked = computed(() => {
  const expected = options.value?.typeToConfirm
  return Boolean(expected) && typed.value.trim() !== expected
})
</script>

<template>
  <UiModal :open="state.open" :title="options?.title ?? ''" :busy="state.busy" @close="cancel">
    <form id="confirm-form" class="flex flex-col gap-4 pt-2" @submit.prevent="!locked && accept()">
      <p v-if="options?.subject" class="break-all border-l-4 border-mark bg-sunken px-4 py-2.5 font-medium">
        {{ options.subject }}
      </p>

      <p class="leading-relaxed text-ink-2">{{ options?.message }}</p>

      <ul v-if="options?.effects?.length" class="flex flex-col gap-2 text-sm text-ink-2">
        <li v-for="e in options.effects" :key="e" class="flex gap-2.5">
          <UiIcon name="right" :size="16" class="mt-0.5 text-ink-3" />
          {{ e }}
        </li>
      </ul>

      <UiChangeList v-if="options?.changes?.length" :changes="options.changes" />

      <UiField
        v-if="options?.typeToConfirm"
        id="confirm-type"
        v-model="typed"
        :label="`Onaylamak için ${options.typeToConfirm} yaz`"
        autocomplete="off"
        :disabled="state.busy"
      />

      <p v-if="state.error" role="alert" class="bg-danger-soft px-4 py-2.5 text-sm text-danger">{{ state.error }}</p>
    </form>

    <template #footer>
      <UiButton variant="quiet" :disabled="state.busy" data-autofocus @click="cancel">
        {{ options?.cancelLabel ?? 'Vazgeç' }}
      </UiButton>
      <UiButton
        type="submit"
        form="confirm-form"
        :variant="options?.tone === 'danger' ? 'destroy' : 'primary'"
        :loading="state.busy"
        loading-label="İşleniyor"
        :disabled="locked"
      >
        {{ options?.confirmLabel }}
      </UiButton>
    </template>
  </UiModal>
</template>
