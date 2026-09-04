package domain

// SeriesPoint est un point de série temporelle (courbe indicative).
type SeriesPoint struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

// DashboardStats regroupe les indicateurs personnels du tableau de bord.
type DashboardStats struct {
	Projects        int64            `json:"projects"`
	Conversations   int64            `json:"conversations"`
	OpenTickets     int64            `json:"open_tickets"`
	CreditsBalance  int64            `json:"credits_balance"`
	CreditsUsed30d  int64            `json:"credits_used_30d"`
	JobsByStatus    map[string]int64 `json:"jobs_by_status"`
	CreditsPerDay   []SeriesPoint    `json:"credits_per_day"`
	ProjectsPerWeek []SeriesPoint    `json:"projects_per_week"`
}
