import { api, type ApiSingle } from "@/lib/api/client"

export interface ChatConversation {
  id: string
  title: string
  status: string
  opportunity_id?: string | null
  created_at: string
  updated_at: string
}

export interface ChatMessage {
  id: string
  role: "user" | "assistant"
  content: string
  payload?: { idea_ids?: string[] } | null
  created_at: string
}

export interface ChatIdea {
  id: string
  opportunity_id?: string | null
  conversation_id?: string | null
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
  status: string
}

export interface ChatOpportunityScores {
  demand: number
  pain: number
  competition: number
  purchasing_power: number
  digital_fit: number
  evidence_strength: number
}

export interface ChatOpportunity {
  id: string
  title: string
  summary: string
  country: string
  sector: string
  difficulty: string
  signal: string
  score: number
  scores: ChatOpportunityScores
  is_saved?: boolean
}

export interface ConversationDetail extends ChatConversation {
  opportunity?: ChatOpportunity
  messages: ChatMessage[]
  ideas: ChatIdea[]
}

export interface SendMessageAccepted {
  conversation_id: string
  message_id: string
}

export function fetchConversation(id: string) {
  return api.get<ApiSingle<ConversationDetail>>(`/api/v1/conversations/${id}`).then((r) => r.data)
}

export function createConversation() {
  return api.post<ApiSingle<ChatConversation>>("/api/v1/conversations").then((r) => r.data)
}

export function sendChatMessage(conversationId: string, content: string) {
  return api
    .post<ApiSingle<SendMessageAccepted>>(`/api/v1/conversations/${conversationId}/messages`, { content })
    .then((r) => r.data)
}

export function confirmIdea(ideaId: string) {
  return api.post<ApiSingle<ChatIdea>>(`/api/v1/ideas/${ideaId}/confirm`).then((r) => r.data)
}
