import { Link, useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import type { ColumnDef } from "@tanstack/react-table"

import { AdminNav } from "@/features/admin/admin-nav"
import { TableToolbar } from "@/features/admin/table-toolbar"
import { DataTable } from "@/components/data-table/data-table"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { useAdminTickets, useAdminResolveTicket } from "@/features/admin/hooks"
import type { AdminTicket } from "@/features/admin/api"

function useTicketListParams() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page")) || 1
  const status = searchParams.get("status") ?? ""
  const search = searchParams.get("search") ?? ""

  const setParam = (key: string, value: string) => {
    const next = new URLSearchParams(searchParams)
    if (value) next.set(key, value)
    else next.delete(key)
    if (key !== "page") next.delete("page")
    setSearchParams(next)
  }
  return { page, status, search, setParam }
}

export default function AdminTicketsPage() {
  const { t } = useTranslation()
  const { page, status, search, setParam } = useTicketListParams()

  const listParams = { page, status, search }
  const { data: tickets } = useAdminTickets(listParams)
  const resolve = useAdminResolveTicket()

  const columns: ColumnDef<AdminTicket>[] = [
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
      cell: ({ row }) => new Date(row.original.created_at).toLocaleString(),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <div className="flex justify-end gap-2">
          {row.original.status === "open" && (
            <Button
              size="sm"
              variant="outline"
              disabled={resolve.isPending}
              onClick={() => resolve.mutate(row.original.id)}
            >
              {t("admin:resolve")}
            </Button>
          )}
          <Button size="sm" asChild>
            <Link to={`/admin/tickets/${row.original.id}`}>{t("admin:openTicket")}</Link>
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <AdminNav />
      <header>
        <h1 className="font-display text-2xl font-bold text-primary">{t("admin:ticketsTitle")}</h1>
      </header>

      <TableToolbar
        search={search}
        onSearch={(v) => setParam("search", v)}
        selects={[
          {
            value: status,
            onChange: (v) => setParam("status", v === "all" ? "" : v),
            placeholder: t("admin:filterAllStatuses"),
            options: [
              { value: "open", label: t("support:status.open") },
              { value: "resolved", label: t("support:status.resolved") },
            ],
          },
        ]}
      />

      <div className="rounded-lg border bg-card">
        <DataTable columns={columns} data={tickets?.data ?? []} emptyMessage={t("admin:noTickets")} />
        <Pagination page={page} totalPages={tickets?.pagination.totalPages ?? 1} onPageChange={(p) => setParam("page", String(p))} />
      </div>
    </div>
  )
}
