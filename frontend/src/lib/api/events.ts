import { useEffect, useRef } from "react"

import { getAccessToken } from "@/lib/auth"

const API_URL = import.meta.env.VITE_API_URL ?? ""

/** Événement interne émis après chaque reconnexion (resync client). */
export const RECONNECTED_EVENT = "__reconnected"

type Listener = (data: unknown) => void
type StatusListener = (connected: boolean) => void

const HEARTBEAT_TIMEOUT_MS = 60_000
const MAX_BACKOFF_MS = 15_000

/**
 * Canal temps réel unique (SSE GET /api/v1/events) : un seul tunnel
 * server→client pour tous les événements (chat.*, job.updated…).
 * Fetch + ReadableStream (EventSource ne supporte pas l'header Authorization).
 */
class EventsClient {
  private listeners = new Map<string, Set<Listener>>()
  private statusListeners = new Set<StatusListener>()
  private controller: AbortController | null = null
  private running = false
  private attempts = 0

  /** Abonne un handler à un type d'événement. Renvoie la désinscription. */
  on(type: string, fn: Listener): () => void {
    let set = this.listeners.get(type)
    if (!set) {
      set = new Set()
      this.listeners.set(type, set)
    }
    set.add(fn)
    return () => {
      set?.delete(fn)
      if (set && set.size === 0) this.listeners.delete(type)
    }
  }

  /** Abonne un handler aux changements de connexion. */
  onStatus(fn: StatusListener): () => void {
    this.statusListeners.add(fn)
    return () => this.statusListeners.delete(fn)
  }

  start() {
    if (this.running) return
    this.running = true
    void this.loop()
  }

  stop() {
    this.running = false
    this.controller?.abort()
    this.controller = null
    this.setConnected(false)
  }

  private setConnected(connected: boolean) {
    for (const fn of this.statusListeners) fn(connected)
  }

  private dispatch(type: string, data: unknown) {
    for (const fn of this.listeners.get(type) ?? []) fn(data)
  }

  private async loop(): Promise<void> {
    while (this.running) {
      try {
        await this.connect()
        this.attempts = 0
        this.setConnected(true)
        this.dispatch(RECONNECTED_EVENT, null)
      } catch {
        this.setConnected(false)
        if (!this.running) break
      }
      if (!this.running) break
      const backoff = Math.min(1000 * 2 ** this.attempts, MAX_BACKOFF_MS)
      this.attempts += 1
      await new Promise((r) => setTimeout(r, backoff))
    }
  }

  private async connect(): Promise<void> {
    const token = await getAccessToken()
    const controller = new AbortController()
    this.controller = controller

    const res = await fetch(`${API_URL}/api/v1/events`, {
      credentials: "include",
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      signal: controller.signal,
    })
    if (!res.ok || !res.body) {
      // 401 : pas authentifié — on stoppe (restart au prochain mount).
      if (res.status === 401) this.running = false
      throw new Error(`events: HTTP ${res.status}`)
    }

    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ""
    let watchdog = this.resetWatchdog(controller)

    try {
      for (;;) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        clearTimeout(watchdog)
        watchdog = this.resetWatchdog(controller)

        let idx: number
        while ((idx = buffer.indexOf("\n\n")) !== -1) {
          const raw = buffer.slice(0, idx)
          buffer = buffer.slice(idx + 2)
          this.parseEvent(raw)
        }
      }
    } finally {
      clearTimeout(watchdog)
      reader.releaseLock()
      if (this.controller === controller) this.controller = null
      this.setConnected(false)
    }
  }

  /** Connexion muette > 60 s (heartbeat perdu) → on coupe et on reconnecte. */
  private resetWatchdog(controller: AbortController): ReturnType<typeof setTimeout> {
    return setTimeout(() => controller.abort(), HEARTBEAT_TIMEOUT_MS)
  }

  private parseEvent(raw: string) {
    let type = "message"
    const dataLines: string[] = []
    for (const line of raw.split("\n")) {
      if (line.startsWith(":")) continue // commentaire/heartbeat
      if (line.startsWith("event:")) type = line.slice(6).trim()
      else if (line.startsWith("data:")) dataLines.push(line.slice(5).trim())
    }
    const data = dataLines.join("\n")
    if (!data) return
    try {
      this.dispatch(type, JSON.parse(data))
    } catch {
      this.dispatch(type, data)
    }
  }
}

export const appEvents = new EventsClient()

/** useAppEvent abonne un handler React à un type d'événement du canal. */
export function useAppEvent(type: string, handler: (data: never) => void) {
  const ref = useRef(handler)
  ref.current = handler

  useEffect(() => {
    return appEvents.on(type, (data) => ref.current(data as never))
  }, [type])
}

/** useEventsConnection démarre/arrête la connexion SSE avec l'app active. */
export function useEventsConnection() {
  useEffect(() => {
    appEvents.start()
    return () => appEvents.stop()
  }, [])
}
