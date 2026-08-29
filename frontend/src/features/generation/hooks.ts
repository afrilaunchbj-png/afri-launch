import { useQuery } from "@tanstack/react-query"

import { fetchJob } from "./api"

export const generationKeys = {
  job: (id: string) => ["generation", "job", id] as const,
}

/** useJob interroge le statut d'un job, en re-pollant tant qu'il n'est pas terminé. */
export function useJob(id: string | null) {
  return useQuery({
    queryKey: generationKeys.job(id ?? ""),
    queryFn: () => fetchJob(id as string),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.status
      return status === "pending" || status === "processing" ? 2000 : false
    },
  })
}
