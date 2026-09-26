import overview from './overview'
import pricing from './pricing'
import accounts from './accounts'
import resources from './resources'
import ops from './ops'
import settings from './settings'
import audit from './audit'

export default {
    protocols: {
      nativeTitle: 'Native upstream protocols',
      nativeHint: 'Only native protocols for this account type are shown. An empty set disables new calls.',
      loadError: 'Failed to load protocol capabilities. Please reopen the form.',
      fallback: 'Convert when unavailable',
      fallbackWhen: 'If unavailable',
      auto: 'Automatic conversion',
      restricted: 'Ordered conversion targets',
      nativeOnly: 'Native only',
      imagePolicy: 'Image generation in conversations (Responses)',
      imagePolicyHint: 'Controls image-generation tools in conversations. Separate image generation and editing endpoints are unaffected.',
      imagePolicyOptions: {
        inherit: {
          label: 'Follow account and global settings',
          description: 'Use the account setting for this group. If the account has no override, use the global setting.',
        },
        enabled: {
          label: 'Automatically add an image tool',
          description: 'Add an image-generation tool to regular conversations, except Responses Lite requests. Keep any image tools provided by the client.',
        },
        disabled: {
          label: 'Keep only client-provided image tools',
          description: 'Do not add an image-generation tool. Keep image tools already included in the client request.',
        },
        block: {
          label: 'Remove image tools',
          description: 'Do not add image tools, and remove those included in the client request. Direct calls to image models or image endpoints are unaffected.',
        },
      },
      groupTitle: 'Protocol controls',
      groupHint: 'Try each account’s native protocol first, then the allowed conversion targets in order. Automatic mode uses the routes supported by the server.',
    },
  ...overview,
  ...pricing,
  ...accounts,
  ...resources,
  ...ops,
  ...settings,
  ...audit,
}
