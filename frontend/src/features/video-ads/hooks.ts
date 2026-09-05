import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { fetchJob } from "@/features/generation/api"

import { generateVideoAd, type GenerateVideoAdInput, type VideoAdJob } from "./api"

export const videoAdKeys = {
  job: (id: string) => ["video-ads", "job", id] as const,
}

/** useGenerateVideoAd lance la génération (chaque appel consomme des crédits). */
export function useGenerateVideoAd(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input?: GenerateVideoAdInput) => generateVideoAd(projectId, input),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["projects", "detail", projectId] })
      queryClient.invalidateQueries({ queryKey: ["projects", "assets", projectId] })
      queryClient.invalidateQueries({ queryKey: ["credits"] })
    },
  })
}

/**
 * useVideoJob suit un job vidéo (poll de secours, 5 s — le détail des étapes
 * arrive en temps réel via le canal SSE job.updated).
 */
export function useVideoJob(id: string | null) {
  return useQuery({
    queryKey: videoAdKeys.job(id ?? ""),
    queryFn: () => fetchJob(id as string),
    enabled: !!id,
    select: (job) => job as VideoAdJob,
    refetchInterval: (query) => {
      const status = query.state.data?.status
      return status === "pending" || status === "processing" ? 5000 : false
    },
  })
}
