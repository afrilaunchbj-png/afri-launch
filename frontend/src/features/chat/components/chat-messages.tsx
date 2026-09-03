import { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"
import { Loader2, Search } from "lucide-react"

import { cn } from "@/lib/utils"

import type { ChatMessage } from "../api"

export interface StreamingState {
  messageId: string
  text: string
  tool: string | null
}

interface ChatMessagesProps {
  messages: ChatMessage[]
  streaming: StreamingState | null
}

function ToolIndicator({ tool }: { tool: string }) {
  const { t } = useTranslation()
  if (tool !== "search") return null
  return (
    <p className="mb-2 inline-flex items-center gap-1.5 text-xs text-muted-foreground">
      <Search className="h-3.5 w-3.5 animate-pulse" />
      {t("chat:searching")}
    </p>
  )
}

export function ChatMessages({ messages, streaming }: ChatMessagesProps) {
  const { t } = useTranslation()
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" })
  }, [messages.length, streaming?.text, streaming?.tool])

  return (
    <div className="flex-1 space-y-4 overflow-y-auto px-4 py-4">
      {messages.map((m) => (
        <div key={m.id} className={cn("flex", m.role === "user" ? "justify-end" : "justify-start")}>
          <div
            className={cn(
              "max-w-[85%] whitespace-pre-wrap rounded-2xl px-4 py-2.5 text-sm leading-relaxed",
              m.role === "user"
                ? "rounded-br-md bg-primary text-primary-foreground"
                : "rounded-bl-md border bg-card text-foreground",
            )}
          >
            {m.content}
          </div>
        </div>
      ))}

      {streaming ? (
        <div className="flex justify-start">
          <div className="max-w-[85%] rounded-2xl rounded-bl-md border bg-card px-4 py-2.5 text-sm leading-relaxed">
            {streaming.tool ? <ToolIndicator tool={streaming.tool} /> : null}
            {streaming.text ? (
              <p className="whitespace-pre-wrap">{streaming.text}</p>
            ) : !streaming.tool ? (
              <p className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                {t("chat:thinking")}
              </p>
            ) : null}
          </div>
        </div>
      ) : null}

      <div ref={bottomRef} />
    </div>
  )
}
