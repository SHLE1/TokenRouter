<template>
  <SettingsCard :title="t('admin.settings.responsesWS.title')" :description="t('admin.settings.responsesWS.description')" data-testid="responses-ws-settings">
    <template #actions>
      <button type="button" class="btn btn-secondary" :disabled="!effective || saving" @click="restore">{{ t('admin.settings.responsesWS.restore') }}</button>
    </template>
    <SettingsNotice v-if="error" tone="error">{{ error }} <button type="button" class="btn btn-secondary" @click="load">{{ t('admin.settings.responsesWS.retry') }}</button></SettingsNotice>
    <SettingsNotice tone="info">{{ t('admin.settings.responsesWS.applies') }}</SettingsNotice>
    <p v-if="resetAll" class="input-hint">{{ t('admin.settings.responsesWS.restoring') }}</p>
    <template v-if="effective">
      <SettingsSection v-for="advanced in [false, true]" :key="String(advanced)">
        <button v-if="advanced" type="button" class="btn btn-secondary" :aria-expanded="expanded" @click="expanded = !expanded">{{ t('admin.settings.responsesWS.advanced') }}</button>
        <div v-show="!advanced || expanded" class="space-y-4">
          <template v-for="field in responsesWsFields.filter(item => item.advanced === advanced)" :key="field.key">
            <SettingToggleRow v-if="field.boolean" :id="`responses-ws-${field.key}`" :model-value="booleanValue(field.key)" :label="t(`admin.settings.responsesWS.fields.${field.key}`)" :disabled="saving" @update:model-value="setValue(field.key, $event)" />
            <SettingRow v-else :id="`responses-ws-${field.key}`" :label-for="`responses-ws-input-${field.key}`" :label="t(`admin.settings.responsesWS.fields.${field.key}`)" :hint="t('admin.settings.responsesWS.inherit', { value: Number(effective[field.key]) / field.scale }) + (field.zero ? ' ' + t('admin.settings.responsesWS.zero') : '')" field>
              <input :id="`responses-ws-input-${field.key}`" class="input" type="number" :min="field.min" :step="field.step" :disabled="saving" :value="numberValue(field.key, field.scale)" :placeholder="String(Number(effective[field.key]) / field.scale)" @input="setNumber(field.key, field.scale, $event)" />
            </SettingRow>
          </template>
        </div>
      </SettingsSection>
    </template>
  </SettingsCard>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getSettings, updateSettings, type ResponsesWSParameters, type ResponsesWSOverrides } from '@/api/admin/settings'
import { useSettingsSaveTarget } from '@/composables/useSettingsSaveRegistry'
import { useDirtyTracker } from '@/composables/useDirtyTracker'
import SettingsCard from '@/components/common/settings/SettingsCard.vue'
import SettingsSection from '@/components/common/settings/SettingsSection.vue'
import SettingsNotice from '@/components/common/settings/SettingsNotice.vue'
import SettingRow from '@/components/common/settings/SettingRow.vue'
import SettingToggleRow from '@/components/common/settings/SettingToggleRow.vue'
import { responsesWsFields } from './responsesWsFields'

const { t } = useI18n()
const effective = ref<ResponsesWSParameters>()
const saved = ref<ResponsesWSOverrides>({})
const patch = ref<ResponsesWSOverrides>({})
const resetAll = ref(false)
const expanded = ref(false)
const saving = ref(false)
const error = ref('')
const proposedOverrides = computed(() => {
  if (resetAll.value) return {}
  const next = { ...saved.value, ...patch.value }
  return Object.fromEntries(Object.entries(next).filter(([, value]) => value != null).sort(([left], [right]) => left.localeCompare(right)))
})
const { dirty, markClean } = useDirtyTracker({ responsesWS: () => proposedOverrides.value })

// 仅发送用户修改的字段，清空输入表示移除该项覆盖。
function setValue(key: keyof ResponsesWSParameters, value: number | boolean | null) {
  if (resetAll.value) {
    patch.value = Object.fromEntries(responsesWsFields.map(field => [field.key, null]))
    resetAll.value = false
  }
  patch.value = { ...patch.value, [key]: value }
}
function setNumber(key: keyof ResponsesWSParameters, scale: number, event: Event) {
  const text = (event.target as HTMLInputElement).value
  setValue(key, text === '' ? null : Number(text) * scale)
}
function numberValue(key: keyof ResponsesWSParameters, scale: number) {
  if (resetAll.value) return ''
  const value = Object.prototype.hasOwnProperty.call(patch.value, key) ? patch.value[key] : saved.value[key]
  return value == null ? '' : Number(value) / scale
}
function booleanValue(key: keyof ResponsesWSParameters) {
  return Boolean(patch.value[key] ?? saved.value[key] ?? effective.value?.[key])
}
function restore() { resetAll.value = true; patch.value = {} }
async function load() {
  try {
    const value = await getSettings()
    effective.value = value.responses_ws_effective
    saved.value = value.responses_ws ?? {}
    patch.value = {}
    resetAll.value = false
    error.value = ''
    markClean('responsesWS')
  } catch { error.value = t('admin.settings.responsesWS.loadError') }
}
async function save() {
  saving.value = true
  error.value = ''
  try {
    await updateSettings({ responses_ws: resetAll.value ? null : patch.value })
    await load()
    return !error.value
  } catch (cause) {
    const data = (cause as { response?: { data?: { reason?: string; code?: string; message?: string } } }).response?.data
    error.value = data?.reason === 'SETTINGS_APPLY_FAILED' || data?.code === 'SETTINGS_APPLY_FAILED'
      ? t('admin.settings.responsesWS.applyError')
      : data?.message ?? t('admin.settings.responsesWS.saveError')
    return false
  } finally { saving.value = false }
}
useSettingsSaveTarget('responses-ws', { dirty, save, discard: load })
onMounted(load)
</script>
