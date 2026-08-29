import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { fetchIdeas, generateIdeas } from "./api"

export const ideaKeys = {
  all: ["ideas"] as const,
  list: (opportunityId?: string) => [...ideaKeys.all, "list", opportunityId ?? ""] as const,
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
