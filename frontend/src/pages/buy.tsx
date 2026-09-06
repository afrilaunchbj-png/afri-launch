import { useState } from "react"
import { useParams } from "react-router"
import { useTranslation } from "react-i18next"
import { useQuery } from "@tanstack/react-query"
import { CheckCircle2, ShoppingCart, Store } from "lucide-react"

import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { createPublicCheckout, fetchPublicProduct } from "@/features/commerce/api"
import { isAppError } from "@/lib/errors"

function money(minor: number, currency: string) {
  const amount = currency === "XOF" || currency === "XAF" ? minor : minor / 100
  return `${amount.toLocaleString()} ${currency}`
}

export default function BuyPage() {
  const { t } = useTranslation()
  const { token = "" } = useParams()
  const { data: product, isLoading, isError, refetch } = useQuery({
    queryKey: ["commerce", "public", token],
    queryFn: () => fetchPublicProduct(token),
    enabled: !!token,
    retry: false,
  })

  const [email, setEmail] = useState("")
  const [firstName, setFirstName] = useState("")
  const [lastName, setLastName] = useState("")
  const [phone, setPhone] = useState("")
  const [country, setCountry] = useState("BJ")
  const [pending, setPending] = useState(false)
  const [done, setDone] = useState<string | null>(null)

  const canSubmit = email.includes("@") && firstName.trim() && lastName.trim() && phone.trim().length >= 6 && country.trim()

  const handleBuy = async () => {
    if (!canSubmit) return
    setPending(true)
    try {
      const res = await createPublicCheckout(token, {
        email: email.trim(),
        first_name: firstName.trim(),
        last_name: lastName.trim(),
        phone: { number: phone.trim().replace(/\s+/g, ""), country_code: country.trim().toUpperCase() },
      })
      if (res.step === "completed") {
        setDone("completed")
      } else if (res.step === "already_purchased") {
        setDone("already")
      } else if (res.checkout_url) {
        window.location.href = res.checkout_url
        return
      } else {
        setDone("completed")
      }
    } catch (e) {
      alert(isAppError(e) ? e.message : t("common.genericError"))
    } finally {
      setPending(false)
    }
  }

  if (isLoading) return <LoadingState label={t("common.loading")} />
  if (isError || !product) return <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />

  return (
    <div className="mx-auto max-w-md space-y-6 py-8">
      <Card>
        <CardHeader className="text-center">
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
            <ShoppingCart className="h-6 w-6" />
          </div>
          <CardTitle className="font-display text-xl text-primary">{product.name}</CardTitle>
          <p className="flex items-center justify-center gap-1 text-sm text-muted-foreground">
            <Store className="h-3.5 w-3.5" />
            {product.store_name}
          </p>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="text-center">
            <span className="font-display text-3xl font-bold text-primary">{money(product.price_minor, product.currency)}</span>
          </div>

          {done ? (
            <div className="flex flex-col items-center gap-2 rounded-lg border border-success/30 bg-success/5 p-4 text-center">
              <CheckCircle2 className="h-8 w-8 text-success" />
              <p className="text-sm font-medium">
                {done === "already" ? t("commerce:alreadyPurchased") : t("commerce:purchaseCompleted")}
              </p>
            </div>
          ) : (
            <form
              className="space-y-3"
              onSubmit={(e) => {
                e.preventDefault()
                void handleBuy()
              }}
            >
              <div className="space-y-1.5">
                <Label htmlFor="buy-email">{t("commerce:emailLabel")}</Label>
                <Input id="buy-email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="email" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="buy-first">{t("commerce:firstNameLabel")}</Label>
                  <Input id="buy-first" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="buy-last">{t("commerce:lastNameLabel")}</Label>
                  <Input id="buy-last" value={lastName} onChange={(e) => setLastName(e.target.value)} />
                </div>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="buy-phone">{t("commerce:phoneLabel")}</Label>
                <div className="grid grid-cols-[100px_1fr] gap-2">
                  <Input
                    value={country}
                    onChange={(e) => setCountry(e.target.value.toUpperCase())}
                    placeholder="BJ"
                    aria-label={t("commerce:phoneCountryLabel")}
                    maxLength={10}
                  />
                  <Input id="buy-phone" type="tel" value={phone} onChange={(e) => setPhone(e.target.value)} autoComplete="tel" />
                </div>
              </div>
              <Button type="submit" size="touch" className="w-full" disabled={!canSubmit || pending} loading={pending}>
                {pending ? t("commerce:buying") : t("commerce:buyNow")}
              </Button>
              <p className="text-center text-xs text-muted-foreground">{t("commerce:buySubtitle")}</p>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
