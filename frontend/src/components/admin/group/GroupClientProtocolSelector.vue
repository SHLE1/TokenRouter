<template>
  <section class="space-y-3 border-t border-gray-200 pt-4 dark:border-dark-400">
    <h4 class="text-sm font-medium">{{ t('admin.protocols.groupTitle') }}</h4>
    <p class="input-hint">{{ t('admin.protocols.groupHint') }}</p>
    <div v-if="protocolCatalogError" class="flex items-center gap-2 text-sm text-red-500" role="alert">
      <span>{{ t('admin.protocols.loadError') }}</span>
      <button type="button" class="underline" data-testid="protocol-catalog-retry" :disabled="protocolCatalogLoading" @click="retryCatalog">{{ t('common.retry') }}</button>
    </div>
    <p v-else-if="!protocolCatalog" class="input-hint">{{ t('common.loading') }}</p>
    <div data-testid="client-protocol-list" class="divide-y divide-gray-100 dark:divide-dark-700">
      <div v-for="protocol in protocols" :key="protocol.id" class="grid items-center gap-3 py-3 sm:grid-cols-2">
        <div class="flex items-center justify-between gap-3">
          <div><span class="block text-sm font-medium">{{ protocol.name }}</span><code class="block break-all text-xs text-gray-500" :data-protocol-endpoint="protocol.id">{{ protocol.endpoint }}</code></div>
          <Toggle :model-value="modelValue.includes(protocol.id)" :data-protocol="protocol.id" :aria-label="protocol.name" @update:model-value="toggle(protocol.id)" />
        </div>
        <div v-if="profile?.fallback_targets[protocol.id]?.length">
          <label class="input-label text-xs">{{ t('admin.protocols.fallback') }}</label>
          <Select :model-value="fallbackMode(protocol.id)" :options="modeOptions" @update:model-value="setMode(protocol.id, String($event))" />
          <div v-if="fallbackMode(protocol.id) === 'restricted'" class="mt-2 space-y-2">
            <div v-for="(target, index) in fallbacks?.[protocol.id]" :key="index" class="flex items-center gap-2">
              <Select :model-value="target" :options="targetOptions(protocol.id)" @update:model-value="setTarget(protocol.id, index, String($event) as ProtocolID)" />
              <button type="button" class="btn btn-secondary btn-sm" :aria-label="t('common.delete')" @click="removeTarget(protocol.id, index)">{{ t('common.delete') }}</button>
            </div>
            <button type="button" class="btn btn-secondary btn-sm" :disabled="!remainingTarget(protocol.id)" @click="addTarget(protocol.id)">{{ t('common.add') }}</button>
          </div>
        </div>
      </div>
    </div>
    <div class="space-y-2">
      <label class="input-label">{{ t('admin.protocols.imagePolicy') }}</label>
      <CodexImageToolModeSelector :model-value="imagePolicy ?? 'inherit'" @update:model-value="emit('update:imagePolicy', $event)" />
    </div>
  </section>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Select from '@/components/common/Select.vue'
import CodexImageToolModeSelector from '@/components/account/CodexImageToolModeSelector.vue'
import { loadProtocolCatalog, protocolCatalog, protocolCatalogError, protocolCatalogLoading } from '@/api/admin/protocolCapabilities'
import type { ProtocolID } from '@/types'
import type { CodexImageToolMode } from '@/utils/codexImageToolMode'
import { setGroupClientProtocol } from '@/utils/groupClientProtocols'
const props = defineProps<{ modelValue: ProtocolID[]; fallbacks?: Partial<Record<ProtocolID, ProtocolID[]>>; imagePolicy?: CodexImageToolMode }>()
const emit = defineEmits<{
  'update:modelValue': [value: ProtocolID[]]
  'update:fallbacks': [value: Partial<Record<ProtocolID, ProtocolID[]>>]
  'update:imagePolicy': [value: CodexImageToolMode]
}>()
const { t } = useI18n()
// 所有表单共享加载状态；任一入口重试成功后同时恢复。
function retryCatalog() { void loadProtocolCatalog().catch(() => {}) }
retryCatalog()
const profile = computed(() => protocolCatalog.value?.groups[0])
const protocols = computed(() => protocolCatalog.value?.protocols.filter(protocol => profile.value?.protocols.includes(protocol.id)) ?? [])
function targetOptions(source: ProtocolID) {
  return (profile.value?.fallback_targets[source] ?? []).map(id => ({ value: id, label: protocolCatalog.value?.protocols.find(protocol => protocol.id === id)?.name ?? id }))
}
function toggle(id: ProtocolID) { emit('update:modelValue', setGroupClientProtocol(props.modelValue, id, !props.modelValue.includes(id))) }
// 缺少入口采用自动转换；空数组仅允许原生，显式列表按顺序尝试。
const modeOptions = computed(() => [
  { value: 'auto', label: t('admin.protocols.auto') },
  { value: 'native', label: t('admin.protocols.nativeOnly') },
  { value: 'restricted', label: t('admin.protocols.restricted') },
])
function fallbackMode(source: ProtocolID) {
  const targets = props.fallbacks?.[source]
  return targets === undefined ? 'auto' : targets.length ? 'restricted' : 'native'
}
function setMode(source: ProtocolID, mode: string) {
  const next = { ...props.fallbacks }
  if (mode === 'auto') delete next[source]
  else next[source] = mode === 'restricted' ? (profile.value?.fallback_targets[source] ?? []).slice(0, 1) : []
  emit('update:fallbacks', next)
}
function remainingTarget(source: ProtocolID) {
  return profile.value?.fallback_targets[source]?.find(target => !props.fallbacks?.[source]?.includes(target))
}
function addTarget(source: ProtocolID) {
  const target = remainingTarget(source)
  if (target) emit('update:fallbacks', { ...props.fallbacks, [source]: [...(props.fallbacks?.[source] ?? []), target] })
}
function setTarget(source: ProtocolID, index: number, target: ProtocolID) {
  const targets = [...(props.fallbacks?.[source] ?? [])]
  const existing = targets.indexOf(target)
  if (existing >= 0 && existing !== index) targets[existing] = targets[index]
  targets[index] = target
  emit('update:fallbacks', { ...props.fallbacks, [source]: targets })
}
function removeTarget(source: ProtocolID, index: number) {
  emit('update:fallbacks', { ...props.fallbacks, [source]: props.fallbacks?.[source]?.filter((_, i) => i !== index) ?? [] })
}
</script>
