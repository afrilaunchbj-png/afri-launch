import { NavLink } from "react-router"
import { useTranslation } from "react-i18next"

const sections = [
  { to: "/admin", labelKey: "admin:navOverview", end: true },
  { to: "/admin/users", labelKey: "admin:usersTitle" },
  { to: "/admin/tickets", labelKey: "admin:ticketsTitle" },
  { to: "/admin/projects", labelKey: "admin:statProjects" },
  { to: "/admin/conversations", labelKey: "admin:statConversations" },
  { to: "/admin/assets", labelKey: "admin:statAssets" },
  { to: "/admin/jobs", labelKey: "admin:jobsTitle" },
  { to: "/admin/transactions", labelKey: "admin:transactionsTitle" },
  { to: "/admin/audit-logs", labelKey: "admin:auditTitle" },
]

/** Navigation secondaire de l'administration (affichée en haut de chaque page). */
export function AdminNav() {
  const { t } = useTranslation()
  return (
    <nav className="flex gap-1 overflow-x-auto border-b pb-2">
      {sections.map((s) => (
        <NavLink
          key={s.to}
          to={s.to}
          end={s.end}
          className={({ isActive }) =>
            [
              "whitespace-nowrap rounded-lg px-3 py-1.5 text-sm font-medium transition-colors",
              isActive
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:bg-muted",
            ].join(" ")
          }
        >
          {t(s.labelKey)}
        </NavLink>
      ))}
    </nav>
  )
}
