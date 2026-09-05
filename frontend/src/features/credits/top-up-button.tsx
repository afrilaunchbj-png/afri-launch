import { Link } from "react-router"
import { useTranslation } from "react-i18next"
import { Coins } from "lucide-react"

import { Button, type ButtonProps } from "@/components/ui/button"
import { usePlans } from "@/features/credits/hooks"

export type TopUpButtonProps = Omit<ButtonProps, "children" | "asChild">

/**
 * TopUpButton : CTA « Recharger » menant à la page /credits (packs de
 * crédits + checkout Mobile Money). Masqué tant qu'aucun provider de
 * paiement n'est actif côté backend (plans.enabled).
 */
export function TopUpButton(props: TopUpButtonProps) {
  const { t } = useTranslation()
  const { data: plans } = usePlans()

  if (!plans?.enabled) return null

  return (
    <Button asChild {...props}>
      <Link to="/credits">
        <Coins className="h-4 w-4" />
        {t("credits:recharge")}
      </Link>
    </Button>
  )
}
