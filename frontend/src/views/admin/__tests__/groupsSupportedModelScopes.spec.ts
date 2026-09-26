import { describe, expect, it } from 'vitest'
import { normalizeSupportedModelScopesForPlatform } from '../groupsSupportedModelScopes'

describe('分组模型系列策略', () => {
  it('保留显式策略，空配置不增加限制', () => {
    expect(normalizeSupportedModelScopesForPlatform(['claude', 'gemini_text'])).toEqual(['claude', 'gemini_text'])
    expect(normalizeSupportedModelScopesForPlatform(undefined)).toEqual([])
  })
})
