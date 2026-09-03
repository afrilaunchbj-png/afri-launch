import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { isAppError } from "@/lib/errors"

import { fetchAdminStats, fetchAdminTickets, fetchAdminUsers, resolveTicket } from "./api"

export const adminKeys = {
  all: ["admin"] as const,
  stats: () => [...adminKeys.all, "stats"] as const,
  users: (page: number) => [...adminKeys.all, "users", page] as const,
  tickets: (page: number) => [...adminKeys.all, "tickets", page] as const,
}

export function useAdminStats() {
  return useQuery({ queryKey: adminKeys.stats(), queryFn: fetchAdminStats, refetchInterval: 15000 })
}

export function useAdminUsers(page: number) {
  return useQuery({ queryKey: adminKeys.users(page), queryFn: () => fetchAdminUsers(page) })
}

export function useAdminTickets(page: number) {
  return useQuery({ queryKey: adminKeys.tickets(page), queryFn: () => fetchAdminTickets(page) })
}

export function useResolveTicket(page: number) {
  const queryClient = useQueryClient()
  const { t } = useTranslation()
  return useMutation({
    mutationFn: resolveTicket,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.tickets(page) })
      queryClient.invalidateQueries({ queryKey: adminKeys.stats() })
      toast.success(t("admin:ticketResolved"))
    },
    onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
  })
}
