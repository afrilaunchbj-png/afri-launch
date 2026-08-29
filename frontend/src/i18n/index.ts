import i18n from "i18next"
import { initReactI18next } from "react-i18next"

import authEn from "./locales/en/auth.json"
import commonEn from "./locales/en/common.json"
import creditsEn from "./locales/en/credits.json"
import dashboardEn from "./locales/en/dashboard.json"
import opportunitiesEn from "./locales/en/opportunities.json"
import authFr from "./locales/fr/auth.json"
import commonFr from "./locales/fr/common.json"
import creditsFr from "./locales/fr/credits.json"
import dashboardFr from "./locales/fr/dashboard.json"
import opportunitiesFr from "./locales/fr/opportunities.json"

export const resources = {
  fr: {
    common: commonFr,
    auth: authFr,
    credits: creditsFr,
    dashboard: dashboardFr,
    opportunities: opportunitiesFr,
  },
  en: {
    common: commonEn,
    auth: authEn,
    credits: creditsEn,
    dashboard: dashboardEn,
    opportunities: opportunitiesEn,
  },
} as const

void i18n.use(initReactI18next).init({
  resources,
  lng: "fr",
  fallbackLng: "fr",
  supportedLngs: ["fr", "en"],
  defaultNS: "common",
  interpolation: { escapeValue: false },
})

export default i18n
