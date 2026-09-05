import { api, type ApiSingle } from "@/lib/api/client"

export type VideoRatio = "9:16" | "1:1" | "16:9"

export interface VideoAdScene {
  id: string
  type: "avatar" | "product" | "text" | string
  duration: number
  script?: string
  text_overlay?: string
  visual_description?: string
}

export interface VideoStoryboard {
  angle?: string
  hook?: string
  cta?: string
  duration: number
  aspect_ratio: string
  scenes: VideoAdScene[]
}

export interface VideoAdResult {
  asset_id?: string
  thumb_asset_id?: string
  storyboard?: VideoStoryboard
  provider_name?: string
  provider_video_id?: string
  duration?: number
}

export interface VideoAdJob {
  id: string
  kind: "video_ad" | string
  status: "pending" | "processing" | "completed" | "failed"
  error?: string
  cost: number
  result?: VideoAdResult | null
  created_at: string
  completed_at?: string | null
}

export interface GenerateVideoAdInput {
  duration?: number
  aspect_ratio?: VideoRatio
  cta?: string
  instructions?: string
}

export function generateVideoAd(projectId: string, input?: GenerateVideoAdInput) {
  return api
    .post<ApiSingle<VideoAdJob>>(`/api/v1/projects/${projectId}/video-ads`, input)
    .then((r) => r.data)
}

export interface JobUpdatedEvent {
  id: string
  kind: string
  status: "pending" | "processing" | "completed" | "failed"
  project_id?: string | null
  stage?: string
  error?: string
}
