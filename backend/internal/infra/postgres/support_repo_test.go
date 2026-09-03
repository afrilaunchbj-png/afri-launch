package postgres_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authapp "afrilaunch/backend/internal/application/auth"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres"
)

// TestSuperadminPromotionAndTickets valide la promotion SUPERADMIN_EMAILS au
// login et le cycle de vie des tickets de support, contre une vraie base.
// Activé uniquement lorsque AFRILAUNCH_TEST_DB est défini.
func TestSuperadminPromotionAndTickets(t *testing.T) {
	url := os.Getenv("AFRILAUNCH_TEST_DB")
	if url == "" {
		t.Skip("AFRILAUNCH_TEST_DB non défini — test d'intégration ignoré")
	}

	ctx := context.Background()
	pool, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)
	users := postgres.NewUserRepository(store)
	credits := postgres.NewCreditRepository(store)
	support := postgres.NewSupportRepository(store)

	// Le service promeut les emails déclarés superadmin à chaque login.
	stamp := time.Now().UnixNano()
	bossEmail := fmt.Sprintf("boss-%d@example.com", stamp)
	plainEmail := fmt.Sprintf("simple-%d@example.com", stamp)
	svc := authapp.NewService(users, credits, []string{bossEmail})

	identity := port.AuthUser{ID: uuid.NewString(), Email: strings.ToUpper(bossEmail), Name: "La Boss"}
	user, err := svc.GetOrCreateUser(ctx, identity, 0)
	if err != nil {
		t.Fatalf("get or create: %v", err)
	}
	if user.Role != domain.RoleSuperadmin {
		t.Fatalf("role = %q, want superadmin", user.Role)
	}

	// Re-login : le rôle est conservé (upsert n'écrase pas).
	user, err = svc.GetOrCreateUser(ctx, identity, 0)
	if err != nil || user.Role != domain.RoleSuperadmin {
		t.Fatalf("re-login: role=%q err=%v", user.Role, err)
	}

	// Un utilisateur normal n'est pas promu.
	plain, err := svc.GetOrCreateUser(ctx, port.AuthUser{ID: uuid.NewString(), Email: plainEmail, Name: "Simple"}, 0)
	if err != nil || plain.Role != domain.RoleUser {
		t.Fatalf("plain user: role=%q err=%v", plain.Role, err)
	}

	// Tickets : création, listing utilisateur.
	t1, err := support.Create(ctx, domain.SupportTicket{UserID: plain.ID, Subject: "Probleme ebook", Message: "La generation echoue depuis ce matin."})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	if t1.Status != domain.TicketOpen {
		t.Fatalf("status = %q, want open", t1.Status)
	}
	if _, err := support.Create(ctx, domain.SupportTicket{UserID: user.ID, Subject: "Question credits", Message: "Comment fonctionnent les credits ?"}); err != nil {
		t.Fatalf("create ticket 2: %v", err)
	}

	mine, err := support.ListByUser(ctx, plain.ID)
	if err != nil || len(mine) != 1 {
		t.Fatalf("list mine: %d err=%v", len(mine), err)
	}

	// Listing admin (avec auteur) + résolution.
	all, total, err := support.ListAll(ctx, 20, 0)
	if err != nil || total < 2 || len(all) < 2 {
		t.Fatalf("list all: total=%d len=%d err=%v", total, len(all), err)
	}
	found := false
	for _, t2 := range all {
		if t2.ID == t1.ID {
			found = true
			if t2.UserEmail != plainEmail {
				t.Errorf("user_email = %q", t2.UserEmail)
			}
		}
	}
	if !found {
		t.Error("ticket introuvable dans le listing admin")
	}

	resolved, err := support.SetStatus(ctx, t1.ID, domain.TicketResolved)
	if err != nil || resolved.Status != domain.TicketResolved {
		t.Fatalf("resolve: %+v err=%v", resolved, err)
	}
	open, err := support.CountOpen(ctx)
	if err != nil {
		t.Fatalf("count open: %v", err)
	}
	if open < 1 { // le ticket du superadmin reste ouvert
		t.Errorf("open = %d", open)
	}

	// Stats admin cohérentes.
	adminRepo := postgres.NewAdminRepository(store)
	stats, err := adminRepo.Stats(ctx)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.Users < 2 || stats.OpenTickets < 1 {
		t.Errorf("stats incohérentes : %+v", stats)
	}
	usersPage, totalUsers, err := adminRepo.ListUsers(ctx, 10, 0)
	if err != nil || totalUsers < 2 || len(usersPage) < 2 {
		t.Errorf("list users: %d/%d err=%v", len(usersPage), totalUsers, err)
	}

	fmt.Println("stats:", stats)
	_ = time.Now()
}
