import { api, type ApiList, type ApiSingle } from "@/lib/api/client"

import type { Opportunity, OpportunityFacets, OpportunityFilters } from "./types"

export const opportunityKeys = {
  all: ["opportunities"] as const,
  list: (filters: OpportunityFilters) => [...opportunityKeys.all, "list", filters] as const,
  facets: () => [...opportunityKeys.all, "facets"] as const,
}

export function fetchOpportunities(filters: OpportunityFilters) {
  const qs = new URLSearchParams({
    page: String(filters.page),
    pageSize: String(filters.pageSize),
  })
  if (filters.country) qs.set("country", filters.country)
  if (filters.sector) qs.set("sector", filters.sector)
  if (filters.difficulty) qs.set("difficulty", filters.difficulty)
  if (filters.q) qs.set("q", filters.q)

  return api.get<ApiList<Opportunity>>(`/api/v1/opportunities?${qs.toString()}`)
}

export function fetchFacets() {
  return api.get<ApiSingle<OpportunityFacets>>("/api/v1/opportunities/filters").then((r) => r.data)
}

export function saveOpportunity(id: string) {
  return api.post<void>(`/api/v1/opportunities/${id}/save`)
}

export function unsaveOpportunity(id: string) {
  return api.delete<void>(`/api/v1/opportunities/${id}/save`)
}
