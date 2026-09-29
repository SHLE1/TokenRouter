import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UserDashboardCharts from '../UserDashboardCharts.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key === 'dashboard.noDataAvailable' ? 'No data available' : key,
    }),
  }
})

vi.mock('@/composables/useBalanceDisplay', () => ({
  useBalanceDisplay: () => ({
    formatBalanceAmount: (value: number) => String(value),
    formatUsdAmount: (value: number) => String(value),
  }),
}))

describe('UserDashboardCharts', () => {
  it('图表数据加载期间显示骨架，不显示暂无数据', async () => {
    const wrapper = mount(UserDashboardCharts, {
      props: { loading: true, startDate: '', endDate: '', granularity: 'day', trend: [], models: [] },
      global: { stubs: { DateRangePicker: true, Select: true, TokenUsageTrend: true } }
    })
    expect(wrapper.find('[aria-busy="true"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="model-distribution-empty"]').exists()).toBe(false)
    await wrapper.setProps({ loading: false })
    expect(wrapper.find('[data-testid="model-distribution-empty"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('only renders the centered empty state when model data is absent', () => {
    const wrapper = mount(UserDashboardCharts, {
      props: {
        loading: false,
        startDate: '2026-08-06',
        endDate: '2026-08-12',
        granularity: 'day',
        trend: [],
        models: [],
      },
      global: {
        stubs: {
          DateRangePicker: true,
          Select: true,
          LoadingSpinner: true,
          TokenUsageTrend: true,
        },
      },
    })

    const emptyState = wrapper.get('[data-testid="model-distribution-empty"]')
    expect(emptyState.text()).toBe('No data available')
    expect(emptyState.classes()).toContain('h-48')
    expect(wrapper.find('[data-testid="model-distribution-content"]').exists()).toBe(false)
    expect(wrapper.find('table').exists()).toBe(false)
  })
})
