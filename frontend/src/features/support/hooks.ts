import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { isAppError } from "@/lib/errors"

import {
  createTicket,
  fetchMyTickets,
  fetchTicketDetail,
  replyTicket,
  uploadAttachments,
} from "./api"

export const supportKeys = {
  all: ["support"] as const,
  mine: () => [...supportKeys.all, "mine"] as const,
  ticket: (id: string) => [...supportKeys.all, "ticket", id] as const,
}

export function useMyTickets() {
  return useQuery({ queryKey: supportKeys.mine(), queryFn: fetchMyTickets })
}

export function useTicketDetail(id: string) {
  return useQuery({ queryKey: supportKeys.ticket(id), queryFn: () => fetchTicketDetail(id) })
}

export function useCreateTicket() {
  const queryClient = useQueryClient()
  const { t } = useTranslation()
  return useMutation({
    mutationFn: async ({ subject, message, files }: { subject: string; message: string; files: File[] }) => {
      const attachmentIds = await uploadAttachments(files)
      return createTicket(subject, message, attachmentIds)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: supportKeys.mine() })
      toast.success(t("support:created"))
    },
    onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
  })
}

export function useReplyTicket(id: string) {
  const queryClient = useQueryClient()
  const { t } = useTranslation()
  return useMutation({
    mutationFn: async ({ content, files }: { content: string; files: File[] }) => {
      const attachmentIds = await uploadAttachments(files)
      return replyTicket(id, content, attachmentIds)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: supportKeys.ticket(id) })
      queryClient.invalidateQueries({ queryKey: supportKeys.mine() })
      toast.success(t("support:replySent"))
    },
    onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
  })
}
