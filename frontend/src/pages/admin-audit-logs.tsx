import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import type { ColumnDef } from "@tanstack/react-table"

import { AdminNav } from "@/features/admin/admin-nav"
import { TableToolbar } from "@/features/admin/table-toolbar"
import { DataTable } from "@/components/data-table/data-table"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { useAdminAuditLogs } from "@/features/admin/hooks"
import type { AdminAuditLog } from "@/features/admin/api"

const AUDIT_ACTIONS = [
  "user.register",
  "user.role_promoted",
  "ticket.create",
  "ticket.reply",
  "ticket.admin_reply",
  "ticket.resolve",
] as const

const AUDIT_ENTITIES = ["user", "support_ticket"] as const

export default function AdminAuditLogsPage() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page")) || 1
  const action = searchParams.get("action") ?? ""
  const entity = searchParams.get("entity") ?? ""

  const setParam = (key: string, value: string) => {
    const next = new URLSearchParams(searchParams)
    if (value) next.set(key, value)
    else next.delete(key)
    if (key !== "page") next.delete("page")
    setSearchParams(next)
  }

  const { data: logs } = useAdminAuditLogs({ page, action, entity })

  const columns: ColumnDef<AdminAuditLog>[] = [
    {
      accessorKey: "created_at",
      header: t("admin:date"),
      cell: ({ row }) => (
        <span className="whitespace-nowrap text-xs text-muted-foreground">
          {new Date(row.original.created_at).toLocaleString()}
        </span>
      ),
    },
    {
      accessorKey: "action",
      header: t("admin:auditAction"),
      cell: ({ row }) => <Badge variant="secondary">{row.original.action}</Badge>,
    },
    {
      accessorKey: "entity",
      header: t("admin:auditEntity"),
      cell: ({ row }) => (
        <div className="text-xs">
          <p>{row.original.entity}</p>
          {row.original.entity_id && (
            <p className="truncate font-mono text-muted-foreground">{row.original.entity_id}</p>
          )}
        </div>
      ),
    },
    {
      accessorKey: "user_id",
      header: t("admin:auditUser"),
      cell: ({ row }) => (
        <span className="font-mono text-xs text-muted-foreground">{row.original.user_id || "—"}</span>
      ),
    },
    {
      id: "metadata",
      header: t("admin:auditDetails"),
      cell: ({ row }) => {
        const entries = Object.entries(row.original.metadata ?? {})
        if (entries.length === 0) return <span className="text-muted-foreground">—</span>
        return (
          <span className="text-xs text-muted-foreground">
            {entries.map(([k, v]) => `${k}=${String(v)}`).join(", ")}
          </span>
        )
      },
    },
  ]

  return (
    <div className="space-y-6">
      <AdminNav />
      <header>
        <h1 className="font-display text-2xl font-bold text-primary">{t("admin:auditTitle")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("admin:auditSubtitle")}</p>
      </header>

      <TableToolbar
        selects={[
          {
            value: action,
            onChange: (v) => setParam("action", v === "all" ? "" : v),
            placeholder: t("admin:auditAllActions"),
            options: AUDIT_ACTIONS.map((a) => ({ value: a, label: t(`admin:auditActions.${a}`) })),
          },
          {
            value: entity,
            onChange: (v) => setParam("entity", v === "all" ? "" : v),
            placeholder: t("admin:auditAllEntities"),
            options: AUDIT_ENTITIES.map((e) => ({ value: e, label: t(`admin:auditEntities.${e}`) })),
          },
        ]}
      />

      <div className="rounded-lg border bg-card">
        <DataTable columns={columns} data={logs?.data ?? []} emptyMessage={t("admin:noAuditEntries")} />
        <Pagination page={page} totalPages={logs?.pagination.totalPages ?? 1} onPageChange={(p) => setParam("page", String(p))} />
      </div>
    </div>
  )
}
