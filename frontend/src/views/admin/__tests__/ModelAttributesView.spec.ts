import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ModelAttributesView from '../ModelAttributesView.vue'
import { modelAttributesAPI } from '@/api/admin/modelAttributes'

vi.mock('@/api/admin/modelAttributes', () => ({ modelAttributesAPI: { list: vi.fn(), defaults: vi.fn(), update: vi.fn(), save: vi.fn(), remove: vi.fn() } }))
vi.mock('@/api/admin', () => ({ adminAPI: { groups: { getAll: vi.fn().mockResolvedValue([{ id: 7, name: 'Group' }]) } } }))
vi.mock('vue-i18n', async () => ({ ...(await vi.importActual<typeof import('vue-i18n')>('vue-i18n')), useI18n: () => ({ t: (key: string) => key }) }))

const stubs = {
  AppLayout: { template: '<div><slot /></div>' },
  TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
  DataTable: { props: ['data'], template: '<div><div v-for="row in data" :key="row.id || row.model"><span>{{ row.name || row.model }}</span><slot name="cell-actions" :row="row" /></div></div>' },
  BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' },
  RuleListEditor: { props: ['items'], template: '<div><div v-for="(item, index) in items" :key="index"><slot name="row" :item="item" :index="index" /></div></div>' },
  Icon: true, Select: true, Toggle: true, Pagination: true, ConfirmDialog: true,
}

beforeEach(() => {
  vi.clearAllMocks()
  vi.mocked(modelAttributesAPI.list).mockResolvedValue({ items: [{ id: 1, name: 'Profile', description: '', status: 'active', group_ids: [7], rules: [{ models: ['upstream'], attributes: { tool_call: false } }] }], total: 1 })
  vi.mocked(modelAttributesAPI.defaults).mockResolvedValue({ items: [{ model: 'default-keep', provider: 'original', source: 'models.dev', attributes: {} }], total: 1, providers: ['original'], version: 'abc123', last_updated: '2026-09-30T00:00:00Z' })
})

describe('属性管理页面', () => {
  it('默认属性的筛选面板支持组合查询、计数、重置和关闭', async () => {
    const wrapper = mount(ModelAttributesView, { attachTo: document.body, global: { stubs } })
    await flushPromises()
    expect(wrapper.find('button[aria-label="common.filter"]').exists()).toBe(false)

    await wrapper.findAll('[role="tab"]')[1]!.trigger('click')
    await flushPromises()
    const filter = wrapper.get('button[aria-label="common.filter"]')
    expect(filter.attributes('aria-expanded')).toBe('false')
    expect(wrapper.findAllComponents({ name: 'Select' })).toHaveLength(0)
    await filter.trigger('click')

    // 同时调整两个条件只查询一次，条件计数与实际请求保持一致。
    vi.mocked(modelAttributesAPI.defaults).mockClear()
    const selects = wrapper.findAllComponents({ name: 'Select' })
    selects[0]!.vm.$emit('update:modelValue', 'original')
    selects[1]!.vm.$emit('update:modelValue', 'tool_call')
    await flushPromises()
    expect(filter.text()).toBe('2')
    expect(modelAttributesAPI.defaults).toHaveBeenCalledTimes(1)
    expect(modelAttributesAPI.defaults).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1, provider: 'original', capability: 'tool_call' }))

    await wrapper.findAll('button').find(button => button.text() === 'common.reset')!.trigger('click')
    await flushPromises()
    expect(filter.text()).toBe('')
    expect(modelAttributesAPI.defaults).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1, provider: '', capability: '' }))

    await wrapper.get('.filter-panel').trigger('keydown', { key: 'Escape' })
    expect(filter.attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(filter.element)
    await filter.trigger('click')
    document.body.click()
    await flushPromises()
    expect(filter.attributes('aria-expanded')).toBe('false')
    wrapper.unmount()
  })

  it('默认属性页只读查询，更新失败保留现有目录', async () => {
    const wrapper = mount(ModelAttributesView, { global: { stubs } })
    await flushPromises()
    await wrapper.findAll('[role="tab"]')[1]!.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('default-keep')
    expect(modelAttributesAPI.update).not.toHaveBeenCalled()
    vi.mocked(modelAttributesAPI.update).mockRejectedValueOnce(new Error('offline'))
    await wrapper.findAll('button').find(button => button.text() === 'admin.pricing.defaults.update')!.trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('default-keep')
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    wrapper.unmount()
  })
  it('编辑回显与保存保留显式 false 及分组关联', async () => {
    const wrapper = mount(ModelAttributesView, { global: { stubs } })
    await flushPromises()
    await wrapper.find('button[aria-label="common.edit"]').trigger('click')
    await flushPromises()
    await wrapper.find('#attribute-form').trigger('submit')
    await flushPromises()
    expect(modelAttributesAPI.save).toHaveBeenCalledWith(expect.objectContaining({ id: 1, group_ids: [7], rules: [{ models: ['upstream'], attributes: { tool_call: false } }] }))
    wrapper.unmount()
  })

  it('模型按回车添加标签，删除后保存仍提交模型数组和分组选择', async () => {
    const wrapper = mount(ModelAttributesView, { global: { stubs } })
    await flushPromises()
    await wrapper.get('button[aria-label="common.edit"]').trigger('click')
    await flushPromises()

    const input = wrapper.get('input[aria-label="admin.modelAttributes.models"]')
    await input.setValue('claude-*')
    await input.trigger('keydown', { key: 'Enter' })
    expect(modelAttributesAPI.save).not.toHaveBeenCalled()
    expect(wrapper.get('button[aria-label="common.delete claude-*"]').exists()).toBe(true)
    await wrapper.get('button[aria-label="common.delete upstream"]').trigger('click')
    await wrapper.get('input[type="checkbox"]').setValue(false)
    await wrapper.get('#attribute-form').trigger('submit')
    await flushPromises()

    expect(modelAttributesAPI.save).toHaveBeenCalledWith(expect.objectContaining({
      group_ids: [],
      rules: [{ models: ['claude-*'], attributes: { tool_call: false } }],
    }))
    wrapper.unmount()
  })

  it('清空模型标签后阻止保存空规则', async () => {
    const wrapper = mount(ModelAttributesView, { global: { stubs } })
    await flushPromises()
    await wrapper.get('button[aria-label="common.edit"]').trigger('click')
    await flushPromises()
    await wrapper.get('button[aria-label="common.delete upstream"]').trigger('click')
    await wrapper.get('#attribute-form').trigger('submit')
    await flushPromises()

    expect(modelAttributesAPI.save).not.toHaveBeenCalled()
    expect(wrapper.get('[role="alert"]').text()).toBe('admin.modelAttributes.modelsRequired')
    wrapper.unmount()
  })
})
