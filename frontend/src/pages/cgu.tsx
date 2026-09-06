import { useTranslation } from "react-i18next"

import { LegalLayout, type LegalDoc } from "./legal"

const fr: LegalDoc = {
  updatedLabel: "Dernière mise à jour",
  updated: "6 septembre 2026",
  intro:
    "Bienvenue sur AfriLaunch. Les présentes Conditions Générales d'Utilisation (« CGU ») encadrent l'accès et l'usage de la plateforme AfriLaunch accessible depuis le site afrilaunch (la « Plateforme »), éditée par AfriLaunch. En créant un compte et/ou en utilisant la Plateforme, vous acceptez sans réserve les présentes CGU.",
  sections: [
    {
      heading: "1. Objet du service",
      paragraphs: [
        "AfriLaunch est un service en ligne qui aide les entrepreneurs, créateurs et experts africains à transformer une opportunité de marché en produit digital commercialisable : recherche de marché, génération d'idées, rédaction d'ebook, création d'assets marketing (couvertures, affiches, vidéos publicitaires) et de pages de vente, ainsi que la mise en relation de votre boutique avec des solutions de commerce et de paiement (dont Chariow et des providers de paiement Mobile Money).",
      ],
    },
    {
      heading: "2. Compte utilisateur",
      paragraphs: [
        "L'accès à la Plateforme nécessite la création d'un compte. Vous vous engagez à fournir des informations exactes et à jour, à préserver la confidentialité de vos accès et à nous notifier toute utilisation non autorisée. Vous êtes responsable de toute activité réalisée depuis votre compte.",
      ],
    },
    {
      heading: "3. Crédits et utilisation",
      paragraphs: [
        "Certaines fonctionnalités (génération de contenu, recherches avancées, vidéos…) consomment des crédits achetables en packs. Les crédits sont personnels, non transférables et non remboursables. Leur solde est visible dans votre espace. Toute utilisation abusive (automatisation excessive, contournement des limites, comptes multiples frauduleux) peut entraîner la suspension du compte sans préavis.",
      ],
    },
    {
      heading: "4. Contenus générés par l'intelligence artificielle",
      paragraphs: [
        "Les contenus (textes, ebooks, images, vidéos) sont générés par des modèles d'intelligence artificielle à partir de vos instructions. Ils sont fournis « en l'état » :",
      ],
      bullets: [
        "vous restez responsable de la vérification, de la correction et de l'utilisation de ces contenus (exactitude factuelle, droits d'auteur, conformité réglementaire de votre secteur) ;",
        "AfriLaunch s'efforce d'éviter les statistiques inventées en s'appuyant sur des sources, mais ne garantit ni l'exactitude, ni l'exhaustivité, ni l'absence d'erreurs ou d'hallucinations des contenus générés ;",
        "il vous appartient de contrôler le résultat avant toute publication ou commercialisation.",
      ],
    },
    {
      heading: "5. Ventes de vos produits et plateformes tierces",
      paragraphs: [
        "AfriLaunch peut vous permettre de connecter votre boutique (ex. Chariow) et de générer des liens de vente. Les ventes conclues entre vous et vos clients finaux sont vos propres ventes : AfriLaunch agit comme un outil et n'est pas partie au contrat de vente, au paiement, à la délivrance ou à la gestion client. Ces opérations sont régies par les conditions de la plateforme concernée (notamment Chariow) et par vos propres obligations (CGV, droit de la consommation applicable, fiscalité).",
      ],
    },
    {
      heading: "6. Propriété intellectuelle",
      paragraphs: [
        "La Plateforme et son interface, sa marque et ses contenus éditoriaux appartiennent à AfriLaunch. Vous conservez la propriété des contenus que vous créez ou importez via le service. Vous nous concédez une licence limitée pour exploiter ces contenus aux seules fins de vous fournir le service (stockage, traitement, génération, rendu). Vous garantissez disposer des droits nécessaires sur les éléments que vous fournissez (notamment les visuels générés à partir de vos instructions et le matériel que vous téléversez).",
      ],
    },
    {
      heading: "7. Données personnelles",
      paragraphs: [
        "Le traitement de vos données est décrit dans notre politique de confidentialité (données de compte, préférences, contenus, journaux d'activité). Vos contenus ne sont pas revendus. Vous pouvez demander la suppression de votre compte et de vos données dans les conditions prévues par la réglementation applicable (dont le RGPD pour les utilisateurs concernés).",
      ],
    },
    {
      heading: "8. Disponibilité et responsabilité",
      paragraphs: [
        "Le service est fourni « en l'état » et accessible au mieux avec les moyens techniques disponibles. AfriLaunch ne garantit pas une disponibilité ininterrompue. Dans les limites autorisées par la loi, AfriLaunch n'est pas responsable des dommages indirects (perte de données, perte de chiffre d'affaires, interruption d'activité) résultant de l'usage du service ou de contenus générés.",
      ],
    },
    {
      heading: "9. Comportements interdits",
      bullets: [
        "utiliser le service à des fins illégales, frauduleuses ou portant atteinte aux droits de tiers ;",
        "tenter d'accéder aux comptes ou données d'autres utilisateurs ;",
        "reproduire, revendre ou exploiter la Plateforme sans autorisation ;",
        "diffuser des contenus illicites, diffamatoires ou portant atteinte à l'ordre public.",
      ],
    },
    {
      heading: "10. Suspension et résiliation",
      paragraphs: [
        "Vous pouvez fermer votre compte à tout moment. AfriLaunch peut suspendre ou résilier un compte en cas de manquement aux présentes CGU, de fraude ou d'usage abusif, après information sauf urgence. Les crédits non consommés ne sont pas remboursés en cas de résiliation pour manquement.",
      ],
    },
    {
      heading: "11. Droit applicable et litiges",
      paragraphs: [
        "Les présentes CGU sont régies par la législation applicable au siège d'AfriLaunch (Bénin). En cas de litige, les parties rechercheront d'abord une solution amiable. À défaut, le litige sera porté devant les juridictions compétentes. Pour les consommateurs, les dispositions protectrices impératives de leur pays de résidence restent applicables.",
      ],
    },
    {
      heading: "12. Contact",
      paragraphs: [
        "Pour toute question relative aux présentes CGU : support accessible depuis votre espace AfriLaunch (page Support).",
      ],
    },
  ],
}

const en: LegalDoc = {
  updatedLabel: "Last updated",
  updated: "September 6, 2026",
  intro:
    "Welcome to AfriLaunch. These Terms of Use (« Terms ») govern access to and use of the AfriLaunch platform (the « Platform »), operated by AfriLaunch. By creating an account and/or using the Platform you unreservedly accept these Terms.",
  sections: [
    {
      heading: "1. Purpose of the service",
      paragraphs: [
        "AfriLaunch is an online service that helps African entrepreneurs, creators and experts turn a market opportunity into a sellable digital product: market research, idea generation, ebook writing, marketing assets (covers, posters, video ads), sales pages, and connecting your store to commerce and payment solutions (including Chariow and Mobile Money providers).",
      ],
    },
    {
      heading: "2. User account",
      paragraphs: [
        "Using the Platform requires creating an account. You undertake to provide accurate, up-to-date information, keep your credentials confidential and notify us of any unauthorised use. You are responsible for all activity on your account.",
      ],
    },
    {
      heading: "3. Credits and usage",
      paragraphs: [
        "Some features (content generation, advanced research, videos…) consume credits that you can purchase in packs. Credits are personal, non-transferable and non-refundable. Your balance is visible in your account. Any abusive use (excessive automation, bypassing limits, fraudulent multiple accounts) may lead to suspension without notice.",
      ],
    },
    {
      heading: "4. AI-generated content",
      paragraphs: [
        "Content (texts, ebooks, images, videos) is generated by artificial intelligence from your instructions. It is provided « as is »:",
      ],
      bullets: [
        "you remain responsible for verifying, correcting and using such content (factual accuracy, copyright, sector-specific compliance);",
        "AfriLaunch works to avoid invented statistics by relying on sources, but does not guarantee accuracy, completeness or the absence of errors or hallucinations in generated content;",
        "you must review the output before any publication or sale.",
      ],
    },
    {
      heading: "5. Sales of your products and third-party platforms",
      paragraphs: [
        "AfriLaunch may let you connect your store (e.g. Chariow) and generate sales links. Sales concluded between you and your end customers are your own sales: AfriLaunch acts as a tool and is not a party to the sale contract, payment, delivery or customer management. Those transactions are governed by the relevant platform's terms (including Chariow) and by your own obligations (terms of sale, applicable consumer law, tax).",
      ],
    },
    {
      heading: "6. Intellectual property",
      paragraphs: [
        "The Platform, its interface, brand and editorial content belong to AfriLaunch. You keep ownership of the content you create or import. You grant us a limited licence to process such content solely to provide the service (storage, processing, generation, rendering). You warrant that you hold the necessary rights on the elements you provide.",
      ],
    },
    {
      heading: "7. Personal data",
      paragraphs: [
        "Processing of your data is described in our privacy policy. Your content is not resold. You may request account and data deletion under applicable law (including GDPR where relevant).",
      ],
    },
    {
      heading: "8. Availability and liability",
      paragraphs: [
        "The service is provided « as is ». AfriLaunch does not guarantee uninterrupted availability. To the extent permitted by law, AfriLaunch is not liable for indirect damages (data loss, loss of revenue, business interruption) resulting from the use of the service or generated content.",
      ],
    },
    {
      heading: "9. Prohibited conduct",
      bullets: [
        "using the service for unlawful, fraudulent purposes or infringing third-party rights;",
        "attempting to access other users' accounts or data;",
        "reproducing, reselling or exploiting the Platform without authorisation;",
        "distributing unlawful, defamatory or unlawful content.",
      ],
    },
    {
      heading: "10. Suspension and termination",
      paragraphs: [
        "You may close your account at any time. AfriLaunch may suspend or terminate an account in case of breach of these Terms, fraud or abuse, with notice except in urgent cases. Unused credits are not refunded upon termination for breach.",
      ],
    },
    {
      heading: "11. Governing law and disputes",
      paragraphs: [
        "These Terms are governed by the law applicable at AfriLaunch's registered office (Benin). In case of dispute the parties will first seek an amicable solution; otherwise the competent courts will have jurisdiction. Mandatory protective provisions of consumers' country of residence remain applicable.",
      ],
    },
    {
      heading: "12. Contact",
      paragraphs: [
        "For any question: support accessible from your AfriLaunch account (Support page).",
      ],
    },
  ],
}

export default function CguPage() {
  const { i18n } = useTranslation()
  return <LegalLayout doc={i18n.language.startsWith("en") ? en : fr} />
}
