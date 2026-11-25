<script setup lang="ts">
/** Arama kutusu. Enter ile arar; temizle düğmesi kutunun içinde, yeri sabit. */
const props = defineProps<{ id: string; placeholder: string; loading?: boolean; active?: string }>()
const emit = defineEmits<{ search: [value: string] }>()

const text = ref(props.active ?? '')

function submit() {
  emit('search', text.value.trim())
}

function clear() {
  text.value = ''
  if (props.active) emit('search', '')
}
</script>

<template>
  <form
    role="search"
    class="flex min-h-11 w-full min-w-0 items-center rounded-md border border-line-2 bg-surface focus-within:border-brand focus-within:ring-1 focus-within:ring-brand sm:max-w-md"
    @submit.prevent="submit"
  >
    <label :for="id" class="sr-only">{{ placeholder }}</label>
    <UiIcon name="search" :size="17" class="ml-3.5 text-ink-3" />
    <input
      :id="id"
      v-model="text"
      type="search"
      :placeholder="placeholder"
      class="w-full min-w-0 bg-transparent px-3 py-2 text-[15px] outline-none placeholder:text-ink-3/80"
    />
    <span class="flex w-10 items-center justify-center">
      <UiSpinner v-if="loading && active" :size="15" class="text-ink-3" />
      <button v-else-if="text" type="button" class="text-ink-3 hover:text-ink" aria-label="Aramayı temizle" @click="clear">
        <UiIcon name="x" :size="16" />
      </button>
    </span>
    <button type="submit" class="self-stretch border-l border-line px-4 text-sm font-medium text-ink-2 hover:bg-sunken hover:text-ink">
      Ara
    </button>
  </form>
</template>
