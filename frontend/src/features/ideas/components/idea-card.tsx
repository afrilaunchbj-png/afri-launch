import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useQueryClient } from "@tanstack/react-query"
import { CheckCircle2, MessageSquare, Send } from "lucide-react"
import { toast } from "sonner"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { useJob } from "@/features/generation/hooks"
import { ideaKeys, useConfirmIdea, useIdeaMessages, useSendIdeaMessage } from "@/features/ideas/hooks"
import type { Idea } from "@/features/ideas/api"
import { isAppError } from "@/lib/errors"
import { cn } from "@/lib/utils"

export function IdeaCard({ idea, onCreate }: { idea: Idea; onCreate: (idea: Idea) => void }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [draft, setDraft] = useState("")
  const [jobId, setJobId] = useState<string | null>(null)

  const { data: job } = useJob(jobId)
  const { data: messages, refetch: refetchMessages } = useIdeaMessages(idea.id)
  const send = useSendIdeaMessage(idea.id)
  const confirm = useConfirmIdea(idea.id)

  useEffect(() => {
    if (!job) return
    if (job.status === "completed") {
      refetchMessages()
      queryClient.invalidateQueries({ queryKey: ideaKeys.all })
      setJobId(null)
    } else if (job.status === "failed") {
      toast.error(t("common.generationFailed", { error: job.error ?? "" }))
      setJobId(null)
    }
  }, [job, refetchMessages, queryClient, t])

  const busy = send.isPending || (job && (job.status === "pending" || job.status === "processing"))
  const confirmed = idea.status === "confirmed"

  const handleSend = () => {
    if (!draft.trim() || busy) return
    send.mutate(draft.trim(), {
      onSuccess: (j) => {
        setDraft("")
        setJobId(j.id)
      },
      onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
    })
  }

  const handleConfirm = () => {
    confirm.mutate(undefined, {
      onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
    })
  }

  return (
    <Card className="flex flex-col">
      <CardContent className="flex h-full flex-col gap-3 p-5">
        <div className="flex items-start justify-between gap-2">
          <h3 className="font-display text-base font-semibold text-primary">{idea.title}</h3>
          {confirmed ? (
            <Badge variant="success">
              <CheckCircle2 className="h-3 w-3" />
              {t("ideas:confirmed")}
            </Badge>
          ) : (
            <Badge variant="outline">{t("ideas:draft")}</Badge>
          )}
        </div>

        {idea.hook ? <p className="text-sm font-medium text-foreground">{idea.hook}</p> : null}
        {idea.explanation ? <p className="text-sm text-muted-foreground">{idea.explanation}</p> : null}

        <div className="mt-auto flex flex-wrap gap-2">
          <Button size="sm" variant="outline" onClick={() => setOpen((o) => !o)}>
            <MessageSquare className="h-4 w-4" />
            {t("ideas:challenge")}
          </Button>
          {!confirmed ? (
            <Button size="sm" variant="secondary" onClick={handleConfirm} loading={confirm.isPending}>
              {t("ideas:confirm")}
            </Button>
          ) : (
            <Button size="sm" onClick={() => onCreate(idea)}>
              {t("ideas:createProject")}
            </Button>
          )}
        </div>

        {open ? (
          <div className="space-y-3 rounded-lg border bg-muted/30 p-3">
            <div className="max-h-56 space-y-2 overflow-y-auto">
              {(messages ?? []).map((m) => (
                <div key={m.id} className={cn("flex", m.role === "user" ? "justify-end" : "justify-start")}>
                  <div
                    className={cn(
                      "max-w-[85%] whitespace-pre-wrap rounded-lg px-3 py-2 text-sm",
                      m.role === "user" ? "bg-primary text-primary-foreground" : "bg-card text-foreground",
                    )}
                  >
                    {m.content}
                  </div>
                </div>
              ))}
              {!messages || messages.length === 0 ? (
                <p className="text-xs text-muted-foreground">{t("ideas:noMessages")}</p>
              ) : null}
            </div>
            <div className="flex items-center gap-2">
              <Input
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                placeholder={t("ideas:challengePlaceholder")}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleSend()
                }}
              />
              <Button size="icon" onClick={handleSend} disabled={busy || !draft.trim()} loading={busy} aria-label={t("ideas:send")}>
                <Send className="h-4 w-4" />
              </Button>
            </div>
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}
