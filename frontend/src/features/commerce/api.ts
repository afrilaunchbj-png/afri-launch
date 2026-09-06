import { api, type ApiSingle } from "@/lib/api/client"

export const COMMERCE_PROVIDER = "chariow"

export interface CommerceConnection {
  provider: string
  status: string
  store_id: string
  store_name: string
  store_url: string
  store_currency: string
  webhook_url: string
  webhook_configured?: boolean
  connected_at?: string | null
  last_sync_at?: string | null
}

export interface CommerceStatusResponse {
  connected: boolean
  connection?: CommerceConnection
}

export interface CommerceProduct {
  id: string
  slug: string
  name: string
  type: string
  is_free: boolean
  price_minor: number
  currency: string
}

export interface CommerceProductLink {
  id: string
  project_id: string
  external_product_id: string
  external_product_name: string
  price_minor: number
  currency: string
  is_public: boolean
  public_url: string
}

export interface CommerceSale {
  id: string
  external_sale_id: string
  status: string
  amount_minor: number
  currency: string
  buyer_email: string
  buyer_name: string
  product_link_id?: string
  created_at: string
}

export function fetchCommerceStatus() {
  return api.get<ApiSingle<CommerceStatusResponse>>("/api/v1/commerce/status").then((r) => r.data)
}

export function connectCommerce(payload: { api_key: string; webhook_secret?: string }) {
  return api.post<ApiSingle<CommerceConnection>>("/api/v1/commerce/chariow/connect", payload).then((r) => r.data)
}

export function updateCommerceWebhookSecret(webhookSecret: string) {
  return api.post("/api/v1/commerce/chariow/webhook-secret", { webhook_secret: webhookSecret })
}

export function disconnectCommerce() {
  return api.post("/api/v1/commerce/chariow/disconnect")
}

export function deleteCommerce() {
  return api.delete("/api/v1/commerce/chariow")
}

export function syncCommerceSales() {
  return api.post("/api/v1/commerce/sync")
}

export function fetchCommerceProducts() {
  return api.get<ApiSingle<CommerceProduct[]>>("/api/v1/commerce/products").then((r) => r.data)
}

export function fetchCommerceLinks() {
  return api.get<ApiSingle<CommerceProductLink[]>>("/api/v1/commerce/links").then((r) => r.data)
}

export function linkCommerceProduct(payload: { project_id: string; external_product_id: string }) {
  return api.post<ApiSingle<CommerceProductLink>>("/api/v1/commerce/links", payload).then((r) => r.data)
}

export function publishCommerceLink(id: string, public_: boolean) {
  return api
    .put<ApiSingle<CommerceProductLink>>(`/api/v1/commerce/links/${id}/public`, { public: public_ })
    .then((r) => r.data)
}

export function unlinkCommerceProduct(id: string) {
  return api.delete(`/api/v1/commerce/links/${id}`)
}

export function fetchCommerceSales() {
  return api.get<ApiSingle<CommerceSale[]>>("/api/v1/commerce/sales").then((r) => r.data)
}

export interface PublicProductInfo {
  name: string
  price_minor: number
  currency: string
  is_free: boolean
  store_name: string
}

export function fetchPublicProduct(token: string) {
  return api.get<ApiSingle<PublicProductInfo>>(`/api/v1/commerce/public/${token}`).then((r) => r.data)
}

export interface PublicCheckoutResult {
  step: string
  checkout_url: string
  sale_id: string
  sale_status: string
}

export function createPublicCheckout(
  token: string,
  payload: { email: string; first_name: string; last_name: string; phone: { number: string; country_code: string } },
) {
  return api.post<ApiSingle<PublicCheckoutResult>>(`/api/v1/commerce/public/${token}/checkout`, payload).then((r) => r.data)
}
