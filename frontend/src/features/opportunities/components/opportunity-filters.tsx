import { useTranslation } from "react-i18next"
import { RotateCcw, SlidersHorizontal } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { useFacets } from "@/features/opportunities/hooks"
import type { Difficulty } from "@/features/opportunities/types"
import { cn } from "@/lib/utils"

const ALL = "all"

interface OpportunityFiltersProps {
  country: string
  sector: string
  difficulty: string
  onCountryChange: (v: string) => void
  onSectorChange: (v: string) => void
  onDifficultyChange: (v: string) => void
  onReset: () => void
}

const difficulties: Difficulty[] = ["low", "medium", "high"]

export function OpportunityFilters({
  country,
  sector,
  difficulty,
  onCountryChange,
  onSectorChange,
  onDifficultyChange,
  onReset,
}: OpportunityFiltersProps) {
  const { t } = useTranslation()
  const { data: facets } = useFacets()

  const hasFilters = country !== "" || sector !== "" || difficulty !== ""

  return (
    <div className="space-y-5">
      <div className="flex items-center gap-2 border-b pb-3">
        <SlidersHorizontal className="h-4 w-4 text-primary" />
        <h3 className="text-sm font-semibold">{t("opportunities:refine")}</h3>
        {hasFilters ? (
          <Button
            variant="ghost"
            size="sm"
            className="ml-auto text-xs text-muted-foreground"
            onClick={onReset}
          >
            <RotateCcw className="h-3.5 w-3.5" />
            {t("opportunities:reset")}
          </Button>
        ) : null}
      </div>

      <div className="space-y-1.5">
        <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {t("opportunities:country")}
        </label>
        <Select value={country || ALL} onValueChange={(v) => onCountryChange(v === ALL ? "" : v)}>
          <SelectTrigger className="h-11 w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>{t("opportunities:all")}</SelectItem>
            {facets?.countries.map((c) => (
              <SelectItem key={c} value={c}>
                {c}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1.5">
        <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {t("opportunities:sector")}
        </label>
        <Select value={sector || ALL} onValueChange={(v) => onSectorChange(v === ALL ? "" : v)}>
          <SelectTrigger className="h-11 w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>{t("opportunities:all")}</SelectItem>
            {facets?.sectors.map((s) => (
              <SelectItem key={s} value={s}>
                {s}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <Separator />

      <div className="space-y-2">
        <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {t("opportunities:difficultyLabel")}
        </label>
        <div className="flex flex-wrap gap-2">
          <Button
            variant={difficulty === "" ? "default" : "outline"}
            size="sm"
            className="rounded-full"
            onClick={() => onDifficultyChange("")}
          >
            {t("opportunities:all")}
          </Button>
          {difficulties.map((d) => (
            <Button
              key={d}
              variant={difficulty === d ? "default" : "outline"}
              size="sm"
              className={cn("rounded-full", difficulty !== d && "text-muted-foreground")}
              onClick={() => onDifficultyChange(d)}
            >
              {t(`opportunities:difficulty.${d}`)}
            </Button>
          ))}
        </div>
      </div>
    </div>
  )
}
