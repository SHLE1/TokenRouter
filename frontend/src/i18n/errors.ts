import messages from '../../../backend/internal/pkg/locale/error_messages.json'
import { getLocale } from './index'
import { normalizeLocale, defaultLocale } from './catalog'

// 错误码用于业务判断，用户提示按当前语言生成。
export function localizedErrorMessage(reason: unknown, status: number, original = ''): string {
  const code = typeof reason === 'string' ? reason.toUpperCase() : ''
  const language = normalizeLocale(getLocale()) || defaultLocale
  if (language === 'en' && original.trim() && !/[\u3400-\u9fff]/.test(original)) return original
  if (!code && language === 'zh-Hans' && /[\u3400-\u9fff]/.test(original)) return original
  const dictionary = (messages as Record<string, Record<string, string>>)[language] || messages.en
  return dictionary[code] || dictionary[`HTTP_${status}`] || dictionary.REQUEST_FAILED
}

// 管理员专用错误维持现有行为，翻译编辑的版本和输入错误使用共同提示。
export function responseErrorMessage(path: string, reason: unknown, status: number, original = ''): string {
  const code = typeof reason === 'string' ? reason.toUpperCase() : ''
  const contentError = /^(LOCALIZATION_|SOURCE_LOCALE_|INVALID_LOCALE|INVALID_TRANSLATION_|DUPLICATE_LOCALE|UNKNOWN_LOCALIZED_|LOCALIZED_|SITE_NAME_REQUIRED|LEGAL_|NAVIGATION_LABEL_|INVALID_LINK_URL|DUPLICATE_CONTENT_ID|LOCALIZED_PAGE_|ENDPOINT_NAME_REQUIRED|PLAN_NAME_REQUIRED|PLAN_TEXT_TOO_LONG|GROUP_DISPLAY_|GROUP_DESCRIPTION_|MODEL_DISPLAY_|INVALID_BLOCK_MESSAGE|INVALID_CUSTOM_MESSAGE|PAYMENT_METHOD_NAME_REQUIRED)/.test(code)
  if (path.includes('/admin/') && !contentError) return original
  return localizedErrorMessage(reason, status, original)
}
