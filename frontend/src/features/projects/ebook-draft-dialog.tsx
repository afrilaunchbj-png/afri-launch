import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Eye, Save } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { assetDownloadPath } from "@/features/projects/api"
import { useSaveEbookDraft } from "@/features/projects/hooks"
import { api } from "@/lib/api/client"
import { isAppError } from "@/lib/errors"

interface Block {
  text: string
  tag: string
}

/** Récupère le texte d'un asset HTML (brouillon). */
async function fetchAssetText(assetId: string): Promise<string> {
  const blob = await api.download(assetDownloadPath(assetId))
  return blob.text()
}

/**
 * EbookDraftDialog : prévisualisation + édition « texte simple » d'un
 * brouillon d'ebook. Le HTML/CSS d'origine est conservé ; seuls les textes
 * des blocs (h1-h4, p, li, blockquote) sont modifiables.
 */
export function EbookDraftDialog({
  assetId,
  projectId,
  onClose,
  onSaved,
}: {
  assetId: string
  projectId: string
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const save = useSaveEbookDraft()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [blocks, setBlocks] = useState<Block[]>([])
  const [html, setHtml] = useState("")
  const [preview, setPreview] = useState(false)
  const docRef = useRef<Document | null>(null)
  const elsRef = useRef<HTMLElement[]>([])

  useEffect(() => {
    let active = true
    fetchAssetText(assetId)
      .then((raw) => {
        if (!active) return
        const doc = new DOMParser().parseFromString(raw, "text/html")
        docRef.current = doc
        const els: HTMLElement[] = []
        doc.querySelectorAll("body").forEach((body) => {
          body.querySelectorAll("h1,h2,h3,h4,p,li,blockquote").forEach((el) => {
            const e = el as HTMLElement
            if (e.querySelector("h1,h2,h3,h4,p,li,blockquote")) return
            if (!e.textContent?.trim()) return
            els.push(e)
          })
        })
        elsRef.current = els
        setBlocks(els.map((e) => ({ text: e.textContent ?? "", tag: e.tagName.toLowerCase() })))
        setHtml(raw)
        setLoading(false)
      })
      .catch(() => {
        if (active) {
          setError(t("common.genericError"))
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
  }, [assetId, t])

  const updateBlock = (index: number, value: string) => {
    setBlocks((prev) => prev.map((b, i) => (i === index ? { ...b, text: value } : b)))
    if (elsRef.current[index]) elsRef.current[index].textContent = value
    if (docRef.current) setHtml(docRef.current.documentElement.outerHTML)
  }

  const handleSave = () => {
    if (!docRef.current) return
    const final = docRef.current.documentElement.outerHTML
    save.mutate(
      { id: projectId, content: final },
      {
        onSuccess: () => {
          toast.success(t("projects:draftSaved"))
          onSaved()
          onClose()
        },
        onError: (e) => toast.error(isAppError(e) ? e.message : t("common.genericError")),
      },
    )
  }

  return (
    <Dialog open onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-4xl">
        <DialogHeader>
          <DialogTitle>{t("projects:ebookDraftTitle")}</DialogTitle>
        </DialogHeader>

        {loading ? (
          <p className="text-sm text-muted-foreground">{t("common.loading")}</p>
        ) : error ? (
          <p className="text-sm text-destructive">{error}</p>
        ) : (
          <div className="space-y-3">
            <div className="flex items-center justify-between gap-2">
              <p className="text-xs text-muted-foreground">{t("projects:ebookDraftHint")}</p>
              <Button size="sm" variant="outline" onClick={() => setPreview((p) => !p)}>
                <Eye className="h-4 w-4" />
                {preview ? t("projects:editMode") : t("projects:previewMode")}
              </Button>
            </div>

            {preview ? (
              <iframe
                title="preview"
                sandbox=""
                srcDoc={html}
                className="h-[60vh] w-full rounded-lg border bg-white"
              />
            ) : (
              <div className="max-h-[60vh] space-y-2 overflow-y-auto rounded-lg border p-3">
                {blocks.map((b, i) => (
                  <div key={i}>
                    <span className="mb-0.5 block text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
                      {b.tag}
                    </span>
                    <Textarea
                      rows={b.tag === "li" || b.tag === "h1" || b.tag === "h2" ? 1 : 3}
                      value={b.text}
                      onChange={(e) => updateBlock(i, e.target.value)}
                      className="font-sans text-sm"
                    />
                  </div>
                ))}
              </div>
            )}

            <div className="flex justify-end gap-2">
              <Button variant="ghost" onClick={onClose}>
                {t("integrations:cancel")}
              </Button>
              <Button onClick={handleSave} loading={save.isPending} disabled={loading || !!error}>
                <Save className="h-4 w-4" />
                {t("projects:draftSave")}
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
