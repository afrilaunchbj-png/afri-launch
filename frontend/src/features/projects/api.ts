import { api, type ApiSingle } from "@/lib/api/client"

import type { Job } from "@/features/generation/api"

export type ProjectStatus = "draft" | "idea_selected" | "generating" | "content_ready" | "completed" | "failed"

export interface ProjectPalette {
  primary?: string
  secondary?: string
  accent?: string
  background?: string
  text?: string
  source?: "ai" | "user"
}

export interface ProjectConfig {
  palette?: ProjectPalette | null
  style?: string
  ebook_min_pages?: number
  ebook_max_pages?: number
}

export interface Project {
  id: string
  title: string
  status: ProjectStatus
  credits_consumed: number
  opportunity_id?: string | null
  idea_id?: string | null
  config?: ProjectConfig
  created_at: string
}

export interface Asset {
  id: string
  kind: string
  filename: string
  content_type: string
  size_bytes: number
  created_at: string
}

export interface CreateProjectInput {
  opportunity_id?: string | null
  idea_id?: string | null
  title: string
}

export function createProject(input: CreateProjectInput) {
  return api.post<ApiSingle<Project>>("/api/v1/projects", input).then((r) => r.data)
}

export function fetchProjects() {
  return api.get<ApiSingle<Project[]>>("/api/v1/projects").then((r) => r.data)
}

export function fetchProject(id: string) {
  return api.get<ApiSingle<Project>>(`/api/v1/projects/${id}`).then((r) => r.data)
}

export function generateEbook(id: string) {
  return api.post<ApiSingle<Job>>(`/api/v1/projects/${id}/ebook`).then((r) => r.data)
}

export function generateCover(id: string, instructions?: string) {
  return api
    .post<ApiSingle<Job>>(`/api/v1/projects/${id}/cover`, instructions ? { instructions } : undefined)
    .then((r) => r.data)
}

export function updateProjectConfig(id: string, input: ProjectConfigInput) {
  return api.put<ApiSingle<Project>>(`/api/v1/projects/${id}/config`, input).then((r) => r.data)
}

export interface ProjectConfigInput {
  palette?: ProjectPalette
  clear_palette?: boolean
  style?: string
  ebook_min_pages?: number
  ebook_max_pages?: number
}

export function generatePosters(id: string) {
  return api.post<ApiSingle<Job>>(`/api/v1/projects/${id}/posters`).then((r) => r.data)
}

export function generateSalesPage(id: string) {
  return api.post<ApiSingle<Job>>(`/api/v1/projects/${id}/sales-page`).then((r) => r.data)
}

export function fetchAssets(projectId: string) {
  return api.get<ApiSingle<Asset[]>>(`/api/v1/projects/${projectId}/assets`).then((r) => r.data)
}

export function assetDownloadPath(id: string) {
  return `/api/v1/assets/${id}/download`
}
