import { Link, Outlet } from "react-router"
import { useTranslation } from "react-i18next"
import { Rocket } from "lucide-react"

import { LanguageToggle } from "@/components/language-toggle"
import { ThemeToggle } from "@/components/theme-toggle"

export default function AuthLayout() {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-screen flex-col bg-dot-pattern">
      <header className="flex items-center justify-between px-4 py-4">
        <Link to="/" className="flex items-center gap-2 font-display text-lg font-semibold text-primary">
          <Rocket className="h-5 w-5" />
          {t("appName")}
        </Link>
        <div className="flex items-center gap-1">
          <LanguageToggle />
          <ThemeToggle />
        </div>
      </header>
      <main className="flex flex-1 items-center justify-center px-4 py-8">
        <Outlet />
      </main>
    </div>
  )
}
