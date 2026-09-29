<template>
  <article class="card flex flex-col p-6 transition-colors hover:border-black/20 dark:hover:border-dark-500">
    <!-- 名称固定两行高度，保证同一行卡片的价格区对齐。 -->
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0 flex-1">
        <h3
          :title="plan.name"
          class="h-12 min-w-0 break-words [overflow-wrap:anywhere] text-base font-semibold leading-6 text-gray-900 dark:text-white line-clamp-2"
        >
          {{ plan.name }}
        </h3>
      </div>
      <span
        v-if="discountText"
        class="shrink-0 rounded-compact bg-red-500/10 px-1.5 py-0.5 text-xs font-semibold text-red-600 dark:text-red-400"
      >
        {{ discountText }}
      </span>
    </div>
    <p v-if="plan.description" class="mt-1 text-sm leading-relaxed text-gray-500 dark:text-dark-400 line-clamp-2">
      {{ plan.description }}
    </p>

    <!-- 价格 -->
    <div class="mt-4">
      <div class="flex flex-wrap items-baseline gap-x-1">
        <span class="text-lg font-semibold text-gray-900 dark:text-white">{{ planCurrencySymbol }}</span>
        <span class="text-3xl font-semibold tracking-tight tabular-nums text-gray-900 dark:text-white">{{ plan.price }}</span>
        <span v-if="plan.currency" class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ plan.currency }}</span>
        <span class="ml-1 text-sm text-gray-500 dark:text-dark-400">/ {{ validitySuffix }}</span>
      </div>
      <div v-if="plan.original_price" class="mt-1 text-sm text-gray-400 line-through dark:text-dark-500">
        {{ planCurrencySymbol }}{{ plan.original_price }}<template v-if="plan.currency"> {{ plan.currency }}</template>
      </div>
    </div>

    <!-- 套餐额度信息 -->
    <dl class="mt-5 space-y-2 border-t border-gray-100 pt-5 text-sm dark:border-dark-700">
      <div v-if="hasPlanQuota(plan.daily_limit_usd)" class="flex items-center justify-between gap-3">
        <dt class="text-gray-500 dark:text-dark-400">{{ t('payment.planCard.dailyLimit') }}</dt>
        <dd class="font-medium tabular-nums text-gray-900 dark:text-dark-100">{{ formatPlanQuota(plan.daily_limit_usd) }}</dd>
      </div>
      <div v-if="hasPlanQuota(plan.weekly_limit_usd)" class="flex items-center justify-between gap-3">
        <dt class="text-gray-500 dark:text-dark-400">{{ t('payment.planCard.weeklyLimit') }}</dt>
        <dd class="font-medium tabular-nums text-gray-900 dark:text-dark-100">{{ formatPlanQuota(plan.weekly_limit_usd) }}</dd>
      </div>
      <div v-if="hasPlanQuota(plan.monthly_limit_usd)" class="flex items-center justify-between gap-3">
        <dt class="text-gray-500 dark:text-dark-400">{{ t('payment.planCard.monthlyLimit') }}</dt>
        <dd class="font-medium tabular-nums text-gray-900 dark:text-dark-100">{{ formatPlanQuota(plan.monthly_limit_usd) }}</dd>
      </div>
      <div v-if="isUnlimited" class="flex items-center justify-between gap-3">
        <dt class="text-gray-500 dark:text-dark-400">{{ t('payment.planCard.quota') }}</dt>
        <dd class="font-medium text-gray-900 dark:text-dark-100">{{ t('payment.planCard.unlimited') }}</dd>
      </div>
      <div v-if="modelScopeLabels.length > 0" class="flex items-start justify-between gap-3">
        <dt class="shrink-0 text-gray-500 dark:text-dark-400">{{ t('payment.planCard.models') }}</dt>
        <dd class="flex flex-wrap justify-end gap-1">
          <span
            v-for="scope in modelScopeLabels"
            :key="scope"
            class="rounded-compact bg-gray-100 px-1.5 py-0.5 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-200"
          >
            {{ scope }}
          </span>
        </dd>
      </div>
    </dl>

    <!-- 功能列表 -->
    <ul v-if="plan.features.length > 0" class="mt-4 space-y-2">
      <li v-for="feature in plan.features" :key="feature" class="flex items-start gap-2">
        <Icon
          name="check"
          size="sm"
          :animate-on-hover="false"
          class="mt-0.5 shrink-0 text-primary-500 dark:text-primary-400"
        />
        <span class="text-sm text-gray-600 dark:text-dark-200">{{ feature }}</span>
      </li>
    </ul>

    <div class="flex-1" />

    <button
      type="button"
      class="btn btn-primary mt-6 w-full"
      @click="emit('select', plan)"
    >
      {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
    </button>
  </article>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import { useBalanceDisplay } from '@/composables/useBalanceDisplay'
import { currencySymbol } from '@/components/payment/currency'
import { planValiditySuffix } from './validity'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()
const { formatBalanceAmount } = useBalanceDisplay()
const planCurrencySymbol = computed(() => currencySymbol(props.plan.currency || 'USD'))

const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.plan_id === props.plan.id && s.status === 'active') ?? false
)

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})

function formatPlanQuota(value: number | null | undefined): string {
  const amount = Number(value)
  return formatBalanceAmount(value, { fractionDigits: Number.isInteger(amount) ? 0 : 2 })
}

function hasPlanQuota(value: number | null | undefined): boolean {
  return value != null && value > 0
}

// 日、周、月限额都未设置时展示为无限制。
const isUnlimited = computed(() =>
  !hasPlanQuota(props.plan.daily_limit_usd)
    && !hasPlanQuota(props.plan.weekly_limit_usd)
    && !hasPlanQuota(props.plan.monthly_limit_usd)
)

const MODEL_SCOPE_LABELS: Record<string, string> = {
  claude: 'Claude',
  gemini_text: 'Gemini',
  gemini_image: 'Imagen',
}

const modelScopeLabels = computed(() => {
  // 模型系列限制由分组策略提供，作用于相关提供商。
  const scopes = props.plan.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => MODEL_SCOPE_LABELS[s] || s)
})

const validitySuffix = computed(() => {
  return planValiditySuffix(props.plan, t)
})
</script>
