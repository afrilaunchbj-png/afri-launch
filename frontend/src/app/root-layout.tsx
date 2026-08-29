import { Link, Outlet } from "react-router"
import { useTranslation } from "react-i18next"

import { LanguageToggle } from "@/components/language-toggle"
import { ThemeToggle } from "@/components/theme-toggle"
import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export default function RootLayout() {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur">
        <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <Link to="/" className="font-display text-lg font-semibold text-primary">
            {t("appName")}
          </Link>
          <nav className="flex items-center gap-2">
            <Link to="/login" className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}>
              {t("auth:login")}
            </Link>
            <Link to="/register" className={cn(buttonVariants({ size: "sm" }))}>
              {t("auth:register")}
            </Link>
            <LanguageToggle />
            <ThemeToggle />
          </nav>
        </div>
      </header>
      <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-8 sm:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  )
}
