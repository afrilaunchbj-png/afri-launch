import { api, type ApiSingle } from "@/lib/api/client"

import type { Job } from "@/features/generation/api"

export interface StartResearchInput {
  query: string
  sector: string
  markets: string[]
  language: string
}

export function startResearch(input: StartResearchInput) {
  return api.post<ApiSingle<Job>>("/api/v1/research", input).then((r) => r.data)
}
