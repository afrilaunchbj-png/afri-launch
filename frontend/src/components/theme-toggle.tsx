import { Moon, Sun } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useTheme } from "@/components/theme-provider"
import { useUpdatePreferences } from "@/features/preferences/hooks"
import type { ThemePreference } from "@/features/preferences/api"

export function ThemeToggle() {
  const { setTheme } = useTheme()
  const { t } = useTranslation()
  const update = useUpdatePreferences()

  const change = (theme: ThemePreference) => {
    setTheme(theme)
    update.mutate({ theme })
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="icon">
          <Sun className="h-4 w-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
          <Moon className="absolute h-4 w-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
          <span className="sr-only">{t("theme.system")}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => change("light")}>{t("theme.light")}</DropdownMenuItem>
        <DropdownMenuItem onClick={() => change("dark")}>{t("theme.dark")}</DropdownMenuItem>
        <DropdownMenuItem onClick={() => change("system")}>{t("theme.system")}</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
