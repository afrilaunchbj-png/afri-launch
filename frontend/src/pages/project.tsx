import { useEffect, useMemo, useState } from "react"
import { Link, useParams } from "react-router"
import { useTranslation } from "react-i18next"
import {
  BookOpen,
  Clapperboard,
  Download,
  Image as ImageIcon,
  Images,
  Lock,
  Megaphone,
  Palette as PaletteIcon,
  RefreshCw,
  Save,
  Sparkles,
} from "lucide-react"
import { toast } from "sonner"

import { api } from "@/lib/api/client"
import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { useJob } from "@/features/generation/hooks"
import type { Job } from "@/features/generation/api"
import type { ProjectPalette } from "@/features/projects/api"
import { assetDownloadPath } from "@/features/projects/api"
import { downloadAsset, useAssets, useGenerate, useGenerateCover, useProject, useUpdateProjectConfig } from "@/features/projects/hooks"
import { useLatestVideoAsset, VideoAdsPanel, VideoPreview } from "@/features/video-ads/video-ads-panel"
import { isAppError } from "@/lib/errors"

function isActive(job?: Job) {
  return job?.status === "pending" || job?.status === "processing"
}

const PALETTE_FIELDS = ["primary", "secondary", "accent", "background", "text"] as const
type PaletteField = (typeof PALETTE_FIELDS)[number]

const DEFAULT_FALLBACK_PALETTE: Record<PaletteField, string> = {
  primary: "#003527",
  secondary: "#855300",
  accent: "#fea619",
  background: "#ffffff",
  text: "#1c1917",
}

function StepBadge({ step, done, locked }: { step: number; done: boolean; locked: boolean }) {
  const { t } = useTranslation()
  return (
    <span
      className={
        "flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold " +
        (done
          ? "bg-primary text-primary-foreground"
          : locked
            ? "bg-muted text-muted-foreground"
            : "border border-primary text-primary")
      }
      title={locked ? t("projects:locked") : undefined}
    >
      {locked ? <Lock className="h-3 w-3" /> : step}
    </span>
  )
}

export default function ProjectPage() {
  const { id = "" } = useParams()
  const { t } = useTranslation()

  const { data: project, isLoading, isError, refetch } = useProject(id)
  const { data: assets } = useAssets(id)
  const generateEbook = useGenerate("ebook")
  const generatePosters = useGenerate("posters")
  const generateSalesPage = useGenerate("sales-page")
  const generateCover = useGenerateCover(id)
  const updateConfig = useUpdateProjectConfig(id)

  const config = project?.config
  const hasCover = useMemo(() => !!assets?.some((a) => a.kind === "cover"), [assets])
  const coverAsset = useMemo(() => assets?.filter((a) => a.kind === "cover").slice(-1)[0], [assets])
  const videoAsset = useLatestVideoAsset(assets)

  // Aperçu de la cover (dernière générée).
  const [coverURL, setCoverURL] = useState<string | null>(null)
  useEffect(() => {
    if (!coverAsset) {
      setCoverURL(null)
      return
    }
    let objectURL: string | null = null
    api
      .download(assetDownloadPath(coverAsset.id))
      .then((blob) => {
        objectURL = URL.createObjectURL(blob)
        setCoverURL(objectURL)
      })
      .catch(() => setCoverURL(null))
    return () => {
      if (objectURL) URL.revokeObjectURL(objectURL)
    }
  }, [coverAsset?.id])

  // Palette locale (éditable) initialisée depuis la config du projet.
  const [palette, setPalette] = useState<Record<PaletteField, string>>(DEFAULT_FALLBACK_PALETTE)
  const [hasPalette, setHasPalette] = useState(false)
  useEffect(() => {
    const p = config?.palette
    if (p) {
      setPalette({
        primary: p.primary || DEFAULT_FALLBACK_PALETTE.primary,
        secondary: p.secondary || DEFAULT_FALLBACK_PALETTE.secondary,
        accent: p.accent || DEFAULT_FALLBACK_PALETTE.accent,
        background: p.background || DEFAULT_FALLBACK_PALETTE.background,
        text: p.text || DEFAULT_FALLBACK_PALETTE.text,
      })
      setHasPalette(true)
    } else {
      setHasPalette(false)
    }
  }, [config?.palette])

  const [minPages, setMinPages] = useState<string>("")
  const [maxPages, setMaxPages] = useState<string>("")
  useEffect(() => {
    setMinPages(config?.ebook_min_pages ? String(config.ebook_min_pages) : "")
    setMaxPages(config?.ebook_max_pages ? String(config.ebook_max_pages) : "")
  }, [config?.ebook_min_pages, config?.ebook_max_pages])

  const [instructions, setInstructions] = useState("")

  // Jobs en cours (statuts via le canal + polling de secours).
  const [coverJobId, setCoverJobId] = useState<string | null>(null)
  const [ebookJobId, setEbookJobId] = useState<string | null>(null)
  const [postersJobId, setPostersJobId] = useState<string | null>(null)
  const [salesJobId, setSalesJobId] = useState<string | null>(null)
  const { data: coverJob } = useJob(coverJobId)
  const { data: ebookJob } = useJob(ebookJobId)
  const { data: postersJob } = useJob(postersJobId)
  const { data: salesJob } = useJob(salesJobId)

  useEffect(() => {
    const entries = [
      { job: coverJob, done: () => setCoverJobId(null) },
      { job: ebookJob, done: () => setEbookJobId(null) },
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
  }, [coverJob, ebookJob, postersJob, salesJob, refetch, t])

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

  const handleSavePalette = () => {
    const input: ProjectPalette = {
      primary: palette.primary, secondary: palette.secondary, accent: palette.accent,
      background: palette.background, text: palette.text,
    }
    updateConfig.mutate(
      { palette: input },
      { onError, onSuccess: () => toast.success(t("projects:saved")) },
    )
  }

  const handleSavePages = () => {
    updateConfig.mutate(
      {
        ebook_min_pages: minPages ? Number(minPages) : undefined,
        ebook_max_pages: maxPages ? Number(maxPages) : undefined,
      },
      { onError, onSuccess: () => toast.success(t("projects:saved")) },
    )
  }

  const handleGenerateCover = () => {
    generateCover.mutate(instructions.trim() || undefined, {
      onSuccess: (j) => setCoverJobId(j.id),
      onError,
    })
  }

  if (isLoading) return <LoadingState label={t("common.loading")} />
  if (isError || !project) return <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />

  const hasAsset = (kind: string): boolean => !!assets?.some((a) => a.kind === kind)

  const coverBusy = generateCover.isPending || isActive(coverJob)
  const ebookBusy = generateEbook.isPending || isActive(ebookJob)
  const postersBusy = generatePosters.isPending || isActive(postersJob)
  const salesBusy = generateSalesPage.isPending || isActive(salesJob)

  const steps = [
    { n: 1, label: t("projects:stepCover"), done: hasCover, locked: false },
    { n: 2, label: t("projects:ebook"), done: hasAsset("ebook_pdf"), locked: !hasCover },
    { n: 3, label: t("projects:posters"), done: hasAsset("poster"), locked: !hasCover },
    { n: 4, label: t("projects:salesPage"), done: hasAsset("sales_page"), locked: !hasCover },
    { n: 5, label: t("projects:videoTitle"), done: hasAsset("video_ad"), locked: !hasCover },
  ]

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

      {/* Stepper cover-first */}
      <nav className="flex flex-wrap items-center gap-x-6 gap-y-2">
        {steps.map((s) => (
          <div key={s.n} className="flex items-center gap-2">
            <StepBadge step={s.n} done={s.done} locked={s.locked} />
            <span className={s.done ? "text-sm font-medium text-primary" : "text-sm text-muted-foreground"}>
              {s.label}
            </span>
          </div>
        ))}
      </nav>

      {/* Étape 1 : identité visuelle + couverture */}
      <section>
        <h2 className="mb-3 flex items-center gap-2 font-display text-lg font-semibold text-primary">
          <PaletteIcon className="h-5 w-5" />
          {t("projects:visualTitle")}
        </h2>
        <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
          <Card>
            <CardContent className="flex h-full flex-col items-center justify-center gap-3 p-5">
              {coverURL ? (
                <img src={coverURL} alt={project.title} className="w-full rounded-lg border shadow-sm" />
              ) : (
                <div className="flex aspect-[3/4] w-full items-center justify-center rounded-lg border border-dashed text-muted-foreground">
                  <ImageIcon className="h-8 w-8" />
                </div>
              )}
              <p className="text-xs text-muted-foreground">
                {hasCover ? t("projects:coverReady") : t("projects:coverNone")}
              </p>
            </CardContent>
          </Card>

          <div className="space-y-4">
            <Card>
              <CardContent className="space-y-4 p-5">
                <div className="flex items-start justify-between gap-2">
                  <h3 className="font-semibold">{t("projects:paletteTitle")}</h3>
                  {hasPalette && !coverBusy ? (
                    <Badge variant="secondary" className="gap-1 text-[10px]">
                      <Sparkles className="h-3 w-3" />
                      {t("projects:aiProposed")}
                    </Badge>
                  ) : null}
                </div>
                <p className="text-xs text-muted-foreground">{t("projects:visualHint")}</p>

                <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
                  {PALETTE_FIELDS.map((field) => (
                    <label key={field} className="space-y-1.5">
                      <span className="block text-xs text-muted-foreground">{t(`projects:color.${field}`)}</span>
                      <input
                        type="color"
                        value={palette[field]}
                        onChange={(e) => setPalette((p) => ({ ...p, [field]: e.target.value }))}
                        className="h-9 w-full cursor-pointer rounded-md border border-input bg-transparent p-1"
                        aria-label={t(`projects:color.${field}`)}
                      />
                    </label>
                  ))}
                </div>
                <Button size="sm" variant="outline" onClick={handleSavePalette} loading={updateConfig.isPending}>
                  <Save className="h-4 w-4" />
                  {t("projects:applyColors")}
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="space-y-3 p-5">
                <h3 className="font-semibold">{t("projects:coverTitle")}</h3>
                <Textarea
                  rows={2}
                  value={instructions}
                  placeholder={t("projects:feedbackPlaceholder")}
                  onChange={(e) => setInstructions(e.target.value)}
                  className="text-sm"
                />
                <Button size="touch" onClick={handleGenerateCover} disabled={coverBusy} loading={coverBusy}>
                  <RefreshCw className="h-4 w-4" />
                  {hasCover ? t("projects:regenerateCover") : t("projects:generateCover")}
                </Button>
                {hasCover ? (
                  <p className="text-xs text-muted-foreground">{t("projects:regenHint")}</p>
                ) : (
                  <p className="text-xs text-muted-foreground">{t("projects:coverHint")}</p>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* Config ebook + assets verrouillés sans cover */}
      <section>
        <h2 className="mb-3 font-display text-lg font-semibold text-primary">{t("projects:assetsTitle")}</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card className={hasCover ? "" : "opacity-70"}>
            <CardContent className="flex h-full flex-col gap-3 p-5">
              <div className="flex items-center justify-between">
                <BookOpen className="h-6 w-6 text-primary" />
                {!hasCover ? <Lock className="h-4 w-4 text-muted-foreground" /> : null}
              </div>
              <h3 className="font-semibold">{t("projects:ebook")}</h3>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <Input
                  type="number"
                  min={2}
                  max={40}
                  value={minPages}
                  placeholder="6"
                  onChange={(e) => setMinPages(e.target.value)}
                  className="h-8 w-16"
                  aria-label={t("projects:minPages")}
                />
                <span>→</span>
                <Input
                  type="number"
                  min={2}
                  max={40}
                  value={maxPages}
                  placeholder="14"
                  onChange={(e) => setMaxPages(e.target.value)}
                  className="h-8 w-16"
                  aria-label={t("projects:maxPages")}
                />
                <Button
                  size="sm"
                  variant="ghost"
                  className="h-8 px-2"
                  onClick={handleSavePages}
                  loading={updateConfig.isPending}
                  aria-label={t("projects:save")}
                >
                  <Save className="h-3.5 w-3.5" />
                </Button>
              </div>
              <Button
                size="sm"
                disabled={ebookBusy}
                loading={ebookBusy}
                onClick={() => generateEbook.mutate(id, { onSuccess: (j) => setEbookJobId(j.id), onError })}
              >
                {hasAsset("ebook_pdf") ? t("projects:regenerate") : t("projects:generate")}
              </Button>
              {!hasCover ? <p className="text-xs text-muted-foreground">{t("projects:coverRequired")}</p> : null}
            </CardContent>
          </Card>

          {[
            { icon: Images, title: t("projects:posters"), busy: postersBusy, job: postersJobId, setJob: setPostersJobId, kind: "poster", gen: generatePosters },
            { icon: Megaphone, title: t("projects:salesPage"), busy: salesBusy, job: salesJobId, setJob: setSalesJobId, kind: "sales_page", gen: generateSalesPage },
          ].map((card) => (
            <Card key={card.kind} className={hasCover ? "" : "opacity-70"}>
              <CardContent className="flex h-full flex-col gap-3 p-5">
                <div className="flex items-center justify-between">
                  <card.icon className="h-6 w-6 text-primary" />
                  {!hasCover ? <Lock className="h-4 w-4 text-muted-foreground" /> : null}
                </div>
                <h3 className="font-semibold">{card.title}</h3>
                <div className="flex-1" />
                <Button
                  size="sm"
                  disabled={card.busy}
                  loading={card.busy}
                  onClick={() => card.gen.mutate(id, { onSuccess: (j) => card.setJob(j.id), onError })}
                >
                  {hasAsset(card.kind) ? t("projects:regenerate") : t("projects:generate")}
                </Button>
                {!hasCover ? <p className="text-xs text-muted-foreground">{t("projects:coverRequired")}</p> : null}
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      {/* Étape 5 : vidéo publicitaire (job video_ad, ADR-016) */}
      <section>
        <h2 className="mb-3 flex items-center gap-2 font-display text-lg font-semibold text-primary">
          <Clapperboard className="h-5 w-5" />
          {t("projects:videoSectionTitle")}
        </h2>
        <div className="grid gap-4 lg:grid-cols-[1fr_360px]">
          <div>
            {videoAsset ? (
              <VideoPreview asset={videoAsset} />
            ) : (
              <p className="text-sm text-muted-foreground">{t("projects:videoNone")}</p>
            )}
          </div>
          <VideoAdsPanel projectId={id} locked={!hasCover} />
        </div>
      </section>

      <section>
        <h2 className="mb-3 font-display text-lg font-semibold text-primary">{t("projects:files")}</h2>
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
