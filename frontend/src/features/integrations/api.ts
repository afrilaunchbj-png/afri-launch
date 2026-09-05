import { api, type ApiSingle } from "@/lib/api/client"

export type AdProvider = "meta" | "google_ads" | "tiktok_ads"

export const PROVIDERS: { id: AdProvider; name: string }[] = [
  { id: "meta", name: "Meta Ads" },
  { id: "google_ads", name: "Google Ads" },
  { id: "tiktok_ads", name: "TikTok Ads" },
]

export interface AdConnection {
  id: string
  provider: AdProvider
  status: "pending" | "active" | "expired" | "revoked" | "error" | "disconnected"
  external_account_id?: string
  external_account_name?: string
  last_error?: string
  last_sync_at?: string | null
  created_at: string
}

export interface IntegrationsResponse {
  connections: AdConnection[]
  capabilities: Record<string, Record<string, boolean>>
}

export interface AdAccount {
  id: string
  name: string
  currency?: string
  status?: string
}

export interface AdCampaign {
  id: string
  connection_id: string
  external_campaign_id?: string
  name: string
  objective?: string
  status: "draft" | "active" | "paused" | "deleted"
  budget_minor: number
  currency: string
}

export interface CreateAdCampaignInput {
  provider: AdProvider
  name: string
  objective: string
  budget_minor: number
}

export function fetchIntegrations() {
  return api.get<ApiSingle<IntegrationsResponse>>("/api/v1/integrations").then((r) => r.data)
}

export function fetchConnectURL(provider: AdProvider) {
  return api
    .get<ApiSingle<{ authorization_url: string }>>(`/api/v1/integrations/${provider}/connect`)
    .then((r) => r.data.authorization_url)
}

export function fetchAdAccounts(provider: AdProvider) {
  return api.get<ApiSingle<AdAccount[]>>(`/api/v1/integrations/${provider}/accounts`).then((r) => r.data)
}

export function selectAdAccount(provider: AdProvider, externalAccountId: string) {
  return api
    .post<ApiSingle<AdConnection>>(`/api/v1/integrations/${provider}/accounts/select`, {
      external_account_id: externalAccountId,
    })
    .then((r) => r.data)
}

export function disconnectProvider(provider: AdProvider) {
  return api.delete<ApiSingle<{ ok: boolean }>>(`/api/v1/integrations/${provider}`).then((r) => r.data)
}

export function syncCampaigns(provider: AdProvider) {
  return api
    .post<ApiSingle<AdCampaign[]>>(`/api/v1/integrations/${provider}/campaigns/sync`)
    .then((r) => r.data)
}

export function fetchCampaigns() {
  return api.get<ApiSingle<AdCampaign[]>>("/api/v1/ad-campaigns").then((r) => r.data)
}

export function createCampaign(input: CreateAdCampaignInput) {
  return api.post<ApiSingle<AdCampaign>>("/api/v1/ad-campaigns", input).then((r) => r.data)
}

export function pauseCampaign(id: string) {
  return api.post<ApiSingle<AdCampaign>>(`/api/v1/ad-campaigns/${id}/pause`).then((r) => r.data)
}

export function resumeCampaign(id: string) {
  return api.post<ApiSingle<AdCampaign>>(`/api/v1/ad-campaigns/${id}/resume`).then((r) => r.data)
}
