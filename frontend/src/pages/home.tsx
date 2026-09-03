import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import {
  ArrowRight,
  BookOpen,
  Download,
  Lightbulb,
  Megaphone,
  MessageCircle,
  PenTool,
  ShieldCheck,
  Smartphone,
  Sparkles,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

const MARKETS = ["Bénin", "Sénégal", "Côte d'Ivoire", "Ghana", "Nigeria", "Kenya"]

const STEPS = [
  { icon: MessageCircle, key: "chat" },
  { icon: Lightbulb, key: "idea" },
  { icon: PenTool, key: "assets" },
  { icon: Download, key: "sell" },
] as const

const FEATURES = [
  { icon: ShieldCheck, key: "verified" },
  { icon: BookOpen, key: "identity" },
  { icon: Megaphone, key: "marketing" },
] as const

export default function HomePage() {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col">
      {/* Hero */}
      <section className="relative overflow-hidden">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 bg-gradient-to-b from-primary/10 via-transparent to-transparent dark:from-primary/20"
        />
        <div className="relative mx-auto flex max-w-3xl flex-col items-center py-16 text-center sm:py-24">
          <Badge variant="secondary" className="gap-1.5 rounded-full px-3 py-1">
            <Sparkles className="h-3.5 w-3.5" />
            {t("home.badge")}
          </Badge>
          <h1 className="mt-5 font-display text-4xl font-bold leading-tight tracking-tight text-primary sm:text-5xl">
            {t("home.heroTitle")}
          </h1>
          <p className="mt-4 max-w-2xl text-balance text-base text-muted-foreground sm:text-lg">
            {t("home.heroSubtitle")}
          </p>
          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <Button asChild size="touch">
              <Link to="/register">
                {t("home.ctaStart")} <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>
            <Button asChild size="touch" variant="outline">
              <a href="#how">{t("home.ctaHow")}</a>
            </Button>
          </div>
          <p className="mt-5 flex flex-wrap items-center justify-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
            <span className="inline-flex items-center gap-1.5">
              <ShieldCheck className="h-3.5 w-3.5" />
              {t("home.trustData")}
            </span>
            <span className="inline-flex items-center gap-1.5">
              <Smartphone className="h-3.5 w-3.5" />
              {t("home.trustMobileMoney")}
            </span>
          </p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-2">
            <span className="text-xs text-muted-foreground">{t("home.marketsLabel")}</span>
            {MARKETS.map((m) => (
              <span
                key={m}
                className="rounded-full border bg-card px-3 py-1 text-xs font-medium text-foreground/80"
              >
                {m}
              </span>
            ))}
          </div>
        </div>
      </section>

      {/* Comment ça marche */}
      <section id="how" className="border-t bg-muted/30 py-16">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <h2 className="text-center font-display text-2xl font-bold text-primary sm:text-3xl">
            {t("home.howTitle")}
          </h2>
          <p className="mx-auto mt-2 max-w-xl text-center text-sm text-muted-foreground">
            {t("home.howSubtitle")}
          </p>
          <ol className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {STEPS.map((step, i) => (
              <li key={step.key}>
                <Card className="h-full rounded-2xl">
                  <CardContent className="flex h-full flex-col gap-3 p-5">
                    <div className="flex items-center justify-between">
                      <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-primary-foreground">
                        <step.icon className="h-5 w-5" />
                      </span>
                      <span className="font-display text-2xl font-bold text-primary/20">
                        {String(i + 1).padStart(2, "0")}
                      </span>
                    </div>
                    <h3 className="font-semibold leading-snug">{t(`home.steps.${step.key}.title`)}</h3>
                    <p className="text-sm leading-relaxed text-muted-foreground">
                      {t(`home.steps.${step.key}.description`)}
                    </p>
                  </CardContent>
                </Card>
              </li>
            ))}
          </ol>
        </div>
      </section>

      {/* Ce qui vous est livré */}
      <section className="py-16">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <h2 className="text-center font-display text-2xl font-bold text-primary sm:text-3xl">
            {t("home.featuresTitle")}
          </h2>
          <div className="mt-10 grid gap-6 md:grid-cols-3">
            {FEATURES.map((f) => (
              <Card key={f.key} className="rounded-2xl transition-shadow hover:shadow-md">
                <CardContent className="flex h-full flex-col gap-3 p-6">
                  <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-secondary/10 text-secondary">
                    <f.icon className="h-5 w-5" />
                  </span>
                  <h3 className="font-semibold">{t(`home.features.${f.key}.title`)}</h3>
                  <p className="text-sm leading-relaxed text-muted-foreground">
                    {t(`home.features.${f.key}.description`)}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* CTA final */}
      <section className="px-4 pb-16 sm:px-6">
        <div className="mx-auto max-w-5xl rounded-3xl bg-primary px-6 py-14 text-center text-primary-foreground">
          <h2 className="font-display text-2xl font-bold sm:text-3xl">{t("home.finalCtaTitle")}</h2>
          <p className="mx-auto mt-3 max-w-xl text-sm text-primary-foreground/80 sm:text-base">
            {t("home.finalCtaSubtitle")}
          </p>
          <Button
            asChild
            size="touch"
            variant="secondary"
            className="mt-7 bg-background text-primary hover:bg-background/90"
          >
            <Link to="/register">
              {t("home.ctaStart")} <ArrowRight className="h-4 w-4" />
            </Link>
          </Button>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t py-8">
        <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 px-4 text-sm text-muted-foreground sm:flex-row sm:px-6">
          <p>© {new Date().getFullYear()} AfriLaunch · {t("home.rights")}</p>
          <div className="flex items-center gap-4">
            <Link to="/login" className="hover:text-primary hover:underline">
              {t("auth:login")}
            </Link>
            <Link to="/register" className="hover:text-primary hover:underline">
              {t("auth:register")}
            </Link>
          </div>
        </div>
      </footer>
    </div>
  )
}
