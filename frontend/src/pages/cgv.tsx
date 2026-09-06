import { useTranslation } from "react-i18next"

import { LegalLayout, type LegalDoc } from "./legal"

const fr: LegalDoc = {
  updatedLabel: "Dernière mise à jour",
  updated: "6 septembre 2026",
  intro:
    "Les présentes Conditions Générales de Vente (« CGV ») régissent l'achat de crédits de la Plateforme AfriLaunch et la fourniture des services associés (création de produits digitaux). Elles s'appliquent à tout utilisateur achetant des crédits. Les ventes de vos propres produits digitaux à vos clients finaux, réalisées via votre boutique connectée (ex. Chariow), ne sont pas des ventes AfriLaunch : elles sont régies par vos propres CGV et celles de la plateforme concernée.",
  sections: [
    {
      heading: "1. Produits et prix",
      paragraphs: [
        "AfriLaunch propose des packs de crédits utilisables pour générer du contenu (recherche, idées, ebook, assets marketing, vidéos). Les packs et leurs prix (en Francs CFA — XOF) sont affichés sur la page Crédits. Les prix s'entendent toutes taxes comprises dans la mesure applicable. AfriLaunch se réserve le droit de modifier les prix à tout moment ; le prix applicable est celui affiché au moment de l'achat.",
      ],
    },
    {
      heading: "2. Commande et paiement",
      paragraphs: [
        "La commande est confirmée lorsque le paiement a été accepté par notre partenaire de paiement Mobile Money (ex. PawaPay). Le paiement s'effectue sur une page sécurisée hébergée par le partenaire. Aucune donnée bancaire ou de compte mobile n'est traitée ni stockée par AfriLaunch.",
      ],
      bullets: [
        "Le paiement doit être réalisé dans un pays et un mode de paiement acceptés ;",
        "En cas de refus du paiement ou d'annulation, la commande n'est pas exécutée.",
      ],
    },
    {
      heading: "3. Délivrance des crédits",
      paragraphs: [
        "Les crédits sont crédités sur votre compte de manière automatique et immédiate après confirmation du paiement par le partenaire (généralement quelques secondes à quelques minutes). Si le crédit n'apparaît pas sous 24 h alors que le paiement a été confirmé, contactez le support muni de votre référence de paiement.",
      ],
    },
    {
      heading: "4. Droit de rétractation",
      paragraphs: [
        "Les crédits constituent un contenu numérique fourni immédiatement et exécuté dès la commande. Conformément au droit de la consommation applicable (notamment l'article L.221-28 du Code de la consommation pour les contenus numériques fournis immédiatement avec accord exprès), vous renoncez expressément à votre droit de rétractation en acceptant la fourniture immédiate du service. Aucun remboursement des crédits non consommés ne sera effectué, sauf disposition légale impérative contraire (notamment crédits défectueux non utilisables).",
      ],
    },
    {
      heading: "5. Crédits offerts et promotions",
      paragraphs: [
        "Des crédits peuvent être offerts (bonus de bienvenue, promotions). Ils sont soumis aux mêmes règles d'utilisation que les crédits achetés, à l'exception qu'ils sont non remboursables en toutes circonstances.",
      ],
    },
    {
      heading: "6. Garanties et responsabilité",
      paragraphs: [
        "AfriLaunch s'engage à fournir les crédits commandés et à rendre le service fonctionnel. Les contenus générés par intelligence artificielle étant fournis « en l'état », AfriLaunch ne garantit pas un résultat précis (qualité, exactitude, performance commerciale) et n'est pas responsable de l'usage que vous faites des contenus. En cas de défaillance technique avérée empêchant l'utilisation des crédits achetés, AfriLaunch pourra, à son choix, créditer des crédits de remplacement ou procéder au remboursement des crédits affectés.",
      ],
    },
    {
      heading: "7. Facturation et données",
      paragraphs: [
        "Votre historique d'achats et vos transactions de crédits sont consultables dans votre espace (page Crédits). Aucune facture papier n'est émise par défaut ; sur demande, une facture/justificatif peut être fourni dans la mesure exigée par la réglementation locale.",
      ],
    },
    {
      heading: "8. Litiges",
      paragraphs: [
        "Les présentes CGV sont régies par la législation applicable au siège d'AfriLaunch (Bénin). Les parties privilégient une résolution amiable ; à défaut, les juridictions compétentes seront saisies. Les dispositions impératives protectrices des consommateurs restent applicables.",
      ],
    },
    {
      heading: "9. Contact",
      paragraphs: [
        "Pour toute réclamation relative à une commande : page Support de votre espace AfriLaunch.",
      ],
    },
  ],
}

const en: LegalDoc = {
  updatedLabel: "Last updated",
  updated: "September 6, 2026",
  intro:
    "These Terms of Sale (« Terms of Sale ») govern the purchase of credits on AfriLaunch and the provision of the associated services (digital product creation). They apply to any user purchasing credits. Sales of your own digital products to your end customers through your connected store (e.g. Chariow) are not AfriLaunch sales: they are governed by your own terms and those of the relevant platform.",
  sections: [
    {
      heading: "1. Products and prices",
      paragraphs: [
        "AfriLaunch offers credit packs used to generate content (research, ideas, ebook, marketing assets, videos). Packs and their prices (in CFA Francs — XOF) are displayed on the Credits page. Prices include applicable taxes where relevant. AfriLaunch may change prices at any time; the applicable price is the one displayed at the time of purchase.",
      ],
    },
    {
      heading: "2. Order and payment",
      paragraphs: [
        "Your order is confirmed once payment is accepted by our Mobile Money partner (e.g. PawaPay). Payment takes place on a secure page hosted by the partner. AfriLaunch never processes or stores bank or mobile-account details.",
      ],
      bullets: [
        "Payment must be made from an accepted country and payment method;",
        "If payment is refused or cancelled, the order is not fulfilled.",
      ],
    },
    {
      heading: "3. Delivery of credits",
      paragraphs: [
        "Credits are credited to your account automatically and immediately after payment confirmation by the partner (usually seconds to minutes). If credits do not appear within 24h despite confirmed payment, contact support with your payment reference.",
      ],
    },
    {
      heading: "4. Withdrawal right",
      paragraphs: [
        "Credits are digital content delivered immediately upon order. Under applicable consumer law (including article L.221-28 of the French Consumer Code for digital content supplied immediately with express consent), you expressly waive your withdrawal right by accepting immediate supply. Unused credits are not refunded, except where mandatory law provides otherwise (e.g. unusable defective credits).",
      ],
    },
    {
      heading: "5. Bonus and promotional credits",
      paragraphs: [
        "Credits may be offered (welcome bonus, promotions). They follow the same usage rules as purchased credits, except that they are non-refundable under any circumstances.",
      ],
    },
    {
      heading: "6. Warranties and liability",
      paragraphs: [
        "AfriLaunch undertakes to deliver ordered credits and keep the service functional. AI-generated content is provided « as is »: AfriLaunch does not guarantee a specific result (quality, accuracy, commercial performance) and is not liable for your use of the content. In case of a proven technical failure preventing use of purchased credits, AfriLaunch may, at its discretion, credit replacement credits or refund the affected credits.",
      ],
    },
    {
      heading: "7. Invoicing and data",
      paragraphs: [
        "Your purchase history and credit transactions are available in your account (Credits page). No paper invoice is issued by default; on request, an invoice/receipt may be provided where required by local regulation.",
      ],
    },
    {
      heading: "8. Disputes",
      paragraphs: [
        "These Terms of Sale are governed by the law applicable at AfriLaunch's registered office (Benin). The parties prefer an amicable resolution; otherwise the competent courts will have jurisdiction. Mandatory protective provisions for consumers remain applicable.",
      ],
    },
    {
      heading: "9. Contact",
      paragraphs: [
        "For any complaint about an order: Support page in your AfriLaunch account.",
      ],
    },
  ],
}

export default function CgvPage() {
  const { i18n } = useTranslation()
  return <LegalLayout doc={i18n.language.startsWith("en") ? en : fr} />
}
