import { createAuthClient } from "@neondatabase/neon-js/auth"
import { BetterAuthReactAdapter } from "@neondatabase/neon-js/auth/react/adapters"

/**
 * Client Managed Better Auth (Neon Auth).
 * VITE_NEON_AUTH_URL est fournie par la console Neon (Auth > Configuration).
 */
export const authClient = createAuthClient(import.meta.env.VITE_NEON_AUTH_URL ?? "", {
  adapter: BetterAuthReactAdapter({
    fetchOptions: { credentials: "include" },
  }),
})

let cachedToken: string | null = null
let cachedAt = 0

/**
 * Récupère le JWT de la session (token EdDSA, ~15 min de validité),
 * mis en cache pour éviter un appel réseau à chaque requête API.
 */
export async function getAccessToken(): Promise<string | null> {
  if (cachedToken && Date.now() - cachedAt < 10 * 60 * 1000) {
    return cachedToken
  }
  const { data } = await authClient.getSession()
  const token = data?.session?.token ?? null
  cachedToken = token
  cachedAt = Date.now()
  return token
}

export function clearAccessTokenCache(): void {
  cachedToken = null
  cachedAt = 0
}
