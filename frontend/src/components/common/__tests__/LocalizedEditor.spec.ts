import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { defineComponent, ref } from 'vue'
import LocalizedEditor from '../LocalizedEditor.vue'
import Select from '../Select.vue'
import { resolveContent, type LocalizedUpdate } from '@/i18n/content'

// 使用受控父组件保存草稿，语言标签切换和重新渲染经过相同的 v-model 路径。
function editor() {
  const content = ref<LocalizedUpdate<string>>({ source: '原文', source_locale: 'zh-Hans', revision: 3, source_revision: 2, translations: { en: { value: 'English', source_revision: 2 } } })
  const parent = defineComponent({ components: { LocalizedEditor }, setup: () => ({ content }), template: '<LocalizedEditor v-model="content" />' })
  const wrapper = mount(parent, { global: { plugins: [createI18n({ legacy: false, locale: 'en', missingWarn: false, fallbackWarn: false, messages: {} })] } })
  return { wrapper, content }
}

describe('翻译编辑', () => {
  it('原文修改后回退，译文确认后参与预览', async () => {
    const { wrapper, content } = editor()
    await wrapper.get('input').setValue('修改后')
    expect(resolveContent(content.value, 'en').value).toBe('修改后')
    await wrapper.findAll('button').find(item => item.text().includes('English'))!.trigger('click')
    expect((wrapper.get('input').element as HTMLInputElement).value).toBe('English')
    await wrapper.findAll('button').find(item => item.text() === 'localization.confirmReviewed')!.trigger('click')
    expect(content.value.reviewed_locales).toEqual(['en'])
    expect(resolveContent(content.value, 'en').value).toBe('English')
    wrapper.unmount()
  })

  it('删除译文记录操作并保留原文草稿', async () => {
    const { wrapper, content } = editor()
    await wrapper.findAll('button').find(item => item.text() === 'English')!.trigger('click')
    await wrapper.findAll('button').find(item => item.text() === 'localization.removeTranslation')!.trigger('click')
    expect(content.value.deleted_locales).toEqual(['en'])
    expect(content.value.translations).toEqual({})
    expect(content.value.source).toBe('原文')
    wrapper.unmount()
  })

  it('切换原文语言提示冲突，提升译文保留已知语言的原文', async () => {
    const { wrapper, content } = editor()
    wrapper.findComponent(Select).vm.$emit('update:modelValue', 'en')
    await wrapper.vm.$nextTick()
    expect(content.value.source_locale).toBe('zh-Hans')
    expect(wrapper.text()).toContain('localization.languageConflict')
    await wrapper.findAll('button').find(item => item.text() === 'English')!.trigger('click')
    await wrapper.findAll('button').find(item => item.text() === 'localization.useAsOriginal')!.trigger('click')
    expect(content.value.source).toBe('English')
    expect(content.value.source_locale).toBe('en')
    expect(content.value.translations['zh-Hans'].value).toBe('原文')
    expect(content.value.deleted_locales).toEqual(['en'])
    expect(resolveContent(content.value, 'zh').value).toBe('English')
    wrapper.unmount()
  })
})
