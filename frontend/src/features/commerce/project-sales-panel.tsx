import { useEffect, useState } from "react"
import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import { ExternalLink, Link2, RefreshCw, Settings2, ShoppingBag, Store, Trash2 } from "lucide-react"
import { toast } from "sonner"

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
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { LoadingState } from "@/components/states/loading-state"
import {
  useCommerceLinks,
  useCommerceProducts,
  useCommerceSales,
  useCommerceStatus,
  useLinkCommerceProduct,
  usePublishCommerceLink,
  useSyncCommerce,
  useUnlinkCommerceProduct,
} from "@/features/commerce/hooks"
import { isAppError } from "@/lib/errors"

function money(minor: number, currency: string) {
  const amount = currency === "XOF" || currency === "XAF" ? minor : minor / 100
  return `${amount.toLocaleString()} ${currency}`
}

function onError(t: (k: string) => string) {
  return (e: unknown) => toast.error(isAppError(e) ? e.message : t("common.genericError"))
}

/**
 * CommerceSalesPanel : page de vente d'un projet branchée sur Chariow.
 * « Configurer » = lier un produit existant de la boutique ; « Consulter »
 * = ouvrir la page d'achat publique + voir boutique/ventes.
 */
export function CommerceSalesPanel({ projectId }: { projectId: string }) {
  const { t } = useTranslation()
  const { data: statusQ } = useCommerceStatus()
  const { data: links, isLoading: linksLoading } = useCommerceLinks()
  const link = links?.find((l) => l.project_id === projectId)

  const status = statusQ?.connection
  const connected = statusQ?.connected === true && status?.status === "connected"

  const [pickerOpen, setPickerOpen] = useState(false)

  if (linksLoading) return <LoadingState label={t("common.loading")} />

  return (
    <Card>
      <CardContent className="flex h-full flex-col gap-4 p-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ShoppingBag className="h-6 w-6 text-primary" />
            <h3 className="font-semibold">{t("projects:salesPage")}</h3>
            <Badge variant={connected && link ? "success" : "outline"}>
              {connected && link ? t("commerce:linkedBadge") : t("commerce:notLinkedBadge")}
            </Badge>
          </div>
        </div>

        {!connected ? (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">{t("commerce:needsConnection")}</p>
            <Button asChild size="sm" variant="outline">
              <Link to="/integrations">{t("commerce:goToIntegrations")}</Link>
            </Button>
          </div>
        ) : !link ? (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">{t("commerce:notLinkedHint")}</p>
            <Button size="sm" onClick={() => setPickerOpen(true)}>
              <Settings2 className="h-4 w-4" />
              {t("commerce:configure")}
            </Button>
          </div>
        ) : (
          <LinkedView projectId={projectId} link={link} storeName={status?.store_name ?? ""} storeUrl={status?.store_url ?? ""} />
        )}

        {connected ? (
          <ProductPickerDialog
            open={pickerOpen}
            projectId={projectId}
            onClose={() => setPickerOpen(false)}
          />
        ) : null}
      </CardContent>
    </Card>
  )
}

function LinkedView({
  projectId,
  link,
  storeName,
  storeUrl,
}: {
  projectId: string
  link: { id: string; external_product_name: string; price_minor: number; currency: string; is_public: boolean; public_url: string }
  storeName: string
  storeUrl: string
}) {
  const { t } = useTranslation()
  const publish = usePublishCommerceLink()
  const unlink = useUnlinkCommerceProduct()
  const sync = useSyncCommerce()
  const { data: sales, isLoading: salesLoading } = useCommerceSales()
  const [openPicker, setOpenPicker] = useState(false)

  const linkSales = (sales ?? []).filter((s) => s.product_link_id === link.id).slice(0, 8)
  const completed = linkSales.filter((s) => s.status === "completed" || s.status === "settled").length

  const publicUrl = link.is_public && link.public_url ? link.public_url : null

  return (
    <>
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="space-y-1">
          <p className="font-medium">{link.external_product_name}</p>
          <p className="text-sm text-muted-foreground">
            {money(link.price_minor, link.currency)}
            {storeName ? (
              <span className="ml-3 inline-flex items-center gap-1">
                <Store className="h-3.5 w-3.5" />
                {storeName}
              </span>
            ) : null}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {publicUrl ? (
            <Button size="sm" asChild>
              <a href={publicUrl} target="_blank" rel="noreferrer">
                <ExternalLink className="h-4 w-4" />
                {t("commerce:viewSalesPage")}
              </a>
            </Button>
          ) : (
            <Button
              size="sm"
              variant="outline"
              loading={publish.isPending}
              onClick={() =>
                publish.mutate(
                  { id: link.id, public: true },
                  { onSuccess: () => toast.success(t("commerce:published")), onError: onError(t) },
                )
              }
            >
              {t("commerce:publish")}
            </Button>
          )}
          <Button size="sm" variant="outline" onClick={() => setOpenPicker(true)}>
            <Settings2 className="h-4 w-4" />
            {t("commerce:changeProduct")}
          </Button>
          {storeUrl ? (
            <Button size="sm" variant="ghost" asChild>
              <a href={storeUrl} target="_blank" rel="noreferrer">
                <ExternalLink className="h-4 w-4" />
                {t("commerce:openStore")}
              </a>
            </Button>
          ) : null}
          <Button
            size="sm"
            variant="ghost"
            className="text-destructive hover:text-destructive"
            onClick={() => {
              if (window.confirm(t("commerce:unlinkConfirm"))) unlink.mutate(link.id, { onError: onError(t) })
            }}
          >
            <Trash2 className="h-4 w-4" />
            {t("commerce:unlink")}
          </Button>
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-2">
        <div className="rounded-lg border bg-muted/30 p-3">
          <p className="text-xs text-muted-foreground">{t("commerce:salesCount")}</p>
          <p className="font-display text-2xl font-bold text-primary">{completed}</p>
        </div>
        <div className="rounded-lg border bg-muted/30 p-3">
          <p className="text-xs text-muted-foreground">{t("commerce:salesRevenue")}</p>
          <p className="font-display text-2xl font-bold text-primary">
            {money(
              linkSales.reduce((sum, s) => sum + (s.status === "completed" || s.status === "settled" ? s.amount_minor : 0), 0),
              link.currency,
            )}
          </p>
        </div>
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <h4 className="text-sm font-semibold">{t("commerce:recentSales")}</h4>
          <Button size="sm" variant="ghost" onClick={() => sync.mutate(undefined, { onError: onError(t) })} loading={sync.isPending}>
            <RefreshCw className="h-4 w-4" />
            {t("commerce:sync")}
          </Button>
        </div>
        {salesLoading ? (
          <LoadingState label={t("common.loading")} />
        ) : linkSales.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("commerce:noSales")}</p>
        ) : (
          <div className="divide-y rounded-lg border bg-card">
            {linkSales.map((s) => (
              <div key={s.id} className="flex items-center justify-between gap-3 p-3 text-sm">
                <div className="min-w-0">
                  <p className="truncate font-medium">{s.buyer_email || s.external_sale_id}</p>
                  <p className="text-xs text-muted-foreground">{new Date(s.created_at).toLocaleString()}</p>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <Badge variant={s.status === "completed" ? "success" : s.status === "failed" || s.status === "refunded" ? "destructive" : "warning"}>
                    {s.status}
                  </Badge>
                  <span className="font-semibold">{money(s.amount_minor, s.currency)}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <ProductPickerDialog
        open={openPicker}
        projectId={projectId}
        onClose={() => setOpenPicker(false)}
      />
    </>
  )
}

function ProductPickerDialog({
  open,
  projectId,
  onClose,
}: {
  open: boolean
  projectId: string
  onClose: () => void
}) {
  const { t } = useTranslation()
  const productsQ = useCommerceProducts()
  const link = useLinkCommerceProduct()
  const [productId, setProductId] = useState("")

  useEffect(() => {
    if (open) {
      setProductId("")
      void productsQ.refetch()
    }
  }, [open]) // eslint-disable-line react-hooks/exhaustive-deps

  const submit = () => {
    if (!productId) return
    link.mutate(
      { project_id: projectId, external_product_id: productId },
      {
        onSuccess: () => {
          toast.success(t("commerce:linkSuccess"))
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
          <DialogTitle>{t("commerce:chooseProduct")}</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">{t("commerce:linkHint")}</p>
        <div className="space-y-1.5">
          <Label htmlFor="sales-product">{t("commerce:chooseProduct")}</Label>
          <Select value={productId} onValueChange={setProductId}>
            <SelectTrigger id="sales-product">
              <SelectValue placeholder={t("commerce:chooseProduct")} />
            </SelectTrigger>
            <SelectContent>
              {(productsQ.data ?? []).map((p) => (
                <SelectItem key={p.id} value={p.id}>
                  {p.name} — {money(p.price_minor, p.currency)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {productsQ.isLoading ? <p className="text-xs text-muted-foreground">{t("common.loading")}</p> : null}
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            {t("integrations:cancel")}
          </Button>
          <Button onClick={submit} disabled={!productId} loading={link.isPending}>
            <Link2 className="h-4 w-4" />
            {t("commerce:configure")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
