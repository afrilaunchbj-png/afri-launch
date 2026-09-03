import { api, type ApiSingle } from "@/lib/api/client"

export type UserRole = "user" | "superadmin"

export interface BackendUser {
  id: string
  email: string
  full_name: string
  avatar_url?: string | null
  role: UserRole
  created_at: string
}

/** Déclenche (et retourne) le profil backend : upsert + bonus de bienvenue. */
export function fetchMe() {
  return api.get<ApiSingle<BackendUser>>("/api/v1/auth/me").then((r) => r.data)
}
