import { useEffect, useState } from "react"
import { Link, useParams } from "react-router"
import { useTranslation } from "react-i18next"
import { BookOpen, Download, Image as ImageIcon, Images, Megaphone } from "lucide-react"
import { toast } from "sonner"

import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { useJob } from "@/features/generation/hooks"
import type { Job } from "@/features/generation/api"
import { downloadAsset, useAssets, useGenerate, useProject } from "@/features/projects/hooks"
import { isAppError } from "@/lib/errors"

function isActive(job?: Job) {
  return job?.status === "pending" || job?.status === "processing"
}

export default function ProjectPage() {
  const { id = "" } = useParams()
  const { t } = useTranslation()

  const { data: project, isLoading, isError, refetch } = useProject(id)
  const { data: assets } = useAssets(id)
  const generate = useGenerate("ebook")
  const generateCover = useGenerate("cover")
  const generatePosters = useGenerate("posters")
  const generateSalesPage = useGenerate("sales-page")

  const [ebookJobId, setEbookJobId] = useState<string | null>(null)
  const [coverJobId, setCoverJobId] = useState<string | null>(null)
  const [postersJobId, setPostersJobId] = useState<string | null>(null)
  const [salesJobId, setSalesJobId] = useState<string | null>(null)
  const { data: ebookJob } = useJob(ebookJobId)
  const { data: coverJob } = useJob(coverJobId)
  const { data: postersJob } = useJob(postersJobId)
  const { data: salesJob } = useJob(salesJobId)

  useEffect(() => {
    const entries = [
      { job: ebookJob, done: () => setEbookJobId(null) },
      { job: coverJob, done: () => setCoverJobId(null) },
      { job: postersJob, done: () => setPostersJobId(null) },
      { job: salesJob, done: () => setSalesJobId(null) },
    ]
    for (const entry of entries) {
      if (!entry.job) continue
      if (entry.job.status === "completed") {
        refetch()
        entry.done()
      } else if (entry.job.status === "failed") {
        toast.error(t("common.generationFailed", { error: entry.job.error ?? "" }))
        entry.done()
      }
    }
  }, [ebookJob, coverJob, postersJob, salesJob, refetch, t])

  const [downloadingId, setDownloadingId] = useState<string | null>(null)

  const handleDownload = async (assetId: string, filename: string) => {
    setDownloadingId(assetId)
    try {
      await downloadAsset(assetId, filename)
    } catch (error) {
      toast.error(isAppError(error) ? error.message : t("common.genericError"))
    } finally {
      setDownloadingId(null)
    }
  }

  const onError = (error: unknown) => toast.error(isAppError(error) ? error.message : t("common.genericError"))

  if (isLoading) return <LoadingState label={t("common.loading")} />
  if (isError || !project) return <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />

  const hasAsset = (kind: string) => assets?.some((a) => a.kind === kind)

  const ebookBusy = generate.isPending || isActive(ebookJob)
  const coverBusy = generateCover.isPending || isActive(coverJob)
  const postersBusy = generatePosters.isPending || isActive(postersJob)
  const salesBusy = generateSalesPage.isPending || isActive(salesJob)

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <Link to="/projects" className="text-sm text-primary hover:underline">
            ← {t("projects:back")}
          </Link>
          <h1 className="mt-1 font-display text-2xl font-bold text-primary md:text-3xl">{project.title}</h1>
          <Badge
            variant={project.status === "failed" ? "destructive" : project.status === "completed" ? "success" : "outline"}
            className="mt-2"
          >
            {t(`projects:status.${project.status}`)}
          </Badge>
        </div>
        <p className="text-sm text-muted-foreground">{t("projects:creditsUsed", { count: project.credits_consumed })}</p>
      </header>

      <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="flex h-full flex-col gap-3 p-5">
            <BookOpen className="h-6 w-6 text-primary" />
            <h3 className="font-semibold">{t("projects:ebook")}</h3>
            <Button
              size="sm"
              disabled={ebookBusy || hasAsset("ebook_pdf")}
              loading={ebookBusy}
              onClick={() => generate.mutate(id, { onSuccess: (j) => setEbookJobId(j.id), onError })}
            >
              {hasAsset("ebook_pdf") ? t("projects:generated") : t("projects:generate")}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex h-full flex-col gap-3 p-5">
            <ImageIcon className="h-6 w-6 text-primary" />
            <h3 className="font-semibold">{t("projects:cover")}</h3>
            <Button
              size="sm"
              disabled={coverBusy || hasAsset("cover")}
              loading={coverBusy}
              onClick={() => generateCover.mutate(id, { onSuccess: (j) => setCoverJobId(j.id), onError })}
            >
              {hasAsset("cover") ? t("projects:generated") : t("projects:generate")}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex h-full flex-col gap-3 p-5">
            <Images className="h-6 w-6 text-primary" />
            <h3 className="font-semibold">{t("projects:posters")}</h3>
            <Button
              size="sm"
              disabled={postersBusy || hasAsset("poster")}
              loading={postersBusy}
              onClick={() => generatePosters.mutate(id, { onSuccess: (j) => setPostersJobId(j.id), onError })}
            >
              {hasAsset("poster") ? t("projects:generated") : t("projects:generate")}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex h-full flex-col gap-3 p-5">
            <Megaphone className="h-6 w-6 text-primary" />
            <h3 className="font-semibold">{t("projects:salesPage")}</h3>
            <Button
              size="sm"
              disabled={salesBusy || hasAsset("sales_page")}
              loading={salesBusy}
              onClick={() => generateSalesPage.mutate(id, { onSuccess: (j) => setSalesJobId(j.id), onError })}
            >
              {hasAsset("sales_page") ? t("projects:generated") : t("projects:generate")}
            </Button>
          </CardContent>
        </Card>
      </section>

      <section>
        <h2 className="mb-3 font-display text-lg font-semibold text-primary">{t("projects:assets")}</h2>
        {!assets || assets.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("projects:noAssets")}</p>
        ) : (
          <div className="divide-y rounded-lg border bg-card">
            {assets.map((a) => (
              <div key={a.id} className="flex items-center justify-between gap-3 p-4">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{a.filename}</p>
                  <p className="text-xs text-muted-foreground">
                    {t(`projects:assetKind.${a.kind}`)} · {(a.size_bytes / 1024).toFixed(0)} KB
                  </p>
                </div>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={downloadingId !== null}
                  loading={downloadingId === a.id}
                  onClick={() => handleDownload(a.id, a.filename)}
                >
                  <Download className="h-4 w-4" />
                  {t("projects:download")}
                </Button>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
