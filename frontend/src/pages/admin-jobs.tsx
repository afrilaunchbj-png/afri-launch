import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import type { ColumnDef } from "@tanstack/react-table"

import { AdminNav } from "@/features/admin/admin-nav"
import { TableToolbar } from "@/features/admin/table-toolbar"
import { DataTable } from "@/components/data-table/data-table"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { useAdminJobs } from "@/features/admin/hooks"
import type { AdminJob } from "@/features/admin/api"

const JOB_STATUSES = ["pending", "processing", "completed", "failed"] as const

export default function AdminJobsPage() {
  const { t } = useTranslation()
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

  const { data: jobs } = useAdminJobs({ page, status, search })

  const columns: ColumnDef<AdminJob>[] = [
    {
      accessorKey: "kind",
      header: t("admin:jobKind"),
      cell: ({ row }) => (
        <div className="max-w-xs">
          <p className="truncate font-medium">{row.original.kind}</p>
          <p className="truncate text-xs text-muted-foreground">{row.original.user_email}</p>
        </div>
      ),
    },
    {
      accessorKey: "status",
      header: t("admin:status"),
      cell: ({ row }) => (
        <Badge variant={row.original.status === "failed" ? "destructive" : "outline"}>
          {t(`admin:jobStatus.${row.original.status}`)}
        </Badge>
      ),
    },
    { accessorKey: "cost", header: t("admin:cost") },
    {
      accessorKey: "created_at",
      header: t("admin:date"),
      cell: ({ row }) => new Date(row.original.created_at).toLocaleString(),
    },
  ]

  return (
    <div className="space-y-6">
      <AdminNav />
      <header>
        <h1 className="font-display text-2xl font-bold text-primary">{t("admin:jobsTitle")}</h1>
      </header>

      <TableToolbar
        search={search}
        onSearch={(v) => setParam("search", v)}
        selects={[
          {
            value: status,
            onChange: (v) => setParam("status", v === "all" ? "" : v),
            placeholder: t("admin:filterAllStatuses"),
            options: JOB_STATUSES.map((s) => ({ value: s, label: t(`admin:jobStatus.${s}`) })),
          },
        ]}
      />

      <div className="rounded-lg border bg-card">
        <DataTable columns={columns} data={jobs?.data ?? []} emptyMessage={t("admin:noItems")} />
        <Pagination page={page} totalPages={jobs?.pagination.totalPages ?? 1} onPageChange={(p) => setParam("page", String(p))} />
      </div>
    </div>
  )
}
