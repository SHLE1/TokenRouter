import type { ProtocolID } from '@/api/protocols.types'

// openCodeDraft.ts translates OpenCode form state into mode-scoped credentials.
// Each mode owns its model protocol overrides, while protocol address overrides
// apply to the selected provider. Hydration copies nested maps so canceling a
// dialog cannot modify the provider response retained by its parent.

// OpenCodeDraft contains the editable service mode, model encodings, and addresses.
export interface OpenCodeDraft {
  mode: 'payg' | 'go' // Selected inference service mode.
  protocols: Record<'payg' | 'go', Record<string, ProtocolID>> // Independent exact-model encodings.
  baseUrls: Record<string, string> // Complete protocol endpoint bases.
}

// openCodeDraft copies saved OpenCode settings or creates an empty payg draft.
// Missing mode tables become independently owned objects. Secret credentials
// stay in the owning provider form and are excluded from this draft.
export function openCodeDraft(credentials: Record<string, unknown> = {}): OpenCodeDraft {
  const modes = credentials.opencode_model_protocols as OpenCodeDraft['protocols'] | undefined
  return {
    mode: credentials.provider_mode === 'go' ? 'go' : 'payg',
    protocols: { payg: { ...modes?.payg }, go: { ...modes?.go } },
    baseUrls: { ...(credentials.api_base_urls as Record<string, string> | undefined) },
  }
}

// openCodeCredentials serializes independent mode tables for provider persistence.
// Empty addresses are omitted so the selected mode supplies official defaults.
// Calling forms merge this result into their credential payload before saving.
export function openCodeCredentials(draft: OpenCodeDraft): Record<string, unknown> {
  return {
    provider_mode: draft.mode,
    opencode_model_protocols: { payg: { ...draft.protocols.payg }, go: { ...draft.protocols.go } },
    api_base_urls: Object.fromEntries(
      Object.entries(draft.baseUrls)
        .filter(([, value]) => value.trim())
        .map(([key, value]) => [key, value.trim()]),
    ),
  }
}
