import { Languages } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { setStoredLanguage } from "@/i18n"
import { useUpdatePreferences } from "@/features/preferences/hooks"
import type { Language } from "@/features/preferences/api"

export function LanguageToggle() {
  const { i18n } = useTranslation()
  const update = useUpdatePreferences()

  const change = (language: Language) => {
    void i18n.changeLanguage(language)
    setStoredLanguage(language)
    update.mutate({ language })
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="icon">
          <Languages className="h-4 w-4" />
          <span className="sr-only">{i18n.language.toUpperCase()}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => change("fr")}>Français</DropdownMenuItem>
        <DropdownMenuItem onClick={() => change("en")}>English</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
