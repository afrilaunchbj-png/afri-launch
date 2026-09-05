import { useState } from "react"
import { useTranslation } from "react-i18next"
import type { TFunction } from "i18next"
import type { ColumnDef } from "@tanstack/react-table"
import { ArrowDownCircle, ArrowUpCircle, Coins, Receipt } from "lucide-react"

import { PlansPanel } from "@/features/credits/plans-panel"
import { DataTable } from "@/components/data-table/data-table"
import { EmptyState } from "@/components/states/empty-state"
import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Pagination } from "@/components/pagination"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useCreditsSummary, useTransactions } from "@/features/credits/hooks"
import type { CreditTransaction } from "@/features/credits/types"
import { cn } from "@/lib/utils"

const PAGE_SIZE = 10

type TypeFilter = "" | "credit" | "debit"

function operationLabel(t: TFunction, operation: string) {
  const key = `credits:operations.${operation}`
  const translated = t(key)
  return translated === key ? operation : translated
}

function statusVariant(status: string): "success" | "warning" | "destructive" | "outline" {
  switch (status) {
    case "completed":
      return "success"
    case "pending":
      return "warning"
    case "failed":
      return "destructive"
    default:
      return "outline"
  }
}

function formatDate(value: string, lang: string) {
  return new Intl.DateTimeFormat(lang, { dateStyle: "medium" }).format(new Date(value))
}

function formatTime(value: string, lang: string) {
  return new Intl.DateTimeFormat(lang, { timeStyle: "short" }).format(new Date(value))
}

function buildColumns(t: TFunction, lang: string): ColumnDef<CreditTransaction, unknown>[] {
  return [
    {
      accessorKey: "created_at",
      header: t("credits:date"),
      cell: ({ row }) => (
        <div className="flex flex-col">
          <span className="font-medium">{formatDate(row.original.created_at, lang)}</span>
          <span className="text-xs text-muted-foreground">{formatTime(row.original.created_at, lang)}</span>
        </div>
      ),
    },
    {
      accessorKey: "operation",
      header: t("credits:activity"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-muted">
            {row.original.type === "credit" ? (
              <ArrowUpCircle className="h-4 w-4 text-success" />
            ) : (
              <ArrowDownCircle className="h-4 w-4 text-destructive" />
            )}
          </div>
          <span className="font-medium">{operationLabel(t, row.original.operation)}</span>
        </div>
      ),
    },
    {
      accessorKey: "status",
      header: t("credits:statusLabel"),
      cell: ({ row }) => (
        <Badge variant={statusVariant(row.original.status)}>{t(`credits:status.${row.original.status}`)}</Badge>
      ),
    },
    {
      accessorKey: "amount",
      header: t("credits:amount"),
      cell: ({ row }) => {
        const isCredit = row.original.type === "credit"
        return (
          <span className={cn("font-display font-semibold", isCredit ? "text-success" : "text-destructive")}>
            {isCredit ? "+" : "-"}
            {row.original.amount}
          </span>
        )
      },
    },
  ]
}

export default function CreditsPage() {
  const { t, i18n } = useTranslation()
  const [type, setType] = useState<TypeFilter>("")
  const [page, setPage] = useState(1)

  const { data: summary, isLoading: summaryLoading } = useCreditsSummary()
  const { data, isLoading, isError, refetch } = useTransactions({
    type,
    operation: "",
    page,
    pageSize: PAGE_SIZE,
  })

  const columns = buildColumns(t, i18n.language)

  return (
    <div className="space-y-8">
      <header>
        <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">{t("credits:title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("credits:subtitle")}</p>
      </header>

      <PlansPanel />

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="relative overflow-hidden">
          <CardContent className="p-6">
            <div className="mb-4 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-accent text-accent-foreground">
                <Coins className="h-5 w-5" />
              </div>
              <h3 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
                {t("credits:balance")}
              </h3>
            </div>
            {summaryLoading ? (
              <Skeleton className="h-9 w-28" />
            ) : (
              <div className="flex items-baseline gap-2">
                <span className="font-display text-4xl font-bold text-primary">
                  {summary?.summary.available ?? 0}
                </span>
                <span className="text-muted-foreground">{t("credits:label")}</span>
              </div>
            )}
            <Button size="touch" className="mt-6 w-full" disabled>
              {t("credits:recharge")}
            </Button>
          </CardContent>
        </Card>

        <div className="grid gap-6 sm:grid-cols-2 lg:col-span-2">
          <Card>
            <CardContent className="flex h-full flex-col justify-between p-6">
              <div className="mb-2 flex items-center gap-3">
                <ArrowUpCircle className="h-5 w-5 text-success" />
                <h4 className="text-sm font-semibold text-muted-foreground">{t("credits:addedMonth")}</h4>
              </div>
              {summaryLoading ? (
                <Skeleton className="h-8 w-16" />
              ) : (
                <div className="text-2xl font-bold text-success">+{summary?.summary.added_month ?? 0}</div>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardContent className="flex h-full flex-col justify-between p-6">
              <div className="mb-2 flex items-center gap-3">
                <ArrowDownCircle className="h-5 w-5 text-destructive" />
                <h4 className="text-sm font-semibold text-muted-foreground">{t("credits:usedMonth")}</h4>
              </div>
              {summaryLoading ? (
                <Skeleton className="h-8 w-16" />
              ) : (
                <div className="text-2xl font-bold text-destructive">-{summary?.summary.used_month ?? 0}</div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      <section>
        <div className="mb-4 flex flex-wrap items-center gap-2">
          {(["", "credit", "debit"] as const).map((value) => (
            <Button
              key={value || "all"}
              variant={type === value ? "default" : "outline"}
              size="sm"
              className="rounded-full"
              onClick={() => {
                setType(value)
                setPage(1)
              }}
            >
              {t(value === "" ? "credits:filter.all" : `credits:filter.${value}`)}
            </Button>
          ))}
        </div>

        {isLoading ? (
          <LoadingState label={t("common.loading")} />
        ) : isError ? (
          <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />
        ) : !data || data.data.length === 0 ? (
          <EmptyState
            icon={Receipt}
            title={t("credits:empty")}
            description={t("credits:emptyDesc")}
          />
        ) : (
          <Card className="overflow-hidden">
            <div className="hidden sm:block">
              <DataTable columns={columns} data={data.data} emptyMessage={t("credits:empty")} />
            </div>
            <div className="divide-y sm:hidden">
              {data.data.map((tx) => (
                <div key={tx.id} className="flex items-center gap-3 p-4">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-muted">
                    {tx.type === "credit" ? (
                      <ArrowUpCircle className="h-5 w-5 text-success" />
                    ) : (
                      <ArrowDownCircle className="h-5 w-5 text-destructive" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">{operationLabel(t, tx.operation)}</p>
                    <p className="text-xs text-muted-foreground">
                      {formatDate(tx.created_at, i18n.language)} · {formatTime(tx.created_at, i18n.language)}
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-1">
                    <span className={cn("font-semibold", tx.type === "credit" ? "text-success" : "text-destructive")}>
                      {tx.type === "credit" ? "+" : "-"}
                      {tx.amount}
                    </span>
                    <Badge variant={statusVariant(tx.status)} className="text-[10px]">
                      {t(`credits:status.${tx.status}`)}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
            <Pagination
              page={page}
              totalPages={data.pagination.totalPages}
              onPageChange={(p) => setPage(p)}
            />
          </Card>
        )}
      </section>
    </div>
  )
}
