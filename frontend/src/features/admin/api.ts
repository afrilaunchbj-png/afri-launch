import { api, type ApiList, type ApiSingle } from "@/lib/api/client"

export interface AdminStats {
  users: number
  projects: number
  assets: number
  conversations: number
  jobs_by_status: Record<string, number>
  credits_consumed: number
  open_tickets: number
}

export interface AdminUser {
  id: string
  email: string
  full_name: string
  role: "user" | "superadmin"
  created_at: string
}

export interface AdminTicket {
  id: string
  user_email: string
  user_name: string
  subject: string
  message: string
  status: "open" | "resolved"
  created_at: string
}

export function fetchAdminStats() {
  return api.get<ApiSingle<AdminStats>>("/api/v1/admin/stats").then((r) => r.data)
}

export function fetchAdminUsers(page: number, pageSize = 20) {
  return api
    .get<ApiList<AdminUser>>(`/api/v1/admin/users?page=${page}&pageSize=${pageSize}`)
    .then((r) => r)
}

export function fetchAdminTickets(page: number, pageSize = 20) {
  return api
    .get<ApiList<AdminTicket>>(`/api/v1/admin/tickets?page=${page}&pageSize=${pageSize}`)
    .then((r) => r)
}

export function resolveTicket(id: string) {
  return api.post<ApiSingle<{ id: string; status: string }>>(`/api/v1/admin/tickets/${id}/resolve`)
}
