import { useEffect, useState } from "react"
import { Link, useNavigate, useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import { Lightbulb, Plus } from "lucide-react"

import { EmptyState } from "@/components/states/empty-state"
import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { useJob } from "@/features/generation/hooks"
import { useGenerateIdeas, useIdeas } from "@/features/ideas/hooks"
import { useCreateProject } from "@/features/projects/hooks"

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
    if (job && (job.status === "completed" || job.status === "failed")) {
      refetch()
      setJobId(null)
    }
  }, [job, refetch])

  const handleGenerate = () => {
    generate.mutate(undefined, {
      onSuccess: (j) => setJobId(j.id),
    })
  }

  const handleCreateProject = (ideaId: string, title: string, oppId?: string | null) => {
    createProject.mutate(
      { idea_id: ideaId, opportunity_id: oppId ?? null, title },
      { onSuccess: (p) => navigate(`/projects/${p.id}`) },
    )
  }

  const generating = job && (job.status === "pending" || job.status === "processing")

  return (
    <div className="space-y-6">
      <header className="flex items-center justify-between gap-4">
        <div>
          <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">{t("ideas.title")}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t("ideas.subtitle")}</p>
        </div>
        {opportunityId ? (
          <Button size="touch" onClick={handleGenerate} disabled={generating}>
            <Plus className="h-4 w-4" />
            {generating ? t("common.generating") : t("ideas.generate")}
          </Button>
        ) : null}
      </header>

      {generating ? <LoadingState label={t("ideas.generatingHint")} /> : null}

      {isLoading ? (
        <LoadingState label={t("common.loading")} />
      ) : isError ? (
        <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />
      ) : !ideas || ideas.length === 0 ? (
        <EmptyState
          icon={Lightbulb}
          title={t("ideas.empty")}
          description={opportunityId ? t("ideas.emptyDesc") : t("ideas.noOpportunity")}
          actionLabel={opportunityId ? t("ideas.generate") : undefined}
          onAction={opportunityId ? handleGenerate : undefined}
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {ideas.map((idea) => (
            <Card key={idea.id}>
              <CardContent className="flex h-full flex-col gap-3 p-5">
                <div className="flex items-start justify-between gap-2">
                  <h3 className="font-display text-base font-semibold text-primary">{idea.title}</h3>
                  {idea.difficulty ? <Badge variant="outline">{idea.difficulty}</Badge> : null}
                </div>
                {idea.subtitle ? <p className="text-sm text-muted-foreground">{idea.subtitle}</p> : null}
                {idea.audience ? (
                  <p className="text-xs text-muted-foreground">
                    {t("ideas.audience")}: {idea.audience}
                  </p>
                ) : null}
                {idea.format || idea.estimated_price ? (
                  <p className="text-xs text-muted-foreground">
                    {[idea.format, idea.estimated_price].filter(Boolean).join(" · ")}
                  </p>
                ) : null}
                <Button
                  size="sm"
                  className="mt-auto w-full sm:w-auto"
                  onClick={() => handleCreateProject(idea.id, idea.title, idea.opportunity_id ?? null)}
                  disabled={createProject.isPending}
                >
                  {t("ideas.createProject")}
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <div className="text-sm text-muted-foreground">
        <Link to="/opportunities" className="text-primary hover:underline">
          ← {t("ideas.backToOpportunities")}
        </Link>
      </div>
    </div>
  )
}
