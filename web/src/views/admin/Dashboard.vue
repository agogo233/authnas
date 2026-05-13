<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NGrid, NGi, NStatistic, NSpin, useMessage } from 'naive-ui'
import { adminApi } from '@/api/admin'

const message = useMessage()
const loading = ref(true)
const stats = ref({
  users: 0,
  groups: 0,
  clients: 0,
  invitations: 0,
  activeInvitations: 0,
})

async function fetchStats() {
  loading.value = true
  try {
    const [usersRes, groupsRes, clientsRes, invitationsRes] = await Promise.all([
      adminApi.users.count(),
      adminApi.groups.list(),
      adminApi.clients.list(),
      adminApi.invitations.list(),
    ])

    stats.value.users = usersRes.data.data?.total || 0
    stats.value.groups = groupsRes.data.data?.length || 0
    stats.value.clients = clientsRes.data.data?.length || 0
    
    if (invitationsRes.data.data) {
      stats.value.invitations = invitationsRes.data.data.length
      stats.value.activeInvitations = invitationsRes.data.data.filter(
        (inv: any) => new Date(inv.expiresAt) > new Date()
      ).length
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '获取仪表盘数据失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStats()
})
</script>

<template>
  <div class="page-container">
    <h1>管理后台</h1>
    <NSpin :show="loading">
      <NGrid :cols="4" :x-gap="20" :y-gap="20" responsive="screen" item-responsive>
        <NGi>
          <NCard hoverable>
            <NStatistic label="用户总数" :value="stats.users" />
          </NCard>
        </NGi>
        <NGi>
          <NCard hoverable>
            <NStatistic label="用户组" :value="stats.groups" />
          </NCard>
        </NGi>
        <NGi>
          <NCard hoverable>
            <NStatistic label="OIDC 客户端" :value="stats.clients" />
          </NCard>
        </NGi>
        <NGi>
          <NCard hoverable>
            <NStatistic label="有效邀请" :value="stats.activeInvitations" />
          </NCard>
        </NGi>
      </NGrid>

      <NGrid :cols="2" :x-gap="20" :y-gap="20" style="margin-top: 24px" responsive="screen" item-responsive>
        <NGi>
          <NCard title="快捷操作" hoverable>
            <div class="quick-actions">
              <router-link to="/admin/users">
                <NStatistic label="管理用户" />
              </router-link>
            </div>
          </NCard>
        </NGi>
        <NGi>
          <NCard title="系统状态" hoverable>
            <NStatistic label="已创建邀请总数" :value="stats.invitations" />
          </NCard>
        </NGi>
      </NGrid>
    </NSpin>
  </div>
</template>

<style scoped>
.page-container {
  padding: 40px 0;
}

h1 {
  margin-bottom: 24px;
  font-size: 28px;
  font-weight: 600;
  color: var(--text-h);
}

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

a {
  text-decoration: none;
  color: inherit;
}
</style>
