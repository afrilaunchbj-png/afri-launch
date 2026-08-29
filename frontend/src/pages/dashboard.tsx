import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import { Coins, LineChart } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/features/auth"
import { useCreditsSummary } from "@/features/credits/hooks"

export default function DashboardPage() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const { data, isLoading } = useCreditsSummary()

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
              {isLoading ? (
                <Skeleton className="h-7 w-24" />
              ) : (
                <p className="font-display text-xl font-bold text-warning">
                  {data?.summary.available ?? 0} {t("credits:label")}
                </p>
              )}
            </div>
          </CardContent>
        </Card>
      </header>

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
                <Link to="/opportunities">{t("dashboard:start")}</Link>
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
