import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import {
  createCampaign,
  disconnectProvider,
  fetchAdAccounts,
  fetchCampaigns,
  fetchConnectURL,
  fetchIntegrations,
  pauseCampaign,
  resumeCampaign,
  selectAdAccount,
  syncCampaigns,
  type AdProvider,
  type CreateAdCampaignInput,
} from "./api"

export const integrationKeys = {
  all: ["integrations"] as const,
  detail: () => [...integrationKeys.all, "detail"] as const,
  accounts: (p: string) => [...integrationKeys.all, "accounts", p] as const,
  campaigns: () => ["ad-campaigns"] as const,
}

/** useIntegrations : connexions + capacités par plateforme. */
export function useIntegrations() {
  return useQuery({ queryKey: integrationKeys.detail(), queryFn: fetchIntegrations })
}

/** useConnectProvider renvoie l'URL OAuth à ouvrir (navigation complète). */
export function useConnectProvider() {
  return useMutation({ mutationFn: (provider: AdProvider) => fetchConnectURL(provider) })
}

/** useAdAccounts : comptes accessibles via une connexion active. */
export function useAdAccounts(provider: AdProvider | null) {
  return useQuery({
    queryKey: integrationKeys.accounts(provider ?? ""),
    queryFn: () => fetchAdAccounts(provider as AdProvider),
    enabled: !!provider,
  })
}

/** useSelectAccount confirme le choix de compte (vérifié côté backend). */
export function useSelectAccount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ provider, accountId }: { provider: AdProvider; accountId: string }) =>
      selectAdAccount(provider, accountId),
    onSettled: () => queryClient.invalidateQueries({ queryKey: integrationKeys.all }),
  })
}

/** useDisconnectProvider révoque et déconnecte (historique conservé). */
export function useDisconnectProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (provider: AdProvider) => disconnectProvider(provider),
    onSettled: () => queryClient.invalidateQueries({ queryKey: integrationKeys.all }),
  })
}

/** useCampaigns : campagnes internes (mapping externe conservé). */
export function useCampaigns() {
  return useQuery({ queryKey: integrationKeys.campaigns(), queryFn: fetchCampaigns })
}

/** useSyncCampaigns synchronise les campagnes depuis la plateforme. */
export function useSyncCampaigns() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (provider: AdProvider) => syncCampaigns(provider),
    onSettled: () => queryClient.invalidateQueries({ queryKey: integrationKeys.campaigns() }),
  })
}

/** useCreateCampaign crée une campagne (toujours en pause au départ). */
export function useCreateCampaign() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateAdCampaignInput) => createCampaign(input),
    onSettled: () => queryClient.invalidateQueries({ queryKey: integrationKeys.campaigns() }),
  })
}

/** usePauseResumeCampaign change le statut d'une campagne. */
export function usePauseResumeCampaign() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, action }: { id: string; action: "pause" | "resume" }) =>
      action === "pause" ? pauseCampaign(id) : resumeCampaign(id),
    onSettled: () => queryClient.invalidateQueries({ queryKey: integrationKeys.campaigns() }),
  })
}
