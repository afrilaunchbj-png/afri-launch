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

// ---------- Paiements (checkout, ADR-018) ----------

export interface CreditPlan {
  id: string
  name: string
  credits: number
  price_minor: number
  currency: string
}

export interface PlansResponse {
  plans: CreditPlan[]
  provider: string
  enabled: boolean
}

export interface Payment {
  id: string
  plan_id?: string
  amount_minor: number
  currency: string
  provider: string
  status: "pending" | "succeeded" | "failed" | "refunded"
  checkout_url?: string
}

export function fetchPlans() {
  return api.get<ApiSingle<PlansResponse>>("/api/v1/payments/plans").then((r) => r.data)
}

export function createCheckout(planId: string) {
  return api
    .post<ApiSingle<{ payment: Payment; redirect_url: string }>>("/api/v1/payments/checkout", {
      plan_id: planId,
    })
    .then((r) => r.data)
}

export function fetchPayment(id: string) {
  return api.get<ApiSingle<Payment>>(`/api/v1/payments/${id}`).then((r) => r.data)
}

export function syncPayment(id: string) {
  return api.post<ApiSingle<Payment>>(`/api/v1/payments/${id}/sync`).then((r) => r.data)
}
