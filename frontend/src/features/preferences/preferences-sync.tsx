import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { useTheme } from "@/components/theme-provider"
import { setStoredLanguage } from "@/i18n"

import { usePreferences } from "./hooks"

/**
 * PreferencesSync applique les préférences du compte (langue, thème) une
 * fois l'utilisateur connecté : la même expérience sur tous les appareils,
 * la DB étant la source de vérité (le localStorage ne sert qu'avant login).
 */
export function PreferencesSync() {
  const { data } = usePreferences()
  const { setTheme } = useTheme()
  const { i18n } = useTranslation()

  useEffect(() => {
    if (!data) return
    if (data.language && data.language !== i18n.language) {
      void i18n.changeLanguage(data.language)
    }
    setStoredLanguage(data.language)
    setTheme(data.theme)
  }, [data, i18n, setTheme])

  return null
}
