import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import batchImage from './batchImage'
import creative from './creative'
import admin from './admin'
import misc from './misc'
import team from './team'

export default {
  legal: {
    "login": "登录",
    "loadFailed": "文档加载失败",
    "retry": "请刷新页面重试。",
    "notFound": "文档不存在",
    "notFoundDescription": "当前文档不存在或已被移除。",
    "title": "登录条款",
    "empty": "暂无正文内容",
    "updatedAt": "更新日期：{date}",
    "acceptedPrefix": "我已阅读并同意",
    "consentRequired": "继续登录前需要先同意最新条款。",
    "disabledUntilAccepted": "同意后可以使用账号密码和快捷登录。",
    "viewTerms": "查看条款",
    "updateNotice": "条款更新通知",
    "changedNotice": "服务条款于 {date} 更新，请阅读相关文档后再继续。",
    "recently": "近期",
    "documents": "相关文档",
    "reject": "拒绝",
    "accept": "同意并继续"
  },
  notFoundPage: {
    "description": "页面不存在或已被移动。",
    "back": "返回上一页",
    "dashboard": "前往控制台",
    "help": "需要帮助？",
    "support": "联系客服"
  },

  localization: {
    "useAsOriginal": "设为原文",
    "languageConflict": "此语言已有译文。请打开该译文设为原文，或先删除对应译文。",
    "nameRequired": "请填写名称。",
    "displayName": "展示名称",
    "productName": "支付商品名称",
    "fields": {
      "site_name": "站点名称",
      "site_title": "首页标题",
      "site_subtitle": "副标题"
    },
    "original": "原文",
    "originalLanguage": "原文语言",
    "chooseOriginalLanguage": "选择原文语言",
    "unknownOriginal": "原文语言未知。修改原文前，在右上角选择原文语言；该语言已有译文时，点那条译文的“设为原文”。",
    "addTranslation": "添加翻译",
    "addLocale": "添加译文",
    "languageCount": "{n} 种语言",
    "needsUpdate": "待更新",
    "dialogTitle": "翻译 · {field}",
    "dialogHint": "没有译文或原文修改后，用户看到原文。",
    "status": {
      "translated": "已翻译",
      "stale": "原文已修改",
      "missing": "未翻译，用户看到原文"
    },
    "staleNotice": "这条译文对照的是修改前的原文，暂不展示。",
    "stillValid": "仍然适用",
    "removeTranslation": "删除译文",
    "done": "完成",
    "defaultLanguage": "默认语言"
  },
  ...landing,
  ...common,
  ...dashboard,
  ...batchImage,
  ...creative,
  admin,
  ...misc,
  ...team,
}
