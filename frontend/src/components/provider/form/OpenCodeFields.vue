<template>
  <div class="space-y-4" data-testid="opencode-fields">
    <label class="block">
      <span class="input-label">{{ t('admin.providers.opencode.mode') }}</span>
      <Select v-model="draft.mode" :options="modeOptions" :disabled="syncing" />
    </label>
    <p class="input-hint">{{ t('admin.providers.opencode.description') }}</p>
    <button
      type="button"
      class="btn btn-secondary"
      :disabled="syncing || (!providerId && !apiKey?.trim())"
      @click="syncModels"
    >
      {{
        t(
          syncing
            ? 'admin.providers.syncUpstreamModelsLoading'
            : 'admin.providers.syncUpstreamModels',
        )
      }}
    </button>
    <p v-if="error" role="alert" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
    <details class="space-y-3">
      <summary class="cursor-pointer text-sm font-medium">{{
        t('admin.providers.opencode.modelProtocols')
      }}</summary>
      <p class="input-hint">{{ t('admin.providers.opencode.modelProtocolsHint') }}</p>
      <div
        v-for="(target, model) in draft.protocols[draft.mode]"
        :key="model"
        class="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-center gap-2"
      >
        <span class="break-all text-sm">{{ model }}</span>
        <Select
          :model-value="target"
          :options="protocolOptions"
          @update:model-value="setProtocol(model, String($event))"
        />
        <button
          type="button"
          class="btn btn-ghost btn-sm"
          :aria-label="t('common.delete')"
          @click="removeProtocol(model)"
          ><Icon name="trash" size="sm"
        /></button>
      </div>
      <div class="grid gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
        <input
          v-model="newModel"
          class="input"
          :placeholder="t('admin.providers.fromModel')"
          :aria-label="t('admin.providers.fromModel')"
        />
        <Select
          v-model="newProtocol"
          :options="protocolOptions"
          :placeholder="t('admin.providers.opencode.selectProtocol')"
        />
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="!newModel.trim() || !newProtocol"
          @click="addProtocol"
          >{{ t('common.add') }}</button
        >
      </div>
    </details>
    <details class="space-y-3">
      <summary class="cursor-pointer text-sm font-medium">{{
        t('admin.providers.opencode.endpointOverrides')
      }}</summary>
      <label v-for="item in protocolOptions" :key="item.value" class="block">
        <span class="input-label">{{ item.label }}</span>
        <input
          v-model="draft.baseUrls[item.slot]"
          class="input"
          :placeholder="endpointPlaceholder(item.slot)"
        />
      </label>
    </details>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import * as providersAPI from '@/api/admin/providers'
import type { ProtocolID } from '@/api/protocols.types'
import type { OpenCodeDraft } from './openCodeDraft'

// OpenCodeFields.vue edits inference modes, exact model protocols, and endpoint
// overrides using the provider form's draft. Synchronization loads current
// upstream model IDs and their protocol declarations for the selected mode;
// the parent receives the IDs for its whitelist while the draft retains routes.
//
// A synchronization captures mode and root before awaiting the server. It
// ignores results after unmount or identity changes, preventing an earlier
// provider's results from replacing the active draft. Mode controls are disabled
// during that request, and every successful merge keeps explicit administrator
// entries ahead of suggested catalog routes. Persistence belongs to the parent
// dialog, so synchronizing or canceling this component does not save a provider.

// props supplies saved identity or temporary credentials for model synchronization.
const props = defineProps<{
  providerId?: number // Existing provider identity.
  apiKey?: string // Unsaved inference credential.
  baseUrl?: string // Unsaved inference service root.
}>()
// The parent owns persistence, while this component edits its isolated configuration draft.
const draft = defineModel<OpenCodeDraft>({ required: true })
// emit publishes synchronized model IDs for the parent whitelist draft.
const emit = defineEmits<{ synced: [models: string[]] }>()
// These local fields retain request feedback and a pending manual override for this mounted form.
const { t } = useI18n()
const syncing = ref(false)
const error = ref('')
const newModel = ref('')
const newProtocol = ref('')
let disposed = false

// modeOptions presents the two independently billed OpenCode inference services.
const modeOptions = computed(() => [
  { value: 'payg', label: t('admin.providers.opencode.payg') },
  { value: 'go', label: 'OpenCode Go' },
])
// Protocol choices pair native wire identities with their persisted address slots.
const protocolOptions = [
  { value: 'openai_responses', label: 'Responses', slot: 'responses' },
  { value: 'anthropic_messages', label: 'Messages', slot: 'anthropic' },
  { value: 'openai_chat_completions', label: 'Chat Completions', slot: 'chat_completions' },
  { value: 'gemini_generate_content', label: 'Gemini GenerateContent', slot: 'gemini' },
  { value: 'systemone', label: 'SystemOne', slot: 'systemone' },
]

// endpointPlaceholder displays the selected mode's complete protocol base. A
// custom service root replaces the official root, while each address slot
// retains the version segment expected by its protocol's request builder.
function endpointPlaceholder(slot: string): string {
  const root =
    props.baseUrl?.trim().replace(/\/$/, '') ||
    `https://opencode.ai/inference${draft.value.mode === 'go' ? '/go' : ''}`
  const suffix =
    slot === 'anthropic'
      ? '/anthropic/v1'
      : slot === 'gemini'
        ? '/google/v1beta'
        : slot === 'systemone'
          ? '/systemone/v1'
          : '/openai/v1'
  return root + suffix
}

// setProtocol assigns an administrator-selected encoding to one exact model identity.
function setProtocol(model: string, target: string): void {
  draft.value.protocols[draft.value.mode][model] = target as ProtocolID
}

// removeProtocol removes one explicit override, allowing the backend catalog default.
function removeProtocol(model: string): void {
  delete draft.value.protocols[draft.value.mode][model]
}

// addProtocol stores the selected encoding and clears the new-model input fields.
function addProtocol(): void {
  if (!newModel.value.trim() || !newProtocol.value) return
  setProtocol(newModel.value.trim(), newProtocol.value)
  newModel.value = ''
  newProtocol.value = ''
}

// syncModels loads IDs and encodings for the captured provider and mode.
// Stale results are discarded after unmount or input changes. Errors remain
// visible beside the retry button, and the parent saves successful suggestions.
async function syncModels(): Promise<void> {
  if (syncing.value) return
  const mode = draft.value.mode
  const root = props.baseUrl
  const identity = props.providerId
  const key = props.apiKey
  syncing.value = true
  error.value = ''
  try {
    const result = identity
      ? await providersAPI.syncUpstreamModels(identity, { provider_mode: mode, base_url: root })
      : await providersAPI.syncUpstreamModelsPreview({
          platform: 'opencode',
          type: 'apikey',
          api_key: key ?? '',
          base_url: root,
          provider_mode: mode,
        })
    // A changed draft identifies a different synchronization target. Discard the
    // earlier response before publishing either whitelist entries or protocol overrides.
    if (
      disposed ||
      identity !== props.providerId ||
      mode !== draft.value.mode ||
      root !== props.baseUrl ||
      key !== props.apiKey
    )
      return
    draft.value.protocols[mode] = { ...result.model_protocols, ...draft.value.protocols[mode] }
    emit('synced', result.models)
  } catch (cause) {
    if (
      !disposed &&
      identity === props.providerId &&
      mode === draft.value.mode &&
      root === props.baseUrl &&
      key === props.apiKey
    )
      error.value =
        cause instanceof Error ? cause.message : t('admin.providers.syncUpstreamModelsFailed')
  } finally {
    if (!disposed) syncing.value = false
  }
}

// Invalidate in-flight synchronization before an unmounted form can publish fetched settings.
onBeforeUnmount(() => {
  disposed = true
})
</script>
