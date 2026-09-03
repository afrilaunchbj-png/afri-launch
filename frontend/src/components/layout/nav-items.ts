import {
  Coins,
  FolderOpen,
  LayoutDashboard,
  MessageCircle,
  Megaphone,
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

export const mainNav: NavItem[] = [
  { to: "/dashboard", icon: LayoutDashboard, labelKey: "nav.dashboard" },
  { to: "/discover", icon: MessageCircle, labelKey: "nav.discover" },
  { to: "/projects", icon: FolderOpen, labelKey: "nav.projects" },
  { to: "/credits", icon: Coins, labelKey: "nav.credits" },
]

export const futureNav: DisabledNavItem[] = [
  { icon: Megaphone, labelKey: "nav.marketing" },
]
