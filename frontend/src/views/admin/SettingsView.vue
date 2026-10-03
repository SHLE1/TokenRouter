<template>
  <AppLayout>
    <!-- 设置内容填满主区，切换页签时卡片宽度保持一致。 -->
    <div class="w-full min-w-0 space-y-4">
      <!-- 设置尚未返回时，先保留页签和表单控件的位置。 -->
      <SettingsSkeleton v-if="loading" />

      <!-- Settings Form -->
      <form v-else @submit.prevent="saveSettings" class="space-y-4" novalidate>
        <!-- Tab Navigation -->
        <div
          :class="[
            'settings-tabs-shell',
            activeTab === 'gateway' && 'settings-tabs-shell-stacked',
          ]"
        >
          <nav
            ref="settingsTabsScrollRef"
            class="settings-tabs-scroll"
            role="tablist"
            :aria-label="t('admin.settings.title')"
          >
            <div class="settings-tabs">
              <button
                v-for="tab in settingsTabs"
                :key="tab.key"
                :id="`settings-tab-${tab.key}`"
                type="button"
                role="tab"
                :aria-selected="activeTab === tab.key"
                :tabindex="activeTab === tab.key ? 0 : -1"
                :class="[
                  'settings-tab',
                  activeTab === tab.key && 'settings-tab-active',
                ]"
                @click="selectSettingsTab(tab.key)"
                @keydown="handleSettingsTabKeydown($event, tab.key)"
              >
                <span class="settings-tab-icon">
                  <Icon :name="tab.icon" size="sm" />
                </span>
                <span class="settings-tab-label">
                  {{ t(`admin.settings.tabs.${tab.key}`) }}
                </span>
              </button>
            </div>
          </nav>

          <nav
            v-if="activeTab === 'gateway'"
            ref="gatewaySectionsScrollRef"
            class="gateway-section-tabs-scroll"
            role="tablist"
            :aria-label="t('admin.settings.gatewaySections.label')"
          >
            <div class="gateway-section-tabs">
              <button
                v-for="section in gatewaySections"
                :id="`gateway-section-tab-${section.key}`"
                :key="section.key"
                type="button"
                role="tab"
                :aria-selected="activeGatewaySection === section.key"
                :tabindex="activeGatewaySection === section.key ? 0 : -1"
                :data-testid="`gateway-section-tab-${section.key}`"
                :class="[
                  'gateway-section-tab',
                  activeGatewaySection === section.key &&
                    'gateway-section-tab-active',
                ]"
                @click="selectGatewaySection(section.key)"
                @keydown="handleGatewaySectionKeydown($event, section.key)"
              >
                <span class="gateway-section-tab-icon">
                  <ProviderIcon
                    v-if="section.providerBrand"
                    :brand="section.providerBrand"
                    size="18px"
                    color="currentColor"
                  />
                  <Icon v-else name="cog" size="sm" />
                </span>
                <span class="gateway-section-tab-label">
                  {{ t(`admin.settings.gatewaySections.${section.key}`) }}
                </span>
              </button>
            </div>
          </nav>
        </div>

        <!-- Tab: Security — Admin API Key -->
        <div v-show="activeTab === 'security'" v-content-reveal="activeTab === 'security'" class="space-y-4">
          <!-- Admin API Key Settings -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.adminApiKey.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.adminApiKey.description") }}
              </p>
            </div>
            <div class="space-y-4 p-6">
              <!-- Security Warning -->
              <div
                class="rounded-control border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-900/20"
              >
                <div class="flex items-start">
                  <Icon
                    name="exclamationTriangle"
                    size="md"
                    class="mt-0.5 flex-shrink-0 text-amber-500"
                  />
                  <p class="ml-3 text-sm text-amber-700 dark:text-amber-300">
                    {{ t("admin.settings.adminApiKey.securityWarning") }}
                  </p>
                </div>
              </div>

              <!-- Loading State -->
              <ContentSkeleton v-if="adminApiKeyLoading" variant="form" :rows="1" />

              <!-- No Key Configured -->
              <div
                v-else-if="!adminApiKeyExists"
                class="flex items-center justify-between"
              >
                <span class="text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.adminApiKey.notConfigured") }}
                </span>
                <button
                  type="button"
                  @click="createAdminApiKey"
                  :disabled="adminApiKeyOperating"
                  class="btn btn-primary btn-sm h-9"
                >
                  <Icon
                    name="loader"
                    size="sm"
                    :animate-on-hover="false"
                    v-if="adminApiKeyOperating"
                    class="mr-1 h-4 w-4 animate-spin"
                  />
                  {{
                    adminApiKeyOperating
                      ? t("admin.settings.adminApiKey.creating")
                      : t("admin.settings.adminApiKey.create")
                  }}
                </button>
              </div>

              <!-- Key Exists -->
              <div v-else class="space-y-4">
                <div class="flex items-center justify-between">
                  <div>
                    <label
                      class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.adminApiKey.currentKey") }}
                    </label>
                    <code
                      class="rounded-compact bg-gray-100 px-2 py-1 font-mono text-sm text-gray-900 dark:bg-dark-700 dark:text-gray-100"
                    >
                      {{ adminApiKeyMasked }}
                    </code>
                  </div>
                  <div class="flex gap-2">
                    <button
                      type="button"
                      @click="regenerateAdminApiKey"
                      :disabled="adminApiKeyOperating"
                      class="btn btn-secondary btn-sm h-9"
                    >
                      {{
                        adminApiKeyOperating
                          ? t("admin.settings.adminApiKey.regenerating")
                          : t("admin.settings.adminApiKey.regenerate")
                      }}
                    </button>
                    <button
                      type="button"
                      @click="deleteAdminApiKey"
                      :disabled="adminApiKeyOperating"
                      class="btn btn-secondary btn-sm h-9 text-red-600 hover:text-red-700 dark:text-red-400"
                    >
                      {{ t("admin.settings.adminApiKey.delete") }}
                    </button>
                  </div>
                </div>

                <!-- Newly Generated Key Display -->
                <div
                  v-if="newAdminApiKey"
                  class="space-y-3 rounded-control border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20"
                >
                  <p
                    class="text-sm font-medium text-green-700 dark:text-green-300"
                  >
                    {{ t("admin.settings.adminApiKey.keyWarning") }}
                  </p>
                  <div class="flex items-center gap-2">
                    <code
                      class="flex-1 select-all break-all rounded-compact border border-green-300 bg-white px-3 py-2 font-mono text-sm dark:border-green-700 dark:bg-dark-800"
                    >
                      {{ newAdminApiKey }}
                    </code>
                    <button
                      type="button"
                      @click="copyNewKey"
                      class="btn btn-primary btn-sm h-9 flex-shrink-0"
                    >
                      {{ t("admin.settings.adminApiKey.copyKey") }}
                    </button>
                  </div>
                  <p class="text-xs text-green-600 dark:text-green-400">
                    {{ t("admin.settings.adminApiKey.usage") }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
        <!-- /Tab: Security — Admin API Key -->

        <!-- Tab: Gateway -->
        <div
          v-show="activeTab === 'gateway'" v-content-reveal="activeTab === 'gateway'"
          ref="gatewayContentStartRef"
          class="gateway-settings-content space-y-4"
        >
          <!-- 过载冷却（529） -->
          <SettingsCard
            v-show="activeGatewaySection === 'general'" v-content-reveal="activeGatewaySection === 'general'"
            data-testid="gateway-card-overload-cooldown"
            :title="t('admin.settings.overloadCooldown.title')"
            :description="t('admin.settings.overloadCooldown.description')"
          >
            <ContentSkeleton v-if="overloadCooldownLoading" variant="form" :rows="3" />
            <SettingsSection v-else>
              <SettingToggleRow
                id="overload-cooldown-enabled"
                v-model="overloadCooldownForm.enabled"
                :label="t('admin.settings.overloadCooldown.enabled')"
                :hint="t('admin.settings.overloadCooldown.enabledHint')"
              />
              <Collapse :open="overloadCooldownForm.enabled" unmount-on-hide>
                <SettingsSubpanel>
                  <SettingRow
                    id="overload-cooldown-minutes"
                    field
                    label-for="overload-cooldown-minutes"
                    :label="t('admin.settings.overloadCooldown.cooldownMinutes')"
                    :hint="t('admin.settings.overloadCooldown.cooldownMinutesHint')"
                  >
                    <input
                      id="overload-cooldown-minutes"
                      v-model.number="overloadCooldownForm.cooldown_minutes"
                      type="number"
                      min="1"
                      max="120"
                      class="input"
                    />
                  </SettingRow>
                </SettingsSubpanel>
              </Collapse>
            </SettingsSection>
            <template v-if="!overloadCooldownLoading" #footer>
              <button
                type="button"
                :disabled="overloadCooldownSaving"
                class="btn btn-primary btn-sm h-9"
                @click="saveOverloadCooldownSettings"
              >
                <Icon
                  v-if="overloadCooldownSaving"
                  name="loader"
                  size="sm"
                  :animate-on-hover="false"
                  class="mr-1 h-4 w-4 animate-spin"
                />
                {{ overloadCooldownSaving ? t("common.saving") : t("common.save") }}
              </button>
            </template>
          </SettingsCard>

          <!-- OpenAI OAuth 403 冷却 -->
          <SettingsCard
            v-show="activeGatewaySection === 'openai'" v-content-reveal="activeGatewaySection === 'openai'"
            data-testid="gateway-card-openai-403-cooldown"
            :title="t('admin.settings.openAI403Cooldown.title')"
            :description="t('admin.settings.openAI403Cooldown.description')"
          >
            <ContentSkeleton v-if="openAI403CooldownLoading" variant="form" :rows="3" />
            <SettingsSection v-else>
              <SettingToggleRow
                id="openai-403-cooldown-enabled"
                v-model="openAI403CooldownForm.enabled"
                :label="t('admin.settings.openAI403Cooldown.enabled')"
                :hint="t('admin.settings.openAI403Cooldown.enabledHint')"
              />
              <Collapse :open="openAI403CooldownForm.enabled" unmount-on-hide>
                <SettingsSubpanel>
                  <SettingRow
                    id="openai-403-cooldown-minutes"
                    field
                    label-for="openai-403-cooldown-minutes"
                    :label="t('admin.settings.openAI403Cooldown.cooldownMinutes')"
                    :hint="t('admin.settings.openAI403Cooldown.cooldownMinutesHint')"
                  >
                    <input
                      id="openai-403-cooldown-minutes"
                      v-model.number="openAI403CooldownForm.cooldown_minutes"
                      type="number"
                      min="1"
                      max="120"
                      class="input"
                    />
                  </SettingRow>
                  <SettingToggleRow
                    id="openai-403-error-on-threshold"
                    v-model="openAI403CooldownForm.error_on_threshold_enabled"
                    :label="t('admin.settings.openAI403Cooldown.errorOnThresholdEnabled')"
                    :hint="t('admin.settings.openAI403Cooldown.errorOnThresholdEnabledHint')"
                  />
                  <template v-if="openAI403CooldownForm.error_on_threshold_enabled">
                    <SettingRow
                      id="openai-403-threshold-count"
                      field
                      label-for="openai-403-threshold-count"
                      :label="t('admin.settings.openAI403Cooldown.thresholdCount')"
                      :hint="t('admin.settings.openAI403Cooldown.thresholdCountHint')"
                    >
                      <input
                        id="openai-403-threshold-count"
                        v-model.number="openAI403CooldownForm.threshold_count"
                        type="number"
                        min="1"
                        max="20"
                        class="input"
                      />
                    </SettingRow>
                    <SettingRow
                      id="openai-403-threshold-window"
                      field
                      label-for="openai-403-threshold-window"
                      :label="t('admin.settings.openAI403Cooldown.thresholdWindowMinutes')"
                      :hint="t('admin.settings.openAI403Cooldown.thresholdWindowMinutesHint')"
                    >
                      <input
                        id="openai-403-threshold-window"
                        v-model.number="openAI403CooldownForm.threshold_window_minutes"
                        type="number"
                        min="1"
                        max="1440"
                        class="input"
                      />
                    </SettingRow>
                  </template>
                </SettingsSubpanel>
              </Collapse>
            </SettingsSection>
            <template v-if="!openAI403CooldownLoading" #footer>
              <button
                type="button"
                :disabled="openAI403CooldownSaving"
                class="btn btn-primary btn-sm h-9"
                @click="saveOpenAI403CooldownSettings"
              >
                <Icon
                  v-if="openAI403CooldownSaving"
                  name="loader"
                  size="sm"
                  :animate-on-hover="false"
                  class="mr-1 h-4 w-4 animate-spin"
                />
                {{ openAI403CooldownSaving ? t("common.saving") : t("common.save") }}
              </button>
            </template>
          </SettingsCard>

          <div
            v-show="activeGatewaySection === 'openai'" v-content-reveal="activeGatewaySection === 'openai'"
            data-testid="gateway-card-openai-oauth-defaults"
          >
            <OpenAIOAuthImportDefaultsSettings />
          </div>

          <!-- 限流冷却（429） -->
          <SettingsCard
            v-show="activeGatewaySection === 'general'" v-content-reveal="activeGatewaySection === 'general'"
            data-testid="gateway-card-rate-limit-cooldown"
            :title="t('admin.settings.rateLimit429Cooldown.title')"
            :description="t('admin.settings.rateLimit429Cooldown.description')"
          >
            <ContentSkeleton v-if="rateLimit429CooldownLoading" variant="form" :rows="3" />
            <SettingsSection v-else>
              <SettingToggleRow
                id="rate-limit-429-cooldown-enabled"
                v-model="rateLimit429CooldownForm.enabled"
                :label="t('admin.settings.rateLimit429Cooldown.enabled')"
                :hint="t('admin.settings.rateLimit429Cooldown.enabledHint')"
              />
              <Collapse :open="rateLimit429CooldownForm.enabled" unmount-on-hide>
                <SettingsSubpanel>
                  <SettingRow
                    id="rate-limit-429-cooldown-seconds"
                    field
                    label-for="rate-limit-429-cooldown-seconds"
                    :label="t('admin.settings.rateLimit429Cooldown.cooldownSeconds')"
                    :hint="t('admin.settings.rateLimit429Cooldown.cooldownSecondsHint')"
                  >
                    <input
                      id="rate-limit-429-cooldown-seconds"
                      v-model.number="rateLimit429CooldownForm.cooldown_seconds"
                      type="number"
                      min="1"
                      max="7200"
                      class="input"
                    />
                  </SettingRow>
                </SettingsSubpanel>
              </Collapse>
            </SettingsSection>
            <template v-if="!rateLimit429CooldownLoading" #footer>
              <button
                type="button"
                :disabled="rateLimit429CooldownSaving"
                class="btn btn-primary btn-sm h-9"
                @click="saveRateLimit429CooldownSettings"
              >
                <Icon
                  v-if="rateLimit429CooldownSaving"
                  name="loader"
                  size="sm"
                  :animate-on-hover="false"
                  class="mr-1 h-4 w-4 animate-spin"
                />
                {{ rateLimit429CooldownSaving ? t("common.saving") : t("common.save") }}
              </button>
            </template>
          </SettingsCard>

          <!-- 流超时 -->
          <SettingsCard
            v-show="activeGatewaySection === 'general'" v-content-reveal="activeGatewaySection === 'general'"
            data-testid="gateway-card-stream-timeout"
            :title="t('admin.settings.streamTimeout.title')"
            :description="t('admin.settings.streamTimeout.description')"
          >
            <ContentSkeleton v-if="streamTimeoutLoading" variant="form" :rows="4" />
            <SettingsSection v-else>
              <SettingToggleRow
                id="stream-timeout-enabled"
                v-model="streamTimeoutForm.enabled"
                :label="t('admin.settings.streamTimeout.enabled')"
                :hint="t('admin.settings.streamTimeout.enabledHint')"
              />
              <Collapse :open="streamTimeoutForm.enabled" unmount-on-hide>
                <SettingsSubpanel>
                  <SettingRow
                    id="stream-timeout-action"
                    field
                    :label="t('admin.settings.streamTimeout.action')"
                    :hint="t('admin.settings.streamTimeout.actionHint')"
                  >
                    <Select
                      v-model="streamTimeoutForm.action"
                      :options="streamTimeoutActionOptions"
                      :aria-label="t('admin.settings.streamTimeout.action')"
                    />
                  </SettingRow>
                  <!-- 临时停止调度的时长只在动作为 temp_unsched 时生效。 -->
                  <SettingRow
                    v-if="streamTimeoutForm.action === 'temp_unsched'"
                    id="stream-timeout-temp-unsched-minutes"
                    field
                    label-for="stream-timeout-temp-unsched-minutes"
                    :label="t('admin.settings.streamTimeout.tempUnschedMinutes')"
                    :hint="t('admin.settings.streamTimeout.tempUnschedMinutesHint')"
                  >
                    <input
                      id="stream-timeout-temp-unsched-minutes"
                      v-model.number="streamTimeoutForm.temp_unsched_minutes"
                      type="number"
                      min="1"
                      max="60"
                      class="input"
                    />
                  </SettingRow>
                  <SettingRow
                    id="stream-timeout-threshold-count"
                    field
                    label-for="stream-timeout-threshold-count"
                    :label="t('admin.settings.streamTimeout.thresholdCount')"
                    :hint="t('admin.settings.streamTimeout.thresholdCountHint')"
                  >
                    <input
                      id="stream-timeout-threshold-count"
                      v-model.number="streamTimeoutForm.threshold_count"
                      type="number"
                      min="1"
                      max="10"
                      class="input"
                    />
                  </SettingRow>
                  <SettingRow
                    id="stream-timeout-threshold-window"
                    field
                    label-for="stream-timeout-threshold-window"
                    :label="t('admin.settings.streamTimeout.thresholdWindowMinutes')"
                    :hint="t('admin.settings.streamTimeout.thresholdWindowMinutesHint')"
                  >
                    <input
                      id="stream-timeout-threshold-window"
                      v-model.number="streamTimeoutForm.threshold_window_minutes"
                      type="number"
                      min="1"
                      max="60"
                      class="input"
                    />
                  </SettingRow>
                </SettingsSubpanel>
              </Collapse>
            </SettingsSection>
            <template v-if="!streamTimeoutLoading" #footer>
              <button
                type="button"
                :disabled="streamTimeoutSaving"
                class="btn btn-primary btn-sm h-9"
                @click="saveStreamTimeoutSettings"
              >
                <Icon
                  v-if="streamTimeoutSaving"
                  name="loader"
                  size="sm"
                  :animate-on-hover="false"
                  class="mr-1 h-4 w-4 animate-spin"
                />
                {{ streamTimeoutSaving ? t("common.saving") : t("common.save") }}
              </button>
            </template>
          </SettingsCard>

          <!-- 跨平台请求整流器 -->
          <SettingsCard
            v-show="activeGatewaySection === 'general'" v-content-reveal="activeGatewaySection === 'general'"
            data-testid="gateway-card-request-rectifier"
            :title="t('admin.settings.rectifier.title')"
            :description="t('admin.settings.rectifier.description')"
          >
            <ContentSkeleton v-if="rectifierLoading" variant="form" :rows="4" />
            <SettingsSection v-else>
              <SettingToggleRow
                id="rectifier-enabled"
                v-model="rectifierForm.enabled"
                :label="t('admin.settings.rectifier.enabled')"
                :hint="t('admin.settings.rectifier.enabledHint')"
              />
              <Collapse :open="rectifierForm.enabled" unmount-on-hide>
                <SettingsSubpanel>
                  <SettingToggleRow
                    id="rectifier-thinking-signature"
                    v-model="rectifierForm.thinking_signature_enabled"
                    :label="t('admin.settings.rectifier.thinkingSignature')"
                    :hint="t('admin.settings.rectifier.thinkingSignatureHint')"
                  />
                  <SettingToggleRow
                    id="rectifier-thinking-budget"
                    v-model="rectifierForm.thinking_budget_enabled"
                    :label="t('admin.settings.rectifier.thinkingBudget')"
                    :hint="t('admin.settings.rectifier.thinkingBudgetHint')"
                  />
                  <SettingToggleRow
                    id="rectifier-apikey-signature"
                    v-model="rectifierForm.apikey_signature_enabled"
                    :label="t('admin.settings.rectifier.apikeySignature')"
                    :hint="t('admin.settings.rectifier.apikeySignatureHint')"
                  />
                  <!-- 自定义匹配规则只在 API Key 签名整流开启时生效。 -->
                  <RuleListEditor
                    v-if="rectifierForm.apikey_signature_enabled"
                    :items="rectifierForm.apikey_signature_patterns"
                    :title="t('admin.settings.rectifier.apikeyPatterns')"
                    :hint="t('admin.settings.rectifier.apikeyPatternsHint')"
                    :add-label="t('admin.settings.rectifier.addPattern')"
                    :animated="false"
                    test-id="rectifier-patterns"
                    @add="rectifierForm.apikey_signature_patterns.push('')"
                    @remove="rectifierForm.apikey_signature_patterns.splice($event, 1)"
                  >
                    <template #row="{ index }">
                      <input
                        v-model="rectifierForm.apikey_signature_patterns[index]"
                        type="text"
                        class="input"
                        :placeholder="t('admin.settings.rectifier.apikeyPatternPlaceholder')"
                      />
                    </template>
                  </RuleListEditor>
                </SettingsSubpanel>
              </Collapse>
            </SettingsSection>
            <template v-if="!rectifierLoading" #footer>
              <button
                type="button"
                :disabled="rectifierSaving"
                class="btn btn-primary btn-sm h-9"
                @click="saveRectifierSettings"
              >
                <Icon
                  v-if="rectifierSaving"
                  name="loader"
                  size="sm"
                  :animate-on-hover="false"
                  class="mr-1 h-4 w-4 animate-spin"
                />
                {{ rectifierSaving ? t("common.saving") : t("common.save") }}
              </button>
            </template>
          </SettingsCard>

          <!-- Anthropic Beta 策略 -->
          <SettingsCard
            v-show="activeGatewaySection === 'anthropic'" v-content-reveal="activeGatewaySection === 'anthropic'"
            data-testid="gateway-card-beta-policy"
            :title="t('admin.settings.betaPolicy.title')"
            :description="t('admin.settings.betaPolicy.description')"
          >
            <ContentSkeleton v-if="betaPolicyLoading" variant="form" :rows="5" />
            <SettingsSection v-else>
              <!-- 每个 beta token 一条固定规则。 -->
              <SettingsSubpanel
                v-for="rule in betaPolicyForm.rules"
                :key="rule.beta_token"
              >
                <div class="flex items-center gap-2">
                  <span class="text-sm font-semibold text-primary-900 dark:text-dark-50">
                    {{ getBetaDisplayName(rule.beta_token) }}
                  </span>
                  <span class="rounded-compact bg-gray-100 px-2 py-0.5 font-mono text-xs text-gray-500 dark:bg-dark-700 dark:text-dark-300">
                    {{ rule.beta_token }}
                  </span>
                </div>

                <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <div>
                    <label :for="`beta-policy-${rule.beta_token}-action`" class="input-label">
                      {{ t("admin.settings.betaPolicy.action") }}
                    </label>
                    <Select
                      :id="`beta-policy-${rule.beta_token}-action`"
                      :modelValue="rule.action"
                      @update:modelValue="rule.action = $event as any"
                      :options="betaPolicyActionOptions"
                    />
                  </div>
                  <div>
                    <label :for="`beta-policy-${rule.beta_token}-scope`" class="input-label">
                      {{ t("admin.settings.betaPolicy.scope") }}
                    </label>
                    <Select
                      :id="`beta-policy-${rule.beta_token}-scope`"
                      :modelValue="rule.scope"
                      @update:modelValue="rule.scope = $event as any"
                      :options="betaPolicyScopeOptions"
                    />
                  </div>
                </div>

                <!-- 动作为拦截时填写返回给客户端的错误信息。 -->
                <div v-if="rule.action === 'block'">
                  <label :for="`beta-policy-${rule.beta_token}-error`" class="input-label">
                    {{ t("admin.settings.betaPolicy.errorMessage") }}
                  </label>
                  <input
                    :id="`beta-policy-${rule.beta_token}-error`"
                    v-model="rule.error_message"
                    type="text"
                    class="input"
                    :placeholder="t('admin.settings.betaPolicy.errorMessagePlaceholder')"
                  />
                  <p class="input-hint">
                    {{ t("admin.settings.betaPolicy.errorMessageHint") }}
                  </p>
                </div>

                <div v-if="betaPresets[rule.beta_token]?.length">
                  <p class="input-label">
                    {{ t("admin.settings.betaPolicy.quickPresets") }}
                  </p>
                  <div class="flex flex-wrap gap-2">
                    <button
                      v-for="preset in betaPresets[rule.beta_token]"
                      :key="preset.label"
                      type="button"
                      class="inline-flex items-center gap-1 rounded-compact border border-primary-200 bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 transition-colors hover:bg-primary-100 dark:border-primary-800 dark:bg-primary-900/30 dark:text-primary-300 dark:hover:bg-primary-900/50"
                      @click="applyBetaPreset(rule, preset)"
                      :title="preset.description"
                    >
                      {{ preset.label }}
                    </button>
                  </div>
                </div>

                <RuleListEditor
                  :items="rule.model_whitelist || []"
                  :title="t('admin.settings.betaPolicy.modelWhitelist')"
                  :hint="t('admin.settings.betaPolicy.modelWhitelistHint')"
                  :add-label="t('admin.settings.betaPolicy.addModelPattern')"
                  :animated="false"
                  @add="if (!rule.model_whitelist) rule.model_whitelist = []; rule.model_whitelist.push('');"
                  @remove="rule.model_whitelist!.splice($event, 1)"
                >
                  <template #footer>
                    <div class="flex flex-wrap items-center gap-1.5">
                      <span class="text-xs text-gray-500 dark:text-dark-400"
                        >{{ t("admin.settings.betaPolicy.commonPatterns") }}:</span
                      >
                      <button
                        v-for="pattern in commonModelPatterns"
                        :key="pattern"
                        type="button"
                        class="rounded-compact border border-gray-200 px-2 py-0.5 text-xs text-gray-600 transition-colors hover:border-primary-300 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-600 dark:text-gray-400 dark:hover:border-primary-700 dark:hover:bg-primary-900/30 dark:hover:text-primary-300"
                        @click="addQuickPattern(rule, pattern)"
                      >
                        {{ pattern }}
                      </button>
                    </div>
                  </template>
                  <template #row="{ index }">
                    <input
                      v-model="rule.model_whitelist![index]"
                      type="text"
                      class="input"
                      :placeholder="t('admin.settings.betaPolicy.modelPatternPlaceholder')"
                    />
                  </template>
                </RuleListEditor>

                <!-- 模型白名单非空时，未命中白名单的模型按这里选择的动作处理。 -->
                <div v-if="rule.model_whitelist && rule.model_whitelist.length > 0">
                  <label :for="`beta-policy-${rule.beta_token}-fallback`" class="input-label">
                    {{ t("admin.settings.betaPolicy.fallbackAction") }}
                  </label>
                  <Select
                    :id="`beta-policy-${rule.beta_token}-fallback`"
                    :modelValue="rule.fallback_action || 'pass'"
                    @update:modelValue="rule.fallback_action = $event as any"
                    :options="betaPolicyActionOptions"
                  />
                  <p class="input-hint">
                    {{ t("admin.settings.betaPolicy.fallbackActionHint") }}
                  </p>
                  <div v-if="rule.fallback_action === 'block'" class="mt-2">
                    <input
                      v-model="rule.fallback_error_message"
                      type="text"
                      class="input"
                      :aria-label="t('admin.settings.betaPolicy.errorMessage')"
                      :placeholder="t('admin.settings.betaPolicy.fallbackErrorMessagePlaceholder')"
                    />
                    <p class="input-hint">
                      {{ t("admin.settings.betaPolicy.errorMessageHint") }}
                    </p>
                  </div>
                </div>
              </SettingsSubpanel>
            </SettingsSection>
            <template v-if="!betaPolicyLoading" #footer>
              <button
                type="button"
                :disabled="betaPolicySaving"
                class="btn btn-primary btn-sm h-9"
                @click="saveBetaPolicySettings"
              >
                <Icon
                  v-if="betaPolicySaving"
                  name="loader"
                  size="sm"
                  :animate-on-hover="false"
                  class="mr-1 h-4 w-4 animate-spin"
                />
                {{ betaPolicySaving ? t("common.saving") : t("common.save") }}
              </button>
            </template>
          </SettingsCard>

          <!-- OpenAI Fast/Flex 策略，随全局保存提交。 -->
          <SettingsCard
            v-show="activeGatewaySection === 'openai'" v-content-reveal="activeGatewaySection === 'openai'"
            data-testid="gateway-card-openai-fast-policy"
            :title="t('admin.settings.openaiFastPolicy.title')"
            :description="t('admin.settings.openaiFastPolicy.description')"
          >
            <SettingsSection>
              <RuleListEditor
                :items="openaiFastPolicyForm.rules"
                variant="card"
                :item-label="(index) => t('admin.settings.openaiFastPolicy.ruleHeader', { index: index + 1 })"
                :add-label="t('admin.settings.openaiFastPolicy.addRule')"
                :remove-label="t('admin.settings.openaiFastPolicy.removeRule')"
                :empty-text="t('admin.settings.openaiFastPolicy.empty')"
                add-placement="footer"
                test-id="openai-fast-rules"
                @add="addOpenAIFastPolicyRule"
                @remove="removeOpenAIFastPolicyRule"
              >
                <template #row="{ item: rule, index: ruleIndex }">
                  <div class="space-y-4">
                    <div
                      class="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-gray-500 dark:text-dark-400"
                      :data-testid="`openai-fast-policy-summary-${ruleIndex}`"
                    >
                      <span class="font-medium text-gray-700 dark:text-dark-200">
                        {{
                          t(
                            hasOpenAIFastPolicyTargetModels(rule)
                              ? "admin.settings.openaiFastPolicy.summaryTargetModels"
                              : "admin.settings.openaiFastPolicy.summaryAllModels",
                          )
                        }}
                      </span>
                      <span aria-hidden="true">→</span>
                      <span
                        class="inline-flex items-center rounded-compact bg-primary-50 px-2 py-0.5 font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
                      >
                        {{ openaiFastPolicyActionSummary(rule.action) }}
                      </span>
                      <template v-if="hasOpenAIFastPolicyTargetModels(rule)">
                        <span aria-hidden="true">·</span>
                        <span class="font-medium text-gray-700 dark:text-dark-200">
                          {{ t("admin.settings.openaiFastPolicy.summaryOtherModels") }}
                        </span>
                        <span aria-hidden="true">→</span>
                        <span
                          class="inline-flex items-center rounded-compact bg-gray-100 px-2 py-0.5 font-medium text-gray-700 dark:bg-dark-600 dark:text-gray-300"
                        >
                          {{ openaiFastPolicyActionSummary(rule.fallback_action || "pass") }}
                        </span>
                      </template>
                    </div>
                    <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                      <div>
                        <label :for="`openai-fast-policy-tier-${ruleIndex}`" class="input-label">
                          {{ t("admin.settings.openaiFastPolicy.serviceTier") }}
                        </label>
                        <Select
                          :id="`openai-fast-policy-tier-${ruleIndex}`"
                          :modelValue="rule.service_tier"
                          @update:modelValue="
                            rule.service_tier = $event as
                              | 'all'
                              | 'priority'
                              | 'flex'
                          "
                          :options="openaiFastPolicyTierOptions"
                        />
                      </div>
                      <div>
                        <label :for="`openai-fast-policy-action-${ruleIndex}`" class="input-label">
                          {{ t("admin.settings.openaiFastPolicy.action") }}
                        </label>
                        <Select
                          :id="`openai-fast-policy-action-${ruleIndex}`"
                          :modelValue="rule.action"
                          @update:modelValue="
                            rule.action = $event as
                              | 'pass'
                              | 'filter'
                              | 'block'
                              | 'force_priority'
                              | 'force_ultrafast'
                          "
                          :options="openaiFastPolicyActionOptions"
                        />
                      </div>
                      <div>
                        <label :for="`openai-fast-policy-scope-${ruleIndex}`" class="input-label">
                          {{ t("admin.settings.openaiFastPolicy.scope") }}
                        </label>
                        <Select
                          :id="`openai-fast-policy-scope-${ruleIndex}`"
                          :modelValue="rule.scope"
                          @update:modelValue="
                            rule.scope = $event as
                              | 'all'
                              | 'oauth'
                              | 'apikey'
                              | 'bedrock'
                          "
                          :options="openaiFastPolicyScopeOptions"
                        />
                      </div>
                    </div>
                    <div>
                      <p class="input-label">
                        {{ t("admin.settings.openaiFastPolicy.userIds") }}
                      </p>
                      <OpenAIFastPolicyUserSelector
                        :model-value="rule.user_ids || []"
                        @update:model-value="rule.user_ids = $event"
                      />
                      <p class="input-hint">
                        {{ t("admin.settings.openaiFastPolicy.userIdsHint") }}
                      </p>
                    </div>
                    <div v-if="rule.action === 'block'">
                      <label :for="`openai-fast-policy-error-${ruleIndex}`" class="input-label">
                        {{ t("admin.settings.openaiFastPolicy.errorMessage") }}
                      </label>
                      <input
                        :id="`openai-fast-policy-error-${ruleIndex}`"
                        v-model="rule.error_message"
                        type="text"
                        class="input"
                        :placeholder="t('admin.settings.openaiFastPolicy.errorMessagePlaceholder')"
                      />
                      <p class="input-hint">
                        {{ t("admin.settings.openaiFastPolicy.errorMessageHint") }}
                      </p>
                    </div>
                    <div
                      role="group"
                      :aria-labelledby="`openai-fast-policy-models-label-${ruleIndex}`"
                      :aria-describedby="`openai-fast-policy-models-hint-${ruleIndex}`"
                    >
                      <p
                        :id="`openai-fast-policy-models-label-${ruleIndex}`"
                        class="input-label mb-0"
                      >
                        {{ t("admin.settings.openaiFastPolicy.modelWhitelist") }}
                      </p>
                      <p
                        :id="`openai-fast-policy-models-hint-${ruleIndex}`"
                        class="input-hint mb-2"
                      >
                        {{ t("admin.settings.openaiFastPolicy.modelWhitelistHint") }}
                      </p>
                      <RuleListEditor
                        :items="rule.model_whitelist || []"
                        :add-label="t('admin.settings.openaiFastPolicy.addModelPattern')"
                        add-placement="footer"
                        :animated="false"
                        :test-id="`openai-fast-models-${ruleIndex}`"
                        @add="addOpenAIFastPolicyModelPattern(rule)"
                        @remove="removeOpenAIFastPolicyModelPattern(rule, $event)"
                      >
                        <template #row="{ index: patternIdx }">
                          <input
                            v-model="rule.model_whitelist![patternIdx]"
                            type="text"
                            class="input"
                            :placeholder="t('admin.settings.openaiFastPolicy.modelPatternPlaceholder')"
                          />
                        </template>
                      </RuleListEditor>
                    </div>
                    <div v-if="hasOpenAIFastPolicyTargetModels(rule)">
                      <label :for="`openai-fast-policy-fallback-${ruleIndex}`" class="input-label">
                        {{ t("admin.settings.openaiFastPolicy.fallbackAction") }}
                      </label>
                      <Select
                        :id="`openai-fast-policy-fallback-${ruleIndex}`"
                        :modelValue="rule.fallback_action || 'pass'"
                        @update:modelValue="
                          rule.fallback_action = $event as
                            | 'pass'
                            | 'filter'
                            | 'block'
                            | 'force_priority'
                            | 'force_ultrafast'
                        "
                        :options="openaiFastPolicyActionOptions"
                      />
                      <p class="input-hint">
                        {{ t("admin.settings.openaiFastPolicy.fallbackActionHint") }}
                      </p>
                      <div v-if="rule.fallback_action === 'block'" class="mt-2">
                        <input
                          v-model="rule.fallback_error_message"
                          type="text"
                          class="input"
                          :aria-label="t('admin.settings.openaiFastPolicy.errorMessage')"
                          :placeholder="t('admin.settings.openaiFastPolicy.fallbackErrorMessagePlaceholder')"
                        />
                      </div>
                    </div>
                  </div>
                </template>
              </RuleListEditor>
              <p class="input-hint">{{ t('admin.settings.openaiFastPolicy.saveHint') }}</p>
            </SettingsSection>
          </SettingsCard>
        </div>
        <!-- /Tab: Gateway -->

        <!-- Tab: Security — Registration, Turnstile, LinuxDo -->
        <div v-show="activeTab === 'security'" v-content-reveal="activeTab === 'security'" class="space-y-4">
          <!-- Registration Settings -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.registration.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.registration.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <!-- Enable Registration -->
              <div class="flex items-center justify-between">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.registration.enableRegistration")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{
                      t("admin.settings.registration.enableRegistrationHint")
                    }}
                  </p>
                </div>
                <Toggle v-model="form.registration_enabled" />
              </div>

              <!-- Email Verification -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.registration.emailVerification")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.registration.emailVerificationHint") }}
                  </p>
                </div>
                <Toggle v-model="form.email_verify_enabled" />
              </div>

              <!-- Email Suffix Whitelist -->
              <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
                <label class="font-medium text-gray-900 dark:text-white">{{
                  t("admin.settings.registration.emailSuffixWhitelist")
                }}</label>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{
                    t("admin.settings.registration.emailSuffixWhitelistHint")
                  }}
                </p>
                <div
                  class="mt-3 rounded-control border border-gray-300 bg-white p-2 dark:border-dark-500 dark:bg-dark-700"
                >
                  <div class="flex flex-wrap items-center gap-2">
                    <span
                      v-for="suffix in registrationEmailSuffixWhitelistTags"
                      :key="suffix"
                      class="inline-flex items-center gap-1 rounded-compact bg-gray-100 px-2 py-1 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-200"
                    >
                      <span>{{ suffix }}</span>
                      <button
                        type="button"
                        class="rounded-full text-gray-500 hover:bg-gray-200 hover:text-gray-700 dark:text-gray-300 dark:hover:bg-dark-500 dark:hover:text-white"
                        @click="
                          removeRegistrationEmailSuffixWhitelistTag(suffix)
                        "
                      >
                        <Icon
                          name="x"
                          size="xs"
                          class="h-3.5 w-3.5"
                          :stroke-width="2"
                        />
                      </button>
                    </span>

                    <div
                      class="flex min-w-[220px] flex-1 items-center gap-1 rounded-compact border border-transparent px-2 py-1 focus-within:border-primary-900/10 focus-within:ring-2 focus-within:ring-black/10 dark:focus-within:border-primary-700 dark:focus-within:ring-0"
                    >
                      <input
                        v-model="registrationEmailSuffixWhitelistDraft"
                        type="text"
                        class="w-full bg-transparent text-sm font-mono text-gray-900 outline-none placeholder:text-gray-400 dark:text-white dark:placeholder:text-gray-500"
                        :placeholder="
                          t(
                            'admin.settings.registration.emailSuffixWhitelistPlaceholder',
                          )
                        "
                        @input="
                          handleRegistrationEmailSuffixWhitelistDraftInput
                        "
                        @keydown="
                          handleRegistrationEmailSuffixWhitelistDraftKeydown
                        "
                        @blur="commitRegistrationEmailSuffixWhitelistDraft"
                        @paste="handleRegistrationEmailSuffixWhitelistPaste"
                      />
                    </div>
                  </div>
                </div>
                <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                  {{
                    t(
                      "admin.settings.registration.emailSuffixWhitelistInputHint",
                    )
                  }}
                </p>
              </div>

              <!-- 非白名单邮箱域名额度 -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.registration.emailDomainQuota")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.registration.emailDomainQuotaHint") }}
                  </p>
                </div>
                <Toggle
                  v-model="form.registration_email_domain_quota_enabled"
                />
              </div>

              <!-- 邮箱地址归一化 -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.registration.emailNormalization")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{
                      t("admin.settings.registration.emailNormalizationHint")
                    }}
                  </p>
                </div>
                <Toggle v-model="form.registration_email_normalization" />
              </div>

              <!-- 用户换绑主邮箱 -->
              <div
                data-testid="user-email-change-setting"
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.registration.userEmailChange")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.registration.userEmailChangeHint") }}
                  </p>
                </div>
                <Toggle v-model="form.user_email_change_enabled" />
              </div>

              <!-- Promo Code -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.registration.promoCode")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.registration.promoCodeHint") }}
                  </p>
                </div>
                <Toggle v-model="form.promo_code_enabled" />
              </div>

              <!-- Invitation Code -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.registration.invitationCode")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.registration.invitationCodeHint") }}
                  </p>
                </div>
                <Toggle v-model="form.invitation_code_enabled" />
              </div>
              <!-- Password Reset - Only show when email verification is enabled -->
              <Collapse :open="form.email_verify_enabled" unmount-on-hide>
                <div

                  class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
                >
                  <div>
                    <label class="font-medium text-gray-900 dark:text-white">{{
                      t("admin.settings.registration.passwordReset")
                    }}</label>
                    <p class="text-sm text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.registration.passwordResetHint") }}
                    </p>
                  </div>
                  <Toggle v-model="form.password_reset_enabled" />
                </div>
              </Collapse>
              <!-- 前端地址同时用于密码重置、团队邀请等外部邮件链接，必须始终可配置。 -->
              <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
                <label
                  class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  {{ t("admin.settings.registration.frontendUrl") }}
                </label>
                <input
                  v-model="form.frontend_url"
                  data-testid="frontend-url-input"
                  type="url"
                  class="input"
                  :placeholder="
                    t('admin.settings.registration.frontendUrlPlaceholder')
                  "
                />
                <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.registration.frontendUrlHint") }}
                </p>
              </div>

              <!-- TOTP 2FA -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.registration.totp")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.registration.totpHint") }}
                  </p>
                  <!-- 未配置密钥时只阻止启用，保留关闭已有开关的能力 -->
                  <p
                    v-if="!form.totp_encryption_key_configured"
                    class="mt-2 text-sm text-amber-600 dark:text-amber-400"
                  >
                    {{ t("admin.settings.registration.totpKeyNotConfigured") }}
                  </p>
                </div>
                <Toggle
                  v-model="form.totp_enabled"
                  :disabled="
                    !form.totp_encryption_key_configured && !form.totp_enabled
                  "
                />
              </div>

              <!-- 敏感操作 step-up 2FA -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.security.stepUp")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.security.stepUpHint") }}
                  </p>
                </div>
                <Toggle v-model="form.step_up_enabled" />
              </div>

              <!-- 会话 IP/UA 绑定 -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.security.sessionBinding")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.security.sessionBindingHint") }}
                  </p>
                </div>
                <Toggle v-model="form.session_binding_enabled" />
              </div>

              <!-- 审计日志保留天数 -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.security.auditRetention")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.security.auditRetentionHint") }}
                  </p>
                </div>
                <input
                  v-model.number="form.audit_log_retention_days"
                  type="number"
                  min="0"
                  class="input w-28 text-right"
                />
              </div>
            </div>
          </div>

          <!-- API Key IP ACL Settings -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.apiKeyAcl.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.apiKeyAcl.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">
                    {{ t("admin.settings.apiKeyAcl.trustForwardedIp") }}
                  </label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.apiKeyAcl.trustForwardedIpHint") }}
                  </p>
                </div>
                <Toggle v-model="form.api_key_acl_trust_forwarded_ip" />
              </div>

              <Collapse :open="form.api_key_acl_trust_forwarded_ip" unmount-on-hide>
                <div

                  class="border-t border-gray-100 pt-4 dark:border-dark-700"
                >
                  <label
                    for="forwarded-client-ip-headers"
                    class="font-medium text-gray-900 dark:text-white"
                  >
                    {{ t("admin.settings.apiKeyAcl.forwardedClientIpHeaders") }}
                  </label>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.apiKeyAcl.forwardedClientIpHeadersHint") }}
                  </p>
                  <div
                    class="mt-3 rounded-control border border-gray-300 bg-white p-2 dark:border-dark-500 dark:bg-dark-700"
                  >
                    <div class="flex flex-wrap items-center gap-2">
                      <span
                        v-for="header in form.forwarded_client_ip_headers"
                        :key="header"
                        data-testid="forwarded-client-ip-header-tag"
                        class="inline-flex items-center gap-1 rounded-compact bg-gray-100 px-2 py-1 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-200"
                      >
                        <span>{{ header }}</span>
                        <button
                          type="button"
                          class="rounded-full text-gray-500 hover:bg-gray-200 hover:text-gray-700 dark:text-gray-300 dark:hover:bg-dark-500 dark:hover:text-white"
                          :aria-label="t('admin.settings.apiKeyAcl.removeForwardedClientIpHeader', { header })"
                          @click="removeForwardedClientIpHeader(header)"
                        >
                          <Icon
                            name="x"
                            size="xs"
                            class="h-3.5 w-3.5"
                            :stroke-width="2"
                          />
                        </button>
                      </span>
                      <div
                        class="flex min-w-[220px] flex-1 items-center gap-1 rounded-compact border border-transparent px-2 py-1 focus-within:border-primary-900/10 focus-within:ring-2 focus-within:ring-black/10 dark:focus-within:border-primary-700 dark:focus-within:ring-0"
                      >
                        <input
                          id="forwarded-client-ip-headers"
                          v-model="forwardedClientIpHeaderDraft"
                          data-testid="forwarded-client-ip-headers-input"
                          type="text"
                          class="w-full bg-transparent text-sm font-mono text-gray-900 outline-none placeholder:text-gray-400 dark:text-white dark:placeholder:text-gray-500"
                          :placeholder="t('admin.settings.apiKeyAcl.forwardedClientIpHeadersPlaceholder')"
                          @keydown="handleForwardedClientIpHeaderKeydown"
                          @blur="commitForwardedClientIpHeaderDraft"
                          @paste="handleForwardedClientIpHeaderPaste"
                        />
                      </div>
                    </div>
                  </div>
                  <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.apiKeyAcl.forwardedClientIpHeadersRiskHint") }}
                  </p>
                </div>
              </Collapse>
            </div>
          </div>

          <!-- 面板 API 限流设置 -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <div class="flex items-center gap-2">
                <Icon
                  name="shield"
                  size="md"
                  class="text-primary-500"
                />
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t("admin.settings.panelRateLimit.title") }}
                </h2>
              </div>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.panelRateLimit.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <ContentSkeleton v-if="panelRateLimitLoading" variant="form" :rows="4" />

              <template v-else>
                <!-- 计数维度说明：按用户 ID 计数，反代部署无误伤 -->
                <div
                  class="rounded-control border border-sky-200 bg-sky-50 p-4 dark:border-sky-800 dark:bg-sky-900/20"
                >
                  <div class="flex items-start">
                    <Icon
                      name="infoCircle"
                      size="md"
                      class="mt-0.5 flex-shrink-0 text-sky-500"
                    />
                    <p class="ml-3 text-sm text-sky-700 dark:text-sky-300">
                      {{ t("admin.settings.panelRateLimit.proxySafeNote") }}
                    </p>
                  </div>
                </div>

                <div class="flex items-center justify-between">
                  <div>
                    <label class="font-medium text-gray-900 dark:text-white">{{
                      t("admin.settings.panelRateLimit.enabled")
                    }}</label>
                    <p class="text-sm text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.panelRateLimit.enabledHint") }}
                    </p>
                  </div>
                  <Toggle v-model="panelRateLimitForm.enabled" />
                </div>

                <Collapse :open="panelRateLimitForm.enabled" unmount-on-hide>
                  <div

                    class="space-y-5 border-t border-gray-100 pt-4 dark:border-dark-700"
                  >
                    <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
                      <div>
                        <label
                          class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                        >
                          {{ t("admin.settings.panelRateLimit.userRpm") }}
                        </label>
                        <div class="flex items-center gap-2">
                          <input
                            v-model.number="panelRateLimitForm.user_rpm"
                            data-testid="panel-rate-limit-user-rpm"
                            type="number"
                            min="0"
                            max="100000"
                            class="input w-32"
                          />
                          <span class="text-sm text-gray-500 dark:text-gray-400">
                            {{ t("admin.settings.panelRateLimit.perMinute") }}
                          </span>
                        </div>
                        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                          {{ t("admin.settings.panelRateLimit.userRpmHint") }}
                        </p>
                      </div>

                      <div>
                        <label
                          class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                        >
                          {{ t("admin.settings.panelRateLimit.heavyRpm") }}
                        </label>
                        <div class="flex items-center gap-2">
                          <input
                            v-model.number="panelRateLimitForm.heavy_rpm"
                            type="number"
                            min="0"
                            max="100000"
                            class="input w-32"
                          />
                          <span class="text-sm text-gray-500 dark:text-gray-400">
                            {{ t("admin.settings.panelRateLimit.perMinute") }}
                          </span>
                        </div>
                        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                          {{ t("admin.settings.panelRateLimit.heavyRpmHint") }}
                        </p>
                      </div>

                      <div>
                        <label
                          class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                        >
                          {{ t("admin.settings.panelRateLimit.publicIpRpm") }}
                        </label>
                        <div class="flex items-center gap-2">
                          <input
                            v-model.number="panelRateLimitForm.public_ip_rpm"
                            type="number"
                            min="0"
                            max="100000"
                            class="input w-32"
                          />
                          <span class="text-sm text-gray-500 dark:text-gray-400">
                            {{ t("admin.settings.panelRateLimit.perMinute") }}
                          </span>
                        </div>
                        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                          {{ t("admin.settings.panelRateLimit.publicIpRpmHint") }}
                        </p>
                      </div>
                    </div>

                    <div
                      class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
                    >
                      <div>
                        <label class="font-medium text-gray-900 dark:text-white">{{
                          t("admin.settings.panelRateLimit.exemptAdmin")
                        }}</label>
                        <p class="text-sm text-gray-500 dark:text-gray-400">
                          {{ t("admin.settings.panelRateLimit.exemptAdminHint") }}
                        </p>
                      </div>
                      <Toggle v-model="panelRateLimitForm.exempt_admin" />
                    </div>
                  </div>
                </Collapse>

                <div
                  class="flex justify-end border-t border-gray-100 pt-4 dark:border-dark-700"
                >
                  <button
                    type="button"
                    data-testid="panel-rate-limit-save"
                    @click="savePanelRateLimitSettings"
                    :disabled="panelRateLimitSaving"
                    class="btn btn-primary btn-sm h-9"
                  >
                    <Icon
                      name="loader"
                      size="sm"
                      :animate-on-hover="false"
                      v-if="panelRateLimitSaving"
                      class="mr-1 h-4 w-4 animate-spin"
                    />
                    {{
                      panelRateLimitSaving
                        ? t("common.saving")
                        : t("common.save")
                    }}
                  </button>
                </div>
              </template>
            </div>
          </div>

          <!-- 人机验证设置 -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.captcha.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.captcha.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <!-- 人机验证总开关 -->
              <div class="flex items-center justify-between">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.captcha.enable")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.captcha.enableHint") }}
                  </p>
                </div>
                <Toggle
                  v-model="captchaMasterEnabled"
                  data-testid="captcha-enabled-toggle"
                />
              </div>

              <!-- 启用后显示服务商配置 -->
              <Collapse :open="captchaMasterEnabled" unmount-on-hide>
                <div

                  class="border-t border-gray-100 pt-4 dark:border-dark-700"
                >
                  <!-- 服务商选择 -->
                  <div class="mb-6">
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.captcha.provider") }}
                    </label>
                    <div
                      class="grid grid-cols-3 gap-2 rounded-control bg-gray-100 p-1 dark:bg-dark-700"
                    >
                      <button
                        type="button"
                        data-testid="captcha-provider-turnstile"
                        class="inline-flex items-center justify-center gap-2 rounded-control px-3 py-2 text-sm font-medium transition"
                        :class="
                          captchaProviderSelection === 'turnstile'
                            ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300'
                            : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                        "
                        @click="selectCaptchaProvider('turnstile')"
                      >
                        {{ t("admin.settings.captcha.providerTurnstile") }}
                      </button>
                      <button
                        type="button"
                        data-testid="captcha-provider-tencent"
                        class="inline-flex items-center justify-center gap-2 rounded-control px-3 py-2 text-sm font-medium transition"
                        :class="
                          captchaProviderSelection === 'tencent'
                            ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300'
                            : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                        "
                        @click="selectCaptchaProvider('tencent')"
                      >
                        {{ t("admin.settings.captcha.providerTencent") }}
                      </button>
                      <button
                        type="button"
                        data-testid="captcha-provider-aliyun"
                        class="inline-flex items-center justify-center gap-2 rounded-control px-3 py-2 text-sm font-medium transition"
                        :class="
                          captchaProviderSelection === 'aliyun'
                            ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300'
                            : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                        "
                        @click="selectCaptchaProvider('aliyun')"
                      >
                        {{ t("admin.settings.captcha.providerAliyun") }}
                      </button>
                    </div>
                  </div>

                  <!-- Cloudflare Turnstile 配置 -->
                  <div
                    v-if="captchaProviderSelection === 'turnstile'" v-content-reveal
                    class="grid grid-cols-1 gap-6"
                  >
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.turnstile.siteKey") }}
                      </label>
                      <input
                        v-model="form.turnstile_site_key"
                        type="text"
                        class="input font-mono text-sm"
                        placeholder="0x4AAAAAAA..."
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.turnstile.siteKeyHint") }}
                        <a
                          href="https://dash.cloudflare.com/"
                          target="_blank"
                          class="text-primary-600 hover:text-primary-500"
                          >{{
                            t("admin.settings.turnstile.cloudflareDashboard")
                          }}</a
                        >
                      </p>
                    </div>
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.turnstile.secretKey") }}
                      </label>
                      <input
                        v-model="form.turnstile_secret_key"
                        type="password"
                        class="input font-mono text-sm"
                        placeholder="0x4AAAAAAA..."
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{
                          form.turnstile_secret_key_configured
                            ? t(
                                "admin.settings.turnstile.secretKeyConfiguredHint",
                              )
                            : t("admin.settings.turnstile.secretKeyHint")
                        }}
                      </p>
                    </div>
                  </div>

                  <!-- 腾讯天御验证码配置 -->
                  <div v-else-if="captchaProviderSelection === 'tencent'">
                    <div class="mb-6 max-w-sm">
                      <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                        {{ t("admin.settings.tencentCaptcha.region") }}
                      </label>
                      <div class="grid grid-cols-2 gap-2 rounded-control bg-gray-100 p-1 dark:bg-dark-700">
                        <button
                          type="button"
                          data-testid="tencent-captcha-region-cn"
                          class="inline-flex items-center justify-center rounded-control px-3 py-1.5 text-sm font-medium transition"
                          :class="
                            form.tencent_captcha_region !== 'intl'
                              ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300'
                              : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                          "
                          @click="form.tencent_captcha_region = 'cn'"
                        >
                          {{ t("admin.settings.tencentCaptcha.regionCn") }}
                        </button>
                        <button
                          type="button"
                          data-testid="tencent-captcha-region-intl"
                          class="inline-flex items-center justify-center rounded-control px-3 py-1.5 text-sm font-medium transition"
                          :class="
                            form.tencent_captcha_region === 'intl'
                              ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300'
                              : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                          "
                          @click="form.tencent_captcha_region = 'intl'"
                        >
                          {{ t("admin.settings.tencentCaptcha.regionIntl") }}
                        </button>
                      </div>
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.tencentCaptcha.regionHint") }}
                      </p>
                    </div>
                    <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                      <div class="md:col-span-2">
                        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                          {{ t("admin.settings.tencentCaptcha.appCredentialsTitle") }}
                        </h3>
                        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                          {{ t("admin.settings.tencentCaptcha.appCredentialsHint") }}
                        </p>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                          {{ t("admin.settings.tencentCaptcha.appId") }}
                        </label>
                        <input
                          v-model="form.tencent_captcha_app_id"
                          type="text"
                          inputmode="numeric"
                          class="input font-mono text-sm"
                          placeholder="123456789"
                        />
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                          {{ t("admin.settings.tencentCaptcha.appSecretKey") }}
                        </label>
                        <input
                          v-model="form.tencent_captcha_app_secret_key"
                          type="password"
                          autocomplete="new-password"
                          class="input font-mono text-sm"
                          :placeholder="t('admin.settings.tencentCaptcha.keepExisting')"
                        />
                        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                          {{ form.tencent_captcha_app_secret_key_configured ? t("admin.settings.tencentCaptcha.configured") : t("admin.settings.tencentCaptcha.required") }}
                        </p>
                      </div>
                      <div class="border-t border-gray-100 pt-5 md:col-span-2 dark:border-dark-700">
                        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                          {{ t("admin.settings.tencentCaptcha.cloudCredentialsTitle") }}
                        </h3>
                        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                          {{ t("admin.settings.tencentCaptcha.cloudCredentialsHint") }}
                        </p>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                          {{ t("admin.settings.tencentCaptcha.cloudSecretId") }}
                        </label>
                        <input
                          v-model="form.tencent_captcha_cloud_secret_id"
                          type="password"
                          autocomplete="new-password"
                          class="input font-mono text-sm"
                          :placeholder="t('admin.settings.tencentCaptcha.keepExisting')"
                        />
                        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                          {{ form.tencent_captcha_cloud_secret_id_configured ? t("admin.settings.tencentCaptcha.configured") : t("admin.settings.tencentCaptcha.required") }}
                        </p>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                          {{ t("admin.settings.tencentCaptcha.cloudSecretKey") }}
                        </label>
                        <input
                          v-model="form.tencent_captcha_cloud_secret_key"
                          type="password"
                          autocomplete="new-password"
                          class="input font-mono text-sm"
                          :placeholder="t('admin.settings.tencentCaptcha.keepExisting')"
                        />
                        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                          {{ form.tencent_captcha_cloud_secret_key_configured ? t("admin.settings.tencentCaptcha.configured") : t("admin.settings.tencentCaptcha.required") }}
                        </p>
                      </div>
                    </div>
                    <p class="mt-5 text-xs text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.tencentCaptcha.camPermissionHint") }}
                    </p>
                    <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.tencentCaptcha.aidEncryptedHint") }}
                    </p>
                    <div class="mt-3 flex flex-wrap gap-x-4 gap-y-2 text-sm">
                      <a
                        :href="tencentCaptchaLinks.console"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="text-primary-600 hover:text-primary-500"
                      >
                        {{ t("admin.settings.tencentCaptcha.openCaptchaConsole") }}
                      </a>
                      <a
                        :href="tencentCaptchaLinks.cloudKeys"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="text-primary-600 hover:text-primary-500"
                      >
                        {{ t("admin.settings.tencentCaptcha.createCloudKeys") }}
                      </a>
                      <a
                        :href="tencentCaptchaLinks.webDocs"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="text-primary-600 hover:text-primary-500"
                      >
                        {{ t("admin.settings.tencentCaptcha.openWebDocs") }}
                      </a>
                    </div>
                  </div>

                  <!-- 阿里云验证码 2.0 配置 -->
                  <div v-else class="grid grid-cols-1 gap-6">
                    <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
                      <div>
                        <label
                          class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                        >
                          {{ t("admin.settings.aliyunCaptcha.region") }}
                        </label>
                        <div
                          class="grid grid-cols-2 gap-2 rounded-control bg-gray-100 p-1 dark:bg-dark-700"
                        >
                          <button
                            type="button"
                            class="inline-flex items-center justify-center rounded-control px-3 py-1.5 text-sm font-medium transition"
                            :class="
                              form.aliyun_captcha_region !== 'sgp'
                                ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300'
                                : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                            "
                            @click="form.aliyun_captcha_region = 'cn'"
                          >
                            {{ t("admin.settings.aliyunCaptcha.regionCn") }}
                          </button>
                          <button
                            type="button"
                            class="inline-flex items-center justify-center rounded-control px-3 py-1.5 text-sm font-medium transition"
                            :class="
                              form.aliyun_captcha_region === 'sgp'
                                ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300'
                                : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                            "
                            @click="form.aliyun_captcha_region = 'sgp'"
                          >
                            {{ t("admin.settings.aliyunCaptcha.regionSgp") }}
                          </button>
                        </div>
                        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                          {{ t("admin.settings.aliyunCaptcha.regionHint") }}
                        </p>
                      </div>
                      <div>
                        <label
                          class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                        >
                          {{ t("admin.settings.aliyunCaptcha.prefix") }}
                        </label>
                        <input
                          v-model="form.aliyun_captcha_prefix"
                          type="text"
                          class="input font-mono text-sm"
                          placeholder="14xxxxx"
                        />
                        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                          {{ t("admin.settings.aliyunCaptcha.prefixHint") }}
                        </p>
                      </div>
                    </div>
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.aliyunCaptcha.sceneId") }}
                      </label>
                      <input
                        v-model="form.aliyun_captcha_scene_id"
                        type="text"
                        class="input font-mono text-sm"
                        placeholder="1cxxxxxx"
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.aliyunCaptcha.sceneIdHint") }}
                      </p>
                    </div>
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.aliyunCaptcha.accessKeyId") }}
                      </label>
                      <input
                        v-model="form.aliyun_captcha_access_key_id"
                        type="text"
                        class="input font-mono text-sm"
                        placeholder="LTAI..."
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.aliyunCaptcha.accessKeyIdHint") }}
                      </p>
                    </div>
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.aliyunCaptcha.accessKeySecret") }}
                      </label>
                      <input
                        v-model="form.aliyun_captcha_access_key_secret"
                        type="password"
                        autocomplete="new-password"
                        class="input font-mono text-sm"
                        placeholder="••••••••"
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{
                          form.aliyun_captcha_access_key_secret_configured
                            ? t(
                                "admin.settings.aliyunCaptcha.accessKeySecretConfiguredHint",
                              )
                            : t("admin.settings.aliyunCaptcha.accessKeySecretHint")
                        }}
                      </p>
                    </div>
                  </div>
                </div>
              </Collapse>
            </div>
          </div>

          <!-- LinuxDo Connect OAuth 登录 -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.linuxdo.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.linuxdo.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <div class="flex items-center justify-between">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.linuxdo.enable")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.linuxdo.enableHint") }}
                  </p>
                </div>
                <Toggle v-model="form.linuxdo_connect_enabled" />
              </div>

              <Collapse :open="form.linuxdo_connect_enabled" unmount-on-hide>
                <div

                  class="border-t border-gray-100 pt-4 dark:border-dark-700"
                >
                  <div class="grid grid-cols-1 gap-6">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.linuxdo.clientId") }}
                      </label>
                      <input
                        v-model="form.linuxdo_connect_client_id"
                        type="text"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.linuxdo.clientIdPlaceholder')
                        "
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.linuxdo.clientIdHint") }}
                      </p>
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.linuxdo.clientSecret") }}
                      </label>
                      <input
                        v-model="form.linuxdo_connect_client_secret"
                        type="password"
                        class="input font-mono text-sm"
                        :placeholder="
                          form.linuxdo_connect_client_secret_configured
                            ? t(
                                'admin.settings.linuxdo.clientSecretConfiguredPlaceholder',
                              )
                            : t('admin.settings.linuxdo.clientSecretPlaceholder')
                        "
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{
                          form.linuxdo_connect_client_secret_configured
                            ? t(
                                "admin.settings.linuxdo.clientSecretConfiguredHint",
                              )
                            : t("admin.settings.linuxdo.clientSecretHint")
                        }}
                      </p>
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.linuxdo.redirectUrl") }}
                      </label>
                      <input
                        v-model="form.linuxdo_connect_redirect_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.linuxdo.redirectUrlPlaceholder')
                        "
                      />
                      <div
                        class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3"
                      >
                        <button
                          type="button"
                          class="btn btn-secondary btn-sm h-9 w-fit"
                          @click="setAndCopyLinuxdoRedirectUrl"
                        >
                          {{ t("admin.settings.linuxdo.quickSetCopy") }}
                        </button>
                        <code
                          v-if="linuxdoRedirectUrlSuggestion"
                          class="select-all break-all rounded-compact bg-gray-50 px-2 py-1 font-mono text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300"
                        >
                          {{ linuxdoRedirectUrlSuggestion }}
                        </code>
                      </div>
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.linuxdo.redirectUrlHint") }}
                      </p>
                    </div>
                  </div>
                </div>
              </Collapse>
            </div>
          </div>

          <!-- GitHub / Google 邮箱快捷登录 -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.emailOAuth.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{
                  t("admin.settings.emailOAuth.description")
                }}
              </p>
            </div>
            <div class="space-y-6 p-6">
              <div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
                <div class="rounded-control border border-gray-200 p-4 dark:border-dark-700">
                  <div class="flex items-start justify-between gap-4">
                    <div>
                      <h3 class="font-medium text-gray-900 dark:text-white">
                        GitHub
                      </h3>
                      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                        {{
                          t("admin.settings.emailOAuth.githubHint")
                        }}
                      </p>
                    </div>
                    <Toggle v-model="form.github_oauth_enabled" />
                  </div>

                  <Collapse :open="form.github_oauth_enabled" unmount-on-hide>
                    <div class="mt-4 space-y-4">
                      <div class="rounded-control bg-gray-50 px-3 py-2 text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                        {{ t("admin.settings.emailOAuth.githubGuideBeforeLink") }}
                        <a
                          data-testid="github-oauth-apps-guide-link"
                          href="https://github.com/settings/developers"
                          target="_blank"
                          rel="noopener noreferrer"
                          class="font-medium text-primary-600 hover:underline dark:text-primary-400"
                        >OAuth Apps</a>
                        {{ t("admin.settings.emailOAuth.githubGuideAfterLink") }}
                      </div>

                      <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
                        <div>
                          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">Client ID</label>
                          <input
                            v-model="form.github_oauth_client_id"
                            type="text"
                            class="input font-mono text-sm"
                            placeholder="GitHub OAuth Client ID"
                          />
                        </div>
                        <div>
                          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">Client Secret</label>
                          <input
                            v-model="form.github_oauth_client_secret"
                            type="password"
                            class="input font-mono text-sm"
                            :placeholder="
                              form.github_oauth_client_secret_configured
                                ? t('admin.settings.emailOAuth.secretConfiguredPlaceholder')
                                : 'GitHub OAuth Client Secret'
                            "
                          />
                        </div>
                      </div>

                      <div>
                        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                          {{ t("admin.settings.emailOAuth.backendCallbackUrl") }}
                        </label>
                        <input
                          v-model="form.github_oauth_redirect_url"
                          type="url"
                          class="input font-mono text-sm"
                          placeholder="https://your-domain.com/api/v1/auth/oauth/github/callback"
                        />
                        <div class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
                          <button
                            type="button"
                            class="btn btn-secondary btn-sm h-9 w-fit"
                            @click="setAndCopyEmailOAuthRedirectUrl('github')"
                          >
                            {{ t("admin.settings.emailOAuth.generateAndCopy") }}
                          </button>
                          <code
                            v-if="githubOAuthRedirectUrlSuggestion"
                            class="select-all break-all rounded-compact bg-gray-50 px-2 py-1 font-mono text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300"
                          >
                            {{ githubOAuthRedirectUrlSuggestion }}
                          </code>
                        </div>
                      </div>

                      <div>
                        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                          {{ t("admin.settings.emailOAuth.frontendCallbackUrl") }}
                        </label>
                        <input
                          v-model="form.github_oauth_frontend_redirect_url"
                          type="text"
                          class="input font-mono text-sm"
                          placeholder="/auth/oauth/callback"
                        />
                      </div>
                    </div>
                  </Collapse>
                </div>

                <div class="rounded-control border border-gray-200 p-4 dark:border-dark-700">
                  <div class="flex items-start justify-between gap-4">
                    <div>
                      <h3 class="font-medium text-gray-900 dark:text-white">
                        Google
                      </h3>
                      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                        {{
                          t("admin.settings.emailOAuth.googleHint")
                        }}
                      </p>
                    </div>
                    <Toggle v-model="form.google_oauth_enabled" />
                  </div>

                  <Collapse :open="form.google_oauth_enabled" unmount-on-hide>
                    <div class="mt-4 space-y-4">
                      <div class="rounded-control bg-gray-50 px-3 py-2 text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                        {{
                          t("admin.settings.emailOAuth.googleGuide")
                        }}
                      </div>

                      <div class="border-t border-gray-200 pt-4 dark:border-dark-700">
                        <div class="flex items-start justify-between gap-4">
                          <div>
                            <h4 class="text-sm font-medium text-gray-900 dark:text-white">
                              Google One Tap
                            </h4>
                            <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">
                              {{
                                t("admin.settings.emailOAuth.googleOneTapHint")
                              }}
                            </p>
                          </div>
                          <Toggle v-model="form.google_one_tap_enabled" />
                        </div>

                        <Collapse :open="form.google_one_tap_enabled" unmount-on-hide>
                          <div class="mt-3">
                            <label class="mb-2 block text-xs font-medium text-gray-700 dark:text-gray-300">
                              Authorized JavaScript origin
                            </label>
                            <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
                              <code class="min-w-0 flex-1 select-all break-all rounded-compact bg-gray-50 px-2 py-1.5 font-mono text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                                {{ googleOneTapOriginSuggestion }}
                              </code>
                              <button
                                type="button"
                                class="btn btn-secondary btn-sm h-9 w-fit"
                                :disabled="!googleOneTapOriginSuggestion"
                                @click="copyGoogleOneTapOrigin"
                              >
                                <Icon name="copy" size="sm" class="mr-1.5" />
                                {{ t("admin.settings.emailOAuth.copy") }}
                              </button>
                            </div>
                            <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-gray-400">
                              {{
                                t("admin.settings.emailOAuth.googleOneTapOriginHint")
                              }}
                            </p>
                          </div>
                        </Collapse>
                      </div>

                      <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
                        <div>
                          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">Client ID</label>
                          <input
                            v-model="form.google_oauth_client_id"
                            type="text"
                            class="input font-mono text-sm"
                            placeholder="Google OAuth Client ID"
                          />
                        </div>
                        <div>
                          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">Client Secret</label>
                          <input
                            v-model="form.google_oauth_client_secret"
                            type="password"
                            class="input font-mono text-sm"
                            :placeholder="
                              form.google_oauth_client_secret_configured
                                ? t('admin.settings.emailOAuth.secretConfiguredPlaceholder')
                                : 'Google OAuth Client Secret'
                            "
                          />
                        </div>
                      </div>

                      <div>
                        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                          {{ t("admin.settings.emailOAuth.backendCallbackUrl") }}
                        </label>
                        <input
                          v-model="form.google_oauth_redirect_url"
                          type="url"
                          class="input font-mono text-sm"
                          placeholder="https://your-domain.com/api/v1/auth/oauth/google/callback"
                        />
                        <div class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
                          <button
                            type="button"
                            class="btn btn-secondary btn-sm h-9 w-fit"
                            @click="setAndCopyEmailOAuthRedirectUrl('google')"
                          >
                            {{ t("admin.settings.emailOAuth.generateAndCopy") }}
                          </button>
                          <code
                            v-if="googleOAuthRedirectUrlSuggestion"
                            class="select-all break-all rounded-compact bg-gray-50 px-2 py-1 font-mono text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300"
                          >
                            {{ googleOAuthRedirectUrlSuggestion }}
                          </code>
                        </div>
                      </div>

                      <div>
                        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                          {{ t("admin.settings.emailOAuth.frontendCallbackUrl") }}
                        </label>
                        <input
                          v-model="form.google_oauth_frontend_redirect_url"
                          type="text"
                          class="input font-mono text-sm"
                          placeholder="/auth/oauth/callback"
                        />
                      </div>
                    </div>
                  </Collapse>
                </div>
              </div>
            </div>
          </div>

          <!-- WeChat Connect OAuth 登录 -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.wechatConnect.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.wechatConnect.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <div class="flex items-center justify-between">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.wechatConnect.enabledLabel")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.wechatConnect.enabledHint") }}
                  </p>
                </div>
                <Toggle
                  v-model="form.wechat_connect_enabled"
                  data-testid="wechat-connect-enabled"
                />
              </div>

              <Collapse :open="form.wechat_connect_enabled" unmount-on-hide>
                <div

                  class="space-y-6 border-t border-gray-100 pt-4 dark:border-dark-700"
                >
                  <div class="space-y-4">
                    <div
                      class="rounded-control border border-gray-200 p-4 dark:border-dark-700"
                    >
                      <div class="flex items-start justify-between gap-4">
                        <div>
                          <h3 class="font-medium text-gray-900 dark:text-white">
                            {{ t("admin.settings.wechatConnect.pcApp") }}
                          </h3>
                          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                            {{
                              t("admin.settings.wechatConnect.pcAppHint")
                            }}
                          </p>
                        </div>
                        <Toggle
                          :model-value="form.wechat_connect_open_enabled"
                          data-testid="wechat-connect-open-enabled"
                          @update:model-value="handleWeChatOpenEnabledChange"
                        />
                      </div>
                      <Collapse :open="form.wechat_connect_open_enabled" unmount-on-hide>
                        <div

                          class="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2"
                        >
                          <div>
                            <label
                              class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                            >
                              {{ t("admin.settings.wechatConnect.pcAppId") }}
                            </label>
                            <input
                              v-model="form.wechat_connect_open_app_id"
                              data-testid="wechat-connect-open-app-id"
                              type="text"
                              class="input font-mono text-sm"
                              :placeholder="
                                t('admin.settings.wechatConnect.pcAppIdPlaceholder')
                              "
                            />
                          </div>
                          <div>
                            <label
                              class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                            >
                              {{ t("admin.settings.wechatConnect.pcAppSecret") }}
                            </label>
                            <input
                              v-model="form.wechat_connect_open_app_secret"
                              data-testid="wechat-connect-open-app-secret"
                              type="password"
                              class="input font-mono text-sm"
                              :placeholder="
                                form.wechat_connect_open_app_secret_configured
                                  ? t('admin.settings.wechatConnect.appSecretConfiguredPlaceholder')
                                  : t('admin.settings.wechatConnect.pcAppSecretPlaceholder')
                              "
                            />
                          </div>
                        </div>
                      </Collapse>
                    </div>

                    <div
                      class="rounded-control border border-gray-200 p-4 dark:border-dark-700"
                    >
                      <div class="flex items-start justify-between gap-4">
                        <div>
                          <h3 class="font-medium text-gray-900 dark:text-white">
                            {{ t("admin.settings.wechatConnect.mp") }}
                          </h3>
                          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                            {{
                              t("admin.settings.wechatConnect.mpHint")
                            }}
                          </p>
                        </div>
                        <Toggle
                          :model-value="form.wechat_connect_mp_enabled"
                          data-testid="wechat-connect-mp-enabled"
                          @update:model-value="handleWeChatMPEnabledChange"
                        />
                      </div>
                      <Collapse :open="form.wechat_connect_mp_enabled" unmount-on-hide>
                        <div

                          class="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2"
                        >
                          <div>
                            <label
                              class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                            >
                              {{ t("admin.settings.wechatConnect.mpAppId") }}
                            </label>
                            <input
                              v-model="form.wechat_connect_mp_app_id"
                              data-testid="wechat-connect-mp-app-id"
                              type="text"
                              class="input font-mono text-sm"
                              :placeholder="
                                t('admin.settings.wechatConnect.mpAppId')
                              "
                            />
                          </div>
                          <div>
                            <label
                              class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                            >
                              {{
                                t("admin.settings.wechatConnect.mpAppSecret")
                              }}
                            </label>
                            <input
                              v-model="form.wechat_connect_mp_app_secret"
                              data-testid="wechat-connect-mp-app-secret"
                              type="password"
                              class="input font-mono text-sm"
                              :placeholder="
                                form.wechat_connect_mp_app_secret_configured
                                  ? t('admin.settings.wechatConnect.appSecretConfiguredPlaceholder')
                                  : t('admin.settings.wechatConnect.mpAppSecret')
                              "
                            />
                          </div>
                        </div>
                      </Collapse>
                    </div>

                    <div
                      class="rounded-control border border-gray-200 p-4 dark:border-dark-700"
                    >
                      <div class="flex items-start justify-between gap-4">
                        <div>
                          <h3 class="font-medium text-gray-900 dark:text-white">
                            {{ t("admin.settings.wechatConnect.mobile") }}
                          </h3>
                          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                            {{
                              t("admin.settings.wechatConnect.mobileHint")
                            }}
                          </p>
                        </div>
                        <Toggle
                          :model-value="form.wechat_connect_mobile_enabled"
                          data-testid="wechat-connect-mobile-enabled"
                          @update:model-value="handleWeChatMobileEnabledChange"
                        />
                      </div>
                      <Collapse :open="form.wechat_connect_mobile_enabled" unmount-on-hide>
                        <div

                          class="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2"
                        >
                          <div>
                            <label
                              class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                            >
                              {{ t("admin.settings.wechatConnect.mobileAppId") }}
                            </label>
                            <input
                              v-model="form.wechat_connect_mobile_app_id"
                              data-testid="wechat-connect-mobile-app-id"
                              type="text"
                              class="input font-mono text-sm"
                              :placeholder="
                                t('admin.settings.wechatConnect.mobileAppId')
                              "
                            />
                          </div>
                          <div>
                            <label
                              class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                            >
                              {{ t("admin.settings.wechatConnect.mobileAppSecret") }}
                            </label>
                            <input
                              v-model="form.wechat_connect_mobile_app_secret"
                              data-testid="wechat-connect-mobile-app-secret"
                              type="password"
                              class="input font-mono text-sm"
                              :placeholder="
                                form.wechat_connect_mobile_app_secret_configured
                                  ? t('admin.settings.wechatConnect.appSecretConfiguredPlaceholder')
                                  : t('admin.settings.wechatConnect.mobileAppSecret')
                              "
                            />
                          </div>
                        </div>
                      </Collapse>
                    </div>
                  </div>

                  <Collapse :open="form.wechat_connect_open_enabled &&
                      (form.wechat_connect_mp_enabled ||
                        form.wechat_connect_mobile_enabled)" unmount-on-hide>
                    <div

                      class="rounded-control border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700 dark:border-amber-900/40 dark:bg-amber-900/10 dark:text-amber-300"
                    >
                      {{
                        t("admin.settings.wechatConnect.unionIdHint")
                      }}
                    </div>
                  </Collapse>

                  <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{
                          t("admin.settings.wechatConnect.browserRedirectUrl")
                        }}
                      </label>
                      <input
                        data-testid="wechat-connect-redirect-url"
                        v-model="form.wechat_connect_redirect_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="t('admin.settings.wechatConnect.redirectUrlPlaceholder')"
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{
                          t("admin.settings.wechatConnect.browserRedirectUrlHint")
                        }}
                      </p>
                      <div
                        class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3"
                      >
                        <button
                          type="button"
                          class="btn btn-secondary btn-sm h-9 w-fit"
                          @click="setAndCopyWeChatRedirectUrl"
                        >
                          {{ t("admin.settings.wechatConnect.generateAndCopy") }}
                        </button>
                        <code
                          v-if="wechatRedirectUrlSuggestion"
                          class="select-all break-all rounded-compact bg-gray-50 px-2 py-1 font-mono text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300"
                        >
                          {{ wechatRedirectUrlSuggestion }}
                        </code>
                      </div>
                    </div>
                  </div>

                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.wechatConnect.frontendRedirectUrlLabel") }}
                    </label>
                    <input
                      data-testid="wechat-connect-frontend-redirect-url"
                      v-model="form.wechat_connect_frontend_redirect_url"
                      type="text"
                      class="input font-mono text-sm"
                      :placeholder="t('admin.settings.wechatConnect.frontendRedirectUrlPlaceholder')"
                    />
                    <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.wechatConnect.frontendRedirectUrlHint") }}
                    </p>
                  </div>
                </div>
              </Collapse>
            </div>
          </div>

          <!-- 钉钉 Connect OAuth 登录 -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.dingtalk.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.dingtalk.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <div class="flex items-center justify-between">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.dingtalk.enable")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.dingtalk.enableHint") }}
                  </p>
                </div>
                <Toggle v-model="form.dingtalk_connect_enabled" />
              </div>

              <Collapse :open="form.dingtalk_connect_enabled" unmount-on-hide>
                <div

                  class="border-t border-gray-100 pt-4 dark:border-dark-700"
                >
                  <div class="grid grid-cols-1 gap-6">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.dingtalk.clientId") }}
                      </label>
                      <input
                        v-model="form.dingtalk_connect_client_id"
                        type="text"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.dingtalk.clientIdPlaceholder')
                        "
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.dingtalk.clientIdHint") }}
                      </p>
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.dingtalk.clientSecret") }}
                      </label>
                      <input
                        v-model="form.dingtalk_connect_client_secret"
                        type="password"
                        class="input font-mono text-sm"
                        :placeholder="
                          form.dingtalk_connect_client_secret_configured
                            ? t(
                                'admin.settings.dingtalk.clientSecretConfiguredPlaceholder',
                              )
                            : t('admin.settings.dingtalk.clientSecretPlaceholder')
                        "
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{
                          form.dingtalk_connect_client_secret_configured
                            ? t(
                                "admin.settings.dingtalk.clientSecretConfiguredHint",
                              )
                            : t("admin.settings.dingtalk.clientSecretHint")
                        }}
                      </p>
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.dingtalk.redirectUrl") }}
                      </label>
                      <input
                        v-model="form.dingtalk_connect_redirect_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.dingtalk.redirectUrlPlaceholder')
                        "
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.dingtalk.redirectUrlHint") }}
                      </p>
                    </div>

                    <!-- 企业限制策略 -->
                    <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
                      <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                        {{ t("admin.settings.dingtalk.corpPolicy.label") }}
                      </label>
                      <p class="mb-3 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.dingtalk.corpPolicy.hint") }}
                      </p>
                      <div class="space-y-2">
                        <label class="flex cursor-pointer items-center gap-3">
                          <input
                            v-model="form.dingtalk_connect_corp_restriction_policy"
                            type="radio"
                            value="none"
                            class="h-4 w-4 text-primary-600"
                          />
                          <span class="text-sm text-gray-700 dark:text-gray-300">
                            {{ t("admin.settings.dingtalk.corpPolicy.none") }}
                          </span>
                        </label>
                        <label class="flex cursor-pointer items-center gap-3">
                          <input
                            v-model="form.dingtalk_connect_corp_restriction_policy"
                            type="radio"
                            value="internal_only"
                            class="h-4 w-4 text-primary-600"
                          />
                          <span class="text-sm text-gray-700 dark:text-gray-300">
                            {{ t("admin.settings.dingtalk.corpPolicy.internalOnly") }}
                          </span>
                        </label>
                      </div>
                    </div>

                    <!-- bypass_registration 开关（仅 internal_only 模式下可见可用） -->
                    <div
                      v-if="form.dingtalk_connect_corp_restriction_policy === 'internal_only'"
                      class="flex items-center justify-between pt-4 border-t border-gray-100 dark:border-dark-700"
                    >
                      <div>
                        <label class="font-medium text-gray-900 dark:text-white">{{
                          t("admin.settings.dingtalk.bypassRegistration")
                        }}</label>
                        <p class="text-sm text-gray-500 dark:text-gray-400">
                          {{ t("admin.settings.dingtalk.bypassRegistrationHint") }}
                        </p>
                      </div>
                      <Toggle v-model="form.dingtalk_connect_bypass_registration" />
                    </div>

                    <!-- 身份同步开关（仅 internal_only 模式下可见） -->
                    <div
                      v-if="form.dingtalk_connect_corp_restriction_policy === 'internal_only'"
                      class="pt-4 border-t border-gray-100 dark:border-dark-700 space-y-2"
                    >
                      <div class="flex items-center justify-between">
                        <div>
                          <label class="font-medium text-gray-900 dark:text-white">{{
                            t("admin.settings.dingtalk.syncDisplayName")
                          }}</label>
                          <p class="text-sm text-gray-500 dark:text-gray-400">
                            {{ t("admin.settings.dingtalk.syncDisplayNameHint") }}
                          </p>
                        </div>
                        <Toggle v-model="form.dingtalk_connect_sync_display_name" />
                      </div>
                      <div v-if="form.dingtalk_connect_sync_display_name" class="space-y-2">
                        <div class="flex items-center gap-2">
                          <label class="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap min-w-[5rem]">
                            {{ t("admin.settings.dingtalk.syncDisplayNameTarget") }}
                          </label>
                          <input
                            v-model="form.dingtalk_connect_sync_display_name_attr_key"
                            type="text"
                            placeholder="dingtalk_name"
                            class="input text-sm flex-1 max-w-xs"
                          />
                        </div>
                        <div class="flex items-center gap-2">
                          <label class="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap min-w-[5rem]">
                            {{ t("admin.settings.dingtalk.syncAttrDisplayName") }}
                          </label>
                          <input
                            v-model="form.dingtalk_connect_sync_display_name_attr_name"
                            type="text"
                            placeholder="钉钉姓名"
                            class="input text-sm flex-1 max-w-xs"
                          />
                        </div>
                      </div>
                      <p v-if="form.dingtalk_connect_sync_display_name" class="text-xs text-gray-400 dark:text-gray-500">
                        {{ t("admin.settings.dingtalk.syncDisplayNameTargetHint") }}
                      </p>
                    </div>
                    <div
                      v-if="form.dingtalk_connect_corp_restriction_policy === 'internal_only'"
                      class="pt-4 border-t border-gray-100 dark:border-dark-700 space-y-2"
                    >
                      <div class="flex items-center justify-between">
                        <div>
                          <label class="font-medium text-gray-900 dark:text-white">{{
                            t("admin.settings.dingtalk.syncCorpEmail")
                          }}</label>
                          <p class="text-sm text-gray-500 dark:text-gray-400">
                            {{ t("admin.settings.dingtalk.syncCorpEmailHint") }}
                          </p>
                          <p class="text-xs text-amber-600 dark:text-amber-400 mt-1">
                            {{ t("admin.settings.dingtalk.syncCorpEmailPermissionHint") }}
                          </p>
                        </div>
                        <Toggle v-model="form.dingtalk_connect_sync_corp_email" />
                      </div>
                      <div v-if="form.dingtalk_connect_sync_corp_email" class="space-y-2">
                        <div class="flex items-center gap-2">
                          <label class="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap min-w-[5rem]">
                            {{ t("admin.settings.dingtalk.syncCorpEmailTarget") }}
                          </label>
                          <input
                            v-model="form.dingtalk_connect_sync_corp_email_attr_key"
                            type="text"
                            placeholder="dingtalk_email"
                            class="input text-sm flex-1 max-w-xs"
                          />
                        </div>
                        <div class="flex items-center gap-2">
                          <label class="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap min-w-[5rem]">
                            {{ t("admin.settings.dingtalk.syncAttrDisplayName") }}
                          </label>
                          <input
                            v-model="form.dingtalk_connect_sync_corp_email_attr_name"
                            type="text"
                            placeholder="钉钉企业邮箱"
                            class="input text-sm flex-1 max-w-xs"
                          />
                        </div>
                      </div>
                      <p v-if="form.dingtalk_connect_sync_corp_email" class="text-xs text-gray-400 dark:text-gray-500">
                        {{ t("admin.settings.dingtalk.syncCorpEmailTargetHint") }}
                      </p>
                    </div>
                    <div
                      v-if="form.dingtalk_connect_corp_restriction_policy === 'internal_only'"
                      class="pt-4 border-t border-gray-100 dark:border-dark-700 space-y-2"
                    >
                      <div class="flex items-center justify-between">
                        <div>
                          <label class="font-medium text-gray-900 dark:text-white">{{
                            t("admin.settings.dingtalk.syncDept")
                          }}</label>
                          <p class="text-sm text-gray-500 dark:text-gray-400">
                            {{ t("admin.settings.dingtalk.syncDeptHint") }}
                          </p>
                          <p class="text-xs text-amber-600 dark:text-amber-400 mt-1">
                            {{ t("admin.settings.dingtalk.syncDeptPermissionHint") }}
                          </p>
                        </div>
                        <Toggle v-model="form.dingtalk_connect_sync_dept" />
                      </div>
                      <div v-if="form.dingtalk_connect_sync_dept" class="space-y-2">
                        <div class="flex items-center gap-2">
                          <label class="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap min-w-[5rem]">
                            {{ t("admin.settings.dingtalk.syncDeptTarget") }}
                          </label>
                          <input
                            v-model="form.dingtalk_connect_sync_dept_attr_key"
                            type="text"
                            placeholder="dingtalk_department"
                            class="input text-sm flex-1 max-w-xs"
                          />
                        </div>
                        <div class="flex items-center gap-2">
                          <label class="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap min-w-[5rem]">
                            {{ t("admin.settings.dingtalk.syncAttrDisplayName") }}
                          </label>
                          <input
                            v-model="form.dingtalk_connect_sync_dept_attr_name"
                            type="text"
                            placeholder="钉钉部门"
                            class="input text-sm flex-1 max-w-xs"
                          />
                        </div>
                      </div>
                      <p v-if="form.dingtalk_connect_sync_dept" class="text-xs text-gray-400 dark:text-gray-500">
                        {{ t("admin.settings.dingtalk.syncDeptTargetHint") }}
                      </p>
                    </div>
                  </div>
                </div>
              </Collapse>
            </div>
          </div>

          <!-- Generic OIDC OAuth 登录 -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.oidc.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.oidc.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <div class="flex items-center justify-between">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.oidc.enable")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.oidc.enableHint") }}
                  </p>
                </div>
                <Toggle v-model="form.oidc_connect_enabled" />
              </div>

              <Collapse :open="form.oidc_connect_enabled" unmount-on-hide>
                <div

                  class="space-y-6 border-t border-gray-100 pt-4 dark:border-dark-700"
                >
                  <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.providerName") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_provider_name"
                        type="text"
                        class="input"
                        :placeholder="
                          t('admin.settings.oidc.providerNamePlaceholder')
                        "
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.clientId") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_client_id"
                        type="text"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.clientIdPlaceholder')
                        "
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.clientSecret") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_client_secret"
                        type="password"
                        class="input font-mono text-sm"
                        :placeholder="
                          form.oidc_connect_client_secret_configured
                            ? t(
                                'admin.settings.oidc.clientSecretConfiguredPlaceholder',
                              )
                            : t('admin.settings.oidc.clientSecretPlaceholder')
                        "
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{
                          form.oidc_connect_client_secret_configured
                            ? t("admin.settings.oidc.clientSecretConfiguredHint")
                            : t("admin.settings.oidc.clientSecretHint")
                        }}
                      </p>
                    </div>
                  </div>

                  <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.issuerUrl") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_issuer_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.issuerUrlPlaceholder')
                        "
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.discoveryUrl") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_discovery_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.discoveryUrlPlaceholder')
                        "
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.authorizeUrl") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_authorize_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.authorizeUrlPlaceholder')
                        "
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.tokenUrl") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_token_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.tokenUrlPlaceholder')
                        "
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.userinfoUrl") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_userinfo_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.userinfoUrlPlaceholder')
                        "
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.jwksUrl") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_jwks_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="t('admin.settings.oidc.jwksUrlPlaceholder')"
                      />
                    </div>
                  </div>

                  <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.scopes") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_scopes"
                        type="text"
                        class="input font-mono text-sm"
                        :placeholder="t('admin.settings.oidc.scopesPlaceholder')"
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.oidc.scopesHint") }}
                      </p>
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.redirectUrl") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_redirect_url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.redirectUrlPlaceholder')
                        "
                      />
                      <div
                        class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3"
                      >
                        <button
                          type="button"
                          class="btn btn-secondary btn-sm h-9 w-fit"
                          @click="setAndCopyOIDCRedirectUrl"
                        >
                          {{ t("admin.settings.oidc.quickSetCopy") }}
                        </button>
                        <code
                          v-if="oidcRedirectUrlSuggestion"
                          class="select-all break-all rounded-compact bg-gray-50 px-2 py-1 font-mono text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300"
                        >
                          {{ oidcRedirectUrlSuggestion }}
                        </code>
                      </div>
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.oidc.redirectUrlHint") }}
                      </p>
                    </div>

                    <div class="lg:col-span-2">
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.frontendRedirectUrl") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_frontend_redirect_url"
                        type="text"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.frontendRedirectUrlPlaceholder')
                        "
                      />
                      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                        {{ t("admin.settings.oidc.frontendRedirectUrlHint") }}
                      </p>
                    </div>
                  </div>

                  <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.tokenAuthMethod") }}
                      </label>
                      <Select
                        v-model="form.oidc_connect_token_auth_method"
                        :options="oidcTokenAuthMethodOptions"
                        class="font-mono text-sm"
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.clockSkewSeconds") }}
                      </label>
                      <input
                        v-model.number="form.oidc_connect_clock_skew_seconds"
                        type="number"
                        min="0"
                        max="600"
                        class="input"
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.allowedSigningAlgs") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_allowed_signing_algs"
                        type="text"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.allowedSigningAlgsPlaceholder')
                        "
                      />
                    </div>
                  </div>

                  <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
                    <div
                      class="flex items-center justify-between rounded-compact border border-gray-200 px-4 py-3 dark:border-dark-700"
                    >
                      <div>
                        <label class="font-medium text-gray-900 dark:text-white">
                          {{ t("admin.settings.oidc.usePkce") }}
                        </label>
                      </div>
                      <Toggle
                        v-model="form.oidc_connect_use_pkce"
                        data-testid="oidc-connect-use-pkce"
                      />
                    </div>

                    <div
                      class="flex items-center justify-between rounded-compact border border-gray-200 px-4 py-3 dark:border-dark-700"
                    >
                      <div>
                        <label class="font-medium text-gray-900 dark:text-white">
                          {{ t("admin.settings.oidc.validateIdToken") }}
                        </label>
                      </div>
                      <Toggle
                        v-model="form.oidc_connect_validate_id_token"
                        data-testid="oidc-connect-validate-id-token"
                      />
                    </div>

                    <div
                      class="flex items-center justify-between rounded-compact border border-gray-200 px-4 py-3 dark:border-dark-700"
                    >
                      <div>
                        <label class="font-medium text-gray-900 dark:text-white">
                          {{ t("admin.settings.oidc.requireEmailVerified") }}
                        </label>
                      </div>
                      <Toggle
                        v-model="form.oidc_connect_require_email_verified"
                      />
                    </div>
                  </div>

                  <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.userinfoEmailPath") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_userinfo_email_path"
                        type="text"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.userinfoEmailPathPlaceholder')
                        "
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.userinfoIdPath") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_userinfo_id_path"
                        type="text"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.userinfoIdPathPlaceholder')
                        "
                      />
                    </div>

                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.oidc.userinfoUsernamePath") }}
                      </label>
                      <input
                        v-model="form.oidc_connect_userinfo_username_path"
                        type="text"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.oidc.userinfoUsernamePathPlaceholder')
                        "
                      />
                    </div>
                  </div>
                </div>
              </Collapse>
            </div>
          </div>
        </div>
        <!-- /Tab: Security — Registration, Turnstile, LinuxDo, OIDC -->

        <!-- Tab: Users -->
        <div v-show="activeTab === 'users'" v-content-reveal="activeTab === 'users'" class="space-y-4">
          <!-- Default Settings -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.defaults.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.defaults.description") }}
              </p>
            </div>
            <div class="space-y-6 p-6">
              <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                <div>
                  <label
                    class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    {{ t("admin.settings.defaults.defaultBalance") }}
                  </label>
                  <input
                    v-model.number="form.default_balance"
                    type="number"
                    step="0.01"
                    min="0"
                    class="input"
                    placeholder="0.00"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.defaults.defaultBalanceHint") }}
                  </p>
                </div>
                <div>
                  <label
                    class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    {{ t("admin.settings.defaults.defaultConcurrency") }}
                  </label>
                  <input
                    v-model.number="form.default_concurrency"
                    type="number"
                    min="1"
                    class="input"
                    placeholder="1"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.defaults.defaultConcurrencyHint") }}
                  </p>
                </div>
                <div>
                  <label
                    class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    {{ t("admin.settings.defaults.defaultUserRpmLimit") }}
                  </label>
                  <input
                    v-model.number="form.default_user_rpm_limit"
                    type="number"
                    min="0"
                    :max="MAX_USER_API_KEY_LIMIT"
                    step="1"
                    class="input"
                    placeholder="0"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.defaults.defaultUserRpmLimitHint") }}
                  </p>
                </div>
                <div>
                  <label
                    class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    {{ t("admin.settings.defaults.defaultUserApiKeyLimit") }}
                  </label>
                  <input
                    v-model.number="form.default_user_api_key_limit"
                    type="number"
                    min="0"
                    step="1"
                    class="input"
                    data-test="default-user-api-key-limit"
                    placeholder="100"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.defaults.defaultUserApiKeyLimitHint") }}
                  </p>
                </div>
              </div>

              <RuleListEditor
                class="border-t border-gray-100 pt-4 dark:border-dark-700"
                :items="form.default_subscriptions"
                :title="t('admin.settings.defaults.defaultSubscriptions')"
                :hint="t('admin.settings.defaults.defaultSubscriptionsHint')"
                :add-label="t('admin.settings.defaults.addDefaultSubscription')"
                :empty-text="t('admin.settings.defaults.defaultSubscriptionsEmpty')"
                :add-disabled="subscriptionPlans.length === 0"
                test-id="default-subscriptions"
                @add="addDefaultSubscription"
                @remove="removeDefaultSubscription"
              >
                <template #row="{ item }">
                  <Select
                    v-model="item.plan_id"
                    :aria-label="t('admin.settings.defaults.subscriptionGroup')"
                  :options="defaultSubscriptionPlanOptions"
                    :placeholder="
                      t('admin.settings.defaults.subscriptionGroup')
                    "
                  />
                </template>
              </RuleListEditor>
            </div>
          </div>

          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.authSourceDefaults.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.authSourceDefaults.description") }}
              </p>
            </div>
            <div class="space-y-6 p-6">
              <div
                class="flex items-center justify-between rounded-compact border border-gray-200 px-4 py-3 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">
                    {{ t("admin.settings.authSourceDefaults.requireEmailLabel") }}
                  </label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.authSourceDefaults.requireEmailHint") }}
                  </p>
                </div>
                <Toggle v-model="form.force_email_on_third_party_signup" />
              </div>

              <div class="space-y-4">
                <div
                  v-for="authSource in authSourceDefaultsMeta"
                  :key="authSource.source"
                  class="rounded-surface border border-gray-200 p-4 dark:border-dark-700"
                >
                  <div class="flex items-center justify-between gap-4">
                    <div>
                      <div class="font-medium text-gray-900 dark:text-white">
                        {{ authSource.title }}
                      </div>
                      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                        {{ authSource.description }}
                      </p>
                    </div>
                    <Toggle
                      v-model="
                        authSourceDefaults[authSource.source].grant_on_signup
                      "
                      :data-testid="`auth-source-${authSource.source}-enabled`"
                    />
                  </div>

                  <div
                    v-if="authSourceDefaults[authSource.source].grant_on_signup"
                    :data-testid="`auth-source-${authSource.source}-panel`"
                    class="mt-4 space-y-4 border-t border-gray-100 pt-4 dark:border-dark-700"
                  >
                    <p class="text-sm text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.authSourceDefaults.enabledHint") }}
                    </p>

                    <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                      <div>
                        <label
                          class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                        >
                          {{ t("admin.settings.defaults.defaultBalance") }}
                        </label>
                        <input
                          v-model.number="
                            authSourceDefaults[authSource.source].balance
                          "
                          type="number"
                          step="0.01"
                          min="0"
                          class="input"
                          placeholder="0.00"
                        />
                      </div>
                      <div>
                        <label
                          class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                        >
                          {{ t("admin.settings.defaults.defaultConcurrency") }}
                        </label>
                        <input
                          v-model.number="
                            authSourceDefaults[authSource.source].concurrency
                          "
                          type="number"
                          min="1"
                          class="input"
                          placeholder="5"
                        />
                      </div>
                    </div>

                    <div
                      class="flex items-center justify-between rounded-compact border border-gray-200 px-4 py-3 dark:border-dark-700"
                    >
                      <div>
                        <label
                          class="font-medium text-gray-900 dark:text-white"
                        >
                          {{ t("admin.settings.authSourceDefaults.grantOnFirstBindLabel") }}
                        </label>
                        <p
                          class="mt-0.5 text-xs text-gray-500 dark:text-gray-400"
                        >
                          {{ t("admin.settings.authSourceDefaults.grantOnFirstBindHint") }}
                        </p>
                      </div>
                      <Toggle
                        v-model="
                          authSourceDefaults[authSource.source]
                            .grant_on_first_bind
                        "
                      />
                    </div>

                    <RuleListEditor
                      :items="authSourceDefaults[authSource.source].subscriptions"
                      :title="t('admin.settings.authSourceDefaults.defaultSubscriptionsLabel')"
                      :hint="t('admin.settings.authSourceDefaults.defaultSubscriptionsHint')"
                      :add-label="t('admin.settings.defaults.addDefaultSubscription')"
                      :empty-text="t('admin.settings.authSourceDefaults.noSourceSubscriptions')"
                      :add-disabled="subscriptionPlans.length === 0"
                      :test-id="`auth-source-${authSource.source}-subscriptions`"
                      @add="addAuthSourceDefaultSubscription(authSource.source)"
                      @remove="removeAuthSourceDefaultSubscription(authSource.source, $event)"
                    >
                      <template #row="{ item }">
                        <Select
                          v-model="item.plan_id"
                          :aria-label="t('admin.settings.defaults.subscriptionGroup')"
                        :options="defaultSubscriptionPlanOptions"
                          :placeholder="
                            t('admin.settings.defaults.subscriptionGroup')
                          "
                        />
                      </template>
                    </RuleListEditor>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <!-- /Tab: Users -->

        <!-- Tab: Gateway — Claude Code, Scheduling -->
        <div v-show="activeTab === 'gateway'" v-content-reveal="activeTab === 'gateway'" class="space-y-4">
          <!-- Claude Code 版本限制 -->
          <SettingsCard
            v-show="activeGatewaySection === 'anthropic'" v-content-reveal="activeGatewaySection === 'anthropic'"
            data-testid="gateway-card-claude-code"
            :title="t('admin.settings.claudeCode.title')"
            :description="t('admin.settings.claudeCode.description')"
          >
            <SettingsSection>
              <SettingRow
                id="claude-code-min-version"
                field
                label-for="claude-code-min-version"
                :label="t('admin.settings.claudeCode.minVersion')"
                :hint="t('admin.settings.claudeCode.minVersionHint')"
              >
                <input
                  id="claude-code-min-version"
                  v-model="form.min_claude_code_version"
                  type="text"
                  class="input font-mono text-sm"
                  :placeholder="t('admin.settings.claudeCode.minVersionPlaceholder')"
                />
              </SettingRow>
              <SettingRow
                id="claude-code-max-version"
                field
                label-for="claude-code-max-version"
                :label="t('admin.settings.claudeCode.maxVersion')"
                :hint="t('admin.settings.claudeCode.maxVersionHint')"
              >
                <input
                  id="claude-code-max-version"
                  v-model="form.max_claude_code_version"
                  type="text"
                  class="input font-mono text-sm"
                  :placeholder="t('admin.settings.claudeCode.maxVersionPlaceholder')"
                />
              </SettingRow>
            </SettingsSection>
          </SettingsCard>

          <!-- Ollama Cloud 用量 -->
          <SettingsCard
            v-show="activeGatewaySection === 'ollamaCloud'" v-content-reveal="activeGatewaySection === 'ollamaCloud'"
            data-testid="ollama-cloud-usage-global-settings"
            :title="t('admin.settings.ollamaCloudUsage.title')"
            :description="t('admin.settings.ollamaCloudUsage.description')"
          >
            <ContentSkeleton v-if="ollamaCloudUsageLoading" variant="form" :rows="3" />
            <SettingsSection v-else>
              <SettingToggleRow
                id="ollama-cloud-usage-enabled"
                v-model="ollamaCloudUsageForm.enabled"
                :label="t('admin.settings.ollamaCloudUsage.enabled')"
                :hint="t('admin.settings.ollamaCloudUsage.enabledHint')"
                testid="ollama-cloud-usage-global-enabled"
              />
              <Collapse :open="ollamaCloudUsageForm.enabled" unmount-on-hide>
                <SettingsSubpanel>
                  <SettingRow
                    id="ollama-cloud-usage-debounce"
                    field
                    label-for="ollama-cloud-usage-debounce"
                    :label="t('admin.settings.ollamaCloudUsage.debounceMinutes')"
                    :hint="t('admin.settings.ollamaCloudUsage.debounceHint')"
                  >
                    <input
                      id="ollama-cloud-usage-debounce"
                      v-model.number="ollamaCloudUsageForm.debounce_minutes"
                      type="number"
                      min="1"
                      max="60"
                      class="input"
                      data-testid="ollama-cloud-usage-global-debounce"
                      @keydown.enter.prevent="saveOllamaCloudUsageSettings"
                    />
                  </SettingRow>
                  <SettingRow
                    id="ollama-cloud-usage-interval"
                    field
                    label-for="ollama-cloud-usage-interval"
                    :label="t('admin.settings.ollamaCloudUsage.intervalMinutes')"
                    :hint="t('admin.settings.ollamaCloudUsage.intervalHint')"
                  >
                    <input
                      id="ollama-cloud-usage-interval"
                      v-model.number="ollamaCloudUsageForm.interval_minutes"
                      type="number"
                      min="15"
                      max="1440"
                      class="input"
                      data-testid="ollama-cloud-usage-global-interval"
                      @keydown.enter.prevent="saveOllamaCloudUsageSettings"
                    />
                  </SettingRow>
                </SettingsSubpanel>
              </Collapse>
            </SettingsSection>
            <template v-if="!ollamaCloudUsageLoading" #footer>
              <button
                type="button"
                class="btn btn-primary btn-sm h-9"
                :disabled="ollamaCloudUsageSaving"
                data-testid="ollama-cloud-usage-global-save"
                @click="saveOllamaCloudUsageSettings"
              >
                <Icon
                  v-if="ollamaCloudUsageSaving"
                  name="loader"
                  size="sm"
                  :animate-on-hover="false"
                  class="mr-1 h-4 w-4 animate-spin"
                />
                {{ ollamaCloudUsageSaving ? t("common.saving") : t("common.save") }}
              </button>
            </template>
          </SettingsCard>

          <!-- 通用调度 -->
          <SettingsCard
            v-show="activeGatewaySection === 'general'" v-content-reveal="activeGatewaySection === 'general'"
            data-testid="gateway-card-scheduling"
            :title="t('admin.settings.scheduling.title')"
            :description="t('admin.settings.scheduling.description')"
          >
            <div class="settings-section space-y-6" data-testid="gateway-scheduling-general">
              <SettingsSection
                :title="t('admin.settings.scheduling.providerSchedulingThresholdsTitle')"
                :hint="t('admin.settings.scheduling.providerSchedulingThresholdsDescription')"
              >
                <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
                  <div v-for="platform in schedulingThresholdPlatforms" :key="platform">
                    <label
                      :for="`provider-scheduling-threshold-${platform}`"
                      class="input-label font-mono"
                    >
                      {{ platform }}
                    </label>
                    <div class="input-icon-wrap">
                      <input
                        :id="`provider-scheduling-threshold-${platform}`"
                        v-model.number="form.provider_scheduling_thresholds[platform]"
                        type="number"
                        min="1"
                        max="100"
                        step="1"
                        class="input input-has-icon-right"
                        :data-testid="`provider-scheduling-threshold-${platform}`"
                        placeholder="100"
                      />
                      <span class="input-icon-right text-sm text-gray-400 dark:text-dark-400" aria-hidden="true">%</span>
                    </div>
                  </div>
                </div>
                <div class="space-y-1">
                  <p class="input-hint">
                    {{ t("admin.settings.scheduling.providerSchedulingThresholdsRangeHint") }}
                  </p>
                  <p class="input-hint">
                    {{ t("admin.settings.scheduling.providerSchedulingThresholdsDisabledHint") }}
                  </p>
                  <p class="input-hint">
                    {{ t("admin.settings.scheduling.providerSchedulingThresholdsGlobalHint") }}
                  </p>
                </div>
              </SettingsSection>

              <SettingsSection
                :title="t('admin.settings.scheduling.stickyEscapeTitle')"
                :hint="t('admin.settings.scheduling.stickyEscapeDescription')"
              >
                <SettingToggleRow
                  id="advanced-scheduler-sticky-escape-enabled"
                  v-model="form.advanced_scheduler_sticky_escape_enabled"
                  :label="t('admin.settings.scheduling.stickyEscapeEnabled')"
                />
                <Collapse :open="form.advanced_scheduler_sticky_escape_enabled">
                  <SettingsSubpanel>
                    <SettingRow
                      id="advanced-scheduler-sticky-escape-ttft"
                      field
                      label-for="advanced-scheduler-sticky-escape-ttft"
                      :label="t('admin.settings.scheduling.stickyEscapeTTFT')"
                    >
                      <input
                        id="advanced-scheduler-sticky-escape-ttft"
                        v-model="form.advanced_scheduler_sticky_escape_ttft_ms"
                        class="input"
                        inputmode="numeric"
                        type="text"
                        :placeholder="advancedSchedulerPlaceholder('advanced_scheduler_effective_sticky_escape_ttft_ms', '15000')"
                      />
                    </SettingRow>
                    <SettingRow
                      id="advanced-scheduler-sticky-escape-error-rate"
                      field
                      label-for="advanced-scheduler-sticky-escape-error-rate"
                      :label="t('admin.settings.scheduling.stickyEscapeErrorRate')"
                    >
                      <input
                        id="advanced-scheduler-sticky-escape-error-rate"
                        v-model="form.advanced_scheduler_sticky_escape_error_rate"
                        class="input"
                        inputmode="decimal"
                        type="text"
                        :placeholder="advancedSchedulerPlaceholder('advanced_scheduler_effective_sticky_escape_error_rate', '0.5')"
                      />
                    </SettingRow>
                  </SettingsSubpanel>
                </Collapse>
              </SettingsSection>
            </div>

            <div class="settings-section space-y-6" data-testid="gateway-scheduling-general-advanced">
              <SettingsSection
                :title="t('admin.settings.scheduling.advancedTitle')"
                :hint="t('admin.settings.scheduling.advancedDescription')"
              >
                <template #actions>
                  <HelpTooltip
                    trigger="both"
                    placement="bottom"
                    width-class="w-max max-w-[calc(100vw-1.5rem)] sm:w-96"
                    class="shrink-0"
                    data-testid="advanced-scheduler-help"
                  >
                    <template #trigger>
                      <button
                        type="button"
                        class="inline-flex h-7 w-7 items-center justify-center rounded-control text-gray-400 transition-colors hover:bg-gray-100 hover:text-primary-600 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500/40 dark:text-gray-500 dark:hover:bg-dark-800 dark:hover:text-primary-400"
                        :aria-label="t('admin.settings.scheduling.advancedHelp.trigger')"
                        :title="t('admin.settings.scheduling.advancedHelp.trigger')"
                      >
                        <Icon name="questionCircle" size="sm" :stroke-width="1.75" />
                      </button>
                    </template>
                    <div class="space-y-2.5">
                      <p class="font-semibold text-white">
                        {{ t("admin.settings.scheduling.advancedHelp.title") }}
                      </p>
                      <p class="text-gray-200">
                        {{ t("admin.settings.scheduling.advancedHelp.summary") }}
                      </p>
                      <p class="rounded-control bg-white/5 px-2 py-1.5 text-xs text-gray-200">
                        {{ t("admin.settings.scheduling.advancedHelp.formula") }}
                      </p>
                      <ol class="list-decimal space-y-2 pl-4 text-gray-200">
                        <li>
                          <span class="font-medium text-white">
                            {{ t("admin.settings.scheduling.advancedHelp.hardFilterTitle") }}
                          </span>
                          {{ t("admin.settings.scheduling.advancedHelp.hardFilter") }}
                        </li>
                        <li>
                          <span class="font-medium text-white">
                            {{ t("admin.settings.scheduling.advancedHelp.scoreTitle") }}
                          </span>
                          {{ t("admin.settings.scheduling.advancedHelp.score") }}
                        </li>
                        <li>
                          <span class="font-medium text-white">
                            {{ t("admin.settings.scheduling.advancedHelp.feedbackTitle") }}
                          </span>
                          {{ t("admin.settings.scheduling.advancedHelp.feedback") }}
                        </li>
                        <li>
                          <span class="font-medium text-white">
                            {{ t("admin.settings.scheduling.advancedHelp.selectionTitle") }}
                          </span>
                          {{ t("admin.settings.scheduling.advancedHelp.selection") }}
                        </li>
                      </ol>
                      <p class="border-t border-white/10 pt-2 text-gray-300">
                        {{ t("admin.settings.scheduling.advancedHelp.boundary") }}
                      </p>
                    </div>
                  </HelpTooltip>
                </template>
                <SettingToggleRow
                  id="advanced-scheduler-sticky-weighted"
                  v-model="form.advanced_scheduler_sticky_weighted_enabled"
                  :label="t('admin.settings.scheduling.stickyWeightedTitle')"
                  :hint="t('admin.settings.scheduling.stickyWeightedDescription')"
                />
                <SettingToggleRow
                  id="advanced-scheduler-subscription-priority"
                  v-model="form.advanced_scheduler_subscription_priority_enabled"
                  :label="t('admin.settings.scheduling.subscriptionPriorityTitle')"
                  :hint="t('admin.settings.scheduling.subscriptionPriorityDescription')"
                />
              </SettingsSection>

              <SettingsSection
                :title="t('admin.settings.scheduling.ewmaTitle')"
                :hint="t('admin.settings.scheduling.ewmaDescription')"
              >
                <SettingRow
                  id="advanced-scheduler-ewma-error-rate"
                  field
                  label-for="advanced-scheduler-ewma-error-rate"
                  :label="t('admin.settings.scheduling.ewmaErrorRateAlpha')"
                >
                  <input
                    id="advanced-scheduler-ewma-error-rate"
                    v-model="form.advanced_scheduler_ewma_error_rate_alpha"
                    class="input"
                    inputmode="decimal"
                    type="text"
                    :placeholder="advancedSchedulerPlaceholder('advanced_scheduler_effective_ewma_error_rate_alpha', '0.2')"
                  />
                </SettingRow>
                <SettingRow
                  id="advanced-scheduler-ewma-ttft"
                  field
                  label-for="advanced-scheduler-ewma-ttft"
                  :label="t('admin.settings.scheduling.ewmaTTFTAlpha')"
                >
                  <input
                    id="advanced-scheduler-ewma-ttft"
                    v-model="form.advanced_scheduler_ewma_ttft_alpha"
                    class="input"
                    inputmode="decimal"
                    type="text"
                    :placeholder="advancedSchedulerPlaceholder('advanced_scheduler_effective_ewma_ttft_alpha', '0.2')"
                  />
                </SettingRow>
              </SettingsSection>

              <SettingsSection
                :title="t('admin.settings.scheduling.weightsTitle')"
                :hint="t('admin.settings.scheduling.weightsDescription')"
              >
                <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
                  <div v-for="field in advancedSchedulerWeightFields" :key="field.key">
                    <label :for="`advanced-scheduler-weight-${field.key}`" class="input-label">
                      {{ field.label }}
                    </label>
                    <input
                      :id="`advanced-scheduler-weight-${field.key}`"
                      v-model="form[field.key]"
                      class="input"
                      inputmode="decimal"
                      :placeholder="field.placeholder"
                      type="text"
                    />
                  </div>
                </div>
              </SettingsSection>
            </div>
          </SettingsCard>

          <!-- OpenAI 调度：配额自动暂停 -->
          <SettingsCard
            v-show="activeGatewaySection === 'openai'" v-content-reveal="activeGatewaySection === 'openai'"
            data-testid="gateway-card-openai-scheduling"
            :title="t('admin.settings.openaiScheduling.title')"
            :description="t('admin.settings.openaiScheduling.description')"
          >
            <SettingsSection
              :title="t('admin.settings.openaiQuotaAutoPause.title')"
              :hint="t('admin.settings.openaiQuotaAutoPause.description')"
            >
              <SettingRow
                id="openai-quota-auto-pause-5h"
                field
                label-for="openai-quota-auto-pause-5h"
                :label="t('admin.settings.openaiQuotaAutoPause.default5h')"
              >
                <input
                  id="openai-quota-auto-pause-5h"
                  v-model.number="openAIQuotaAutoPause5hPercent"
                  type="number"
                  min="0"
                  max="100"
                  step="0.1"
                  class="input"
                  data-testid="settings-openai-quota-auto-pause-5h"
                />
              </SettingRow>
              <SettingRow
                id="openai-quota-auto-pause-7d"
                field
                label-for="openai-quota-auto-pause-7d"
                :label="t('admin.settings.openaiQuotaAutoPause.default7d')"
                :hint="t('admin.settings.openaiQuotaAutoPause.thresholdHint')"
              >
                <input
                  id="openai-quota-auto-pause-7d"
                  v-model.number="openAIQuotaAutoPause7dPercent"
                  type="number"
                  min="0"
                  max="100"
                  step="0.1"
                  class="input"
                  data-testid="settings-openai-quota-auto-pause-7d"
                />
              </SettingRow>
            </SettingsSection>
          </SettingsCard>

          <!-- Grok 请求转发 -->
          <SettingsCard
            v-show="activeGatewaySection === 'grok'" v-content-reveal="activeGatewaySection === 'grok'"
            data-testid="gateway-forwarding-grok"
            :title="t('admin.settings.gatewayForwarding.grokTitle')"
            :description="t('admin.settings.gatewayForwarding.grokDescription')"
          >
            <SettingsSection>
              <SettingRow
                id="grok-default-text-model"
                field
                label-for="grok-default-text-model"
                :label="t('admin.settings.gatewayForwarding.grokDefaultTextModel')"
                :hint="t('admin.settings.gatewayForwarding.grokDefaultTextModelHint')"
              >
                <input
                  id="grok-default-text-model"
                  v-model.trim="form.grok_default_text_model"
                  type="text"
                  class="input"
                  list="grok-default-text-model-options"
                  data-testid="grok-default-text-model"
                  placeholder="grok-4.5"
                />
                <datalist id="grok-default-text-model-options">
                  <option value="grok-4.5" />
                  <option value="grok-4.3" />
                  <option value="grok-build-0.1" />
                </datalist>
              </SettingRow>
              <SettingRow
                id="grok-default-base-url-mode"
                field
                :label="t('admin.settings.gatewayForwarding.grokDefaultBaseURLMode')"
                :hint="t('admin.settings.gatewayForwarding.grokDefaultBaseURLModeHint')"
              >
                <Select
                  id="grok-default-base-url-mode"
                  v-model="form.grok_default_base_url_mode"
                  :options="grokDefaultBaseURLOptions"
                  :aria-label="t('admin.settings.gatewayForwarding.grokDefaultBaseURLMode')"
                  data-testid="grok-default-base-url-mode"
                />
              </SettingRow>
            </SettingsSection>
          </SettingsCard>

          <!-- Anthropic 请求转发 -->
          <SettingsCard
            v-show="activeGatewaySection === 'anthropic'" v-content-reveal="activeGatewaySection === 'anthropic'"
            data-testid="gateway-forwarding-anthropic"
            :title="t('admin.settings.gatewayForwarding.anthropicTitle')"
            :description="t('admin.settings.gatewayForwarding.anthropicDescription')"
          >
            <SettingsSection>
              <SettingToggleRow
                id="enable-fingerprint-unification"
                v-model="form.enable_fingerprint_unification"
                :label="t('admin.settings.gatewayForwarding.fingerprintUnification')"
                :hint="t('admin.settings.gatewayForwarding.fingerprintUnificationHint')"
              />
              <SettingToggleRow
                id="enable-metadata-passthrough"
                v-model="form.enable_metadata_passthrough"
                :label="t('admin.settings.gatewayForwarding.metadataPassthrough')"
                :hint="t('admin.settings.gatewayForwarding.metadataPassthroughHint')"
              />
              <SettingToggleRow
                id="enable-anthropic-cache-ttl-1h-injection"
                v-model="form.enable_anthropic_cache_ttl_1h_injection"
                :label="t('admin.settings.gatewayForwarding.anthropicCacheTTL1hInjection')"
                :hint="t('admin.settings.gatewayForwarding.anthropicCacheTTL1hInjectionHint')"
              />
              <SettingToggleRow
                id="rewrite-message-cache-control"
                v-model="form.rewrite_message_cache_control"
                :label="t('admin.settings.gatewayForwarding.rewriteMessageCacheControl')"
                :hint="t('admin.settings.gatewayForwarding.rewriteMessageCacheControlHint')"
              />
              <!-- 只对 Anthropic OAuth 和 Setup Token 请求生效。 -->
              <SettingToggleRow
                id="enable-client-dateline-normalization"
                v-model="form.enable_client_dateline_normalization"
                :label="t('admin.settings.gatewayForwarding.clientDatelineNormalization')"
                :hint="t('admin.settings.gatewayForwarding.clientDatelineNormalizationHint')"
              />
            </SettingsSection>
            <SettingsSection>
              <SettingToggleRow
                id="enable-claude-oauth-system-prompt-injection"
                v-model="form.enable_claude_oauth_system_prompt_injection"
                :label="t('admin.settings.gatewayForwarding.claudeOAuthSystemPromptInjection')"
                :hint="t('admin.settings.gatewayForwarding.claudeOAuthSystemPromptInjectionHint')"
              />
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label for="claude-oauth-system-prompt" class="input-label">
                    {{ t("admin.settings.gatewayForwarding.claudeOAuthSystemPrompt") }}
                  </label>
                  <textarea
                    id="claude-oauth-system-prompt"
                    v-model="form.claude_oauth_system_prompt"
                    rows="6"
                    class="input min-h-32 font-mono text-sm"
                    :placeholder="t('admin.settings.gatewayForwarding.claudeOAuthSystemPromptPlaceholder')"
                  />
                  <p class="input-hint">
                    {{ t("admin.settings.gatewayForwarding.claudeOAuthSystemPromptHint") }}
                  </p>
                </div>
                <div>
                  <label for="claude-oauth-system-prompt-blocks" class="input-label">
                    {{ t("admin.settings.gatewayForwarding.claudeOAuthSystemPromptBlocks") }}
                  </label>
                  <textarea
                    id="claude-oauth-system-prompt-blocks"
                    v-model="form.claude_oauth_system_prompt_blocks"
                    rows="6"
                    class="input min-h-32 font-mono text-sm"
                    :placeholder="t('admin.settings.gatewayForwarding.claudeOAuthSystemPromptBlocksPlaceholder')"
                  />
                  <p class="input-hint">
                    {{ t("admin.settings.gatewayForwarding.claudeOAuthSystemPromptBlocksHint") }}
                  </p>
                </div>
              </div>
            </SettingsSection>
          </SettingsCard>

          <!-- Antigravity 请求转发 -->
          <SettingsCard
            v-show="activeGatewaySection === 'antigravity'" v-content-reveal="activeGatewaySection === 'antigravity'"
            data-testid="gateway-forwarding-antigravity"
            :title="t('admin.settings.gatewayForwarding.antigravityTitle')"
            :description="t('admin.settings.gatewayForwarding.antigravityDescription')"
          >
            <SettingsSection>
              <SettingRow
                id="antigravity-user-agent-version"
                field
                label-for="antigravity-user-agent-version"
                :label="t('admin.settings.gatewayForwarding.antigravityUserAgentVersion')"
                :hint="t('admin.settings.gatewayForwarding.antigravityUserAgentVersionHint')"
              >
                <input
                  id="antigravity-user-agent-version"
                  v-model="form.antigravity_user_agent_version"
                  type="text"
                  class="input font-mono text-sm"
                  :placeholder="t('admin.settings.gatewayForwarding.antigravityUserAgentVersionPlaceholder')"
                />
              </SettingRow>
            </SettingsSection>
          </SettingsCard>

          <!-- OpenAI / Codex 请求转发 -->
          <SettingsCard
            v-show="activeGatewaySection === 'openai'" v-content-reveal="activeGatewaySection === 'openai'"
            data-testid="gateway-forwarding-openai"
            :title="t('admin.settings.gatewayForwarding.openaiTitle')"
            :description="t('admin.settings.gatewayForwarding.openaiDescription')"
          >
            <SettingsSection>
              <!-- OpenAI Responses 首 token 的计时方式。 -->
              <SettingRow
                id="openai-ttft-mode"
                field
                :label="t('admin.settings.gatewayForwarding.openaiTTFTMode')"
                :hint="t('admin.settings.gatewayForwarding.openaiTTFTModeHint')"
              >
                <Select
                  v-model="form.openai_ttft_mode"
                  :options="openAITTFTModeOptions"
                  :aria-label="t('admin.settings.gatewayForwarding.openaiTTFTMode')"
                  data-testid="openai-ttft-mode"
                />
              </SettingRow>
              <!-- 全局开关：是否允许在 Claude Code 中使用 Codex 插件。 -->
              <SettingToggleRow
                id="openai-allow-claude-code-codex-plugin"
                v-model="form.openai_allow_claude_code_codex_plugin"
                :label="t('admin.settings.gatewayForwarding.openaiAllowClaudeCodeCodexPlugin')"
                :hint="t('admin.settings.gatewayForwarding.openaiAllowClaudeCodeCodexPluginDesc')"
              />
              <div>
                <label for="openai-codex-user-agent" class="input-label">
                  {{ t("admin.settings.gatewayForwarding.openaiCodexUserAgent") }}
                </label>
                <input
                  id="openai-codex-user-agent"
                  v-model="form.openai_codex_user_agent"
                  type="text"
                  class="input font-mono text-sm"
                  :placeholder="t('admin.settings.gatewayForwarding.openaiCodexUserAgentPlaceholder')"
                />
                <p class="input-hint">
                  {{ t("admin.settings.gatewayForwarding.openaiCodexUserAgentHint") }}
                </p>
              </div>
            </SettingsSection>
          </SettingsCard>

          <!-- 跨平台用户提示词替换 -->
          <SettingsCard
            v-show="activeGatewaySection === 'general'" v-content-reveal="activeGatewaySection === 'general'"
            data-testid="gateway-card-user-prompt-replacement"
            :title="t('admin.settings.userPromptReplacement.title')"
            :description="t('admin.settings.userPromptReplacement.description')"
          >
            <template #actions>
              <Toggle
                v-model="form.user_prompt_replacement_config.enabled"
                size="md"
                :aria-label="t('admin.settings.userPromptReplacement.title')"
              />
            </template>
            <SettingsSection>
              <RuleListEditor
                :items="form.user_prompt_replacement_config.rules"
                variant="card"
                :item-label="(index) => t('common.ruleIndex', { index: index + 1 })"
                :add-label="t('admin.settings.userPromptReplacement.addRule')"
                :empty-text="t('admin.settings.userPromptReplacement.empty')"
                test-id="prompt-replacement-rules"
                @add="addUserPromptReplacementRule"
                @remove="removeUserPromptReplacementRule"
              >
                <template #header-actions>
                  <button
                    type="button"
                    class="btn btn-secondary"
                    @click="resetUserPromptReplacementRules"
                  >
                    {{ t("admin.settings.userPromptReplacement.resetDefault") }}
                  </button>
                </template>
                <template #row="{ item: rule, index: ruleIndex }">
                  <div class="space-y-4">
                    <div class="flex min-w-0 items-center gap-3">
                      <Toggle
                        v-model="rule.enabled"
                        size="md"
                        :aria-label="rule.name || t('common.ruleIndex', { index: ruleIndex + 1 })"
                      />
                      <input
                        v-model="rule.name"
                        type="text"
                        class="input min-w-0 flex-1"
                        :aria-label="t('admin.settings.userPromptReplacement.namePlaceholder')"
                        :placeholder="t('admin.settings.userPromptReplacement.namePlaceholder')"
                      />
                    </div>
                    <div class="grid gap-4 lg:grid-cols-6">
                      <div class="lg:col-span-4">
                        <label :for="`prompt-replacement-pattern-${ruleIndex}`" class="input-label">
                          {{ t("admin.settings.userPromptReplacement.pattern") }}
                        </label>
                        <textarea
                          :id="`prompt-replacement-pattern-${ruleIndex}`"
                          v-model="rule.pattern"
                          rows="3"
                          class="input font-mono text-xs"
                          :placeholder="t('admin.settings.userPromptReplacement.patternPlaceholder')"
                        />
                      </div>
                      <div>
                        <label :for="`prompt-replacement-group-${ruleIndex}`" class="input-label">
                          {{ t("admin.settings.userPromptReplacement.targetGroup") }}
                        </label>
                        <input
                          :id="`prompt-replacement-group-${ruleIndex}`"
                          v-model.number="rule.target_group"
                          type="number"
                          min="0"
                          step="1"
                          class="input"
                        />
                      </div>
                      <div>
                        <label :for="`prompt-replacement-type-${ruleIndex}`" class="input-label">
                          {{ t("admin.settings.userPromptReplacement.replacementType") }}
                        </label>
                        <Select
                          :id="`prompt-replacement-type-${ruleIndex}`"
                          v-model="rule.replacement_type"
                          :options="userPromptReplacementTypeOptions"
                        />
                      </div>
                    </div>
                    <div class="grid gap-4 md:grid-cols-3">
                      <div v-if="rule.replacement_type === 'static'">
                        <label :for="`prompt-replacement-static-${ruleIndex}`" class="input-label">
                          {{ t("admin.settings.userPromptReplacement.staticText") }}
                        </label>
                        <input
                          :id="`prompt-replacement-static-${ruleIndex}`"
                          v-model="rule.static_text"
                          type="text"
                          class="input"
                          :placeholder="t('admin.settings.userPromptReplacement.staticTextPlaceholder')"
                        />
                      </div>
                      <div v-if="rule.replacement_type !== 'static'">
                        <label :for="`prompt-replacement-timezone-${ruleIndex}`" class="input-label">
                          {{ t("admin.settings.userPromptReplacement.timezone") }}
                        </label>
                        <Select
                          :id="`prompt-replacement-timezone-${ruleIndex}`"
                          v-model="rule.timezone"
                          :options="userPromptReplacementTimezoneOptions"
                          searchable
                          creatable
                          :creatable-prefix="t('admin.settings.userPromptReplacement.useTimezone')"
                        />
                      </div>
                      <div v-if="rule.replacement_type === 'current_time'">
                        <label :for="`prompt-replacement-format-${ruleIndex}`" class="input-label">
                          {{ t("admin.settings.userPromptReplacement.timeFormat") }}
                        </label>
                        <input
                          :id="`prompt-replacement-format-${ruleIndex}`"
                          v-model="rule.time_format"
                          type="text"
                          class="input font-mono text-sm"
                          placeholder="2006-01-02"
                        />
                      </div>
                    </div>
                  </div>
                </template>
              </RuleListEditor>
            </SettingsSection>
          </SettingsCard>

          <!-- 联网搜索模拟，配置随全局保存提交。 -->
          <SettingsCard
            v-show="activeGatewaySection === 'anthropic'" v-content-reveal="activeGatewaySection === 'anthropic'"
            data-testid="gateway-card-web-search-emulation"
            :title="t('admin.settings.webSearchEmulation.title')"
            :description="t('admin.settings.webSearchEmulation.description')"
          >
            <SettingsSection>
              <SettingToggleRow
                id="web-search-emulation-enabled"
                v-model="webSearchConfig.enabled"
                :label="t('admin.settings.webSearchEmulation.enabled')"
                :hint="t('admin.settings.webSearchEmulation.enabledHint')"
              />
              <Collapse :open="webSearchConfig.enabled" unmount-on-hide>
                <RuleListEditor
                  :items="webSearchConfig.providers"
                  variant="card"
                  :title="t('admin.settings.webSearchEmulation.providers')"
                  :add-label="t('admin.settings.webSearchEmulation.addProvider')"
                  :remove-label="t('admin.settings.webSearchEmulation.removeProvider')"
                  :empty-text="t('admin.settings.webSearchEmulation.noProviders')"
                  test-id="web-search-providers"
                  @add="addWebSearchProvider"
                  @remove="removeWebSearchProvider"
                >
                  <template #row="{ item: provider, index: pIdx }">
                    <!-- 点击标题行展开或收起服务商详情。 -->
                    <div
                      data-icon-trigger
                      class="flex cursor-pointer flex-wrap items-center gap-3"
                      @click="toggleProviderExpand(pIdx)"
                    >
                      <Icon
                        name="chevronRight"
                        size="sm"
                        :animate-on-hover="false"
                        class="h-4 w-4 text-gray-400 transition-transform"
                        :class="{ 'rotate-90': expandedProviders[pIdx] }"
                      />
                      <Select
                        v-model="provider.type"
                        :options="[
                          { value: 'brave', label: 'Brave Search' },
                          { value: 'tavily', label: 'Tavily' },
                        ]"
                        :aria-label="t('admin.settings.webSearchEmulation.providers')"
                        class="w-36"
                        @click.stop
                      />
                      <span class="text-xs text-gray-500 dark:text-dark-400">
                        {{ provider.quota_used ?? 0 }} /
                        {{
                          provider.quota_limit != null && provider.quota_limit > 0
                            ? provider.quota_limit
                            : "∞"
                        }}
                      </span>
                      <span
                        v-if="!expandedProviders[pIdx] && provider.api_key_configured"
                        class="text-xs text-green-600 dark:text-green-400"
                      >
                        {{ t("admin.settings.webSearchEmulation.apiKeyConfigured") }}
                      </span>
                    </div>
                    <Collapse :open="expandedProviders[pIdx]" unmount-on-hide>
                      <div class="mt-4 space-y-4 border-t border-gray-200 pt-4 dark:border-dark-600">
                        <div>
                          <label :for="`web-search-api-key-${pIdx}`" class="input-label">
                            {{ t("admin.settings.webSearchEmulation.apiKey") }}
                          </label>
                          <div class="relative">
                            <input
                              :id="`web-search-api-key-${pIdx}`"
                              v-model="provider.api_key"
                              :type="apiKeyVisible[pIdx] ? 'text' : 'password'"
                              class="input"
                              :class="provider.api_key || provider.api_key_configured ? 'pr-16' : ''"
                              :placeholder="
                                provider.api_key_configured
                                  ? '••••••••'
                                  : t('admin.settings.webSearchEmulation.apiKeyPlaceholder')
                              "
                            />
                            <div
                              v-if="provider.api_key || provider.api_key_configured"
                              class="absolute inset-y-0 right-0 flex items-center pr-1.5"
                            >
                              <button
                                type="button"
                                class="rounded-compact p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                                :title="
                                  apiKeyVisible[pIdx]
                                    ? t('admin.settings.webSearchEmulation.hideApiKey')
                                    : t('admin.settings.webSearchEmulation.showApiKey')
                                "
                                @click="apiKeyVisible[pIdx] = !apiKeyVisible[pIdx]"
                              >
                                <Icon v-if="!apiKeyVisible[pIdx]" name="eye" size="sm" class="h-4 w-4" />
                                <Icon v-else name="eyeOff" size="sm" class="h-4 w-4" />
                              </button>
                              <button
                                type="button"
                                class="rounded-compact p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                                :class="{ 'cursor-not-allowed opacity-30': !provider.api_key }"
                                :title="t('admin.settings.webSearchEmulation.copyApiKey')"
                                :disabled="!provider.api_key"
                                @click="copyApiKey(pIdx)"
                              >
                                <Icon name="copy" size="sm" class="h-4 w-4" />
                              </button>
                            </div>
                          </div>
                        </div>
                        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                          <div>
                            <label :for="`web-search-quota-${pIdx}`" class="input-label">
                              {{ t("admin.settings.webSearchEmulation.quotaLimit") }}
                            </label>
                            <input
                              :id="`web-search-quota-${pIdx}`"
                              v-model="provider.quota_limit"
                              type="number"
                              min="1"
                              class="input"
                              :placeholder="'∞'"
                            />
                            <p class="input-hint">
                              {{ t("admin.settings.webSearchEmulation.quotaLimitHint") }}
                            </p>
                          </div>
                          <div>
                            <label :for="`web-search-subscribed-${pIdx}`" class="input-label">
                              {{ t("admin.settings.webSearchEmulation.subscribedAt") }}
                            </label>
                            <input
                              :id="`web-search-subscribed-${pIdx}`"
                              :value="formatSubscribedAt(provider.subscribed_at)"
                              type="date"
                              class="input"
                              @input="
                                provider.subscribed_at = parseSubscribedAt(
                                  ($event.target as HTMLInputElement).value,
                                )
                              "
                            />
                            <p class="input-hint">
                              {{ t("admin.settings.webSearchEmulation.subscribedAtHint") }}
                            </p>
                          </div>
                        </div>
                        <div class="flex items-center gap-2">
                          <span class="text-xs text-gray-500 dark:text-dark-400"
                            >{{ t("admin.settings.webSearchEmulation.quotaUsage") }}:</span
                          >
                          <div
                            v-if="provider.quota_limit != null && provider.quota_limit > 0"
                            class="h-1.5 flex-1 rounded-full bg-gray-200 dark:bg-dark-600"
                          >
                            <div
                              class="h-full rounded-full transition-[width,background-color]"
                              :class="
                                quotaPercentage(provider) > 90
                                  ? 'bg-red-500'
                                  : quotaPercentage(provider) > 70
                                    ? 'bg-yellow-500'
                                    : 'bg-green-500'
                              "
                              :style="{ width: Math.min(quotaPercentage(provider), 100) + '%' }"
                            />
                          </div>
                          <div v-else class="flex-1" />
                          <span class="text-xs text-gray-500 dark:text-dark-400"
                            >{{ provider.quota_used ?? 0 }} /
                            {{
                              provider.quota_limit != null && provider.quota_limit > 0
                                ? provider.quota_limit
                                : "∞"
                            }}</span
                          >
                          <button
                            v-if="(provider.quota_used ?? 0) > 0"
                            type="button"
                            class="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-500"
                            @click="resetWebSearchUsage(pIdx)"
                          >
                            {{ t("admin.settings.webSearchEmulation.resetUsage") }}
                          </button>
                        </div>
                        <div class="flex items-end gap-3">
                          <div class="min-w-0 flex-1">
                            <p class="input-label">
                              {{ t("admin.settings.webSearchEmulation.proxy") }}
                            </p>
                            <ProxySelector
                              v-model="provider.proxy_id"
                              :proxies="webSearchProxies"
                            />
                          </div>
                          <button
                            type="button"
                            class="btn btn-secondary btn-sm h-9 whitespace-nowrap"
                            @click="openTestDialog()"
                          >
                            {{ t("admin.settings.webSearchEmulation.test") }}
                          </button>
                        </div>
                      </div>
                    </Collapse>
                  </template>
                </RuleListEditor>
              </Collapse>
            </SettingsSection>
          </SettingsCard>

          <!-- 联网搜索测试弹窗 -->
          <BaseDialog
            :show="wsTestDialogOpen"
            :title="t('admin.settings.webSearchEmulation.testResultTitle')"
            width="normal"
            close-on-click-outside
            @close="wsTestDialogOpen = false"
          >
            <div class="flex items-center gap-2">
              <input
                v-model="wsTestQuery"
                type="text"
                class="input flex-1 text-sm"
                :placeholder="t('admin.settings.webSearchEmulation.testDefaultQuery')"
                @keyup.enter="testWebSearchProvider()"
              />
              <button
                type="button"
                class="btn btn-primary btn-sm h-9"
                :disabled="wsTestLoading"
                @click="testWebSearchProvider()"
              >
                {{
                  wsTestLoading
                    ? t("admin.settings.webSearchEmulation.testing")
                    : t("admin.settings.webSearchEmulation.test")
                }}
              </button>
            </div>
            <div
              v-if="wsTestResult"
              class="mt-4 max-h-80 overflow-y-auto rounded-control bg-gray-50 p-4 dark:bg-dark-700"
            >
              <p class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t("admin.settings.webSearchEmulation.testResultProvider") }}: {{ wsTestResult.provider }}
              </p>
              <div
                v-if="wsTestResult.results.length === 0"
                class="text-sm text-gray-400"
              >
                {{ t("admin.settings.webSearchEmulation.testNoResults") }}
              </div>
              <div
                v-for="(r, rIdx) in wsTestResult.results"
                :key="rIdx"
                class="mt-2 border-t border-gray-200 pt-2 first:mt-0 first:border-0 first:pt-0 dark:border-dark-600"
              >
                <a
                  :href="r.url"
                  target="_blank"
                  class="text-sm font-medium text-primary-600 hover:underline dark:text-primary-500"
                  >{{ r.title }}</a
                >
                <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ r.snippet }}
                </p>
              </div>
            </div>
            <div class="mt-4 flex justify-end">
              <button
                type="button"
                class="btn btn-secondary btn-sm h-9"
                @click="wsTestDialogOpen = false"
              >
                {{ t("common.close") }}
              </button>
            </div>
          </BaseDialog>

          <!-- 用量记录 -->
          <SettingsCard
            v-show="activeGatewaySection === 'general'" v-content-reveal="activeGatewaySection === 'general'"
            data-testid="gateway-card-usage-records"
            :title="t('admin.settings.usageRecords.title')"
            :description="t('admin.settings.usageRecords.description')"
          >
            <SettingsSection>
              <SettingToggleRow
                id="allow-user-view-error-requests"
                v-model="form.allow_user_view_error_requests"
                :label="t('admin.settings.user_error_view.label')"
                :hint="t('admin.settings.user_error_view.description')"
              />
            </SettingsSection>
          </SettingsCard>
        </div>
        <!-- /Tab: Gateway — Claude Code, Scheduling -->

        <!-- Tab: General -->
        <div v-show="activeTab === 'general'" v-content-reveal="activeTab === 'general'" class="space-y-4">
          <PreAggregationSettings />

          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.usageRanking.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.usageRanking.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t("admin.settings.usageRanking.enabled") }}
                  </label>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.usageRanking.enabledHint") }}
                  </p>
                </div>
                <Toggle v-model="form.usage_ranking_enabled" />
              </div>

              <div
                class="grid grid-cols-1 gap-6 border-t border-gray-100 pt-5 dark:border-dark-700 md:grid-cols-2"
                :class="!form.usage_ranking_enabled && 'opacity-60'"
              >
                <div>
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t("admin.settings.usageRanking.sortBy") }}
                  </label>
                  <Select
                    v-model="form.usage_ranking_sort_by"
                    :options="usageRankingSortOptions"
                    :disabled="!form.usage_ranking_enabled"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.usageRanking.sortByHint") }}
                  </p>
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t("admin.settings.usageRanking.limit") }}
                  </label>
                  <input
                    v-model.number="form.usage_ranking_limit"
                    type="number"
                    min="1"
                    max="100"
                    step="1"
                    class="input w-40"
                    :disabled="!form.usage_ranking_enabled"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.usageRanking.limitHint") }}
                  </p>
                </div>
              </div>

              <div
                class="space-y-3 border-t border-gray-100 pt-5 dark:border-dark-700"
                :class="!form.usage_ranking_enabled && 'opacity-60'"
              >
                <div>
                  <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t("admin.settings.usageRanking.fields") }}
                  </h3>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.usageRanking.fieldsHint") }}
                  </p>
                </div>
                <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
                  <div class="flex items-center justify-between rounded-control border border-gray-200 px-4 py-3 dark:border-dark-600">
                    <span class="text-sm text-gray-700 dark:text-gray-300">
                      {{ t("admin.settings.usageRanking.totalTokens") }}
                    </span>
                    <Toggle
                      v-model="form.usage_ranking_show_total_tokens"
                      size="sm"
                      :disabled="!form.usage_ranking_enabled || form.usage_ranking_sort_by === 'total_tokens'"
                    />
                  </div>
                  <div class="flex items-center justify-between rounded-control border border-gray-200 px-4 py-3 dark:border-dark-600">
                    <span class="text-sm text-gray-700 dark:text-gray-300">
                      {{ t("admin.settings.usageRanking.requests") }}
                    </span>
                    <Toggle
                      v-model="form.usage_ranking_show_requests"
                      size="sm"
                      :disabled="!form.usage_ranking_enabled || form.usage_ranking_sort_by === 'requests'"
                    />
                  </div>
                  <div class="flex items-center justify-between rounded-control border border-gray-200 px-4 py-3 dark:border-dark-600">
                    <span class="text-sm text-gray-700 dark:text-gray-300">
                      {{ t("admin.settings.usageRanking.actualCost") }}
                    </span>
                    <Toggle
                      v-model="form.usage_ranking_show_actual_cost"
                      size="sm"
                      :disabled="!form.usage_ranking_enabled || form.usage_ranking_sort_by === 'actual_cost'"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Balance Display Settings -->
          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.settings.balanceDisplay.title') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.settings.balanceDisplay.description') }}
              </p>
            </div>
            <div class="grid grid-cols-1 gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_280px]">
              <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t('admin.settings.balanceDisplay.unitName') }}
                  </label>
                  <input
                    v-model="form.balance_unit_name"
                    type="text"
                    class="input"
                    :placeholder="t('admin.settings.balanceDisplay.unitNamePlaceholder')"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.settings.balanceDisplay.unitNameHint') }}
                  </p>
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t('admin.settings.balanceDisplay.unitSymbol') }}
                  </label>
                  <input
                    v-model="form.balance_unit_symbol"
                    type="text"
                    class="input"
                    :placeholder="t('admin.settings.balanceDisplay.unitSymbolPlaceholder')"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.settings.balanceDisplay.unitSymbolHint') }}
                  </p>
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t('admin.settings.balanceDisplay.reasoningPointRmbUnitPrice') }}
                  </label>
                  <input
                    v-model.number="form.reasoning_point_rmb_unit_price"
                    type="number"
                    min="0"
                    step="0.0001"
                    class="input"
                    :placeholder="t('admin.settings.balanceDisplay.reasoningPointRmbUnitPricePlaceholder')"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.settings.balanceDisplay.reasoningPointRmbUnitPriceHint') }}
                  </p>
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t('admin.settings.balanceDisplay.usdExchangeRate') }}
                  </label>
                  <input
                    v-model.number="form.usd_exchange_rate"
                    type="number"
                    min="0"
                    step="0.0001"
                    class="input"
                    :placeholder="t('admin.settings.balanceDisplay.usdExchangeRatePlaceholder')"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.settings.balanceDisplay.usdExchangeRateHint') }}
                  </p>
                </div>
                <div class="md:col-span-2">
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t('admin.settings.balanceDisplay.iconSvg') }}
                  </label>
                  <ImageUpload
                    v-model="form.balance_icon_svg"
                    mode="svg"
                    :upload-label="t('admin.settings.balanceDisplay.uploadSvg')"
                    :remove-label="t('admin.settings.balanceDisplay.removeSvg')"
                    :hint="t('admin.settings.balanceDisplay.iconHint')"
                  />
                </div>
              </div>

              <div class="rounded-surface border border-primary-100 bg-primary-50/60 p-5 dark:border-primary-900/40 dark:bg-primary-900/10">
                <p class="text-xs font-medium uppercase tracking-[0.18em] text-primary-600 dark:text-primary-300">
                  {{ t('admin.settings.balanceDisplay.previewLabel') }}
                </p>
                <div class="mt-4 rounded-surface bg-white p-4 shadow-sm dark:bg-dark-800">
                  <div class="flex items-center gap-3">
                    <div class="flex h-11 w-11 items-center justify-center rounded-surface bg-primary-100 text-primary-600 dark:bg-primary-900/30 dark:text-primary-300">
                      <BalanceIcon
                        :svg="form.balance_icon_svg"
                        :use-global-fallback="false"
                        class="h-5 w-5"
                      />
                    </div>
                    <div>
                      <p class="text-xs text-gray-500 dark:text-gray-400">
                        {{ t('admin.settings.balanceDisplay.previewUnit', { unitName: previewBalanceUnitName }) }}
                      </p>
                      <p class="text-xl font-semibold text-gray-900 dark:text-white">
                        {{ previewBalanceAmount }}
                      </p>
                    </div>
                  </div>
                  <p class="mt-4 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.settings.balanceDisplay.previewHint') }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.marketplaceAvailability.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.marketplaceAvailability.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                <div>
                  <label class="input-label">
                    {{ t("admin.settings.marketplaceAvailability.windowDays") }}
                  </label>
                  <input
                    v-model.number="form.marketplace_availability_window_days"
                    type="number"
                    :min="marketplaceAvailabilityWindowDaysMin"
                    :max="marketplaceAvailabilityWindowDaysMax"
                    step="1"
                    class="input"
                  />
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{
                      t("admin.settings.marketplaceAvailability.windowDaysHint", {
                        min: marketplaceAvailabilityWindowDaysMin,
                        max: marketplaceAvailabilityWindowDaysMax,
                      })
                    }}
                  </p>
                </div>

                <div>
                  <label class="input-label">
                    {{ t("admin.settings.marketplaceAvailability.bucketMinutes") }}
                  </label>
                  <input
                    v-model.number="form.marketplace_availability_bucket_minutes"
                    type="number"
                    :min="marketplaceAvailabilityBucketMinutesMin"
                    :max="marketplaceAvailabilityBucketMinutesMax"
                    step="1"
                    class="input"
                  />
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{
                      t("admin.settings.marketplaceAvailability.bucketMinutesHint", {
                        min: marketplaceAvailabilityBucketMinutesMin,
                        max: marketplaceAvailabilityBucketMinutesMax,
                      })
                    }}
                  </p>
                </div>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.marketplaceAvailability.bucketLimitHint") }}
              </p>
            </div>
          </div>

          <!-- Site Settings -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.site.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.site.description") }}
              </p>
            </div>
            <div class="space-y-6 p-6">
              <!-- Backend Mode -->
              <div
                class="flex items-center justify-between rounded-control border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-900/20"
              >
                <div>
                  <h3 class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ t("admin.settings.site.backendMode") }}
                  </h3>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.site.backendModeDescription") }}
                  </p>
                </div>
                <Toggle v-model="form.backend_mode_enabled" />
              </div>

              <!-- 站点文案 -->
              <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
                <h3 class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ t("admin.settings.site.siteCopyTitle") }}
                </h3>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.site.siteCopyDescription") }}
                </p>
                <div class="mt-4 space-y-6">
                  <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.site.siteNameZh") }}
                      </label>
                      <input
                        v-model="form.site_name_zh"
                        type="text"
                        class="input"
                        :placeholder="t('admin.settings.site.siteNameZhPlaceholder')"
                      />
                    </div>
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.site.siteNameEn") }}
                      </label>
                      <input
                        v-model="form.site_name_en"
                        type="text"
                        class="input"
                        :placeholder="t('admin.settings.site.siteNameEnPlaceholder')"
                      />
                    </div>
                  </div>
                  <div>
                    <p class="text-xs text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.site.siteNameHint") }}
                    </p>
                  </div>

                  <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.site.siteTitleZh") }}
                      </label>
                      <input
                        v-model="form.site_title_zh"
                        type="text"
                        class="input"
                        :placeholder="t('admin.settings.site.siteTitleZhPlaceholder')"
                      />
                    </div>
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.site.siteTitleEn") }}
                      </label>
                      <input
                        v-model="form.site_title_en"
                        type="text"
                        class="input"
                        :placeholder="t('admin.settings.site.siteTitleEnPlaceholder')"
                      />
                    </div>
                  </div>
                  <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.site.siteSubtitleZh") }}
                      </label>
                      <textarea
                        v-model="form.site_subtitle_zh"
                        rows="2"
                        class="input min-h-[72px] resize-y"
                        :placeholder="t('admin.settings.site.siteSubtitleZhPlaceholder')"
                      ></textarea>
                    </div>
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ t("admin.settings.site.siteSubtitleEn") }}
                      </label>
                      <textarea
                        v-model="form.site_subtitle_en"
                        rows="2"
                        class="input min-h-[72px] resize-y"
                        :placeholder="t('admin.settings.site.siteSubtitleEnPlaceholder')"
                      ></textarea>
                    </div>
                  </div>
                </div>
              </div>

              <!-- API Base URL -->
              <div>
                <label
                  class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  {{ t("admin.settings.site.apiBaseUrl") }}
                </label>
                <input
                  v-model="form.api_base_url"
                  type="text"
                  class="input font-mono text-sm"
                  :placeholder="t('admin.settings.site.apiBaseUrlPlaceholder')"
                />
                <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.site.apiBaseUrlHint") }}
                </p>
              </div>

              <!-- Global Table Preferences -->
              <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
                <h3 class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ t("admin.settings.site.tablePreferencesTitle") }}
                </h3>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.site.tablePreferencesDescription") }}
                </p>
                <div class="mt-4 grid grid-cols-1 gap-6 md:grid-cols-2">
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.site.tableDefaultPageSize") }}
                    </label>
                    <input
                      v-model.number="form.table_default_page_size"
                      type="number"
                      min="5"
                      max="1000"
                      step="1"
                      class="input w-40"
                    />
                    <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.site.tableDefaultPageSizeHint") }}
                    </p>
                  </div>
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.site.tablePageSizeOptions") }}
                    </label>
                    <input
                      v-model="tablePageSizeOptionsInput"
                      type="text"
                      class="input font-mono text-sm"
                      :placeholder="
                        t('admin.settings.site.tablePageSizeOptionsPlaceholder')
                      "
                    />
                    <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.site.tablePageSizeOptionsHint") }}
                    </p>
                  </div>
                </div>
              </div>

              <!-- Custom Endpoints -->
              <RuleListEditor
                :items="form.custom_endpoints"
                variant="card"
                :title="t('admin.settings.site.customEndpoints.title')"
                :hint="t('admin.settings.site.customEndpoints.description')"
                :item-label="(index) => t('admin.settings.site.customEndpoints.itemLabel', { n: index + 1 })"
                :add-label="t('admin.settings.site.customEndpoints.add')"
                test-id="custom-endpoints"
                @add="addEndpoint"
                @remove="removeEndpoint"
              >
                <template #row="{ item: ep }">
                  <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <div>
                      <label
                        class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400"
                      >
                        {{ t("admin.settings.site.customEndpoints.name") }}
                      </label>
                      <input
                        v-model="ep.name"
                        type="text"
                        class="input text-sm"
                        :placeholder="
                          t(
                            'admin.settings.site.customEndpoints.namePlaceholder',
                          )
                        "
                      />
                    </div>
                    <div>
                      <label
                        class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400"
                      >
                        {{
                          t("admin.settings.site.customEndpoints.endpointUrl")
                        }}
                      </label>
                      <input
                        v-model="ep.endpoint"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t(
                            'admin.settings.site.customEndpoints.endpointUrlPlaceholder',
                          )
                        "
                      />
                    </div>
                    <div class="sm:col-span-2">
                      <label
                        class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400"
                      >
                        {{
                          t(
                            "admin.settings.site.customEndpoints.descriptionLabel",
                          )
                        }}
                      </label>
                      <input
                        v-model="ep.description"
                        type="text"
                        class="input text-sm"
                        :placeholder="
                          t(
                            'admin.settings.site.customEndpoints.descriptionPlaceholder',
                          )
                        "
                      />
                    </div>
                  </div>
                </template>
              </RuleListEditor>

              <!-- Contact Info -->
              <div>
                <label
                  class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  {{ t("admin.settings.site.contactInfo") }}
                </label>
                <input
                  v-model="form.contact_info"
                  type="text"
                  class="input"
                  :placeholder="t('admin.settings.site.contactInfoPlaceholder')"
                />
                <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.site.contactInfoHint") }}
                </p>
              </div>

              <!-- Doc URL -->
              <div>
                <label
                  class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  {{ t("admin.settings.site.docUrl") }}
                </label>
                <input
                  v-model="form.doc_url"
                  type="url"
                  class="input font-mono text-sm"
                  :placeholder="t('admin.settings.site.docUrlPlaceholder')"
                />
                <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.site.docUrlHint") }}
                </p>
              </div>

              <!-- Site Logo Upload -->
              <div>
                <label
                  class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  {{ t("admin.settings.site.siteLogo") }}
                </label>
                <ImageUpload
                  v-model="form.site_logo"
                  mode="image"
                  :upload-label="t('admin.settings.site.uploadImage')"
                  :remove-label="t('admin.settings.site.remove')"
                  :hint="t('admin.settings.site.logoHint')"
                  :max-size="300 * 1024"
                />
              </div>

              <!-- Home Content -->
              <div>
                <label
                  class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  {{ t("admin.settings.site.homeContent") }}
                </label>
                <textarea
                  v-model="form.home_content"
                  rows="6"
                  class="input font-mono text-sm"
                  :placeholder="t('admin.settings.site.homeContentPlaceholder')"
                ></textarea>
                <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.site.homeContentHint") }}
                </p>
                <!-- iframe CSP Warning -->
                <p class="mt-2 text-xs text-amber-600 dark:text-amber-400">
                  {{ t("admin.settings.site.homeContentIframeWarning") }}
                </p>
              </div>

              <!-- Hide CCS Import Button -->
              <div
                class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
              >
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.site.hideCcsImportButton")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.site.hideCcsImportButtonHint") }}
                  </p>
                </div>
                <Toggle v-model="form.hide_ccs_import_button" />
              </div>
            </div>
          </div>

          <!-- Custom Menu Items -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.customMenu.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.customMenu.description") }}
              </p>
            </div>
            <div class="p-6">
              <RuleListEditor
                :items="form.custom_menu_items"
                variant="card"
                :item-label="(index) => t('admin.settings.customMenu.itemLabel', { n: index + 1 })"
                :add-label="t('admin.settings.customMenu.add')"
                :remove-label="t('admin.settings.customMenu.remove')"
                add-placement="footer"
                reorderable
                test-id="custom-menu-items"
                @add="addMenuItem"
                @remove="removeMenuItem"
                @move="(from, to) => moveMenuItem(from, to > from ? 1 : -1)"
              >
                <template #row="{ item }">
                  <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <!-- Label -->
                    <div>
                      <label
                        class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400"
                      >
                        {{ t("admin.settings.customMenu.name") }}
                      </label>
                      <input
                        v-model="item.label"
                        type="text"
                        class="input text-sm"
                        :placeholder="
                          t('admin.settings.customMenu.namePlaceholder')
                        "
                      />
                    </div>
                    <!-- Visibility -->
                    <div>
                      <label
                        class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400"
                      >
                        {{ t("admin.settings.customMenu.visibility") }}
                      </label>
                      <Select v-model="item.visibility" :options="customMenuVisibilityOptions" class="text-sm" />
                    </div>
                    <!-- URL (full width) -->
                    <div class="sm:col-span-2">
                      <label
                        class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400"
                      >
                        {{ t("admin.settings.customMenu.url") }}
                      </label>
                      <input
                        v-model="item.url"
                        type="url"
                        class="input font-mono text-sm"
                        :placeholder="
                          t('admin.settings.customMenu.urlPlaceholder')
                        "
                      />
                    </div>
                    <!-- SVG Icon (full width) -->
                    <div class="sm:col-span-2">
                      <label
                        class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400"
                      >
                        {{ t("admin.settings.customMenu.iconSvg") }}
                      </label>
                      <ImageUpload
                        :model-value="item.icon_svg"
                        mode="svg"
                        size="sm"
                        :upload-label="t('admin.settings.customMenu.uploadSvg')"
                        :remove-label="t('admin.settings.customMenu.removeSvg')"
                        @update:model-value="(v: string) => (item.icon_svg = v)"
                      />
                    </div>
                  </div>
                </template>
              </RuleListEditor>
            </div>
          </div>

          <!-- Home Featured Models -->
          <div class="card">
            <div
              class="flex items-end justify-between gap-4 border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <div>
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t("admin.settings.homeFeaturedModels.title") }}
                </h2>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{
                    t("admin.settings.homeFeaturedModels.description")
                  }}
                </p>
              </div>
              <button
                v-if="form.home_featured_models.length > 0"
                type="button"
                class="btn btn-secondary btn-sm h-9 shrink-0"
                @click="form.home_featured_models = []"
              >
                {{ t("admin.settings.homeFeaturedModels.clear") }}
              </button>
            </div>
            <div class="space-y-4 p-6">
              <RuleListEditor
                :items="form.home_featured_models"
                :add-label="t('admin.settings.homeFeaturedModels.add')"
                :remove-label="t('admin.settings.homeFeaturedModels.remove')"
                :max="homeFeaturedModelsMax"
                :animated="false"
                add-placement="footer"
                reorderable
                test-id="home-featured-models"
                @add="form.home_featured_models.push('')"
                @remove="removeHomeFeaturedModel"
                @move="(from, to) => moveHomeFeaturedModel(from, to > from ? 1 : -1)"
              >
                <template #row="{ index: mIndex }">
                  <Select
                    v-model="form.home_featured_models[mIndex]"
                    :options="homeFeaturedModelOptions"
                    searchable
                    class="min-w-0"
                    :placeholder="t('admin.settings.homeFeaturedModels.select')"
                  />
                </template>
              </RuleListEditor>
              <p
                v-if="homeFeaturedModelOptions.length === 0"
                class="text-xs text-gray-400 dark:text-gray-500"
              >
                {{ t("admin.settings.homeFeaturedModels.empty") }}
              </p>
            </div>
          </div>

          <!-- Footer Settings -->
          <div class="card">
            <div
              class="flex items-end justify-between gap-4 border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <div>
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t("admin.settings.homeFooter.title") }}
                </h2>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{
                    t("admin.settings.homeFooter.description")
                  }}
                </p>
              </div>
              <button
                type="button"
                class="btn btn-secondary btn-sm h-9 shrink-0"
                @click="applyDefaultFooterLinks"
              >
                {{ t("admin.settings.homeFooter.useDefault") }}
              </button>
            </div>
            <div class="space-y-4 p-6">
              <!-- Link groups -->
              <RuleListEditor
                :items="form.footer_links"
                variant="card"
                :item-label="(index) => t('admin.settings.homeFooter.groupIndex', { index: index + 1 })"
                :add-label="t('admin.settings.homeFooter.addGroup')"
                :remove-label="t('admin.settings.homeFooter.removeGroup')"
                add-placement="footer"
                reorderable
                test-id="footer-groups"
                @add="addFooterGroup"
                @remove="removeFooterGroup"
                @move="(from, to) => moveFooterGroup(from, to > from ? 1 : -1)"
              >
                <template #row="{ item: group, index: gIndex }">
                  <div class="space-y-3">
                    <input
                      v-model="group.title"
                      type="text"
                      class="input max-w-xs text-sm font-medium"
                      :placeholder="t('admin.settings.homeFooter.groupTitlePlaceholder')"
                    />
                    <RuleListEditor
                      :items="group.links"
                      :add-label="t('admin.settings.homeFooter.addLink')"
                      :remove-label="t('admin.settings.homeFooter.removeLink')"
                      add-placement="footer"
                      :test-id="`footer-links-${gIndex}`"
                      @add="group.links.push({ label: '', url: '' })"
                      @remove="group.links.splice($event, 1)"
                    >
                      <template #row="{ item: link }">
                        <div class="grid gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(0,2fr)]">
                          <input
                            v-model="link.label"
                            type="text"
                            class="input text-sm"
                            :placeholder="t('admin.settings.homeFooter.linkLabel')"
                          />
                          <input
                            v-model="link.url"
                            type="text"
                            class="input min-w-0 font-mono text-sm"
                            :placeholder="t('admin.settings.homeFooter.linkUrlPlaceholder')"
                          />
                        </div>
                      </template>
                    </RuleListEditor>
                  </div>
                </template>
              </RuleListEditor>

              <!-- Footer text -->
              <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
                <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t("admin.settings.homeFooter.extraText") }}
                </label>
                <textarea
                  v-model="form.footer_text"
                  rows="2"
                  class="input min-h-[64px] resize-y text-sm"
                  :placeholder="t('admin.settings.homeFooter.extraTextPlaceholder')"
                ></textarea>
              </div>
            </div>
          </div>
        </div>
        <!-- /分页：通用设置 -->

        <!-- 分页：登录条款 -->
        <div v-show="activeTab === 'agreement'" v-content-reveal="activeTab === 'agreement'" class="space-y-4">
          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div>
                  <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                    {{ t("admin.settings.loginAgreement.title") }}
                  </h2>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {{
                      t("admin.settings.loginAgreement.description")
                    }}
                  </p>
                </div>
                <div class="flex items-center gap-3">
                  <span class="text-sm text-gray-600 dark:text-gray-300">
                    {{ form.login_agreement_enabled ? t("admin.settings.loginAgreement.enabled") : t("admin.settings.loginAgreement.disabled") }}
                  </span>
                  <Toggle v-model="form.login_agreement_enabled" />
                </div>
              </div>
            </div>

            <div class="space-y-6 p-6">
              <div class="grid grid-cols-1 gap-5 lg:grid-cols-[minmax(0,1fr)_220px]">
                <div>
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t("admin.settings.loginAgreement.mode") }}
                  </label>
                  <div class="grid grid-cols-2 gap-2 rounded-control bg-gray-100 p-1 dark:bg-dark-700">
                    <button
                      type="button"
                      class="inline-flex items-center justify-center gap-2 rounded-control px-3 py-2 text-sm font-medium transition"
                      :class="
                        form.login_agreement_mode === 'modal'
                          ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300'
                          : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                      "
                      @click="form.login_agreement_mode = 'modal'"
                    >
                      <Icon name="shield" size="sm" />
                      {{ t("admin.settings.loginAgreement.modeModal") }}
                    </button>
                    <button
                      type="button"
                      class="inline-flex items-center justify-center gap-2 rounded-control px-3 py-2 text-sm font-medium transition"
                      :class="
                        form.login_agreement_mode === 'checkbox'
                          ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300'
                          : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                      "
                      @click="form.login_agreement_mode = 'checkbox'"
                    >
                      <Icon name="checkCircle" size="sm" :animate-on-hover="false" />
                      {{ t("admin.settings.loginAgreement.modeCheckbox") }}
                    </button>
                  </div>
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{
                      form.login_agreement_mode === "checkbox"
                        ? t("admin.settings.loginAgreement.modeCheckboxHint")
                        : t("admin.settings.loginAgreement.modeModalHint")
                    }}
                  </p>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t("admin.settings.loginAgreement.updatedAt") }}
                  </label>
                  <input
                    v-model="form.login_agreement_updated_at"
                    type="date"
                    class="input"
                  />
                  <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.loginAgreement.updatedAtHint") }}
                  </p>
                </div>
              </div>

              <RuleListEditor
                :items="form.login_agreement_documents"
                :title="t('admin.settings.loginAgreement.documents')"
                :hint="t('admin.settings.loginAgreement.documentsHint')"
                :add-label="t('admin.settings.loginAgreement.addDocument')"
                variant="card"
                :item-label="(index) => t('admin.settings.loginAgreement.documentIndex', { index: index + 1 })"
                :min="form.login_agreement_enabled ? 1 : 0"
                test-id="login-agreement-documents"
                @add="addLoginAgreementDocument"
                @remove="removeLoginAgreementDocument"
              >
                <template #row="{ item: doc, index }">
                  <div class="mb-3 flex items-center justify-between gap-3">
                    <div class="flex min-w-0 items-center gap-3">
                      <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-control bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-200">
                        <Icon
                          :name="
                            index === 1
                              ? 'shield'
                              : index === 2
                                ? 'globe'
                                : index === 3
                                  ? 'cog'
                                  : 'document'
                          "
                          size="sm"
                        />
                      </span>
                      <div class="min-w-0">
                        <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">
                          {{ doc.title || t("admin.settings.loginAgreement.untitledDocument") }}
                        </p>
                        <p class="truncate text-xs text-gray-500 dark:text-gray-400">
                          {{ loginAgreementRoutePath(doc, index) }}
                        </p>
                      </div>
                    </div>
                  </div>
                  <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
                    <div>
                      <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
                        {{ t("admin.settings.loginAgreement.documentTitle") }}
                      </label>
                      <input
                        v-model="doc.title"
                        type="text"
                        class="input text-sm"
                        :placeholder="t('admin.settings.loginAgreement.documentTitlePlaceholder')"
                      />
                    </div>
                    <div>
                      <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
                        {{ t("admin.settings.loginAgreement.documentSlug") }}
                      </label>
                      <div class="flex overflow-hidden rounded-control border border-primary-900/10 bg-white focus-within:border-primary-900/10 focus-within:ring-2 focus-within:ring-black/10 dark:border-dark-600 dark:bg-dark-900 dark:focus-within:border-primary-500 dark:focus-within:ring-primary-500">
                        <span class="inline-flex flex-shrink-0 items-center border-r border-gray-200 bg-gray-50 px-3 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-400">
                          /legal/
                        </span>
                        <input
                          v-model="doc.id"
                          type="text"
                          class="min-w-0 flex-1 border-0 bg-transparent px-3 py-2 text-sm text-gray-900 outline-none placeholder:text-gray-400 focus:ring-0 dark:text-white dark:placeholder:text-dark-500"
                          placeholder="usage-policy"
                        />
                      </div>
                    </div>
                  </div>
                  <div class="mt-3">
                    <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
                      {{ t("admin.settings.loginAgreement.documentContent") }}
                    </label>
                      <textarea
                        v-model="doc.content_md"
                        rows="8"
                        class="input font-mono text-sm"
                        :placeholder="t('admin.settings.loginAgreement.documentContentPlaceholder')"
                      ></textarea>
                  </div>
                </template>
              </RuleListEditor>
            </div>
          </div>
        </div>
        <!-- /分页：登录条款 -->

        <!-- 分页：功能特性 -->
        <div v-show="activeTab === 'features'" v-content-reveal="activeTab === 'features'" class="space-y-4">
          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.features.team.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.features.team.description") }}
              </p>
            </div>
            <div class="p-6">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">
                    {{ t("admin.settings.features.team.enabled") }}
                  </label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.team.enabledHint") }}
                  </p>
                </div>
                <Toggle v-model="form.team_enabled" />
              </div>
            </div>
          </div>

          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.features.creative.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.features.creative.description") }}
              </p>
            </div>
            <div class="p-6">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">
                    {{ t("admin.settings.features.creative.enabled") }}
                  </label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.creative.enabledHint") }}
                  </p>
                </div>
                <Toggle v-model="form.creative_enabled" />
              </div>

              <div class="mt-5 flex flex-wrap items-center justify-between gap-4 border-t border-gray-100 pt-5 dark:border-dark-700">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white" for="creative-worker-count">
                    {{ t("admin.settings.features.creative.workerCount") }}
                  </label>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.creative.workerCountHint") }}
                  </p>
                </div>
                <input
                  id="creative-worker-count"
                  v-model.number="form.creative_worker_count"
                  type="number"
                  min="1"
                  step="1"
                  required
                  class="input w-32"
                />
              </div>

              <!-- 当前 worker 使用情况：进度条 + 忙碌/总数，数据来自运行时状态轮询。 -->
              <div class="mt-4">
                <p class="mb-1.5 text-sm text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.features.creative.workerUsage") }}
                </p>
                <div class="flex items-center gap-3">
                  <div class="h-2.5 flex-1 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
                    <div
                      class="h-full rounded-full bg-emerald-500 transition-[width,background-color] duration-layout"
                      :style="{ width: `${creativeWorkerUsagePercent}%` }"
                    ></div>
                  </div>
                  <span class="shrink-0 text-sm tabular-nums text-gray-900 dark:text-white">
                    {{ creativeWorkerUsageText }}
                  </span>
                </div>
              </div>

              <RuleListEditor
                class="mt-6 border-t border-gray-100 pt-5 dark:border-dark-700"
                :items="form.creative_model_settings"
                :title="t('admin.settings.features.creative.modelSettings.title')"
                :hint="t('admin.settings.features.creative.modelSettings.description')"
                :add-label="t('admin.settings.features.creative.modelSettings.add')"
                :remove-label="t('admin.settings.features.creative.modelSettings.remove')"
                :empty-text="t('admin.settings.features.creative.modelSettings.empty')"
                :add-disabled="creativeModelCandidatesLoading || !creativeModelCandidates.some((candidate) => !form.creative_model_settings.some((item) => creativeModelSettingKey(item) === creativeModelSettingKey(candidate)))"
                test-id="creative-model-settings"
                @add="addCreativeModelSetting"
                @remove="removeCreativeModelSetting"
              >
                <template #header-extra>
                  <p v-if="creativeModelCandidatesLoading" class="mt-4 text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.creative.modelSettings.loading") }}
                  </p>
                  <p v-else-if="creativeModelCandidatesError" class="mt-4 text-sm text-amber-600 dark:text-amber-400">
                    {{ t("admin.settings.features.creative.modelSettings.loadError") }}
                  </p>
                  <div
                    class="hidden grid-cols-[minmax(0,1fr)_auto] items-center gap-2 text-xs font-medium text-gray-500 sm:grid dark:text-dark-300"
                  >
                    <div class="grid grid-cols-2 items-center gap-4">
                      <span>{{ t("admin.settings.features.creative.modelSettings.modelColumn") }}</span>
                      <span>{{ t("admin.settings.features.creative.modelSettings.operationsColumn") }}</span>
                    </div>
                    <span class="w-9" aria-hidden="true"></span>
                  </div>
                </template>
                <template #row="{ item, index }">
                  <div class="grid grid-cols-1 items-center gap-3 sm:grid-cols-2 sm:gap-4">
                    <div class="min-w-0">
                      <Select
                        :model-value="creativeModelSettingKey(item)"
                        :options="creativeModelOptionsForRow(index)"
                        :placeholder="t('admin.settings.features.creative.modelSettings.selectModel')"
                        :searchable="'auto'"
                        class="w-full sm:max-w-xs"
                        @change="onCreativeModelSelected(index, $event)"
                      />
                      <p v-if="!creativeCandidateForSetting(item)" class="mt-2 text-xs text-amber-600 dark:text-amber-400">
                        {{ t("admin.settings.features.creative.modelSettings.unavailableHint") }}
                      </p>
                    </div>
                    <!-- 能力开关：胶囊按钮替代原生复选框，选中态带对勾，禁用态沿用“至少保留一项能力”等约束。 -->
                    <div class="flex flex-wrap items-center gap-2">
                      <button
                        v-for="operation in creativeOperationChoices"
                        :key="operation"
                        type="button"
                        class="inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-medium transition-colors"
                        :class="[
                          item.operations.includes(operation)
                            ? 'border-primary-500/60 bg-primary-50 text-primary-700 dark:border-primary-500/15 dark:bg-primary-500/8 dark:text-primary-500'
                            : 'border-gray-200 text-gray-600 hover:border-gray-300 hover:text-gray-900 dark:border-dark-600 dark:text-dark-300 dark:hover:border-dark-400 dark:hover:text-dark-100',
                          creativeOperationCheckboxDisabled(index, operation) && 'cursor-not-allowed opacity-50',
                        ]"
                        :disabled="creativeOperationCheckboxDisabled(index, operation)"
                        :aria-pressed="item.operations.includes(operation)"
                        @click="toggleCreativeOperation(index, operation, !item.operations.includes(operation))"
                      >
                        <Icon
                          v-if="item.operations.includes(operation)"
                          name="check"
                          size="xs"
                          :animate-on-hover="false"
                        />
                        {{ t(`admin.settings.features.creative.modelSettings.operations.${operation}`) }}
                      </button>
                    </div>
                  </div>
                </template>
              </RuleListEditor>
            </div>
          </div>

          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.features.affiliate.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.features.affiliate.description") }}
              </p>
            </div>
            <div class="space-y-5 p-6">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">
                    {{ t("admin.settings.features.affiliate.enabled") }}
                  </label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.affiliate.enabledHint") }}
                  </p>
                </div>
                <Toggle v-model="form.affiliate_enabled" />
              </div>

              <!-- 管理员充值返利仅在邀请返利总开关开启时可配置。 -->
              <Collapse :open="form.affiliate_enabled" unmount-on-hide>
                <div

                  class="flex items-center justify-between gap-4 border-t border-gray-100 pt-5 dark:border-dark-700"
                >
                  <div>
                    <label class="font-medium text-gray-900 dark:text-white">
                      {{ t("admin.settings.features.affiliate.adminRechargeRebate") }}
                    </label>
                    <p class="text-sm text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.features.affiliate.adminRechargeRebateHint") }}
                    </p>
                  </div>
                  <Toggle v-model="form.affiliate_admin_recharge_enabled" />
                </div>
              </Collapse>

              <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                <div>
                  <label class="input-label">
                    {{ t("admin.settings.features.affiliate.rebateRate") }}
                  </label>
                  <div class="relative">
                    <input
                      v-model.number="form.affiliate_rebate_rate"
                      type="number"
                      min="0"
                      max="100"
                      step="0.01"
                      class="input pr-8"
                      placeholder="20"
                    />
                    <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-gray-400">%</span>
                  </div>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.affiliate.rebateRateHint") }}
                  </p>
                </div>

                <div>
                  <label class="input-label">
                    {{ t("admin.settings.features.affiliate.freezeHours") }}
                  </label>
                  <input
                    v-model.number="form.affiliate_rebate_freeze_hours"
                    type="number"
                    min="0"
                    max="720"
                    step="1"
                    class="input"
                  />
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.affiliate.freezeHoursDesc") }}
                  </p>
                </div>

                <div>
                  <label class="input-label">
                    {{ t("admin.settings.features.affiliate.durationDays") }}
                  </label>
                  <input
                    v-model.number="form.affiliate_rebate_duration_days"
                    type="number"
                    min="0"
                    max="3650"
                    step="1"
                    class="input"
                  />
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.affiliate.durationDaysDesc") }}
                  </p>
                </div>

                <div>
                  <label class="input-label">
                    {{ t("admin.settings.features.affiliate.perInviteeCap", { unitName: previewBalanceUnitName }) }}
                  </label>
                  <input
                    v-model.number="form.affiliate_rebate_per_invitee_cap"
                    type="number"
                    min="0"
                    step="0.01"
                    class="input"
                  />
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.affiliate.perInviteeCapDesc") }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.features.riskControl.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.features.riskControl.description") }}
              </p>
              <p class="mt-1.5 text-xs">
                <router-link
                  to="/admin/risk-control"
                  class="inline-flex items-center gap-1 text-primary-600 hover:underline dark:text-primary-400"
                >
                  {{ t("admin.settings.features.riskControl.configureLink") }}
                  <span aria-hidden="true">-&gt;</span>
                </router-link>
              </p>
            </div>
            <div class="space-y-5 p-6">
              <div class="flex items-center justify-between">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">
                    {{ t("admin.settings.features.riskControl.enabled") }}
                  </label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.features.riskControl.enabledHint") }}
                  </p>
                </div>
                <Toggle v-model="form.risk_control_enabled" />
              </div>

              <div class="border-t border-gray-100 pt-5 dark:border-dark-700">
                <div class="flex items-center justify-between gap-4">
                  <div>
                    <label class="font-medium text-gray-900 dark:text-white">
                      {{ t("admin.settings.features.riskControl.cyberSessionBlockEnabled") }}
                    </label>
                    <p class="text-sm text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.features.riskControl.cyberSessionBlockEnabledHint") }}
                    </p>
                  </div>
                  <Toggle v-model="form.cyber_session_block_enabled" />
                </div>

                <Collapse :open="form.cyber_session_block_enabled" unmount-on-hide>
                  <div

                    class="mt-4 max-w-xs"
                  >
                    <label class="input-label">
                      {{ t("admin.settings.features.riskControl.cyberSessionBlockTTLSeconds") }}
                    </label>
                    <input
                      v-model.number="form.cyber_session_block_ttl_seconds"
                      type="number"
                      min="1"
                      step="1"
                      class="input"
                    />
                    <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.features.riskControl.cyberSessionBlockTTLSecondsHint") }}
                    </p>
                  </div>
                </Collapse>
              </div>
            </div>
          </div>
        </div>
        <!-- /分页：功能特性 -->

        <!-- Tab: Email -->
        <!-- Tab: Payment -->
        <div v-show="activeTab === 'payment'" v-content-reveal="activeTab === 'payment'" class="space-y-4">
          <!-- Payment System Settings -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t("admin.settings.payment.title") }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.payment.description") }}
                <a
                  :href="paymentGuideHref"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="ml-2 inline-flex items-center text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
                >
                  <Icon name="externalLink" size="xs" class="mr-0.5 h-3.5 w-3.5" />
                  {{ t("admin.settings.payment.configGuide") }}
                </a>
              </p>
            </div>
            <div class="space-y-4 p-6">
              <!-- Enable toggle -->
              <div class="flex items-center justify-between">
                <div>
                  <label class="font-medium text-gray-900 dark:text-white">{{
                    t("admin.settings.payment.enabled")
                  }}</label>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.payment.enabledHint") }}
                  </p>
                </div>
                <Toggle v-model="form.payment_enabled" />
              </div>
              <template v-if="form.payment_enabled">
                <!-- Row 1: Product name -->
                <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.productNamePrefix")
                    }}</label
                    ><input
                      v-model="form.payment_product_name_prefix"
                      type="text"
                      class="input"
                      placeholder="TokenRouter"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.productNameSuffix")
                    }}</label
                    ><input
                      v-model="form.payment_product_name_suffix"
                      type="text"
                      class="input"
                      placeholder="CNY"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.preview")
                    }}</label>
                    <div
                      class="rounded-control border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-600 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300"
                    >
                      {{
                        (form.payment_product_name_prefix || "TokenRouter") +
                        " 100 " +
                        (form.payment_product_name_suffix || "CNY")
                      }}
                    </div>
                  </div>
                </div>
                <!-- Row 2: Balance toggle + amounts -->
                <div class="grid grid-cols-2 gap-3 sm:grid-cols-5">
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.minAmount")
                    }}</label
                    ><input
                      :value="form.payment_min_amount || ''"
                      @input="
                        form.payment_min_amount =
                          parseFloat(
                            ($event.target as HTMLInputElement).value,
                          ) || 0
                      "
                      type="number"
                      step="0.01"
                      min="0"
                      class="input"
                      :placeholder="t('admin.settings.payment.noLimit')"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.maxAmount")
                    }}</label
                    ><input
                      :value="form.payment_max_amount || ''"
                      @input="
                        form.payment_max_amount =
                          parseFloat(
                            ($event.target as HTMLInputElement).value,
                          ) || 0
                      "
                      type="number"
                      step="0.01"
                      min="0"
                      class="input"
                      :placeholder="t('admin.settings.payment.noLimit')"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.dailyLimit")
                    }}</label
                    ><input
                      :value="form.payment_daily_limit || ''"
                      @input="
                        form.payment_daily_limit =
                          parseFloat(
                            ($event.target as HTMLInputElement).value,
                          ) || 0
                      "
                      type="number"
                      step="0.01"
                      min="0"
                      class="input"
                      :placeholder="t('admin.settings.payment.noLimit')"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.balanceRechargeMultiplier")
                    }}</label>
                    <input
                      :value="form.payment_balance_recharge_multiplier || ''"
                      @input="
                        form.payment_balance_recharge_multiplier =
                          parseFloat(
                            ($event.target as HTMLInputElement).value,
                          ) || 1
                      "
                      type="number"
                      step="0.01"
                      min="0.01"
                      class="input"
                    />
                    <p class="mt-0.5 text-xs text-gray-400">
                      {{
                        t("admin.settings.payment.balanceRechargeMultiplierHint", { unitName: previewBalanceUnitName })
                      }}
                    </p>
                    <p
                      class="mt-1 text-xs font-medium text-primary-600 dark:text-primary-400"
                    >
                      {{
                        t("admin.settings.payment.balanceRechargePreview", { amount: (Number(form.payment_balance_recharge_multiplier) || 1).toFixed(2), unitName: previewBalanceUnitName })
                      }}
                    </p>
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.subscriptionUsdToCnyRate")
                    }}</label>
                    <input
                      :value="form.payment_subscription_usd_to_cny_rate || ''"
                      @input="
                        form.payment_subscription_usd_to_cny_rate =
                          parseFloat(
                            ($event.target as HTMLInputElement).value,
                          ) || 0
                      "
                      type="number"
                      step="0.01"
                      min="0"
                      class="input"
                      :placeholder="
                        t(
                          'admin.settings.payment.subscriptionUsdToCnyRateDisabled',
                        )
                      "
                    />
                    <p class="mt-0.5 text-xs text-gray-400">
                      {{
                        t("admin.settings.payment.subscriptionUsdToCnyRateHint")
                      }}
                    </p>
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.rechargeFeeRate")
                    }}</label>
                    <div class="relative">
                      <input
                        :value="form.payment_recharge_fee_rate ?? ''"
                        @input="
                          form.payment_recharge_fee_rate = Math.min(
                            100,
                            Math.max(
                              0,
                              Math.round(
                                parseFloat(
                                  ($event.target as HTMLInputElement).value ||
                                    '0',
                                ) * 100,
                              ) / 100,
                            ),
                          )
                        "
                        type="number"
                        step="0.01"
                        min="0"
                        max="100"
                        class="input pr-8"
                      />
                      <span
                        class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400"
                        >%</span
                      >
                    </div>
                    <p class="mt-0.5 text-xs text-gray-400">
                      {{ t("admin.settings.payment.rechargeFeeRateHint") }}
                    </p>
                    <p
                      v-if="(Number(form.payment_recharge_fee_rate) || 0) > 0"
                      class="mt-1 text-xs font-medium text-primary-600 dark:text-primary-400"
                    >
                      {{
                        t("admin.settings.payment.rechargeFeePreview", {
                          fee: (
                            Number(form.payment_recharge_fee_rate) || 0
                          ).toFixed(2),
                        })
                      }}
                    </p>
                  </div>
                  <div class="col-span-2 sm:col-span-5">
                    <div class="rounded-control border border-gray-200 dark:border-dark-700">
                      <div class="border-b border-gray-100 px-4 py-3 dark:border-dark-700">
                        <p
                          class="text-sm font-semibold text-gray-900 dark:text-white"
                        >
                          {{ t("admin.settings.payment.methodFeesTitle") }}
                        </p>
                        <p class="mt-0.5 text-xs text-gray-400">
                          {{ t("admin.settings.payment.methodFeesHint") }}
                        </p>
                      </div>
                      <div
                        class="hidden grid-cols-[minmax(120px,1fr)_minmax(140px,180px)_minmax(140px,180px)_minmax(180px,1.2fr)] gap-4 bg-gray-50 px-4 py-2 text-xs font-medium text-gray-500 dark:bg-dark-800 dark:text-gray-400 md:grid"
                      >
                        <span>{{
                          t("admin.settings.payment.paymentMethod")
                        }}</span>
                        <span>{{ t("admin.settings.payment.fixedFee") }}</span>
                        <span>{{ t("admin.settings.payment.feeRate") }}</span>
                        <span>{{
                          t("admin.settings.payment.methodFeeResult")
                        }}</span>
                      </div>
                      <div class="divide-y divide-gray-100 dark:divide-dark-700">
                        <div
                          v-for="method in paymentMethodFeeOptions"
                          :key="method.value"
                          class="grid gap-3 px-4 py-3 md:grid-cols-[minmax(120px,1fr)_minmax(140px,180px)_minmax(140px,180px)_minmax(180px,1.2fr)] md:items-center md:gap-4"
                        >
                          <label
                            class="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-300"
                          >
                            <input
                              type="checkbox"
                              class="rounded-compact border-gray-300 text-primary-600 focus:ring-primary-500"
                              :checked="methodFeeEnabled(method.value)"
                              @change="setMethodFeeEnabled(method.value, ($event.target as HTMLInputElement).checked)"
                            />
                            {{ method.label }}
                          </label>
                          <div class="min-w-0">
                            <label
                              class="mb-1 block text-xs text-gray-500 dark:text-gray-400 md:hidden"
                              >{{ t("admin.settings.payment.fixedFee") }}</label
                            >
                            <div class="relative">
                              <span
                                class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-gray-400"
                                >¥</span
                              >
                              <input
                                type="number"
                                step="0.01"
                                min="0"
                                class="input w-full pl-8"
                                :disabled="!methodFeeEnabled(method.value)"
                                :value="methodFeeValue(method.value, 'fixed_fee')"
                                @input="setMethodFeeValue(method.value, 'fixed_fee', ($event.target as HTMLInputElement).value)"
                              />
                            </div>
                          </div>
                          <div class="min-w-0">
                            <label
                              class="mb-1 block text-xs text-gray-500 dark:text-gray-400 md:hidden"
                              >{{ t("admin.settings.payment.feeRate") }}</label
                            >
                            <div class="relative">
                              <input
                                type="number"
                                step="0.01"
                                min="0"
                                max="100"
                                class="input w-full pr-8"
                                :disabled="!methodFeeEnabled(method.value)"
                                :value="methodFeeValue(method.value, 'fee_rate')"
                                @input="setMethodFeeValue(method.value, 'fee_rate', ($event.target as HTMLInputElement).value)"
                              />
                              <span
                                class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400"
                                >%</span
                              >
                            </div>
                          </div>
                          <div class="min-w-0 text-xs">
                            <p
                              class="mb-1 text-gray-500 dark:text-gray-400 md:hidden"
                            >
                              {{ t("admin.settings.payment.methodFeeResult") }}
                            </p>
                            <p
                              :class="methodFeeEnabled(method.value)
                                ? 'font-medium text-primary-600 dark:text-primary-400'
                                : 'text-gray-400 dark:text-gray-500'"
                            >
                              {{
                                methodFeeEnabled(method.value)
                                  ? methodFeePreview(method.value)
                                  : t("admin.settings.payment.methodFeeFallback")
                              }}
                            </p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div>
                    <label class="input-label"
                      >{{ t("admin.settings.payment.orderTimeout") }}
                      <span class="text-red-500">*</span></label
                    ><input
                      v-model.number="form.payment_order_timeout_minutes"
                      type="number"
                      min="1"
                      class="input"
                      required
                    />
                    <p class="mt-0.5 text-xs text-gray-400">
                      {{ t("admin.settings.payment.orderTimeoutHint") }}
                    </p>
                  </div>
                </div>
                <!-- Row 3: Pending orders + load balance + cancel rate limit (all in one row) -->
                <div class="flex flex-wrap items-end gap-4">
                  <div class="w-28">
                    <label class="input-label">{{
                      t("admin.settings.payment.maxPendingOrders")
                    }}</label
                    ><input
                      v-model.number="form.payment_max_pending_orders"
                      type="number"
                      min="1"
                      class="input"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.loadBalanceStrategy")
                    }}</label>
                    <Select
                      v-model="form.payment_load_balance_strategy"
                      :options="loadBalanceOptions"
                      class="w-40"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.cancelRateLimit")
                    }}</label>
                    <div class="flex items-center gap-2">
                      <Toggle v-model="form.payment_cancel_rate_limit_enabled" variant="flush" on-class="bg-primary-500" />
                      <Select
                        v-model="form.payment_cancel_rate_limit_window_mode"
                        :options="cancelRateLimitModeOptions"
                        class="w-24"
                        :disabled="!form.payment_cancel_rate_limit_enabled"
                      />
                      <span
                        :class="[
                          'text-sm whitespace-nowrap',
                          form.payment_cancel_rate_limit_enabled
                            ? 'text-gray-700 dark:text-gray-300'
                            : 'text-gray-400 dark:text-gray-600',
                        ]"
                        >{{
                          t("admin.settings.payment.cancelRateLimitEvery")
                        }}</span
                      >
                      <input
                        v-model.number="form.payment_cancel_rate_limit_window"
                        type="number"
                        min="1"
                        required
                        class="input w-14 text-center"
                        :disabled="!form.payment_cancel_rate_limit_enabled"
                      />
                      <Select
                        v-model="form.payment_cancel_rate_limit_unit"
                        :options="cancelRateLimitUnitOptions"
                        class="w-28"
                        :disabled="!form.payment_cancel_rate_limit_enabled"
                      />
                      <span
                        :class="[
                          'text-sm whitespace-nowrap',
                          form.payment_cancel_rate_limit_enabled
                            ? 'text-gray-700 dark:text-gray-300'
                            : 'text-gray-400 dark:text-gray-600',
                        ]"
                        >{{
                          t("admin.settings.payment.cancelRateLimitAllowMax")
                        }}</span
                      >
                      <input
                        v-model.number="form.payment_cancel_rate_limit_max"
                        type="number"
                        min="1"
                        required
                        class="input w-14 text-center"
                        :disabled="!form.payment_cancel_rate_limit_enabled"
                      />
                      <span
                        :class="[
                          'text-sm whitespace-nowrap',
                          form.payment_cancel_rate_limit_enabled
                            ? 'text-gray-700 dark:text-gray-300'
                            : 'text-gray-400 dark:text-gray-600',
                        ]"
                        >{{
                          t("admin.settings.payment.cancelRateLimitTimes")
                        }}</span
                      >
                    </div>
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.alipayForceQRCode")
                    }}</label>
                    <div class="flex items-center gap-2">
                      <Toggle
                        :model-value="!!form.payment_alipay_force_qrcode"
                        variant="flush"
                        on-class="bg-primary-500"
                        @update:model-value="form.payment_alipay_force_qrcode = $event"
                      />
                      <span class="text-sm text-gray-500 dark:text-gray-400">{{
                        t("admin.settings.payment.alipayForceQRCodeHint")
                      }}</span>
                    </div>
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.alipayMobilePrecreateDeepLink")
                    }}</label>
                    <div class="flex items-center gap-2">
                      <Toggle
                        :model-value="!!form.payment_alipay_mobile_precreate_deep_link"
                        variant="flush"
                        on-class="bg-primary-500"
                        @update:model-value="form.payment_alipay_mobile_precreate_deep_link = $event"
                      />
                      <span class="text-sm text-gray-500 dark:text-gray-400">{{
                        t("admin.settings.payment.alipayMobilePrecreateDeepLinkHint")
                      }}</span>
                    </div>
                  </div>
                </div>
                <!-- Row 4: Enabled payment types (provider badges like sub2apipay) -->
                <div>
                  <label class="input-label">{{
                    t("admin.settings.payment.enabledPaymentTypes")
                  }}</label>
                  <div class="mt-1.5 flex flex-wrap gap-2">
                    <button
                      v-for="pt in allPaymentTypes"
                      :key="pt.value"
                      type="button"
                      @click="togglePaymentType(pt.value)"
                      :class="[
                        'rounded-control border px-3 py-1.5 text-sm font-medium transition',
                        isPaymentTypeEnabled(pt.value)
                          ? 'border-primary-600 bg-primary-600 text-white shadow-sm'
                          : 'border-gray-300 bg-white text-gray-600 hover:border-gray-400 hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:border-dark-500',
                      ]"
                    >
                      {{ pt.label }}
                    </button>
                  </div>
                  <p class="mt-2 text-xs text-gray-400 dark:text-gray-500">
                    {{ t("admin.settings.payment.enabledPaymentTypesHint") }}
                    <a
                      :href="paymentMethodsHref"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="ml-1 text-primary-500 hover:text-primary-600 dark:text-primary-400 dark:hover:text-primary-300"
                    >
                      {{ t("admin.settings.payment.findProvider") }}
                      <Icon name="externalLink" size="xs" class="mb-0.5 ml-0.5 inline h-3 w-3" />
                    </a>
                  </p>
                </div>
                <!-- Row 5: Help image + text -->
                <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.helpImage")
                    }}</label>
                    <ImageUpload
                      v-model="form.payment_help_image_url"
                      :upload-label="t('admin.settings.site.uploadImage')"
                      :remove-label="t('admin.settings.site.remove')"
                      :placeholder="
                        t('admin.settings.payment.helpImagePlaceholder')
                      "
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.settings.payment.helpText")
                    }}</label>
                    <p class="mb-1 text-xs text-gray-400">
                      {{ t("admin.settings.payment.helpTextHint") }}
                    </p>
                    <textarea
                      v-model="form.payment_help_text"
                      rows="3"
                      class="input"
                      :placeholder="
                        t('admin.settings.payment.helpTextPlaceholder')
                      "
                    ></textarea>
                  </div>
                </div>
              </template>
            </div>
          </div>

          <!-- Provider Management -->
          <PaymentProviderList
            v-if="form.payment_enabled"
            :providers="providers"
            :loading="providersLoading"
            :can-create="hasAnyPaymentTypeEnabled"
            :enabled-payment-types="form.payment_enabled_types"
            :all-payment-types="allPaymentTypes"
            :redirect-label="t('admin.settings.payment.easypayRedirect')"
            :updating-provider-ids="updatingProviderIds"
            @refresh="loadProviders"
            @create="openCreateProvider"
            @edit="openEditProvider"
            @delete="confirmDeleteProvider"
            @toggle-field="handleToggleField"
            @toggle-type="handleToggleType"
            @reorder="handleReorderProviders"
          />
        </div>

        <div v-show="activeTab === 'email'" v-content-reveal="activeTab === 'email'" class="space-y-4">
          <!-- Email disabled hint - show when email_verify_enabled is off -->
          <div v-if="!form.email_verify_enabled" class="card">
            <div class="p-6">
              <div class="flex items-start gap-3">
                <Icon
                  name="mail"
                  size="md"
                  class="mt-0.5 flex-shrink-0 text-gray-400 dark:text-gray-500"
                />
                <div>
                  <h3 class="font-medium text-gray-900 dark:text-white">
                    {{ t("admin.settings.emailTabDisabledTitle") }}
                  </h3>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.emailTabDisabledHint") }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- SMTP Settings - Only show when email verification is enabled -->
          <Collapse :open="form.email_verify_enabled" unmount-on-hide>
            <div class="card">
              <div
                class="flex items-end justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700"
              >
                <div>
                  <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                    {{ t("admin.settings.smtp.title") }}
                  </h2>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.smtp.description") }}
                  </p>
                </div>
                <button
                  type="button"
                  @click="testSmtpConnection"
                  :disabled="testingSmtp || loadFailed"
                  class="btn btn-secondary btn-sm h-9"
                >
                  <Icon
                    name="loader"
                    size="sm"
                    :animate-on-hover="false"
                    v-if="testingSmtp"
                    class="h-4 w-4 animate-spin"
                  />
                  {{
                    testingSmtp
                      ? t("admin.settings.smtp.testing")
                      : t("admin.settings.smtp.testConnection")
                  }}
                </button>
              </div>
              <div class="space-y-6 p-6">
                <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.smtp.host") }}
                    </label>
                    <input
                      v-model="form.smtp_host"
                      type="text"
                      class="input"
                      :placeholder="t('admin.settings.smtp.hostPlaceholder')"
                    />
                  </div>
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.smtp.port") }}
                    </label>
                    <input
                      v-model.number="form.smtp_port"
                      type="number"
                      min="1"
                      max="65535"
                      class="input"
                      :placeholder="t('admin.settings.smtp.portPlaceholder')"
                    />
                  </div>
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.smtp.username") }}
                    </label>
                    <input
                      v-model="form.smtp_username"
                      type="text"
                      class="input"
                      :placeholder="t('admin.settings.smtp.usernamePlaceholder')"
                    />
                  </div>
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.smtp.password") }}
                    </label>
                    <input
                      v-model="form.smtp_password"
                      type="password"
                      class="input"
                      autocomplete="new-password"
                      autocapitalize="off"
                      spellcheck="false"
                      @keydown="smtpPasswordManuallyEdited = true"
                      @paste="smtpPasswordManuallyEdited = true"
                      :placeholder="
                        form.smtp_password_configured
                          ? t('admin.settings.smtp.passwordConfiguredPlaceholder')
                          : t('admin.settings.smtp.passwordPlaceholder')
                      "
                    />
                    <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                      {{
                        form.smtp_password_configured
                          ? t("admin.settings.smtp.passwordConfiguredHint")
                          : t("admin.settings.smtp.passwordHint")
                      }}
                    </p>
                  </div>
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.smtp.fromEmail") }}
                    </label>
                    <input
                      v-model="form.smtp_from_email"
                      type="email"
                      class="input"
                      :placeholder="t('admin.settings.smtp.fromEmailPlaceholder')"
                    />
                  </div>
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.smtp.fromName") }}
                    </label>
                    <input
                      v-model="form.smtp_from_name"
                      type="text"
                      class="input"
                      :placeholder="t('admin.settings.smtp.fromNamePlaceholder')"
                    />
                  </div>
                </div>

                <!-- Use TLS Toggle -->
                <div
                  class="flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700"
                >
                  <div>
                    <label class="font-medium text-gray-900 dark:text-white">{{
                      t("admin.settings.smtp.useTls")
                    }}</label>
                    <p class="text-sm text-gray-500 dark:text-gray-400">
                      {{ t("admin.settings.smtp.useTlsHint") }}
                    </p>
                  </div>
                  <Toggle v-model="form.smtp_use_tls" />
                </div>
              </div>
            </div>
          </Collapse>

          <!-- Send Test Email - Only show when email verification is enabled -->
          <Collapse :open="form.email_verify_enabled" unmount-on-hide>
            <div class="card">
              <div
                class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
              >
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t("admin.settings.testEmail.title") }}
                </h2>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.testEmail.description") }}
                </p>
              </div>
              <div class="p-6">
                <div class="flex items-end gap-4">
                  <div class="flex-1">
                    <label
                      class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      {{ t("admin.settings.testEmail.recipientEmail") }}
                    </label>
                    <input
                      v-model="testEmailAddress"
                      type="email"
                      class="input"
                      :placeholder="
                        t('admin.settings.testEmail.recipientEmailPlaceholder')
                      "
                    />
                  </div>
                  <button
                    type="button"
                    @click="sendTestEmail"
                    :disabled="
                      sendingTestEmail || !testEmailAddress || loadFailed
                    "
                    class="btn btn-secondary"
                  >
                    <Icon
                      name="loader"
                      size="sm"
                      :animate-on-hover="false"
                      v-if="sendingTestEmail"
                      class="h-4 w-4 animate-spin"
                    />
                    {{
                      sendingTestEmail
                        ? t("admin.settings.testEmail.sending")
                        : t("admin.settings.testEmail.sendTestEmail")
                    }}
                  </button>
                </div>
              </div>
            </div>
          </Collapse>

          <!-- 订阅到期提醒 -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h3 class="text-base font-medium text-gray-900 dark:text-white">
                {{ t("admin.settings.subscriptionExpiryNotify.title") }}
              </h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.subscriptionExpiryNotify.description") }}
              </p>
            </div>
            <div class="px-6 py-6">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <label
                    class="mb-0 block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    {{ t("admin.settings.subscriptionExpiryNotify.enabled") }}
                  </label>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.subscriptionExpiryNotify.enabledHint") }}
                  </p>
                </div>
                <Toggle v-model="form.subscription_expiry_notify_enabled" />
              </div>
            </div>
          </div>

          <EmailTemplateEditor />

          <!-- Balance Low Notification -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h3 class="text-base font-medium text-gray-900 dark:text-white">
                {{ t("admin.settings.balanceNotify.title") }}
              </h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.balanceNotify.description") }}
              </p>
            </div>
            <div class="px-6 py-6 space-y-4">
              <div class="flex items-center justify-between">
                <label
                  class="mb-0 block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t("admin.settings.balanceNotify.enabled") }}</label
                >
                <Toggle v-model="form.balance_low_notify_enabled" />
              </div>
              <Collapse :open="form.balance_low_notify_enabled" unmount-on-hide>
                <div>
                  <label
                    class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >{{ t("admin.settings.balanceNotify.threshold") }}</label
                  >
                  <div class="relative">
                    <span
                      class="input-icon text-gray-400"
                      >{{ previewBalanceUnitSymbol }}</span>
                    <input
                      v-model.number="form.balance_low_notify_threshold"
                      type="number"
                      min="0"
                      step="0.01"
                      class="input input-has-icon input-icon-text"
                    />
                  </div>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t("admin.settings.balanceNotify.thresholdHint") }}
                  </p>
                </div>
              </Collapse>
              <div>
                <label
                  class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t("admin.settings.balanceNotify.rechargeUrl") }}</label
                >
                <input
                  v-model="form.balance_low_notify_recharge_url"
                  type="url"
                  class="input"
                  :placeholder="currentOrigin"
                />
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ t("admin.settings.balanceNotify.rechargeUrlHint") }}
                </p>
              </div>
            </div>
          </div>

          <!-- Provider Quota Notification -->
          <div class="card">
            <div
              class="border-b border-gray-100 px-6 py-4 dark:border-dark-700"
            >
              <h3 class="text-base font-medium text-gray-900 dark:text-white">
                {{ t("admin.settings.quotaNotify.title") }}
              </h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t("admin.settings.quotaNotify.description") }}
              </p>
            </div>
            <div class="px-6 py-6 space-y-4">
              <div class="flex items-center justify-between">
                <label
                  class="mb-0 block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t("admin.settings.quotaNotify.enabled") }}</label
                >
                <Toggle v-model="form.provider_quota_notify_enabled" />
              </div>
              <Collapse :open="form.provider_quota_notify_enabled" unmount-on-hide>
                <RuleListEditor
                  :items="form.provider_quota_notify_emails || []"
                  :title="t('admin.settings.quotaNotify.emails')"
                  :hint="t('admin.settings.quotaNotify.emailsHint')"
                  :add-label="t('admin.settings.quotaNotify.addEmail')"
                  test-id="quota-notify-emails"
                  @add="addQuotaNotifyEmail"
                  @remove="form.provider_quota_notify_emails.splice($event, 1)"
                >
                  <template #row="{ item: entry }">
                    <div class="flex items-center gap-2">
                      <label
                        class="relative inline-flex items-center cursor-pointer shrink-0"
                      >
                        <Toggle
                          :model-value="!entry.disabled"
                          size="sm"
                          off-class="bg-gray-200 dark:bg-gray-600"
                          @update:model-value="entry.disabled = !entry.disabled"
                        />
                      </label>
                      <input
                        v-model="entry.email"
                        type="email"
                        class="input min-w-0 flex-1"
                        :placeholder="
                          t('admin.settings.quotaNotify.emailPlaceholder')
                        "
                      />
                    </div>
                  </template>
                </RuleListEditor>
              </Collapse>
            </div>
          </div>
        </div>
        <!-- /Tab: Email -->

        <!-- Tab: Backup -->
        <div v-show="activeTab === 'backup'" v-content-reveal="activeTab === 'backup'">
          <BackupSettings />
        </div>

        <!-- 有未保存的修改时，保存按钮吸附在视口底部。 -->
        <SettingsSaveBar
          :visible="settingsDirty && !loadFailed"
          :message="t('admin.settings.unsavedChanges')"
        >
          <button
            type="submit"
            :disabled="saving"
            class="btn btn-primary"
          >
            <Icon
              name="loader"
              size="sm"
              :animate-on-hover="false"
              v-if="saving"
              class="h-4 w-4 animate-spin"
            />
            {{
              saving
                ? t("admin.settings.saving")
                : t("admin.settings.saveSettings")
            }}
          </button>
        </SettingsSaveBar>
      </form>

      <!-- Provider dialogs placed outside the settings form to prevent form submission bubbling -->
      <PaymentProviderDialog
        ref="providerDialogRef"
        :show="showProviderDialog"
        :saving="providerSaving"
        :testing="providerTesting"
        :editing="editingProvider"
        :all-key-options="providerKeyOptions"
        :enabled-key-options="enabledProviderKeyOptions"
        :all-payment-types="allPaymentTypes"
        :redirect-label="t('admin.settings.payment.easypayRedirect')"
        @close="showProviderDialog = false"
        @save="handleSaveProvider"
        @test="handleTestProviderDraft"
      />
      <ConfirmDialog
        :show="showDeleteProviderDialog"
        :title="t('admin.settings.payment.deleteProvider')"
        :message="t('admin.settings.payment.deleteProviderConfirm')"
        :confirm-text="t('common.delete')"
        danger
        @confirm="handleDeleteProvider"
        @cancel="showDeleteProviderDialog = false"
      />
      <!-- 关闭 step-up 开关等敏感保存操作触发的 TOTP 二次验证 -->
      <TotpStepUpDialog :controller="settingsStepUp" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import ContentSkeleton from '@/components/common/ContentSkeleton.vue'
import { vContentReveal } from '@/directives/contentReveal'
import Collapse from '@/components/common/Collapse.vue'

import { ref, reactive, computed, onMounted, onUnmounted, nextTick, watch } from "vue";
import SettingsSkeleton from "@/components/admin/SettingsSkeleton.vue";
import SettingRow from "@/components/common/settings/SettingRow.vue";
import SettingToggleRow from "@/components/common/settings/SettingToggleRow.vue";
import SettingsCard from "@/components/common/settings/SettingsCard.vue";
import SettingsSaveBar from "@/components/common/settings/SettingsSaveBar.vue";
import SettingsSection from "@/components/common/settings/SettingsSection.vue";
import SettingsSubpanel from "@/components/common/settings/SettingsSubpanel.vue";
import { useDirtyTracker } from "@/composables/useDirtyTracker";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { adminAPI } from "@/api";
import {
  appendAuthSourceDefaultsToUpdateRequest,
  buildAuthSourceDefaultsState,
  normalizeProviderSchedulingThresholdsMap,
  sanitizeProviderSchedulingThresholdsMap,
  SCHEDULING_THRESHOLD_PLATFORMS,
  defaultWeChatConnectScopesForMode,
  deriveWeChatConnectStoredMode,
  normalizeDefaultSubscriptionSettings,
  resolveWeChatConnectModeCapabilities,
} from "@/api/admin/settings";
import type {
  AuthSourceDefaultsState,
  AuthSourceType,
  SystemSettings,
  UpdateSettingsRequest,
  DefaultSubscriptionSetting,
  OpenAIFastPolicyRule,
  WeChatConnectMode,
  WebSearchEmulationConfig,
  WebSearchProviderConfig,
  WebSearchTestResult,
  PaymentMethodFeeConfig,
  OpenAIQuotaAutoPauseSettings,
  UserPromptReplacementConfig,
  UserPromptReplacementRule,
  UserPromptReplacementType,
  UsageRankingSortBy,
  CreativeModelCandidate,
  CreativeModelSetting,
  CreativeWorkerStatus,
} from "@/api/admin/settings";
import type { CreativeOperation } from "@/api/creative";
import type { LoginAgreementDocument, MarketplaceGroup, NotifyEmailEntry, Proxy } from "@/types";
import type { ProviderInstance, SubscriptionPlan } from "@/types/payment";
import AppLayout from "@/components/layout/AppLayout.vue";
import Icon from "@/components/icons/Icon.vue";
import HelpTooltip from "@/components/common/HelpTooltip.vue";
import ProviderIcon from "@/components/common/ProviderIcon.vue";
import Select from "@/components/common/Select.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import ConfirmDialog from "@/components/common/ConfirmDialog.vue";
import PaymentProviderList from "@/components/payment/PaymentProviderList.vue";
import PaymentProviderDialog from "@/components/payment/PaymentProviderDialog.vue";
import RuleListEditor from "@/components/common/RuleListEditor.vue";
import Toggle from "@/components/common/Toggle.vue";
import ProxySelector from "@/components/common/ProxySelector.vue";
import ImageUpload from "@/components/common/ImageUpload.vue";
import BalanceIcon from "@/components/common/BalanceIcon.vue";
import OpenAIOAuthImportDefaultsSettings from "@/components/admin/provider/OpenAIOAuthImportDefaultsSettings.vue";
import BackupSettings from "@/views/admin/BackupView.vue";
import { useBalanceDisplay } from "@/composables/useBalanceDisplay";
import EmailTemplateEditor from "@/views/admin/settings/EmailTemplateEditor.vue";
import OpenAIFastPolicyUserSelector from "@/views/admin/settings/OpenAIFastPolicyUserSelector.vue";
import PreAggregationSettings from "@/views/admin/settings/PreAggregationSettings.vue";
import { useClipboard } from "@/composables/useClipboard";
import {
  useStepUp,
  isStepUpCancelled,
  isStepUpBlocked,
  stepUpBlockReason,
} from "@/composables/useStepUp";
import TotpStepUpDialog from "@/components/auth/TotpStepUpDialog.vue";
import { extractApiErrorMessage, extractI18nErrorMessage } from "@/utils/apiError";
import { useAppStore } from "@/stores";
import { useAdminSettingsStore } from "@/stores/adminSettings";
import { getMarketplaceModels } from "@/api/marketplace";
import { normalizeVisibleMethod } from "@/components/payment/paymentFlow";
import { MAX_USER_API_KEY_LIMIT } from "@/constants/user";
import {
  isRegistrationEmailSuffixDomainValid,
  normalizeRegistrationEmailSuffixDomain,
  normalizeRegistrationEmailSuffixDomains,
  parseRegistrationEmailSuffixWhitelistInput,
} from "@/utils/registrationEmailPolicy";

const { t, locale } = useI18n();
const route = useRoute();
const appStore = useAppStore();
// 关闭 step-up 开关是敏感操作：后端返回 STEP_UP_REQUIRED 时弹 TOTP 码重试
const settingsStepUp = useStepUp();
const adminSettingsStore = useAdminSettingsStore();

// 支付帮助指向本仓库的支付配置指南。
const paymentGuideHref =
  "https://github.com/TokenFlux/TokenRouter/blob/main/docs/guides/payments/configuration.md";

const paymentMethodsHref =
  "https://github.com/TokenFlux/TokenRouter/blob/main/docs/guides/payments/configuration.md#支持的支付方式";

type SettingsTab =
  | "general"
  | "agreement"
  | "features"
  | "security"
  | "users"
  | "gateway"
  | "payment"
  | "email"
  | "backup";
const settingsTabs = [
  { key: "general" as SettingsTab, icon: "home" as const },
  { key: "agreement" as SettingsTab, icon: "document" as const },
  { key: "features" as SettingsTab, icon: "bolt" as const },
  { key: "security" as SettingsTab, icon: "shield" as const },
  { key: "users" as SettingsTab, icon: "user" as const },
  { key: "gateway" as SettingsTab, icon: "server" as const },
  { key: "payment" as SettingsTab, icon: "creditCard" as const },
  { key: "email" as SettingsTab, icon: "mail" as const },
  { key: "backup" as SettingsTab, icon: "database" as const },
];
// 允许通过路由参数直接定位到指定标签。
const settingsTabKeys = new Set<SettingsTab>(settingsTabs.map((tab) => tab.key));
const initialSettingsTab = settingsTabKeys.has(route.query.tab as SettingsTab)
  ? (route.query.tab as SettingsTab)
  : "general";
const activeTab = ref<SettingsTab>(initialSettingsTab);
const settingsTabsScrollRef = ref<HTMLElement | null>(null);

type GatewaySection =
  | "general"
  | "anthropic"
  | "openai"
  | "grok"
  | "antigravity"
  | "ollamaCloud";
const gatewaySections = [
  { key: "general" as GatewaySection },
  { key: "anthropic" as GatewaySection, providerBrand: "Anthropic" },
  { key: "openai" as GatewaySection, providerBrand: "OpenAI" },
  { key: "grok" as GatewaySection, providerBrand: "Grok" },
  { key: "antigravity" as GatewaySection, providerBrand: "Google" },
  { key: "ollamaCloud" as GatewaySection, providerBrand: "Ollama" },
];
const activeGatewaySection = ref<GatewaySection>("general");
const gatewaySectionsScrollRef = ref<HTMLElement | null>(null);
const gatewayContentStartRef = ref<HTMLElement | null>(null);
const openAITTFTModeOptions = computed(() => [
  {
    value: "semantic",
    label: t("admin.settings.gatewayForwarding.openaiTTFTModeSemantic"),
  },
  {
    value: "visible",
    label: t("admin.settings.gatewayForwarding.openaiTTFTModeVisible"),
  },
]);

// 支持方向键和 Home / End 在标签之间快速切换。
const settingsTabKeyboardActions = {
  ArrowLeft: -1,
  ArrowUp: -1,
  ArrowRight: 1,
  ArrowDown: 1,
  Home: "first",
  End: "last",
} as const;

function selectSettingsTab(tab: SettingsTab): void {
  activeTab.value = tab;
}

function focusSettingsTab(tab: SettingsTab): void {
  window.requestAnimationFrame(() => {
    document.getElementById(`settings-tab-${tab}`)?.focus();
  });
}

function handleSettingsTabKeydown(event: KeyboardEvent, tab: SettingsTab): void {
  const action =
    settingsTabKeyboardActions[
      event.key as keyof typeof settingsTabKeyboardActions
    ];
  if (action === undefined) {
    return;
  }

  event.preventDefault();
  const currentIndex = settingsTabs.findIndex((item) => item.key === tab);
  let nextIndex = currentIndex < 0 ? 0 : currentIndex;

  if (action === "first") {
    nextIndex = 0;
  } else if (action === "last") {
    nextIndex = settingsTabs.length - 1;
  } else {
    nextIndex =
      (nextIndex + action + settingsTabs.length) % settingsTabs.length;
  }

  const nextTab = settingsTabs[nextIndex]?.key;
  if (!nextTab) {
    return;
  }

  selectSettingsTab(nextTab);
  focusSettingsTab(nextTab);
}

function selectGatewaySection(section: GatewaySection): void {
  activeGatewaySection.value = section;
  void nextTick(() => {
    gatewayContentStartRef.value?.scrollIntoView?.({ block: "start" });
  });
}

function focusGatewaySection(section: GatewaySection): void {
  window.requestAnimationFrame(() => {
    document.getElementById(`gateway-section-tab-${section}`)?.focus();
  });
}

// 二级标签沿用顶层标签的键盘交互，方便在横向滚动区域内快速切换。
function handleGatewaySectionKeydown(
  event: KeyboardEvent,
  section: GatewaySection,
): void {
  const action =
    settingsTabKeyboardActions[
      event.key as keyof typeof settingsTabKeyboardActions
    ];
  if (action === undefined) {
    return;
  }

  event.preventDefault();
  const currentIndex = gatewaySections.findIndex(
    (item) => item.key === section,
  );
  let nextIndex = currentIndex < 0 ? 0 : currentIndex;

  if (action === "first") {
    nextIndex = 0;
  } else if (action === "last") {
    nextIndex = gatewaySections.length - 1;
  } else {
    nextIndex =
      (nextIndex + action + gatewaySections.length) % gatewaySections.length;
  }

  const nextSection = gatewaySections[nextIndex]?.key;
  if (!nextSection) {
    return;
  }

  selectGatewaySection(nextSection);
  focusGatewaySection(nextSection);
}

const { copyToClipboard } = useClipboard();

// 按横轴滚动标签容器，页面保持当前滚动位置。
function scrollTabHorizontallyIntoView(
  container: HTMLElement | null,
  activeSelector: string,
): void {
  const tabElement = container?.querySelector<HTMLElement>(activeSelector);
  if (!container || !tabElement) {
    return;
  }

  const containerRect = container.getBoundingClientRect();
  const tabRect = tabElement.getBoundingClientRect();
  const nextScrollLeft = Math.max(
    0,
    container.scrollLeft +
      tabRect.left -
      containerRect.left -
      (container.clientWidth - tabElement.clientWidth) / 2,
  );

  if (typeof container.scrollTo === "function") {
    container.scrollTo({ left: nextScrollLeft, behavior: "smooth" });
    return;
  }

  container.scrollLeft = nextScrollLeft;
}

// 切换标签后将选中项滚动到容器中央。
function scrollActiveSettingsTabIntoView() {
  void nextTick(() => {
    scrollTabHorizontallyIntoView(
      settingsTabsScrollRef.value,
      ".settings-tab-active",
    );
  });
}

watch(activeTab, scrollActiveSettingsTabIntoView, {
  immediate: true,
  flush: "post",
});

// 平台切换后让选中项保持可见，窄屏下不需要手动寻找当前标签。
function scrollActiveGatewaySectionIntoView(): void {
  if (activeTab.value !== "gateway") {
    return;
  }

  void nextTick(() => {
    scrollTabHorizontallyIntoView(
      gatewaySectionsScrollRef.value,
      ".gateway-section-tab-active",
    );
  });
}

watch(
  [activeTab, activeGatewaySection],
  scrollActiveGatewaySectionIntoView,
  { immediate: true, flush: "post" },
);

const loading = ref(true);
const loadFailed = ref(false);
const saving = ref(false);
const testingSmtp = ref(false);
const sendingTestEmail = ref(false);
const smtpPasswordManuallyEdited = ref(false);
const testEmailAddress = ref("");
const registrationEmailSuffixWhitelistTags = ref<string[]>([]);
const registrationEmailSuffixWhitelistDraft = ref("");
const forwardedClientIpHeaderDraft = ref("");
const tablePageSizeOptionsInput = ref("10, 20, 50, 100");
const paymentMethodFeeOptions = [
  { value: "stripe", label: "Stripe" },
  { value: "alipay", label: "支付宝" },
  { value: "wxpay", label: "微信支付" },
] as const;

// Admin API Key 状态
const adminApiKeyLoading = ref(true);
const adminApiKeyExists = ref(false);
const adminApiKeyMasked = ref("");
const adminApiKeyOperating = ref(false);
const newAdminApiKey = ref("");
const subscriptionPlans = ref<SubscriptionPlan[]>([]);

const ollamaCloudUsageLoading = ref(true);
const ollamaCloudUsageSaving = ref(false);
const ollamaCloudUsageForm = reactive({
  enabled: false,
  interval_minutes: 60,
  debounce_minutes: 1,
});

// Overload Cooldown (529) 状态
const overloadCooldownLoading = ref(true);
const overloadCooldownSaving = ref(false);
const overloadCooldownForm = reactive({
  enabled: true,
  cooldown_minutes: 10,
});

// OpenAI OAuth 403 Cooldown 状态
const openAI403CooldownLoading = ref(true);
const openAI403CooldownSaving = ref(false);
const openAI403CooldownForm = reactive({
  enabled: true,
  cooldown_minutes: 10,
  error_on_threshold_enabled: true,
  threshold_count: 3,
  threshold_window_minutes: 180,
});

// Rate Limit Cooldown (429) 状态
const rateLimit429CooldownLoading = ref(true);
const rateLimit429CooldownSaving = ref(false);
const rateLimit429CooldownForm = reactive({
  enabled: true,
  cooldown_seconds: 5,
});

// 面板 API 限流状态
const panelRateLimitLoading = ref(true);
const panelRateLimitSaving = ref(false);
const panelRateLimitForm = reactive({
  enabled: true,
  user_rpm: 240,
  heavy_rpm: 60,
  exempt_admin: true,
  public_ip_rpm: 300,
});

// Stream Timeout 状态
const streamTimeoutLoading = ref(true);
const streamTimeoutSaving = ref(false);
const streamTimeoutForm = reactive({
  enabled: true,
  action: "temp_unsched" as "temp_unsched" | "error" | "none",
  temp_unsched_minutes: 5,
  threshold_count: 3,
  threshold_window_minutes: 10,
});

const streamTimeoutActionOptions = computed(() => [
  {
    value: "temp_unsched",
    label: t("admin.settings.streamTimeout.actionTempUnsched"),
  },
  { value: "error", label: t("admin.settings.streamTimeout.actionError") },
  { value: "none", label: t("admin.settings.streamTimeout.actionNone") },
]);

// Rectifier 状态
const rectifierLoading = ref(true);
const rectifierSaving = ref(false);
const rectifierForm = reactive({
  enabled: true,
  thinking_signature_enabled: true,
  thinking_budget_enabled: true,
  apikey_signature_enabled: false,
  apikey_signature_patterns: [] as string[],
});

// Beta Policy 状态
const betaPolicyLoading = ref(true);
const betaPolicySaving = ref(false);
const betaPolicyForm = reactive({
  rules: [] as Array<{
    beta_token: string;
    action: "pass" | "filter" | "block";
    scope: "all" | "oauth" | "apikey" | "bedrock";
    error_message?: string;
    model_whitelist?: string[];
    fallback_action?: "pass" | "filter" | "block";
    fallback_error_message?: string;
  }>,
});

// OpenAI Fast/Flex Policy 状态
const openaiFastPolicyForm = reactive({
  rules: [] as OpenAIFastPolicyRule[],
});
// 标记 openai_fast_policy_settings 是否已成功从后端加载，
// 避免后端 GET 出错或字段缺失时，保存把默认规则覆盖成空数组。
const openaiFastPolicyLoaded = ref(false);

const tablePageSizeMin = 5;
const tablePageSizeMax = 1000;
const tablePageSizeDefault = 20;
const usageRankingLimitMin = 1;
const usageRankingLimitMax = 100;
const usageRankingLimitDefault = 20;
const usageRankingSortOptions = computed(() => [
  {
    value: "total_tokens" as UsageRankingSortBy,
    label: t("admin.settings.usageRanking.sortOptions.totalTokens"),
  },
  {
    value: "requests" as UsageRankingSortBy,
    label: t("admin.settings.usageRanking.sortOptions.requests"),
  },
  {
    value: "actual_cost" as UsageRankingSortBy,
    label: t("admin.settings.usageRanking.sortOptions.actualCost"),
  },
]);
const marketplaceAvailabilityWindowDaysMin = 1;
const marketplaceAvailabilityWindowDaysMax = 90;
const marketplaceAvailabilityWindowDaysDefault = 7;
const marketplaceAvailabilityBucketMinutesMin = 5;
const marketplaceAvailabilityBucketMinutesMax = 1440;
const marketplaceAvailabilityBucketMinutesDefault = 120;

function defaultLoginAgreementDocuments(): LoginAgreementDocument[] {
  return [
    {
      id: "terms",
      title: "服务条款",
      content_md: "",
    },
    {
      id: "usage-policy",
      title: "使用政策",
      content_md: "",
    },
    {
      id: "supported-regions",
      title: "支持的国家和地区",
      content_md: "",
    },
    {
      id: "service-specific-terms",
      title: "服务特定条款",
      content_md: "",
    },
  ];
}

// 默认规则与后端内置配置保持一致，便于新环境或旧后端缺字段时回显。
function defaultUserPromptReplacementConfig(): UserPromptReplacementConfig {
  return {
    enabled: true,
    rules: [
      {
        id: "environment-context-timezone-japan",
        name: "environment_context timezone -> Asia/Tokyo",
        enabled: true,
        pattern:
          "(?s)(<environment_context\\b[^>]*>.*?<timezone>)([^<]*)(</timezone>.*?</environment_context>)",
        target_group: 2,
        replacement_type: "timezone_name",
        scope: "environment_context",
        timezone: "Asia/Tokyo",
      },
      {
        id: "environment-context-current-date-japan",
        name: "environment_context current_date -> Asia/Tokyo today",
        enabled: true,
        pattern:
          "(?s)(<environment_context\\b[^>]*>.*?<current_date>)([^<]*)(</current_date>.*?</environment_context>)",
        target_group: 2,
        replacement_type: "current_time",
        scope: "environment_context",
        timezone: "Asia/Tokyo",
        time_format: "2006-01-02",
      },
    ],
  };
}

// 后端可能返回旧格式或空值，这里统一整理为表单可直接绑定的结构。
function normalizeUserPromptReplacementConfig(
  raw: UserPromptReplacementConfig | null | undefined,
): UserPromptReplacementConfig {
  const source = raw ?? defaultUserPromptReplacementConfig();
  return {
    enabled: source.enabled !== false,
    rules: Array.isArray(source.rules)
      ? source.rules.map((rule, index) => ({
          id: String(rule.id || `rule-${index + 1}`).trim(),
          name: String(rule.name || rule.id || `Rule ${index + 1}`).trim(),
          enabled: rule.enabled !== false,
          pattern: String(rule.pattern || "").trim(),
          target_group: Number.isFinite(Number(rule.target_group))
            ? Math.max(0, Math.floor(Number(rule.target_group)))
            : 0,
          replacement_type: normalizeUserPromptReplacementType(
            rule.replacement_type,
          ),
          scope: String(rule.scope || "").trim(),
          static_text: String(rule.static_text ?? ""),
          timezone: String(rule.timezone || "Asia/Tokyo").trim(),
          time_format: String(rule.time_format || "2006-01-02").trim(),
        }))
      : [],
  };
}

// 替换类型只允许后端支持的三种内置值，未知值降级为固定文本。
function normalizeUserPromptReplacementType(
  value: unknown,
): UserPromptReplacementType {
  return value === "static" ||
    value === "timezone_name" ||
    value === "current_time"
    ? value
    : "static";
}

// 新规则默认使用固定文本，可直接保存。
function createUserPromptReplacementRule(): UserPromptReplacementRule {
  return {
    id: `rule-${Date.now()}`,
    name: "",
    enabled: true,
    pattern: "",
    target_group: 0,
    replacement_type: "static",
    scope: "",
    static_text: "",
    timezone: "Asia/Tokyo",
    time_format: "2006-01-02",
  };
}

function normalizeLoginAgreementDocumentId(raw: string): string {
  return raw
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/[-_]{2,}/g, "-")
    .replace(/^[-_]+|[-_]+$/g, "");
}

function loginAgreementRoutePath(
  doc: LoginAgreementDocument,
  index: number,
): string {
  const id =
    normalizeLoginAgreementDocumentId(doc.id || doc.title) || `doc-${index + 1}`;
  return `/legal/${id}`;
}

interface DefaultSubscriptionPlanOption {
  value: number;
  label: string;
  description: string | null;
  [key: string]: unknown;
}

type SettingsForm = Omit<
  SystemSettings,
  | "wechat_connect_open_enabled"
  | "wechat_connect_mp_enabled"
  | "wechat_connect_mobile_enabled"
> & {
  smtp_password: string;
  turnstile_secret_key: string;
  tencent_captcha_app_secret_key: string;
  tencent_captcha_cloud_secret_id: string;
  tencent_captcha_cloud_secret_key: string;
  aliyun_captcha_access_key_secret: string;
  linuxdo_connect_client_secret: string;
  dingtalk_connect_client_secret: string;
  wechat_connect_app_secret: string;
  wechat_connect_open_app_secret: string;
  wechat_connect_mp_app_secret: string;
  wechat_connect_mobile_app_secret: string;
  wechat_connect_open_enabled: boolean;
  wechat_connect_mp_enabled: boolean;
  wechat_connect_mobile_enabled: boolean;
  oidc_connect_client_secret: string;
  github_oauth_client_secret: string;
  google_oauth_client_secret: string;
  force_email_on_third_party_signup: boolean;
  advanced_scheduler_sticky_weighted_enabled: boolean;
  advanced_scheduler_subscription_priority_enabled: boolean;
  advanced_scheduler_ewma_error_rate_alpha: string;
  advanced_scheduler_ewma_ttft_alpha: string;
  advanced_scheduler_sticky_escape_enabled: boolean;
  advanced_scheduler_sticky_escape_ttft_ms: string;
  advanced_scheduler_sticky_escape_error_rate: string;
  advanced_scheduler_lb_top_k: string;
  advanced_scheduler_weight_priority: string;
  advanced_scheduler_weight_load: string;
  advanced_scheduler_weight_queue: string;
  advanced_scheduler_weight_error_rate: string;
  advanced_scheduler_weight_ttft: string;
  advanced_scheduler_weight_reset: string;
  advanced_scheduler_weight_quota_headroom: string;
  advanced_scheduler_weight_previous_response: string;
  advanced_scheduler_weight_session_sticky: string;
  openai_provider_quota_auto_pause: OpenAIQuotaAutoPauseSettings;
  // 系统全局平台限额 map；form 内始终归一化为全平台对象（模板非空绑定依赖此不变量）
  provider_scheduling_thresholds: ReturnType<typeof normalizeProviderSchedulingThresholdsMap>;
};

const schedulingThresholdPlatforms = SCHEDULING_THRESHOLD_PLATFORMS;

const form = reactive<SettingsForm>({
  registration_enabled: true,
  email_verify_enabled: false,
  registration_email_suffix_whitelist: [],
  registration_email_normalization: false,
  registration_email_domain_quota_enabled: false,
  user_email_change_enabled: false,
  promo_code_enabled: true,
  invitation_code_enabled: false,
  password_reset_enabled: false,
  totp_enabled: false,
  totp_encryption_key_configured: false,
  session_binding_enabled: false,
  step_up_enabled: false,
  audit_log_retention_days: 180,
  login_agreement_enabled: false,
  login_agreement_mode: "modal",
  login_agreement_updated_at: "2026-03-31",
  login_agreement_documents: defaultLoginAgreementDocuments(),
  default_balance: 0,
  affiliate_enabled: false,
  provider_scheduling_thresholds: normalizeProviderSchedulingThresholdsMap(),
  affiliate_rebate_rate: 20,
  affiliate_rebate_freeze_hours: 0,
  affiliate_rebate_duration_days: 0,
  affiliate_rebate_per_invitee_cap: 0,
  affiliate_admin_recharge_enabled: false,
  default_concurrency: 1,
  default_subscriptions: [],
  balance_unit_name: "USD",
  balance_unit_symbol: "$",
  balance_icon_svg: "",
  reasoning_point_rmb_unit_price: 0,
  usd_exchange_rate: 0,
  marketplace_availability_window_days:
    marketplaceAvailabilityWindowDaysDefault,
  marketplace_availability_bucket_minutes:
    marketplaceAvailabilityBucketMinutesDefault,
  force_email_on_third_party_signup: false,
  default_user_rpm_limit: 0,
  default_user_api_key_limit: 100,
  site_name: "TokenRouter",
  site_logo: "",
  site_subtitle: "Subscription to API Conversion Platform",
  site_name_zh: "",
  site_name_en: "",
  site_title_zh: "",
  site_title_en: "",
  site_subtitle_zh: "",
  site_subtitle_en: "",
  api_base_url: "",
  contact_info: "",
  doc_url: "",
  home_content: "",
  backend_mode_enabled: false,
  hide_ccs_import_button: false,
  payment_enabled: false,
  payment_min_amount: 1,
  payment_max_amount: 10000,
  payment_daily_limit: 50000,
  payment_max_pending_orders: 3,
  payment_order_timeout_minutes: 30,
  payment_balance_disabled: false,
  payment_balance_recharge_multiplier: 1,
  payment_subscription_usd_to_cny_rate: 0,
  payment_recharge_fee_rate: 0,
  payment_method_fees: {},
  payment_enabled_types: [],
  payment_help_image_url: "",
  payment_help_text: "",
  payment_product_name_prefix: "",
  payment_product_name_suffix: "",
  payment_load_balance_strategy: "round-robin",
  payment_cancel_rate_limit_enabled: false,
  payment_cancel_rate_limit_max: 10,
  payment_cancel_rate_limit_window: 1,
  payment_cancel_rate_limit_unit: "day",
  payment_cancel_rate_limit_window_mode: "rolling",
  payment_alipay_force_qrcode: false,
  payment_alipay_mobile_precreate_deep_link: false,
  table_default_page_size: tablePageSizeDefault,
  table_page_size_options: [10, 20, 50, 100],
  usage_ranking_limit: usageRankingLimitDefault,
  usage_ranking_enabled: true,
  usage_ranking_sort_by: "total_tokens" as UsageRankingSortBy,
  usage_ranking_show_total_tokens: true,
  usage_ranking_show_requests: true,
  usage_ranking_show_actual_cost: true,
  custom_menu_items: [] as Array<{
    id: string;
    label: string;
    icon_svg: string;
    url: string;
    visibility: "user" | "admin";
    sort_order: number;
  }>,
  custom_endpoints: [] as Array<{
    name: string;
    endpoint: string;
    description: string;
  }>,
  footer_links: [] as Array<{
    title: string;
    links: Array<{ label: string; url: string }>;
  }>,
  footer_text: "",
  home_featured_models: [] as string[],
  frontend_url: "",
  smtp_host: "",
  smtp_port: 587,
  smtp_username: "",
  smtp_password: "",
  smtp_password_configured: false,
  smtp_from_email: "",
  smtp_from_name: "",
  smtp_use_tls: true,
  // Cloudflare Turnstile
  turnstile_enabled: false,
  turnstile_site_key: "",
  turnstile_secret_key: "",
  turnstile_secret_key_configured: false,
  tencent_captcha_enabled: false,
  tencent_captcha_app_id: "",
  tencent_captcha_app_secret_key: "",
  tencent_captcha_app_secret_key_configured: false,
  tencent_captcha_cloud_secret_id: "",
  tencent_captcha_cloud_secret_id_configured: false,
  tencent_captcha_cloud_secret_key: "",
  tencent_captcha_cloud_secret_key_configured: false,
  tencent_captcha_region: "cn",
  aliyun_captcha_enabled: false,
  aliyun_captcha_access_key_id: "",
  aliyun_captcha_access_key_secret: "",
  aliyun_captcha_access_key_secret_configured: false,
  aliyun_captcha_scene_id: "",
  aliyun_captcha_prefix: "",
  aliyun_captcha_region: "cn",
  api_key_acl_trust_forwarded_ip: true,
  forwarded_client_ip_headers: [],
  // LinuxDo Connect OAuth 登录
  linuxdo_connect_enabled: false,
  linuxdo_connect_client_id: "",
  linuxdo_connect_client_secret: "",
  linuxdo_connect_client_secret_configured: false,
  linuxdo_connect_redirect_url: "",
  // 钉钉 Connect OAuth 登录
  dingtalk_connect_enabled: false,
  dingtalk_connect_client_id: "",
  dingtalk_connect_client_secret: "",
  dingtalk_connect_client_secret_configured: false,
  dingtalk_connect_redirect_url: "",
  dingtalk_connect_corp_restriction_policy: "none",
  dingtalk_connect_internal_corp_id: "",
  dingtalk_connect_bypass_registration: false,
  dingtalk_connect_sync_corp_email: false,
  dingtalk_connect_sync_display_name: false,
  dingtalk_connect_sync_dept: false,
  dingtalk_connect_sync_corp_email_attr_key: "dingtalk_email",
  dingtalk_connect_sync_display_name_attr_key: "dingtalk_name",
  dingtalk_connect_sync_dept_attr_key: "dingtalk_department",
  dingtalk_connect_sync_corp_email_attr_name: "钉钉企业邮箱",
  dingtalk_connect_sync_display_name_attr_name: "钉钉姓名",
  dingtalk_connect_sync_dept_attr_name: "钉钉部门",
  wechat_connect_enabled: false,
  wechat_connect_app_id: "",
  wechat_connect_app_secret: "",
  wechat_connect_app_secret_configured: false,
  wechat_connect_open_app_id: "",
  wechat_connect_open_app_secret: "",
  wechat_connect_open_app_secret_configured: false,
  wechat_connect_mp_app_id: "",
  wechat_connect_mp_app_secret: "",
  wechat_connect_mp_app_secret_configured: false,
  wechat_connect_mobile_app_id: "",
  wechat_connect_mobile_app_secret: "",
  wechat_connect_mobile_app_secret_configured: false,
  wechat_connect_open_enabled: false,
  wechat_connect_mp_enabled: false,
  wechat_connect_mobile_enabled: false,
  wechat_connect_mode: "open",
  wechat_connect_scopes: "snsapi_login",
  wechat_connect_redirect_url: "",
  wechat_connect_frontend_redirect_url: "/auth/wechat/callback",
  // Generic OIDC OAuth 登录
  oidc_connect_enabled: false,
  oidc_connect_provider_name: "OIDC",
  oidc_connect_client_id: "",
  oidc_connect_client_secret: "",
  oidc_connect_client_secret_configured: false,
  oidc_connect_issuer_url: "",
  oidc_connect_discovery_url: "",
  oidc_connect_authorize_url: "",
  oidc_connect_token_url: "",
  oidc_connect_userinfo_url: "",
  oidc_connect_jwks_url: "",
  oidc_connect_scopes: "openid email profile",
  oidc_connect_redirect_url: "",
  oidc_connect_frontend_redirect_url: "/auth/oidc/callback",
  oidc_connect_token_auth_method: "client_secret_post",
  oidc_connect_use_pkce: false,
  oidc_connect_validate_id_token: false,
  oidc_connect_allowed_signing_algs: "RS256,ES256,PS256",
  oidc_connect_clock_skew_seconds: 120,
  oidc_connect_require_email_verified: false,
  oidc_connect_userinfo_email_path: "",
  oidc_connect_userinfo_id_path: "",
  oidc_connect_userinfo_username_path: "",
  // GitHub / Google 邮箱快捷登录
  github_oauth_enabled: false,
  github_oauth_client_id: "",
  github_oauth_client_secret: "",
  github_oauth_client_secret_configured: false,
  github_oauth_redirect_url: "",
  github_oauth_frontend_redirect_url: "/auth/oauth/callback",
  google_oauth_enabled: false,
  google_one_tap_enabled: false,
  google_oauth_client_id: "",
  google_oauth_client_secret: "",
  google_oauth_client_secret_configured: false,
  google_oauth_redirect_url: "",
  google_oauth_frontend_redirect_url: "/auth/oauth/callback",
  // Model fallback
  enable_model_fallback: false,
  fallback_model_anthropic: "claude-3-5-sonnet-20241022",
  fallback_model_openai: "gpt-4o",
  fallback_model_gemini: "gemini-2.5-pro",
  fallback_model_antigravity: "gemini-2.5-pro",
  grok_default_text_model: "grok-4.5",
  grok_default_base_url_mode: "cli",
  // Identity patch (Claude -> Gemini)
  enable_identity_patch: true,
  identity_patch_prompt: "",
  // Ops monitoring (vNext)
  ops_monitoring_enabled: true,
  ops_realtime_monitoring_enabled: true,
  ops_metrics_interval_seconds: 60,
  // Claude Code version check
  min_claude_code_version: "",
  max_claude_code_version: "",
  // 分组隔离
  advanced_scheduler_sticky_weighted_enabled: false,
  advanced_scheduler_subscription_priority_enabled: false,
  advanced_scheduler_ewma_error_rate_alpha: "",
  advanced_scheduler_ewma_ttft_alpha: "",
  advanced_scheduler_sticky_escape_enabled: true,
  advanced_scheduler_sticky_escape_ttft_ms: "",
  advanced_scheduler_sticky_escape_error_rate: "",
  advanced_scheduler_lb_top_k: "",
  advanced_scheduler_weight_priority: "",
  advanced_scheduler_weight_load: "",
  advanced_scheduler_weight_queue: "",
  advanced_scheduler_weight_error_rate: "",
  advanced_scheduler_weight_ttft: "",
  advanced_scheduler_weight_reset: "",
  advanced_scheduler_weight_quota_headroom: "",
  advanced_scheduler_weight_previous_response: "",
  advanced_scheduler_weight_session_sticky: "",
  openai_provider_quota_auto_pause: {
    default_threshold_5h: 0,
    default_threshold_7d: 0,
  },
  // Gateway forwarding behavior
  openai_ttft_mode: "semantic",
  enable_fingerprint_unification: true,
  enable_metadata_passthrough: false,
  enable_cch_signing: false,
  enable_claude_oauth_system_prompt_injection: true,
  claude_oauth_system_prompt: "",
  claude_oauth_system_prompt_blocks: "",
  enable_anthropic_cache_ttl_1h_injection: false,
  rewrite_message_cache_control: false,
  enable_client_dateline_normalization: true,
  // 页面功能开关默认开启，兼容升级前行为。
  team_enabled: true,
  creative_enabled: true,
  creative_model_settings: [] as CreativeModelSetting[],
  creative_worker_count: 128,
  risk_control_enabled: false,
  cyber_session_block_enabled: false,
  cyber_session_block_ttl_seconds: 3600,
  antigravity_user_agent_version: "",
  openai_codex_user_agent: "",
  openai_allow_claude_code_codex_plugin: false,
  user_prompt_replacement_config: defaultUserPromptReplacementConfig(),
  // 余额、订阅到期与提供商限额通知
  balance_low_notify_enabled: false,
  balance_low_notify_threshold: 0,
  balance_low_notify_recharge_url: "",
  subscription_expiry_notify_enabled: true,
  provider_quota_notify_enabled: false,
  provider_quota_notify_emails: [] as NotifyEmailEntry[],
  allow_user_view_error_requests: false,
});

const creativeOperationChoices: CreativeOperation[] = [
  "generate",
  "edit",
  "inpaint",
];
const creativeModelCandidates = ref<CreativeModelCandidate[]>([]);
const creativeModelCandidatesLoading = ref(false);
const creativeModelCandidatesError = ref(false);

// 进入功能标签页时轮询创作台 worker 状态，离开后停止轮询。
const creativeWorkerStatus = ref<CreativeWorkerStatus | null>(null);
let creativeWorkerStatusTimer: number | null = null;

const creativeWorkerUsageTotal = computed(() => {
  const status = creativeWorkerStatus.value;
  if (status?.running && status.worker_count > 0) {
    return status.worker_count;
  }
  return Math.max(0, Math.floor(Number(form.creative_worker_count)) || 0);
});
const creativeWorkerUsageBusy = computed(() => {
  const status = creativeWorkerStatus.value;
  if (!status?.running) {
    return 0;
  }
  return Math.min(Math.max(status.busy_workers, 0), creativeWorkerUsageTotal.value);
});
const creativeWorkerUsagePercent = computed(() => {
  const total = creativeWorkerUsageTotal.value;
  if (total <= 0) {
    return 0;
  }
  return Math.min(Math.round((creativeWorkerUsageBusy.value / total) * 100), 100);
});
const creativeWorkerUsageText = computed(
  () => `${creativeWorkerUsageBusy.value}/${creativeWorkerUsageTotal.value}`,
);

async function loadCreativeWorkerStatus() {
  try {
    creativeWorkerStatus.value = await adminAPI.settings.getCreativeWorkerStatus();
  } catch {
    // 轮询失败静默处理：保留上一次成功快照，不打断设置页操作。
  }
}

function startCreativeWorkerStatusPolling() {
  if (creativeWorkerStatusTimer !== null) {
    return;
  }
  void loadCreativeWorkerStatus();
  creativeWorkerStatusTimer = window.setInterval(() => {
    void loadCreativeWorkerStatus();
  }, 5000);
}

function stopCreativeWorkerStatusPolling() {
  if (creativeWorkerStatusTimer === null) {
    return;
  }
  window.clearInterval(creativeWorkerStatusTimer);
  creativeWorkerStatusTimer = null;
}

watch(
  activeTab,
  (tab) => {
    if (tab === "features") {
      startCreativeWorkerStatusPolling();
    } else {
      stopCreativeWorkerStatusPolling();
    }
  },
  { immediate: true },
);

onUnmounted(() => {
  stopCreativeWorkerStatusPolling();
});

function creativeModelSettingKey(item: Pick<CreativeModelSetting, "group_id" | "model">): string {
  return `${item.group_id}::${item.model}`;
}

const creativeModelCandidateOptions = computed(() => {
  const options: Array<{ value: string; label: string; kind?: string; disabled?: boolean; candidate?: CreativeModelCandidate }> = [];
  const groups = new Map<number, CreativeModelCandidate[]>();
  for (const candidate of creativeModelCandidates.value) {
    const items = groups.get(candidate.group_id) ?? [];
    items.push(candidate);
    groups.set(candidate.group_id, items);
  }
  for (const [groupID, candidates] of groups) {
    const groupName = candidates[0]?.group_name || String(groupID);
    options.push({ value: `__creative_group_${groupID}`, label: groupName, kind: "group", disabled: true });
    for (const candidate of candidates) {
      options.push({
        value: creativeModelSettingKey(candidate),
        label: `${candidate.model} / ${candidate.platform}`,
        candidate,
      });
    }
  }
  return options;
});

function creativeModelOptionsForRow(index: number) {
  const current = form.creative_model_settings[index];
  const used = new Set(
    form.creative_model_settings
      .filter((_, itemIndex) => itemIndex !== index)
      .map((item) => creativeModelSettingKey(item)),
  );
  const currentKey = current ? creativeModelSettingKey(current) : "";
  const options = creativeModelCandidateOptions.value.filter(
    (option) => option.kind === "group" || option.value === currentKey || !used.has(option.value),
  );
  if (current && !creativeModelCandidates.value.some((candidate) => creativeModelSettingKey(candidate) === currentKey)) {
    options.unshift({
      value: currentKey,
      label: `${current.group_id} / ${current.model} (${t("admin.settings.features.creative.modelSettings.unavailable")})`,
      kind: "stale",
      disabled: true,
    });
  }
  return options;
}

function creativeCandidateForSetting(item: CreativeModelSetting): CreativeModelCandidate | undefined {
  const key = creativeModelSettingKey(item);
  return creativeModelCandidates.value.find((candidate) => creativeModelSettingKey(candidate) === key);
}

function creativeOperationSupported(index: number, operation: CreativeOperation): boolean {
  const item = form.creative_model_settings[index];
  if (!item) return false;
  return creativeCandidateForSetting(item)?.operations.includes(operation) ?? false;
}

function creativeOperationCheckboxDisabled(index: number, operation: CreativeOperation): boolean {
  const item = form.creative_model_settings[index];
  if (!item || !creativeOperationSupported(index, operation)) return true;
  return item.operations.length <= 1 && item.operations.includes(operation);
}

function onCreativeModelSelected(index: number, value: string | number | boolean | null) {
  if (typeof value !== "string") return;
  const candidate = creativeModelCandidates.value.find(
    (item) => creativeModelSettingKey(item) === value,
  );
  if (!candidate) return;
  const row = form.creative_model_settings[index];
  if (!row) return;
  row.group_id = candidate.group_id;
  row.model = candidate.model;
  row.operations = [...candidate.operations];
}

function addCreativeModelSetting() {
  const candidate = creativeModelCandidates.value.find(
    (item) => !form.creative_model_settings.some((setting) => creativeModelSettingKey(setting) === creativeModelSettingKey(item)),
  );
  if (!candidate) return;
  form.creative_model_settings.push({
    group_id: candidate.group_id,
    model: candidate.model,
    operations: [...candidate.operations],
  });
}

function removeCreativeModelSetting(index: number) {
  form.creative_model_settings.splice(index, 1);
}

function toggleCreativeOperation(index: number, operation: CreativeOperation, checked: boolean) {
  const item = form.creative_model_settings[index];
  if (!item || !creativeOperationSupported(index, operation)) return;
  const next = new Set(item.operations);
  if (checked) next.add(operation);
  else next.delete(operation);
  item.operations = creativeOperationChoices.filter((choice) => next.has(choice));
}

function normalizeCreativeModelSettingsForSave(): CreativeModelSetting[] | null {
  const seen = new Set<string>();
  const normalized: CreativeModelSetting[] = [];
  for (const item of form.creative_model_settings) {
    const groupID = Number(item.group_id);
    const model = String(item.model || "").trim();
    const candidate = creativeModelCandidates.value.find(
      (entry) => creativeModelSettingKey(entry) === creativeModelSettingKey({ group_id: groupID, model }),
    );
    // 按平台候选能力过滤提交值，已移除的 Gemini inpaint 会被过滤。
    const operations = creativeOperationChoices.filter(
      (operation) => item.operations.includes(operation) && (candidate?.operations.includes(operation) ?? true),
    );
    if (!Number.isSafeInteger(groupID) || groupID <= 0 || !model) {
      return null;
    }
    if (operations.length === 0) {
      // 已知候选但没有任何可用能力时删除该条目；未知历史分组仍按 fail-closed 保留原值。
      if (candidate) continue;
      return null;
    }
    const key = creativeModelSettingKey({ group_id: groupID, model });
    if (seen.has(key)) return null;
    seen.add(key);
    normalized.push({ group_id: groupID, model, operations });
  }
  return normalized;
}

async function loadCreativeModelCandidates() {
  creativeModelCandidatesLoading.value = true;
  creativeModelCandidatesError.value = false;
  try {
    creativeModelCandidates.value = await adminAPI.settings.getCreativeModelCandidates();
  } catch {
    creativeModelCandidatesError.value = true;
  } finally {
    creativeModelCandidatesLoading.value = false;
  }
}

// 显示用于排序的指标，供用户查看排名依据。
function ensureUsageRankingSortMetricVisible() {
  switch (form.usage_ranking_sort_by) {
    case "requests":
      form.usage_ranking_show_requests = true;
      break;
    case "actual_cost":
      form.usage_ranking_show_actual_cost = true;
      break;
    default:
      form.usage_ranking_show_total_tokens = true;
      break;
  }
}

watch(
  () => form.usage_ranking_sort_by,
  () => ensureUsageRankingSortMetricVisible(),
);

const {
  balanceUnitName: previewBalanceUnitName,
  balanceUnitSymbol: previewBalanceUnitSymbol,
  formatBalanceAmount: formatPreviewBalanceAmount,
} = useBalanceDisplay({
  unitName: computed(() => form.balance_unit_name),
  unitSymbol: computed(() => form.balance_unit_symbol),
  iconSvg: computed(() => form.balance_icon_svg),
});
const previewBalanceAmount = computed(() => formatPreviewBalanceAmount(123.45));

function quotaThresholdToPercent(value: number | undefined): number | null {
  if (!value || value <= 0) return null;
  return Math.round(value * 1000) / 10;
}

function percentToQuotaThreshold(value: number | null): number {
  return value != null && value > 0 ? value / 100 : 0;
}

// OpenAI 配额自动暂停在后端以 0~1 存储，系统设置页按百分比展示。
const openAIQuotaAutoPause5hPercent = computed<number | null>({
  get() {
    return quotaThresholdToPercent(form.openai_provider_quota_auto_pause?.default_threshold_5h);
  },
  set(value) {
    form.openai_provider_quota_auto_pause.default_threshold_5h = percentToQuotaThreshold(value);
  },
});

const openAIQuotaAutoPause7dPercent = computed<number | null>({
  get() {
    return quotaThresholdToPercent(form.openai_provider_quota_auto_pause?.default_threshold_7d);
  },
  set(value) {
    form.openai_provider_quota_auto_pause.default_threshold_7d = percentToQuotaThreshold(value);
  },
});

const oidcTokenAuthMethodOptions = [
  { value: "client_secret_post", label: "client_secret_post" },
  { value: "client_secret_basic", label: "client_secret_basic" },
  { value: "none", label: "none" },
];

const grokDefaultBaseURLOptions = computed(() => [
  {
    value: "cli",
    label: t("admin.settings.gatewayForwarding.grokBaseURLModeCLI"),
  },
  {
    value: "api",
    label: t("admin.settings.gatewayForwarding.grokBaseURLModeAPI"),
  },
  {
    value: "us-east-1",
    label: t("admin.settings.gatewayForwarding.grokBaseURLModeUSEast1"),
  },
  {
    value: "us-west-2",
    label: t("admin.settings.gatewayForwarding.grokBaseURLModeUSWest2"),
  },
  {
    value: "eu-west-1",
    label: t("admin.settings.gatewayForwarding.grokBaseURLModeEUWest1"),
  },
]);

const customMenuVisibilityOptions = computed(() => [
  { value: "user", label: t("admin.settings.customMenu.visibilityUser") },
  { value: "admin", label: t("admin.settings.customMenu.visibilityAdmin") },
]);

const userPromptReplacementTypeOptions = computed(() => [
  {
    value: "static",
    label: t("admin.settings.userPromptReplacement.typeStatic"),
  },
  {
    value: "timezone_name",
    label: t("admin.settings.userPromptReplacement.typeTimezoneName"),
  },
  {
    value: "current_time",
    label: t("admin.settings.userPromptReplacement.typeCurrentTime"),
  },
]);

const userPromptReplacementTimezoneOptions = [
  { value: "Asia/Tokyo", label: "Asia/Tokyo" },
  { value: "Asia/Shanghai", label: "Asia/Shanghai" },
  { value: "UTC", label: "UTC" },
  { value: "America/Los_Angeles", label: "America/Los_Angeles" },
  { value: "America/New_York", label: "America/New_York" },
  { value: "Europe/London", label: "Europe/London" },
];

function addUserPromptReplacementRule(): void {
  form.user_prompt_replacement_config.rules.push(
    createUserPromptReplacementRule(),
  );
}

function removeUserPromptReplacementRule(index: number): void {
  form.user_prompt_replacement_config.rules.splice(index, 1);
}

function resetUserPromptReplacementRules(): void {
  form.user_prompt_replacement_config = defaultUserPromptReplacementConfig();
}

function defaultPaymentMethodFee(enabled = false): PaymentMethodFeeConfig {
  return { enabled, fixed_fee: 0, fee_rate: 0 };
}

function ensurePaymentMethodFee(method: string): PaymentMethodFeeConfig {
  if (!form.payment_method_fees) form.payment_method_fees = {};
  if (!form.payment_method_fees[method]) {
    form.payment_method_fees[method] = defaultPaymentMethodFee(false);
  }
  return form.payment_method_fees[method];
}

function methodFeeEnabled(method: string): boolean {
  return ensurePaymentMethodFee(method).enabled;
}

function setMethodFeeEnabled(method: string, enabled: boolean) {
  ensurePaymentMethodFee(method).enabled = enabled;
}

function normalizeFeeInput(value: string, max = Number.POSITIVE_INFINITY): number {
  const parsed = Number.parseFloat(value || "0");
  if (!Number.isFinite(parsed)) return 0;
  return Math.min(max, Math.max(0, Math.round(parsed * 100) / 100));
}

function methodFeeValue(method: string, field: "fixed_fee" | "fee_rate"): number {
  return ensurePaymentMethodFee(method)[field] || 0;
}

function setMethodFeeValue(method: string, field: "fixed_fee" | "fee_rate", value: string) {
  ensurePaymentMethodFee(method)[field] = normalizeFeeInput(value, field === "fee_rate" ? 100 : Number.POSITIVE_INFINITY);
}

function methodFeePreview(method: string): string {
  const cfg = ensurePaymentMethodFee(method);
  const base = 100;
  const rateFee = Math.ceil(((base * (Number(cfg.fee_rate) || 0)) / 100) * 100 - 1e-9) / 100;
  const totalFee = Math.round(((Number(cfg.fixed_fee) || 0) + rateFee + Number.EPSILON) * 100) / 100;
  return t("admin.settings.payment.methodFeePreview", {
    fee: totalFee.toFixed(2),
    amount: (base + totalFee).toFixed(2),
  });
}

// 人机验证 UI 状态：单卡片「总开关 + 服务商单选」，落库仍是三个独立
// enabled 键（与上游一致），由下面的映射保证同一时间至多一家启用。
type CaptchaProviderSelection = "turnstile" | "tencent" | "aliyun";

const captchaProviderSelection = ref<CaptchaProviderSelection>("turnstile");

function applyCaptchaSelection(provider: CaptchaProviderSelection | null): void {
  form.turnstile_enabled = provider === "turnstile";
  form.tencent_captcha_enabled = provider === "tencent";
  form.aliyun_captcha_enabled = provider === "aliyun";
}

const captchaMasterEnabled = computed({
  get: () =>
    form.turnstile_enabled ||
    form.tencent_captcha_enabled ||
    form.aliyun_captcha_enabled,
  set: (enabled: boolean) =>
    applyCaptchaSelection(enabled ? captchaProviderSelection.value : null),
});

function selectCaptchaProvider(provider: CaptchaProviderSelection): void {
  captchaProviderSelection.value = provider;
  applyCaptchaSelection(provider);
}

// 天御中国站与国际站是两套独立提供商体系，控制台与文档入口不通用，
// 按所选站点提供控制台链接，供管理员查找 CaptchaAppId。
const tencentCaptchaLinks = computed(() =>
  form.tencent_captcha_region === "intl"
    ? {
        console: "https://console.tencentcloud.com/captcha/graphical",
        cloudKeys: "https://console.tencentcloud.com/cam/capi",
        webDocs: "https://www.tencentcloud.com/document/product/1159/49680",
      }
    : {
        console: "https://console.cloud.tencent.com/captcha",
        cloudKeys: "https://console.cloud.tencent.com/cam/capi",
        webDocs: "https://cloud.tencent.com/document/product/1110/36841",
      },
);

function syncCaptchaProviderSelection(): void {
  if (form.tencent_captcha_enabled) {
    captchaProviderSelection.value = "tencent";
  } else if (form.aliyun_captcha_enabled) {
    captchaProviderSelection.value = "aliyun";
  } else if (form.turnstile_enabled) {
    captchaProviderSelection.value = "turnstile";
  }
}

type AdvancedSchedulerOverrideKey =
  | "advanced_scheduler_lb_top_k"
  | "advanced_scheduler_weight_priority"
  | "advanced_scheduler_weight_load"
  | "advanced_scheduler_weight_queue"
  | "advanced_scheduler_weight_error_rate"
  | "advanced_scheduler_weight_ttft"
  | "advanced_scheduler_weight_reset"
  | "advanced_scheduler_weight_quota_headroom"
  | "advanced_scheduler_weight_previous_response"
  | "advanced_scheduler_weight_session_sticky";

type AdvancedSchedulerEffectiveKey =
  | "advanced_scheduler_effective_lb_top_k"
  | "advanced_scheduler_effective_weight_priority"
  | "advanced_scheduler_effective_weight_load"
  | "advanced_scheduler_effective_weight_queue"
  | "advanced_scheduler_effective_weight_error_rate"
  | "advanced_scheduler_effective_weight_ttft"
  | "advanced_scheduler_effective_weight_reset"
  | "advanced_scheduler_effective_weight_quota_headroom"
  | "advanced_scheduler_effective_weight_previous_response"
  | "advanced_scheduler_effective_weight_session_sticky"
  | "advanced_scheduler_effective_ewma_error_rate_alpha"
  | "advanced_scheduler_effective_ewma_ttft_alpha"
  | "advanced_scheduler_effective_sticky_escape_enabled"
  | "advanced_scheduler_effective_sticky_escape_ttft_ms"
  | "advanced_scheduler_effective_sticky_escape_error_rate";

const advancedSchedulerPlaceholder = (
  effectiveKey: AdvancedSchedulerEffectiveKey,
  fallbackValue: string,
) => {
  const effectiveValue = String(
    (form as Record<string, unknown>)[effectiveKey] ?? "",
  ).trim();
  return t("admin.settings.scheduling.defaultPlaceholder", {
    value: effectiveValue || fallbackValue,
  });
};

const advancedSchedulerWeightFields = computed<
  Array<{
    key: AdvancedSchedulerOverrideKey;
    label: string;
    placeholder: string;
  }>
>(() => {
  return [
    {
      key: "advanced_scheduler_lb_top_k",
      label: t("admin.settings.scheduling.topKLabel"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_lb_top_k", "7"),
    },
    {
      key: "advanced_scheduler_weight_priority",
      label: t("admin.settings.scheduling.priorityWeight"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_weight_priority", "1"),
    },
    {
      key: "advanced_scheduler_weight_load",
      label: t("admin.settings.scheduling.loadWeight"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_weight_load", "1"),
    },
    {
      key: "advanced_scheduler_weight_queue",
      label: t("admin.settings.scheduling.queueWeight"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_weight_queue", "0.7"),
    },
    {
      key: "advanced_scheduler_weight_error_rate",
      label: t("admin.settings.scheduling.errorRateWeight"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_weight_error_rate", "0.8"),
    },
    {
      key: "advanced_scheduler_weight_ttft",
      label: t("admin.settings.scheduling.ttftWeight"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_weight_ttft", "0.5"),
    },
    {
      key: "advanced_scheduler_weight_reset",
      label: t("admin.settings.scheduling.resetWeight"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_weight_reset", "0"),
    },
    {
      key: "advanced_scheduler_weight_quota_headroom",
      label: t("admin.settings.scheduling.quotaHeadroomWeight"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_weight_quota_headroom", "0"),
    },
    {
      key: "advanced_scheduler_weight_previous_response",
      label: t("admin.settings.scheduling.previousResponseWeight"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_weight_previous_response", "5"),
    },
    {
      key: "advanced_scheduler_weight_session_sticky",
      label: t("admin.settings.scheduling.sessionStickyWeight"),
      placeholder: advancedSchedulerPlaceholder("advanced_scheduler_effective_weight_session_sticky", "3"),
    },
  ];
});

const authSourceDefaults = reactive<AuthSourceDefaultsState>(
  buildAuthSourceDefaultsState({}),
);

const authSourceDefaultsMeta = computed(() => [
  {
    source: "email" as AuthSourceType,
    title: t("admin.settings.authSourceDefaults.sources.email.title"),
    description: t("admin.settings.authSourceDefaults.sources.email.description"),
  },
  {
    source: "linuxdo" as AuthSourceType,
    title: t("admin.settings.authSourceDefaults.sources.linuxdo.title"),
    description: t("admin.settings.authSourceDefaults.sources.linuxdo.description"),
  },
  {
    source: "oidc" as AuthSourceType,
    title: t("admin.settings.authSourceDefaults.sources.oidc.title"),
    description: t("admin.settings.authSourceDefaults.sources.oidc.description"),
  },
  {
    source: "wechat" as AuthSourceType,
    title: t("admin.settings.authSourceDefaults.sources.wechat.title"),
    description: t("admin.settings.authSourceDefaults.sources.wechat.description"),
  },
  {
    source: "github" as AuthSourceType,
    title: t("admin.settings.authSourceDefaults.sources.github.title"),
    description: t("admin.settings.authSourceDefaults.sources.github.description"),
  },
  {
    source: "google" as AuthSourceType,
    title: t("admin.settings.authSourceDefaults.sources.google.title"),
    description: t("admin.settings.authSourceDefaults.sources.google.description"),
  },
  {
    source: "dingtalk" as AuthSourceType,
    title: t("admin.settings.authSourceDefaults.sources.dingtalk.title"),
    description: t("admin.settings.authSourceDefaults.sources.dingtalk.description"),
  },
]);

// Proxies for web search emulation ProxySelector
const webSearchProxies = ref<Proxy[]>([]);

// Web Search Emulation config (loaded/saved separately)
const DEFAULT_WEB_SEARCH_QUOTA_LIMIT = 1000;

const webSearchConfig = reactive<WebSearchEmulationConfig>({
  enabled: false,
  providers: [],
});

const expandedProviders = reactive<Record<number, boolean>>({});
const apiKeyVisible = reactive<Record<number, boolean>>({});
const wsTestQuery = ref("");
const wsTestLoading = ref(false);
const wsTestResult = ref<WebSearchTestResult | null>(null);
const wsTestDialogOpen = ref(false);

function openTestDialog() {
  wsTestResult.value = null;
  wsTestDialogOpen.value = true;
}

function toggleProviderExpand(idx: number) {
  expandedProviders[idx] = !expandedProviders[idx];
}

function removeWebSearchProvider(idx: number) {
  webSearchConfig.providers.splice(idx, 1);
  // Re-index expandedProviders and apiKeyVisible after removal
  const newExpanded: Record<number, boolean> = {};
  const newVisible: Record<number, boolean> = {};
  for (let i = 0; i < webSearchConfig.providers.length; i++) {
    const oldIdx = i >= idx ? i + 1 : i;
    newExpanded[i] = expandedProviders[oldIdx] ?? false;
    newVisible[i] = apiKeyVisible[oldIdx] ?? false;
  }
  Object.keys(expandedProviders).forEach(
    (k) => delete expandedProviders[Number(k)],
  );
  Object.keys(apiKeyVisible).forEach((k) => delete apiKeyVisible[Number(k)]);
  Object.assign(expandedProviders, newExpanded);
  Object.assign(apiKeyVisible, newVisible);
}

function addWebSearchProvider() {
  const idx = webSearchConfig.providers.length;
  webSearchConfig.providers.push({
    type: "brave",
    api_key: "",
    api_key_configured: false,
    quota_limit: DEFAULT_WEB_SEARCH_QUOTA_LIMIT,
    subscribed_at: null,
    proxy_id: null,
    expires_at: null,
  } as WebSearchProviderConfig);
  expandedProviders[idx] = true;
}

function formatSubscribedAt(ts: number | null): string {
  if (!ts) return "";
  // Use UTC to avoid timezone drift on repeated edits
  const d = new Date(ts * 1000);
  const y = d.getUTCFullYear();
  const m = String(d.getUTCMonth() + 1).padStart(2, "0");
  const day = String(d.getUTCDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function parseSubscribedAt(dateStr: string): number | null {
  if (!dateStr) return null;
  // Parse as UTC to match formatSubscribedAt
  return Math.floor(new Date(dateStr + "T00:00:00Z").getTime() / 1000);
}

function quotaPercentage(provider: WebSearchProviderConfig): number {
  if (!provider.quota_limit || provider.quota_limit <= 0) return 0;
  return ((provider.quota_used ?? 0) / provider.quota_limit) * 100;
}

async function resetWebSearchUsage(idx: number) {
  const provider = webSearchConfig.providers[idx];
  if (!provider) return;
  if (!confirm(t("admin.settings.webSearchEmulation.resetUsageConfirm")))
    return;
  try {
    await adminAPI.settings.resetWebSearchUsage({
      provider_type: provider.type,
    });
    provider.quota_used = 0;
    appStore.showSuccess(
      t("admin.settings.webSearchEmulation.resetUsageSuccess"),
    );
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t("common.error")));
  }
}

async function copyApiKey(idx: number) {
  const key = webSearchConfig.providers[idx]?.api_key;
  if (!key) {
    appStore.showError(
      t("admin.settings.webSearchEmulation.apiKeyPlaceholder"),
    );
    return;
  }
  try {
    await navigator.clipboard.writeText(key);
    appStore.showSuccess(t("admin.settings.webSearchEmulation.copied"));
  } catch {
    appStore.showError(t("common.error"));
  }
}

async function testWebSearchProvider() {
  wsTestLoading.value = true;
  wsTestResult.value = null;
  try {
    const query =
      wsTestQuery.value.trim() ||
      t("admin.settings.webSearchEmulation.testDefaultQuery");
    wsTestResult.value = await adminAPI.settings.testWebSearchEmulation(query);
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t("common.error")));
  } finally {
    wsTestLoading.value = false;
  }
}

async function loadWebSearchConfig() {
  try {
    const [resp, proxiesResp] = await Promise.all([
      adminAPI.settings.getWebSearchEmulationConfig(),
      adminAPI.proxies.list().catch(() => ({ items: [] as Proxy[] })),
    ]);
    if (resp) {
      webSearchConfig.enabled = resp.enabled || false;
      webSearchConfig.providers = resp.providers || [];
    }
    webSearchProxies.value = proxiesResp.items || [];
  } catch (err: unknown) {
    // 404 is expected when config hasn't been created yet; show error for other failures
    const status = (err as { status?: number })?.status;
    if (status !== 404 && status !== undefined) {
      appStore.showError(extractApiErrorMessage(err, t("common.error")));
    }
  }
}

async function saveWebSearchConfig(): Promise<boolean> {
  try {
    for (const p of webSearchConfig.providers) {
      const raw = p.quota_limit;
      if (raw != null && Number(raw) !== 0 && Number(raw) < 1) {
        appStore.showError(
          t("admin.settings.webSearchEmulation.quotaLimitMustBePositive"),
        );
        return false;
      }
    }
    const providers = webSearchConfig.providers.map(
      (p: WebSearchProviderConfig) => ({
        ...p,
        quota_limit: Number(p.quota_limit) > 0 ? Number(p.quota_limit) : null,
      }),
    );
    await adminAPI.settings.updateWebSearchEmulationConfig({
      enabled: webSearchConfig.enabled,
      providers,
    });
    return true;
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t("common.error")));
    return false;
  }
}

const defaultSubscriptionPlanOptions = computed<
  DefaultSubscriptionPlanOption[]
>(() =>
  subscriptionPlans.value.map((plan) => ({
    value: plan.id,
    label: plan.name,
    description: plan.description,
  })),
);

const registrationEmailSuffixWhitelistSeparatorKeys = new Set([
  " ",
  ",",
  "，",
  "Enter",
  "Tab",
]);

function removeRegistrationEmailSuffixWhitelistTag(suffix: string) {
  registrationEmailSuffixWhitelistTags.value =
    registrationEmailSuffixWhitelistTags.value.filter(
      (item) => item !== suffix,
    );
}

function addRegistrationEmailSuffixWhitelistTag(raw: string) {
  const suffix = normalizeRegistrationEmailSuffixDomain(raw);
  if (
    !isRegistrationEmailSuffixDomainValid(suffix) ||
    registrationEmailSuffixWhitelistTags.value.includes(suffix)
  ) {
    return;
  }
  registrationEmailSuffixWhitelistTags.value = [
    ...registrationEmailSuffixWhitelistTags.value,
    suffix,
  ];
}

function commitRegistrationEmailSuffixWhitelistDraft() {
  if (!registrationEmailSuffixWhitelistDraft.value) {
    return;
  }
  addRegistrationEmailSuffixWhitelistTag(
    registrationEmailSuffixWhitelistDraft.value,
  );
  registrationEmailSuffixWhitelistDraft.value = "";
}

function handleRegistrationEmailSuffixWhitelistDraftInput() {
  registrationEmailSuffixWhitelistDraft.value =
    normalizeRegistrationEmailSuffixDomain(
      registrationEmailSuffixWhitelistDraft.value,
    );
}

function handleRegistrationEmailSuffixWhitelistDraftKeydown(
  event: KeyboardEvent,
) {
  if (event.isComposing) {
    return;
  }

  if (registrationEmailSuffixWhitelistSeparatorKeys.has(event.key)) {
    event.preventDefault();
    commitRegistrationEmailSuffixWhitelistDraft();
    return;
  }

  if (
    event.key === "Backspace" &&
    !registrationEmailSuffixWhitelistDraft.value &&
    registrationEmailSuffixWhitelistTags.value.length > 0
  ) {
    registrationEmailSuffixWhitelistTags.value.pop();
  }
}

function handleRegistrationEmailSuffixWhitelistPaste(event: ClipboardEvent) {
  const text = event.clipboardData?.getData("text") || "";
  if (!text.trim()) {
    return;
  }
  event.preventDefault();
  const tokens = parseRegistrationEmailSuffixWhitelistInput(text);
  for (const token of tokens) {
    addRegistrationEmailSuffixWhitelistTag(token);
  }
}

const forwardedClientIpHeaderSeparatorKeys = new Set([
  " ",
  ",",
  "，",
  "Enter",
  "Tab",
]);
const forwardedClientIpHeaderTokenPattern = /^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/;
const maxForwardedClientIpHeaders = 16;

type ForwardedClientIpHeaderResult = "added" | "duplicate" | "invalid" | "full";

function normalizeForwardedClientIpHeader(raw: string): string {
  const header = raw.trim();
  if (!forwardedClientIpHeaderTokenPattern.test(header)) {
    return "";
  }

  return header
    .toLowerCase()
    .split("-")
    .map((part) => `${part.charAt(0).toUpperCase()}${part.slice(1)}`)
    .join("-");
}

function normalizeForwardedClientIpHeaders(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }

  const headers: string[] = [];
  const seen = new Set<string>();
  for (const raw of value) {
    if (typeof raw !== "string") {
      continue;
    }
    const header = normalizeForwardedClientIpHeader(raw);
    const key = header.toLowerCase();
    if (!header || seen.has(key) || headers.length >= maxForwardedClientIpHeaders) {
      continue;
    }
    seen.add(key);
    headers.push(header);
  }
  return headers;
}

function removeForwardedClientIpHeader(header: string) {
  form.forwarded_client_ip_headers = form.forwarded_client_ip_headers.filter(
    (item) => item !== header,
  );
}

function addForwardedClientIpHeader(raw: string): ForwardedClientIpHeaderResult {
  const header = normalizeForwardedClientIpHeader(raw);
  if (!header) {
    return "invalid";
  }
  if (
    form.forwarded_client_ip_headers.some(
      (item) => item.toLowerCase() === header.toLowerCase(),
    )
  ) {
    return "duplicate";
  }
  if (form.forwarded_client_ip_headers.length >= maxForwardedClientIpHeaders) {
    return "full";
  }
  form.forwarded_client_ip_headers = [
    ...form.forwarded_client_ip_headers,
    header,
  ];
  return "added";
}

function showForwardedClientIpHeaderError(result: ForwardedClientIpHeaderResult) {
  if (result === "invalid") {
    appStore.showError(t("admin.settings.apiKeyAcl.forwardedClientIpHeaderInvalid"));
  } else if (result === "full") {
    appStore.showError(
      t("admin.settings.apiKeyAcl.forwardedClientIpHeadersLimit", {
        max: maxForwardedClientIpHeaders,
      }),
    );
  }
}

function commitForwardedClientIpHeaderDraft() {
  const draft = forwardedClientIpHeaderDraft.value;
  if (!draft) {
    return;
  }
  const result = addForwardedClientIpHeader(draft);
  showForwardedClientIpHeaderError(result);
  forwardedClientIpHeaderDraft.value = "";
}

function handleForwardedClientIpHeaderKeydown(event: KeyboardEvent) {
  if (event.isComposing) {
    return;
  }
  if (forwardedClientIpHeaderSeparatorKeys.has(event.key)) {
    event.preventDefault();
    commitForwardedClientIpHeaderDraft();
    return;
  }
  if (
    event.key === "Backspace" &&
    !forwardedClientIpHeaderDraft.value &&
    form.forwarded_client_ip_headers.length > 0
  ) {
    form.forwarded_client_ip_headers.pop();
  }
}

function handleForwardedClientIpHeaderPaste(event: ClipboardEvent) {
  const text = event.clipboardData?.getData("text") || "";
  if (!text.trim()) {
    return;
  }
  event.preventDefault();

  let error: ForwardedClientIpHeaderResult | undefined;
  for (const token of text.split(/[,，;\r\n]+/)) {
    if (!token.trim()) {
      continue;
    }
    const result = addForwardedClientIpHeader(token);
    if (result === "invalid" || result === "full") {
      error = result;
    }
  }
  if (error) {
    showForwardedClientIpHeaderError(error);
  }
}

// Quota notify email helpers
const addQuotaNotifyEmail = () => {
  if (!form.provider_quota_notify_emails) {
    form.provider_quota_notify_emails = [];
  }
  form.provider_quota_notify_emails.push({
    email: "",
    disabled: false,
    verified: true,
  });
};

const currentOrigin =
  typeof window !== "undefined" ? window.location.origin : "";

function buildApiCallbackUrl(path: string): string {
  const base = (form.api_base_url || currentOrigin).replace(/\/+$/, "");
  const apiRoot = base.endsWith("/api/v1") ? base : `${base}/api/v1`;
  return `${apiRoot}${path.startsWith("/") ? path : `/${path}`}`;
}

// LinuxDo OAuth redirect URL suggestion
const linuxdoRedirectUrlSuggestion = computed(() => {
  return buildApiCallbackUrl("/auth/oauth/linuxdo/callback");
});

async function setAndCopyLinuxdoRedirectUrl() {
  const url = linuxdoRedirectUrlSuggestion.value;
  if (!url) return;

  form.linuxdo_connect_redirect_url = url;
  await copyToClipboard(
    url,
    t("admin.settings.linuxdo.redirectUrlSetAndCopied"),
  );
}

type EmailOAuthProvider = "github" | "google";

const githubOAuthRedirectUrlSuggestion = computed(() => {
  return buildApiCallbackUrl("/auth/oauth/github/callback");
});

const googleOAuthRedirectUrlSuggestion = computed(() => {
  return buildApiCallbackUrl("/auth/oauth/google/callback");
});

const googleOneTapOriginSuggestion = computed(() => {
  if (typeof window === "undefined") return "";
  return window.location.origin;
});

async function copyGoogleOneTapOrigin() {
  const origin = googleOneTapOriginSuggestion.value;
  if (!origin) return;
  await copyToClipboard(
    origin,
    t("admin.settings.emailOAuth.originCopied"),
  );
}

async function setAndCopyEmailOAuthRedirectUrl(provider: EmailOAuthProvider) {
  const url =
    provider === "github"
      ? githubOAuthRedirectUrlSuggestion.value
      : googleOAuthRedirectUrlSuggestion.value;
  if (!url) return;

  if (provider === "github") {
    form.github_oauth_redirect_url = url;
  } else {
    form.google_oauth_redirect_url = url;
  }
  await copyToClipboard(url, t("admin.settings.emailOAuth.callbackCopied"));
}

const wechatRedirectUrlSuggestion = computed(() => {
  return buildApiCallbackUrl("/auth/oauth/wechat/callback");
});

function syncWeChatConnectMode(preferredMode?: WeChatConnectMode) {
  if (form.wechat_connect_mp_enabled && form.wechat_connect_mobile_enabled) {
    if (preferredMode === "mobile") {
      form.wechat_connect_mp_enabled = false;
    } else {
      form.wechat_connect_mobile_enabled = false;
    }
  }

  const capabilities = resolveWeChatConnectModeCapabilities(
    form.wechat_connect_open_enabled,
    form.wechat_connect_mp_enabled,
    form.wechat_connect_mobile_enabled,
    form.wechat_connect_mode,
  );
  form.wechat_connect_open_enabled = capabilities.openEnabled;
  form.wechat_connect_mp_enabled = capabilities.mpEnabled;
  form.wechat_connect_mobile_enabled = capabilities.mobileEnabled;
  form.wechat_connect_mode = deriveWeChatConnectStoredMode(
    capabilities.openEnabled,
    capabilities.mpEnabled,
    capabilities.mobileEnabled,
    form.wechat_connect_mode,
  );
  form.wechat_connect_scopes = defaultWeChatConnectScopesForMode(
    form.wechat_connect_mode,
  );
}

function handleWeChatOpenEnabledChange(value: boolean) {
  form.wechat_connect_open_enabled = value;
  syncWeChatConnectMode(value ? "open" : undefined);
}

function handleWeChatMPEnabledChange(value: boolean) {
  form.wechat_connect_mp_enabled = value;
  if (value) {
    form.wechat_connect_mobile_enabled = false;
  }
  syncWeChatConnectMode(value ? "mp" : undefined);
}

function handleWeChatMobileEnabledChange(value: boolean) {
  form.wechat_connect_mobile_enabled = value;
  if (value) {
    form.wechat_connect_mp_enabled = false;
  }
  syncWeChatConnectMode(value ? "mobile" : undefined);
}

async function setAndCopyWeChatRedirectUrl() {
  const url = wechatRedirectUrlSuggestion.value;
  if (!url) return;

  form.wechat_connect_redirect_url = url;
  await copyToClipboard(
    url,
    t("admin.settings.wechatConnect.redirectUrlSetAndCopied"),
  );
}

const oidcRedirectUrlSuggestion = computed(() => {
  return buildApiCallbackUrl("/auth/oauth/oidc/callback");
});

async function setAndCopyOIDCRedirectUrl() {
  const url = oidcRedirectUrlSuggestion.value;
  if (!url) return;

  form.oidc_connect_redirect_url = url;
  await copyToClipboard(url, t("admin.settings.oidc.redirectUrlSetAndCopied"));
}

// Custom menu item management
function addMenuItem() {
  form.custom_menu_items.push({
    id: "",
    label: "",
    icon_svg: "",
    url: "",
    visibility: "user",
    sort_order: form.custom_menu_items.length,
  });
}

function removeMenuItem(index: number) {
  form.custom_menu_items.splice(index, 1);
  // Re-index sort_order
  form.custom_menu_items.forEach((item, i) => {
    item.sort_order = i;
  });
}

function moveMenuItem(index: number, direction: -1 | 1) {
  const targetIndex = index + direction;
  if (targetIndex < 0 || targetIndex >= form.custom_menu_items.length) return;
  const items = form.custom_menu_items;
  const temp = items[index];
  items[index] = items[targetIndex];
  items[targetIndex] = temp;
  // Re-index sort_order
  items.forEach((item, i) => {
    item.sort_order = i;
  });
}

// Custom endpoint management
function addEndpoint() {
  form.custom_endpoints.push({ name: "", endpoint: "", description: "" });
}

function removeEndpoint(index: number) {
  form.custom_endpoints.splice(index, 1);
}

// Footer link group management
function addFooterGroup() {
  form.footer_links.push({ title: "", links: [{ label: "", url: "" }] });
}

// 默认底栏模板:基于站内已有页面 + 已配置的文档/联系方式,可直接用或小改
function applyDefaultFooterLinks() {
  const zh = locale.value.toLowerCase().startsWith("zh");
  const groups: Array<{ title: string; links: Array<{ label: string; url: string }> }> = [
    {
      title: zh ? "产品" : "Product",
      links: [
        { label: zh ? "模型广场" : "Models", url: "/models" },
        { label: zh ? "用量查询" : "Key usage", url: "/key-usage" },
        { label: zh ? "控制台" : "Dashboard", url: "/dashboard" },
      ],
    },
    {
      title: zh ? "开发者" : "Developer",
      links: [
        ...(form.doc_url ? [{ label: zh ? "文档" : "Documentation", url: form.doc_url }] : []),
        ...(form.api_base_url
          ? [{ label: "API Base URL", url: form.api_base_url }]
          : []),
      ],
    },
    {
      title: zh ? "支持" : "Support",
      links: [
        { label: zh ? "登录 / 注册" : "Sign in / Sign up", url: "/login" },
      ],
    },
  ];
  form.footer_links = groups.filter((g) => g.links.length > 0);
}

function removeFooterGroup(index: number) {
  form.footer_links.splice(index, 1);
}

function moveFooterGroup(index: number, direction: -1 | 1) {
  const targetIndex = index + direction;
  if (targetIndex < 0 || targetIndex >= form.footer_links.length) return;
  const groups = form.footer_links;
  const temp = groups[index];
  groups[index] = groups[targetIndex];
  groups[targetIndex] = temp;
}

// 保存前清理:去掉空链接行和无标题且无链接的分组
function normalizeFooterLinksForSave() {
  return form.footer_links
    .map((group) => ({
      title: group.title.trim(),
      links: group.links
        .map((link) => ({ label: link.label.trim(), url: link.url.trim() }))
        .filter((link) => link.label && link.url),
    }))
    .filter((group) => group.title && group.links.length > 0);
}

// 首页展示模型上限，与后端校验保持一致
const homeFeaturedModelsMax = 12;

// 首页模型展示卡片：选项来自公开模型广场接口，按分组分片展示
const homeFeaturedModelOptions = ref<
  Array<{ value: string; label: string; kind?: string; disabled?: boolean }>
>([]);

async function loadHomeFeaturedModelOptions() {
  try {
    const groups: MarketplaceGroup[] = await getMarketplaceModels();
    const options: Array<{ value: string; label: string; kind?: string; disabled?: boolean }> = [];
    for (const group of groups) {
      if (!group.models?.length) continue;
      options.push({ value: `__group_${group.id}`, label: group.name, kind: "group", disabled: true });
      for (const model of group.models) {
        options.push({
          value: model.id,
          label: model.display_name ? `${model.display_name}（${model.id}）` : model.id,
        });
      }
    }
    homeFeaturedModelOptions.value = options;
  } catch {
    // 选项加载失败不阻塞设置页，已配置的模型 ID 仍会原样保存
    homeFeaturedModelOptions.value = [];
  }
}

function removeHomeFeaturedModel(index: number) {
  form.home_featured_models.splice(index, 1);
}

function moveHomeFeaturedModel(index: number, direction: -1 | 1) {
  const targetIndex = index + direction;
  if (targetIndex < 0 || targetIndex >= form.home_featured_models.length) return;
  const models = form.home_featured_models;
  const temp = models[index];
  models[index] = models[targetIndex];
  models[targetIndex] = temp;
}

// 保存前清理:去掉空白项并去重，保持管理员配置的顺序
function normalizeHomeFeaturedModelsForSave() {
  const seen = new Set<string>();
  return form.home_featured_models
    .map((id) => id.trim())
    .filter((id) => {
      if (!id || seen.has(id)) return false;
      seen.add(id);
      return true;
    });
}

function addLoginAgreementDocument() {
  form.login_agreement_documents.push({
    id: `custom-${Date.now().toString(36)}`,
    title: "",
    content_md: "",
  });
}

function removeLoginAgreementDocument(index: number) {
  form.login_agreement_documents.splice(index, 1);
}

function normalizeLoginAgreementDocumentsForSave(): LoginAgreementDocument[] {
  return form.login_agreement_documents
    .map((doc, index) => ({
      id:
        normalizeLoginAgreementDocumentId(doc.id || doc.title) ||
        `doc-${index + 1}`,
      title: doc.title.trim(),
      content_md: doc.content_md.trim(),
    }))
    .filter((doc) => doc.title || doc.content_md);
}

// 保存前清洗空白与数值字段，让后端只负责正则和时区等强校验。
function normalizeUserPromptReplacementConfigForSave(): UserPromptReplacementConfig {
  const config = normalizeUserPromptReplacementConfig(
    form.user_prompt_replacement_config,
  );
  return {
    enabled: config.enabled,
    rules: config.rules.map((rule, index) => ({
      id: rule.id || `rule-${index + 1}`,
      name: rule.name || rule.id || `Rule ${index + 1}`,
      enabled: rule.enabled !== false,
      pattern: rule.pattern.trim(),
      target_group: Math.max(0, Math.floor(Number(rule.target_group) || 0)),
      replacement_type: normalizeUserPromptReplacementType(
        rule.replacement_type,
      ),
      scope: rule.scope || "",
      static_text: rule.static_text || "",
      timezone: rule.timezone || "Asia/Tokyo",
      time_format: rule.time_format || "2006-01-02",
    })),
  };
}

function findDuplicateLoginAgreementDocumentId(
  documents: LoginAgreementDocument[],
): string | null {
  const seen = new Set<string>();
  for (const doc of documents) {
    if (seen.has(doc.id)) {
      return doc.id;
    }
    seen.add(doc.id);
  }
  return null;
}

function formatTablePageSizeOptions(options: number[]): string {
  return options.join(", ");
}

function parseTablePageSizeOptionsInput(raw: string): number[] | null {
  const tokens = raw
    .split(",")
    .map((token) => token.trim())
    .filter((token) => token.length > 0);

  if (tokens.length === 0) {
    return null;
  }

  const parsed = tokens.map((token) => Number(token));
  if (parsed.some((value) => !Number.isInteger(value))) {
    return null;
  }

  const deduped = Array.from(new Set(parsed)).sort((a, b) => a - b);
  if (
    deduped.some(
      (value) => value < tablePageSizeMin || value > tablePageSizeMax,
    )
  ) {
    return null;
  }

  return deduped;
}

async function loadSettings() {
  loading.value = true;
  loadFailed.value = false;
  try {
    const settings = await adminAPI.settings.getSettings();
    settings.payment_load_balance_strategy =
      settings.payment_load_balance_strategy || "round-robin";
    // Only assign non-null values from backend (null means unconfigured, keep defaults)
    for (const [key, value] of Object.entries(settings)) {
      if (value !== null && value !== undefined) {
        (form as Record<string, unknown>)[key] = value;
      }
    }
    form.site_name_zh = form.site_name_zh || settings.site_name || "TokenRouter";
    form.site_name_en = form.site_name_en || "";
    form.site_subtitle_zh =
      form.site_subtitle_zh || settings.site_subtitle || "";
    form.site_subtitle_en = form.site_subtitle_en || "";
    syncCaptchaProviderSelection();
    form.login_agreement_mode =
      settings.login_agreement_mode === "checkbox" ? "checkbox" : "modal";
    form.login_agreement_updated_at =
      settings.login_agreement_updated_at || "2026-03-31";
    form.login_agreement_documents =
      Array.isArray(settings.login_agreement_documents) &&
      settings.login_agreement_documents.length > 0
        ? settings.login_agreement_documents.map((doc) => ({
            id: doc.id || "",
            title: doc.title || "",
            content_md: doc.content_md || "",
          }))
        : defaultLoginAgreementDocuments();
    Object.assign(authSourceDefaults, buildAuthSourceDefaultsState(settings));
    form.user_prompt_replacement_config =
      normalizeUserPromptReplacementConfig(
        settings.user_prompt_replacement_config,
      );
    form.openai_provider_quota_auto_pause = {
      default_threshold_5h:
        settings.openai_provider_quota_auto_pause?.default_threshold_5h ?? 0,
      default_threshold_7d:
        settings.openai_provider_quota_auto_pause?.default_threshold_7d ?? 0,
    };
    form.backend_mode_enabled = settings.backend_mode_enabled;
    form.default_subscriptions = normalizeDefaultSubscriptionSettings(
      settings.default_subscriptions,
    );
    registrationEmailSuffixWhitelistTags.value =
      normalizeRegistrationEmailSuffixDomains(
        settings.registration_email_suffix_whitelist,
      );
    form.forwarded_client_ip_headers = normalizeForwardedClientIpHeaders(
      settings.forwarded_client_ip_headers,
    );
    forwardedClientIpHeaderDraft.value = "";
    tablePageSizeOptionsInput.value = formatTablePageSizeOptions(
      Array.isArray(settings.table_page_size_options)
        ? settings.table_page_size_options
        : [10, 20, 50, 100],
    );
    registrationEmailSuffixWhitelistDraft.value = "";
    form.smtp_password = "";
    smtpPasswordManuallyEdited.value = false;
    form.turnstile_secret_key = "";
    form.tencent_captcha_app_secret_key = "";
    form.tencent_captcha_cloud_secret_id = "";
    form.tencent_captcha_cloud_secret_key = "";
    form.aliyun_captcha_access_key_secret = "";
    form.linuxdo_connect_client_secret = "";
    form.dingtalk_connect_client_secret = "";
    form.wechat_connect_app_secret = "";
    form.wechat_connect_open_app_secret = "";
    form.wechat_connect_mp_app_secret = "";
    form.wechat_connect_mobile_app_secret = "";
    form.github_oauth_client_secret = "";
    form.google_oauth_client_secret = "";
    const wechatCapabilities = resolveWeChatConnectModeCapabilities(
      settings.wechat_connect_open_enabled,
      settings.wechat_connect_mp_enabled,
      settings.wechat_connect_mobile_enabled,
      settings.wechat_connect_mode,
    );
    form.wechat_connect_open_enabled = wechatCapabilities.openEnabled;
    form.wechat_connect_mp_enabled = wechatCapabilities.mpEnabled;
    form.wechat_connect_mobile_enabled = wechatCapabilities.mobileEnabled;
    form.wechat_connect_mode = deriveWeChatConnectStoredMode(
      wechatCapabilities.openEnabled,
      wechatCapabilities.mpEnabled,
      wechatCapabilities.mobileEnabled,
      settings.wechat_connect_mode,
    );
    const legacyWeChatAppID = String(settings.wechat_connect_app_id || "").trim();
    const legacyWeChatSecretConfigured = Boolean(
      settings.wechat_connect_app_secret_configured,
    );
    if (!form.wechat_connect_open_app_id && wechatCapabilities.openEnabled) {
      form.wechat_connect_open_app_id = legacyWeChatAppID;
    }
    if (!form.wechat_connect_mp_app_id && wechatCapabilities.mpEnabled) {
      form.wechat_connect_mp_app_id = legacyWeChatAppID;
    }
    if (!form.wechat_connect_mobile_app_id && wechatCapabilities.mobileEnabled) {
      form.wechat_connect_mobile_app_id = legacyWeChatAppID;
    }
    if (
      !form.wechat_connect_open_app_secret_configured &&
      wechatCapabilities.openEnabled
    ) {
      form.wechat_connect_open_app_secret_configured =
        legacyWeChatSecretConfigured;
    }
    if (
      !form.wechat_connect_mp_app_secret_configured &&
      wechatCapabilities.mpEnabled
    ) {
      form.wechat_connect_mp_app_secret_configured = legacyWeChatSecretConfigured;
    }
    if (
      !form.wechat_connect_mobile_app_secret_configured &&
      wechatCapabilities.mobileEnabled
    ) {
      form.wechat_connect_mobile_app_secret_configured =
        legacyWeChatSecretConfigured;
    }
    form.wechat_connect_scopes = defaultWeChatConnectScopesForMode(
      form.wechat_connect_mode,
    );
    form.oidc_connect_client_secret = "";

    // Load OpenAI fast/flex policy rules from bulk settings.
    // 仅当 payload 真的包含该字段时填充并标记为已加载；否则保持表单空值，
    // 让 saveSettings 在未加载时跳过该字段，防止覆盖后端默认规则。
    if (
      settings.openai_fast_policy_settings &&
      Array.isArray(settings.openai_fast_policy_settings.rules)
    ) {
      openaiFastPolicyForm.rules =
        settings.openai_fast_policy_settings.rules.map((rule) => ({
          ...rule,
          user_ids: rule.user_ids ? [...rule.user_ids] : [],
          model_whitelist: rule.model_whitelist
            ? [...rule.model_whitelist]
            : [],
        }));
      openaiFastPolicyLoaded.value = true;
    }

    // Load web search emulation config separately
    await loadWebSearchConfig();
    await loadCreativeModelCandidates();
  } catch (error: unknown) {
    loadFailed.value = true;
    appStore.showError(
      extractApiErrorMessage(error, t("admin.settings.failedToLoad")),
    );
  } finally {
    loading.value = false;
    // 首次加载期间标签尚未挂载，渲染完成后补做一次横向定位。
    scrollActiveSettingsTabIntoView();
    scrollActiveGatewaySectionIntoView();
  }
  if (!loadFailed.value) {
    // 表单控件挂载时可能规整初始值，等渲染完成后再记录快照。
    await nextTick();
    markSettingsClean();
  }
}

async function loadSubscriptionPlans() {
  try {
    const response = await adminAPI.payment.getPlans();
    subscriptionPlans.value = response.data || [];
  } catch (_error: unknown) {
    subscriptionPlans.value = [];
  }
}

function findNextAvailableSubscriptionPlan(
  existingPlanIDs: number[],
): SubscriptionPlan | undefined {
  const existing = new Set(existingPlanIDs);
  return subscriptionPlans.value.find((plan) => !existing.has(plan.id));
}

function addDefaultSubscription() {
  if (subscriptionPlans.value.length === 0) return;
  const candidate = findNextAvailableSubscriptionPlan(
    form.default_subscriptions.map((item) => item.plan_id),
  );
  if (!candidate) return;
  form.default_subscriptions.push({
    plan_id: candidate.id,
  });
}

function removeDefaultSubscription(index: number) {
  form.default_subscriptions.splice(index, 1);
}

function addAuthSourceDefaultSubscription(source: AuthSourceType) {
  if (subscriptionPlans.value.length === 0) return;
  const candidate = findNextAvailableSubscriptionPlan(
    authSourceDefaults[source].subscriptions.map((item) => item.plan_id),
  );
  if (!candidate) return;
  authSourceDefaults[source].subscriptions.push({
    plan_id: candidate.id,
  });
}

function removeAuthSourceDefaultSubscription(
  source: AuthSourceType,
  index: number,
) {
  authSourceDefaults[source].subscriptions.splice(index, 1);
}

function findDuplicateDefaultSubscription(
  subscriptions: DefaultSubscriptionSetting[],
): DefaultSubscriptionSetting | undefined {
  const seenPlanIDs = new Set<number>();

  return subscriptions.find((item) => {
    if (seenPlanIDs.has(item.plan_id)) {
      return true;
    }
    seenPlanIDs.add(item.plan_id);
    return false;
  });
}

// 全局保存提交的数据分两组记录快照，联网搜索配置调用独立接口保存。
// 快照里去掉联网搜索的已用额度，它由“重置用量”按钮直接写回服务端。
const { dirty: settingsDirty, markClean: markSettingsClean } = useDirtyTracker({
  settings: () => ({
    form,
    authSourceDefaults,
    registrationEmailSuffixWhitelist: registrationEmailSuffixWhitelistTags.value,
    tablePageSizeOptions: tablePageSizeOptionsInput.value,
    openaiFastPolicyRules: openaiFastPolicyLoaded.value
      ? openaiFastPolicyForm.rules
      : null,
  }),
  webSearch: () => ({
    enabled: webSearchConfig.enabled,
    providers: webSearchConfig.providers.map(
      ({ quota_used: _quotaUsed, ...provider }) => provider,
    ),
  }),
});

async function saveSettings() {
  saving.value = true;
  try {
    const normalizedCreativeWorkerCount = Math.floor(Number(form.creative_worker_count));
    if (!Number.isSafeInteger(normalizedCreativeWorkerCount) || normalizedCreativeWorkerCount <= 0) {
      appStore.showError(t("admin.settings.features.creative.workerCountInvalid"));
      return;
    }
    form.creative_worker_count = normalizedCreativeWorkerCount;

    const rawDefaultUserAPIKeyLimit = String(
      form.default_user_api_key_limit,
    ).trim();
    const normalizedDefaultUserAPIKeyLimit = Number(rawDefaultUserAPIKeyLimit);
    if (
      rawDefaultUserAPIKeyLimit === "" ||
      !Number.isSafeInteger(normalizedDefaultUserAPIKeyLimit) ||
      normalizedDefaultUserAPIKeyLimit < 0 ||
      normalizedDefaultUserAPIKeyLimit > MAX_USER_API_KEY_LIMIT
    ) {
      appStore.showError(
        t("admin.settings.defaults.defaultUserApiKeyLimitInvalid"),
      );
      return;
    }
    form.default_user_api_key_limit = normalizedDefaultUserAPIKeyLimit;

    const normalizedTableDefaultPageSize = Math.floor(
      Number(form.table_default_page_size),
    );
    if (
      !Number.isInteger(normalizedTableDefaultPageSize) ||
      normalizedTableDefaultPageSize < tablePageSizeMin ||
      normalizedTableDefaultPageSize > tablePageSizeMax
    ) {
      appStore.showError(
        t("admin.settings.site.tableDefaultPageSizeRangeError", {
          min: tablePageSizeMin,
          max: tablePageSizeMax,
        }),
      );
      return;
    }

    const normalizedTablePageSizeOptions = parseTablePageSizeOptionsInput(
      tablePageSizeOptionsInput.value,
    );
    if (!normalizedTablePageSizeOptions) {
      appStore.showError(
        t("admin.settings.site.tablePageSizeOptionsFormatError", {
          min: tablePageSizeMin,
          max: tablePageSizeMax,
        }),
      );
      return;
    }

    const normalizedUsageRankingLimit = Math.floor(
      Number(form.usage_ranking_limit),
    );
    if (
      !Number.isInteger(normalizedUsageRankingLimit) ||
      normalizedUsageRankingLimit < usageRankingLimitMin ||
      normalizedUsageRankingLimit > usageRankingLimitMax
    ) {
      appStore.showError(
        t("admin.settings.site.usageRankingLimitRangeError", {
          min: usageRankingLimitMin,
          max: usageRankingLimitMax,
        }),
      );
      return;
    }

    form.table_default_page_size = normalizedTableDefaultPageSize;
    form.table_page_size_options = normalizedTablePageSizeOptions;
    form.usage_ranking_limit = normalizedUsageRankingLimit;
    ensureUsageRankingSortMetricVisible();
    form.balance_unit_name = form.balance_unit_name.trim() || "USD";
    form.balance_unit_symbol = form.balance_unit_symbol.trim() || "$";
    form.balance_icon_svg = form.balance_icon_svg.trim();
    form.reasoning_point_rmb_unit_price = Math.max(
      0,
      Number(form.reasoning_point_rmb_unit_price) || 0,
    );
    form.usd_exchange_rate = Math.max(0, Number(form.usd_exchange_rate) || 0);
    form.marketplace_availability_window_days = Math.min(
      marketplaceAvailabilityWindowDaysMax,
      Math.max(
        marketplaceAvailabilityWindowDaysMin,
        Math.floor(
          Number(form.marketplace_availability_window_days) ||
            marketplaceAvailabilityWindowDaysDefault,
        ),
      ),
    );
    form.marketplace_availability_bucket_minutes = Math.min(
      marketplaceAvailabilityBucketMinutesMax,
      Math.max(
        marketplaceAvailabilityBucketMinutesMin,
        Math.floor(
          Number(form.marketplace_availability_bucket_minutes) ||
            marketplaceAvailabilityBucketMinutesDefault,
        ),
      ),
    );
    const normalizedCreativeModelSettings = normalizeCreativeModelSettingsForSave();
    if (!normalizedCreativeModelSettings) {
      appStore.showError(
        t("admin.settings.features.creative.modelSettings.validationError"),
      );
      return;
    }
    if (
      form.openai_provider_quota_auto_pause.default_threshold_5h < 0 ||
      form.openai_provider_quota_auto_pause.default_threshold_5h > 1 ||
      form.openai_provider_quota_auto_pause.default_threshold_7d < 0 ||
      form.openai_provider_quota_auto_pause.default_threshold_7d > 1
    ) {
      appStore.showError(t("admin.settings.openaiQuotaAutoPause.rangeError"));
      return;
    }

    const normalizedLoginAgreementDocuments =
      normalizeLoginAgreementDocumentsForSave();
    if (form.login_agreement_enabled && normalizedLoginAgreementDocuments.length === 0) {
      appStore.showError(
        t("admin.settings.loginAgreement.documentsRequired"),
      );
      return;
    }
    const emptyTitleDocument = normalizedLoginAgreementDocuments.find(
      (doc) => !doc.title,
    );
    if (emptyTitleDocument) {
      appStore.showError(
        t("admin.settings.loginAgreement.documentTitleRequired"),
      );
      return;
    }
    const duplicateLoginAgreementDocumentId =
      findDuplicateLoginAgreementDocumentId(normalizedLoginAgreementDocuments);
    if (duplicateLoginAgreementDocumentId) {
      appStore.showError(
        t("admin.settings.loginAgreement.documentSlugDuplicate", {
          id: duplicateLoginAgreementDocumentId,
        }),
      );
      return;
    }
    form.login_agreement_mode =
      form.login_agreement_mode === "checkbox" ? "checkbox" : "modal";
    form.login_agreement_documents = normalizedLoginAgreementDocuments;
    form.forwarded_client_ip_headers = normalizeForwardedClientIpHeaders(
      form.forwarded_client_ip_headers,
    );

    const normalizedDefaultSubscriptions = normalizeDefaultSubscriptionSettings(
      form.default_subscriptions,
    );
    const duplicateDefaultSubscription = findDuplicateDefaultSubscription(
      normalizedDefaultSubscriptions,
    );
    if (duplicateDefaultSubscription) {
      appStore.showError(
        t("admin.settings.defaults.defaultSubscriptionsDuplicate", {
          planId: duplicateDefaultSubscription.plan_id,
        }),
      );
      return;
    }

    for (const authSource of authSourceDefaultsMeta.value) {
      authSourceDefaults[authSource.source].subscriptions =
        normalizeDefaultSubscriptionSettings(
          authSourceDefaults[authSource.source].subscriptions,
        );
      const duplicate = findDuplicateDefaultSubscription(
        authSourceDefaults[authSource.source].subscriptions,
      );
      if (duplicate) {
        appStore.showError(
          `${authSource.title}: ${t(
            "admin.settings.defaults.defaultSubscriptionsDuplicate",
            {
              planId: duplicate.plan_id,
            },
          )}`,
        );
        return;
      }
    }

    if (form.wechat_connect_mp_enabled && form.wechat_connect_mobile_enabled) {
      appStore.showError(
        t("admin.settings.wechatConnect.mpMobileConflict"),
      );
      return;
    }
    // 表单设置了 novalidate，URL 字段由此处校验。
    const isValidHttpUrl = (url: string): boolean => {
      if (!url) return true;
      try {
        const u = new URL(url);
        return u.protocol === "http:" || u.protocol === "https:";
      } catch {
        return false;
      }
    };
    // Optional URL fields: auto-clear invalid values so they don't cause backend 400 errors
    if (!isValidHttpUrl(form.frontend_url)) form.frontend_url = "";
    if (!isValidHttpUrl(form.doc_url)) form.doc_url = "";
    syncWeChatConnectMode();
    const wechatStoredMode = deriveWeChatConnectStoredMode(
      form.wechat_connect_open_enabled,
      form.wechat_connect_mp_enabled,
      form.wechat_connect_mobile_enabled,
      form.wechat_connect_mode,
    );

    const payload: UpdateSettingsRequest = {
      registration_enabled: form.registration_enabled,
      email_verify_enabled: form.email_verify_enabled,
      registration_email_suffix_whitelist:
        registrationEmailSuffixWhitelistTags.value.map((suffix) =>
          suffix.startsWith('*.') ? suffix : `@${suffix}`,
        ),
      registration_email_normalization:
        form.registration_email_normalization,
      registration_email_domain_quota_enabled:
        form.registration_email_domain_quota_enabled,
      user_email_change_enabled: form.user_email_change_enabled,
      promo_code_enabled: form.promo_code_enabled,
      invitation_code_enabled: form.invitation_code_enabled,
      password_reset_enabled: form.password_reset_enabled,
      totp_enabled: form.totp_enabled,
      session_binding_enabled: form.session_binding_enabled,
      step_up_enabled: form.step_up_enabled,
      // 清空数字框时 v-model.number 会得到空串，后端 int 字段解析空串会 400 拒绝整次保存；
      // 空/非法值回退默认 180（与后端 parseAuditLogRetentionDays("") 语义一致，0 仍表示永久保留）。
      audit_log_retention_days: Number.isFinite(form.audit_log_retention_days)
        ? form.audit_log_retention_days
        : 180,
      login_agreement_enabled: form.login_agreement_enabled,
      login_agreement_mode: form.login_agreement_mode,
      login_agreement_updated_at: form.login_agreement_updated_at,
      login_agreement_documents: form.login_agreement_documents,
      default_balance: form.default_balance,
      affiliate_enabled: form.affiliate_enabled,
      affiliate_rebate_rate: Math.min(
        100,
        Math.max(0, Number(form.affiliate_rebate_rate) || 0),
      ),
      affiliate_rebate_freeze_hours: Math.max(
        0,
        Math.min(720, Math.floor(Number(form.affiliate_rebate_freeze_hours) || 0)),
      ),
      affiliate_rebate_duration_days: Math.max(
        0,
        Math.min(3650, Math.floor(Number(form.affiliate_rebate_duration_days) || 0)),
      ),
      affiliate_rebate_per_invitee_cap: Math.max(
        0,
        Number(form.affiliate_rebate_per_invitee_cap) || 0,
      ),
      affiliate_admin_recharge_enabled: form.affiliate_admin_recharge_enabled,
      default_concurrency: form.default_concurrency,
      default_subscriptions: normalizedDefaultSubscriptions,
      balance_unit_name: form.balance_unit_name,
      balance_unit_symbol: form.balance_unit_symbol,
      balance_icon_svg: form.balance_icon_svg,
      reasoning_point_rmb_unit_price: form.reasoning_point_rmb_unit_price,
      usd_exchange_rate: form.usd_exchange_rate,
      marketplace_availability_window_days:
        form.marketplace_availability_window_days,
      marketplace_availability_bucket_minutes:
        form.marketplace_availability_bucket_minutes,
      force_email_on_third_party_signup: form.force_email_on_third_party_signup,
      default_user_rpm_limit: form.default_user_rpm_limit,
      default_user_api_key_limit: form.default_user_api_key_limit,
      site_name: form.site_name_zh || form.site_name_en || form.site_name,
      site_logo: form.site_logo,
      site_subtitle:
        form.site_subtitle_zh || form.site_subtitle_en || form.site_subtitle,
      site_name_zh: form.site_name_zh,
      site_name_en: form.site_name_en,
      site_title_zh: form.site_title_zh,
      site_title_en: form.site_title_en,
      site_subtitle_zh: form.site_subtitle_zh,
      site_subtitle_en: form.site_subtitle_en,
      api_base_url: form.api_base_url,
      contact_info: form.contact_info,
      doc_url: form.doc_url,
      home_content: form.home_content,
      backend_mode_enabled: form.backend_mode_enabled,
      hide_ccs_import_button: form.hide_ccs_import_button,
      table_default_page_size: form.table_default_page_size,
      table_page_size_options: form.table_page_size_options,
      usage_ranking_limit: form.usage_ranking_limit,
      usage_ranking_enabled: form.usage_ranking_enabled,
      usage_ranking_sort_by: form.usage_ranking_sort_by,
      usage_ranking_show_total_tokens: form.usage_ranking_show_total_tokens,
      usage_ranking_show_requests: form.usage_ranking_show_requests,
      usage_ranking_show_actual_cost: form.usage_ranking_show_actual_cost,
      custom_menu_items: form.custom_menu_items,
      custom_endpoints: form.custom_endpoints,
      footer_links: normalizeFooterLinksForSave(),
      footer_text: form.footer_text,
      home_featured_models: normalizeHomeFeaturedModelsForSave(),
      frontend_url: form.frontend_url,
      smtp_host: form.smtp_host,
      smtp_port: form.smtp_port,
      smtp_username: form.smtp_username,
      smtp_password: form.smtp_password || undefined,
      smtp_from_email: form.smtp_from_email,
      smtp_from_name: form.smtp_from_name,
      smtp_use_tls: form.smtp_use_tls,
      turnstile_enabled: form.turnstile_enabled,
      turnstile_site_key: form.turnstile_site_key,
      turnstile_secret_key: form.turnstile_secret_key || undefined,
      tencent_captcha_enabled: form.tencent_captcha_enabled,
      tencent_captcha_app_id: form.tencent_captcha_app_id,
      tencent_captcha_app_secret_key:
        form.tencent_captcha_app_secret_key || undefined,
      tencent_captcha_cloud_secret_id:
        form.tencent_captcha_cloud_secret_id || undefined,
      tencent_captcha_cloud_secret_key:
        form.tencent_captcha_cloud_secret_key || undefined,
      tencent_captcha_region: form.tencent_captcha_region,
      aliyun_captcha_enabled: form.aliyun_captcha_enabled,
      aliyun_captcha_access_key_id: form.aliyun_captcha_access_key_id,
      aliyun_captcha_access_key_secret:
        form.aliyun_captcha_access_key_secret || undefined,
      aliyun_captcha_scene_id: form.aliyun_captcha_scene_id,
      aliyun_captcha_prefix: form.aliyun_captcha_prefix,
      aliyun_captcha_region: form.aliyun_captcha_region,
      api_key_acl_trust_forwarded_ip: form.api_key_acl_trust_forwarded_ip,
      forwarded_client_ip_headers: form.forwarded_client_ip_headers,
      linuxdo_connect_enabled: form.linuxdo_connect_enabled,
      linuxdo_connect_client_id: form.linuxdo_connect_client_id,
      linuxdo_connect_client_secret:
        form.linuxdo_connect_client_secret || undefined,
      linuxdo_connect_redirect_url: form.linuxdo_connect_redirect_url,
      dingtalk_connect_enabled: form.dingtalk_connect_enabled,
      dingtalk_connect_client_id: form.dingtalk_connect_client_id,
      dingtalk_connect_client_secret:
        form.dingtalk_connect_client_secret || undefined,
      dingtalk_connect_redirect_url: form.dingtalk_connect_redirect_url,
      dingtalk_connect_corp_restriction_policy:
        form.dingtalk_connect_corp_restriction_policy,
      dingtalk_connect_internal_corp_id: form.dingtalk_connect_internal_corp_id,
      dingtalk_connect_bypass_registration: form.dingtalk_connect_bypass_registration,
      dingtalk_connect_sync_corp_email: form.dingtalk_connect_sync_corp_email,
      dingtalk_connect_sync_display_name: form.dingtalk_connect_sync_display_name,
      dingtalk_connect_sync_dept: form.dingtalk_connect_sync_dept,
      dingtalk_connect_sync_corp_email_attr_key: form.dingtalk_connect_sync_corp_email_attr_key,
      dingtalk_connect_sync_display_name_attr_key: form.dingtalk_connect_sync_display_name_attr_key,
      dingtalk_connect_sync_dept_attr_key: form.dingtalk_connect_sync_dept_attr_key,
      dingtalk_connect_sync_corp_email_attr_name: form.dingtalk_connect_sync_corp_email_attr_name,
      dingtalk_connect_sync_display_name_attr_name: form.dingtalk_connect_sync_display_name_attr_name,
      dingtalk_connect_sync_dept_attr_name: form.dingtalk_connect_sync_dept_attr_name,
      wechat_connect_enabled: form.wechat_connect_enabled,
      wechat_connect_app_id:
        form.wechat_connect_open_app_id ||
        form.wechat_connect_mp_app_id ||
        form.wechat_connect_mobile_app_id ||
        form.wechat_connect_app_id,
      wechat_connect_app_secret: form.wechat_connect_app_secret || undefined,
      wechat_connect_open_app_id: form.wechat_connect_open_app_id,
      wechat_connect_open_app_secret:
        form.wechat_connect_open_app_secret || undefined,
      wechat_connect_mp_app_id: form.wechat_connect_mp_app_id,
      wechat_connect_mp_app_secret:
        form.wechat_connect_mp_app_secret || undefined,
      wechat_connect_mobile_app_id: form.wechat_connect_mobile_app_id,
      wechat_connect_mobile_app_secret:
        form.wechat_connect_mobile_app_secret || undefined,
      wechat_connect_open_enabled: form.wechat_connect_open_enabled,
      wechat_connect_mp_enabled: form.wechat_connect_mp_enabled,
      wechat_connect_mobile_enabled: form.wechat_connect_mobile_enabled,
      wechat_connect_mode: wechatStoredMode,
      wechat_connect_scopes:
        defaultWeChatConnectScopesForMode(wechatStoredMode),
      wechat_connect_redirect_url: form.wechat_connect_redirect_url,
      wechat_connect_frontend_redirect_url:
        form.wechat_connect_frontend_redirect_url,
      oidc_connect_enabled: form.oidc_connect_enabled,
      oidc_connect_provider_name: form.oidc_connect_provider_name,
      oidc_connect_client_id: form.oidc_connect_client_id,
      oidc_connect_client_secret: form.oidc_connect_client_secret || undefined,
      oidc_connect_issuer_url: form.oidc_connect_issuer_url,
      oidc_connect_discovery_url: form.oidc_connect_discovery_url,
      oidc_connect_authorize_url: form.oidc_connect_authorize_url,
      oidc_connect_token_url: form.oidc_connect_token_url,
      oidc_connect_userinfo_url: form.oidc_connect_userinfo_url,
      oidc_connect_jwks_url: form.oidc_connect_jwks_url,
      oidc_connect_scopes: form.oidc_connect_scopes,
      oidc_connect_redirect_url: form.oidc_connect_redirect_url,
      oidc_connect_frontend_redirect_url:
        form.oidc_connect_frontend_redirect_url,
      oidc_connect_token_auth_method: form.oidc_connect_token_auth_method,
      oidc_connect_use_pkce: form.oidc_connect_use_pkce,
      oidc_connect_validate_id_token: form.oidc_connect_validate_id_token,
      oidc_connect_allowed_signing_algs: form.oidc_connect_allowed_signing_algs,
      oidc_connect_clock_skew_seconds: form.oidc_connect_clock_skew_seconds,
      oidc_connect_require_email_verified:
        form.oidc_connect_require_email_verified,
      oidc_connect_userinfo_email_path: form.oidc_connect_userinfo_email_path,
      oidc_connect_userinfo_id_path: form.oidc_connect_userinfo_id_path,
      oidc_connect_userinfo_username_path:
        form.oidc_connect_userinfo_username_path,
      github_oauth_enabled: form.github_oauth_enabled,
      github_oauth_client_id: form.github_oauth_client_id,
      github_oauth_client_secret:
        form.github_oauth_client_secret || undefined,
      github_oauth_redirect_url: form.github_oauth_redirect_url,
      github_oauth_frontend_redirect_url:
        form.github_oauth_frontend_redirect_url,
      google_oauth_enabled: form.google_oauth_enabled,
      google_one_tap_enabled: form.google_one_tap_enabled,
      google_oauth_client_id: form.google_oauth_client_id,
      google_oauth_client_secret:
        form.google_oauth_client_secret || undefined,
      google_oauth_redirect_url: form.google_oauth_redirect_url,
      google_oauth_frontend_redirect_url:
        form.google_oauth_frontend_redirect_url,
      enable_model_fallback: form.enable_model_fallback,
      fallback_model_anthropic: form.fallback_model_anthropic,
      fallback_model_openai: form.fallback_model_openai,
      fallback_model_gemini: form.fallback_model_gemini,
      fallback_model_antigravity: form.fallback_model_antigravity,
      grok_default_text_model:
        form.grok_default_text_model.trim() || "grok-4.5",
      grok_default_base_url_mode: form.grok_default_base_url_mode,
      enable_identity_patch: form.enable_identity_patch,
      identity_patch_prompt: form.identity_patch_prompt,
      min_claude_code_version: form.min_claude_code_version,
      max_claude_code_version: form.max_claude_code_version,
      openai_ttft_mode:
        form.openai_ttft_mode === "visible" ? "visible" : "semantic",
      enable_fingerprint_unification: form.enable_fingerprint_unification,
      enable_metadata_passthrough: form.enable_metadata_passthrough,
      enable_claude_oauth_system_prompt_injection:
        form.enable_claude_oauth_system_prompt_injection,
      claude_oauth_system_prompt:
        form.claude_oauth_system_prompt?.trim() || "",
      claude_oauth_system_prompt_blocks:
        form.claude_oauth_system_prompt_blocks?.trim() || "",
      enable_anthropic_cache_ttl_1h_injection:
        form.enable_anthropic_cache_ttl_1h_injection,
      rewrite_message_cache_control: form.rewrite_message_cache_control,
      enable_client_dateline_normalization:
        form.enable_client_dateline_normalization,
      antigravity_user_agent_version:
        form.antigravity_user_agent_version?.trim() || "",
      openai_codex_user_agent:
        form.openai_codex_user_agent?.trim() || "",
      openai_allow_claude_code_codex_plugin: form.openai_allow_claude_code_codex_plugin,
      user_prompt_replacement_config:
        normalizeUserPromptReplacementConfigForSave(),
      // Payment configuration
      payment_enabled: form.payment_enabled,
      // 页面功能开关
      team_enabled: form.team_enabled,
      creative_enabled: form.creative_enabled,
      creative_model_settings: normalizedCreativeModelSettings,
      creative_worker_count: normalizedCreativeWorkerCount,
      risk_control_enabled: form.risk_control_enabled,
      cyber_session_block_enabled: form.cyber_session_block_enabled,
      cyber_session_block_ttl_seconds: Math.max(
        1,
        Math.floor(Number(form.cyber_session_block_ttl_seconds) || 3600),
      ),
      payment_min_amount: Number(form.payment_min_amount) || 0,
      payment_max_amount: Number(form.payment_max_amount) || 0,
      payment_daily_limit: Number(form.payment_daily_limit) || 0,
      payment_max_pending_orders: Number(form.payment_max_pending_orders) || 0,
      payment_order_timeout_minutes:
        Number(form.payment_order_timeout_minutes) || 0,
      payment_balance_disabled: form.payment_balance_disabled,
      payment_balance_recharge_multiplier:
        Number(form.payment_balance_recharge_multiplier) || 1,
      payment_subscription_usd_to_cny_rate:
        Number(form.payment_subscription_usd_to_cny_rate) || 0,
      payment_recharge_fee_rate: Number(form.payment_recharge_fee_rate) || 0,
      payment_method_fees: form.payment_method_fees,
      payment_enabled_types: form.payment_enabled_types,
      payment_load_balance_strategy: form.payment_load_balance_strategy,
      payment_product_name_prefix: form.payment_product_name_prefix,
      payment_product_name_suffix: form.payment_product_name_suffix,
      payment_help_image_url: form.payment_help_image_url,
      payment_help_text: form.payment_help_text,
      payment_cancel_rate_limit_enabled: form.payment_cancel_rate_limit_enabled,
      payment_cancel_rate_limit_max:
        Number(form.payment_cancel_rate_limit_max) || 10,
      payment_cancel_rate_limit_window:
        Number(form.payment_cancel_rate_limit_window) || 1,
      payment_cancel_rate_limit_unit: form.payment_cancel_rate_limit_unit,
      payment_cancel_rate_limit_window_mode:
        form.payment_cancel_rate_limit_window_mode,
      payment_alipay_force_qrcode: form.payment_alipay_force_qrcode,
      payment_alipay_mobile_precreate_deep_link:
        form.payment_alipay_mobile_precreate_deep_link,
      advanced_scheduler_sticky_weighted_enabled:
        form.advanced_scheduler_sticky_weighted_enabled,
      advanced_scheduler_subscription_priority_enabled:
        form.advanced_scheduler_subscription_priority_enabled,
      advanced_scheduler_ewma_error_rate_alpha:
        form.advanced_scheduler_ewma_error_rate_alpha.trim(),
      advanced_scheduler_ewma_ttft_alpha:
        form.advanced_scheduler_ewma_ttft_alpha.trim(),
      advanced_scheduler_sticky_escape_enabled:
        form.advanced_scheduler_sticky_escape_enabled,
      advanced_scheduler_sticky_escape_ttft_ms:
        form.advanced_scheduler_sticky_escape_ttft_ms.trim(),
      advanced_scheduler_sticky_escape_error_rate:
        form.advanced_scheduler_sticky_escape_error_rate.trim(),
      advanced_scheduler_lb_top_k:
        form.advanced_scheduler_lb_top_k.trim(),
      advanced_scheduler_weight_priority:
        form.advanced_scheduler_weight_priority.trim(),
      advanced_scheduler_weight_load:
        form.advanced_scheduler_weight_load.trim(),
      advanced_scheduler_weight_queue:
        form.advanced_scheduler_weight_queue.trim(),
      advanced_scheduler_weight_error_rate:
        form.advanced_scheduler_weight_error_rate.trim(),
      advanced_scheduler_weight_ttft:
        form.advanced_scheduler_weight_ttft.trim(),
      advanced_scheduler_weight_reset:
        form.advanced_scheduler_weight_reset.trim(),
      advanced_scheduler_weight_quota_headroom:
        form.advanced_scheduler_weight_quota_headroom.trim(),
      advanced_scheduler_weight_previous_response:
        form.advanced_scheduler_weight_previous_response.trim(),
      advanced_scheduler_weight_session_sticky:
        form.advanced_scheduler_weight_session_sticky.trim(),
      openai_provider_quota_auto_pause: {
        default_threshold_5h:
          form.openai_provider_quota_auto_pause.default_threshold_5h,
        default_threshold_7d:
          form.openai_provider_quota_auto_pause.default_threshold_7d,
      },
      // 余额、订阅到期与提供商限额通知
      balance_low_notify_enabled: form.balance_low_notify_enabled,
      balance_low_notify_threshold:
        Number(form.balance_low_notify_threshold) || 0,
      balance_low_notify_recharge_url: (form.balance_low_notify_recharge_url =
        form.balance_low_notify_recharge_url || currentOrigin),
      subscription_expiry_notify_enabled:
        form.subscription_expiry_notify_enabled,
      provider_quota_notify_enabled: form.provider_quota_notify_enabled,
      provider_quota_notify_emails: (
        form.provider_quota_notify_emails || []
      ).filter((e) => e.email.trim() !== ""),
      allow_user_view_error_requests: form.allow_user_view_error_requests,
    };

    // 仅当 openai_fast_policy_settings 已成功从后端加载时才回写，
    // 否则省略整个字段，让后端保留既有规则（含默认值）。
    if (openaiFastPolicyLoaded.value) {
      payload.openai_fast_policy_settings = {
        rules: openaiFastPolicyForm.rules.map((rule) => {
          const whitelist = (rule.model_whitelist || [])
            .map((p) => p.trim())
            .filter((p) => p !== "");
          const hasWhitelist = whitelist.length > 0;
          return {
            service_tier: rule.service_tier,
            action: rule.action,
            scope: rule.scope,
            user_ids:
              rule.user_ids && rule.user_ids.length > 0
                ? [...rule.user_ids]
                : undefined,
            error_message:
              rule.action === "block" ? rule.error_message : undefined,
            model_whitelist: hasWhitelist ? whitelist : undefined,
            fallback_action: hasWhitelist
              ? rule.fallback_action || "pass"
              : undefined,
            fallback_error_message:
              hasWhitelist && rule.fallback_action === "block"
                ? rule.fallback_error_message
                : undefined,
          };
        }),
      };
    }

    payload.provider_scheduling_thresholds = sanitizeProviderSchedulingThresholdsMap(
      form.provider_scheduling_thresholds,
    );
    appendAuthSourceDefaultsToUpdateRequest(payload, authSourceDefaults);

    const updated = await settingsStepUp.run(() =>
      adminAPI.settings.updateSettings(payload),
    );
    for (const [key, value] of Object.entries(updated)) {
      if (key === "openai_fast_policy_settings") continue;
      if (value !== null && value !== undefined) {
        (form as Record<string, unknown>)[key] = value;
      }
    }
    Object.assign(authSourceDefaults, buildAuthSourceDefaultsState(updated));
    form.openai_provider_quota_auto_pause = {
      default_threshold_5h:
        updated.openai_provider_quota_auto_pause?.default_threshold_5h ?? 0,
      default_threshold_7d:
        updated.openai_provider_quota_auto_pause?.default_threshold_7d ?? 0,
    };
    registrationEmailSuffixWhitelistTags.value =
      normalizeRegistrationEmailSuffixDomains(
        updated.registration_email_suffix_whitelist,
      );
    form.forwarded_client_ip_headers = normalizeForwardedClientIpHeaders(
      updated.forwarded_client_ip_headers,
    );
    forwardedClientIpHeaderDraft.value = "";
    tablePageSizeOptionsInput.value = formatTablePageSizeOptions(
      Array.isArray(updated.table_page_size_options)
        ? updated.table_page_size_options
        : [10, 20, 50, 100],
    );
    registrationEmailSuffixWhitelistDraft.value = "";
    form.smtp_password = "";
    smtpPasswordManuallyEdited.value = false;
    form.turnstile_secret_key = "";
    form.aliyun_captcha_access_key_secret = "";
    form.linuxdo_connect_client_secret = "";
    form.dingtalk_connect_client_secret = "";
    form.wechat_connect_app_secret = "";
    form.wechat_connect_open_app_secret = "";
    form.wechat_connect_mp_app_secret = "";
    form.wechat_connect_mobile_app_secret = "";
    const updatedWechatCapabilities = resolveWeChatConnectModeCapabilities(
      updated.wechat_connect_open_enabled,
      updated.wechat_connect_mp_enabled,
      updated.wechat_connect_mobile_enabled,
      updated.wechat_connect_mode,
    );
    form.wechat_connect_open_enabled = updatedWechatCapabilities.openEnabled;
    form.wechat_connect_mp_enabled = updatedWechatCapabilities.mpEnabled;
    form.wechat_connect_mobile_enabled =
      updatedWechatCapabilities.mobileEnabled;
    form.wechat_connect_mode = deriveWeChatConnectStoredMode(
      updatedWechatCapabilities.openEnabled,
      updatedWechatCapabilities.mpEnabled,
      updatedWechatCapabilities.mobileEnabled,
      updated.wechat_connect_mode,
    );
    form.wechat_connect_scopes = defaultWeChatConnectScopesForMode(
      form.wechat_connect_mode,
    );
    form.oidc_connect_client_secret = "";
    form.github_oauth_client_secret = "";
    form.google_oauth_client_secret = "";
    // Refresh OpenAI fast/flex policy from server response
    if (
      updated.openai_fast_policy_settings &&
      Array.isArray(updated.openai_fast_policy_settings.rules)
    ) {
      openaiFastPolicyForm.rules =
        updated.openai_fast_policy_settings.rules.map((rule) => ({
          ...rule,
          user_ids: rule.user_ids ? [...rule.user_ids] : [],
          model_whitelist: rule.model_whitelist
            ? [...rule.model_whitelist]
            : [],
        }));
      openaiFastPolicyLoaded.value = true;
    }
    // Save web search emulation config separately (errors handled internally)
    const wsOk = await saveWebSearchConfig();
    // Refresh cached settings so sidebar/header update immediately
    await appStore.fetchPublicSettings(true);
    await adminSettingsStore.fetch(true);
    await nextTick();
    // 联网搜索配置保存失败时保留它的未保存状态，用户可以修正后再次提交。
    if (wsOk) {
      markSettingsClean();
    } else {
      markSettingsClean("settings");
    }
    if (wsOk) {
      appStore.showSuccess(t("admin.settings.settingsSaved"));
    }
  } catch (error: unknown) {
    // 用户取消 step-up 验证：静默返回，不弹错误
    if (isStepUpCancelled(error)) {
      return;
    }
    if (isStepUpBlocked(error)) {
      appStore.showError(
        stepUpBlockReason(error) === "STEP_UP_ADMIN_API_KEY_FORBIDDEN"
          ? t("stepUp.adminApiKeyForbidden")
          : t("stepUp.notEnabled"),
      );
      return;
    }
    // 开启 step-up 开关但本人未启用 2FA：给出可操作的专用提示
    if (
      (error as { reason?: string })?.reason === "STEP_UP_ENABLE_REQUIRES_TOTP"
    ) {
      appStore.showError(t("admin.settings.security.stepUpEnableRequiresTotp"));
      return;
    }
    appStore.showError(
      extractApiErrorMessage(error, t("admin.settings.failedToSave")),
    );
  } finally {
    saving.value = false;
  }
}

async function testSmtpConnection() {
  testingSmtp.value = true;
  try {
    const smtpPasswordForTest = smtpPasswordManuallyEdited.value
      ? form.smtp_password
      : "";
    const result = await adminAPI.settings.testSmtpConnection({
      smtp_host: form.smtp_host,
      smtp_port: form.smtp_port,
      smtp_username: form.smtp_username,
      smtp_password: smtpPasswordForTest,
      smtp_use_tls: form.smtp_use_tls,
    });
    // API returns { message: "..." } on success, errors are thrown as exceptions
    appStore.showSuccess(
      result.message || t("admin.settings.smtpConnectionSuccess"),
    );
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t("admin.settings.failedToTestSmtp")),
    );
  } finally {
    testingSmtp.value = false;
  }
}

async function sendTestEmail() {
  if (!testEmailAddress.value) {
    appStore.showError(t("admin.settings.testEmail.enterRecipientHint"));
    return;
  }

  sendingTestEmail.value = true;
  try {
    const smtpPasswordForSend = smtpPasswordManuallyEdited.value
      ? form.smtp_password
      : "";
    const result = await adminAPI.settings.sendTestEmail({
      email: testEmailAddress.value,
      smtp_host: form.smtp_host,
      smtp_port: form.smtp_port,
      smtp_username: form.smtp_username,
      smtp_password: smtpPasswordForSend,
      smtp_from_email: form.smtp_from_email,
      smtp_from_name: form.smtp_from_name,
      smtp_use_tls: form.smtp_use_tls,
    });
    // API returns { message: "..." } on success, errors are thrown as exceptions
    appStore.showSuccess(result.message || t("admin.settings.testEmailSent"));
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t("admin.settings.failedToSendTestEmail")),
    );
  } finally {
    sendingTestEmail.value = false;
  }
}

// Admin API Key 方法
async function loadAdminApiKey() {
  adminApiKeyLoading.value = true;
  try {
    const status = await adminAPI.settings.getAdminApiKey();
    adminApiKeyExists.value = status.exists;
    adminApiKeyMasked.value = status.masked_key;
  } catch (_error: unknown) {
    // Silent fail - admin API key status is non-critical
  } finally {
    adminApiKeyLoading.value = false;
  }
}

async function createAdminApiKey() {
  adminApiKeyOperating.value = true;
  try {
    const result = await adminAPI.settings.regenerateAdminApiKey();
    newAdminApiKey.value = result.key;
    adminApiKeyExists.value = true;
    adminApiKeyMasked.value =
      result.key.substring(0, 10) + "..." + result.key.slice(-4);
    appStore.showSuccess(t("admin.settings.adminApiKey.keyGenerated"));
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t("common.error")));
  } finally {
    adminApiKeyOperating.value = false;
  }
}

async function regenerateAdminApiKey() {
  if (!confirm(t("admin.settings.adminApiKey.regenerateConfirm"))) return;
  await createAdminApiKey();
}

async function deleteAdminApiKey() {
  if (!confirm(t("admin.settings.adminApiKey.deleteConfirm"))) return;
  adminApiKeyOperating.value = true;
  try {
    await adminAPI.settings.deleteAdminApiKey();
    adminApiKeyExists.value = false;
    adminApiKeyMasked.value = "";
    newAdminApiKey.value = "";
    appStore.showSuccess(t("admin.settings.adminApiKey.keyDeleted"));
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t("common.error")));
  } finally {
    adminApiKeyOperating.value = false;
  }
}

function copyNewKey() {
  navigator.clipboard
    .writeText(newAdminApiKey.value)
    .then(() => {
      appStore.showSuccess(t("admin.settings.adminApiKey.keyCopied"));
    })
    .catch(() => {
      appStore.showError(t("common.copyFailed"));
    });
}

async function loadOllamaCloudUsageSettings() {
  ollamaCloudUsageLoading.value = true;
  try {
    Object.assign(
      ollamaCloudUsageForm,
      await adminAPI.providers.getOllamaCloudUsageSettings(),
    );
  } catch (_error: unknown) {
    // 可选设置加载失败时保留默认关闭的安全配置。
  } finally {
    ollamaCloudUsageLoading.value = false;
  }
}

async function saveOllamaCloudUsageSettings() {
  ollamaCloudUsageSaving.value = true;
  try {
    const updated = await adminAPI.providers.updateOllamaCloudUsageSettings({
      ...ollamaCloudUsageForm,
    });
    Object.assign(ollamaCloudUsageForm, updated);
    appStore.showSuccess(t("admin.settings.ollamaCloudUsage.saved"));
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t("admin.settings.ollamaCloudUsage.saveFailed")),
    );
  } finally {
    ollamaCloudUsageSaving.value = false;
  }
}

// Overload Cooldown 方法
async function loadOverloadCooldownSettings() {
  overloadCooldownLoading.value = true;
  try {
    const settings = await adminAPI.settings.getOverloadCooldownSettings();
    Object.assign(overloadCooldownForm, settings);
  } catch (_error: unknown) {
    // Silent fail - settings will use defaults
  } finally {
    overloadCooldownLoading.value = false;
  }
}

async function saveOverloadCooldownSettings() {
  overloadCooldownSaving.value = true;
  try {
    const updated = await adminAPI.settings.updateOverloadCooldownSettings({
      enabled: overloadCooldownForm.enabled,
      cooldown_minutes: overloadCooldownForm.cooldown_minutes,
    });
    Object.assign(overloadCooldownForm, updated);
    appStore.showSuccess(t("admin.settings.overloadCooldown.saved"));
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(
        error,
        t("admin.settings.overloadCooldown.saveFailed"),
      ),
    );
  } finally {
    overloadCooldownSaving.value = false;
  }
}

async function loadOpenAI403CooldownSettings() {
  openAI403CooldownLoading.value = true;
  try {
    const settings = await adminAPI.settings.getOpenAI403CooldownSettings();
    Object.assign(openAI403CooldownForm, settings);
  } catch (_error: unknown) {
    // Silent fail - settings will use defaults
  } finally {
    openAI403CooldownLoading.value = false;
  }
}

async function saveOpenAI403CooldownSettings() {
  openAI403CooldownSaving.value = true;
  try {
    const updated = await adminAPI.settings.updateOpenAI403CooldownSettings({
      enabled: openAI403CooldownForm.enabled,
      cooldown_minutes: openAI403CooldownForm.cooldown_minutes,
      error_on_threshold_enabled: openAI403CooldownForm.error_on_threshold_enabled,
      threshold_count: openAI403CooldownForm.threshold_count,
      threshold_window_minutes: openAI403CooldownForm.threshold_window_minutes,
    });
    Object.assign(openAI403CooldownForm, updated);
    appStore.showSuccess(t("admin.settings.openAI403Cooldown.saved"));
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(
        error,
        t("admin.settings.openAI403Cooldown.saveFailed"),
      ),
    );
  } finally {
    openAI403CooldownSaving.value = false;
  }
}

// 面板 API 限流方法
async function loadPanelRateLimitSettings() {
  panelRateLimitLoading.value = true;
  try {
    const settings = await adminAPI.settings.getPanelRateLimitSettings();
    Object.assign(panelRateLimitForm, settings);
  } catch (_error: unknown) {
    // 静默失败，界面继续使用默认值。
  } finally {
    panelRateLimitLoading.value = false;
  }
}

async function savePanelRateLimitSettings() {
  panelRateLimitSaving.value = true;
  try {
    const updated = await adminAPI.settings.updatePanelRateLimitSettings({
      enabled: panelRateLimitForm.enabled,
      user_rpm: panelRateLimitForm.user_rpm,
      heavy_rpm: panelRateLimitForm.heavy_rpm,
      exempt_admin: panelRateLimitForm.exempt_admin,
      public_ip_rpm: panelRateLimitForm.public_ip_rpm,
    });
    Object.assign(panelRateLimitForm, updated);
    appStore.showSuccess(t("admin.settings.panelRateLimit.saved"));
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(
        error,
        t("admin.settings.panelRateLimit.saveFailed"),
      ),
    );
  } finally {
    panelRateLimitSaving.value = false;
  }
}

// Rate Limit Cooldown (429) 方法
async function loadRateLimit429CooldownSettings() {
  rateLimit429CooldownLoading.value = true;
  try {
    const settings = await adminAPI.settings.getRateLimit429CooldownSettings();
    Object.assign(rateLimit429CooldownForm, settings);
  } catch (_error: unknown) {
    // Silent fail - settings will use defaults
  } finally {
    rateLimit429CooldownLoading.value = false;
  }
}

async function saveRateLimit429CooldownSettings() {
  rateLimit429CooldownSaving.value = true;
  try {
    const updated = await adminAPI.settings.updateRateLimit429CooldownSettings({
      enabled: rateLimit429CooldownForm.enabled,
      cooldown_seconds: rateLimit429CooldownForm.cooldown_seconds,
    });
    Object.assign(rateLimit429CooldownForm, updated);
    appStore.showSuccess(t("admin.settings.rateLimit429Cooldown.saved"));
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(
        error,
        t("admin.settings.rateLimit429Cooldown.saveFailed"),
      ),
    );
  } finally {
    rateLimit429CooldownSaving.value = false;
  }
}

// Stream Timeout 方法
async function loadStreamTimeoutSettings() {
  streamTimeoutLoading.value = true;
  try {
    const settings = await adminAPI.settings.getStreamTimeoutSettings();
    Object.assign(streamTimeoutForm, settings);
  } catch (_error: unknown) {
    // Silent fail - settings will use defaults
  } finally {
    streamTimeoutLoading.value = false;
  }
}

async function saveStreamTimeoutSettings() {
  streamTimeoutSaving.value = true;
  try {
    const updated = await adminAPI.settings.updateStreamTimeoutSettings({
      enabled: streamTimeoutForm.enabled,
      action: streamTimeoutForm.action,
      temp_unsched_minutes: streamTimeoutForm.temp_unsched_minutes,
      threshold_count: streamTimeoutForm.threshold_count,
      threshold_window_minutes: streamTimeoutForm.threshold_window_minutes,
    });
    Object.assign(streamTimeoutForm, updated);
    appStore.showSuccess(t("admin.settings.streamTimeout.saved"));
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(
        error,
        t("admin.settings.streamTimeout.saveFailed"),
      ),
    );
  } finally {
    streamTimeoutSaving.value = false;
  }
}

// Rectifier 方法
async function loadRectifierSettings() {
  rectifierLoading.value = true;
  try {
    const settings = await adminAPI.settings.getRectifierSettings();
    Object.assign(rectifierForm, settings);
    // 确保 patterns 是数组（旧数据可能为 null）
    if (!Array.isArray(rectifierForm.apikey_signature_patterns)) {
      rectifierForm.apikey_signature_patterns = [];
    }
  } catch (_error: unknown) {
    // Silent fail - settings will use defaults
  } finally {
    rectifierLoading.value = false;
  }
}

async function saveRectifierSettings() {
  rectifierSaving.value = true;
  try {
    const updated = await adminAPI.settings.updateRectifierSettings({
      enabled: rectifierForm.enabled,
      thinking_signature_enabled: rectifierForm.thinking_signature_enabled,
      thinking_budget_enabled: rectifierForm.thinking_budget_enabled,
      apikey_signature_enabled: rectifierForm.apikey_signature_enabled,
      apikey_signature_patterns: rectifierForm.apikey_signature_patterns.filter(
        (p) => p.trim() !== "",
      ),
    });
    Object.assign(rectifierForm, updated);
    if (!Array.isArray(rectifierForm.apikey_signature_patterns)) {
      rectifierForm.apikey_signature_patterns = [];
    }
    appStore.showSuccess(t("admin.settings.rectifier.saved"));
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t("admin.settings.rectifier.saveFailed")),
    );
  } finally {
    rectifierSaving.value = false;
  }
}

const betaPolicyActionOptions = computed(() => [
  { value: "pass", label: t("admin.settings.betaPolicy.actionPass") },
  { value: "filter", label: t("admin.settings.betaPolicy.actionFilter") },
  { value: "block", label: t("admin.settings.betaPolicy.actionBlock") },
]);

const betaPolicyScopeOptions = computed(() => [
  { value: "all", label: t("admin.settings.betaPolicy.scopeAll") },
  { value: "oauth", label: t("admin.settings.betaPolicy.scopeOAuth") },
  { value: "apikey", label: t("admin.settings.betaPolicy.scopeAPIKey") },
  { value: "bedrock", label: t("admin.settings.betaPolicy.scopeBedrock") },
]);

// Beta Policy 方法
const betaDisplayNames: Record<string, string> = {
  "fast-mode-2026-02-01": "Fast Mode",
  "context-1m-2025-08-07": "Context 1M",
};

// 快捷预设：按 beta_token 定义预设方案
const betaPresets: Record<
  string,
  Array<{
    label: string;
    description: string;
    action: "pass" | "filter" | "block";
    model_whitelist: string[];
    fallback_action: "pass" | "filter" | "block";
  }>
> = {
  "context-1m-2025-08-07": [
    {
      label: t("admin.settings.betaPolicy.presetOpusOnly"),
      description: t("admin.settings.betaPolicy.presetOpusOnlyDesc"),
      action: "pass",
      model_whitelist: ["claude-opus-4-6"],
      fallback_action: "filter",
    },
  ],
};

// 常用模型模式（具体 ID + 通配符示例）
const commonModelPatterns = [
  "claude-opus-4-6",
  "claude-sonnet-4-6",
  "claude-opus-*",
  "claude-sonnet-*",
];

function getBetaDisplayName(token: string): string {
  return betaDisplayNames[token] || token;
}

function applyBetaPreset(
  rule: (typeof betaPolicyForm.rules)[number],
  preset: {
    action: "pass" | "filter" | "block";
    model_whitelist: string[];
    fallback_action: "pass" | "filter" | "block";
  },
) {
  rule.action = preset.action;
  rule.model_whitelist = [...preset.model_whitelist];
  rule.fallback_action = preset.fallback_action;
}

function addQuickPattern(
  rule: (typeof betaPolicyForm.rules)[number],
  pattern: string,
) {
  if (!rule.model_whitelist) rule.model_whitelist = [];
  if (!rule.model_whitelist.includes(pattern)) {
    rule.model_whitelist.push(pattern);
  }
}

async function loadBetaPolicySettings() {
  betaPolicyLoading.value = true;
  try {
    const settings = await adminAPI.settings.getBetaPolicySettings();
    betaPolicyForm.rules = settings.rules;
  } catch (_error: unknown) {
    // Silent fail - settings will use defaults
  } finally {
    betaPolicyLoading.value = false;
  }
}

// ==================== OpenAI Fast/Flex Policy ====================

const openaiFastPolicyTierOptions = computed(() => [
  { value: "all", label: t("admin.settings.openaiFastPolicy.tierAll") },
  {
    value: "priority",
    label: t("admin.settings.openaiFastPolicy.tierPriority"),
  },
  {
    value: "ultrafast",
    label: t("admin.settings.openaiFastPolicy.tierUltrafast"),
  },
  { value: "flex", label: t("admin.settings.openaiFastPolicy.tierFlex") },
]);

const openaiFastPolicyActionOptions = computed(() => [
  { value: "pass", label: t("admin.settings.openaiFastPolicy.actionPass") },
  { value: "filter", label: t("admin.settings.openaiFastPolicy.actionFilter") },
  {
    value: "force_priority",
    label: t("admin.settings.openaiFastPolicy.actionForcePriority"),
  },
  {value:"force_ultrafast",label:t("admin.settings.openaiFastPolicy.actionForceUltrafast")},
  { value: "block", label: t("admin.settings.openaiFastPolicy.actionBlock") },
]);

function openaiFastPolicyActionSummary(
  action: OpenAIFastPolicyRule["action"],
) {
  return t(`admin.settings.openaiFastPolicy.summaryAction.${action}`);
}

function hasOpenAIFastPolicyTargetModels(rule: OpenAIFastPolicyRule) {
  return Boolean(rule.model_whitelist?.some((pattern) => pattern.trim() !== ""));
}

const openaiFastPolicyScopeOptions = computed(() => [
  { value: "all", label: t("admin.settings.openaiFastPolicy.scopeAll") },
  { value: "oauth", label: t("admin.settings.openaiFastPolicy.scopeOAuth") },
  { value: "apikey", label: t("admin.settings.openaiFastPolicy.scopeAPIKey") },
  {
    value: "bedrock",
    label: t("admin.settings.openaiFastPolicy.scopeBedrock"),
  },
]);

function addOpenAIFastPolicyRule() {
  openaiFastPolicyForm.rules.push({
    service_tier: "priority",
    action: "filter",
    scope: "all",
    user_ids: [],
    error_message: "",
    model_whitelist: [],
    fallback_action: "pass",
    fallback_error_message: "",
  });
}

function removeOpenAIFastPolicyRule(index: number) {
  openaiFastPolicyForm.rules.splice(index, 1);
}

function addOpenAIFastPolicyModelPattern(rule: OpenAIFastPolicyRule) {
  if (!rule.model_whitelist) rule.model_whitelist = [];
  rule.model_whitelist.push("");
}

function removeOpenAIFastPolicyModelPattern(
  rule: OpenAIFastPolicyRule,
  idx: number,
) {
  rule.model_whitelist?.splice(idx, 1);
}

async function saveBetaPolicySettings() {
  betaPolicySaving.value = true;
  try {
    // Clean up empty patterns before saving
    const cleanedRules = betaPolicyForm.rules.map((rule) => {
      const whitelist = rule.model_whitelist?.filter((p) => p.trim() !== "");
      const hasWhitelist = whitelist && whitelist.length > 0;
      return {
        beta_token: rule.beta_token,
        action: rule.action,
        scope: rule.scope,
        error_message: rule.error_message,
        model_whitelist: hasWhitelist ? whitelist : undefined,
        fallback_action: hasWhitelist
          ? rule.fallback_action || "pass"
          : undefined,
        fallback_error_message:
          hasWhitelist && rule.fallback_action === "block"
            ? rule.fallback_error_message
            : undefined,
      };
    });
    const updated = await adminAPI.settings.updateBetaPolicySettings({
      rules: cleanedRules,
    });
    betaPolicyForm.rules = updated.rules;
    appStore.showSuccess(t("admin.settings.betaPolicy.saved"));
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t("admin.settings.betaPolicy.saveFailed")),
    );
  } finally {
    betaPolicySaving.value = false;
  }
}

// ==================== Provider Management ====================

const allPaymentTypes = computed(() => [
  { value: "easypay", label: t("payment.methods.easypay") },
  { value: "alipay", label: t("payment.methods.alipay") },
  { value: "wxpay", label: t("payment.methods.wxpay") },
  { value: "stripe", label: t("payment.methods.stripe") },
  { value: "airwallex", label: t("payment.methods.airwallex") },
]);

function isPaymentTypeEnabled(type: string): boolean {
  return form.payment_enabled_types.includes(type);
}

const hasAnyPaymentTypeEnabled = computed(
  () => form.payment_enabled_types.length > 0,
);

function togglePaymentType(type: string) {
  if (form.payment_enabled_types.includes(type)) {
    form.payment_enabled_types = form.payment_enabled_types.filter(
      (t) => t !== type,
    );
    // Disable all provider instances matching this type
    disableProvidersByType(type);
  } else {
    form.payment_enabled_types = [...form.payment_enabled_types, type];
  }
}

async function disableProvidersByType(type: string) {
  const matching = providers.value.filter(
    (p) => p.provider_key === type && p.enabled,
  );
  for (const p of matching) {
    try {
      await adminAPI.payment.updateProvider(p.id, { enabled: false });
      p.enabled = false;
    } catch (err: unknown) {
      slog("disable provider failed", p.id, err);
    }
  }
}

function slog(...args: unknown[]) {
  console.warn("[payment]", ...args);
}

const providersLoading = ref(false);
const providerSaving = ref(false);
const providerTesting = ref(false);
const providers = ref<ProviderInstance[]>([]);
const providerUpdatingIds = ref<Set<number>>(new Set());
const showProviderDialog = ref(false);
const showDeleteProviderDialog = ref(false);
const editingProvider = ref<ProviderInstance | null>(null);
const deletingProviderId = ref<number | null>(null);
const providerDialogRef = ref<InstanceType<
  typeof PaymentProviderDialog
> | null>(null);
let providersLoadSeq = 0;

function normalizePaymentProvider(provider: ProviderInstance): ProviderInstance {
  return {
    ...provider,
    // 将接口返回的 null supported_types 转为空数组，供卡片和弹窗调用 includes。
    supported_types: Array.isArray(provider.supported_types)
      ? provider.supported_types
      : [],
  };
}

const updatingProviderIds = computed(() => Array.from(providerUpdatingIds.value));

function isProviderUpdating(providerId: number): boolean {
  return providerUpdatingIds.value.has(providerId);
}

function setProviderUpdating(providerId: number, updating: boolean) {
  // 同一服务商的 PATCH 和列表回写依次执行。
  const next = new Set(providerUpdatingIds.value);
  if (updating) {
    next.add(providerId);
  } else {
    next.delete(providerId);
  }
  providerUpdatingIds.value = next;
}

const providerKeyOptions = computed(() => [
  { value: "easypay", label: t("admin.settings.payment.providerEasypay") },
  { value: "alipay", label: t("admin.settings.payment.providerAlipay") },
  { value: "wxpay", label: t("admin.settings.payment.providerWxpay") },
  { value: "stripe", label: t("admin.settings.payment.providerStripe") },
  { value: "airwallex", label: t("admin.settings.payment.providerAirwallex") },
]);

const enabledProviderKeyOptions = computed(() => {
  const enabled = form.payment_enabled_types;
  return providerKeyOptions.value.filter((opt) => enabled.includes(opt.value));
});

const loadBalanceOptions = computed(() => [
  {
    value: "round-robin",
    label: t("admin.settings.payment.strategyRoundRobin"),
  },
  {
    value: "least-amount",
    label: t("admin.settings.payment.strategyLeastAmount"),
  },
]);

const cancelRateLimitUnitOptions = computed(() => [
  {
    value: "minute",
    label: t("admin.settings.payment.cancelRateLimitUnitMinute"),
  },
  { value: "hour", label: t("admin.settings.payment.cancelRateLimitUnitHour") },
  { value: "day", label: t("admin.settings.payment.cancelRateLimitUnitDay") },
]);

const cancelRateLimitModeOptions = computed(() => [
  {
    value: "rolling",
    label: t("admin.settings.payment.cancelRateLimitWindowModeRolling"),
  },
  {
    value: "fixed",
    label: t("admin.settings.payment.cancelRateLimitWindowModeFixed"),
  },
]);

type ProviderEnablementCandidate = Pick<
  ProviderInstance,
  "id" | "provider_key" | "supported_types" | "enabled" | "name"
>;

function getProviderVisibleMethods(
  provider: ProviderEnablementCandidate,
): Array<"alipay" | "wxpay"> {
  if (!provider.enabled) {
    return [];
  }

  const supportedTypes = Array.isArray(provider.supported_types)
    ? provider.supported_types
    : [];
  const methods = new Set<"alipay" | "wxpay">();
  const addMethod = (type: string) => {
    const method = normalizeVisibleMethod(type);
    if (method === "alipay" || method === "wxpay") {
      methods.add(method);
    }
  };

  if (provider.provider_key === "alipay") {
    if (supportedTypes.length === 0) {
      methods.add("alipay");
    } else {
      supportedTypes.forEach((type) => {
        if (normalizeVisibleMethod(type) === "alipay") {
          methods.add("alipay");
        }
      });
    }
  } else if (provider.provider_key === "wxpay") {
    if (supportedTypes.length === 0) {
      methods.add("wxpay");
    } else {
      supportedTypes.forEach((type) => {
        if (normalizeVisibleMethod(type) === "wxpay") {
          methods.add("wxpay");
        }
      });
    }
  } else if (provider.provider_key === "easypay") {
    supportedTypes.forEach(addMethod);
  }

  return Array.from(methods);
}

function findProviderEnablementConflict(
  candidate: ProviderEnablementCandidate,
): { method: "alipay" | "wxpay"; conflicting: ProviderInstance } | null {
  const claimedMethods = getProviderVisibleMethods(candidate);
  if (claimedMethods.length === 0) {
    return null;
  }

  for (const other of providers.value) {
    if (other.id === candidate.id || !other.enabled) {
      continue;
    }

    const otherMethods = getProviderVisibleMethods(other);
    const matchedMethod = claimedMethods.find((method) =>
      otherMethods.includes(method),
    );
    if (matchedMethod) {
      return {
        method: matchedMethod,
        conflicting: other,
      };
    }
  }

  return null;
}

function showProviderEnablementConflict(
  conflict: { method: "alipay" | "wxpay"; conflicting: ProviderInstance },
) {
  appStore.showError(
    t("admin.settings.payment.enableConflict", {
      method: t(`payment.methods.${conflict.method}`),
      provider: conflict.conflicting.name,
    }),
  );
}

async function loadProviders() {
  const seq = ++providersLoadSeq;
  providersLoading.value = true;
  try {
    const res = await adminAPI.payment.getProviders();
    // 按请求序号接收最后一次加载的服务商列表。
    if (seq === providersLoadSeq) {
      providers.value = (res.data || []).map(normalizePaymentProvider);
    }
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  } finally {
    if (seq === providersLoadSeq) {
      providersLoading.value = false;
    }
  }
}

function openCreateProvider() {
  editingProvider.value = null;
  providerDialogRef.value?.reset(
    enabledProviderKeyOptions.value[0]?.value || "easypay",
  );
  showProviderDialog.value = true;
}

function openEditProvider(provider: ProviderInstance) {
  editingProvider.value = provider;
  providerDialogRef.value?.loadProvider(provider);
  showProviderDialog.value = true;
}

async function handleSaveProvider(payload: Partial<ProviderInstance>) {
  providerSaving.value = true;
  try {
    const candidate: ProviderEnablementCandidate = {
      id: editingProvider.value?.id ?? 0,
      provider_key:
        payload.provider_key ?? editingProvider.value?.provider_key ?? "",
      supported_types:
        payload.supported_types ?? editingProvider.value?.supported_types ?? [],
      enabled: payload.enabled ?? editingProvider.value?.enabled ?? false,
      name: payload.name ?? editingProvider.value?.name ?? "",
    };
    const conflict = findProviderEnablementConflict(candidate);
    if (conflict) {
      showProviderEnablementConflict(conflict);
      return;
    }

    if (editingProvider.value) {
      await adminAPI.payment.updateProvider(editingProvider.value.id, payload);
    } else {
      await adminAPI.payment.createProvider(payload);
    }
    showProviderDialog.value = false;
    // Reload full list (API returns decrypted/formatted data with correct sort order)
    await loadProviders();
    // Auto-save settings so provider changes take effect immediately
    await saveSettings();
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  } finally {
    providerSaving.value = false;
  }
}

async function handleTestProviderDraft(payload: {
  provider_key: string;
  instance_id?: number;
  config: Record<string, string>;
}) {
  providerTesting.value = true;
  try {
    const result = await adminAPI.payment.testProviderDraft(payload);
    if (result.data?.reachable) {
      appStore.showSuccess(t("admin.settings.payment.connectionTestSuccess"));
      return;
    }
    appStore.showError(t("payment.errors.PAYMENT_PROVIDER_TEST_FAILED"));
  } catch (err: unknown) {
    appStore.showError(
      extractI18nErrorMessage(err, t, "payment.errors", t("common.error")),
    );
  } finally {
    providerTesting.value = false;
  }
}

async function handleToggleField(
  provider: ProviderInstance,
  field: "enabled" | "refund_enabled" | "allow_user_refund",
) {
  if (isProviderUpdating(provider.id)) {
    return;
  }
  setProviderUpdating(provider.id, true);
  try {
    const currentProvider =
      providers.value.find((item) => item.id === provider.id) ?? provider;
    let newValue: boolean;
    if (field === "enabled") newValue = !currentProvider.enabled;
    else if (field === "refund_enabled") newValue = !currentProvider.refund_enabled;
    else newValue = !currentProvider.allow_user_refund;

    if (field === "enabled" && newValue) {
      const conflict = findProviderEnablementConflict({
        id: currentProvider.id,
        provider_key: currentProvider.provider_key,
        supported_types: currentProvider.supported_types,
        enabled: true,
        name: currentProvider.name,
      });
      if (conflict) {
        showProviderEnablementConflict(conflict);
        return;
      }
    }

    const payload: Record<string, boolean> = { [field]: newValue };
    // 关闭退款能力时同步关闭用户自助退款，保持前后端状态一致。
    if (field === "refund_enabled" && !newValue) {
      payload.allow_user_refund = false;
    }
    await adminAPI.payment.updateProvider(currentProvider.id, payload);
    await loadProviders();
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  } finally {
    setProviderUpdating(provider.id, false);
  }
}

async function handleToggleType(provider: ProviderInstance, type: string) {
  if (isProviderUpdating(provider.id)) {
    return;
  }
  setProviderUpdating(provider.id, true);
  try {
    const currentProvider =
      providers.value.find((item) => item.id === provider.id) ?? provider;
    if (currentProvider.provider_key === "stripe") {
      // Stripe Checkout 的支付方式由 Stripe Dashboard 控制。
      return;
    }
    const supportedTypes = Array.isArray(currentProvider.supported_types)
      ? currentProvider.supported_types
      : [];
    const updated = supportedTypes.includes(type)
      ? supportedTypes.filter((t) => t !== type)
      : [...supportedTypes, type];
    const conflict = findProviderEnablementConflict({
      id: currentProvider.id,
      provider_key: currentProvider.provider_key,
      supported_types: updated,
      enabled: currentProvider.enabled,
      name: currentProvider.name,
    });
    if (conflict) {
      showProviderEnablementConflict(conflict);
      return;
    }
    await adminAPI.payment.updateProvider(currentProvider.id, {
      supported_types: updated,
    } as Partial<ProviderInstance>);
    await loadProviders();
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  } finally {
    setProviderUpdating(provider.id, false);
  }
}

function confirmDeleteProvider(provider: ProviderInstance) {
  deletingProviderId.value = provider.id;
  showDeleteProviderDialog.value = true;
}

async function handleReorderProviders(
  updates: { id: number; sort_order: number }[],
) {
  try {
    await Promise.all(
      updates.map((u) =>
        adminAPI.payment.updateProvider(u.id, {
          sort_order: u.sort_order,
        } as Partial<ProviderInstance>),
      ),
    );
    await loadProviders();
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
    loadProviders();
  }
}

async function handleDeleteProvider() {
  if (!deletingProviderId.value) return;
  try {
    await adminAPI.payment.deleteProvider(deletingProviderId.value);
    appStore.showSuccess(t("common.deleted"));
    showDeleteProviderDialog.value = false;
    loadProviders();
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, "payment.errors", t("common.error")));
  }
}

onMounted(() => {
  loadSettings();
  loadHomeFeaturedModelOptions();
  loadSubscriptionPlans();
  loadAdminApiKey();
  loadOllamaCloudUsageSettings();
  loadOverloadCooldownSettings();
  loadOpenAI403CooldownSettings();
  loadRateLimit429CooldownSettings();
  loadPanelRateLimitSettings();
  loadStreamTimeoutSettings();
  loadRectifierSettings();
  loadBetaPolicySettings();
  loadProviders();
});

// bypass_registration 与三个身份同步开关在 internal_only 模式下生效。
// policy 切到其它值时，将这四个字段重置为 false，保存请求随之更新。
// 后端 admin handler 和配置加载层也会规范化这些值。
watch(
  () => form.dingtalk_connect_corp_restriction_policy,
  (policy) => {
    if (policy !== "internal_only") {
      if (form.dingtalk_connect_bypass_registration) form.dingtalk_connect_bypass_registration = false;
      if (form.dingtalk_connect_sync_corp_email) form.dingtalk_connect_sync_corp_email = false;
      if (form.dingtalk_connect_sync_display_name) form.dingtalk_connect_sync_display_name = false;
      if (form.dingtalk_connect_sync_dept) form.dingtalk_connect_sync_dept = false;
    }
  },
);
</script>

<style scoped>
/* ============ 系统设置 Tab 导航 ============ */
.settings-tabs-shell {
  @apply sticky z-20 -mx-1 rounded-full border border-gray-200 bg-white/90 p-1.5 backdrop-blur-xl dark:border-dark-600/70 dark:bg-dark-900/90;
  /* 顶栏高度 + 1.25rem 间距,合成原 4.75rem;顶栏调高时吸顶位置自动跟随。 */
  top: calc(var(--header-h) + 1.25rem);
  box-shadow: 0 1px 0 rgb(255 255 255 / 0.9) inset;
}

/* 网关页有两行标签，外壳使用弹窗圆角。 */
.settings-tabs-shell-stacked {
  @apply rounded-dialog;
}

.settings-tabs-scroll {
  @apply overflow-x-auto rounded-full;
  -ms-overflow-style: none;
  scrollbar-width: none;
  scroll-padding-inline: 0.5rem;
}

.settings-tabs-scroll::-webkit-scrollbar {
  display: none;
}

.settings-tabs {
  @apply flex min-w-max items-center gap-1;
}

.settings-tab {
  @apply relative isolate flex h-9 min-w-[6.75rem] shrink-0 items-center justify-center gap-1.5 whitespace-nowrap rounded-full border border-transparent px-3 text-sm font-medium text-gray-600 outline-none transition-colors duration-fast ease-standard dark:text-gray-300;
}

@media (min-width: 768px) {
  .settings-tabs {
    @apply min-w-full;
  }

  .settings-tab {
    @apply min-w-0 flex-1 basis-0 overflow-hidden px-2 text-[13px];
  }

  .settings-tab-icon {
    @apply h-6 w-6;
  }
}

.settings-tab::before {
  @apply absolute inset-0 -z-10 rounded-full opacity-0 transition-opacity duration-normal;
  content: "";
  background: linear-gradient(135deg, rgb(248 250 252 / 0.95), rgb(241 245 249 / 0.8));
}

.settings-tab:hover::before,
.settings-tab:focus-visible::before {
  opacity: 1;
}

.settings-tab:focus-visible {
  @apply ring-2 ring-primary-500/40 ring-offset-2 ring-offset-white dark:ring-offset-dark-900;
}

.settings-tab-active {
  @apply border-primary-200/80 bg-white text-primary-700 shadow-sm dark:border-primary-500/15 dark:bg-primary-500/8 dark:text-primary-500;
  box-shadow: 0 1px 0 rgb(255 255 255 / 0.92) inset;
}

.settings-tab-active::before {
  opacity: 0;
}

.settings-tab-icon {
  @apply flex h-7 w-7 shrink-0 items-center justify-center text-gray-500 transition-colors duration-normal dark:text-gray-400;
}

.settings-tab:hover .settings-tab-icon,
.settings-tab:focus-visible .settings-tab-icon {
  @apply text-gray-700 dark:text-gray-200;
}

.settings-tab-active .settings-tab-icon {
  @apply text-primary-600 dark:text-primary-500;
}

.settings-tab-label {
  @apply min-w-0 overflow-hidden text-ellipsis whitespace-nowrap leading-none;
}

/* 网关二级标签与主标签放在同一吸顶容器中。 */
.gateway-section-tabs-scroll {
  @apply mt-1.5 overflow-x-auto border-t border-gray-100 pt-1.5 dark:border-dark-700;
  -ms-overflow-style: none;
  scrollbar-width: none;
  scroll-padding-inline: 0.5rem;
}

.gateway-section-tabs-scroll::-webkit-scrollbar {
  display: none;
}

.gateway-section-tabs {
  @apply flex min-w-max items-center gap-1;
}

.gateway-section-tab {
  @apply flex h-9 min-w-[7.75rem] shrink-0 items-center justify-center gap-1.5 whitespace-nowrap rounded-full border border-transparent px-3 text-sm font-medium text-gray-600 outline-none transition-colors duration-normal dark:text-gray-300;
}

.gateway-section-tab:hover,
.gateway-section-tab:focus-visible {
  @apply bg-gray-50 text-gray-900 dark:bg-dark-700 dark:text-white;
}

.gateway-section-tab:focus-visible {
  @apply ring-2 ring-primary-500/40 ring-offset-1 ring-offset-white dark:ring-offset-dark-900;
}

.gateway-section-tab-active {
  @apply border-primary-200 bg-primary-50 text-primary-700 dark:border-primary-400/30 dark:bg-primary-400/10 dark:text-primary-200;
}

.gateway-section-tab-icon {
  @apply flex h-5 w-5 shrink-0 items-center justify-center;
}

.gateway-section-tab-label {
  @apply min-w-0 whitespace-nowrap leading-none;
}

.gateway-settings-content {
  /* 锚点跳转避开顶栏 + 吸顶 tabs 块:9.25rem = tabs 偏移 1.25rem + tabs 高度与下方留白(经验值),
     合成原 12.75rem,数值不变。 */
  scroll-margin-top: calc(var(--header-h) + 9.25rem);
}

@media (min-width: 768px) {
  .gateway-section-tabs {
    @apply min-w-full;
  }

  .gateway-section-tab {
    @apply min-w-0 flex-1 basis-0;
  }
}
</style>

<style>
/* 暗色 Tab 样式放在非 scoped 块中，供生产构建保留。 */
.dark .settings-tabs-shell {
  border-color: theme('borderColor.dark.600');
  background: rgb(15 15 16 / 0.86);
  box-shadow: 0 1px 0 rgb(255 255 255 / 0.06) inset;
}

.dark .settings-tab::before {
  background: linear-gradient(135deg, rgb(39 39 42 / 0.9), rgb(63 63 70 / 0.62));
}

.dark .settings-tab-active {
  box-shadow: 0 1px 0 rgb(255 255 255 / 0.08) inset;
}
</style>
