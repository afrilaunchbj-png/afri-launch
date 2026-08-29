import { Navigate, useLocation } from "react-router"

import { LoadingState } from "@/components/states/loading-state"
import { useAuth } from "@/features/auth/auth-provider"

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return <LoadingState />
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  return <>{children}</>
}
