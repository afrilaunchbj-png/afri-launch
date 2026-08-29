import { createContext, useContext, useMemo } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"

import { authKeys, fetchMe } from "./api"
import type { User } from "./types"

interface AuthContextValue {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  invalidate: () => void
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient()

  const { data: user, isLoading } = useQuery({
    queryKey: authKeys.me(),
    queryFn: fetchMe,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })

  const value = useMemo<AuthContextValue>(
    () => ({
      user: user ?? null,
      isLoading,
      isAuthenticated: user != null,
      invalidate: () => queryClient.invalidateQueries({ queryKey: authKeys.me() }),
    }),
    [user, isLoading, queryClient],
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
