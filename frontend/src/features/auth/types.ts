export interface User {
  id: string
  email: string
  full_name: string
  avatar_url?: string | null
}

export interface LoginInput {
  email: string
  password: string
}

export interface RegisterInput {
  email: string
  password: string
  full_name: string
}
