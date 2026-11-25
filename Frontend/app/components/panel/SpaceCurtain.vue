<script setup lang="ts">
import type { Space } from '~/composables/useSpace'

// Panel değişirken hedef panelin matı ekranı kaplar; kullanıcı yer değiştirdiğini görür.
const { switchingTo } = useSpace()

const shown = ref<Space>('user')
watch(switchingTo, (v) => {
  if (v) shown.value = v
})

const mats: Record<Space, { bg: string; label: string }> = {
  user: { bg: '#0c5c47', label: 'Kullanıcı paneli' },
  admin: { bg: '#173f80', label: 'Yönetim paneli' }
}
</script>

<template>
  <Transition name="space">
    <div
      v-if="switchingTo"
      class="mat-grid fixed inset-0 z-[70] flex items-center justify-center"
      :style="{ backgroundColor: mats[shown].bg }"
      role="status"
    >
      <p class="display text-4xl text-white md:text-6xl">{{ mats[shown].label }}</p>
    </div>
  </Transition>
</template>
