import { api, type ApiList, type ApiSingle } from "@/lib/api/client"

import type {
  CreditSummary,
  CreditTransaction,
  GenerationCost,
  TransactionFilters,
} from "./types"

export const creditsKeys = {
  all: ["credits"] as const,
  summary: () => [...creditsKeys.all, "summary"] as const,
  transactions: (filters: TransactionFilters) => [...creditsKeys.all, "transactions", filters] as const,
}

export interface CreditsSummaryResponse {
  summary: CreditSummary
  costs: GenerationCost[]
}

export function fetchCreditsSummary() {
  return api.get<ApiSingle<CreditsSummaryResponse>>("/api/v1/credits").then((r) => r.data)
}

export function fetchTransactions(filters: TransactionFilters) {
  const qs = new URLSearchParams({
    page: String(filters.page),
    pageSize: String(filters.pageSize),
  })
  if (filters.type) qs.set("type", filters.type)
  if (filters.operation) qs.set("operation", filters.operation)

  return api.get<ApiList<CreditTransaction>>(`/api/v1/credits/transactions?${qs.toString()}`)
}
