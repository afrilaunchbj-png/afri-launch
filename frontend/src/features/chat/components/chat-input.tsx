import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { SendHorizonal } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"

interface ChatInputProps {
  disabled: boolean
  onSend: (content: string) => void
}

export function ChatInput({ disabled, onSend }: ChatInputProps) {
  const { t } = useTranslation()
  const [draft, setDraft] = useState("")
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const resize = () => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = "auto"
    el.style.height = `${Math.min(el.scrollHeight, 140)}px`
  }

  const submit = () => {
    const content = draft.trim()
    if (!content || disabled) return
    onSend(content)
    setDraft("")
    requestAnimationFrame(resize)
  }

  return (
    <div className="flex items-end gap-2 border-t bg-background p-3">
      <Textarea
        ref={textareaRef}
        rows={1}
        value={draft}
        placeholder={t("chat:placeholder")}
        className="max-h-[140px] flex-1 resize-none rounded-2xl"
        onChange={(e) => {
          setDraft(e.target.value)
          resize()
        }}
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault()
            submit()
          }
        }}
      />
      <Button
        size="icon"
        className="shrink-0 rounded-full"
        onClick={submit}
        disabled={disabled || !draft.trim()}
        aria-label={t("chat:send")}
      >
        <SendHorizonal className="h-4 w-4" />
      </Button>
    </div>
  )
}
