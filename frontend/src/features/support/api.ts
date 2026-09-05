import { api, type ApiSingle } from "@/lib/api/client"

export interface SupportTicket {
  id: string
  subject: string
  message: string
  status: "open" | "resolved"
  created_at: string
  attachments?: SupportAttachment[]
}

export interface SupportAttachment {
  id: string
  filename: string
  content_type: string
  size_bytes: number
  is_image: boolean
}

export interface TicketMessage {
  id: string
  author_id: string
  author_name: string
  is_admin: boolean
  content: string
  created_at: string
  attachments?: SupportAttachment[]
}

export interface TicketDetail {
  ticket: SupportTicket
  messages: TicketMessage[]
  attachments?: SupportAttachment[]
}

export function createTicket(subject: string, message: string, attachmentIds: string[] = []) {
  return api
    .post<ApiSingle<SupportTicket>>("/api/v1/support/tickets", {
      subject,
      message,
      attachment_ids: attachmentIds,
    })
    .then((r) => r.data)
}

export function fetchMyTickets() {
  return api.get<ApiSingle<SupportTicket[]>>("/api/v1/support/tickets").then((r) => r.data)
}

export function fetchTicketDetail(id: string) {
  return api.get<ApiSingle<TicketDetail>>(`/api/v1/support/tickets/${id}`).then((r) => r.data)
}

export function replyTicket(id: string, content: string, attachmentIds: string[] = []) {
  return api
    .post<ApiSingle<TicketDetail>>(`/api/v1/support/tickets/${id}/messages`, {
      content,
      attachment_ids: attachmentIds,
    })
    .then((r) => r.data)
}

/** uploadAttachments envoie des fichiers (multipart) et renvoie leurs IDs. */
export async function uploadAttachments(files: File[]): Promise<string[]> {
  if (files.length === 0) return []
  const formData = new FormData()
  for (const file of files) {
    formData.append("file", file)
  }
  const attachments = await api
    .upload<ApiSingle<SupportAttachment[]>>("/api/v1/support/attachments", formData)
    .then((r) => r.data)
  return attachments.map((a) => a.id)
}

export function attachmentDownloadPath(id: string) {
  return `/api/v1/support/attachments/${id}/download`
}

export async function downloadAttachment(id: string, filename: string) {
  const blob = await api.download(attachmentDownloadPath(id))
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
