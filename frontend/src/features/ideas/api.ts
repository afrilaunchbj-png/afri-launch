import { api, type ApiSingle } from "@/lib/api/client"
import { getAccessToken } from "@/lib/auth"

import type { Job } from "@/features/generation/api"

const API_URL = import.meta.env.VITE_API_URL ?? ""

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

export function confirmIdea(ideaId: string) {
  return api.post<ApiSingle<Idea>>(`/api/v1/ideas/${ideaId}/confirm`).then((r) => r.data)
}

export interface StreamCallbacks {
  onDelta: (delta: string) => void
  onDone: (idea: Idea) => void
  onError: (message: string) => void
}

function parseSSEEvent(raw: string, cb: StreamCallbacks) {
  let event = "message"
  const dataLines: string[] = []
  for (const line of raw.split("\n")) {
    if (line.startsWith("event:")) event = line.slice(6).trim()
    else if (line.startsWith("data:")) dataLines.push(line.slice(5).trim())
  }
  const data = dataLines.join("\n")
  if (!data) return
  if (event === "done") {
    try {
      cb.onDone(JSON.parse(data) as Idea)
    } catch {
      /* ignore */
    }
  } else if (event === "error") {
    try {
      cb.onError(JSON.parse(data) as string)
    } catch {
      cb.onError(data)
    }
  } else {
    try {
      cb.onDelta(JSON.parse(data) as string)
    } catch {
      /* ignore */
    }
  }
}

export async function streamIdeaMessage(ideaId: string, content: string, cb: StreamCallbacks): Promise<void> {
  const token = await getAccessToken()
  const res = await fetch(`${API_URL}/api/v1/ideas/${ideaId}/messages`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ content }),
  })

  if (!res.ok || !res.body) {
    let message = `HTTP ${res.status}`
    try {
      const problem = await res.json()
      if (problem?.detail) message = problem.detail
    } catch {
      /* ignore */
    }
    cb.onError(message)
    return
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ""
  try {
    for (;;) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      let idx
      while ((idx = buffer.indexOf("\n\n")) !== -1) {
        const raw = buffer.slice(0, idx)
        buffer = buffer.slice(idx + 2)
        parseSSEEvent(raw, cb)
      }
    }
  } finally {
    reader.releaseLock()
  }
}
