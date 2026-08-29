export type Difficulty = "low" | "medium" | "high"
export type Signal = "verified" | "estimated" | "inferred" | "hypothesis"

export interface OpportunityScores {
  demand: number
  pain: number
  competition: number
  purchasing_power: number
  digital_fit: number
  evidence_strength: number
}

export interface Opportunity {
  id: string
  title: string
  summary: string
  country: string
  sector: string
  language: string
  difficulty: Difficulty
  signal: Signal
  score: number
  scores: OpportunityScores
  evidence: unknown[]
  is_saved: boolean
  created_at: string
}

export interface OpportunityFilters {
  country: string
  sector: string
  difficulty: "" | Difficulty
  q: string
  page: number
  pageSize: number
}

export interface OpportunityFacets {
  countries: string[]
  sectors: string[]
  difficulties: Difficulty[]
}
