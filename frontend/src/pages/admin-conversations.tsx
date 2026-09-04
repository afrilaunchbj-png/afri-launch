import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import type { ColumnDef } from "@tanstack/react-table"

import { AdminNav } from "@/features/admin/admin-nav"
import { TableToolbar } from "@/features/admin/table-toolbar"
import { DataTable } from "@/components/data-table/data-table"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { useAdminConversations } from "@/features/admin/hooks"
import type { AdminConversation } from "@/features/admin/api"

export default function AdminConversationsPage() {
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

  const { data: conversations } = useAdminConversations({ page, search })

  const columns: ColumnDef<AdminConversation>[] = [
    {
      accessorKey: "title",
      header: t("admin:titleColumn"),
      cell: ({ row }) => (
        <div className="max-w-xs">
          <p className="truncate font-medium">{row.original.title || "—"}</p>
          <p className="truncate text-xs text-muted-foreground">{row.original.user_email}</p>
        </div>
      ),
    },
    {
      accessorKey: "status",
      header: t("admin:status"),
      cell: ({ row }) => <Badge variant="outline">{t(`admin:conversationStatus.${row.original.status}`)}</Badge>,
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
        <h1 className="font-display text-2xl font-bold text-primary">{t("admin:statConversations")}</h1>
      </header>

      <TableToolbar search={search} onSearch={(v) => setParam("search", v)} />

      <div className="rounded-lg border bg-card">
        <DataTable columns={columns} data={conversations?.data ?? []} emptyMessage={t("admin:noItems")} />
        <Pagination page={page} totalPages={conversations?.pagination.totalPages ?? 1} onPageChange={(p) => setParam("page", String(p))} />
      </div>
    </div>
  )
}
