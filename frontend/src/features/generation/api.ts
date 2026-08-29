import { api, type ApiSingle } from "@/lib/api/client"

export type JobStatus = "pending" | "processing" | "completed" | "failed"

export interface Job {
  id: string
  kind: string
  status: JobStatus
  error?: string
  cost: number
  result?: {
    idea_ids?: string[]
    asset_id?: string
  } | null
  created_at: string
  completed_at?: string | null
}

export function fetchJob(id: string) {
  return api.get<ApiSingle<Job>>(`/api/v1/jobs/${id}`).then((r) => r.data)
}
