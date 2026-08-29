import { keepPreviousData, useQuery } from "@tanstack/react-query"

import { creditsKeys, fetchCreditsSummary, fetchTransactions } from "./api"
import type { TransactionFilters } from "./types"

export function useCreditsSummary() {
  return useQuery({
    queryKey: creditsKeys.summary(),
    queryFn: fetchCreditsSummary,
    staleTime: 30 * 1000,
  })
}

export function useTransactions(filters: TransactionFilters) {
  return useQuery({
    queryKey: creditsKeys.transactions(filters),
    queryFn: () => fetchTransactions(filters),
    placeholderData: keepPreviousData,
  })
}
