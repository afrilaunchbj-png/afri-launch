import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { confirmIdea, fetchIdeas, fetchIdeaMessages, generateIdeas, sendIdeaMessage } from "./api"

export const ideaKeys = {
  all: ["ideas"] as const,
  list: (opportunityId?: string) => [...ideaKeys.all, "list", opportunityId ?? ""] as const,
  messages: (ideaId: string) => [...ideaKeys.all, "messages", ideaId] as const,
}

export function useIdeas(opportunityId?: string) {
  return useQuery({
    queryKey: ideaKeys.list(opportunityId),
    queryFn: () => fetchIdeas(opportunityId),
  })
}

export function useGenerateIdeas(opportunityId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => generateIdeas(opportunityId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ideaKeys.list(opportunityId) })
    },
  })
}

export function useIdeaMessages(ideaId: string) {
  return useQuery({
    queryKey: ideaKeys.messages(ideaId),
    queryFn: () => fetchIdeaMessages(ideaId),
  })
}

export function useSendIdeaMessage(ideaId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (content: string) => sendIdeaMessage(ideaId, content),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ideaKeys.messages(ideaId) })
      queryClient.invalidateQueries({ queryKey: ideaKeys.all })
    },
  })
}

export function useConfirmIdea(ideaId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => confirmIdea(ideaId),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ideaKeys.all })
    },
  })
}
