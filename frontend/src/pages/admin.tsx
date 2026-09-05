import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import {
  BookOpen,
  ChevronRight,
  FolderOpen,
  Headset,
  MessageCircle,
  ReceiptText,
  Users,
  Workflow,
  type LucideIcon,
} from "lucide-react"

import { AdminNav } from "@/features/admin/admin-nav"
import { useAdminStats } from "@/features/admin/hooks"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

function StatCard({ icon: Icon, label, value, to }: { icon: LucideIcon; label: string; value: number; to: string }) {
  return (
    <Link to={to} className="group">
      <Card className="h-full transition-shadow group-hover:shadow-md">
        <CardContent className="flex items-center gap-3 p-4">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Icon className="h-5 w-5" />
          </span>
          <div className="min-w-0 flex-1">
            <p className="font-display text-xl font-bold leading-none">{value.toLocaleString("fr-FR")}</p>
            <p className="mt-1 truncate text-xs text-muted-foreground">{label}</p>
          </div>
          <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground/50 transition-transform group-hover:translate-x-0.5 group-hover:text-primary" />
        </CardContent>
      </Card>
    </Link>
  )
}

export default function AdminPage() {
  const { t } = useTranslation()

  const { data: stats } = useAdminStats()

  const runningJobs =
    (stats?.jobs_by_status.pending ?? 0) + (stats?.jobs_by_status.processing ?? 0)

  return (
    <div className="space-y-8">
      <header>
        <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">{t("admin:title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("admin:subtitle")}</p>
      </header>
      <AdminNav />

      <section className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard icon={Users} label={t("admin:statUsers")} value={stats?.users ?? 0} to="/admin/users" />
        <StatCard icon={MessageCircle} label={t("admin:statConversations")} value={stats?.conversations ?? 0} to="/admin/conversations" />
        <StatCard icon={FolderOpen} label={t("admin:statProjects")} value={stats?.projects ?? 0} to="/admin/projects" />
        <StatCard icon={BookOpen} label={t("admin:statAssets")} value={stats?.assets ?? 0} to="/admin/assets" />
        <StatCard icon={Workflow} label={t("admin:statRunningJobs")} value={runningJobs} to="/admin/jobs?status=pending" />
        <StatCard icon={ReceiptText} label={t("admin:statCredits")} value={stats?.credits_consumed ?? 0} to="/admin/transactions?type=debit" />
        <StatCard icon={Headset} label={t("admin:statOpenTickets")} value={stats?.open_tickets ?? 0} to="/admin/tickets?status=open" />
      </section>

      <section className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <CardTitle className="text-base">{t("admin:ticketsTitle")}</CardTitle>
            <Button asChild variant="ghost" size="sm">
              <Link to="/admin/tickets">
                {t("admin:viewAll")}
                <ChevronRight className="h-4 w-4" />
              </Link>
            </Button>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">{t("admin:ticketsOverviewDesc")}</p>
            <Button asChild variant="outline" className="mt-4 w-full sm:w-auto">
              <Link to="/admin/tickets?status=open">{t("admin:filterOpenTickets")}</Link>
            </Button>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <CardTitle className="text-base">{t("admin:auditTitle")}</CardTitle>
            <Button asChild variant="ghost" size="sm">
              <Link to="/admin/audit-logs">
                {t("admin:viewAll")}
                <ChevronRight className="h-4 w-4" />
              </Link>
            </Button>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">{t("admin:auditOverviewDesc")}</p>
            <Button asChild variant="outline" className="mt-4 w-full sm:w-auto">
              <Link to="/admin/audit-logs">{t("admin:auditOpen")}</Link>
            </Button>
          </CardContent>
        </Card>
      </section>
    </div>
  )
}
