import { useEffect, useMemo, useState } from "react"
import { useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import { CheckCircle2, Link2, Pause, Play, RefreshCw, Trash2 } from "lucide-react"
import { toast } from "sonner"

import { ErrorState } from "@/components/states/error-state"
import { LoadingState } from "@/components/states/loading-state"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  PROVIDERS,
  type AdConnection,
  type AdProvider,
} from "@/features/integrations/api"
import {
  useAdAccounts,
  useCampaigns,
  useConnectProvider,
  useCreateCampaign,
  useDisconnectProvider,
  useIntegrations,
  usePauseResumeCampaign,
  useSelectAccount,
  useSyncCampaigns,
} from "@/features/integrations/hooks"
import { isAppError } from "@/lib/errors"

const OBJECTIVES = ["OUTCOME_TRAFFIC", "OUTCOME_SALES", "OUTCOME_LEADS", "OUTCOME_ENGAGEMENT", "OUTCOME_AWARENESS"]

function isActive(c?: AdConnection) {
  return c?.status === "active"
}

export default function IntegrationsPage() {
  const { t } = useTranslation()
  const { data, isLoading, isError, refetch } = useIntegrations()
  const [searchParams, setSearchParams] = useSearchParams()

  const byProvider = useMemo(() => {
    const map = new Map<string, AdConnection>()
    for (const c of data?.connections ?? []) map.set(c.provider, c)
    return map
  }, [data?.connections])

  // Retour du callback OAuth : notifier puis nettoyer l'URL.
  useEffect(() => {
    const status = searchParams.get("status")
    const provider = searchParams.get("connect")
    if (!status || !provider) return
    if (status === "success") {
      toast.success(t("integrations:connectSuccess", { provider: t(`integrations:provider.${provider}`) }))
    } else {
      toast.error(t("integrations:connectError", { provider: t(`integrations:provider.${provider}`) }))
    }
    searchParams.delete("status")
    searchParams.delete("connect")
    setSearchParams(searchParams, { replace: true })
  }, [searchParams, setSearchParams, t])

  if (isLoading) return <LoadingState label={t("common.loading")} />
  if (isError || !data) return <ErrorState title={t("common.genericError")} onRetry={() => refetch()} />

  return (
    <div className="space-y-6">
      <header>
        <h1 className="font-display text-2xl font-bold text-primary md:text-3xl">{t("integrations:title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("integrations:subtitle")}</p>
      </header>

      <section className="space-y-4">
        {PROVIDERS.map((p) => (
          <ProviderCard key={p.id} provider={p.id} name={p.name} connection={byProvider.get(p.id)} />
        ))}
      </section>

      <CampaignsSection />
    </div>
  )
}

function ProviderCard({
  provider,
  name,
  connection,
}: {
  provider: AdProvider
  name: string
  connection?: AdConnection
}) {
  const { t } = useTranslation()
  const connect = useConnectProvider()
  const disconnect = useDisconnectProvider()
  const [picking, setPicking] = useState(false)

  const handleConnect = () => {
    connect.mutate(provider, {
      onSuccess: (url) => {
        window.location.href = url
      },
      onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
    })
  }

  const handleDisconnect = () => {
    disconnect.mutate(provider, {
      onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
    })
  }

  const statusVariant = isActive(connection) ? "success" : connection?.status === "error" ? "destructive" : "outline"

  return (
    <Card>
      <CardContent className="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <h2 className="font-semibold">{name}</h2>
            {connection ? (
              <Badge variant={statusVariant}>{t(`integrations:status.${connection.status}`)}</Badge>
            ) : (
              <Badge variant="outline">{t("integrations:status.none")}</Badge>
            )}
          </div>
          {isActive(connection) ? (
            <p className="text-sm text-muted-foreground">
              {t("integrations:accountLabel")} : {connection?.external_account_name || connection?.external_account_id}
            </p>
          ) : connection?.status === "pending" ? (
            <p className="text-sm text-muted-foreground">{t("integrations:pendingHint")}</p>
          ) : null}
          {connection?.last_error ? (
            <p className="text-xs text-destructive">{connection.last_error}</p>
          ) : null}
        </div>
        <div className="flex shrink-0 items-center gap-2">
          {isActive(connection) ? (
            <>
              <Button size="sm" variant="outline" onClick={() => setPicking(true)}>
                {t("integrations:manageAccount")}
              </Button>
              <Button size="sm" variant="ghost" onClick={handleDisconnect} loading={disconnect.isPending}>
                <Trash2 className="h-4 w-4" />
                {t("integrations:disconnect")}
              </Button>
            </>
          ) : connection?.status === "pending" ? (
            <Button size="sm" variant="outline" onClick={() => setPicking(true)}>
              <Link2 className="h-4 w-4" />
              {t("integrations:chooseAccount")}
            </Button>
          ) : (
            <Button size="sm" onClick={handleConnect} loading={connect.isPending}>
              {t("integrations:connect", { provider: name })}
            </Button>
          )}
        </div>
        {picking ? <AccountPicker provider={provider} onClose={() => setPicking(false)} /> : null}
      </CardContent>
    </Card>
  )
}

function AccountPicker({ provider, onClose }: { provider: AdProvider; onClose: () => void }) {
  const { t } = useTranslation()
  const { data: accounts, isLoading } = useAdAccounts(provider)
  const select = useSelectAccount()
  const [accountId, setAccountId] = useState("")

  const handleConfirm = () => {
    if (!accountId) return
    select.mutate(
      { provider, accountId },
      {
        onSuccess: () => {
          toast.success(t("integrations:accountSelected"))
          onClose()
        },
        onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
      },
    )
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("integrations:chooseAccountTitle")}</DialogTitle>
        </DialogHeader>
        {isLoading ? (
          <LoadingState label={t("common.loading")} />
        ) : !accounts || accounts.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("integrations:noAccounts")}</p>
        ) : (
          <div className="space-y-2">
            {accounts.map((a) => (
              <label
                key={a.id}
                className={
                  "flex cursor-pointer items-center gap-3 rounded-lg border p-3 text-sm " +
                  (accountId === a.id ? "border-primary bg-muted/50" : "")
                }
              >
                <input
                  type="radio"
                  name="ad-account"
                  className="h-4 w-4"
                  checked={accountId === a.id}
                  onChange={() => setAccountId(a.id)}
                />
                <span className="font-medium">{a.name}</span>
                <span className="ml-auto text-xs text-muted-foreground">{a.id}</span>
                {a.currency ? <Badge variant="secondary">{a.currency}</Badge> : null}
              </label>
            ))}
          </div>
        )}
        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            {t("integrations:cancel")}
          </Button>
          <Button onClick={handleConfirm} disabled={!accountId || select.isPending} loading={select.isPending}>
            <CheckCircle2 className="h-4 w-4" />
            {t("integrations:confirmAccount")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function CampaignsSection() {
  const { t } = useTranslation()
  const { data: campaigns, isLoading } = useCampaigns()
  const sync = useSyncCampaigns()
  const setStatus = usePauseResumeCampaign()
  const create = useCreateCampaign()
  const [creating, setCreating] = useState(false)

  const onError = (e: unknown) => toast.error(isAppError(e) ? e.message : t("common.genericError"))

  const activeProvider: AdProvider | null = PROVIDERS[0].id // meta au MVP

  return (
    <section>
      <div className="mb-3 flex items-center justify-between">
        <h2 className="font-display text-lg font-semibold text-primary">{t("integrations:campaignsTitle")}</h2>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={() => setCreating(true)}>
            {t("integrations:createCampaign")}
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => activeProvider && sync.mutate(activeProvider, { onError })}
            loading={sync.isPending}
          >
            <RefreshCw className="h-4 w-4" />
            {t("integrations:syncCampaigns")}
          </Button>
        </div>
      </div>

      {isLoading ? (
        <LoadingState label={t("common.loading")} />
      ) : !campaigns || campaigns.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("integrations:noCampaigns")}</p>
      ) : (
        <div className="divide-y rounded-lg border bg-card">
          {campaigns.map((c) => (
            <div key={c.id} className="flex items-center justify-between gap-3 p-4">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{c.name}</p>
                <p className="text-xs text-muted-foreground">
                  {c.objective} · {c.budget_minor.toLocaleString()} {c.currency}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant={c.status === "active" ? "success" : c.status === "paused" ? "secondary" : "outline"}>
                  {t(`integrations:campaignStatus.${c.status}`)}
                </Badge>
                {c.status === "active" ? (
                  <Button
                    size="sm"
                    variant="ghost"
                    aria-label={t("integrations:pause")}
                    onClick={() => setStatus.mutate({ id: c.id, action: "pause" }, { onError })}
                  >
                    <Pause className="h-4 w-4" />
                  </Button>
                ) : c.status === "paused" ? (
                  <Button
                    size="sm"
                    variant="ghost"
                    aria-label={t("integrations:resume")}
                    onClick={() => setStatus.mutate({ id: c.id, action: "resume" }, { onError })}
                  >
                    <Play className="h-4 w-4" />
                  </Button>
                ) : null}
              </div>
            </div>
          ))}
        </div>
      )}

      <CreateCampaignDialog open={creating} onClose={() => setCreating(false)} onCreate={create} />
    </section>
  )
}

function CreateCampaignDialog({
  open,
  onClose,
  onCreate,
}: {
  open: boolean
  onClose: () => void
  onCreate: ReturnType<typeof useCreateCampaign>
}) {
  const { t } = useTranslation()
  const [name, setName] = useState("")
  const [objective, setObjective] = useState(OBJECTIVES[0])
  const [budget, setBudget] = useState("")

  const handleSubmit = () => {
    if (!name.trim() || !budget) return
    onCreate.mutate(
      {
        provider: "meta",
        name: name.trim(),
        objective,
        budget_minor: Math.round(Number(budget) * 100),
      },
      {
        onSuccess: () => {
          toast.success(t("integrations:campaignCreated"))
          setName("")
          setBudget("")
          onClose()
        },
        onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
      },
    )
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("integrations:createCampaignTitle")}</DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <div className="space-y-1.5">
            <Label htmlFor="camp-name">{t("integrations:campaignName")}</Label>
            <Input id="camp-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="camp-objective">{t("integrations:campaignObjective")}</Label>
            <Select value={objective} onValueChange={setObjective}>
              <SelectTrigger id="camp-objective">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {OBJECTIVES.map((o) => (
                  <SelectItem key={o} value={o}>
                    {o}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="camp-budget">{t("integrations:campaignBudget")}</Label>
            <Input
              id="camp-budget"
              type="number"
              min={0}
              step="0.01"
              value={budget}
              onChange={(e) => setBudget(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">{t("integrations:campaignBudgetHint")}</p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            {t("integrations:cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={!name.trim() || !budget} loading={onCreate.isPending}>
            {t("integrations:createCampaign")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
