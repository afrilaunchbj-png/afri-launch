import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import {
  connectCommerce,
  disconnectCommerce,
  fetchCommerceLinks,
  fetchCommerceProducts,
  fetchCommerceSales,
  fetchCommerceStatus,
  linkCommerceProduct,
  publishCommerceLink,
  syncCommerceSales,
  unlinkCommerceProduct,
  updateCommerceWebhookSecret,
} from "./api"

export const commerceKeys = {
  all: ["commerce"] as const,
  status: () => [...commerceKeys.all, "status"] as const,
  products: () => [...commerceKeys.all, "products"] as const,
  links: () => [...commerceKeys.all, "links"] as const,
  sales: () => [...commerceKeys.all, "sales"] as const,
}

export function useCommerceStatus() {
  return useQuery({ queryKey: commerceKeys.status(), queryFn: fetchCommerceStatus })
}

export function useCommerceProducts() {
  return useQuery({ queryKey: commerceKeys.products(), queryFn: fetchCommerceProducts, enabled: false })
}

export function useCommerceLinks() {
  return useQuery({ queryKey: commerceKeys.links(), queryFn: fetchCommerceLinks })
}

export function useCommerceSales() {
  return useQuery({ queryKey: commerceKeys.sales(), queryFn: fetchCommerceSales })
}

export function useConnectCommerce() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: connectCommerce,
    onSettled: () => {
      qc.invalidateQueries({ queryKey: commerceKeys.status() })
      qc.invalidateQueries({ queryKey: commerceKeys.links() })
    },
  })
}

export function useUpdateCommerceWebhookSecret() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: updateCommerceWebhookSecret,
    onSettled: () => qc.invalidateQueries({ queryKey: commerceKeys.status() }),
  })
}

export function useDisconnectCommerce() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: disconnectCommerce,
    onSettled: () => {
      qc.invalidateQueries({ queryKey: commerceKeys.status() })
      qc.invalidateQueries({ queryKey: commerceKeys.links() })
      qc.invalidateQueries({ queryKey: commerceKeys.sales() })
    },
  })
}

export function useSyncCommerce() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: syncCommerceSales,
    onSettled: () => qc.invalidateQueries({ queryKey: commerceKeys.sales() }),
  })
}

export function useLinkCommerceProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: linkCommerceProduct,
    onSettled: () => qc.invalidateQueries({ queryKey: commerceKeys.links() }),
  })
}

export function usePublishCommerceLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, public: p }: { id: string; public: boolean }) => publishCommerceLink(id, p),
    onSettled: () => qc.invalidateQueries({ queryKey: commerceKeys.links() }),
  })
}

export function useUnlinkCommerceProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: unlinkCommerceProduct,
    onSettled: () => qc.invalidateQueries({ queryKey: commerceKeys.links() }),
  })
}
