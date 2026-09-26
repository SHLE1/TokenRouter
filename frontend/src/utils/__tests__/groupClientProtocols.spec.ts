import { useProtocolCatalogFixture } from '@/__tests__/helpers/protocolCatalog'
import { describe, expect, it } from 'vitest'
import { protocolCatalog } from '@/api/admin/protocolCapabilities'
import { defaultGroupClientProtocols, effectiveGroupClientProtocols, setGroupClientProtocol, supportedGroupClientProtocols } from '../groupClientProtocols'

useProtocolCatalogFixture()
describe('通用分组协议', () => {
  it('所有分组使用统一入口目录，新组只默认开放三个文本入口', () => {
    expect(supportedGroupClientProtocols()).toEqual(protocolCatalog.value!.protocols.filter(protocol => !protocol.upstream_only).map(protocol => protocol.id))
    expect(defaultGroupClientProtocols()).toEqual(['anthropic_messages', 'openai_responses', 'openai_chat_completions'])
  })
  it('缺少配置和显式空集合都不自动开放入口', () => {
    expect(effectiveGroupClientProtocols(undefined)).toEqual([])
    expect(effectiveGroupClientProtocols([])).toEqual([])
  })
  it('全部入口均可关闭，排序由服务端目录维护', () => {
    expect(setGroupClientProtocol(['openai_chat_completions', 'openai_responses'], 'openai_responses', false)).toEqual(['openai_chat_completions'])
    expect(setGroupClientProtocol(['openai_chat_completions'], 'openai_chat_completions', false)).toEqual([])
    expect(setGroupClientProtocol([], 'qoder_chat', true)).toEqual([])
  })
  it('用户侧直接读取后端集合，不需要管理员目录', () => {
    protocolCatalog.value = null
    expect(effectiveGroupClientProtocols(['openai_responses', 'openai_images_edits'])).toEqual(['openai_responses', 'openai_images_edits'])
  })
})
