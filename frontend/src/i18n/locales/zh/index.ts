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
    "chooseLanguage": "选择语言",
    "addTranslation": "添加翻译",
    "reviewNeeded": "待核对",
    "confirmReviewed": "确认已核对",
    "removeTranslation": "删除译文",
    "fallbackHint": "译文缺失或尚未核对时显示原文。",
    "preview": "预览",
    "showingOriginal": "当前显示原文。",
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
