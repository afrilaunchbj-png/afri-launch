import { useEffect, useState } from "react"
import { Link, useParams } from "react-router"
import { useTranslation } from "react-i18next"
import { BookOpen, Download, Image as ImageIcon, Megaphone } from "lucide-react"

import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Spinner } from "@/components/states/loading-state"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { useJob } from "@/features/generation/hooks"
import { downloadAsset, useAssets, useGenerate, useProject } from "@/features/projects/hooks"

export default function ProjectPage() {
  const { id = "" } = useParams()
  const { t } = useTranslation()

  const { data: project, isLoading, isError, refetch } = useProject(id)
  const { data: assets } = useAssets(id)
  const generate = useGenerate("ebook")
  const generateCover = useGenerate("cover")
  const generateSalesPage = useGenerate("sales-page")

  const [jobId, setJobId] = useState<string | null>(null)
  const { data: job } = useJob(jobId)

  useEffect(() => {
    if (job && (job.status === "completed" || job.status === "failed")) {
      refetch()
      setJobId(null)
    }
  }, [job, refetch])

  const generating = job && (job.status === "pending" || job.status === "processing")

  const handleDownload = async (assetId: string, filename: string) => {
    await downloadAsset(assetId, filename)
  }

  if (isLoading) return <LoadingState label={t("common.loading")} />
  if (isError || !project) return <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />

  const hasAsset = (kind: string) => assets?.some((a) => a.kind === kind)

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <Link to="/projects" className="text-sm text-primary hover:underline">
            ← {t("projects.back")}
          </Link>
          <h1 className="mt-1 font-display text-2xl font-bold text-primary md:text-3xl">{project.title}</h1>
          <Badge
            variant={project.status === "failed" ? "destructive" : project.status === "completed" ? "success" : "outline"}
            className="mt-2"
          >
            {t(`projects.status.${project.status}`)}
          </Badge>
        </div>
        <p className="text-sm text-muted-foreground">{t("projects.creditsUsed", { count: project.credits_consumed })}</p>
      </header>

      {generating ? (
        <Card>
          <CardContent className="flex items-center gap-3 p-4">
            <Spinner className="h-5 w-5" />
            <span>{t("common.generating")}</span>
            {job?.status ? <Badge variant="outline">{job.status}</Badge> : null}
          </CardContent>
        </Card>
      ) : null}

      <section className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardContent className="flex h-full flex-col gap-3 p-5">
            <BookOpen className="h-6 w-6 text-primary" />
            <h3 className="font-semibold">{t("projects.ebook")}</h3>
            <Button
              size="sm"
              disabled={generating || hasAsset("ebook_pdf")}
              onClick={() => generate.mutate(id, { onSuccess: (j) => setJobId(j.id) })}
            >
              {hasAsset("ebook_pdf") ? t("projects.generated") : t("projects.generate")}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex h-full flex-col gap-3 p-5">
            <ImageIcon className="h-6 w-6 text-primary" />
            <h3 className="font-semibold">{t("projects.cover")}</h3>
            <Button
              size="sm"
              disabled={generating || hasAsset("cover")}
              onClick={() => generateCover.mutate(id, { onSuccess: (j) => setJobId(j.id) })}
            >
              {hasAsset("cover") ? t("projects.generated") : t("projects.generate")}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex h-full flex-col gap-3 p-5">
            <Megaphone className="h-6 w-6 text-primary" />
            <h3 className="font-semibold">{t("projects.salesPage")}</h3>
            <Button
              size="sm"
              disabled={generating || hasAsset("sales_page")}
              onClick={() => generateSalesPage.mutate(id, { onSuccess: (j) => setJobId(j.id) })}
            >
              {hasAsset("sales_page") ? t("projects.generated") : t("projects.generate")}
            </Button>
          </CardContent>
        </Card>
      </section>

      <section>
        <h2 className="mb-3 font-display text-lg font-semibold text-primary">{t("projects.assets")}</h2>
        {!assets || assets.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("projects.noAssets")}</p>
        ) : (
          <div className="divide-y rounded-lg border bg-card">
            {assets.map((a) => (
              <div key={a.id} className="flex items-center justify-between gap-3 p-4">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{a.filename}</p>
                  <p className="text-xs text-muted-foreground">
                    {t(`projects.assetKind.${a.kind}`)} · {(a.size_bytes / 1024).toFixed(0)} KB
                  </p>
                </div>
                <Button size="sm" variant="outline" onClick={() => handleDownload(a.id, a.filename)}>
                  <Download className="h-4 w-4" />
                  {t("projects.download")}
                </Button>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
