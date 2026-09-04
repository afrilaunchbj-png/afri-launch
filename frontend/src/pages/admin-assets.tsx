import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import type { ColumnDef } from "@tanstack/react-table"

import { AdminNav } from "@/features/admin/admin-nav"
import { TableToolbar } from "@/features/admin/table-toolbar"
import { DataTable } from "@/components/data-table/data-table"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { useAdminAssets } from "@/features/admin/hooks"
import type { AdminAsset } from "@/features/admin/api"

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export default function AdminAssetsPage() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page")) || 1
  const search = searchParams.get("search") ?? ""

  const setParam = (key: string, value: string) => {
    const next = new URLSearchParams(searchParams)
    if (value) next.set(key, value)
    else next.delete(key)
    if (key !== "page") next.delete("page")
    setSearchParams(next)
  }

  const { data: assets } = useAdminAssets({ page, search })

  const columns: ColumnDef<AdminAsset>[] = [
    {
      accessorKey: "filename",
      header: t("admin:filename"),
      cell: ({ row }) => (
        <div className="max-w-xs">
          <p className="truncate font-medium">{row.original.filename}</p>
          <p className="truncate text-xs text-muted-foreground">{row.original.user_email}</p>
        </div>
      ),
    },
    {
      accessorKey: "kind",
      header: t("admin:assetKind"),
      cell: ({ row }) => <Badge variant="outline">{row.original.kind}</Badge>,
    },
    {
      accessorKey: "project_title",
      header: t("admin:projectColumn"),
      cell: ({ row }) => <span className="truncate text-sm">{row.original.project_title}</span>,
    },
    {
      accessorKey: "size_bytes",
      header: t("admin:size"),
      cell: ({ row }) => formatSize(row.original.size_bytes),
    },
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
        <h1 className="font-display text-2xl font-bold text-primary">{t("admin:statAssets")}</h1>
      </header>

      <TableToolbar search={search} onSearch={(v) => setParam("search", v)} />

      <div className="rounded-lg border bg-card">
        <DataTable columns={columns} data={assets?.data ?? []} emptyMessage={t("admin:noItems")} />
        <Pagination page={page} totalPages={assets?.pagination.totalPages ?? 1} onPageChange={(p) => setParam("page", String(p))} />
      </div>
    </div>
  )
}
