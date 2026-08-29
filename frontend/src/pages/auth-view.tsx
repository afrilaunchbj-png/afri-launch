import { AuthView } from "@neondatabase/auth-ui"
import { useParams } from "react-router"

/**
 * Page d'authentification générique pour les routes /auth/* (callback,
 * sign-up, forgot-password, reset-password, etc.) pilotées par l'UI Better Auth.
 */
export default function AuthViewPage() {
  const params = useParams()
  const pathname = params["*"] ?? "sign-in"

  return <AuthView pathname={pathname} />
}
