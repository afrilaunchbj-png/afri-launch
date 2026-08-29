import { api, type ApiSingle } from "@/lib/api/client"

import type { Job } from "@/features/generation/api"

export interface Idea {
  id: string
  opportunity_id?: string | null
  title: string
  hook: string
  explanation: string
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
  status: string
}

export interface IdeaMessage {
  id: string
  role: "user" | "assistant"
  content: string
  created_at: string
}

export function generateIdeas(opportunityId: string) {
  return api.post<ApiSingle<Job>>(`/api/v1/opportunities/${opportunityId}/ideas`).then((r) => r.data)
}

export function fetchIdeas(opportunityId?: string) {
  const path = opportunityId ? `/api/v1/opportunities/${opportunityId}/ideas` : "/api/v1/ideas"
  return api.get<ApiSingle<Idea[]>>(path).then((r) => r.data)
}

export function fetchIdeaMessages(ideaId: string) {
  return api.get<ApiSingle<IdeaMessage[]>>(`/api/v1/ideas/${ideaId}/messages`).then((r) => r.data)
}

export function sendIdeaMessage(ideaId: string, content: string) {
  return api.post<ApiSingle<Job>>(`/api/v1/ideas/${ideaId}/messages`, { content }).then((r) => r.data)
}

export function confirmIdea(ideaId: string) {
  return api.post<ApiSingle<Idea>>(`/api/v1/ideas/${ideaId}/confirm`).then((r) => r.data)
}
