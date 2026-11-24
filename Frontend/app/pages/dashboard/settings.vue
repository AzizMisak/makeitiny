<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })
useHead({ title: 'Hesap Ayarları — makeiTiny' })

const { user, primaryRoleLabel, fetchCurrentUser } = useAuth()

onMounted(() => {
  fetchCurrentUser().catch(() => {})
})
</script>

<template>
  <DashboardPageHeader
    title="Hesap ayarları"
    description="Hesap bilgilerin ve güvenlik tercihlerin."
  />

  <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
    <DashboardStatTile
      label="Rol"
      :value="primaryRoleLabel"
    />
    <DashboardStatTile
      label="Hesap durumu"
      :value="user?.isActive === false ? 'Askıda' : 'Aktif'"
    />
    <DashboardStatTile
      label="Oluşturulan link"
      :value="(user?.linkCount ?? 0).toLocaleString('tr-TR')"
    />
    <DashboardStatTile
      label="Kayıt tarihi"
      :value="formatDateTr(user?.createdAt)"
    />
  </div>

  <DashboardSettingsForm />
</template>
