import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import { ArrowRight } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function HomePage() {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-8 py-8">
      <section className="text-center">
        <h1 className="font-display text-3xl font-bold text-primary sm:text-4xl">{t("home.title")}</h1>
        <p className="mx-auto mt-3 max-w-2xl text-muted-foreground">{t("tagline")}</p>
        <Button asChild size="touch" className="mt-6">
          <Link to="/dashboard">
            {t("home.cta")} <ArrowRight className="h-4 w-4" />
          </Link>
        </Button>
      </section>
      <section className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>{t("home.research.title")}</CardTitle>
            <CardDescription>{t("home.research.description")}</CardDescription>
          </CardHeader>
          <CardContent />
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{t("home.creation.title")}</CardTitle>
            <CardDescription>{t("home.creation.description")}</CardDescription>
          </CardHeader>
          <CardContent />
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{t("home.marketing.title")}</CardTitle>
            <CardDescription>{t("home.marketing.description")}</CardDescription>
          </CardHeader>
          <CardContent />
        </Card>
      </section>
    </div>
  )
}
