import { useMutation, useQuery } from "@tanstack/react-query"

import { authClient, clearAccessTokenCache } from "@/lib/auth"

import { fetchMe } from "./api"

export function useSignOut() {
  return useMutation({
    mutationFn: () => authClient.signOut(),
    onSettled: () => clearAccessTokenCache(),
  })
}

/** useMe expose le profil backend (rôle, date d'inscription). Partage le
 * cache de la requête de warm-up lancée par AuthProvider. */
export function useMe() {
  return useQuery({
    queryKey: ["auth", "me"],
    queryFn: fetchMe,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })
}
