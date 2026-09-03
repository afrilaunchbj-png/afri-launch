import { useState } from "react"
import { Navigate } from "react-router"
import { useTranslation } from "react-i18next"
import type { ColumnDef } from "@tanstack/react-table"
import { BookOpen, CheckCircle2, FolderOpen, LifeBuoy, MessageCircle, Users } from "lucide-react"

import { DataTable } from "@/components/data-table/data-table"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { useMe } from "@/features/auth/hooks"
import {
  useAdminStats,
  useAdminTickets,
  useAdminUsers,
  useResolveTicket,
} from "@/features/admin/hooks"
import type { AdminTicket, AdminUser } from "@/features/admin/api"

function StatCard({ icon: Icon, label, value }: { icon: typeof Users; label: string; value: number }) {
  return (
    <Card>
      <CardContent className="flex items-center gap-3 p-4">
        <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
          <Icon className="h-5 w-5" />
        </span>
        <div>
          <p className="font-display text-xl font-bold leading-none">{value.toLocaleString("fr-FR")}</p>
          <p className="mt-1 text-xs text-muted-foreground">{label}</p>
        </div>
      </CardContent>
    </Card>
  )
}

export default function AdminPage() {
  const { t } = useTranslation()
  const { data: me, isLoading: meLoading } = useMe()
  const [userPage, setUserPage] = useState(1)
  const [ticketPage, setTicketPage] = useState(1)

  const { data: stats } = useAdminStats()
  const { data: users } = useAdminUsers(userPage)
  const { data: tickets } = useAdminTickets(ticketPage)
  const resolve = useResolveTicket(ticketPage)

  if (meLoading) return null
  // Garde côté client (le backend revalide le rôle à chaque appel).
  if (me?.role !== "superadmin") return <Navigate to="/dashboard" replace />

  const runningJobs =
    (stats?.jobs_by_status.pending ?? 0) + (stats?.jobs_by_status.processing ?? 0)

  const userColumns: ColumnDef<AdminUser>[] = [
    {
      accessorKey: "full_name",
      header: t("admin:name"),
      cell: ({ row }) => <span className="font-medium">{row.original.full_name || "—"}</span>,
    },
    { accessorKey: "email", header: t("admin:email") },
    {
      accessorKey: "role",
      header: t("admin:role"),
      cell: ({ row }) =>
        row.original.role === "superadmin" ? (
          <Badge variant="secondary">{t("settings:superadmin")}</Badge>
        ) : (
          <Badge variant="outline">{t("admin:roleUser")}</Badge>
        ),
    },
    {
      accessorKey: "created_at",
      header: t("admin:signupDate"),
      cell: ({ row }) => new Date(row.original.created_at).toLocaleDateString(),
    },
  ]

  const ticketColumns: ColumnDef<AdminTicket>[] = [
    {
      accessorKey: "subject",
      header: t("support:subject"),
      cell: ({ row }) => (
        <div className="max-w-xs">
          <p className="truncate font-medium">{row.original.subject}</p>
          <p className="truncate text-xs text-muted-foreground">{row.original.message}</p>
        </div>
      ),
    },
    {
      accessorKey: "user_email",
      header: t("admin:requester"),
      cell: ({ row }) => (
        <div className="text-xs">
          <p className="font-medium">{row.original.user_name || "—"}</p>
          <p className="text-muted-foreground">{row.original.user_email}</p>
        </div>
      ),
    },
    {
      accessorKey: "status",
      header: t("admin:status"),
      cell: ({ row }) => (
        <Badge variant={row.original.status === "resolved" ? "success" : "secondary"}>
          {t(`support:status.${row.original.status}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "created_at",
      header: t("admin:date"),
      cell: ({ row }) => new Date(row.original.created_at).toLocaleDateString(),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) =>
        row.original.status === "open" ? (
          <Button
            size="sm"
            variant="outline"
            disabled={resolve.isPending}
            onClick={() => resolve.mutate(row.original.id)}
          >
            <CheckCircle2 className="h-4 w-4" />
            {t("admin:resolve")}
          </Button>
        ) : null,
    },
  ]

  return (
    <div className="space-y-8">
      <header>
        <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">{t("admin:title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("admin:subtitle")}</p>
      </header>

      <section className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard icon={Users} label={t("admin:statUsers")} value={stats?.users ?? 0} />
        <StatCard icon={MessageCircle} label={t("admin:statConversations")} value={stats?.conversations ?? 0} />
        <StatCard icon={FolderOpen} label={t("admin:statProjects")} value={stats?.projects ?? 0} />
        <StatCard icon={BookOpen} label={t("admin:statAssets")} value={stats?.assets ?? 0} />
        <StatCard icon={BookOpen} label={t("admin:statRunningJobs")} value={runningJobs} />
        <StatCard icon={BookOpen} label={t("admin:statCredits")} value={stats?.credits_consumed ?? 0} />
        <StatCard icon={LifeBuoy} label={t("admin:statOpenTickets")} value={stats?.open_tickets ?? 0} />
      </section>

      <section>
        <h2 className="mb-3 font-display text-lg font-semibold text-primary">{t("admin:ticketsTitle")}</h2>
        <div className="rounded-lg border bg-card">
          <DataTable columns={ticketColumns} data={tickets?.data ?? []} emptyMessage={t("admin:noTickets")} />
          <Pagination page={ticketPage} totalPages={tickets?.pagination.totalPages ?? 1} onPageChange={setTicketPage} />
        </div>
      </section>

      <section>
        <h2 className="mb-3 font-display text-lg font-semibold text-primary">{t("admin:usersTitle")}</h2>
        <div className="rounded-lg border bg-card">
          <DataTable columns={userColumns} data={users?.data ?? []} emptyMessage={t("admin:noUsers")} />
          <Pagination page={userPage} totalPages={users?.pagination.totalPages ?? 1} onPageChange={setUserPage} />
        </div>
      </section>
    </div>
  )
}
