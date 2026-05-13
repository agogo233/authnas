import { config } from '@vue/test-utils'
import { vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

setActivePinia(createPinia())

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    go: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
  }),
  useRoute: () => ({
    path: '/',
    name: 'test',
    query: {},
    params: {},
    fullPath: '/',
    matched: [],
  }),
  RouterLink: {
    name: 'RouterLink',
    props: ['to'],
    template: '<a :href="to"><slot /></a>',
  },
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'style'],
    template: '<div class="n-card"><slot /></div>',
  },
  NForm: {
    name: 'NForm',
    template: '<form @submit.prevent="$emit(\'submit\')"><slot /></form>',
  },
  NFormItem: {
    name: 'NFormItem',
    props: ['label'],
    template: '<div class="n-form-item"><label v-if="label">{{ label }}</label><slot /></div>',
  },
  NInput: {
    name: 'NInput',
    props: ['modelValue', 'type', 'placeholder', 'size', 'maxlength', 'disabled'],
    emits: ['update:modelValue'],
    template: '<input :type="type || \'text\'" :value="modelValue" :placeholder="placeholder" :maxlength="maxlength" :disabled="disabled" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  NButton: {
    name: 'NButton',
    props: ['type', 'loading', 'disabled', 'attrType', 'block'],
    emits: ['click'],
    template: '<button :type="attrType" :disabled="disabled || loading" @click="$emit(\'click\')"><slot /></button>',
  },
  NSpace: {
    name: 'NSpace',
    props: ['justify', 'align'],
    template: '<div class="n-space"><slot /></div>',
  },
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  NCheckbox: {
    name: 'NCheckbox',
    props: ['checked'],
    emits: ['update:checked'],
    template: '<input type="checkbox" :checked="checked" @change="$emit(\'update:checked\', $event.target.checked)" />',
  },
  NProgress: {
    name: 'NProgress',
    props: ['percentage', 'color', 'showIndicator', 'height'],
    template: '<div class="n-progress" />',
  },
  NResult: {
    name: 'NResult',
    props: ['status', 'title', 'description'],
    template: '<div class="n-result"><slot name="footer" /></div>',
  },
  NSpin: {
    name: 'NSpin',
    template: '<div class="n-spin" />',
  },
  NModal: {
    name: 'NModal',
    props: ['show', 'preset', 'title'],
    emits: ['update:show'],
    template: '<div class="n-modal" v-if="show"><slot /></div>',
  },
  NImage: {
    name: 'NImage',
    props: ['src', 'width', 'height'],
    template: '<img :src="src" :width="width" :height="height" />',
  },
  NPopconfirm: {
    name: 'NPopconfirm',
    props: [],
    emits: ['positive-click'],
    template: '<div><slot name="trigger" @click="$emit(\'positive-click\')" /></div>',
  },
  NTabs: {
    name: 'NTabs',
    props: ['type', 'animated'],
    template: '<div class="n-tabs"><slot /></div>',
  },
  NTabPane: {
    name: 'NTabPane',
    props: ['name', 'tab'],
    template: '<div class="n-tab-pane"><slot /></div>',
  },
  NDataTable: {
    name: 'NDataTable',
    props: ['columns', 'data'],
    template: '<table class="n-data-table"><slot /></table>',
  },
  NTag: {
    name: 'NTag',
    props: ['type'],
    template: '<span class="n-tag"><slot /></span>',
  },
  NDescriptions: {
    name: 'NDescriptions',
    props: ['column', 'bordered'],
    template: '<div class="n-descriptions"><slot /></div>',
  },
  NDescriptionsItem: {
    name: 'NDescriptionsItem',
    props: ['label'],
    template: '<div class="n-descriptions-item"><slot /></div>',
  },
  useMessage: () => ({
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  }),
}))

config.global.mocks = {
  $route: {
    path: '/',
    name: 'test',
    query: {},
    params: {},
    fullPath: '/',
  },
}
