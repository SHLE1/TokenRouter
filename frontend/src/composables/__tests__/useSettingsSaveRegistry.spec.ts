import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { createSettingsSaveRegistry } from '@/composables/useSettingsSaveRegistry'

describe('createSettingsSaveRegistry', () => {
  it('saves only dirty targets in registration order and keeps going after a failure', async () => {
    const { dirty, saveDirty, registry } = createSettingsSaveRegistry()
    const calls: string[] = []
    const firstDirty = ref(true)
    const secondDirty = ref(false)
    const thirdDirty = ref(true)
    registry.register('first', {
      dirty: firstDirty,
      save: vi.fn(async () => {
        calls.push('first')
        return false
      }),
    })
    registry.register('second', {
      dirty: secondDirty,
      save: vi.fn(async () => {
        calls.push('second')
        return true
      }),
    })
    registry.register('third', {
      dirty: thirdDirty,
      save: vi.fn(async () => {
        calls.push('third')
        return true
      }),
    })

    expect(dirty.value).toBe(true)
    await expect(saveDirty()).resolves.toBe(false)
    expect(calls).toEqual(['first', 'third'])
  })

  it('stops tracking a target after it is unregistered', () => {
    const { dirty, registry } = createSettingsSaveRegistry()
    registry.register('only', { dirty: ref(true), save: async () => true })
    expect(dirty.value).toBe(true)

    registry.unregister('only')
    expect(dirty.value).toBe(false)
  })
})
