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
      :disabled="syncing || syncNeedsSave || (!providerId && !apiKey?.trim())"
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
    <p v-if="providerId" class="input-hint" data-testid="opencode-sync-settings">
      {{
        t(
          syncNeedsSave
            ? 'admin.providers.opencode.saveBeforeSync'
            : 'admin.providers.opencode.syncSavedSettings',
        )
      }}
    </p>
    <p v-if="error" role="alert" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
    <details class="space-y-3">
      <summary class="cursor-pointer text-sm font-medium">{{
        t('admin.providers.opencode.modelProtocols')
      }}</summary>
      <p class="input-hint">{{ t('admin.providers.opencode.modelProtocolsHint') }}</p>
      <div
        v-for="(target, model) in displayedProtocols"
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
          v-if="Object.prototype.hasOwnProperty.call(draft.protocols[draft.mode], model)"
          type="button"
          class="btn btn-ghost btn-sm"
          :aria-label="t('common.delete')"
          @click="removeProtocol(model)"
          ><Icon name="trash" size="sm"
        /></button>
        <span v-else class="text-xs text-gray-500 dark:text-gray-400">{{
          t('admin.providers.opencode.catalogSuggestion')
        }}</span>
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
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import * as providersAPI from '@/api/admin/providers'
import type { ProtocolID } from '@/api/protocols.types'
import type { OpenCodeDraft } from './openCodeDraft'

// OpenCodeFields.vue edits inference modes, exact model protocols, and endpoint
// overrides using the provider form's draft. Synchronization loads current
// upstream model IDs and their protocol declarations for the selected mode;
// the parent receives IDs for its whitelist while the draft keeps catalogue
// suggestions separate from administrator overrides. Both tables are saved by
// the parent, so newly discovered models retain their accepted catalogue route.
//
// A synchronization captures mode and root before awaiting the server. It
// ignores results after unmount or identity changes, preventing an earlier
// provider's results from replacing the active draft. Mode controls are disabled
// during that request. Configuration changes cancel the HTTP request, and saved
// records require their key, root, and mode edits to be saved first. Persistence
// belongs to the parent dialog, so synchronization changes its draft until saving.

// props supplies saved identity or temporary credentials for model synchronization.
const props = defineProps<{
  // providerId selects saved synchronization. When absent, the form sends the
  // supplied temporary key and root to the separate preview endpoint.
  providerId?: number
  // apiKey is blank for an unchanged saved credential. Entering a replacement
  // blocks saved synchronization until the parent saves the edited provider.
  apiKey?: string
  baseUrl?: string // Unsaved inference service root.
  // savedMode and savedBaseUrl identify the persisted query configuration. The
  // parent supplies these separately from editable values to detect unsaved changes.
  savedMode?: 'payg' | 'go' // Persisted inference service mode.
  savedBaseUrl?: string // Persisted inference service root.
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
// request identifies the invocation permitted to publish results; replacing or
// clearing it invalidates older responses even when their input values match again.
let request: AbortController | undefined

// syncNeedsSave prevents combining a saved secret with a temporary query target.
const syncNeedsSave = computed(
  () =>
    Boolean(props.providerId) &&
    (Boolean(props.apiKey?.trim()) ||
      draft.value.mode !== props.savedMode ||
      (props.baseUrl?.trim() ?? '') !== (props.savedBaseUrl?.trim() ?? '')),
)
// displayedProtocols keeps current manual choices ahead of the selected mode's catalogue suggestions.
const displayedProtocols = computed(() => ({
  ...draft.value.suggestions[draft.value.mode],
  ...draft.value.protocols[draft.value.mode],
}))

// Invalidate the active request as soon as its identity changes, including changes
// that return to their earlier value before the pending response reaches the form.
watch(
  () => [
    props.providerId,
    props.apiKey,
    props.baseUrl,
    props.savedMode,
    props.savedBaseUrl,
    draft.value,
    draft.value.mode,
  ],
  () => {
    request?.abort()
    request = undefined
    syncing.value = false
    error.value = ''
  },
  { flush: 'sync' },
)

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

// removeProtocol removes a manual override and restores its accepted catalogue choice.
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
// Saved records require a clean key, root, and mode; creation sends the caller's
// temporary credential. Each invocation owns its abort controller, so a changed
// configuration can start another request while the previous transport settles.
// Completion checks that ownership before publishing suggestions and whitelist
// IDs. Manual overrides remain untouched, including deletions made during the
// request, and current-mode suggestions are replaced as one catalogue result.
async function syncModels(): Promise<void> {
  if (syncing.value || syncNeedsSave.value || (!props.providerId && !props.apiKey?.trim())) return
  const mode = draft.value.mode
  const root = props.baseUrl
  const identity = props.providerId
  const key = props.apiKey
  const controller = new AbortController()
  request = controller
  syncing.value = true
  error.value = ''
  try {
    const result = identity
      ? await providersAPI.syncUpstreamModels(identity, controller.signal)
      : await providersAPI.syncUpstreamModelsPreview(
          {
            platform: 'opencode',
            type: 'apikey',
            api_key: key ?? '',
            base_url: root,
            provider_mode: mode,
          },
          controller.signal,
        )
    // A changed draft identifies a different synchronization target. Discard the
    // earlier response before publishing whitelist entries or accepted catalogue
    // suggestions for the currently edited provider.
    if (
      disposed ||
      request !== controller ||
      controller.signal.aborted ||
      identity !== props.providerId ||
      mode !== draft.value.mode ||
      root !== props.baseUrl ||
      key !== props.apiKey
    )
      return
    draft.value.suggestions[mode] = { ...result.model_protocols }
    emit('synced', result.models)
  } catch (cause) {
    if (
      !disposed &&
      request === controller &&
      !controller.signal.aborted &&
      identity === props.providerId &&
      mode === draft.value.mode &&
      root === props.baseUrl &&
      key === props.apiKey
    )
      error.value =
        cause instanceof Error ? cause.message : t('admin.providers.syncUpstreamModelsFailed')
  } finally {
    // A canceled invocation may settle after another starts. Release busy state
    // only while this invocation still owns the request.
    if (request === controller) {
      request = undefined
      if (!disposed) syncing.value = false
    }
  }
}

// Cancel the owned HTTP request and clear its identity before a disposed form
// can publish model IDs or protocol suggestions into the retained parent draft.
onBeforeUnmount(() => {
  disposed = true
  request?.abort()
  request = undefined
})
</script>
