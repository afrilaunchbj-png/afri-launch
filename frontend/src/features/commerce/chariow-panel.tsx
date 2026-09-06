import { useState } from "react"
import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import { CheckCircle2, Copy, ExternalLink, Link2, RefreshCw, Store, Trash2, Unplug } from "lucide-react"
import { toast } from "sonner"

import { LoadingState } from "@/components/states/loading-state"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useProjects } from "@/features/projects/hooks"
import {
  useCommerceLinks,
  useCommerceProducts,
  useCommerceSales,
  useCommerceStatus,
  useConnectCommerce,
  useDisconnectCommerce,
  useLinkCommerceProduct,
  usePublishCommerceLink,
  useSyncCommerce,
  useUnlinkCommerceProduct,
} from "@/features/commerce/hooks"
import { isAppError } from "@/lib/errors"
import { cn } from "@/lib/utils"

function onError(t: (k: string) => string) {
  return (e: unknown) => toast.error(isAppError(e) ? e.message : t("common.genericError"))
}

/** Section « Vente — Chariow » de la page Intégrations. */
export function ChariowPanel() {
  const { t } = useTranslation()
  const statusQ = useCommerceStatus()
  const connect = useConnectCommerce()
  const disconnect = useDisconnectCommerce()
  const sync = useSyncCommerce()
  const [connectOpen, setConnectOpen] = useState(false)
  const [copied, setCopied] = useState(false)

  if (statusQ.isLoading) return <LoadingState label={t("common.loading")} />
  if (statusQ.isError || !statusQ.data) return null

  const connected = statusQ.data.connected === true
  const conn = statusQ.data.connection

  const copyWebhook = () => {
    if (!conn?.webhook_url) return
    void navigator.clipboard?.writeText(conn.webhook_url)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <Card>
      <CardContent className="flex flex-col gap-4 p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <h2 className="font-semibold">Chariow</h2>
              <Badge variant={connected ? "success" : "outline"}>
                {connected
                  ? t(`commerce:status.${conn?.status ?? "connected"}`)
                  : t("commerce:status.not_connected")}
              </Badge>
            </div>
            {connected ? (
              <p className="text-sm text-muted-foreground">
                <Store className="mr-1 inline h-4 w-4" />
                {conn?.store_name}
                {conn?.store_url ? (
                  <a
                    href={conn.store_url}
                    target="_blank"
                    rel="noreferrer"
                    className="ml-2 inline-flex items-center gap-1 text-primary hover:underline"
                  >
                    <ExternalLink className="h-3 w-3" />
                    {conn.store_url.replace(/^https?:\/\//, "")}
                  </a>
                ) : null}
              </p>
            ) : (
              <p className="text-sm text-muted-foreground">{t("commerce:notConnectedHint")}</p>
            )}
          </div>
          <div className="flex shrink-0 items-center gap-2">
            {connected ? (
              <>
                <Button size="sm" variant="outline" onClick={() => sync.mutate(undefined, { onError: onError(t) })} loading={sync.isPending}>
                  <RefreshCw className="h-4 w-4" />
                  {t("commerce:sync")}
                </Button>
                <Button size="sm" variant="ghost" onClick={() => disconnect.mutate(undefined, { onError: onError(t) })} loading={disconnect.isPending}>
                  <Unplug className="h-4 w-4" />
                  {t("commerce:disconnect")}
                </Button>
              </>
            ) : (
              <Button size="sm" onClick={() => setConnectOpen(true)}>
                <Link2 className="h-4 w-4" />
                {t("commerce:connect")}
              </Button>
            )}
          </div>
        </div>

        {connected && conn ? (
          <>
            <div className="flex flex-wrap items-center gap-2 rounded-lg border bg-muted/30 px-3 py-2 text-xs">
              <span className="text-muted-foreground">{t("commerce:webhookUrl")} :</span>
              <code className="flex-1 truncate text-foreground">{conn.webhook_url}</code>
              <Button size="sm" variant="ghost" onClick={copyWebhook}>
                {copied ? <CheckCircle2 className="h-4 w-4 text-success" /> : <Copy className="h-4 w-4" />}
                {copied ? t("commerce:copied") : t("commerce:copy")}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">{t("commerce:webhookHint")}</p>
            <LinkedProductsSection connected={connected} />
            <SalesSection />
          </>
        ) : null}
      </CardContent>

      <ConnectDialog open={connectOpen} onClose={() => setConnectOpen(false)} onSubmit={connect} />
    </Card>
  )
}

function ConnectDialog({
  open,
  onClose,
  onSubmit,
}: {
  open: boolean
  onClose: () => void
  onSubmit: ReturnType<typeof useConnectCommerce>
}) {
  const { t } = useTranslation()
  const [apiKey, setApiKey] = useState("")
  const [webhookSecret, setWebhookSecret] = useState("")

  const submit = () => {
    if (!apiKey.trim()) return
    onSubmit.mutate(
      { api_key: apiKey.trim(), webhook_secret: webhookSecret.trim() || undefined },
      {
        onSuccess: () => {
          toast.success(t("commerce:connectSuccess"))
          setApiKey("")
          setWebhookSecret("")
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
          <DialogTitle>{t("commerce:connectTitle")}</DialogTitle>
        </DialogHeader>
        <ol className="list-inside list-decimal space-y-1 text-sm text-muted-foreground">
          {[t("commerce:step1"), t("commerce:step2"), t("commerce:step3"), t("commerce:step4"), t("commerce:step5")].map(
            (s, i) => (
              <li key={i}>{s}</li>
            ),
          )}
        </ol>
        <div className="space-y-1.5">
          <Label htmlFor="com-api-key">{t("commerce:apiKey")}</Label>
          <Input id="com-api-key" value={apiKey} onChange={(e) => setApiKey(e.target.value)} placeholder="sk_live_…" />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="com-whsec">{t("commerce:webhookSecret")}</Label>
          <Input
            id="com-whsec"
            value={webhookSecret}
            onChange={(e) => setWebhookSecret(e.target.value)}
            placeholder="whsec_…"
          />
          <p className="text-xs text-muted-foreground">{t("commerce:webhookSecretHint")}</p>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            {t("integrations:cancel")}
          </Button>
          <Button onClick={submit} disabled={!apiKey.trim()} loading={onSubmit.isPending}>
            {t("commerce:connect")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function LinkedProductsSection({ connected }: { connected: boolean }) {
  const { t } = useTranslation()
  const { data: links, isLoading } = useCommerceLinks()
  const link = useLinkCommerceProduct()
  const publish = usePublishCommerceLink()
  const unlink = useUnlinkCommerceProduct()
  const productsQ = useCommerceProducts()
  const { data: projects } = useProjects()
  const [open, setOpen] = useState(false)
  const [projectId, setProjectId] = useState("")
  const [productId, setProductId] = useState("")

  const openDialog = () => {
    setOpen(true)
    setProjectId("")
    setProductId("")
    void productsQ.refetch()
  }

  const doLink = () => {
    if (!projectId || !productId) return
    link.mutate(
      { project_id: projectId, external_product_id: productId },
      {
        onSuccess: () => {
          toast.success(t("commerce:linkSuccess"))
          setOpen(false)
        },
        onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
      },
    )
  }

  return (
    <section className="space-y-2">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">{t("commerce:linkedProducts")}</h3>
        {connected ? (
          <Button size="sm" variant="outline" onClick={openDialog} loading={productsQ.isLoading}>
            <Link2 className="h-4 w-4" />
            {t("commerce:linkProduct")}
          </Button>
        ) : null}
      </div>

      {isLoading ? (
        <LoadingState label={t("common.loading")} />
      ) : !links || links.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("commerce:noLinkedProducts")}</p>
      ) : (
        <div className="divide-y rounded-lg border bg-card">
          {links.map((l) => (
            <div key={l.id} className="flex flex-wrap items-center justify-between gap-3 p-3">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{l.external_product_name}</p>
                <p className="text-xs text-muted-foreground">
                  {t("commerce:linkedTo")} {projects?.find((p) => p.id === l.project_id)?.title ?? l.project_id}
                  {l.is_public && l.public_url ? (
                    <Link to={l.public_url} className="ml-2 text-primary hover:underline">
                      {t("commerce:publicLink")}
                    </Link>
                  ) : null}
                </p>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <Badge variant={l.is_public ? "success" : "outline"}>{t("commerce:published")}</Badge>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={publish.isPending}
                  onClick={() =>
                    publish.mutate({ id: l.id, public: !l.is_public }, { onError: onError(t) })
                  }
                >
                  {l.is_public ? t("commerce:unpublish") : t("commerce:publish")}
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  aria-label={t("commerce:unlink")}
                  onClick={() => unlink.mutate(l.id, { onError: onError(t) })}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <Dialog open={open} onOpenChange={(v) => !v && setOpen(false)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("commerce:linkProduct")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label htmlFor="com-project">{t("commerce:chooseProject")}</Label>
              <Select value={projectId} onValueChange={setProjectId}>
                <SelectTrigger id="com-project">
                  <SelectValue placeholder={t("commerce:chooseProject")} />
                </SelectTrigger>
                <SelectContent>
                  {(projects ?? []).map((p) => (
                    <SelectItem key={p.id} value={p.id}>
                      {p.title}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="com-product">{t("commerce:chooseProduct")}</Label>
              <Select value={productId} onValueChange={setProductId}>
                <SelectTrigger id="com-product">
                  <SelectValue placeholder={t("commerce:chooseProduct")} />
                </SelectTrigger>
                <SelectContent>
                  {(productsQ.data ?? []).map((p) => (
                    <SelectItem key={p.id} value={p.id}>
                      {p.name} — {p.price_minor} {p.currency}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {productsQ.isLoading ? (
                <p className="text-xs text-muted-foreground">{t("common.loading")}</p>
              ) : null}
            </div>
            <p className="text-xs text-muted-foreground">{t("commerce:linkHint")}</p>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setOpen(false)}>
              {t("integrations:cancel")}
            </Button>
            <Button onClick={doLink} disabled={!projectId || !productId} loading={link.isPending}>
              {t("commerce:linkProduct")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  )
}

function SalesSection() {
  const { t } = useTranslation()
  const { data: sales, isLoading } = useCommerceSales()

  return (
    <section className="space-y-2">
      <h3 className="text-sm font-semibold">{t("commerce:recentSales")}</h3>
      {isLoading ? (
        <LoadingState label={t("common.loading")} />
      ) : !sales || sales.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("commerce:noSales")}</p>
      ) : (
        <div className="divide-y rounded-lg border bg-card">
          {sales.slice(0, 10).map((s) => (
            <div key={s.id} className="flex items-center justify-between gap-3 p-3 text-sm">
              <div className="min-w-0">
                <p className="truncate font-medium">{s.buyer_email || s.external_sale_id}</p>
                <p className="text-xs text-muted-foreground">
                  {s.buyer_name || "—"} · {new Date(s.created_at).toLocaleString()}
                </p>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <Badge variant={s.status === "completed" ? "success" : s.status === "failed" || s.status === "refunded" ? "destructive" : "warning"}>
                  {s.status}
                </Badge>
                <span className={cn("font-semibold", s.status === "completed" ? "text-success" : "")}>
                  {(s.amount_minor / 100).toFixed(2)} {s.currency}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
