import { useEffect, useMemo, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Clapperboard, Download, Lock } from "lucide-react"
import { toast } from "sonner"

import { api } from "@/lib/api/client"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { assetDownloadPath, type Asset } from "@/features/projects/api"
import { downloadAsset } from "@/features/projects/hooks"
import { useAppEvent } from "@/lib/api/events"
import { isAppError } from "@/lib/errors"

import { useGenerateVideoAd, useVideoJob } from "./hooks"
import type { JobUpdatedEvent } from "./api"

const STAGES = ["analyzing", "storyboarding", "generating_avatar", "rendering"] as const
type Stage = (typeof STAGES)[number]

function isActive(status?: string) {
  return status === "pending" || status === "processing"
}

export function VideoAdsPanel({ projectId, locked }: { projectId: string; locked: boolean }) {
  const { t } = useTranslation()
  const generate = useGenerateVideoAd(projectId)

  const [duration, setDuration] = useState("30")
  const [ratio, setRatio] = useState("9:16")
  const [instructions, setInstructions] = useState("")
  const [jobId, setJobId] = useState<string | null>(null)
  const [stage, setStage] = useState<Stage | null>(null)
  const jobRef = useRef(jobId)
  jobRef.current = jobId

  const { data: job } = useVideoJob(jobId)
  const busy = generate.isPending || isActive(job?.status)

  // Progression temps réel : les events job.updated portent le champ stage.
  useAppEvent("job.updated", (raw) => {
    const ev = raw as unknown as JobUpdatedEvent
    if (!ev || ev.id !== jobRef.current || ev.kind !== "video_ad") return
    if (ev.stage && (STAGES as readonly string[]).includes(ev.stage)) {
      setStage(ev.stage as Stage)
    }
  })

  useEffect(() => {
    if (!job) return
    if (job.status === "completed") {
      setJobId(null)
      setStage(null)
    } else if (job.status === "failed") {
      toast.error(t("common.generationFailed", { error: job.error ?? "" }))
      setJobId(null)
      setStage(null)
    }
  }, [job, t])

  const handleGenerate = () => {
    setStage("analyzing")
    generate.mutate(
      {
        duration: Number(duration),
        aspect_ratio: ratio as "9:16" | "1:1" | "16:9",
        instructions: instructions.trim() || undefined,
      },
      {
        onSuccess: (j) => setJobId(j.id),
        onError: (error) => {
          setStage(null)
          toast.error(isAppError(error) ? error.message : t("common.genericError"))
        },
      },
    )
  }

  const currentStep = stage ? STAGES.indexOf(stage) : -1

  return (
    <Card className={locked ? "opacity-70" : ""}>
      <CardContent className="flex h-full flex-col gap-4 p-5">
        <div className="flex items-center justify-between">
          <h3 className="font-semibold">{t("projects:videoTitle")}</h3>
          {busy ? <Badge variant="secondary">{t("projects:videoProcessing")}</Badge> : null}
        </div>
        <p className="text-xs text-muted-foreground">{t("projects:videoHint")}</p>

        <div className="flex flex-wrap items-center gap-3">
          <label className="space-y-1.5">
            <span className="block text-xs text-muted-foreground">{t("projects:videoDurationLabel")}</span>
            <Select value={duration} onValueChange={setDuration} disabled={busy}>
              <SelectTrigger className="h-9 w-[110px]" aria-label={t("projects:videoDurationLabel")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {["15", "30", "45", "60"].map((d) => (
                  <SelectItem key={d} value={d}>
                    {t("projects:videoDurationValue", { count: Number(d) })}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </label>

          <label className="space-y-1.5">
            <span className="block text-xs text-muted-foreground">{t("projects:videoRatioLabel")}</span>
            <Select value={ratio} onValueChange={setRatio} disabled={busy}>
              <SelectTrigger className="h-9 w-[140px]" aria-label={t("projects:videoRatioLabel")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="9:16">{t("projects:videoRatioPortrait")}</SelectItem>
                <SelectItem value="1:1">{t("projects:videoRatioSquare")}</SelectItem>
                <SelectItem value="16:9">{t("projects:videoRatioLandscape")}</SelectItem>
              </SelectContent>
            </Select>
          </label>
        </div>

        <Textarea
          rows={2}
          value={instructions}
          placeholder={t("projects:videoInstructionsPlaceholder")}
          onChange={(e) => setInstructions(e.target.value)}
          className="text-sm"
          disabled={busy}
        />

        {/* Progression multi-étapes (cf. ADR-016) */}
        {busy && stage ? (
          <ol className="space-y-1.5 rounded-lg border bg-muted/40 p-3 text-sm">
            {STAGES.map((s, i) => (
              <li key={s} className="flex items-center gap-2">
                <span
                  className={
                    i < currentStep
                      ? "text-primary"
                      : i === currentStep
                        ? "text-foreground"
                        : "text-muted-foreground/50"
                  }
                >
                  {i < currentStep ? "✓" : i === currentStep ? "⏳" : "○"}
                </span>
                <span className={i <= currentStep ? "text-foreground" : "text-muted-foreground/60"}>
                  {t(`projects:videoStage.${s}`)}
                </span>
              </li>
            ))}
          </ol>
        ) : null}

        <div className="mt-auto">
          <Button size="sm" disabled={busy || locked} loading={busy} onClick={handleGenerate}>
            <Clapperboard className="h-4 w-4" />
            {t("projects:generateVideo")}
          </Button>
          {locked ? (
            <p className="mt-2 flex items-center gap-1 text-xs text-muted-foreground">
              <Lock className="h-3 w-3" />
              {t("projects:coverRequired")}
            </p>
          ) : null}
        </div>
      </CardContent>
    </Card>
  )
}

/** Aperçu + téléchargement de la dernière vidéo publicitaire générée. */
export function VideoPreview({ asset }: { asset: Asset }) {
  const { t } = useTranslation()
  const [videoURL, setVideoURL] = useState<string | null>(null)

  useEffect(() => {
    let objectURL: string | null = null
    api
      .download(assetDownloadPath(asset.id))
      .then((blob) => {
        objectURL = URL.createObjectURL(blob)
        setVideoURL(objectURL)
      })
      .catch(() => setVideoURL(null))
    return () => {
      if (objectURL) URL.revokeObjectURL(objectURL)
    }
  }, [asset.id])

  if (!videoURL) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        {t("projects:videoLoadingPreview")}
      </div>
    )
  }
  return (
    <div className="flex flex-col gap-2">
      {/* eslint-disable-next-line jsx-a11y/media-has-caption -- vidéo générée, sans piste de dialogue sous-titrée */}
      <video src={videoURL} controls className="max-h-[420px] rounded-lg border bg-black" />
      <Button
        size="sm"
        variant="outline"
        className="w-fit"
        onClick={() => downloadAsset(asset.id, asset.filename).catch(() => toast.error(t("common.genericError")))}
      >
        <Download className="h-4 w-4" />
        {t("projects:download")}
      </Button>
    </div>
  )
}

/** Hook utilitaire : dernière vidéo pub dans la liste d'assets du projet. */
export function useLatestVideoAsset(assets: Asset[] | undefined) {
  return useMemo(
    () => assets?.filter((a) => a.kind === "video_ad").slice(-1)[0] ?? null,
    [assets],
  )
}
