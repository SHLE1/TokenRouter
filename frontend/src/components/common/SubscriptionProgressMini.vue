<template>
  <!-- 默认变体没有生效订阅时整体隐藏；状态变体始终渲染插槽内容，只在有订阅时显示状态点并可展开。 -->
  <div v-if="variant === 'status' || hasActiveSubscriptions" ref="containerRef" class="relative">
    <button
      :class="triggerClass"
      :title="hasActiveSubscriptions ? t('subscriptionProgress.viewDetails') : undefined"
      :disabled="!hasActiveSubscriptions"
      @click="toggleTooltip"
    >
      <template v-if="variant === 'status'">
        <slot />
        <!-- 状态点取用量最高的订阅，颜色表示用量是否接近上限。 -->
        <span
          v-if="statusDotClass"
          data-testid="subscription-status-dot"
          class="ml-1 h-2 w-2 shrink-0 rounded-full"
          :class="statusDotClass"
        />
      </template>
      <div v-else class="flex items-center gap-2">
        <Icon name="creditCard" size="sm" class="text-primary-600 dark:text-primary-400" />
        <div class="flex items-center gap-0.5">
          <div
            v-for="(dotClass, index) in displayDots"
            :key="index"
            class="rounded-full"
            :class="['h-2 w-2', dotClass]"
          />
        </div>
        <span class="text-xs font-medium text-primary-700 dark:text-primary-300">
          {{ statusCount }}
        </span>
      </div>
    </button>

    <MotionTransition name="dropdown-fade">
      <div
        v-if="tooltipOpen"
        class="dropdown subscription-progress-popover right-0 z-50 mt-2 w-[340px] overflow-hidden py-0"
      >
        <div class="border-b border-gray-100 p-3 dark:border-dark-700">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('subscriptionProgress.title') }}
          </h3>
          <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
            {{ t('subscriptionProgress.activeCount', { count: activeSubscriptions.length }) }}
          </p>
        </div>

        <div class="max-h-64 overflow-y-auto">
          <div
            v-for="subscription in displaySubscriptions"
            :key="subscription.id"
            class="border-b border-gray-50 p-3 last:border-b-0 dark:border-dark-700/50"
          >
            <div class="mb-2 flex items-center justify-between gap-2">
              <span class="min-w-0 truncate text-sm font-medium text-gray-900 dark:text-white">
                {{ subscription.plan?.name || `Plan #${subscription.plan_id}` }}
              </span>
              <span class="shrink-0 whitespace-nowrap text-xs" :class="getDaysRemainingClass(subscription.expires_at)">
                {{ formatExpiration(subscription.expires_at) }}
              </span>
            </div>

            <div class="space-y-1.5">
              <div
                v-if="isUnlimited(subscription)"
                class="flex items-center gap-2 rounded-control bg-gradient-to-r from-emerald-50 to-teal-50 px-2.5 py-1.5 dark:from-emerald-900/20 dark:to-teal-900/20"
              >
                <span class="text-lg text-emerald-600 dark:text-emerald-400">∞</span>
                <span class="text-xs font-medium text-emerald-700 dark:text-emerald-300">
                  {{ t('subscriptionProgress.unlimited') }}
                </span>
              </div>

              <template v-else>
                <div
                  v-for="window in usageWindows(subscription)"
                  :key="window.key"
                  class="flex items-center gap-2"
                >
                  <span class="w-8 flex-shrink-0 text-xs text-gray-500">
                    {{ window.label }}
                  </span>
                  <div class="h-1.5 min-w-0 flex-1 rounded-full bg-gray-200 dark:bg-dark-600">
                    <div
                      class="h-1.5 rounded-full transition-[width,background-color]"
                      :class="getProgressBarClass(window.used, window.limit)"
                      :style="{ width: getProgressWidth(window.used, window.limit) }"
                    />
                  </div>
                  <span class="w-24 flex-shrink-0 text-right text-xs text-gray-500">
                    {{ formatUsage(window.used, window.limit) }}
                  </span>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div class="border-t border-gray-100 p-2 dark:border-dark-700">
          <router-link
            to="/subscriptions"
            class="block w-full py-1 text-center text-xs text-primary-600 hover:underline dark:text-primary-400"
            @click="closeTooltip"
          >
            {{ t('subscriptionProgress.viewAll') }}
          </router-link>
        </div>
      </div>
    </MotionTransition>
  </div>
</template>

<script setup lang="ts">
import MotionTransition from '@/components/common/MotionTransition.vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useBalanceDisplay } from '@/composables/useBalanceDisplay'
import { useSubscriptionStore } from '@/stores'
import type { UserSubscription } from '@/types'
import { formatDateTimeToMinute } from '@/utils/format'

const props = withDefaults(defineProps<{
  variant?: 'default' | 'status'
}>(), {
  variant: 'default'
})

const { t } = useI18n()
const subscriptionStore = useSubscriptionStore()
const { formatBalanceAmount } = useBalanceDisplay()

const containerRef = ref<HTMLElement | null>(null)
const tooltipOpen = ref(false)

const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)
const hasActiveSubscriptions = computed(() => subscriptionStore.hasActiveSubscriptions)
const variant = computed(() => props.variant)
const statusCount = computed(() => activeSubscriptions.value.length)
const triggerClass = computed(() => {
  if (variant.value === 'status') {
    // 状态变体与顶栏图标按钮同为无边框按钮，没有订阅时不响应悬停。
    return 'flex h-9 items-center gap-1 rounded-control px-3 text-sm transition-colors hover:bg-primary-100 disabled:cursor-default disabled:hover:bg-transparent dark:hover:bg-dark-700 dark:disabled:hover:bg-transparent'
  }
  return 'flex cursor-pointer items-center gap-2 rounded-control bg-primary-50 px-3 py-1.5 transition-colors hover:bg-primary-100 dark:bg-primary-900/20 dark:hover:bg-primary-900/30'
})

const displaySubscriptions = computed(() =>
  [...activeSubscriptions.value].sort((a, b) => getMaxUsagePercentage(b) - getMaxUsagePercentage(a))
)
const displayDots = computed(() =>
  displaySubscriptions.value.slice(0, 3).map((subscription) => getProgressDotClass(subscription))
)
const statusDotClass = computed(() => displayDots.value[0] ?? '')

function usageWindows(subscription: UserSubscription) {
  return [
    {
      key: 'daily',
      label: t('subscriptionProgress.daily'),
      used: subscription.daily_usage_usd || 0,
      limit: subscription.daily_limit_usd
    },
    {
      key: 'weekly',
      label: t('subscriptionProgress.weekly'),
      used: subscription.weekly_usage_usd || 0,
      limit: subscription.weekly_limit_usd
    },
    {
      key: 'monthly',
      label: t('subscriptionProgress.monthly'),
      used: subscription.monthly_usage_usd || 0,
      limit: subscription.monthly_limit_usd
    }
  ].filter((window) => window.limit != null && window.limit > 0)
}

function getMaxUsagePercentage(subscription: UserSubscription): number {
  const windows = usageWindows(subscription)
  if (windows.length === 0) return 0
  return Math.max(...windows.map((window) => ((window.used || 0) / (window.limit || 1)) * 100))
}

function isUnlimited(subscription: UserSubscription): boolean {
  return usageWindows(subscription).length === 0
}

function getProgressDotClass(subscription: UserSubscription): string {
  if (isUnlimited(subscription)) return 'bg-emerald-500'
  const percentage = getMaxUsagePercentage(subscription)
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function getProgressBarClass(used: number, limit: number | null): string {
  if (!limit || limit === 0) return 'bg-gray-400'
  const percentage = (used / limit) * 100
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function getProgressWidth(used: number, limit: number | null): string {
  if (!limit || limit === 0) return '0%'
  return `${Math.min((used / limit) * 100, 100)}%`
}

function formatUsage(used: number, limit: number | null): string {
  const usedValue = formatBalanceAmount(used, { fractionDigits: 2 })
  const limitValue = limit == null ? '∞' : formatBalanceAmount(limit, { fractionDigits: 2 })
  return `${usedValue}/${limitValue}`
}

function formatExpiration(expiresAt: string): string {
  const time = formatDateTimeToMinute(expiresAt)
  return time ? t('subscriptionProgress.expiresAt', { time }) : ''
}

function getDaysRemainingClass(expiresAt: string): string {
  const days = Math.ceil((new Date(expiresAt).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-500 dark:text-dark-400'
}

function toggleTooltip() {
  if (!hasActiveSubscriptions.value) return
  tooltipOpen.value = !tooltipOpen.value
}

function closeTooltip() {
  tooltipOpen.value = false
}

function handleClickOutside(event: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    closeTooltip()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  subscriptionStore.fetchActiveSubscriptions().catch((error) => {
    console.error('Failed to load subscriptions in SubscriptionProgressMini:', error)
  })
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>

/* 手机端顶部状态栏空间有限，弹层按视口留边居中，避免右侧按钮定位把内容挤出左边界。 */
@media (max-width: 639px) {
  .subscription-progress-popover {
    position: fixed;
    top: 4.5rem;
    right: 0.75rem;
    left: 0.75rem;
    width: auto;
    margin-top: 0;
  }
}
</style>
