import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { Globe2 } from "lucide-react"
import { toast } from "sonner"

import { AdminNav } from "@/features/admin/admin-nav"
import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { api, type ApiSingle } from "@/lib/api/client"
import { isAppError } from "@/lib/errors"

interface PaymentCountry {
  code: string
  name: string
  currency: string
  enabled: boolean
}

function fetchCountries() {
  return api.get<ApiSingle<PaymentCountry[]>>("/api/v1/admin/payment-countries").then((r) => r.data)
}

function setCountryEnabled(code: string, enabled: boolean) {
  return api.put(`/api/v1/admin/payment-countries/${code}`, { enabled })
}

export default function AdminPaymentCountriesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["admin", "payment-countries"],
    queryFn: fetchCountries,
  })
  const mutate = useMutation({
    mutationFn: ({ code, enabled }: { code: string; enabled: boolean }) => setCountryEnabled(code, enabled),
    onSettled: () => qc.invalidateQueries({ queryKey: ["admin", "payment-countries"] }),
    onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
  })

  return (
    <div className="space-y-6">
      <header>
        <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">{t("admin:paymentCountries")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("admin:paymentCountriesDesc")}</p>
      </header>
      <AdminNav />

      {isLoading ? (
        <LoadingState label={t("common.loading")} />
      ) : isError || !data ? (
        <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />
      ) : (
        <Card>
          <CardContent className="p-2">
            <div className="divide-y">
              {data.map((c) => (
                <div key={c.code} className="flex items-center justify-between gap-3 px-3 py-3">
                  <div className="flex items-center gap-3">
                    <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                      <Globe2 className="h-4 w-4" />
                    </span>
                    <div>
                      <p className="text-sm font-medium">{c.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {c.code} · {c.currency}
                      </p>
                    </div>
                  </div>
                  <div className="flex shrink-0 items-center gap-2">
                    <Badge variant={c.enabled ? "success" : "outline"}>
                      {c.enabled ? t("admin:enabled") : t("admin:disabled")}
                    </Badge>
                    <Button
                      size="sm"
                      variant={c.enabled ? "outline" : "default"}
                      loading={mutate.isPending}
                      onClick={() => mutate.mutate({ code: c.code, enabled: !c.enabled })}
                    >
                      {c.enabled ? t("common.disable") : t("common.enable")}
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
