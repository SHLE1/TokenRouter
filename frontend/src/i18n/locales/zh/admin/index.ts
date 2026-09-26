import overview from './overview'
import pricing from './pricing'
import accounts from './accounts'
import resources from './resources'
import ops from './ops'
import settings from './settings'
import audit from './audit'

export default {
    protocols: {
      nativeTitle: '原生支持协议',
      nativeHint: '仅列出此账号类型原生支持的协议。全部关闭后不承接新调用。',
      loadError: '协议目录加载失败，请重新打开表单。',
      fallback: '账号不支持时转换为',
      fallbackWhen: '不支持时',
      auto: '自动匹配转换',
      restricted: '限制转换目标（按顺序）',
      nativeOnly: '仅原生，不转换',
      imagePolicy: '对话中的图片生成（Responses）',
      imagePolicyHint: '控制对话请求中的生图工具，不影响独立的图片生成、编辑接口。',
      imagePolicyOptions: {
        inherit: {
          label: '跟随账号和全局设置',
          description: '本分组不单独指定，先使用账号设置；账号未配置时，使用全局设置。',
        },
        enabled: {
          label: '自动添加生图工具',
          description: '普通对话会自动添加生图工具，精简模式（Responses Lite）除外。客户端自带的生图工具仍会保留。',
        },
        disabled: {
          label: '仅保留客户端自带的生图工具',
          description: '网关不额外添加生图工具；客户端请求中已有的生图工具会保留。',
        },
        block: {
          label: '移除生图工具',
          description: '不自动添加生图工具，并移除客户端请求中的生图工具。直接调用图片模型或图片接口不受影响。',
        },
      },
      groupTitle: '协议控制',
      groupHint: '先尝试账号原生协议，再按允许的转换目标顺序尝试。自动模式使用服务端支持的转换路线。',
    },
  ...overview,
  ...pricing,
  ...accounts,
  ...resources,
  ...ops,
  ...settings,
  ...audit,
}
