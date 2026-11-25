<script setup lang="ts">
import { NuxtLink } from '#components'

/**
 * Her görünümün tek bir anlamı var:
 *  primary : sayfanın ana eylemi (kesik köşe)
 *  mark    : mat zemin üstündeki ana eylem (sarı)
 *  outline : yardımcı eylem
 *  quiet   : geri dönüş, vazgeçme
 *  danger  : geri alınamayan eyleme götüren düğme
 *  destroy : onay penceresindeki geri alınamayan son adım
 *  mat     : mat zemin üstündeki yardımcı eylem
 */
const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'mark' | 'outline' | 'quiet' | 'danger' | 'destroy' | 'mat'
    size?: 'sm' | 'md' | 'lg'
    type?: 'button' | 'submit'
    to?: string
    href?: string
    icon?: string
    iconEnd?: string
    loading?: boolean
    loadingLabel?: string
    disabled?: boolean
    block?: boolean
  }>(),
  { variant: 'primary', size: 'md', type: 'button' }
)

const isOff = computed(() => props.disabled || props.loading)

const sizes = {
  sm: 'min-h-9 px-3.5 text-sm',
  md: 'min-h-11 px-5 text-[15px]',
  lg: 'min-h-14 px-7 text-base'
}

const variants = {
  primary: 'cut-corner font-semibold text-on-brand hover:[--cut-bg:var(--color-brand-hover)]',
  mark: 'cut-corner font-semibold text-on-mark [--cut-bg:var(--color-mark)] hover:[--cut-bg:var(--color-mark-hover)]',
  outline: 'rounded-md border border-line-2 font-medium text-ink hover:border-ink hover:bg-sunken',
  quiet: 'rounded-md font-medium text-ink-2 hover:bg-sunken hover:text-ink',
  danger: 'rounded-md font-medium text-danger hover:bg-danger-soft',
  destroy: 'rounded-md bg-danger font-semibold text-on-danger hover:bg-danger-hover',
  mat: 'rounded-md border border-white/45 font-medium text-on-mat hover:border-white hover:bg-white/10'
}

const tag = computed(() => (props.to ? NuxtLink : props.href ? 'a' : 'button'))
const attrs = computed(() => {
  if (props.to) return { to: props.to }
  if (props.href) return { href: props.href }
  return { type: props.type, disabled: isOff.value, 'aria-busy': props.loading || undefined }
})

const iconSize = computed(() => (props.size === 'sm' ? 16 : 18))
</script>

<template>
  <component
    :is="tag"
    v-bind="attrs"
    class="inline-flex shrink-0 select-none items-center justify-center whitespace-nowrap transition-colors duration-100 disabled:opacity-55"
    :class="[sizes[size], variants[variant], block ? 'w-full' : '']"
  >
    <!-- Yükleme sırasında içerik yerini korur; düğmenin genişliği değişmez. -->
    <span class="grid items-center justify-items-center">
      <span class="col-start-1 row-start-1 inline-flex items-center gap-2" :class="loading ? 'invisible' : ''">
        <UiIcon v-if="icon" :name="icon" :size="iconSize" />
        <slot />
        <UiIcon v-if="iconEnd" :name="iconEnd" :size="iconSize" />
      </span>
      <span v-if="loadingLabel || loading" class="col-start-1 row-start-1 inline-flex items-center gap-2" :class="loading ? '' : 'invisible'" :aria-hidden="!loading">
        <UiSpinner :size="iconSize - 2" />
        <span v-if="loadingLabel">{{ loadingLabel }}</span>
      </span>
    </span>
  </component>
</template>
