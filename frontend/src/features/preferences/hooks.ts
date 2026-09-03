import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import type { Preferences } from "./api"
import { fetchPreferences, updatePreferences } from "./api"

export const preferencesKeys = {
  all: ["preferences"] as const,
}

/**
 * usePreferences charge les préférences du compte (langue, thème) — la
 * source de vérité une fois l'utilisateur connecté.
 */
export function usePreferences() {
  return useQuery({
    queryKey: preferencesKeys.all,
    queryFn: fetchPreferences,
    staleTime: 5 * 60 * 1000,
  })
}

/**
 * useUpdatePreferences persiste un changement de préférence avec mise à
 * jour optimiste (l'UI réagit instantanément, la DB suit).
 */
export function useUpdatePreferences() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updatePreferences,
    onMutate: async (input: Partial<Preferences>) => {
      await queryClient.cancelQueries({ queryKey: preferencesKeys.all })
      const previous = queryClient.getQueryData<Preferences>(preferencesKeys.all)
      queryClient.setQueryData<Preferences>(preferencesKeys.all, (old) =>
        old ? { ...old, ...input } : ({ language: "fr", theme: "system", ...input } as Preferences),
      )
      return { previous }
    },
    onError: (_error, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(preferencesKeys.all, context.previous)
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: preferencesKeys.all })
    },
  })
}
