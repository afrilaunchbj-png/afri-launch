import { api, type ApiSingle } from "@/lib/api/client"

export interface SupportTicket {
  id: string
  subject: string
  message: string
  status: "open" | "resolved"
  created_at: string
}

export interface TicketMessage {
  id: string
  author_id: string
  author_name: string
  is_admin: boolean
  content: string
  created_at: string
}

export interface TicketDetail {
  ticket: SupportTicket
  messages: TicketMessage[]
}

export function createTicket(subject: string, message: string) {
  return api
    .post<ApiSingle<SupportTicket>>("/api/v1/support/tickets", { subject, message })
    .then((r) => r.data)
}

export function fetchMyTickets() {
  return api.get<ApiSingle<SupportTicket[]>>("/api/v1/support/tickets").then((r) => r.data)
}

export function fetchTicketDetail(id: string) {
  return api.get<ApiSingle<TicketDetail>>(`/api/v1/support/tickets/${id}`).then((r) => r.data)
}

export function replyTicket(id: string, content: string) {
  return api
    .post<ApiSingle<TicketDetail>>(`/api/v1/support/tickets/${id}/messages`, { content })
    .then((r) => r.data)
}
