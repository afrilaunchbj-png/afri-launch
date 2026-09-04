import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import type { ColumnDef } from "@tanstack/react-table"

import { AdminNav } from "@/features/admin/admin-nav"
import { TableToolbar } from "@/features/admin/table-toolbar"
import { DataTable } from "@/components/data-table/data-table"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { useAdminUsers } from "@/features/admin/hooks"
import type { AdminUser } from "@/features/admin/api"

export default function AdminUsersPage() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page")) || 1
  const role = searchParams.get("role") ?? ""
  const search = searchParams.get("search") ?? ""

  const setParam = (key: string, value: string) => {
    const next = new URLSearchParams(searchParams)
    if (value) next.set(key, value)
    else next.delete(key)
    if (key !== "page") next.delete("page")
    setSearchParams(next)
  }

  const { data: users } = useAdminUsers({ page, role, search })

  const columns: ColumnDef<AdminUser>[] = [
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

  return (
    <div className="space-y-6">
      <AdminNav />
      <header>
        <h1 className="font-display text-2xl font-bold text-primary">{t("admin:usersTitle")}</h1>
      </header>

      <TableToolbar
        search={search}
        onSearch={(v) => setParam("search", v)}
        selects={[
          {
            value: role,
            onChange: (v) => setParam("role", v === "all" ? "" : v),
            placeholder: t("admin:filterAllRoles"),
            options: [
              { value: "user", label: t("admin:roleUser") },
              { value: "superadmin", label: t("settings:superadmin") },
            ],
          },
        ]}
      />

      <div className="rounded-lg border bg-card">
        <DataTable columns={columns} data={users?.data ?? []} emptyMessage={t("admin:noUsers")} />
        <Pagination page={page} totalPages={users?.pagination.totalPages ?? 1} onPageChange={(p) => setParam("page", String(p))} />
      </div>
    </div>
  )
}
