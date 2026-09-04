package postgres

import (
	"context"
	"time"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres/db"
)

// dashboardRepo implémente port.DashboardRepository.
type dashboardRepo struct {
	s *Store
}

// NewDashboardRepository construit le repository du tableau de bord.
func NewDashboardRepository(s *Store) *dashboardRepo { return &dashboardRepo{s: s} }

func (r *dashboardRepo) Stats(ctx context.Context, userID string) (domain.DashboardStats, error) {
	stats := domain.DashboardStats{
		JobsByStatus:    make(map[string]int64),
		CreditsPerDay:   []domain.SeriesPoint{},
		ProjectsPerWeek: []domain.SeriesPoint{},
	}

	var err error
	if stats.Projects, err = r.s.q.CountUserProjects(ctx, userID); err != nil {
		return stats, err
	}
	if stats.Conversations, err = r.s.q.CountUserConversations(ctx, userID); err != nil {
		return stats, err
	}
	if stats.OpenTickets, err = r.s.q.CountUserOpenTickets(ctx, userID); err != nil {
		return stats, err
	}
	jobs, err := r.s.q.CountUserJobsByStatus(ctx, userID)
	if err != nil {
		return stats, err
	}
	for _, j := range jobs {
		stats.JobsByStatus[j.Status] = j.Total
	}
	if account, err := r.s.q.GetUserCreditAccount(ctx, userID); err == nil {
		stats.CreditsBalance = account.Balance
	} else if !isNoRows(err) {
		return stats, err
	}

	since := time.Now().AddDate(0, 0, -30)
	if stats.CreditsUsed30d, err = r.s.q.SumUserCreditsUsedSince(ctx, db.SumUserCreditsUsedSinceParams{UserID: userID, CreatedAt: since}); err != nil {
		return stats, err
	}
	days, err := r.s.q.UserCreditsPerDay(ctx, db.UserCreditsPerDayParams{UserID: userID, CreatedAt: since})
	if err != nil {
		return stats, err
	}
	for _, d := range days {
		stats.CreditsPerDay = append(stats.CreditsPerDay, domain.SeriesPoint{
			Date:  d.Day.UTC().Format("2006-01-02"),
			Value: d.Total,
		})
	}

	weeks, err := r.s.q.UserProjectsPerWeek(ctx, db.UserProjectsPerWeekParams{UserID: userID, CreatedAt: time.Now().AddDate(0, 0, -84)})
	if err != nil {
		return stats, err
	}
	for _, w := range weeks {
		stats.ProjectsPerWeek = append(stats.ProjectsPerWeek, domain.SeriesPoint{
			Date:  w.Week.UTC().Format("2006-01-02"),
			Value: w.Total,
		})
	}
	return stats, nil
}
