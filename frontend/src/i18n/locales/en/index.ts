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
    "login": "Log in",
    "loadFailed": "Could not load document",
    "retry": "Refresh the page to try again.",
    "notFound": "Document not found",
    "notFoundDescription": "This document does not exist or has been removed.",
    "title": "Login terms",
    "empty": "No content has been provided.",
    "updatedAt": "Updated: {date}",
    "acceptedPrefix": "I have read and agree to",
    "consentRequired": "Accept the latest terms to continue signing in.",
    "disabledUntilAccepted": "Password and quick sign-in become available after you accept.",
    "viewTerms": "View terms",
    "updateNotice": "Terms updated",
    "changedNotice": "Our terms were updated on {date}. Read the documents before continuing.",
    "recently": "a recent date",
    "documents": "Documents",
    "reject": "Decline",
    "accept": "Accept and continue"
  },
  notFoundPage: {
    "description": "The page you are looking for does not exist or has moved.",
    "back": "Go back",
    "dashboard": "Go to dashboard",
    "help": "Need help?",
    "support": "Contact support"
  },

  localization: {
    "useAsOriginal": "Use as original",
    "languageConflict": "This language already has a translation. Open that translation to use it as the original, or remove it first.",
    "nameRequired": "Enter a name.",
    "displayName": "Display name",
    "productName": "Payment product name",
    "fields": {
      "site_name": "Site name",
      "site_title": "Homepage title",
      "site_subtitle": "Subtitle"
    },
    "original": "Original",
    "originalLanguage": "Original language",
    "chooseOriginalLanguage": "Choose original language",
    "unknownOriginal": "The original language is unknown. Before editing the original, choose its language at the top right. If that language already has a translation, use \"Use as original\" on that translation.",
    "addTranslation": "Add translation",
    "addLocale": "Add translation",
    "languageCount": "{n} languages",
    "needsUpdate": "Needs update",
    "dialogTitle": "Translate · {field}",
    "dialogHint": "Users see the original when a translation is missing or the original has changed.",
    "status": {
      "translated": "Translated",
      "stale": "Original changed",
      "missing": "Not translated, users see the original"
    },
    "staleNotice": "This translation matches an earlier original and is hidden for now.",
    "stillValid": "Still fits",
    "removeTranslation": "Remove translation",
    "done": "Done",
    "defaultLanguage": "Default language"
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
