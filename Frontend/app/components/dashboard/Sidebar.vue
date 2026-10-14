<script setup lang="ts">
const { user } = useAuth()
</script>

<template>
  <aside class="flex h-full w-64 flex-col border-r border-line bg-paper-2/60 p-4">
    <div class="px-2 py-2">
      <BrandLogo />
    </div>

    <nav class="mt-8 flex flex-1 flex-col gap-7" aria-label="Panel menüsü">
      <div v-for="group in dashboardNav" :key="group.title">
        <p class="eyebrow mb-2 px-3 !text-ink-3">{{ group.title }}</p>
        <ul class="flex flex-col gap-1.5">
          <li v-for="item in group.items" :key="item.to">
            <DashboardSidebarLink :item="item" />
          </li>
        </ul>
      </div>
    </nav>

    <div class="border-t border-line pt-4">
      <div class="mb-3 px-3">
        <p class="truncate text-sm font-semibold">{{ user?.name }}</p>
        <p class="truncate text-xs text-ink-3">{{ user?.email }}</p>
      </div>
      <NuxtLink
        to="/auth/logout"
        class="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-ink-2 transition-colors hover:bg-accent-soft hover:text-accent-dark"
      >
        <UiIcon name="logout" /> Çıkış yap
      </NuxtLink>
    </div>
  </aside>
</template>
