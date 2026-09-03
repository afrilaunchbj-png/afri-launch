import { useTranslation } from "react-i18next"
import { Lightbulb, MapPin, Sparkles } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

import type { ChatIdea, ConversationDetail } from "../api"
import { ChatIdeaCard } from "./chat-idea-card"

interface ContextPanelProps {
  detail: ConversationDetail | undefined
  onConfirmIdea: (idea: ChatIdea) => void
  onCreateProject: (idea: ChatIdea) => void
  creating: boolean
}

export function ContextPanel({ detail, onConfirmIdea, onCreateProject, creating }: ContextPanelProps) {
  const { t } = useTranslation()
  const opportunity = detail?.opportunity
  const ideas = detail?.ideas ?? []

  return (
    <div className="flex h-full flex-col gap-4 overflow-y-auto">
      <section>
        <h2 className="mb-2 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          <MapPin className="h-3.5 w-3.5" />
          {t("chat:opportunity")}
        </h2>
        {opportunity ? (
          <Card>
            <CardHeader className="p-4 pb-2">
              <div className="flex items-start justify-between gap-2">
                <CardTitle className="text-sm leading-snug">{opportunity.title}</CardTitle>
                <Badge variant="secondary" className="shrink-0">
                  {opportunity.score}/100
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground">
                {opportunity.country} · {opportunity.sector}
              </p>
            </CardHeader>
            <CardContent className="p-4 pt-0">
              <p className="line-clamp-4 text-xs leading-relaxed text-muted-foreground">{opportunity.summary}</p>
            </CardContent>
          </Card>
        ) : (
          <p className="rounded-lg border border-dashed p-4 text-xs text-muted-foreground">
            {t("chat:noOpportunity")}
          </p>
        )}
      </section>

      <section>
        <h2 className="mb-2 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          <Sparkles className="h-3.5 w-3.5" />
          {t("chat:ideas")}
        </h2>
        {ideas.length === 0 ? (
          <p className="rounded-lg border border-dashed p-4 text-xs text-muted-foreground">{t("chat:noIdeas")}</p>
        ) : (
          <div className="space-y-3">
            {ideas.map((idea) => (
              <ChatIdeaCard
                key={idea.id}
                idea={idea}
                pending={creating}
                onConfirm={() => onConfirmIdea(idea)}
                onCreateProject={() => onCreateProject(idea)}
              />
            ))}
          </div>
        )}
      </section>

      {ideas.some((i) => i.status === "confirmed") ? null : (
        <p className="mt-auto flex items-start gap-1.5 rounded-lg bg-muted/50 p-3 text-xs text-muted-foreground">
          <Lightbulb className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          {t("chat:contextHint")}
        </p>
      )}
    </div>
  )
}
