import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Check, FolderPlus, Lightbulb, Pencil, WandSparkles } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

import type { ChatIdea } from "../api"

interface ChatIdeaCardProps {
  idea: ChatIdea
  onConfirm: (idea: ChatIdea) => void
  onCreateProject: (idea: ChatIdea) => void
  onRefine: () => void
  onSaveIdea: (title: string, subtitle: string) => void
  pending: boolean
  saving?: boolean
}

export function ChatIdeaCard({ idea, onConfirm, onCreateProject, onRefine, onSaveIdea, pending, saving }: ChatIdeaCardProps) {
  const { t } = useTranslation()
  const confirmed = idea.status === "confirmed"
  const [editOpen, setEditOpen] = useState(false)
  const [title, setTitle] = useState(idea.title)
  const [subtitle, setSubtitle] = useState(idea.subtitle ?? "")

  const openEdit = () => {
    setTitle(idea.title)
    setSubtitle(idea.subtitle ?? "")
    setEditOpen(true)
  }

  return (
    <Card className={confirmed ? "border-primary" : "border-primary/20"}>
      <CardContent className="space-y-3 p-4">
        <div className="flex items-start justify-between gap-2">
          <div className="flex min-w-0 items-start gap-2">
            <Lightbulb className="mt-0.5 h-4 w-4 shrink-0 text-accent" />
            <div className="min-w-0">
              <p className="text-sm font-semibold leading-snug">{idea.title}</p>
              {idea.subtitle ? <p className="mt-0.5 text-xs text-muted-foreground">{idea.subtitle}</p> : null}
              {!idea.subtitle && idea.hook ? <p className="mt-0.5 text-xs text-muted-foreground">{idea.hook}</p> : null}
            </div>
          </div>
          <Badge variant={confirmed ? "default" : "outline"} className="shrink-0 text-[10px]">
            {confirmed ? t("chat:ideaConfirmed") : t("chat:ideaDraft")}
          </Badge>
        </div>

        {idea.explanation ? (
          <p className="line-clamp-3 text-xs leading-relaxed text-muted-foreground">{idea.explanation}</p>
        ) : null}

        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={onRefine} className="flex-1">
            <WandSparkles className="h-3.5 w-3.5" />
            {t("chat:refineIdea")}
          </Button>
          <Button size="sm" variant="outline" onClick={openEdit}>
            <Pencil className="h-3.5 w-3.5" />
            {t("chat:editIdea")}
          </Button>
        </div>

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

      <Dialog open={editOpen} onOpenChange={(v) => !v && setEditOpen(false)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("chat:editIdeaTitle")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label htmlFor="idea-title">{t("chat:title")}</Label>
              <Input id="idea-title" value={title} onChange={(e) => setTitle(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="idea-subtitle">{t("chat:subtitle")}</Label>
              <Input id="idea-subtitle" value={subtitle} onChange={(e) => setSubtitle(e.target.value)} />
            </div>
            <p className="text-xs text-muted-foreground">{t("chat:editIdeaHint")}</p>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setEditOpen(false)}>
              {t("integrations:cancel")}
            </Button>
            <Button
              onClick={() => {
                onSaveIdea(title.trim(), subtitle.trim())
                setEditOpen(false)
              }}
              disabled={!title.trim()}
              loading={saving}
            >
              <Check className="h-4 w-4" />
              {t("chat:saveConfirmIdea")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  )
}
