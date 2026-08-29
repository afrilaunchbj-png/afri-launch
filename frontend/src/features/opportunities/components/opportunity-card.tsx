import { useTranslation } from "react-i18next"
import { Bookmark, MapPin, TrendingUp } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { useToggleSave } from "@/features/opportunities/hooks"
import type { Opportunity, Signal } from "@/features/opportunities/types"
import { cn } from "@/lib/utils"

function signalVariant(signal: Signal): "success" | "warning" | "secondary" | "outline" {
  switch (signal) {
    case "verified":
      return "success"
    case "estimated":
      return "warning"
    case "inferred":
      return "secondary"
    default:
      return "outline"
  }
}

function scoreColor(score: number) {
  if (score >= 75) return "text-success"
  if (score >= 50) return "text-warning"
  return "text-muted-foreground"
}

function MetricBar({ label, value, tone }: { label: string; value: number; tone: "success" | "warning" | "primary" }) {
  const toneClass = {
    success: "bg-success",
    warning: "bg-warning",
    primary: "bg-primary",
  }[tone]
  return (
    <div>
      <div className="mb-1 flex justify-between text-xs">
        <span className="text-foreground">{label}</span>
        <span className="font-semibold">{value}</span>
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
        <div className={cn("h-full rounded-full", toneClass)} style={{ width: `${value}%` }} />
      </div>
    </div>
  )
}

export function OpportunityCard({ opportunity }: { opportunity: Opportunity }) {
  const { t } = useTranslation()
  const toggle = useToggleSave(opportunity.id)

  return (
    <Card className="overflow-hidden">
      <CardContent className="p-6 md:p-8">
        <div className="flex flex-col gap-6 md:flex-row">
          <div className="flex-1">
            <div className="mb-3 flex flex-wrap items-center gap-2">
              <Badge variant={signalVariant(opportunity.signal)}>
                {t(`opportunities:signal.${opportunity.signal}`)}
              </Badge>
              <span className="flex items-center gap-1 text-xs text-muted-foreground">
                <MapPin className="h-3.5 w-3.5" />
                {opportunity.country}
              </span>
              <span className="text-xs text-muted-foreground">· {opportunity.sector}</span>
            </div>

            <h3 className="mb-2 font-display text-lg font-semibold text-primary">{opportunity.title}</h3>
            <p className="mb-4 line-clamp-2 text-sm text-muted-foreground">{opportunity.summary}</p>

            <div className="flex items-center justify-end gap-3">
              <Button
                variant="outline"
                size="sm"
                onClick={() => toggle.mutate(opportunity.is_saved)}
                disabled={toggle.isPending}
                aria-label={opportunity.is_saved ? t("opportunities:unsave") : t("opportunities:save")}
              >
                <Bookmark className={cn("h-4 w-4", opportunity.is_saved && "fill-current")} />
                {opportunity.is_saved ? t("opportunities:saved") : t("opportunities:save")}
              </Button>
            </div>
          </div>

          <div className="flex flex-col justify-center rounded-xl border bg-muted/40 p-5 md:w-72">
            <div className="mb-6 flex items-end justify-between">
              <div>
                <p className="text-xs uppercase tracking-wider text-muted-foreground">
                  {t("opportunities:score")}
                </p>
                <div className={cn("font-display text-4xl font-bold leading-none", scoreColor(opportunity.score))}>
                  {opportunity.score}
                  <span className="text-base font-normal text-muted-foreground">/100</span>
                </div>
              </div>
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-success/20 text-success">
                <TrendingUp className="h-6 w-6" />
              </div>
            </div>
            <div className="space-y-3">
              <MetricBar label={t("opportunities:demand")} value={opportunity.scores.demand} tone="success" />
              <MetricBar label={t("opportunities:pain")} value={opportunity.scores.pain} tone="warning" />
              <MetricBar label={t("opportunities:competition")} value={opportunity.scores.competition} tone="primary" />
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
