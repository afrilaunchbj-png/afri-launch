import { Link, NavLink, Outlet, useNavigate } from "react-router"
import { useTranslation } from "react-i18next"
import {
  Coins,
  HelpCircle,
  LogOut,
  Rocket,
  Settings,
} from "lucide-react"

import { LanguageToggle } from "@/components/language-toggle"
import { futureNav, mainNav } from "@/components/layout/nav-items"
import { ThemeToggle } from "@/components/theme-toggle"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/features/auth/auth-provider"
import { useSignOut } from "@/features/auth/hooks"
import { useCreditsSummary } from "@/features/credits/hooks"
import { cn } from "@/lib/utils"

function UserAvatar({ name }: { name: string }) {
  const initials = name
    .split(/\s+/)
    .map((p) => p[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()

  return (
    <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">
      {initials || "?"}
    </div>
  )
}

function CreditsBalance() {
  const { t } = useTranslation()
  const { data, isLoading } = useCreditsSummary()

  if (isLoading) {
    return <Skeleton className="h-4 w-20" />
  }

  return (
    <span className="flex items-center gap-1 text-xs font-medium text-warning">
      <Coins className="h-3.5 w-3.5" />
      {data?.summary.available ?? 0} {t("credits:label")}
    </span>
  )
}

function SidebarNav() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const signOut = useSignOut()
  const navigate = useNavigate()

  const handleLogout = () => {
    signOut.mutate(undefined, { onSettled: () => navigate("/login") })
  }

  return (
    <aside className="fixed inset-y-0 left-0 z-40 hidden w-64 flex-col border-r bg-muted/30 md:flex">
      <div className="flex items-center gap-2 px-6 pb-4 pt-6">
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <Rocket className="h-5 w-5" />
        </div>
        <Link to="/dashboard" className="font-display text-lg font-semibold text-primary">
          {t("appName")}
        </Link>
      </div>

      <div className="px-4">
        <div className="flex items-center gap-3 rounded-lg border bg-card p-3">
          <UserAvatar name={user?.name ?? ""} />
          <div className="min-w-0">
            <p className="truncate text-sm font-medium">{user?.name}</p>
            <CreditsBalance />
          </div>
        </div>
      </div>

      <nav className="flex-1 space-y-1 overflow-y-auto px-4 py-4">
        {mainNav.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                isActive
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:bg-muted",
              )
            }
          >
            <item.icon className="h-4 w-4" />
            {t(item.labelKey)}
          </NavLink>
        ))}

        <div className="pt-3">
          <p className="px-3 pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
            {t("nav.comingSoon")}
          </p>
          {futureNav.map((item) => (
            <div
              key={item.labelKey}
              className="flex cursor-not-allowed items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-muted-foreground/60"
            >
              <item.icon className="h-4 w-4" />
              {t(item.labelKey)}
              <Badge variant="outline" className="ml-auto text-[10px]">
                {t("nav.soon")}
              </Badge>
            </div>
          ))}
        </div>
      </nav>

      <div className="space-y-1 border-t px-4 py-4">
        <div className="flex items-center gap-2">
          <LanguageToggle />
          <ThemeToggle />
          <Button
            variant="ghost"
            size="icon"
            className="ml-auto text-muted-foreground"
            onClick={handleLogout}
            disabled={signOut.isPending}
            aria-label={t("auth:logout")}
          >
            <LogOut className="h-4 w-4" />
          </Button>
        </div>
        <div className="flex cursor-not-allowed items-center gap-3 rounded-lg px-3 py-2 text-sm text-muted-foreground/60">
          <Settings className="h-4 w-4" />
          {t("nav.settings")}
        </div>
        <div className="flex cursor-not-allowed items-center gap-3 rounded-lg px-3 py-2 text-sm text-muted-foreground/60">
          <HelpCircle className="h-4 w-4" />
          {t("nav.support")}
        </div>
      </div>
    </aside>
  )
}

function MobileTopBar() {
  const { t } = useTranslation()
  return (
    <header className="sticky top-0 z-30 flex items-center justify-between border-b bg-background px-4 py-3 md:hidden">
      <div className="flex items-center gap-2">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <Rocket className="h-4 w-4" />
        </div>
        <span className="font-display text-base font-semibold text-primary">{t("appName")}</span>
      </div>
      <div className="flex items-center gap-1">
        <LanguageToggle />
        <ThemeToggle />
      </div>
    </header>
  )
}

function BottomNav() {
  const { t } = useTranslation()
  return (
    <nav className="fixed inset-x-0 bottom-0 z-40 flex items-center justify-around border-t bg-card py-2 md:hidden">
      {mainNav.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          className={({ isActive }) =>
            cn(
              "flex min-w-[64px] flex-col items-center justify-center gap-1 rounded-2xl px-3 py-1 text-xs font-medium transition-colors",
              isActive ? "text-primary" : "text-muted-foreground",
            )
          }
        >
          <item.icon className="h-5 w-5" />
          {t(item.labelKey)}
        </NavLink>
      ))}
    </nav>
  )
}

export default function AppLayout() {
  return (
    <div className="min-h-screen md:pl-64">
      <SidebarNav />
      <MobileTopBar />
      <main className="mx-auto w-full max-w-7xl px-4 py-6 pb-24 md:px-8 md:pb-10">
        <Outlet />
      </main>
      <BottomNav />
    </div>
  )
}
