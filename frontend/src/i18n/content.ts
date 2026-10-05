import { availableLocales, normalizeLocale } from './catalog'

// 译文记录其已经核对的原文版本。
export interface Translation<T> {
  value: T
  source_revision: number
}

export interface Localized<T> {
  source_locale: string | null
  source: T
  translations: Record<string, Translation<T>>
  revision: number
  source_revision: number
}

export interface LocalizedUpdate<T> extends Localized<T> {
  reviewed_locales?: string[]
  deleted_locales?: string[]
}

// 历史内容在管理员确认前保持未知原文语言。
export function originalContent<T>(source: T, sourceLocale: string | null = null): Localized<T> {
  return { source_locale: sourceLocale, source, translations: {}, revision: 0, source_revision: 0 }
}

// 预览与服务端使用相同的语言顺序和译文核对条件。
export function resolveContent<T>(content: Localized<T>, requested: string): { value: T; locale: string | null; fallback: boolean } {
  const code = normalizeLocale(requested)
  const definition = availableLocales.find(item => item.code === code)
  for (const candidate of definition ? [definition.code, ...definition.fallbacks] : []) {
    if (normalizeLocale(content.source_locale) === candidate) {
      return { value: content.source, locale: content.source_locale, fallback: candidate !== code }
    }
    const translation = content.translations[candidate]
    if (translation && translation.source_revision === content.source_revision) {
      return { value: translation.value, locale: candidate, fallback: candidate !== code }
    }
  }
  return { value: content.source, locale: content.source_locale, fallback: !content.source_locale || normalizeLocale(content.source_locale) !== code }
}

let contentSequence = 0

// 内容标识用于译文对应和列表重排，普通 HTTP 环境也能创建条目。
export function newContentID(): string {
  return globalThis.crypto?.randomUUID?.() || `content-${Date.now().toString(36)}-${++contentSequence}-${Math.random().toString(36).slice(2)}`
}
