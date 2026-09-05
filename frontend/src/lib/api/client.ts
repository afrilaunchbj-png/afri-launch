import { getAccessToken } from "@/lib/auth"
import { AppError, type ErrorKind, type FieldError } from "@/lib/errors"

const API_URL = import.meta.env.VITE_API_URL ?? ""

interface ProblemDetails {
  type?: string
  title?: string
  status?: number
  detail?: string
  errors?: FieldError[]
}

function kindFromStatus(status: number, type?: string): ErrorKind {
  if (type?.includes("validation")) return "validation"
  if (status === 401) return "unauthorized"
  if (status === 403) return "forbidden"
  if (status === 404) return "not_found"
  if (status === 409) return "conflict"
  if (status === 422) return "business"
  if (status === 429) return "business"
  if (status >= 500) return "internal"
  return "internal"
}

async function toAppError(response: Response): Promise<AppError> {
  let problem: ProblemDetails = {}
  try {
    problem = (await response.json()) as ProblemDetails
  } catch {
    // réponse non-JSON : on utilise le statut seul
  }

  const status = problem.status ?? response.status
  const kind = kindFromStatus(status, problem.type)
  const message = problem.detail || problem.title || "Erreur inconnue"

  return new AppError(kind, status, message, problem.errors ?? [])
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response
  try {
    const token = await getAccessToken()
    response = await fetch(`${API_URL}${path}`, {
      ...init,
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...init?.headers,
      },
    })
  } catch {
    throw new AppError("network", 0, "Impossible de contacter le serveur.")
  }

  if (!response.ok) {
    throw await toAppError(response)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

async function requestBlob(path: string): Promise<Blob> {
  let response: Response
  try {
    const token = await getAccessToken()
    response = await fetch(`${API_URL}${path}`, {
      credentials: "include",
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    })
  } catch {
    throw new AppError("network", 0, "Impossible de contacter le serveur.")
  }

  if (!response.ok) {
    throw await toAppError(response)
  }
  return response.blob()
}

/** upload POSTe un FormData (multipart) — laisse le navigateur fixer le Content-Type. */
async function requestUpload<T>(path: string, formData: FormData): Promise<T> {
  let response: Response
  try {
    const token = await getAccessToken()
    response = await fetch(`${API_URL}${path}`, {
      method: "POST",
      body: formData,
      credentials: "include",
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    })
  } catch {
    throw new AppError("network", 0, "Impossible de contacter le serveur.")
  }
  if (!response.ok) {
    throw await toAppError(response)
  }
  return (await response.json()) as T
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "POST", body: body === undefined ? undefined : JSON.stringify(body) }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "PUT", body: body === undefined ? undefined : JSON.stringify(body) }),
  patch: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "PATCH", body: body === undefined ? undefined : JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
  upload: <T>(path: string, formData: FormData) => requestUpload<T>(path, formData),
  download: (path: string) => requestBlob(path),
}

/** Enveloppe de réponse standard de l'API : { data, pagination? }. */
export interface ApiList<T> {
  data: T[]
  pagination: {
    page: number
    pageSize: number
    totalItems: number
    totalPages: number
  }
}

export interface ApiSingle<T> {
  data: T
}
