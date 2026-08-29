import {
  Coins,
  FilePen,
  LayoutDashboard,
  Lightbulb,
  LineChart,
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
  { to: "/opportunities", icon: LineChart, labelKey: "nav.opportunities" },
  { to: "/credits", icon: Coins, labelKey: "nav.credits" },
]

export const futureNav: DisabledNavItem[] = [
  { icon: Lightbulb, labelKey: "nav.ideation" },
  { icon: FilePen, labelKey: "nav.creation" },
  { icon: Megaphone, labelKey: "nav.marketing" },
]
