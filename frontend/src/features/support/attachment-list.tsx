import { useTranslation } from "react-i18next"
import { Download, FileText } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  attachmentDownloadPath,
  downloadAttachment,
  type SupportAttachment,
} from "@/features/support/api"

/** AttachmentList affiche les pièces jointes d'un ticket/message (téléchargeables). */
export function AttachmentList({ attachments }: { attachments?: SupportAttachment[] }) {
  const { t } = useTranslation()
  if (!attachments || attachments.length === 0) return null
  return (
    <ul className="flex flex-wrap gap-2">
      {attachments.map((a) => (
        <li key={a.id}>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="h-7 gap-1.5 text-xs"
            onClick={() => downloadAttachment(a.id, a.filename).catch(() => {})}
            title={t("support:downloadAttachment")}
          >
            <FileText className="h-3.5 w-3.5 shrink-0" />
            <span className="max-w-[160px] truncate">{a.filename}</span>
            <Download className="h-3 w-3 shrink-0 opacity-60" />
          </Button>
        </li>
      ))}
    </ul>
  )
}

export { attachmentDownloadPath }
