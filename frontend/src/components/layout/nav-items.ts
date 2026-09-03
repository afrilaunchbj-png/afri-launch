import {
  Coins,
  FolderOpen,
  LayoutDashboard,
  MessageCircle,
  Megaphone,
  Settings,
  ShieldCheck,
  LifeBuoy,
  type LucideIcon,
} from "lucide-react"

export interface NavItem {
  to: string
  icon: LucideIcon
  labelKey: string
}

export interface DisabledNavItem {
  icon: LucideIcon
  labelKey: string
}

/**
 * buildMainNav construit la navigation principale selon le rôle.
 * Les items Administration n'apparaissent que pour un superadmin.
 */
export function buildMainNav(isSuperadmin: boolean): NavItem[] {
  const items: NavItem[] = [
    { to: "/dashboard", icon: LayoutDashboard, labelKey: "nav.dashboard" },
    { to: "/discover", icon: MessageCircle, labelKey: "nav.discover" },
    { to: "/projects", icon: FolderOpen, labelKey: "nav.projects" },
    { to: "/credits", icon: Coins, labelKey: "nav.credits" },
    { to: "/support", icon: LifeBuoy, labelKey: "nav.support" },
    { to: "/settings", icon: Settings, labelKey: "nav.settings" },
  ]
  if (isSuperadmin) {
    items.push({ to: "/admin", icon: ShieldCheck, labelKey: "nav.admin" })
  }
  return items
}

/** Navigation mobile : 5 entrées max, l'administration remplace les crédits. */
export function buildMobileNav(isSuperadmin: boolean): NavItem[] {
  return buildMainNav(isSuperadmin).filter(
    (i) => i.to !== "/dashboard" && i.to !== (isSuperadmin ? "/credits" : "/admin"),
  )
}

export const futureNav: DisabledNavItem[] = [
  { icon: Megaphone, labelKey: "nav.marketing" },
]
