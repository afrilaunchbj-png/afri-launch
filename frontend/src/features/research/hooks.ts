import { useMutation, useQueryClient } from "@tanstack/react-query"

import { startResearch, type StartResearchInput } from "./api"
import { opportunityKeys } from "@/features/opportunities/api"

export function useStartResearch() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: StartResearchInput) => startResearch(input),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: opportunityKeys.all })
    },
  })
}
