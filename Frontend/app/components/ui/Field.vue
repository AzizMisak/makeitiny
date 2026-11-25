<script setup lang="ts">
/**
 * Metin alanı. Alt satır her zaman yer tutar: ipucu, durum ve hata aynı satırda
 * birbirinin yerine geçer, böylece mesaj çıkınca form zıplamaz.
 */
const props = withDefaults(
  defineProps<{
    id: string
    label?: string
    type?: string
    placeholder?: string
    hint?: string
    error?: string
    ok?: string
    prefix?: string
    suffix?: string
    autocomplete?: string
    inputmode?: 'text' | 'numeric' | 'decimal' | 'url' | 'email'
    min?: string | number
    max?: string | number
    step?: string | number
    maxlength?: number
    disabled?: boolean
    busy?: boolean
    mono?: boolean
    large?: boolean
    /** Alt satır hiç kullanılmayacaksa yer ayırma. */
    bare?: boolean
  }>(),
  { type: 'text' }
)

const model = defineModel<string>({ default: '' })
const revealed = ref(false)
const inputType = computed(() => (props.type === 'password' && revealed.value ? 'text' : props.type))
const message = computed(() => props.error || props.ok || props.hint || '')
</script>

<template>
  <div class="flex min-w-0 flex-col">
    <label v-if="label" :for="id" class="mb-1.5 flex items-center gap-2 text-sm font-semibold text-ink">
      {{ label }}
      <slot name="label-end" />
    </label>

    <div
      class="flex items-stretch overflow-hidden rounded-md border bg-surface transition-colors focus-within:border-brand focus-within:ring-1 focus-within:ring-brand"
      :class="[
        error ? 'border-danger' : 'border-line-2',
        disabled ? 'border-dashed bg-transparent' : '',
        large ? 'min-h-14 text-lg' : 'min-h-11 text-[15px]'
      ]"
    >
      <span v-if="prefix" class="flex items-center pl-3.5 text-ink-3" :class="mono ? 'font-mono text-sm' : ''">{{ prefix }}</span>
      <input
        :id="id"
        :value="model"
        :type="inputType"
        :placeholder="placeholder"
        :autocomplete="autocomplete"
        :inputmode="inputmode"
        :min="min"
        :max="max"
        :step="step"
        :maxlength="maxlength"
        :disabled="disabled"
        :aria-invalid="!!error || undefined"
        :aria-describedby="bare ? undefined : `${id}-msg`"
        class="w-full min-w-0 bg-transparent py-2 pr-3.5 outline-none placeholder:text-ink-3/80 disabled:cursor-not-allowed disabled:text-ink-3"
        :class="[mono ? 'font-mono text-sm' : '', prefix ? 'pl-0.5' : 'pl-3.5']"
        @input="model = ($event.target as HTMLInputElement).value"
      />
      <span v-if="suffix" class="flex items-center pr-3.5 text-sm text-ink-3">{{ suffix }}</span>
      <span v-if="busy || ok" class="flex w-10 items-center justify-center" :class="ok ? 'text-ok' : 'text-ink-3'">
        <UiSpinner v-if="busy" :size="15" />
        <UiIcon v-else name="check" :size="17" />
      </span>
      <button
        v-if="type === 'password' && !disabled"
        type="button"
        class="flex w-11 items-center justify-center text-ink-3 hover:text-ink"
        :aria-label="revealed ? 'Parolayı gizle' : 'Parolayı göster'"
        :aria-pressed="revealed"
        @click="revealed = !revealed"
      >
        <UiIcon :name="revealed ? 'eye-off' : 'eye'" :size="17" />
      </button>
    </div>

    <p
      v-if="!bare"
      :id="`${id}-msg`"
      class="mt-1.5 min-h-5 text-[13px] leading-5"
      :class="error ? 'text-danger' : ok ? 'text-ok' : 'text-ink-3'"
      aria-live="polite"
    >
      {{ message }}
    </p>
  </div>
</template>
