import { useTranslation } from "react-i18next"
import { Check, FolderPlus, Lightbulb } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

import type { ChatIdea } from "../api"

interface ChatIdeaCardProps {
  idea: ChatIdea
  onConfirm: (idea: ChatIdea) => void
  onCreateProject: (idea: ChatIdea) => void
  pending: boolean
}

export function ChatIdeaCard({ idea, onConfirm, onCreateProject, pending }: ChatIdeaCardProps) {
  const { t } = useTranslation()
  const confirmed = idea.status === "confirmed"

  return (
    <Card className="border-primary/20">
      <CardContent className="space-y-3 p-4">
        <div className="flex items-start justify-between gap-2">
          <div className="flex min-w-0 items-start gap-2">
            <Lightbulb className="mt-0.5 h-4 w-4 shrink-0 text-accent" />
            <div className="min-w-0">
              <p className="truncate text-sm font-semibold">{idea.title}</p>
              {idea.hook ? <p className="mt-0.5 text-xs text-muted-foreground">{idea.hook}</p> : null}
            </div>
          </div>
          <Badge variant={confirmed ? "default" : "outline"} className="shrink-0 text-[10px]">
            {confirmed ? t("chat:ideaConfirmed") : t("chat:ideaDraft")}
          </Badge>
        </div>

        {idea.explanation ? (
          <p className="line-clamp-3 text-xs leading-relaxed text-muted-foreground">{idea.explanation}</p>
        ) : null}

        {confirmed ? (
          <Button size="sm" className="w-full" onClick={() => onCreateProject(idea)} disabled={pending}>
            <FolderPlus className="h-4 w-4" />
            {t("chat:createProject")}
          </Button>
        ) : (
          <Button
            size="sm"
            variant="outline"
            className="w-full"
            onClick={() => onConfirm(idea)}
            disabled={pending}
          >
            <Check className="h-4 w-4" />
            {t("chat:confirmIdea")}
          </Button>
        )}
      </CardContent>
    </Card>
  )
}
