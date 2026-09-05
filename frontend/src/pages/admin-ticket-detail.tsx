import { useState } from "react"
import { Link, useNavigate, useParams } from "react-router"
import { useTranslation } from "react-i18next"
import { ArrowLeft, CheckCircle2, Send } from "lucide-react"

import { AdminNav } from "@/features/admin/admin-nav"
import { useAdminReplyTicket, useAdminResolveTicket, useAdminTicketDetail } from "@/features/admin/hooks"
import { AttachmentList } from "@/features/support/attachment-list"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/utils"

export default function AdminTicketDetailPage() {
  const { id = "" } = useParams()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data, isLoading } = useAdminTicketDetail(id)
  const reply = useAdminReplyTicket(id)
  const resolve = useAdminResolveTicket(data?.ticket.id)
  const [content, setContent] = useState("")

  if (isLoading) {
    return (
      <div className="space-y-6">
        <AdminNav />
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-64 w-full" />
      </div>
    )
  }

  if (!data) {
    return (
      <div className="space-y-6">
        <AdminNav />
        <p className="text-sm text-muted-foreground">{t("admin:noTickets")}</p>
      </div>
    )
  }

  const { ticket, messages } = data

  const submitReply = () => {
    const trimmed = content.trim()
    if (!trimmed) return
    reply.mutate(trimmed, { onSuccess: () => setContent("") })
  }

  return (
    <div className="space-y-6">
      <AdminNav />
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <Button variant="ghost" size="sm" className="-ml-2 mb-1" onClick={() => navigate("/admin/tickets")}>
            <ArrowLeft className="h-4 w-4" />
            {t("admin:back")}
          </Button>
          <h1 className="font-display text-2xl font-bold text-primary">{ticket.subject}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {ticket.user_name || "—"} · {ticket.user_email} ·{" "}
            {new Date(ticket.created_at).toLocaleString()}
          </p>
        </div>
        <div className="flex shrink-0 flex-col items-end gap-2">
          <Badge variant={ticket.status === "resolved" ? "success" : "secondary"}>
            {t(`support:status.${ticket.status}`)}
          </Badge>
          {ticket.status === "open" && (
            <Button size="sm" variant="outline" disabled={resolve.isPending} onClick={() => resolve.mutate(ticket.id)}>
              <CheckCircle2 className="h-4 w-4" />
              {t("admin:resolve")}
            </Button>
          )}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("admin:discussion")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Message initial du ticket */}
          <div className="flex justify-start">
            <div className="max-w-[80%] space-y-1">
              <p className="text-xs text-muted-foreground">{ticket.user_name || ticket.user_email}</p>
              <div className="rounded-2xl rounded-tl-sm bg-muted px-4 py-3 text-sm">
                <p className="whitespace-pre-wrap">{ticket.message}</p>
              </div>
              <AttachmentList attachments={data.attachments} />
              <p className="text-[11px] text-muted-foreground">{new Date(ticket.created_at).toLocaleString()}</p>
            </div>
          </div>

          {messages.map((m) => (
            <div key={m.id} className={cn("flex", m.is_admin ? "justify-end" : "justify-start")}>
              <div className="max-w-[80%] space-y-1">
                <p className={cn("text-xs text-muted-foreground", m.is_admin && "text-right")}>
                  {m.is_admin ? t("admin:supportTeam") : m.author_name || m.author_id}
                </p>
                <div
                  className={cn(
                    "rounded-2xl px-4 py-3 text-sm",
                    m.is_admin ? "rounded-tr-sm bg-primary text-primary-foreground" : "rounded-tl-sm bg-muted",
                  )}
                >
                  <p className="whitespace-pre-wrap">{m.content}</p>
                </div>
                <AttachmentList attachments={m.attachments} />
                <p className={cn("text-[11px] text-muted-foreground", m.is_admin && "text-right")}>
                  {new Date(m.created_at).toLocaleString()}
                </p>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardContent className="space-y-3 p-4">
          <Textarea
            rows={3}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder={t("admin:replyPlaceholder")}
          />
          <div className="flex justify-end">
            <Button onClick={submitReply} loading={reply.isPending} disabled={!content.trim()}>
              <Send className="h-4 w-4" />
              {t("admin:sendReply")}
            </Button>
          </div>
        </CardContent>
      </Card>

      <p className="text-center text-xs text-muted-foreground">
        <Link to="/admin/tickets" className="underline">
          {t("admin:ticketsTitle")}
        </Link>
      </p>
    </div>
  )
}
