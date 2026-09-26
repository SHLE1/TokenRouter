<template>
  <GroupFormSection
    :title="t('admin.protocols.groupTitle')"
    :hint="t('admin.protocols.groupHint')"
  >
    <div
      v-if="protocolCatalogError"
      class="flex items-center gap-2 text-sm text-red-500"
      role="alert"
    >
      <span>{{ t('admin.protocols.loadError') }}</span>
      <button
        type="button"
        class="btn btn-secondary"
        data-testid="protocol-catalog-retry"
        :disabled="protocolCatalogLoading"
        @click="retryCatalog"
      >
        {{ t('common.retry') }}
      </button>
    </div>
    <p v-else-if="!protocolCatalog" class="input-hint">
      {{ t('common.loading') }}
    </p>
    <div
      data-testid="client-protocol-list"
      class="divide-y divide-gray-100 dark:divide-dark-700"
    >
      <div
        v-for="protocol in protocols"
        :key="protocol.id"
        class="space-y-4 py-4"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <span
              class="block text-sm font-medium text-primary-900 dark:text-dark-50"
              >{{ protocol.name }}</span
            ><code
              class="block break-all text-xs text-gray-500"
              :data-protocol-endpoint="protocol.id"
              >{{ protocol.endpoint }}</code
            >
          </div>
          <Toggle
            size="md"
            class="shrink-0"
            :model-value="modelValue.includes(protocol.id)"
            :data-protocol="protocol.id"
            :aria-label="protocol.name"
            @update:model-value="toggle(protocol.id)"
          />
        </div>
        <div v-if="profile?.fallback_targets[protocol.id]?.length">
          <label
            :for="`${idPrefix}-${protocol.id}-fallback`"
            class="input-label"
            >{{ t('admin.protocols.fallback') }}</label
          >
          <Select
            :id="`${idPrefix}-${protocol.id}-fallback`"
            :aria-label="`${protocol.name}: ${t('admin.protocols.fallback')}`"
            :model-value="fallbackMode(protocol.id)"
            :options="modeOptions"
            @update:model-value="setMode(protocol.id, String($event))"
          />
          <div
            v-if="fallbackMode(protocol.id) === 'restricted'"
            class="mt-2 space-y-2"
          >
            <div
              v-for="(target, index) in fallbacks?.[protocol.id]"
              :key="index"
              class="flex items-center gap-2"
            >
              <Select
                class="min-w-0 flex-1"
                :aria-label="`${protocol.name}: ${t('admin.protocols.fallback')} ${index + 1}`"
                :model-value="target"
                :options="targetOptions(protocol.id)"
                @update:model-value="
                  setTarget(protocol.id, index, String($event) as ProtocolID)
                "
              />
              <button
                type="button"
                class="btn btn-ghost btn-icon shrink-0 text-red-500"
                :aria-label="t('common.delete')"
                @click="removeTarget(protocol.id, index)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="!remainingTarget(protocol.id)"
              @click="addTarget(protocol.id)"
            >
              {{ t('common.add') }}
            </button>
          </div>
        </div>
      </div>
    </div>
    <div class="space-y-2 border-t border-gray-200 pt-6 dark:border-dark-600">
      <label :for="`${idPrefix}-image-policy`" class="input-label">{{
        t('admin.protocols.imagePolicy')
      }}</label>
      <p :id="`${idPrefix}-image-policy-scope`" class="input-hint">
        {{ t('admin.protocols.imagePolicyHint') }}
      </p>
      <Select
        :aria-label="t('admin.protocols.imagePolicy')"
        :aria-describedby="`${idPrefix}-image-policy-scope ${idPrefix}-image-policy-description`"
        :id="`${idPrefix}-image-policy`"
        :model-value="imagePolicy ?? 'inherit'"
        :options="imageOptions"
        @update:model-value="
          emit('update:imagePolicy', String($event) as CodexImageToolMode)
        "
      />
      <p :id="`${idPrefix}-image-policy-description`" class="input-hint">
        {{ t(`admin.protocols.imagePolicyOptions.${imagePolicy ?? 'inherit'}.description`) }}
      </p>
    </div>
  </GroupFormSection>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Select from '@/components/common/Select.vue'
import GroupFormSection from './GroupFormSection.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  loadProtocolCatalog,
  protocolCatalog,
  protocolCatalogError,
  protocolCatalogLoading,
} from '@/api/admin/protocolCapabilities'
import type { ProtocolID } from '@/types'
import type { CodexImageToolMode } from '@/utils/codexImageToolMode'
import { setGroupClientProtocol } from '@/utils/groupClientProtocols'
const props = withDefaults(
  defineProps<{
    idPrefix?: string
    modelValue: ProtocolID[]
    fallbacks?: Partial<Record<ProtocolID, ProtocolID[]>>
    imagePolicy?: CodexImageToolMode
  }>(),
  { idPrefix: 'group-protocol' },
)
const emit = defineEmits<{
  'update:modelValue': [value: ProtocolID[]]
  'update:fallbacks': [value: Partial<Record<ProtocolID, ProtocolID[]>>]
  'update:imagePolicy': [value: CodexImageToolMode]
}>()
const { t } = useI18n()
// 分组单独说明继承来源和请求行为，选项值沿用现有四态契约。
const imageModes = ['inherit', 'enabled', 'disabled', 'block'] as const
const imageOptions = computed(() =>
  imageModes.map((value) => ({
    value,
    label: t(`admin.protocols.imagePolicyOptions.${value}.label`),
  })),
)
// 所有表单共享加载状态；任一入口重试成功后同时恢复。
function retryCatalog() {
  void loadProtocolCatalog().catch(() => {})
}
retryCatalog()
const profile = computed(() => protocolCatalog.value?.groups[0])
const protocols = computed(
  () =>
    protocolCatalog.value?.protocols.filter((protocol) =>
      profile.value?.protocols.includes(protocol.id),
    ) ?? [],
)
function targetOptions(source: ProtocolID) {
  return (profile.value?.fallback_targets[source] ?? []).map((id) => ({
    value: id,
    label:
      protocolCatalog.value?.protocols.find((protocol) => protocol.id === id)
        ?.name ?? id,
  }))
}
function toggle(id: ProtocolID) {
  emit(
    'update:modelValue',
    setGroupClientProtocol(
      props.modelValue,
      id,
      !props.modelValue.includes(id),
    ),
  )
}
// 缺少入口采用自动转换；空数组仅允许原生，显式列表按顺序尝试。
const modeOptions = computed(() => [
  { value: 'auto', label: t('admin.protocols.auto') },
  { value: 'native', label: t('admin.protocols.nativeOnly') },
  { value: 'restricted', label: t('admin.protocols.restricted') },
])
function fallbackMode(source: ProtocolID) {
  const targets = props.fallbacks?.[source]
  return targets === undefined
    ? 'auto'
    : targets.length
      ? 'restricted'
      : 'native'
}
function setMode(source: ProtocolID, mode: string) {
  const next = { ...props.fallbacks }
  if (mode === 'auto') delete next[source]
  else
    next[source] =
      mode === 'restricted'
        ? (profile.value?.fallback_targets[source] ?? []).slice(0, 1)
        : []
  emit('update:fallbacks', next)
}
function remainingTarget(source: ProtocolID) {
  return profile.value?.fallback_targets[source]?.find(
    (target) => !props.fallbacks?.[source]?.includes(target),
  )
}
function addTarget(source: ProtocolID) {
  const target = remainingTarget(source)
  if (target)
    emit('update:fallbacks', {
      ...props.fallbacks,
      [source]: [...(props.fallbacks?.[source] ?? []), target],
    })
}
function setTarget(source: ProtocolID, index: number, target: ProtocolID) {
  const targets = [...(props.fallbacks?.[source] ?? [])]
  const existing = targets.indexOf(target)
  if (existing >= 0 && existing !== index) targets[existing] = targets[index]
  targets[index] = target
  emit('update:fallbacks', { ...props.fallbacks, [source]: targets })
}
function removeTarget(source: ProtocolID, index: number) {
  emit('update:fallbacks', {
    ...props.fallbacks,
    [source]: props.fallbacks?.[source]?.filter((_, i) => i !== index) ?? [],
  })
}
</script>
