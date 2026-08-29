import { useEffect, useState } from "react"
import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import { Globe, Search, SearchX, SlidersHorizontal } from "lucide-react"
import { toast } from "sonner"

import { EmptyState } from "@/components/states/empty-state"
import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Pagination } from "@/components/pagination"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { useJob } from "@/features/generation/hooks"
import { OpportunityCard } from "@/features/opportunities/components/opportunity-card"
import { OpportunityFilters } from "@/features/opportunities/components/opportunity-filters"
import { useOpportunities } from "@/features/opportunities/hooks"
import type { OpportunityFilters as Filters } from "@/features/opportunities/types"
import { useStartResearch } from "@/features/research/hooks"
import { isAppError } from "@/lib/errors"

const PAGE_SIZE = 10

const MARKETS = [
  { code: "SN", name: "Sénégal" },
  { code: "CI", name: "Côte d'Ivoire" },
  { code: "BJ", name: "Bénin" },
  { code: "NG", name: "Nigeria" },
  { code: "KE", name: "Kenya" },
  { code: "GH", name: "Ghana" },
]

export default function OpportunitiesPage() {
  const { t, i18n } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  const [searchDraft, setSearchDraft] = useState(searchParams.get("q") ?? "")

  const country = searchParams.get("country") ?? ""
  const sector = searchParams.get("sector") ?? ""
  const difficulty = searchParams.get("difficulty") ?? ""
  const q = searchParams.get("q") ?? ""
  const page = Math.max(1, Number(searchParams.get("page") ?? "1"))

  const filters: Filters = { country, sector, difficulty: difficulty as Filters["difficulty"], q, page, pageSize: PAGE_SIZE }

  const { data, isLoading, isError, refetch, isFetching } = useOpportunities(filters)

  const research = useStartResearch()
  const [researchQuery, setResearchQuery] = useState("")
  const [researchSector, setResearchSector] = useState("")
  const [researchMarkets, setResearchMarkets] = useState<string[]>([])
  const [researchJobId, setResearchJobId] = useState<string | null>(null)
  const { data: researchJob } = useJob(researchJobId)

  useEffect(() => {
    if (!researchJob) return
    if (researchJob.status === "completed") {
      toast.success(t("opportunities:researchDone"))
      refetch()
      setResearchJobId(null)
    } else if (researchJob.status === "failed") {
      toast.error(t("common.generationFailed", { error: researchJob.error ?? "" }))
      setResearchJobId(null)
    }
  }, [researchJob, refetch, t])

  const updateParams = (patch: Record<string, string>) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      for (const [key, value] of Object.entries(patch)) {
        if (value === "") {
          next.delete(key)
        } else {
          next.set(key, value)
        }
      }
      if (!("page" in patch)) {
        next.delete("page")
      }
      return next
    })
  }

  const resetFilters = () => {
    setSearchDraft("")
    setSearchParams({})
  }

  useEffect(() => {
    const timer = setTimeout(() => {
      updateParams({ q: searchDraft })
    }, 400)
    return () => clearTimeout(timer)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchDraft])

  const toggleMarket = (code: string) => {
    setResearchMarkets((prev) => (prev.includes(code) ? prev.filter((m) => m !== code) : [...prev, code]))
  }

  const handleResearch = () => {
    if (!researchQuery.trim()) return
    research.mutate(
      { query: researchQuery.trim(), sector: researchSector.trim(), markets: researchMarkets, language: i18n.language },
      {
        onSuccess: (j) => setResearchJobId(j.id),
        onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
      },
    )
  }

  const researching = research.isPending || (researchJob && (researchJob.status === "pending" || researchJob.status === "processing"))

  const totalPages = data?.pagination.totalPages ?? 0

  const filterProps = {
    country,
    sector,
    difficulty,
    onCountryChange: (v: string) => updateParams({ country: v }),
    onSectorChange: (v: string) => updateParams({ sector: v }),
    onDifficultyChange: (v: string) => updateParams({ difficulty: v }),
    onReset: resetFilters,
  }

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">
            {t("opportunities:title")}
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-muted-foreground">{t("opportunities:subtitle")}</p>
        </div>
        <div className="relative w-full lg:w-96">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="h-11 pl-10"
            placeholder={t("opportunities:searchPlaceholder")}
            value={searchDraft}
            onChange={(e) => setSearchDraft(e.target.value)}
          />
        </div>
      </header>

      <Sheet>
        <SheetTrigger asChild>
          <Button variant="outline" size="sm">
            <Globe className="h-4 w-4" />
            {t("opportunities:exploreOnline")}
          </Button>
        </SheetTrigger>
        <SheetContent side="right" className="overflow-y-auto">
          <SheetHeader>
            <SheetTitle>{t("opportunities:researchTitle")}</SheetTitle>
            <SheetDescription>{t("opportunities:researchDesc")}</SheetDescription>
          </SheetHeader>
          <div className="mt-4 space-y-4">
            <div className="space-y-2">
              <Label htmlFor="research-query">{t("opportunities:queryLabel")}</Label>
              <Input
                id="research-query"
                placeholder={t("opportunities:queryPlaceholder")}
                value={researchQuery}
                onChange={(e) => setResearchQuery(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="research-sector">{t("opportunities:sectorLabel")}</Label>
              <Input
                id="research-sector"
                placeholder={t("opportunities:sectorPlaceholder")}
                value={researchSector}
                onChange={(e) => setResearchSector(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>{t("opportunities:marketsLabel")}</Label>
              <div className="grid grid-cols-2 gap-2">
                {MARKETS.map((m) => (
                  <label key={m.code} className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={researchMarkets.includes(m.code)}
                      onCheckedChange={() => toggleMarket(m.code)}
                    />
                    {m.name}
                  </label>
                ))}
              </div>
            </div>
            <Button size="touch" className="w-full" onClick={handleResearch} disabled={researching || !researchQuery.trim() || researchMarkets.length === 0} loading={researching}>
              {researching ? t("opportunities:researching") : t("opportunities:launchResearch")}
            </Button>
          </div>
        </SheetContent>
      </Sheet>

      <div className="grid gap-6 lg:grid-cols-12">
        <aside className="hidden lg:col-span-3 lg:block">
          <Card className="p-5">
            <OpportunityFilters {...filterProps} />
          </Card>
        </aside>

        <div className="lg:col-span-9">
          <div className="mb-4 flex items-center justify-between lg:hidden">
            <Sheet>
              <SheetTrigger asChild>
                <Button variant="outline" size="sm">
                  <SlidersHorizontal className="h-4 w-4" />
                  {t("opportunities:filters")}
                </Button>
              </SheetTrigger>
              <SheetContent side="left">
                <SheetHeader>
                  <SheetTitle>{t("opportunities:filters")}</SheetTitle>
                  <SheetDescription>{t("opportunities:refine")}</SheetDescription>
                </SheetHeader>
                <div className="mt-4">
                  <OpportunityFilters {...filterProps} />
                </div>
              </SheetContent>
            </Sheet>
            <span className="text-xs text-muted-foreground">
              {t("opportunities:resultCount", { count: data?.pagination.totalItems ?? 0 })}
            </span>
          </div>

          {isLoading ? (
            <LoadingState label={t("common.loading")} />
          ) : isError ? (
            <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />
          ) : !data || data.data.length === 0 ? (
            <EmptyState
              icon={SearchX}
              title={t("opportunities:empty")}
              description={t("opportunities:emptyDesc")}
            />
          ) : (
            <div className="space-y-6">
              {isFetching && !isLoading ? (
                <p className="text-right text-xs text-muted-foreground">{t("common.refreshing")}</p>
              ) : null}
              {data.data.map((opp) => (
                <OpportunityCard key={opp.id} opportunity={opp} />
              ))}
              <Pagination page={page} totalPages={totalPages} onPageChange={(p) => updateParams({ page: String(p) })} />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
