import i18n from "i18next"
import { initReactI18next } from "react-i18next"

import authEn from "./locales/en/auth.json"
import chatEn from "./locales/en/chat.json"
import commonEn from "./locales/en/common.json"
import creditsEn from "./locales/en/credits.json"
import dashboardEn from "./locales/en/dashboard.json"
import projectsEn from "./locales/en/projects.json"
import authFr from "./locales/fr/auth.json"
import chatFr from "./locales/fr/chat.json"
import commonFr from "./locales/fr/common.json"
import creditsFr from "./locales/fr/credits.json"
import dashboardFr from "./locales/fr/dashboard.json"
import projectsFr from "./locales/fr/projects.json"

export const resources = {
  fr: {
    common: commonFr,
    auth: authFr,
    credits: creditsFr,
    dashboard: dashboardFr,
    chat: chatFr,
    projects: projectsFr,
  },
  en: {
    common: commonEn,
    auth: authEn,
    credits: creditsEn,
    dashboard: dashboardEn,
    chat: chatEn,
    projects: projectsEn,
  },
} as const

const LANGUAGE_STORAGE_KEY = "afrilaunch-lang"

/** setStoredLanguage mémorise la langue (utilisée avant le login / hors compte). */
export function setStoredLanguage(language: string) {
  try {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, language)
  } catch {
    /* stockage indisponible : on ignore */
  }
}

function getStoredLanguage(): string | undefined {
  try {
    const stored = window.localStorage.getItem(LANGUAGE_STORAGE_KEY)
    return stored === "fr" || stored === "en" ? stored : undefined
  } catch {
    return undefined
  }
}

void i18n.use(initReactI18next).init({
  resources,
  lng: getStoredLanguage() ?? "fr",
  fallbackLng: "fr",
  supportedLngs: ["fr", "en"],
  defaultNS: "common",
  interpolation: { escapeValue: false },
})

export default i18n
