<template>
  <header class="site-header fixed inset-x-0 top-0 z-header border-b border-primary-900/10">
    <!-- 水平内边距与主内容区保持同一条链，两侧边缘在所有断点对齐。 -->
    <div class="flex h-[var(--header-h)] items-center justify-between gap-3 px-4 md:px-6 lg:px-8">
      <!-- 品牌固定在全局顶栏，避免与侧栏和页面标题争夺层级。 -->
      <div class="flex min-w-0 shrink-0 items-center gap-2 sm:gap-4">
        <button
          @click="handlePrimaryNavigation"
          :class="['btn-ghost btn-icon', !isCreativeStudio && 'lg:hidden']"
          :aria-label="isCreativeStudio ? t('creative.canvas.backToDashboard') : t('common.toggleMenu')"
          :title="isCreativeStudio ? t('creative.canvas.backToDashboard') : t('common.toggleMenu')"
        >
          <Icon :name="isCreativeStudio ? 'home' : 'menu'" size="md" />
        </button>

        <!-- 版本标签与首页链接分离，避免按钮嵌套在链接内触发错误跳转。 -->
        <div class="header-brand flex min-w-0 items-center gap-2.5 rounded-control px-1.5 py-1 transition-colors hover:bg-primary-100/70 dark:hover:bg-dark-700">
          <router-link
            :to="homePath"
            class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-control bg-primary-100 dark:bg-dark-800"
            :aria-label="siteName"
          >
            <img v-if="settingsLoaded" :src="siteLogo || '/logo.svg'" :alt="siteName" class="h-full w-full object-contain" />
          </router-link>
          <span class="hidden min-w-0 sm:block">
            <router-link
              :to="homePath"
              class="block max-w-44 truncate text-base font-bold leading-tight text-gray-900 dark:text-white"
            >{{ siteName }}</router-link>
            <VersionBadge :version="siteVersion" />
          </span>
        </div>
      </div>

      <!-- 右侧分为工具区和账户区：工具区是同尺寸图标按钮，账户区是余额按钮和头像。 -->
      <div class="header-status-actions">
        <div class="header-status-icon-group">
          <div v-if="user" class="hidden sm:block">
            <AnnouncementBell variant="status" />
          </div>

          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="header-status-icon-button hidden sm:flex"
            :aria-label="t('nav.docs')"
            :title="t('nav.docs')"
          >
            <Icon name="book" size="md" />
          </a>

          <LocaleSwitcher variant="status" />

          <!-- 主题切换在窄屏始终保留，公告和文档入口优先让出空间。 -->
          <button
            type="button"
            data-testid="theme-toggle"
            class="header-status-icon-button"
            :aria-label="isDark ? t('nav.lightMode') : t('nav.darkMode')"
            :title="isDark ? t('nav.lightMode') : t('nav.darkMode')"
            @click="toggleTheme"
          >
            <!-- 太阳和月亮保留各自的颜色，让主题切换一眼可辨。 -->
            <Icon
              :name="isDark ? 'sun' : 'moon'"
              size="md"
              :class="isDark ? 'text-amber-500' : 'text-blue-500'"
            />
          </button>
        </div>

        <template v-if="user">
          <div class="header-status-divider hidden sm:block"></div>

          <!-- 余额按钮右侧的状态点表示订阅用量，点击展开订阅详情；窄屏余额收进用户菜单。 -->
          <SubscriptionProgressMini variant="status" class="hidden sm:block">
            <span class="text-primary-900/60 dark:text-dark-400">{{ balanceUnitSymbol }}</span>
            <span class="font-semibold tabular-nums text-primary-900 dark:text-dark-100">
              {{ formatHeaderMoney(availableBalance, false) }}
            </span>
            <span
              v-if="frozenBalance > 0"
              class="ml-1 text-xs font-medium text-amber-600 dark:text-amber-300"
              :title="balanceFrozenLabel"
            >
              {{ balanceFrozenLabel }}
            </span>
          </SubscriptionProgressMini>
        </template>

        <!-- 用户下拉菜单入口只保留头像，尺寸与图标按钮等高。 -->
        <div v-if="user" class="relative" ref="dropdownRef">
          <button
            @click="toggleDropdown"
            class="header-status-user-button"
            :aria-label="t('common.userMenu')"
          >
            <UserAvatar
              :avatar-url="avatarUrl"
              :user-id="user.id"
              :alt="displayName"
              size-class="h-9 w-9"
            />
          </button>

          <!-- Dropdown Menu -->
          <MotionTransition name="dropdown-fade">
            <div v-if="dropdownOpen" class="dropdown right-0 z-50 mt-2 w-64 origin-top-right">
              <!-- User Info -->
              <div class="border-b border-primary-900/10 px-4 py-3 dark:border-dark-600">
                <div class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ displayName }}
                </div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ user.email }}</div>
              </div>

              <!-- Balance (mobile only) -->
              <div class="border-b border-primary-900/10 px-4 py-2 dark:border-dark-600 sm:hidden">
                <div class="text-xs text-gray-500 dark:text-dark-400">
                  {{ t('common.balance') }}
                </div>
                <div class="text-sm font-semibold text-primary-600 dark:text-primary-400">
                  {{ formatHeaderMoney(availableBalance) }}
                </div>
                <div v-if="frozenBalance > 0" class="mt-1 text-xs text-amber-600 dark:text-amber-300">
                  {{ balanceFrozenText }} {{ formatHeaderMoney(frozenBalance) }}
                </div>
              </div>

              <div class="py-1">
                <router-link to="/profile" @click="closeDropdown" class="dropdown-item dropdown-item-brand">
                  <Icon name="user" size="sm" />
                  {{ t('nav.profile') }}
                </router-link>

                <router-link to="/keys" @click="closeDropdown" class="dropdown-item dropdown-item-brand">
                  <Icon name="key" size="sm" />
                  {{ t('nav.apiKeys') }}
                </router-link>

                <a
                  v-if="authStore.isAdmin"
                  href="https://github.com/TokenFlux/TokenRouter"
                  target="_blank"
                  rel="noopener noreferrer"
                  @click="closeDropdown"
                  class="dropdown-item dropdown-item-brand"
                >
                  <svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
                    <path
                      fill-rule="evenodd"
                      clip-rule="evenodd"
                      d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.604-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.464-1.11-1.464-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.167 22 16.418 22 12c0-5.523-4.477-10-10-10z"
                    />
                  </svg>
                  {{ t('nav.github') }}
                </a>

              </div>

              <!-- Contact Support (only show if configured) -->
              <div
                v-if="contactInfo"
                class="border-t border-primary-900/10 px-4 py-3 dark:border-dark-600"
              >
                <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                  <Icon name="chat" size="xs" class="h-3.5 w-3.5 flex-shrink-0" />
                  <span>{{ t('common.contactSupport') }}</span>
                </div>
                <ul class="mt-2 space-y-1.5">
                  <li
                    v-for="(entry, idx) in contactEntries"
                    :key="idx"
                    class="text-xs leading-relaxed"
                  >
                    <template v-if="entry.label">
                      <span class="text-gray-500 dark:text-gray-400">{{ entry.label }}</span>
                      <span class="text-gray-400 dark:text-gray-500">：</span>
                    </template>
                    <a
                      v-if="entry.url"
                      :href="entry.url"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="break-all font-medium text-primary-600 hover:underline dark:text-primary-400"
                    >{{ entry.value }}</a>
                    <span v-else class="break-all font-medium text-gray-700 dark:text-gray-200">{{ entry.value }}</span>
                  </li>
                </ul>
              </div>

              <div v-if="showOnboardingButton" class="border-t border-primary-900/10 py-1 dark:border-dark-600">
                <button @click="handleReplayGuide" class="dropdown-item dropdown-item-brand w-full">
                  <Icon name="questionCircle" size="sm" class="h-4 w-4" />
                  {{ $t('onboarding.restartTour') }}
                </button>
              </div>

              <div class="border-t border-primary-900/10 py-1 dark:border-dark-600">
                <button
                  @click="handleLogout"
                  class="dropdown-item w-full text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
                >
                  <Icon name="logout" size="sm" class="h-4 w-4" />
                  {{ t('nav.logout') }}
                </button>
              </div>
            </div>
          </MotionTransition>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import MotionTransition from '@/components/common/MotionTransition.vue'
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import Icon from '@/components/icons/Icon.vue'
import VersionBadge from '@/components/common/VersionBadge.vue'
import { sanitizeUrl } from '@/utils/url'
import { useBalanceDisplay } from '@/composables/useBalanceDisplay'
import { useTheme } from '@/composables/useTheme'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const onboardingStore = useOnboardingStore()
const { formatBalanceAmount, balanceUnitSymbol } = useBalanceDisplay()
const { isDark, toggleTheme } = useTheme()

const user = computed(() => authStore.user)
const isCreativeStudio = computed(() => route.path === '/creative')
const homePath = '/home'
const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteVersion = computed(() => appStore.siteVersion)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)
const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const contactInfo = computed(() => appStore.contactInfo)

// 联系客服是自由文本(如"闲聊群(QQ)：123，TG群：https://t.me/xxx"),
// 按逗号/分号拆条,再按冒号拆"标签：值",URL 渲染为可点击链接
interface ContactEntry {
  label: string
  value: string
  url: string
}

const contactEntries = computed<ContactEntry[]>(() => {
  const raw = contactInfo.value?.trim()
  if (!raw) return []
  return raw
    .split(/[，,;；\n]+/)
    .map(part => part.trim())
    .filter(Boolean)
    .map(part => {
      // 找第一个不属于协议(://)的冒号作为"标签：值"分隔符
      let sep = -1
      for (let i = 0; i < part.length; i++) {
        const ch = part[i]
        if (ch === '：') { sep = i; break }
        if (ch === ':' && part.slice(i, i + 3) !== '://') { sep = i; break }
      }
      let label = ''
      let value = part
      if (sep > 0) {
        label = part.slice(0, sep).trim()
        value = part.slice(sep + 1).trim()
      }
      const url = /^https?:\/\/\S+$/.test(value) ? value : ''
      return { label, value, url }
    })
    .filter(e => e.value)
})
const docUrl = computed(() => sanitizeUrl(appStore.docUrl))
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')
const availableBalance = computed(() => Number(user.value?.balance || 0))
const frozenBalance = computed(() => Number(user.value?.frozen_balance || 0))
const balanceFrozenText = computed(() => t('common.frozenBalance') === 'common.frozenBalance' ? '冻结金额' : t('common.frozenBalance'))
const balanceFrozenLabel = computed(() => `${balanceFrozenText.value} ${formatHeaderMoney(frozenBalance.value)}`)

// 只向管理员显示新手引导按钮
const showOnboardingButton = computed(() => {
  return user.value?.role === 'admin'
})

const displayName = computed(() => {
  if (!user.value) return ''
  return user.value.username || user.value.email?.split('@')[0] || ''
})

function toggleMobileSidebar() {
  appStore.toggleMobileSidebar()
}

function handlePrimaryNavigation() {
  if (isCreativeStudio.value) {
    void router.push('/dashboard')
    return
  }
  toggleMobileSidebar()
}

function toggleDropdown() {
  dropdownOpen.value = !dropdownOpen.value
}

function closeDropdown() {
  dropdownOpen.value = false
}

async function handleLogout() {
  closeDropdown()
  try {
    await authStore.logout()
  } catch (error) {
    // Ignore logout errors - still redirect to login
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

function handleReplayGuide() {
  closeDropdown()
  onboardingStore.replay()
}

// withSymbol 为 false 时只返回数字，供顶栏余额按钮单独排版货币符号。
function formatHeaderMoney(value: number, withSymbol = true) {
  return formatBalanceAmount(Number.isFinite(value) ? value : 0, { fractionDigits: 2, withSymbol })
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    closeDropdown()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.header-status-actions {
  @apply ml-auto flex min-w-0 shrink-0 items-center gap-2 sm:gap-5;
}

.header-brand {
  max-width: min(18rem, 42vw);
}

.header-status-icon-group {
  @apply flex items-center gap-1 sm:gap-2;
}

.header-status-divider {
  @apply h-5 w-px shrink-0 bg-primary-900/10 dark:bg-dark-600;
}

.header-status-icon-button {
  @apply flex h-9 w-9 items-center justify-center rounded-control text-primary-900 transition-colors hover:bg-primary-100 dark:text-dark-100 dark:hover:bg-dark-700 dark:hover:text-white;
}

.header-status-user-button {
  @apply flex h-9 w-9 items-center justify-center rounded-full ring-1 ring-primary-200/70 transition-shadow hover:ring-primary-300 dark:ring-dark-600 dark:hover:ring-dark-400;
}
</style>
