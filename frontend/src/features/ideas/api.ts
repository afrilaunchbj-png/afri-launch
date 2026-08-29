import { api, type ApiSingle } from "@/lib/api/client"

import type { Job } from "@/features/generation/api"

export interface Idea {
  id: string
  opportunity_id?: string | null
  title: string
  subtitle: string
  audience: string
  problem: string
  promise: string
  format: string
  estimated_price: string
  difficulty: string
  market_evidence: string
  why_now: string
  competitive_angle: string
  is_selected: boolean
}

export function generateIdeas(opportunityId: string) {
  return api.post<ApiSingle<Job>>(`/api/v1/opportunities/${opportunityId}/ideas`).then((r) => r.data)
}

export function fetchIdeas(opportunityId?: string) {
  const path = opportunityId ? `/api/v1/opportunities/${opportunityId}/ideas` : "/api/v1/ideas"
  return api.get<ApiSingle<Idea[]>>(path).then((r) => r.data)
}
