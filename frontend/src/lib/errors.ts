export type ErrorKind =
  | "validation"
  | "network"
  | "unauthorized"
  | "forbidden"
  | "not_found"
  | "conflict"
  | "business"
  | "internal"

export interface FieldError {
  field: string
  message: string
}

/**
 * AppError est l'erreur applicative frontend, dérivée du format
 * RFC 9457 (Problem Details) renvoyé par l'API Go.
 */
export class AppError extends Error {
  readonly kind: ErrorKind
  readonly status: number
  readonly fields: FieldError[]

  constructor(kind: ErrorKind, status: number, message: string, fields: FieldError[] = []) {
    super(message)
    this.name = "AppError"
    this.kind = kind
    this.status = status
    this.fields = fields
  }
}

export function isAppError(error: unknown): error is AppError {
  return error instanceof AppError
}
