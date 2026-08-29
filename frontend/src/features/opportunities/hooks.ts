import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import {
  fetchFacets,
  fetchOpportunities,
  opportunityKeys,
  saveOpportunity,
  unsaveOpportunity,
} from "./api"
import type { OpportunityFilters } from "./types"

export function useOpportunities(filters: OpportunityFilters) {
  return useQuery({
    queryKey: opportunityKeys.list(filters),
    queryFn: () => fetchOpportunities(filters),
    placeholderData: keepPreviousData,
  })
}

export function useFacets() {
  return useQuery({
    queryKey: opportunityKeys.facets(),
    queryFn: fetchFacets,
    staleTime: Infinity,
  })
}

export function useToggleSave(id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (isSaved: boolean) => (isSaved ? unsaveOpportunity(id) : saveOpportunity(id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: opportunityKeys.all })
    },
  })
}
