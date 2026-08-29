import { api, type ApiSingle } from "@/lib/api/client"

import type { LoginInput, RegisterInput, User } from "./types"

export const authKeys = {
  all: ["auth"] as const,
  me: () => [...authKeys.all, "me"] as const,
}

export function fetchMe() {
  return api.get<ApiSingle<User>>("/api/v1/auth/me").then((r) => r.data)
}

export function login(input: LoginInput) {
  return api.post<ApiSingle<User>>("/api/v1/auth/login", input).then((r) => r.data)
}

export function register(input: RegisterInput) {
  return api.post<ApiSingle<User>>("/api/v1/auth/register", input).then((r) => r.data)
}

export async function logout() {
  await api.post<void>("/api/v1/auth/logout")
}

/** URL d'autorisation Google (le serveur redirige vers Google). */
export function googleAuthUrl() {
  return "/api/v1/auth/google"
}
