import { useEffect, useState } from "react"
import { Link, useNavigate, useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import { Lightbulb, Plus } from "lucide-react"
import { toast } from "sonner"

import { EmptyState } from "@/components/states/empty-state"
import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Button } from "@/components/ui/button"
import { useJob } from "@/features/generation/hooks"
import { IdeaCard } from "@/features/ideas/components/idea-card"
import { useGenerateIdeas, useIdeas } from "@/features/ideas/hooks"
import type { Idea } from "@/features/ideas/api"
import { useCreateProject } from "@/features/projects/hooks"
import { isAppError } from "@/lib/errors"

export default function IdeasPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const opportunityId = searchParams.get("opportunity") ?? undefined

  const { data: ideas, isLoading, isError, refetch } = useIdeas(opportunityId)
  const generate = useGenerateIdeas(opportunityId ?? "")
  const createProject = useCreateProject()

  const [jobId, setJobId] = useState<string | null>(null)
  const { data: job } = useJob(jobId)

  useEffect(() => {
    if (!job) return
    if (job.status === "completed") {
      refetch()
      setJobId(null)
    } else if (job.status === "failed") {
      toast.error(t("common.generationFailed", { error: job.error ?? "" }))
      setJobId(null)
    }
  }, [job, refetch, t])

  const handleGenerate = () => {
    generate.mutate(undefined, {
      onSuccess: (j) => setJobId(j.id),
      onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
    })
  }

  const handleCreateProject = (idea: Idea) => {
    createProject.mutate(
      { idea_id: idea.id, opportunity_id: idea.opportunity_id ?? null, title: idea.title },
      {
        onSuccess: (p) => navigate(`/projects/${p.id}`),
        onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
      },
    )
  }

  const generating = generate.isPending || (job && (job.status === "pending" || job.status === "processing"))

  return (
    <div className="space-y-6">
      <header className="flex items-center justify-between gap-4">
        <div>
          <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">{t("ideas:title")}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t("ideas:subtitle")}</p>
        </div>
        {opportunityId ? (
          <Button size="touch" onClick={handleGenerate} disabled={generating} loading={generating}>
            <Plus className="h-4 w-4" />
            {generating ? t("common.generating") : t("ideas:generate")}
          </Button>
        ) : null}
      </header>

      {generating ? <LoadingState label={t("ideas:generatingHint")} /> : null}

      {isLoading ? (
        <LoadingState label={t("common.loading")} />
      ) : isError ? (
        <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />
      ) : !ideas || ideas.length === 0 ? (
        <EmptyState
          icon={Lightbulb}
          title={t("ideas:empty")}
          description={opportunityId ? t("ideas:emptyDesc") : t("ideas:noOpportunity")}
          actionLabel={opportunityId ? t("ideas:generate") : undefined}
          onAction={opportunityId ? handleGenerate : undefined}
          actionLoading={generating}
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {ideas.map((idea) => (
            <IdeaCard key={idea.id} idea={idea} onCreate={handleCreateProject} />
          ))}
        </div>
      )}

      <div className="text-sm text-muted-foreground">
        <Link to="/opportunities" className="text-primary hover:underline">
          ← {t("ideas:backToOpportunities")}
        </Link>
      </div>
    </div>
  )
}
