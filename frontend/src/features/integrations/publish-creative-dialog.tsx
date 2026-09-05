import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Megaphone } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
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
import { Textarea } from "@/components/ui/textarea"
import type { Asset } from "@/features/projects/api"
import {
  useCampaigns,
  usePublishCreative,
} from "@/features/integrations/hooks"
import { isAppError } from "@/lib/errors"

const CTAS = ["LEARN_MORE", "SHOP_NOW", "SIGN_UP", "DOWNLOAD"]

/**
 * PublishCreativeDialog publie une vidéo du projet comme créative sur une
 * campagne synchronisée (créée en pause côté plateforme, cf. ADR-017).
 */
export function PublishCreativeDialog({
  open,
  onClose,
  asset,
}: {
  open: boolean
  onClose: () => void
  asset: Asset
}) {
  const { t } = useTranslation()
  const { data: campaigns, isLoading } = useCampaigns()
  const publish = usePublishCreative()

  const usable = (campaigns ?? []).filter((c) => c.status !== "deleted")
  const [campaignId, setCampaignId] = useState("")
  const [headline, setHeadline] = useState("")
  const [primaryText, setPrimaryText] = useState("")
  const [cta, setCta] = useState(CTAS[0])

  const handleSubmit = () => {
    if (!campaignId || !headline.trim()) return
    publish.mutate(
      {
        campaignId,
        input: {
          asset_id: asset.id,
          headline: headline.trim(),
          primary_text: primaryText.trim(),
          cta,
        },
      },
      {
        onSuccess: () => {
          toast.success(t("integrations:creativePublished"))
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
          <DialogTitle>{t("integrations:publishCreativeTitle")}</DialogTitle>
          <DialogDescription>{t("integrations:publishCreativeHint")}</DialogDescription>
        </DialogHeader>

        {isLoading ? (
          <p className="text-sm text-muted-foreground">{t("common.loading")}</p>
        ) : usable.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("integrations:noCampaignForCreative")}</p>
        ) : (
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label htmlFor="creative-campaign">{t("integrations:creativeCampaign")}</Label>
              <Select value={campaignId} onValueChange={setCampaignId}>
                <SelectTrigger id="creative-campaign">
                  <SelectValue placeholder={t("integrations:creativeCampaignPlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  {usable.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="creative-headline">{t("integrations:creativeHeadline")}</Label>
              <Input
                id="creative-headline"
                value={headline}
                maxLength={40}
                onChange={(e) => setHeadline(e.target.value)}
                placeholder={t("integrations:creativeHeadlinePlaceholder")}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="creative-text">{t("integrations:creativePrimaryText")}</Label>
              <Textarea
                id="creative-text"
                rows={2}
                value={primaryText}
                maxLength={200}
                onChange={(e) => setPrimaryText(e.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="creative-cta">{t("integrations:creativeCTA")}</Label>
              <Select value={cta} onValueChange={setCta}>
                <SelectTrigger id="creative-cta">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {CTAS.map((c) => (
                    <SelectItem key={c} value={c}>
                      {t(`integrations:cta.${c}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        )}

        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            {t("integrations:cancel")}
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!campaignId || !headline.trim() || publish.isPending}
            loading={publish.isPending}
          >
            <Megaphone className="h-4 w-4" />
            {t("integrations:publishCreative")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
