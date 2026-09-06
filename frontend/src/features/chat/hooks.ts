import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

import { useTranslation } from "react-i18next"

import { isAppError } from "@/lib/errors"

import { confirmIdea, createConversation, fetchConversation, fetchConversations, sendChatMessage, updateChatIdea } from "./api"

export const chatKeys = {
  all: ["chat"] as const,
  list: () => [...chatKeys.all, "list"] as const,
  detail: (id: string) => [...chatKeys.all, "detail", id] as const,
}

export function useConversations() {
  return useQuery({ queryKey: chatKeys.list(), queryFn: fetchConversations })
}

export function useConversation(id: string | undefined) {
  return useQuery({
    queryKey: chatKeys.detail(id ?? ""),
    queryFn: () => fetchConversation(id as string),
    enabled: !!id,
  })
}

export function useUpdateChatIdea(conversationId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ ideaId, title, subtitle, confirm = true }: { ideaId: string; title?: string; subtitle?: string; confirm?: boolean }) =>
      updateChatIdea(conversationId, ideaId, { title, subtitle, confirm }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: chatKeys.detail(conversationId) }),
  })
}

export function useCreateConversation() {
  return useMutation({ mutationFn: createConversation })
}

export function useSendChatMessage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ conversationId, content }: { conversationId: string; content: string }) =>
      sendChatMessage(conversationId, content),
    onSuccess: (_data, vars) => {
      queryClient.invalidateQueries({ queryKey: chatKeys.detail(vars.conversationId) })
    },
  })
}

export function useConfirmIdea(conversationId: string | undefined) {
  const queryClient = useQueryClient()
  const { t } = useTranslation()
  return useMutation({
    mutationFn: confirmIdea,
    onSuccess: () => {
      if (conversationId) {
        queryClient.invalidateQueries({ queryKey: chatKeys.detail(conversationId) })
      }
    },
    onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
  })
}
