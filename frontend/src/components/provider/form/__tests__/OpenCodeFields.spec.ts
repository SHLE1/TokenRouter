import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import OpenCodeFields from '../OpenCodeFields.vue'
import { openCodeDraft } from '../openCodeDraft'

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
  // The request carries Go mode, suggestions fill missing entries, and the parent receives model IDs.
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
    expect(api.syncUpstreamModelsPreview).toHaveBeenCalledWith({
      platform: 'opencode',
      type: 'apikey',
      api_key: 'test',
      base_url: '',
      provider_mode: 'go',
    })
    expect(draft.protocols.go).toEqual({ custom: 'systemone', new: 'anthropic_messages' })
    expect(draft.protocols.payg).toEqual({})
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
})
