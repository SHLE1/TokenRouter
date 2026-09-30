import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import FilterDropdown from '../FilterDropdown.vue'
import { BREAKPOINT_LG } from '@/constants/layout'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

// 把触发区放在指定横坐标，模拟按钮位于页面左侧或右侧。
const openAt = async (left: number, wide: boolean) => {
  const wrapper = mount(FilterDropdown, {
    props: { activeCount: 0, wide },
    global: { stubs: { Icon: true, MotionTransition: { template: '<div><slot /></div>' } } },
  })
  vi.spyOn(wrapper.element, 'getBoundingClientRect').mockReturnValue({ left } as DOMRect)
  await wrapper.get('button').trigger('click')
  return wrapper.get('.dropdown')
}

describe('FilterDropdown', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('加宽面板右侧放得下时向右展开', async () => {
    window.innerWidth = BREAKPOINT_LG
    const panel = await openAt(100, true)
    expect(panel.classes()).toContain('left-0')
  })

  it('加宽面板右侧放不下时向左展开', async () => {
    window.innerWidth = BREAKPOINT_LG
    const panel = await openAt(1000, true)
    expect(panel.classes()).toContain('right-0')
  })

  it('默认面板保持原有定位', async () => {
    const panel = await openAt(1000, false)
    expect(panel.classes()).toEqual(expect.arrayContaining(['right-0', 'w-72', 'sm:left-0']))
  })
})
