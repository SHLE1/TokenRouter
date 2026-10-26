import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import OpenCodeFields from '../OpenCodeFields.vue'
import { openCodeCredentials, openCodeDraft } from '../openCodeDraft'

// OpenCodeFields.spec.ts exercises model synchronization and draft ownership
// through the rendered form. Mocked HTTP responses isolate how the form captures
// the requested service mode, merges protocol suggestions, and rejects responses
// that belong to a form which has already been closed.

// The mock records preview payloads while tests control when upstream results arrive.
const api = vi.hoisted(() => ({ syncUpstreamModels: vi.fn(), syncUpstreamModelsPreview: vi.fn() }))
vi.mock('@/api/admin/providers', () => api)
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

beforeEach(() => vi.clearAllMocks())

describe('OpenCodeFields', () => {
  // A Go draft already contains an administrator override before synchronization.
  // The request carries Go mode, suggestions remain separate, and the parent receives model IDs.
  it('synchronizes the selected mode and keeps explicit model choices', async () => {
    const draft = openCodeDraft({ provider_mode: 'go' })
    draft.protocols.go.custom = 'systemone'
    api.syncUpstreamModelsPreview.mockResolvedValue({
      models: ['custom', 'new'],
      model_protocols: { custom: 'openai_chat_completions', new: 'anthropic_messages' },
    })
    const wrapper = mount(OpenCodeFields, {
      props: { modelValue: draft, apiKey: 'test', baseUrl: '' },
      global: { stubs: { Select: true, Icon: true } },
    })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(api.syncUpstreamModelsPreview).toHaveBeenCalledWith(
      {
        platform: 'opencode',
        type: 'apikey',
        api_key: 'test',
        base_url: '',
        provider_mode: 'go',
      },
      expect.any(AbortSignal),
    )
    expect(draft.protocols.go).toEqual({ custom: 'systemone' })
    expect(draft.suggestions.go).toEqual({
      custom: 'openai_chat_completions',
      new: 'anthropic_messages',
    })
    expect(draft.protocols.payg).toEqual({})
    expect(openCodeCredentials(draft).opencode_catalog_protocols).toEqual({
      go: { custom: 'openai_chat_completions', new: 'anthropic_messages' },
      payg: {},
    })
    expect(wrapper.emitted('synced')).toEqual([[['custom', 'new']]])
    wrapper.unmount()
  })

  // A synchronization completes after the administrator closes the form. Its
  // result must not replace the retained draft or emit model updates to the parent.
  it('discards synchronization results after unmount', async () => {
    let finish!: (value: unknown) => void
    api.syncUpstreamModelsPreview.mockReturnValue(
      new Promise((resolve) => {
        finish = resolve
      }),
    )
    const draft = openCodeDraft()
    const wrapper = mount(OpenCodeFields, {
      props: { modelValue: draft, apiKey: 'test' },
      global: { stubs: { Select: true, Icon: true } },
    })
    await wrapper.find('button').trigger('click')
    wrapper.unmount()
    finish({ models: ['stale'], model_protocols: { stale: 'systemone' } })
    await flushPromises()
    expect(draft.protocols.payg).toEqual({})
    expect(wrapper.emitted('synced')).toBeUndefined()
  })
  // A saved Go provider has a manual protocol removed before or during synchronization.
  // Catalogue responses remain visible while persistence retains the administrator's deletion.
  it.each(['before', 'during'])(
    'keeps overrides deleted %s synchronization removed',
    async (timing) => {
      let finish!: (value: unknown) => void
      api.syncUpstreamModels.mockReturnValue(
        new Promise((resolve) => {
          finish = resolve
        }),
      )
      const draft = reactive(
        openCodeDraft({
          provider_mode: 'go',
          opencode_model_protocols: { go: { 'minimax-m3': 'openai_chat_completions' } },
        }),
      )
      const wrapper = mount(OpenCodeFields, {
        props: {
          modelValue: draft,
          providerId: 44,
          savedMode: 'go',
          apiKey: '',
          baseUrl: '',
          savedBaseUrl: '',
        },
        global: { stubs: { Select: true, Icon: true } },
      })
      if (timing === 'before')
        await wrapper.find('button[aria-label="common.delete"]').trigger('click')
      await wrapper.find('button').trigger('click')
      if (timing === 'during')
        await wrapper.find('button[aria-label="common.delete"]').trigger('click')
      finish({ models: ['minimax-m3'], model_protocols: { 'minimax-m3': 'anthropic_messages' } })
      await flushPromises()
      expect(api.syncUpstreamModels).toHaveBeenCalledWith(44, expect.any(AbortSignal))
      expect(api.syncUpstreamModelsPreview).not.toHaveBeenCalled()
      expect(draft.protocols.go).toEqual({})
      expect(draft.suggestions.go).toEqual({ 'minimax-m3': 'anthropic_messages' })
      expect(openCodeCredentials(draft).opencode_model_protocols).toEqual({ go: {}, payg: {} })
      expect(openCodeCredentials(draft).opencode_catalog_protocols).toEqual({
        go: { 'minimax-m3': 'anthropic_messages' },
        payg: {},
      })
      expect(wrapper.find('button[aria-label="common.delete"]').exists()).toBe(false)
      expect(wrapper.text()).toContain('admin.providers.opencode.catalogSuggestion')
      wrapper.unmount()
    },
  )

  // A saved provider's key, root, or mode differs from the edited draft.
  // The form explains the required save and prevents both saved and preview queries.
  it.each(['key', 'root', 'mode'])(
    'blocks unsaved %s changes for saved providers',
    async (field) => {
      const draft = reactive(openCodeDraft({ provider_mode: field === 'mode' ? 'go' : 'payg' }))
      const wrapper = mount(OpenCodeFields, {
        props: {
          modelValue: draft,
          providerId: 44,
          savedMode: 'payg',
          savedBaseUrl: '',
          apiKey: field === 'key' ? 'replacement' : '',
          baseUrl: field === 'root' ? 'https://temporary.example' : '',
        },
        global: { stubs: { Select: true, Icon: true } },
      })
      expect(wrapper.find('button').attributes('disabled')).toBeDefined()
      expect(wrapper.find('[data-testid="opencode-sync-settings"]').text()).toBe(
        'admin.providers.opencode.saveBeforeSync',
      )
      await wrapper.find('button').trigger('click')
      expect(api.syncUpstreamModels).not.toHaveBeenCalled()
      expect(api.syncUpstreamModelsPreview).not.toHaveBeenCalled()
      wrapper.unmount()
    },
  )

  // A preview root changes and then returns while its first request is pending.
  // Cancellation prevents that stale response from publishing even though the final inputs match.
  it('cancels changed preview identity and discards change-and-revert responses', async () => {
    let finish!: (value: unknown) => void
    api.syncUpstreamModelsPreview.mockReturnValue(
      new Promise((resolve) => {
        finish = resolve
      }),
    )
    const draft = reactive(openCodeDraft())
    const wrapper = mount(OpenCodeFields, {
      props: { modelValue: draft, apiKey: 'test', baseUrl: 'https://first.example' },
      global: { stubs: { Select: true, Icon: true } },
    })
    await wrapper.find('button').trigger('click')
    const signal = api.syncUpstreamModelsPreview.mock.calls[0]![1] as AbortSignal
    await wrapper.setProps({ baseUrl: 'https://second.example' })
    await wrapper.setProps({ baseUrl: 'https://first.example' })
    expect(signal.aborted).toBe(true)
    finish({ models: ['stale'], model_protocols: { stale: 'systemone' } })
    await flushPromises()
    expect(draft.suggestions.payg).toEqual({})
    expect(wrapper.emitted('synced')).toBeUndefined()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    wrapper.unmount()
  })

  // A user selects a protocol while an upstream catalogue request is pending.
  // Completion updates suggestions while retaining that current manual selection in persistence.
  it('keeps a manual edit made while synchronization is pending', async () => {
    let finish!: (value: unknown) => void
    api.syncUpstreamModelsPreview.mockReturnValue(
      new Promise((resolve) => {
        finish = resolve
      }),
    )
    const draft = reactive(
      openCodeDraft({ opencode_model_protocols: { payg: { custom: 'systemone' } } }),
    )
    const wrapper = mount(OpenCodeFields, {
      props: { modelValue: draft, apiKey: 'test' },
      global: { stubs: { Select: true, Icon: true } },
    })
    await wrapper.find('button').trigger('click')
    wrapper
      .findAllComponents({ name: 'Select' })[1]!
      .vm.$emit('update:model-value', 'openai_responses')
    finish({ models: ['custom'], model_protocols: { custom: 'anthropic_messages' } })
    await flushPromises()
    expect(draft.protocols.payg.custom).toBe('openai_responses')
    expect(draft.suggestions.payg.custom).toBe('anthropic_messages')
    expect(openCodeCredentials(draft).opencode_model_protocols).toEqual({
      payg: { custom: 'openai_responses' },
      go: {},
    })
    wrapper.unmount()
  })
})
