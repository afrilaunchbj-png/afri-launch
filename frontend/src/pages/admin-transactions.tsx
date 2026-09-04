import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import type { ColumnDef } from "@tanstack/react-table"

import { AdminNav } from "@/features/admin/admin-nav"
import { TableToolbar } from "@/features/admin/table-toolbar"
import { DataTable } from "@/components/data-table/data-table"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { useAdminCreditTransactions } from "@/features/admin/hooks"
import type { AdminCreditTransaction } from "@/features/admin/api"

export default function AdminTransactionsPage() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page")) || 1
  const type = searchParams.get("type") ?? ""
  const search = searchParams.get("search") ?? ""

  const setParam = (key: string, value: string) => {
    const next = new URLSearchParams(searchParams)
    if (value) next.set(key, value)
    else next.delete(key)
    if (key !== "page") next.delete("page")
    setSearchParams(next)
  }

  const { data: transactions } = useAdminCreditTransactions({ page, status: type, search })

  const columns: ColumnDef<AdminCreditTransaction>[] = [
    {
      accessorKey: "operation",
      header: t("admin:operation"),
      cell: ({ row }) => (
        <div className="max-w-xs">
          <p className="truncate font-medium">{row.original.operation}</p>
          <p className="truncate text-xs text-muted-foreground">{row.original.user_email}</p>
        </div>
      ),
    },
    {
      accessorKey: "type",
      header: t("admin:type"),
      cell: ({ row }) => (
        <Badge variant={row.original.type === "credit" ? "success" : "secondary"}>
          {t(`admin:transactionType.${row.original.type}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "amount",
      header: t("admin:amount"),
      cell: ({ row }) => row.original.amount.toLocaleString("fr-FR"),
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
        <h1 className="font-display text-2xl font-bold text-primary">{t("admin:transactionsTitle")}</h1>
      </header>

      <TableToolbar
        search={search}
        onSearch={(v) => setParam("search", v)}
        selects={[
          {
            value: type,
            onChange: (v) => setParam("type", v === "all" ? "" : v),
            placeholder: t("admin:filterAllTypes"),
            options: [
              { value: "credit", label: t("admin:transactionType.credit") },
              { value: "debit", label: t("admin:transactionType.debit") },
            ],
          },
        ]}
      />

      <div className="rounded-lg border bg-card">
        <DataTable columns={columns} data={transactions?.data ?? []} emptyMessage={t("admin:noItems")} />
        <Pagination page={page} totalPages={transactions?.pagination.totalPages ?? 1} onPageChange={(p) => setParam("page", String(p))} />
      </div>
    </div>
  )
}
