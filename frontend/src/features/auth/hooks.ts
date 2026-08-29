import { useMutation } from "@tanstack/react-query"

import { authClient, clearAccessTokenCache } from "@/lib/auth"

export function useSignOut() {
  return useMutation({
    mutationFn: () => authClient.signOut(),
    onSettled: () => clearAccessTokenCache(),
  })
}
