import { useState } from "react"
import { Link, useParams } from "react-router"
import { useTranslation } from "react-i18next"
import { ArrowLeft, Headset, Send } from "lucide-react"

import { useReplyTicket, useTicketDetail } from "@/features/support/hooks"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/utils"

export default function SupportTicketPage() {
  const { id = "" } = useParams()
  const { t } = useTranslation()
  const { data, isLoading } = useTicketDetail(id)
  const reply = useReplyTicket(id)
  const [content, setContent] = useState("")

  if (isLoading) {
    return (
      <div className="mx-auto max-w-3xl space-y-6">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-64 w-full" />
      </div>
    )
  }

  if (!data) {
    return (
      <div className="mx-auto max-w-3xl space-y-6">
        <p className="text-sm text-muted-foreground">{t("support:empty")}</p>
        <Button asChild variant="outline">
          <Link to="/support">{t("support:back")}</Link>
        </Button>
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
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <Button variant="ghost" size="sm" className="-ml-2 mb-1" asChild>
          <Link to="/support">
            <ArrowLeft className="h-4 w-4" />
            {t("support:back")}
          </Link>
        </Button>
        <h1 className="flex items-center gap-2 font-display text-2xl font-bold text-primary md:text-3xl">
          <Headset className="h-6 w-6" />
          {ticket.subject}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">{new Date(ticket.created_at).toLocaleString()}</p>
      </div>

      <div className="flex items-center gap-2">
        <Badge variant={ticket.status === "resolved" ? "success" : "secondary"}>
          {t(`support:status.${ticket.status}`)}
        </Badge>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("support:discussion")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Message initial */}
          <div className="flex justify-start">
            <div className="max-w-[80%] space-y-1">
              <div className="rounded-2xl rounded-tl-sm bg-muted px-4 py-3 text-sm">
                <p className="whitespace-pre-wrap">{ticket.message}</p>
              </div>
              <p className="text-[11px] text-muted-foreground">{new Date(ticket.created_at).toLocaleString()}</p>
            </div>
          </div>

          {messages.map((m) => (
            <div key={m.id} className={cn("flex", m.is_admin ? "justify-end" : "justify-start")}>
              <div className="max-w-[80%] space-y-1">
                <p className={cn("text-xs text-muted-foreground", m.is_admin && "text-right")}>
                  {m.is_admin ? t("support:supportTeam") : t("support:you")}
                </p>
                <div
                  className={cn(
                    "rounded-2xl px-4 py-3 text-sm",
                    m.is_admin ? "rounded-tr-sm bg-primary text-primary-foreground" : "rounded-tl-sm bg-muted",
                  )}
                >
                  <p className="whitespace-pre-wrap">{m.content}</p>
                </div>
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
            placeholder={t("support:replyPlaceholder")}
          />
          <div className="flex justify-end">
            <Button onClick={submitReply} loading={reply.isPending} disabled={!content.trim()}>
              <Send className="h-4 w-4" />
              {t("support:sendReply")}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
