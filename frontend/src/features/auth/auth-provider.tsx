import { createContext, useContext, useMemo } from "react"
import { useQuery } from "@tanstack/react-query"

import { authClient } from "@/lib/auth"

import { fetchMe } from "./api"

export interface SessionUser {
  id: string
  name: string
  email: string
  image?: string | null
}

interface AuthContextValue {
  user: SessionUser | null
  isLoading: boolean
  isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data: session, isPending } = authClient.useSession()

  const user = session?.user ?? null
  const isAuthenticated = user != null

  // Déclenche l'upsert du profil local + le bonus de bienvenue côté backend.
  useQuery({
    queryKey: ["auth", "me"],
    queryFn: fetchMe,
    enabled: isAuthenticated,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })

  const value = useMemo<AuthContextValue>(
    () => ({
      user: user
        ? { id: user.id, name: user.name ?? "", email: user.email ?? "", image: user.image ?? null }
        : null,
      isLoading: isPending,
      isAuthenticated,
    }),
    [user, isPending, isAuthenticated],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (ctx === undefined) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return ctx
}
