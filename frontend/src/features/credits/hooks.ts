import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import {
  creditsKeys,
  createCheckout,
  fetchCreditsSummary,
  fetchPayment,
  fetchPlans,
  fetchTransactions,
  syncPayment,
} from "./api"
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

export const paymentKeys = {
  plans: () => ["payments", "plans"] as const,
  payment: (id: string) => ["payments", "payment", id] as const,
}

/** usePlans : packs de crédits achetables (provider actif via config). */
export function usePlans() {
  return useQuery({ queryKey: paymentKeys.plans(), queryFn: fetchPlans })
}

/** useCreateCheckout initie un paiement et renvoie l'URL de redirection. */
export function useCreateCheckout() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (planId: string) => createCheckout(planId),
    onSettled: () => queryClient.invalidateQueries({ queryKey: creditsKeys.summary() }),
  })
}

/** usePayment suit un paiement après retour du provider. */
export function usePayment(id: string | null) {
  return useQuery({
    queryKey: paymentKeys.payment(id ?? ""),
    queryFn: () => fetchPayment(id as string),
    enabled: !!id,
  })
}

/** useSyncPayment reconfirme le statut auprès du provider. */
export function useSyncPayment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => syncPayment(id),
    onSettled: () => queryClient.invalidateQueries({ queryKey: creditsKeys.summary() }),
  })
}
