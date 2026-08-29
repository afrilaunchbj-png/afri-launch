export interface CreditSummary {
  balance: number
  reserved: number
  available: number
  added_month: number
  used_month: number
}

export interface GenerationCost {
  operation: string
  name: string
  credits: number
  is_active: boolean
}

export interface CreditTransaction {
  id: string
  type: "credit" | "debit"
  amount: number
  operation: string
  status: string
  reference?: string | null
  created_at: string
}

export interface TransactionFilters {
  type: "" | "credit" | "debit"
  operation: string
  page: number
  pageSize: number
}
