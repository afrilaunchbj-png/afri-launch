import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import { FolderOpen } from "lucide-react"

import { EmptyState } from "@/components/states/empty-state"
import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { useProjects } from "@/features/projects/hooks"

export default function ProjectsPage() {
  const { t } = useTranslation()
  const { data: projects, isLoading, isError, refetch } = useProjects()

  return (
    <div className="space-y-6">
      <header>
        <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">{t("projects.title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("projects.subtitle")}</p>
      </header>

      {isLoading ? (
        <LoadingState label={t("common.loading")} />
      ) : isError ? (
        <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />
      ) : !projects || projects.length === 0 ? (
        <EmptyState icon={FolderOpen} title={t("projects.empty")} description={t("projects.emptyDesc")} />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {projects.map((p) => (
            <Link key={p.id} to={`/projects/${p.id}`}>
              <Card className="transition-shadow hover:shadow-md">
                <CardContent className="p-5">
                  <div className="flex items-center justify-between gap-2">
                    <h3 className="font-display text-base font-semibold text-primary">{p.title}</h3>
                    <Badge variant={p.status === "failed" ? "destructive" : p.status === "completed" ? "success" : "outline"}>
                      {t(`projects.status.${p.status}`)}
                    </Badge>
                  </div>
                  <p className="mt-2 text-xs text-muted-foreground">
                    {t("projects.creditsUsed", { count: p.credits_consumed })}
                  </p>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
