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

export interface AdminProject {
  id: string
  title: string
  status: string
  credits_consumed: number
  created_at: string
  user_email: string
  user_name: string
}

export interface AdminConversation {
  id: string
  title: string
  status: string
  created_at: string
  user_email: string
  user_name: string
}

export interface AdminAsset {
  id: string
  kind: string
  filename: string
  content_type: string
  size_bytes: number
  created_at: string
  project_title: string
  user_email: string
}

export interface AdminJob {
  id: string
  kind: string
  status: string
  cost: number
  created_at: string
  updated_at: string
  user_email: string
  user_name: string
}

export interface AdminCreditTransaction {
  id: string
  type: "credit" | "debit"
  amount: number
  operation: string
  status: string
  created_at: string
  user_email: string
}

export interface AdminAuditLog {
  id: string
  user_id: string
  action: string
  entity: string
  entity_id: string
  metadata: Record<string, unknown>
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
  ticket: AdminTicket
  messages: TicketMessage[]
}

/** Paramètres de liste admin : pagination + filtres serveur. */
export interface AdminListParams {
  page?: number
  pageSize?: number
  search?: string
  status?: string
  role?: string
}

function query(params: AdminListParams): string {
  const q = new URLSearchParams()
  if (params.page) q.set("page", String(params.page))
  if (params.pageSize) q.set("pageSize", String(params.pageSize))
  if (params.search) q.set("search", params.search)
  if (params.status) q.set("status", params.status)
  if (params.role) q.set("role", params.role)
  const s = q.toString()
  return s ? `?${s}` : ""
}

export function fetchAdminStats() {
  return api.get<ApiSingle<AdminStats>>("/api/v1/admin/stats").then((r) => r.data)
}

export function fetchAdminUsers(params: AdminListParams) {
  return api.get<ApiList<AdminUser>>(`/api/v1/admin/users${query(params)}`)
}

export function fetchAdminTickets(params: AdminListParams) {
  return api.get<ApiList<AdminTicket>>(`/api/v1/admin/tickets${query(params)}`)
}

export function fetchAdminProjects(params: AdminListParams) {
  return api.get<ApiList<AdminProject>>(`/api/v1/admin/projects${query(params)}`)
}

export function fetchAdminConversations(params: AdminListParams) {
  return api.get<ApiList<AdminConversation>>(`/api/v1/admin/conversations${query(params)}`)
}

export function fetchAdminAssets(params: AdminListParams) {
  return api.get<ApiList<AdminAsset>>(`/api/v1/admin/assets${query(params)}`)
}

export function fetchAdminJobs(params: AdminListParams) {
  return api.get<ApiList<AdminJob>>(`/api/v1/admin/jobs${query(params)}`)
}

export function fetchAdminCreditTransactions(params: AdminListParams) {
  return api.get<ApiList<AdminCreditTransaction>>(`/api/v1/admin/credit-transactions${query(params)}`)
}

export interface AuditListParams {
  page?: number
  pageSize?: number
  action?: string
  entity?: string
  userId?: string
}

function auditQuery(params: AuditListParams): string {
  const q = new URLSearchParams()
  if (params.page) q.set("page", String(params.page))
  if (params.pageSize) q.set("pageSize", String(params.pageSize))
  if (params.action) q.set("action", params.action)
  if (params.entity) q.set("entity", params.entity)
  if (params.userId) q.set("userId", params.userId)
  const s = q.toString()
  return s ? `?${s}` : ""
}

export function fetchAdminAuditLogs(params: AuditListParams) {
  return api.get<ApiList<AdminAuditLog>>(`/api/v1/admin/audit-logs${auditQuery(params)}`)
}

export function fetchAdminTicketDetail(id: string) {
  return api.get<ApiSingle<TicketDetail>>(`/api/v1/admin/tickets/${id}`).then((r) => r.data)
}

export function replyAdminTicket(id: string, content: string) {
  return api
    .post<ApiSingle<TicketMessage>>(`/api/v1/admin/tickets/${id}/messages`, { content })
    .then((r) => r.data)
}

export function resolveTicket(id: string) {
  return api.post<ApiSingle<{ id: string; status: string }>>(`/api/v1/admin/tickets/${id}/resolve`)
}
