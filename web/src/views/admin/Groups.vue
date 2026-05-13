<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import {
  NCard,
  NDataTable,
  NButton,
  NSpace,
  NModal,
  NForm,
  NFormItem,
  NInput,
  useMessage,
  NPopconfirm,
  NDrawer,
  NDrawerContent,
  NAlert,
} from 'naive-ui'
import type { Group, User } from '@/types'
import { adminApi, type CreateGroupRequest, type UpdateGroupRequest } from '@/api/admin'

const message = useMessage()

const groups = ref<Group[]>([])
const loading = ref(false)

const showGroupModal = ref(false)
const editingGroup = ref<Group | null>(null)
const groupForm = ref<CreateGroupRequest>({
  name: '',
  description: '',
})

const showMembersDrawer = ref(false)
const selectedGroup = ref<Group | null>(null)
const groupMembers = ref<User[]>([])

const columns = [
  { title: '名称', key: 'name' },
  { title: '描述', key: 'description', ellipsis: true },
  {
    title: '创建时间',
    key: 'createdAt',
    render: (row: Group) => new Date(row.createdAt).toLocaleDateString('zh-CN'),
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    render: (row: Group) =>
      h(
        NSpace,
        { size: 'small' },
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                disabled: true,
                onClick: () => message.warning('成员管理功能正在开发中'),
              },
              { default: () => '成员' }
            ),
            h(
              NButton,
              { size: 'small', onClick: () => openEditModal(row) },
              { default: () => '编辑' }
            ),
            h(
              NPopconfirm,
              { onConfirm: () => handleDelete(row.id) },
              {
                trigger: () =>
                  h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
                default: () => '确定要删除此用户组吗？',
              }
            ),
          ],
        }
      ),
  },
]

async function fetchGroups() {
  loading.value = true
  try {
    const res = await adminApi.groups.list()
    if (res.data.success && res.data.data) {
      groups.value = res.data.data
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '获取用户组列表失败')
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  editingGroup.value = null
  groupForm.value = { name: '', description: '' }
  showGroupModal.value = true
}

function openEditModal(group: Group) {
  editingGroup.value = group
  groupForm.value = { name: group.name, description: group.description || '' }
  showGroupModal.value = true
}

async function handleSaveGroup() {
  try {
    if (editingGroup.value) {
      await adminApi.groups.update(editingGroup.value.id, groupForm.value as UpdateGroupRequest)
      message.success('用户组更新成功')
    } else {
      await adminApi.groups.create(groupForm.value)
      message.success('用户组创建成功')
    }
    showGroupModal.value = false
    fetchGroups()
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存用户组失败')
  }
}

async function handleDelete(id: string) {
  try {
    await adminApi.groups.delete(id)
    message.success('用户组删除成功')
    fetchGroups()
  } catch (err: any) {
    message.error(err.response?.data?.message || '删除用户组失败')
  }
}

async function openMembersDrawer(group: Group) {
  selectedGroup.value = group
  groupMembers.value = []
  showMembersDrawer.value = true
}

onMounted(() => {
  fetchGroups()
})
</script>

<template>
  <div class="page-container">
    <NCard title="用户组管理">
      <template #header-extra>
        <NButton type="primary" @click="openCreateModal">创建用户组</NButton>
      </template>

      <NDataTable :columns="columns" :data="groups" :loading="loading" :bordered="false" />
    </NCard>

    <NModal
      v-model:show="showGroupModal"
      preset="card"
      :title="editingGroup ? '编辑用户组' : '创建用户组'"
      style="width: 500px"
    >
      <NForm :model="groupForm" label-placement="top">
        <NFormItem label="名称" required>
          <NInput v-model:value="groupForm.name" />
        </NFormItem>
        <NFormItem label="描述">
          <NInput v-model:value="groupForm.description" type="textarea" :rows="3" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showGroupModal = false">取消</NButton>
          <NButton type="primary" @click="handleSaveGroup">保存</NButton>
        </NSpace>
      </template>
    </NModal>

    <NDrawer v-model:show="showMembersDrawer" :width="400" placement="right">
      <NDrawerContent :title="`用户组: ${selectedGroup?.name || ''}`">
        <template #header>
          <span>用户组成员</span>
        </template>
        <NAlert type="warning"> 成员管理功能正在开发中 </NAlert>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
.page-container {
  padding: 40px 0;
}
</style>
