import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import type { ColumnDef } from "@tanstack/react-table"

import { AdminNav } from "@/features/admin/admin-nav"
import { TableToolbar } from "@/features/admin/table-toolbar"
import { DataTable } from "@/components/data-table/data-table"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { useAdminProjects } from "@/features/admin/hooks"
import type { AdminProject } from "@/features/admin/api"

const PROJECT_STATUSES = ["draft", "idea_selected", "generating", "content_ready", "completed", "failed"] as const

export default function AdminProjectsPage() {
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

  const { data: projects } = useAdminProjects({ page, status, search })

  const columns: ColumnDef<AdminProject>[] = [
    {
      accessorKey: "title",
      header: t("admin:titleColumn"),
      cell: ({ row }) => (
        <div className="max-w-xs">
          <p className="truncate font-medium">{row.original.title}</p>
          <p className="truncate text-xs text-muted-foreground">{row.original.user_email}</p>
        </div>
      ),
    },
    {
      accessorKey: "status",
      header: t("admin:status"),
      cell: ({ row }) => <Badge variant="outline">{t(`admin:projectStatus.${row.original.status}`)}</Badge>,
    },
    {
      accessorKey: "credits_consumed",
      header: t("admin:creditsUsed"),
      cell: ({ row }) => row.original.credits_consumed.toLocaleString("fr-FR"),
    },
    {
      accessorKey: "created_at",
      header: t("admin:date"),
      cell: ({ row }) => new Date(row.original.created_at).toLocaleDateString(),
    },
  ]

  return (
    <div className="space-y-6">
      <AdminNav />
      <header>
        <h1 className="font-display text-2xl font-bold text-primary">{t("admin:statProjects")}</h1>
      </header>

      <TableToolbar
        search={search}
        onSearch={(v) => setParam("search", v)}
        selects={[
          {
            value: status,
            onChange: (v) => setParam("status", v === "all" ? "" : v),
            placeholder: t("admin:filterAllStatuses"),
            options: PROJECT_STATUSES.map((s) => ({ value: s, label: t(`admin:projectStatus.${s}`) })),
          },
        ]}
      />

      <div className="rounded-lg border bg-card">
        <DataTable columns={columns} data={projects?.data ?? []} emptyMessage={t("admin:noItems")} />
        <Pagination page={page} totalPages={projects?.pagination.totalPages ?? 1} onPageChange={(p) => setParam("page", String(p))} />
      </div>
    </div>
  )
}
