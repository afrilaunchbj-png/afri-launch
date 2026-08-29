import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api/client"

import {
  assetDownloadPath,
  createProject,
  fetchAssets,
  fetchProject,
  fetchProjects,
  generateCover,
  generateEbook,
  generatePosters,
  generateSalesPage,
} from "./api"

export const projectKeys = {
  all: ["projects"] as const,
  list: () => [...projectKeys.all, "list"] as const,
  detail: (id: string) => [...projectKeys.all, "detail", id] as const,
  assets: (id: string) => [...projectKeys.all, "assets", id] as const,
}

export function useProjects() {
  return useQuery({ queryKey: projectKeys.list(), queryFn: fetchProjects })
}

export function useProject(id: string) {
  return useQuery({ queryKey: projectKeys.detail(id), queryFn: () => fetchProject(id) })
}

export function useAssets(projectId: string) {
  return useQuery({
    queryKey: projectKeys.assets(projectId),
    queryFn: () => fetchAssets(projectId),
    refetchInterval: 3000,
  })
}

export function useCreateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createProject,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: projectKeys.list() }),
  })
}

export function useGenerate(kind: "ebook" | "cover" | "posters" | "sales-page") {
  const queryClient = useQueryClient()
  const fn = { ebook: generateEbook, cover: generateCover, posters: generatePosters, "sales-page": generateSalesPage }[kind]
  return useMutation({
    mutationFn: (id: string) => fn(id),
    onSettled: (_data, _err, id) => {
      queryClient.invalidateQueries({ queryKey: projectKeys.detail(id) })
      queryClient.invalidateQueries({ queryKey: projectKeys.assets(id) })
    },
  })
}

export async function downloadAsset(id: string, filename: string) {
  const blob = await api.download(assetDownloadPath(id))
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
