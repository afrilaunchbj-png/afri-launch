import { api, type ApiSingle } from "@/lib/api/client"

export type Language = "fr" | "en"
export type ThemePreference = "light" | "dark" | "system"

export interface Preferences {
  language: Language
  theme: ThemePreference
}

export function fetchPreferences() {
  return api.get<ApiSingle<Preferences>>("/api/v1/preferences").then((r) => r.data)
}

export function updatePreferences(input: Partial<Preferences>) {
  return api.put<ApiSingle<Preferences>>("/api/v1/preferences", input).then((r) => r.data)
}
