<script setup lang="ts">
/** Üç nokta menüsü. İçerik body'ye taşınır, tablo taşmasından etkilenmez. */
export interface MenuItem {
  label: string
  icon?: string
  tone?: 'normal' | 'danger'
  to?: string
  href?: string
  disabled?: boolean
  divider?: boolean
  onSelect?: () => void
}

defineProps<{ label: string; items: MenuItem[]; icon?: string; tone?: 'normal' | 'mat' }>()

const open = ref(false)
const trigger = ref<HTMLElement | null>(null)
const list = ref<HTMLElement | null>(null)
const pos = ref<{ top: string; right: string; bottom: string }>({ top: 'auto', right: '0', bottom: 'auto' })

function place() {
  const r = trigger.value?.getBoundingClientRect()
  if (!r) return
  const below = window.innerHeight - r.bottom > 260
  pos.value = {
    top: below ? `${r.bottom + 4}px` : 'auto',
    bottom: below ? 'auto' : `${window.innerHeight - r.top + 4}px`,
    right: `${Math.max(8, document.documentElement.clientWidth - r.right)}px`
  }
}

async function toggle() {
  open.value = !open.value
  if (!open.value) return
  place()
  await nextTick()
  list.value?.querySelector<HTMLElement>('[role="menuitem"]')?.focus()
}

function close(refocus = false) {
  if (!open.value) return
  open.value = false
  if (refocus) trigger.value?.querySelector('button')?.focus()
}

function select(item: MenuItem) {
  close()
  item.onSelect?.()
}

function step(dir: number) {
  const items = [...(list.value?.querySelectorAll<HTMLElement>('[role="menuitem"]:not([disabled])') ?? [])]
  const i = items.indexOf(document.activeElement as HTMLElement)
  items[(i + dir + items.length) % items.length]?.focus()
}

function onDown(e: MouseEvent) {
  const t = e.target as Node
  if (trigger.value?.contains(t) || list.value?.contains(t)) return
  close()
}

function onScroll() {
  close()
}

watch(open, (v) => {
  if (!import.meta.client) return
  const fn = v ? 'addEventListener' : 'removeEventListener'
  document[fn]('mousedown', onDown as EventListener)
  window[fn]('scroll', onScroll, true)
  window[fn]('resize', onScroll)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDown)
  window.removeEventListener('scroll', onScroll, true)
  window.removeEventListener('resize', onScroll)
})
</script>

<template>
  <span ref="trigger" class="inline-flex" @click.stop>
    <slot name="trigger" :open="open" :toggle="toggle">
      <button
        type="button"
        class="inline-flex h-10 w-10 items-center justify-center rounded-md transition-colors"
        :class="
          tone === 'mat'
            ? 'text-on-mat hover:bg-white/15'
            : open
              ? 'bg-sunken text-ink'
              : 'text-ink-2 hover:bg-sunken hover:text-ink'
        "
        :aria-label="label"
        :title="label"
        aria-haspopup="menu"
        :aria-expanded="open"
        @click="toggle"
      >
        <UiIcon :name="icon ?? 'more'" />
      </button>
    </slot>
  </span>

  <Teleport to="body">
    <Transition name="pop">
      <div
        v-if="open"
        ref="list"
        role="menu"
        :aria-label="label"
        class="fixed z-40 min-w-56 rounded-md border border-line bg-surface py-1.5 shadow-pop"
        :style="pos"
        @keydown.esc.stop="close(true)"
        @keydown.down.prevent="step(1)"
        @keydown.up.prevent="step(-1)"
        @keydown.tab="close()"
      >
        <template v-for="item in items" :key="item.label">
          <hr v-if="item.divider" class="my-1.5 border-line" />
          <NuxtLink
            v-if="item.to"
            :to="item.to"
            role="menuitem"
            class="flex min-h-10 items-center gap-3 px-4 text-sm text-ink hover:bg-sunken focus:bg-sunken focus:outline-none"
            @click="close()"
          >
            <UiIcon v-if="item.icon" :name="item.icon" :size="17" class="text-ink-3" /> {{ item.label }}
          </NuxtLink>
          <a
            v-else-if="item.href"
            :href="item.href"
            target="_blank"
            rel="noopener"
            role="menuitem"
            class="flex min-h-10 items-center gap-3 px-4 text-sm text-ink hover:bg-sunken focus:bg-sunken focus:outline-none"
            @click="close()"
          >
            <UiIcon v-if="item.icon" :name="item.icon" :size="17" class="text-ink-3" /> {{ item.label }}
          </a>
          <button
            v-else
            type="button"
            role="menuitem"
            :disabled="item.disabled"
            class="flex min-h-10 w-full items-center gap-3 px-4 text-left text-sm hover:bg-sunken focus:bg-sunken focus:outline-none disabled:opacity-45"
            :class="item.tone === 'danger' ? 'text-danger hover:bg-danger-soft focus:bg-danger-soft' : 'text-ink'"
            @click="select(item)"
          >
            <UiIcon v-if="item.icon" :name="item.icon" :size="17" :class="item.tone === 'danger' ? '' : 'text-ink-3'" />
            {{ item.label }}
          </button>
        </template>
      </div>
    </Transition>
  </Teleport>
</template>
