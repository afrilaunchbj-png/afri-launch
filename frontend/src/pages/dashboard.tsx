import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts"
import { Coins, FolderOpen, MessageCircle, LineChart, Workflow } from "lucide-react"

import { fillDailySeries, fillWeeklySeries } from "@/features/dashboard/api"
import { useDashboardStats } from "@/features/dashboard/hooks"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/features/auth"
import { useCreditsSummary } from "@/features/credits/hooks"

function StatCard({ icon: Icon, label, value }: { icon: typeof FolderOpen; label: string; value: number }) {
  return (
    <Card>
      <CardContent className="flex items-center gap-3 p-4">
        <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
          <Icon className="h-5 w-5" />
        </span>
        <div className="min-w-0">
          <p className="font-display text-xl font-bold leading-none">{value.toLocaleString("fr-FR")}</p>
          <p className="mt-1 truncate text-xs text-muted-foreground">{label}</p>
        </div>
      </CardContent>
    </Card>
  )
}

function formatDateLabel(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { day: "numeric", month: "short" })
}

export default function DashboardPage() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const { data: credits, isLoading: creditsLoading } = useCreditsSummary()
  const { data: stats, isLoading: statsLoading } = useDashboardStats()

  const activeJobs = (stats?.jobs_by_status.pending ?? 0) + (stats?.jobs_by_status.processing ?? 0)
  const creditsSeries = fillDailySeries(stats?.credits_per_day ?? [])
  const projectsSeries = fillWeeklySeries(stats?.projects_per_week ?? [])

  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div>
          <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">
            {t("dashboard:welcome", { name: user?.name ?? "" })}
          </h1>
          <p className="mt-1 text-muted-foreground">{t("dashboard:subtitle")}</p>
        </div>
        <Card className="md:w-auto">
          <CardContent className="flex items-center gap-4 p-4">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-accent text-accent-foreground">
              <Coins className="h-6 w-6" />
            </div>
            <div>
              <p className="text-xs uppercase tracking-wider text-muted-foreground">{t("credits:balance")}</p>
              {creditsLoading ? (
                <Skeleton className="h-7 w-24" />
              ) : (
                <p className="font-display text-xl font-bold text-warning">
                  {credits?.summary.available ?? 0} {t("credits:label")}
                </p>
              )}
            </div>
          </CardContent>
        </Card>
      </header>

      <section>
        <h2 className="mb-4 font-display text-lg font-semibold text-primary">{t("dashboard:statsTitle")}</h2>
        {statsLoading ? (
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-[72px] rounded-xl" />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
            <StatCard icon={FolderOpen} label={t("dashboard:statProjects")} value={stats?.projects ?? 0} />
            <StatCard icon={MessageCircle} label={t("dashboard:statConversations")} value={stats?.conversations ?? 0} />
            <StatCard icon={Workflow} label={t("dashboard:statActiveJobs")} value={activeJobs} />
            <StatCard icon={Coins} label={t("dashboard:statCreditsUsed")} value={stats?.credits_used_30d ?? 0} />
          </div>
        )}
      </section>

      <section className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("dashboard:creditsChartTitle")}</CardTitle>
            <CardDescription>{t("dashboard:creditsChartDesc")}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={creditsSeries} margin={{ top: 8, right: 8, left: -16, bottom: 0 }}>
                  <defs>
                    <linearGradient id="creditsGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.25} />
                      <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" vertical={false} />
                  <XAxis
                    dataKey="date"
                    tickFormatter={formatDateLabel}
                    tick={{ fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    minTickGap={32}
                  />
                  <YAxis tick={{ fontSize: 11 }} tickLine={false} axisLine={false} allowDecimals={false} />
                  <Tooltip
                    labelFormatter={(label) => formatDateLabel(String(label))}
                    formatter={(value) => [value, t("dashboard:creditsChartTitle")]}
                  />
                  <Area
                    type="monotone"
                    dataKey="value"
                    stroke="hsl(var(--primary))"
                    strokeWidth={2}
                    fill="url(#creditsGradient)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("dashboard:projectsChartTitle")}</CardTitle>
            <CardDescription>{t("dashboard:projectsChartDesc")}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={projectsSeries} margin={{ top: 8, right: 8, left: -16, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" vertical={false} />
                  <XAxis
                    dataKey="date"
                    tickFormatter={formatDateLabel}
                    tick={{ fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    minTickGap={24}
                  />
                  <YAxis tick={{ fontSize: 11 }} tickLine={false} axisLine={false} allowDecimals={false} />
                  <Tooltip
                    labelFormatter={(label) => formatDateLabel(String(label))}
                    formatter={(value) => [value, t("dashboard:projectsChartTitle")]}
                  />
                  <Bar dataKey="value" fill="hsl(var(--primary))" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </section>

      <section>
        <h2 className="mb-4 font-display text-lg font-semibold text-primary">{t("dashboard:quickActions")}</h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <Card className="transition-shadow hover:shadow-md">
            <CardHeader>
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted text-primary">
                <LineChart className="h-6 w-6" />
              </div>
              <CardTitle>{t("dashboard:searchOpportunity")}</CardTitle>
              <CardDescription>{t("dashboard:searchOpportunityDesc")}</CardDescription>
            </CardHeader>
            <CardContent>
              <Button asChild size="touch" className="w-full sm:w-auto">
                <Link to="/discover">{t("dashboard:start")}</Link>
              </Button>
            </CardContent>
          </Card>

          <Card className="transition-shadow hover:shadow-md">
            <CardHeader>
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted text-primary">
                <Coins className="h-6 w-6" />
              </div>
              <CardTitle>{t("credits:title")}</CardTitle>
              <CardDescription>{t("dashboard:creditsDesc")}</CardDescription>
            </CardHeader>
            <CardContent>
              <Button asChild size="touch" variant="outline" className="w-full sm:w-auto">
                <Link to="/credits">{t("credits:viewHistory")}</Link>
              </Button>
            </CardContent>
          </Card>
        </div>
      </section>
    </div>
  )
}
