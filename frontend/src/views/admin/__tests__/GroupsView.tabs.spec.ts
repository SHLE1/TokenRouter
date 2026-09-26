import { useProtocolCatalogFixture } from '@/__tests__/helpers/protocolCatalog'
import { defineComponent } from 'vue'
import { createPinia } from 'pinia'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import GroupsView from '../GroupsView.vue'
import Select from '@/components/common/Select.vue'
import GroupClientProtocolSelector from '@/components/admin/group/GroupClientProtocolSelector.vue'
import { defaultRoutingPolicy } from '@/components/admin/group/routingPolicy'
import type { AdminGroup } from '@/types'

const { groups, showError } = vi.hoisted(() => ({
  groups: {
    list: vi.fn(), getAll: vi.fn(), getModelsListCandidates: vi.fn(),
    getUsageSummary: vi.fn(), getCapacitySummary: vi.fn(), getLiveCapability: vi.fn(),
    create: vi.fn(), update: vi.fn(),
  },
  showError: vi.fn(),
}))
vi.mock('@/api/admin', () => ({ adminAPI: { groups, accounts: { list: vi.fn(), getById: vi.fn() } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess: vi.fn() }) }))
vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({ isCurrentStep: vi.fn(() => false), nextStep: vi.fn() }),
}))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key }),
}))

const Dialog = defineComponent({
  props: ['show'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})
const Page = defineComponent({ template: '<div><slot name="filters" /><slot name="table" /></div>' })
const Layout = defineComponent({ template: '<div><slot /></div>' })
const Table = defineComponent({ props: ['data'], template: '<div><div v-for="row in data" :key="row.id"><slot name="cell-actions" :row="row" /></div></div>' })
const wrappers: VueWrapper[] = []

function group(): AdminGroup {
  // 只提供界面依赖的存量字段，其他配置由编辑初始化逻辑使用默认值。
  return {
    id: 42, name: 'Existing', rate_multiplier: 1, status: 'active',
    scheduler_type: 'basic', is_exclusive: false, model_routing: null,
    supported_model_scopes: ['claude', 'gemini_text', 'gemini_image'],
  } as AdminGroup
}

async function open(mode: 'create' | 'edit', _accountType: string, overrides: Partial<AdminGroup> = {}) {
  groups.list.mockResolvedValue({ items: [{ ...group(), ...overrides }], total: 1, pages: 1 })
  const wrapper = mount(GroupsView, {
    attachTo: document.body,
    global: { plugins: [createPinia()], stubs: {
      AppLayout: Layout, TablePageLayout: Page, BaseDialog: Dialog, DataTable: Table,
      Select: true, Icon: true, PlatformIcon: true, ProviderIcon: true,
      Pagination: true, ConfirmDialog: true, EmptyState: true, GroupCapacityBadge: true,
      GroupRateMultipliersModal: true, GroupRPMOverridesModal: true,
      GroupAdvancedSchedulerOverridesModal: true, VueDraggable: true,
    } },
  })
  wrappers.push(wrapper)
  await flushPromises()
  const button = mode === 'create'
    ? wrapper.get('[data-tour="groups-create-btn"]')
    : wrapper.findAll('button').find(button => button.text() === 'common.edit')!
  await button.trigger('click')
  await flushPromises()
  if (mode === 'create') {
    await wrapper.get('[data-group-field="name"] input').setValue('New group')
  }
  return wrapper
}

async function tab(wrapper: VueWrapper, name: string) {
  await wrapper.get(`[data-group-tab-button="${name}"]`).trigger('click')
  await flushPromises()
}

beforeEach(() => {
  vi.clearAllMocks()
  localStorage.clear()
  groups.getAll.mockResolvedValue([])
  groups.getModelsListCandidates.mockResolvedValue(['gpt-test'])
  groups.getUsageSummary.mockResolvedValue([])
  groups.getCapacitySummary.mockResolvedValue([])
  groups.getLiveCapability.mockResolvedValue({ supported: true })
  groups.create.mockResolvedValue({ id: 43 })
  groups.update.mockResolvedValue({ id: 42 })
})
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  document.body.innerHTML = ''
})

it('编辑历史停用策略后保存即应用，打开表单时不提前更新服务端', async () => {
  const policy = {
    ...defaultRoutingPolicy(), enabled: false,
    model_mapping: { alias: 'gpt-test' }, features: '历史文字',
  }
  const wrapper = await open('edit', 'openai', { routing_policy: policy })
  await tab(wrapper, 'routing')
  expect(wrapper.get('input[aria-label="admin.groups.routingPolicy.source"]').isVisible()).toBe(true)
  expect(wrapper.text()).not.toContain('admin.groups.routingPolicy.enabled')
  expect(groups.update).not.toHaveBeenCalled()
  expect(policy.enabled).toBe(false)
  await wrapper.get('#edit-group-form').trigger('submit')
  await flushPromises()
  expect(groups.update.mock.calls[0]?.[1].routing_policy).toMatchObject({
    enabled: true, model_mapping: policy.model_mapping, features: '历史文字',
  })
})

describe.each(['create', 'edit'] as const)('GroupsView %s tabs', mode => {
  it('分组只保留基本、功能、模型和协议页签', async () => {
    const wrapper = await open(mode, 'mixed')
    const keys = wrapper.findAll('[data-group-tab-button]').map(button => button.attributes('data-group-tab-button'))
    expect(wrapper.find('[data-group-tab-button="pricing"]').exists()).toBe(false)
    expect(keys).toEqual(['general', 'features', 'routing', 'protocol'])
    expect(wrapper.get('[data-group-tab="general"]').isVisible()).toBe(true)
    expect(wrapper.get('[data-tour="group-form-multiplier"]').element.closest('[data-group-tab]')?.getAttribute('data-group-tab')).toBe('general')
    expect(wrapper.getComponent(GroupClientProtocolSelector).element.closest('[data-group-tab]')?.getAttribute('data-group-tab')).toBe('protocol')
    expect(wrapper.find('[data-group-field="reasoning"]').exists()).toBe(true)
    expect(wrapper.find('[data-group-field="image-capabilities"]').exists()).toBe(false)
  })

  it('跨页草稿一次提交，基础倍率与 Fast 路由策略保存，重新打开回到通用', async () => {
    const wrapper = await open(mode, 'openai')
    await tab(wrapper, 'routing')
    await wrapper.get('[data-tour="group-form-multiplier"]').setValue('1.5')
    await tab(wrapper, 'features')
    const force = wrapper.get(`[data-testid="${mode}-openai-fast"]`).getComponent(Select)
    expect(force.props('modelValue')).toBe('follow_request')
    expect(force.props('options').map((option: { value: string }) => option.value)).toEqual(['follow_request', 'force_priority', 'force_ultrafast', 'force_off'])
    force.vm.$emit('update:modelValue', 'force_ultrafast')
    await flushPromises()
    await tab(wrapper, 'protocol')
    wrapper.getComponent(GroupClientProtocolSelector).vm.$emit('update:modelValue', ['anthropic_messages'])
    await flushPromises()
    expect(wrapper.text()).not.toContain('admin.groups.openaiMessages.exactMappingTitle')
    await tab(wrapper, 'routing')
    await wrapper.get('[data-group-tab="routing"]').findAll('button').find(button => button.text() === 'common.add')!.trigger('click')
    await wrapper.get('input[aria-label="admin.groups.routingPolicy.source"]').setValue('claude-sonnet-4-6')
    await wrapper.get('input[aria-label="admin.groups.routingPolicy.target"]').setValue('gpt-test')
    await tab(wrapper, 'routing')
    expect((wrapper.get('[data-tour="group-form-multiplier"]').element as HTMLInputElement).value).toBe('1.5')
    await wrapper.get(`#${mode}-group-form`).trigger('submit')
    await flushPromises()
    const payload = mode === 'create' ? groups.create.mock.calls[0]?.[0] : groups.update.mock.calls[0]?.[1]
    expect(payload.routing_policy.model_mapping).toEqual({ 'claude-sonnet-4-6': 'gpt-test' })
    expect(payload.messages_dispatch_model_config).toBeUndefined()
    expect(wrapper.find(`#${mode}-group-form`).exists()).toBe(false)
    await wrapper.get('[data-tour="groups-create-btn"]').trigger('click')
    expect(wrapper.get('[data-group-tab="general"]').isVisible()).toBe(true)
  })

  it('隐藏页签的名称、倍率和推理错误均可定位且阻止提交', async () => {
    const wrapper = await open(mode, 'openai')
    await wrapper.get('[data-group-field="name"] input').setValue('   ')
    await tab(wrapper, 'routing')
    await wrapper.get(`#${mode}-group-form`).trigger('submit')
    await flushPromises()
    expect(wrapper.get('[data-group-tab="general"]').isVisible()).toBe(true)
    await wrapper.get('[data-group-field="name"] input').setValue('Valid')
    await wrapper.get('[data-tour="group-form-multiplier"]').setValue('-1')
    await wrapper.get(`#${mode}-group-form`).trigger('submit')
    await flushPromises()
    expect(wrapper.get('[data-group-tab="general"]').isVisible()).toBe(true)
    await wrapper.get('[data-tour="group-form-multiplier"]').setValue('1')
    await tab(wrapper, 'features')
    await wrapper.get('[data-group-field="reasoning"] button').trigger('click')
    await tab(wrapper, 'general')
    await wrapper.get(`#${mode}-group-form`).trigger('submit')
    await flushPromises()
    expect(wrapper.get('[data-group-tab="features"]').isVisible()).toBe(true)
    expect(wrapper.find('[data-group-field="reasoning"] [role="alert"]').exists()).toBe(true)
    expect(groups[mode === 'create' ? 'create' : 'update']).not.toHaveBeenCalled()
  })

  it('探测缺少模型或提示词时回到通用并定位具体字段', async () => {
    const wrapper = await open(mode, 'openai')
    await wrapper.get('[data-group-field="probe"] button').trigger('click')
    await tab(wrapper, 'routing')
    await wrapper.get(`#${mode}-group-form`).trigger('submit')
    await flushPromises()
    expect(wrapper.get('[data-group-tab="general"]').isVisible()).toBe(true)
    expect(showError).toHaveBeenLastCalledWith('admin.groups.availabilityProbe.modelRequired')
    wrapper.findAllComponents(Select).find(select => select.attributes('data-group-field') === 'probe-model')!
      .vm.$emit('update:modelValue', 'gpt-test')
    await wrapper.get('[data-group-field="probe-prompt"]').setValue('   ')
    await tab(wrapper, 'protocol')
    await wrapper.get(`#${mode}-group-form`).trigger('submit')
    await flushPromises()
    expect(wrapper.get('[data-group-tab="general"]').isVisible()).toBe(true)
    expect(document.activeElement).toBe(wrapper.get('[data-group-field="probe-prompt"]').element)
    expect(showError).toHaveBeenLastCalledWith('admin.groups.availabilityProbe.promptRequired')
    expect(groups[mode === 'create' ? 'create' : 'update']).not.toHaveBeenCalled()
  })

  it('模型列表保留展示选择结果，提交不包含已移除的分组设置', async () => {
    const wrapper = await open(mode, 'antigravity')
    await tab(wrapper, 'features')
    for (const setting of ['claude', 'gemini_text', 'gemini_image', 'mcp_xml_inject']) {
      expect(wrapper.find(`[data-group-setting="${setting}"]`).exists()).toBe(false)
    }
    await tab(wrapper, 'protocol')
    await wrapper.get('[data-group-setting="enabled"]').trigger('click')
    const model = wrapper.get('[data-model-visibility="gpt-test"]')
    expect(model.attributes('aria-checked')).toBe('true')
    await model.trigger('click')
    expect(model.attributes('aria-checked')).toBe('false')
    expect(wrapper.find('input[type="checkbox"]').exists()).toBe(false)
    await wrapper.get(`#${mode}-group-form`).trigger('submit')
    await flushPromises()
    const payload = mode === 'create' ? groups.create.mock.calls[0]?.[0] : groups.update.mock.calls[0]?.[1]
    expect(payload).not.toHaveProperty('supported_model_scopes')
    expect(payload).not.toHaveProperty('mcp_xml_inject')
    expect(payload.models_list_config).toMatchObject({ enabled: true, models: [] })
  })
})

useProtocolCatalogFixture()

it('首次加载目录后初始化创建默认值，清空后提交不恢复默认值', async () => {
  const { protocolCatalog } = await import('@/api/admin/protocolCapabilities')
  const { default: client } = await import('@/api/client')
  const { default: fixture } = await import('@/__tests__/fixtures/protocol-catalog.json')
  protocolCatalog.value = null
  const request = vi.spyOn(client, 'get').mockResolvedValue({ data: structuredClone(fixture) })
  try {
    const wrapper = await open('create', 'anthropic')
    const selector = wrapper.getComponent(GroupClientProtocolSelector)
    const defaults = fixture.groups[0]
    expect(selector.props('modelValue')).toEqual(defaults.defaults)
    expect(selector.props('fallbacks')).toEqual(defaults.default_fallbacks)
    selector.vm.$emit('update:modelValue', [])
    selector.vm.$emit('update:fallbacks', {})
    await flushPromises()
    expect(selector.props('modelValue')).toEqual([])
    await wrapper.get('#create-group-form').trigger('submit')
    await flushPromises()
    expect(groups.create.mock.calls[0]?.[0]).toMatchObject({ allowed_protocols: [], protocol_fallbacks: {} })
    await wrapper.get('[data-tour="groups-create-btn"]').trigger('click')
    await flushPromises()
    expect(wrapper.getComponent(GroupClientProtocolSelector).props('modelValue')).toEqual(defaults.defaults)
    expect(wrapper.getComponent(GroupClientProtocolSelector).props('fallbacks')).toEqual(defaults.default_fallbacks)
  } finally { request.mockRestore() }
})

it.each(['create', 'edit'] as const)('目录失败时阻止 %s 提交，重试恢复且保留编辑的空配置', async mode => {
  const { protocolCatalog } = await import('@/api/admin/protocolCapabilities')
  const { default: client } = await import('@/api/client')
  const { default: fixture } = await import('@/__tests__/fixtures/protocol-catalog.json')
  protocolCatalog.value = null
  const request = vi.spyOn(client, 'get').mockRejectedValue(new Error('offline'))
  try {
    const wrapper = await open(mode, 'anthropic', { allowed_protocols: [], protocol_fallbacks: {} })
    await wrapper.get(`#${mode}-group-form`).trigger('submit')
    await flushPromises()
    expect(groups.create).not.toHaveBeenCalled()
    expect(groups.update).not.toHaveBeenCalled()
    expect(wrapper.getComponent(GroupClientProtocolSelector).find('[role="alert"]').exists()).toBe(true)
    request.mockResolvedValue({ data: structuredClone(fixture) })
    await wrapper.get('[data-testid="protocol-catalog-retry"]').trigger('click')
    await flushPromises()
    const selector = wrapper.getComponent(GroupClientProtocolSelector)
    expect(selector.find('[role="alert"]').exists()).toBe(false)
    expect(selector.props('modelValue')).toEqual(mode === 'create' ? fixture.groups[0].defaults : [])
    await wrapper.get(`#${mode}-group-form`).trigger('submit')
    await flushPromises()
    expect(groups[mode === 'create' ? 'create' : 'update']).toHaveBeenCalledTimes(1)
  } finally { request.mockRestore() }
})
