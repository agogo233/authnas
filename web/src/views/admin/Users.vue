<script setup lang="ts">
import { ref, h, onMounted, onUnmounted, watch } from 'vue'
import {
  NCard,
  NDataTable,
  NButton,
  NSpace,
  NTag,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSwitch,
  NEmpty,
  useMessage,
  NPopconfirm,
  NInputGroup,
  NPagination,
} from 'naive-ui'
import type { User } from '@/types'
import { adminApi, type CreateUserRequest, type UpdateUserRequest } from '@/api/admin'

const message = useMessage()

const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const searchQuery = ref('')
const loading = ref(false)

let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null

const showUserModal = ref(false)
const editingUser = ref<User | null>(null)
const userForm = ref<CreateUserRequest>({
  email: '',
  username: '',
  password: '',
  name: '',
  isAdmin: false,
  approved: false,
  mfaRequired: false,
})

const showPasswordModal = ref(false)
const resetPasswordUserId = ref<string>('')
const newPassword = ref('')

const columns = [
  { title: '用户名', key: 'username' },
  { title: '邮箱', key: 'email' },
  { title: '姓名', key: 'name' },
  {
    title: '管理员',
    key: 'isAdmin',
    render: (row: User) =>
      h(
        NTag,
        { type: row.isAdmin ? 'error' : 'default' },
        { default: () => (row.isAdmin ? '是' : '否') }
      ),
  },
  {
    title: '已批准',
    key: 'approved',
    render: (row: User) =>
      h(
        NTag,
        { type: row.approved ? 'success' : 'warning' },
        { default: () => (row.approved ? '是' : '否') }
      ),
  },
  {
    title: 'MFA',
    key: 'mfaRequired',
    render: (row: User) =>
      h(
        NTag,
        { type: row.mfaRequired ? 'info' : 'default' },
        { default: () => (row.mfaRequired ? '已启用' : '已禁用') }
      ),
  },
  {
    title: '创建时间',
    key: 'createdAt',
    render: (row: User) => new Date(row.createdAt).toLocaleDateString('zh-CN'),
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    render: (row: User) =>
      h(
        NSpace,
        { size: 'small' },
        {
          default: () => [
            h(
              NButton,
              { size: 'small', onClick: () => openEditModal(row) },
              { default: () => '编辑' }
            ),
            !row.approved &&
              h(
                NButton,
                { size: 'small', type: 'success', onClick: () => handleApprove(row.id) },
                { default: () => '批准' }
              ),
            h(
              NButton,
              { size: 'small', type: 'warning', onClick: () => openPasswordModal(row.id) },
              { default: () => '重置密码' }
            ),
            h(
              NPopconfirm,
              { onConfirm: () => handleDelete(row.id) },
              {
                trigger: () =>
                  h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
                default: () => '确定要删除此用户吗？',
              }
            ),
          ],
        }
      ),
  },
]

async function fetchUsers() {
  loading.value = true
  try {
    const res = await adminApi.users.list({
      page: page.value,
      pageSize: pageSize.value,
      search: searchQuery.value,
    })
    if (res.data.success && res.data.data) {
      users.value = res.data.data
      total.value = res.data.total || res.data.data.length
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  editingUser.value = null
  userForm.value = {
    email: '',
    username: '',
    password: '',
    name: '',
    isAdmin: false,
    approved: false,
    mfaRequired: false,
  }
  showUserModal.value = true
}

function openEditModal(user: User) {
  editingUser.value = user
  userForm.value = {
    email: user.email || '',
    username: user.username,
    name: user.name || '',
    isAdmin: user.isAdmin || false,
    approved: user.approved,
    mfaRequired: user.mfaRequired || false,
  }
  showUserModal.value = true
}

async function handleSaveUser() {
  try {
    if (editingUser.value) {
      await adminApi.users.update(editingUser.value.id, userForm.value as UpdateUserRequest)
      message.success('用户更新成功')
    } else {
      await adminApi.users.create(userForm.value)
      message.success('用户创建成功')
    }
    showUserModal.value = false
    fetchUsers()
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存用户失败')
  }
}

async function handleDelete(id: string) {
  try {
    await adminApi.users.delete(id)
    message.success('用户删除成功')
    fetchUsers()
  } catch (err: any) {
    message.error(err.response?.data?.message || '删除用户失败')
  }
}

async function handleApprove(id: string) {
  try {
    await adminApi.users.approve(id, { approved: true })
    message.success('用户批准成功')
    fetchUsers()
  } catch (err: any) {
    message.error(err.response?.data?.message || '批准用户失败')
  }
}

function openPasswordModal(userId: string) {
  resetPasswordUserId.value = userId
  newPassword.value = ''
  showPasswordModal.value = true
}

async function handleResetPassword() {
  try {
    await adminApi.users.resetPassword(resetPasswordUserId.value, {
      newPassword: newPassword.value,
    })
    message.success('密码重置成功')
    showPasswordModal.value = false
  } catch (err: any) {
    message.error(err.response?.data?.message || '重置密码失败')
  }
}

function handleSearch() {
  page.value = 1
  fetchUsers()
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchUsers()
}

function handlePageSizeChange(newSize: number) {
  pageSize.value = newSize
  page.value = 1
  fetchUsers()
}

watch(searchQuery, () => {
  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer)
  }
  searchDebounceTimer = setTimeout(() => {
    page.value = 1
    fetchUsers()
  }, 300)
})

onMounted(() => {
  fetchUsers()
})

onUnmounted(() => {
  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer)
    searchDebounceTimer = null
  }
})
</script>

<template>
  <div class="page-container">
    <NCard title="用户管理">
      <template #header-extra>
        <NButton type="primary" @click="openCreateModal">创建用户</NButton>
      </template>

      <NSpace vertical :size="16">
        <NInputGroup>
          <NInput
            v-model:value="searchQuery"
            placeholder="搜索用户..."
            @keyup.enter="handleSearch"
          />
          <NButton type="primary" @click="handleSearch">搜索</NButton>
        </NInputGroup>

        <NDataTable
          v-if="users.length > 0"
          :columns="columns"
          :data="users"
          :loading="loading"
          :bordered="false"
          :pagination="false"
        />
        <NEmpty v-else description="暂无用户">
          <template #extra>
            <NButton size="small" type="primary" @click="openCreateModal">创建用户</NButton>
          </template>
        </NEmpty>

        <NSpace justify="end">
          <NPagination
            v-model:page="page"
            :page-size="pageSize"
            :page-sizes="[10, 20, 50]"
            :item-count="total"
            show-size-picker
            @update:page="handlePageChange"
            @update:page-size="handlePageSizeChange"
          />
        </NSpace>
      </NSpace>
    </NCard>

    <NModal
      v-model:show="showUserModal"
      preset="card"
      :title="editingUser ? '编辑用户' : '创建用户'"
      style="width: 500px"
    >
      <NForm :model="userForm" label-placement="top">
        <NFormItem label="邮箱" required>
          <NInput v-model:value="userForm.email" :disabled="!!editingUser" />
        </NFormItem>
        <NFormItem label="用户名" required>
          <NInput v-model:value="userForm.username" :disabled="!!editingUser" />
        </NFormItem>
        <NFormItem v-if="!editingUser" label="密码" required>
          <NInput v-model:value="userForm.password" type="password" />
        </NFormItem>
        <NFormItem label="姓名">
          <NInput v-model:value="userForm.name" />
        </NFormItem>
        <NFormItem label="管理员">
          <NSwitch v-model:value="userForm.isAdmin" />
        </NFormItem>
        <NFormItem label="已批准">
          <NSwitch v-model:value="userForm.approved" />
        </NFormItem>
        <NFormItem label="需要 MFA">
          <NSwitch v-model:value="userForm.mfaRequired" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showUserModal = false">取消</NButton>
          <NButton type="primary" @click="handleSaveUser">保存</NButton>
        </NSpace>
      </template>
    </NModal>

    <NModal v-model:show="showPasswordModal" preset="card" title="重置密码" style="width: 400px">
      <NForm :model="{ newPassword }" label-placement="top">
        <NFormItem label="新密码" required>
          <NInput v-model:value="newPassword" type="password" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showPasswordModal = false">取消</NButton>
          <NButton type="primary" @click="handleResetPassword">重置</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.page-container {
  padding: 40px 0;
}
</style>
