import { api, type ApiSingle } from "@/lib/api/client"

export interface SupportTicket {
  id: string
  subject: string
  message: string
  status: "open" | "resolved"
  created_at: string
}

export function createTicket(subject: string, message: string) {
  return api
    .post<ApiSingle<SupportTicket>>("/api/v1/support/tickets", { subject, message })
    .then((r) => r.data)
}

export function fetchMyTickets() {
  return api.get<ApiSingle<SupportTicket[]>>("/api/v1/support/tickets").then((r) => r.data)
}
