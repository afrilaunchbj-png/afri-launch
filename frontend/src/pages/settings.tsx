import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import { Coins, LifeBuoy, Settings as SettingsIcon, User as UserIcon } from "lucide-react"

import { useAuth } from "@/features/auth/auth-provider"
import { useMe } from "@/features/auth/hooks"
import { useCreditsSummary } from "@/features/credits/hooks"
import type { Language, ThemePreference } from "@/features/preferences/api"
import { usePreferences, useUpdatePreferences } from "@/features/preferences/hooks"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

export default function SettingsPage() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const { data: me } = useMe()
  const { data: prefs } = usePreferences()
  const updatePrefs = useUpdatePreferences()
  const { data: credits } = useCreditsSummary()

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <header>
        <h1 className="flex items-center gap-2 font-display text-2xl font-bold text-primary md:text-3xl">
          <SettingsIcon className="h-6 w-6" />
          {t("settings:title")}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("settings:subtitle")}</p>
      </header>

      {/* Profil (identité gérée par Neon Auth : lecture seule) */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <UserIcon className="h-4 w-4 text-primary" />
            {t("settings:profile")}
          </CardTitle>
        </CardHeader>
        <CardContent className="flex items-center gap-4">
          <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-primary text-lg font-semibold text-primary-foreground">
            {(user?.name ?? "?")
              .split(/\s+/)
              .map((p) => p[0])
              .filter(Boolean)
              .slice(0, 2)
              .join("")
              .toUpperCase()}
          </div>
          <div className="min-w-0">
            <p className="truncate font-medium">{user?.name || me?.full_name}</p>
            <p className="truncate text-sm text-muted-foreground">{user?.email}</p>
            {me?.role === "superadmin" ? (
              <Badge className="mt-1" variant="secondary">
                {t("settings:superadmin")}
              </Badge>
            ) : null}
          </div>
          {me?.created_at ? (
            <p className="ml-auto hidden text-xs text-muted-foreground sm:block">
              {t("settings:memberSince", { date: new Date(me.created_at).toLocaleDateString() })}
            </p>
          ) : null}
        </CardContent>
      </Card>

      {/* Préférences (persistées en DB — synchronisées sur tous les appareils) */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("settings:preferences")}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1.5">
            <label htmlFor="pref-language" className="text-sm font-medium">
              {t("settings:language")}
            </label>
            <Select
              value={prefs?.language ?? "fr"}
              onValueChange={(v) => updatePrefs.mutate({ language: v as Language })}
            >
              <SelectTrigger id="pref-language">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="fr">Français</SelectItem>
                <SelectItem value="en">English</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <label htmlFor="pref-theme" className="text-sm font-medium">
              {t("settings:theme")}
            </label>
            <Select
              value={prefs?.theme ?? "system"}
              onValueChange={(v) => updatePrefs.mutate({ theme: v as ThemePreference })}
            >
              <SelectTrigger id="pref-theme">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="light">{t("theme.light")}</SelectItem>
                <SelectItem value="dark">{t("theme.dark")}</SelectItem>
                <SelectItem value="system">{t("theme.system")}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <p className="text-xs text-muted-foreground sm:col-span-2">{t("settings:preferencesHint")}</p>
        </CardContent>
      </Card>

      {/* Compte */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Coins className="h-4 w-4 text-primary" />
            {t("settings:account")}
          </CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-between gap-4">
          <div>
            <p className="font-display text-xl font-bold text-warning">
              {credits?.summary.available ?? 0} {t("credits:label")}
            </p>
            <p className="text-xs text-muted-foreground">{t("settings:balanceHint")}</p>
          </div>
          <Link
            to="/support"
            className="inline-flex items-center gap-1.5 text-sm text-primary hover:underline"
          >
            <LifeBuoy className="h-4 w-4" />
            {t("settings:needHelp")}
          </Link>
        </CardContent>
      </Card>
    </div>
  )
}
