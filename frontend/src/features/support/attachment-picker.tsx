import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { FileText, Paperclip, X } from "lucide-react"

import { Button } from "@/components/ui/button"

const ACCEPT = "image/png,image/jpeg,image/webp,image/gif,application/pdf"
const MAX_FILES = 4
const MAX_SIZE = 5 * 1024 * 1024 // 5 Mo — doit rester aligné avec le backend

/**
 * AttachmentPicker permet de joindre des captures d'écran / PDF à un ticket
 * (max 4 fichiers de 5 Mo). Le téléversement a lieu à la soumission.
 */
export function AttachmentPicker({
  files,
  onChange,
  disabled,
}: {
  files: File[]
  onChange: (files: File[]) => void
  disabled?: boolean
}) {
  const { t } = useTranslation()

  const handleSelect = (list: FileList | null) => {
    if (!list) return
    const next = [...files]
    for (const file of Array.from(list)) {
      if (next.length >= MAX_FILES) break
      if (file.size > MAX_SIZE) {
        toast(t("support:attachmentTooLarge", { name: file.name }))
        continue
      }
      next.push(file)
    }
    onChange(next)
  }

  return (
    <div className="space-y-2">
      <label className="inline-flex cursor-pointer items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
        <Paperclip className="h-4 w-4" />
        {t("support:attachFiles")}
        <input
          type="file"
          multiple
          accept={ACCEPT}
          className="sr-only"
          disabled={disabled}
          onChange={(e) => {
            handleSelect(e.target.files)
            e.target.value = ""
          }}
        />
      </label>
      {files.length > 0 ? (
        <ul className="space-y-1">
          {files.map((file, i) => (
            <li key={`${file.name}-${i}`} className="flex items-center gap-2 text-sm">
              <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="min-w-0 truncate">{file.name}</span>
              <span className="shrink-0 text-xs text-muted-foreground">
                {(file.size / 1024).toFixed(0)} Ko
              </span>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="ml-auto h-6 w-6"
                aria-label={t("support:removeAttachment")}
                disabled={disabled}
                onClick={() => onChange(files.filter((_, j) => j !== i))}
              >
                <X className="h-3.5 w-3.5" />
              </Button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}
