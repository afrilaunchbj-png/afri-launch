import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import { useEffect } from "react"
import { toast } from "sonner"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useCreateCheckout, usePlans, useSyncPayment } from "@/features/credits/hooks"
import { isAppError } from "@/lib/errors"

/** formatPrice affiche un montant en unités mineures (XOF = zéro décimale). */
function formatPrice(minor: number, currency: string) {
  const amount = currency === "XOF" || currency === "XAF" ? minor : minor / 100
  return `${amount.toLocaleString("fr-FR")} ${currency}`
}

/**
 * PlansPanel : packs de crédits achetables. Redirige vers la page de
 * paiement hébergée par le provider (PawaPay / FedaPay / PayDunya).
 * Le retour sur ?payment={id} déclenche la reconfirmation du statut.
 */
export function PlansPanel() {
  const { t } = useTranslation()
  const { data, isLoading } = usePlans()
  const checkout = useCreateCheckout()
  const sync = useSyncPayment()
  const [searchParams, setSearchParams] = useSearchParams()

  // Retour du provider : reconfirme le statut puis nettoie l'URL.
  useEffect(() => {
    const paymentId = searchParams.get("payment")
    if (!paymentId) return
    sync.mutate(paymentId, {
      onSuccess: (payment) => {
        if (payment.status === "succeeded") {
          toast.success(t("credits:paymentSucceeded"))
        } else if (payment.status === "failed") {
          toast.error(t("credits:paymentFailed"))
        } else {
          toast.info(t("credits:paymentPending"))
        }
        searchParams.delete("payment")
        setSearchParams(searchParams, { replace: true })
      },
      onError: () => {
        searchParams.delete("payment")
        setSearchParams(searchParams, { replace: true })
      },
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps -- déclenché une fois par ?payment=
  }, [searchParams])

  if (isLoading) {
    return (
      <div className="grid gap-4 sm:grid-cols-3">
        {[0, 1, 2].map((i) => (
          <Skeleton key={i} className="h-40 rounded-xl" />
        ))}
      </div>
    )
  }
  if (!data || !data.enabled) return null

  const handleBuy = (planId: string) => {
    checkout.mutate(planId, {
      onSuccess: ({ redirect_url }) => {
        window.location.href = redirect_url
      },
      onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
    })
  }

  return (
    <section id="plans" className="scroll-mt-24">
      <div className="mb-3 flex items-center gap-2">
        <h2 className="font-display text-lg font-semibold text-primary">{t("credits:plansTitle")}</h2>
        <Badge variant="outline" className="text-[10px] uppercase">
          {t("credits:paymentProvider", { provider: data.provider })}
        </Badge>
      </div>
      <div className="grid gap-4 sm:grid-cols-3">
        {data.plans.map((plan) => (
          <Card key={plan.id} className="flex flex-col">
            <CardContent className="flex flex-1 flex-col gap-3 p-5">
              <h3 className="font-semibold">{plan.name}</h3>
              <p className="font-display text-2xl font-bold text-primary">
                {t("credits:planCredits", { count: plan.credits })}
              </p>
              <p className="text-sm text-muted-foreground">{formatPrice(plan.price_minor, plan.currency)}</p>
              <Button className="mt-auto" size="sm" loading={checkout.isPending} onClick={() => handleBuy(plan.id)}>
                {t("credits:buy")}
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
      <p className="mt-2 text-xs text-muted-foreground">{t("credits:paymentHint")}</p>
    </section>
  )
}
