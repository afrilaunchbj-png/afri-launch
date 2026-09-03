import { useState } from "react"
import { useNavigate, useSearchParams } from "react-router"
import { useTranslation } from "react-i18next"
import { useQueryClient } from "@tanstack/react-query"
import { PanelRightOpen, Rocket, SquarePen } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { RECONNECTED_EVENT, useAppEvent, useEventsConnection } from "@/lib/api/events"
import { isAppError } from "@/lib/errors"

import type { ChatIdea } from "@/features/chat/api"
import { ChatInput } from "@/features/chat/components/chat-input"
import { ChatMessages, type StreamingState } from "@/features/chat/components/chat-messages"
import { ContextPanel } from "@/features/chat/components/context-panel"
import { chatKeys, useConfirmIdea, useConversation, useCreateConversation, useSendChatMessage } from "@/features/chat/hooks"
import { useCreateProject } from "@/features/projects/hooks"

interface ChatDeltaEvent {
  conversation_id: string
  message_id: string
  delta: string
}

interface ChatToolEvent {
  conversation_id: string
  message_id: string
  tool: string
  status: string
  error?: string
}

interface ChatCompletedEvent {
  conversation_id: string
  ideas: ChatIdea[]
}

interface ChatErrorEvent {
  conversation_id: string
  message_id: string
  error: string
}

export default function DiscoverPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [searchParams, setSearchParams] = useSearchParams()
  const conversationId = searchParams.get("c") ?? undefined

  useEventsConnection()

  const { data: detail, isLoading } = useConversation(conversationId)
  const send = useSendChatMessage()
  const createConv = useCreateConversation()
  const confirmIdea = useConfirmIdea(conversationId)
  const createProject = useCreateProject()

  const [streaming, setStreaming] = useState<StreamingState | null>(null)
  const [contextOpen, setContextOpen] = useState(false)

  const invalidate = (id: string) => {
    queryClient.invalidateQueries({ queryKey: chatKeys.detail(id) })
    queryClient.invalidateQueries({ queryKey: ["credits"] })
  }

  useAppEvent("chat.delta", (raw) => {
    const d = raw as ChatDeltaEvent
    if (d.conversation_id !== conversationId) return
    setStreaming((s) =>
      s && s.messageId === d.message_id ? { ...s, text: s.text + d.delta } : { messageId: d.message_id, text: d.delta, tool: null },
    )
  })

  useAppEvent("chat.tool", (raw) => {
    const d = raw as ChatToolEvent
    if (d.conversation_id !== conversationId) return
    setStreaming((s) => (s && s.messageId === d.message_id ? { ...s, tool: d.status === "running" ? d.tool : null } : s))
  })

  useAppEvent("chat.completed", (raw) => {
    const d = raw as ChatCompletedEvent
    if (d.conversation_id !== conversationId) return
    setStreaming(null)
    invalidate(d.conversation_id)
  })

  useAppEvent("chat.error", (raw) => {
    const d = raw as ChatErrorEvent
    if (d.conversation_id !== conversationId) return
    setStreaming(null)
    invalidate(d.conversation_id)
    toast.error(d.error)
  })

  // Reconnexion SSE : resynchroniser la conversation affichée.
  useAppEvent(RECONNECTED_EVENT, () => {
    if (conversationId) queryClient.invalidateQueries({ queryKey: chatKeys.detail(conversationId) })
  })

  const handleSend = (content: string) => {
    if (conversationId) {
      send.mutate({ conversationId, content })
      return
    }
    createConv.mutate(undefined, {
      onSuccess: (conv) => {
        setSearchParams({ c: conv.id }, { replace: true })
        send.mutate({ conversationId: conv.id, content })
      },
      onError: (error) => toast.error(isAppError(error) ? error.message : t("common.genericError")),
    })
  }

  const handleNewChat = () => {
    setStreaming(null)
    setSearchParams({}, { replace: true })
  }

  const handleConfirmIdea = (idea: ChatIdea) => confirmIdea.mutate(idea.id)

  const handleCreateProject = (idea: ChatIdea) => {
    createProject.mutate(
      { idea_id: idea.id, opportunity_id: idea.opportunity_id ?? null, title: idea.title },
      { onSuccess: (p) => navigate(`/projects/${p.id}`) },
    )
  }

  const sending = send.isPending || createConv.isPending
  const suggestions = [t("chat:suggestions.s1"), t("chat:suggestions.s2"), t("chat:suggestions.s3")]

  const panel = (
    <ContextPanel
      detail={detail}
      onConfirmIdea={handleConfirmIdea}
      onCreateProject={handleCreateProject}
      creating={createProject.isPending || confirmIdea.isPending}
    />
  )

  return (
    <div className="flex h-[calc(100dvh-13rem)] min-h-[420px] gap-4 md:h-[calc(100dvh-9rem)]">
      <Card className="flex min-h-0 flex-1 flex-col overflow-hidden p-0">
        <header className="flex items-center justify-between gap-2 border-b px-4 py-3">
          <div className="min-w-0">
            <h1 className="truncate font-display text-base font-semibold text-primary">
              {detail?.title || t("chat:title")}
            </h1>
            <p className="text-xs text-muted-foreground">{t("chat:subtitle")}</p>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            <Sheet open={contextOpen} onOpenChange={setContextOpen}>
              <SheetTrigger asChild>
                <Button variant="ghost" size="icon" className="lg:hidden" aria-label={t("chat:context")}>
                  <PanelRightOpen className="h-4 w-4" />
                </Button>
              </SheetTrigger>
              <SheetContent side="right" className="w-full max-w-sm overflow-y-auto p-4 sm:max-w-md">
                <SheetHeader className="p-0 pb-2">
                  <SheetTitle>{t("chat:context")}</SheetTitle>
                  <SheetDescription className="sr-only">{t("chat:subtitle")}</SheetDescription>
                </SheetHeader>
                {panel}
              </SheetContent>
            </Sheet>
            <Button variant="ghost" size="icon" onClick={handleNewChat} aria-label={t("chat:newChat")}>
              <SquarePen className="h-4 w-4" />
            </Button>
          </div>
        </header>

        {conversationId ? (
          <ChatMessages messages={detail?.messages ?? []} streaming={streaming} />
        ) : (
          <div className="flex flex-1 flex-col items-center justify-center gap-4 px-6 text-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary text-primary-foreground">
              <Rocket className="h-7 w-7" />
            </div>
            <div>
              <h2 className="font-display text-lg font-semibold text-primary">{t("chat:welcomeTitle")}</h2>
              <p className="mt-1 max-w-md text-sm text-muted-foreground">{t("chat:welcomeDesc")}</p>
            </div>
            <div className="flex flex-wrap justify-center gap-2">
              {suggestions.map((s) => (
                <Button
                  key={s}
                  variant="outline"
                  size="sm"
                  className="rounded-full"
                  disabled={sending}
                  onClick={() => handleSend(s)}
                >
                  {s}
                </Button>
              ))}
            </div>
            {isLoading ? <p className="text-xs text-muted-foreground">{t("common.loading")}</p> : null}
          </div>
        )}

        <ChatInput disabled={sending} onSend={handleSend} />
      </Card>

      <aside className="hidden w-80 shrink-0 lg:block xl:w-96">
        {panel}
      </aside>
    </div>
  )
}
