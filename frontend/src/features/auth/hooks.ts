import { useMutation } from "@tanstack/react-query"

import { login, logout, register } from "./api"
import { useAuth } from "./auth-provider"

export function useLogin() {
  const { invalidate } = useAuth()
  return useMutation({
    mutationFn: login,
    onSuccess: () => invalidate(),
  })
}

export function useRegister() {
  const { invalidate } = useAuth()
  return useMutation({
    mutationFn: register,
    onSuccess: () => invalidate(),
  })
}

export function useLogout() {
  const { invalidate } = useAuth()
  return useMutation({
    mutationFn: logout,
    onSettled: () => invalidate(),
  })
}
