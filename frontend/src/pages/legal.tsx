import { useTranslation } from "react-i18next"
import { Link } from "react-router"

export interface LegalSection {
  heading: string
  paragraphs?: string[]
  bullets?: string[]
}

export interface LegalDoc {
  updatedLabel: string
  updated: string
  intro: string
  sections: LegalSection[]
}

/** Gabarit de page légale (CGU/CGV) rendue depuis la landing. */
export function LegalLayout({ doc }: { doc: LegalDoc }) {
  const { t } = useTranslation()
  return (
    <div className="mx-auto max-w-3xl space-y-8 py-10">
      <header className="space-y-2">
        <Link to="/" className="text-sm text-primary hover:underline">
          ← {t("home:legalBack")}
        </Link>
        <p className="text-xs uppercase tracking-wide text-muted-foreground">
          {doc.updatedLabel} : {doc.updated}
        </p>
      </header>

      <section className="prose-sm space-y-6">
        <p className="text-sm leading-relaxed text-muted-foreground">{doc.intro}</p>
        {doc.sections.map((s) => (
          <div key={s.heading}>
            <h2 className="font-display text-xl font-semibold text-primary">{s.heading}</h2>
            {s.paragraphs?.map((p, i) => (
              <p key={i} className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {p}
              </p>
            ))}
            {s.bullets ? (
              <ul className="mt-2 list-disc space-y-1 pl-5 text-sm leading-relaxed text-muted-foreground">
                {s.bullets.map((b, i) => (
                  <li key={i}>{b}</li>
                ))}
              </ul>
            ) : null}
          </div>
        ))}
      </section>
    </div>
  )
}
