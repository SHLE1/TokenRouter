import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ResponsesWSSettings from '../ResponsesWSSettings.vue'
import { createSettingsSaveRegistry, settingsSaveRegistryKey } from '@/composables/useSettingsSaveRegistry'

const { getSettings, updateSettings } = vi.hoisted(() => ({ getSettings: vi.fn(), updateSettings: vi.fn() }))
vi.mock('@/api/admin/settings', () => ({ getSettings, updateSettings }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

const initial = {
  responses_ws: { read_timeout_seconds: 90 },
  responses_ws_effective: { read_timeout_seconds: 90, client_read_limit_bytes: 67108864, max_conns_per_provider: 128 }
}
async function setup() {
  const registry = createSettingsSaveRegistry()
  const wrapper = mount(ResponsesWSSettings, {
    global: {
      provide: { [settingsSaveRegistryKey as symbol]: registry.registry },
      stubs: {
        SettingsCard: { template: '<div><slot name="actions"/><slot/></div>' },
        SettingsSection: { template: '<section><slot/></section>' },
        SettingsNotice: { template: '<div><slot/></div>' },
        SettingRow: { template: '<div><slot/></div>' },
        SettingToggleRow: true
      }
    }
  })
  await flushPromises()
  return { wrapper, registry }
}

beforeEach(() => {
  vi.clearAllMocks()
  getSettings.mockResolvedValue(initial)
  updateSettings.mockResolvedValue(initial)
})

describe('Responses WS online settings', () => {
  it('submits only edited fields and converts MiB to bytes', async () => {
    const { wrapper, registry } = await setup()
    expect(registry.dirty.value).toBe(false)
    await wrapper.get('#responses-ws-input-client_read_limit_bytes').setValue('32')
    expect(registry.dirty.value).toBe(true)
    expect(await registry.saveDirty()).toBe(true)
    expect(updateSettings).toHaveBeenCalledWith({ responses_ws: { client_read_limit_bytes: 33554432 } })
    expect(registry.dirty.value).toBe(false)
  })
  it('does not save when an edit is reverted', async () => {
    const { wrapper, registry } = await setup()
    await wrapper.get('#responses-ws-input-read_timeout_seconds').setValue('100')
    expect(registry.dirty.value).toBe(true)
    await wrapper.get('#responses-ws-input-read_timeout_seconds').setValue('90')
    expect(registry.dirty.value).toBe(false)
    await registry.saveDirty()
    expect(updateSettings).not.toHaveBeenCalled()
  })
  it('clears a single override with null', async () => {
    const { wrapper, registry } = await setup()
    await wrapper.get('#responses-ws-input-read_timeout_seconds').setValue('')
    await registry.saveDirty()
    expect(updateSettings).toHaveBeenCalledWith({ responses_ws: { read_timeout_seconds: null } })
  })
  it('restores all deployment defaults with a null object', async () => {
    const { wrapper, registry } = await setup()
    await wrapper.get('button').trigger('click')
    await registry.saveDirty()
    expect(updateSettings).toHaveBeenCalledWith({ responses_ws: null })
  })
  it('keeps edits when persisted settings could not be applied', async () => {
    const { wrapper, registry } = await setup()
    updateSettings.mockRejectedValue({ response: { data: { reason: 'SETTINGS_APPLY_FAILED', metadata: { persisted: true } } } })
    await wrapper.get('#responses-ws-input-read_timeout_seconds').setValue('100')
    expect(await registry.saveDirty()).toBe(false)
    expect(registry.dirty.value).toBe(true)
    expect(wrapper.text()).toContain('admin.settings.responsesWS.applyError')
  })
})
