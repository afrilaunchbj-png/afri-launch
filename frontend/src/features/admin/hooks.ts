import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { isAppError } from "@/lib/errors"

import {
  fetchAdminAssets,
  fetchAdminAuditLogs,
  fetchAdminConversations,
  fetchAdminCreditTransactions,
  fetchAdminJobs,
  fetchAdminProjects,
  fetchAdminStats,
  fetchAdminTicketDetail,
  fetchAdminTickets,
  fetchAdminUsers,
  replyAdminTicket,
  resolveTicket,
  type AdminListParams,
  type AuditListParams,
} from "./api"

export const adminKeys = {
  all: ["admin"] as const,
  stats: () => [...adminKeys.all, "stats"] as const,
  users: (params: AdminListParams) => [...adminKeys.all, "users", params] as const,
  tickets: (params: AdminListParams) => [...adminKeys.all, "tickets", params] as const,
  ticket: (id: string) => [...adminKeys.all, "ticket", id] as const,
  projects: (params: AdminListParams) => [...adminKeys.all, "projects", params] as const,
  conversations: (params: AdminListParams) => [...adminKeys.all, "conversations", params] as const,
  assets: (params: AdminListParams) => [...adminKeys.all, "assets", params] as const,
  jobs: (params: AdminListParams) => [...adminKeys.all, "jobs", params] as const,
  transactions: (params: AdminListParams) => [...adminKeys.all, "transactions", params] as const,
  auditLogs: (params: AuditListParams) => [...adminKeys.all, "audit-logs", params] as const,
}

export function useAdminStats() {
  return useQuery({ queryKey: adminKeys.stats(), queryFn: fetchAdminStats, refetchInterval: 15000 })
}

export function useAdminUsers(params: AdminListParams) {
  return useQuery({ queryKey: adminKeys.users(params), queryFn: () => fetchAdminUsers(params) })
}

export function useAdminTickets(params: AdminListParams) {
  return useQuery({ queryKey: adminKeys.tickets(params), queryFn: () => fetchAdminTickets(params) })
}

export function useAdminProjects(params: AdminListParams) {
  return useQuery({ queryKey: adminKeys.projects(params), queryFn: () => fetchAdminProjects(params) })
}

export function useAdminConversations(params: AdminListParams) {
  return useQuery({ queryKey: adminKeys.conversations(params), queryFn: () => fetchAdminConversations(params) })
}

export function useAdminAssets(params: AdminListParams) {
  return useQuery({ queryKey: adminKeys.assets(params), queryFn: () => fetchAdminAssets(params) })
}

export function useAdminJobs(params: AdminListParams) {
  return useQuery({ queryKey: adminKeys.jobs(params), queryFn: () => fetchAdminJobs(params) })
}

export function useAdminCreditTransactions(params: AdminListParams) {
  return useQuery({ queryKey: adminKeys.transactions(params), queryFn: () => fetchAdminCreditTransactions(params) })
}

export function useAdminAuditLogs(params: AuditListParams) {
  return useQuery({ queryKey: adminKeys.auditLogs(params), queryFn: () => fetchAdminAuditLogs(params) })
}

export function useAdminTicketDetail(id: string) {
  return useQuery({ queryKey: adminKeys.ticket(id), queryFn: () => fetchAdminTicketDetail(id) })
}

export function useAdminReplyTicket(id: string) {
  const queryClient = useQueryClient()
  const { t } = useTranslation()
  return useMutation({
    mutationFn: (content: string) => replyAdminTicket(id, content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.ticket(id) })
      toast.success(t("admin:replySent"))
    },
    onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
  })
}

export function useAdminResolveTicket(ticketId?: string) {
  const queryClient = useQueryClient()
  const { t } = useTranslation()
  return useMutation({
    mutationFn: resolveTicket,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [...adminKeys.all, "tickets"] })
      queryClient.invalidateQueries({ queryKey: adminKeys.stats() })
      if (ticketId) queryClient.invalidateQueries({ queryKey: adminKeys.ticket(ticketId) })
      toast.success(t("admin:ticketResolved"))
    },
    onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
  })
}
