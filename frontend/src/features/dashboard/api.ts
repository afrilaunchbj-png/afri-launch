import { api, type ApiSingle } from "@/lib/api/client"

export interface SeriesPoint {
  date: string
  value: number
}

export interface DashboardStats {
  projects: number
  conversations: number
  open_tickets: number
  credits_balance: number
  credits_used_30d: number
  jobs_by_status: Record<string, number>
  credits_per_day: SeriesPoint[]
  projects_per_week: SeriesPoint[]
}

export function fetchDashboardStats() {
  return api.get<ApiSingle<DashboardStats>>("/api/v1/dashboard/stats").then((r) => r.data)
}

/** Complète une série journalière avec des zéros sur les `days` derniers jours. */
export function fillDailySeries(points: SeriesPoint[], days = 30): SeriesPoint[] {
  const map = new Map(points.map((p) => [p.date, p.value]))
  const out: SeriesPoint[] = []
  const now = new Date()
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() - i))
    const key = d.toISOString().slice(0, 10)
    out.push({ date: key, value: map.get(key) ?? 0 })
  }
  return out
}

/** Complète une série hebdomadaire (semaines commençant lundi) avec des zéros. */
export function fillWeeklySeries(points: SeriesPoint[], weeks = 12): SeriesPoint[] {
  const map = new Map(points.map((p) => [p.date, p.value]))
  const out: SeriesPoint[] = []
  const now = new Date()
  const mondayOffset = (now.getUTCDay() + 6) % 7
  const monday = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() - mondayOffset)
  for (let i = weeks - 1; i >= 0; i--) {
    const key = new Date(monday - i * 7 * 86400000).toISOString().slice(0, 10)
    out.push({ date: key, value: map.get(key) ?? 0 })
  }
  return out
}
