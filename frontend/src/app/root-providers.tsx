import { Outlet, useNavigate } from "react-router"
import { NeonAuthUIProvider } from "@neondatabase/auth-ui"

import { AuthProvider } from "@/features/auth"
import { authClient } from "@/lib/auth"

/**
 * Fournit l'UI Better Auth (login/signup/callback) en la branchant sur
 * React Router (navigation interne via `navigate`).
 */
export default function RootProviders() {
  const navigate = useNavigate()

  return (
    <NeonAuthUIProvider
      authClient={authClient}
      navigate={navigate}
      redirectTo="/dashboard"
      social={{ providers: ["google"] }}
    >
      <AuthProvider>
        <Outlet />
      </AuthProvider>
    </NeonAuthUIProvider>
  )
}
