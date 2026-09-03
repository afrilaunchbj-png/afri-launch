package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres"
)

// TestUserPreferencesLifecycle valide GetOrCreate (défauts au premier appel)
// et Upsert, contre une vraie base PostgreSQL.
// Activé uniquement lorsque AFRILAUNCH_TEST_DB est défini.
func TestUserPreferencesLifecycle(t *testing.T) {
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
	prefs := postgres.NewPreferenceRepository(store)

	user, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: fmt.Sprintf("prefs-test-%d@example.com", time.Now().UnixNano()), FullName: "Prefs Test"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Premier accès : valeurs par défaut (fr / system).
	got, err := prefs.GetOrCreate(ctx, user.ID)
	if err != nil {
		t.Fatalf("get or create: %v", err)
	}
	if got.Language != domain.LanguageFr || got.Theme != domain.ThemeSystem {
		t.Fatalf("défauts inattendus : %+v", got)
	}

	// GetOrCreate idempotent.
	again, err := prefs.GetOrCreate(ctx, user.ID)
	if err != nil || again.Language != domain.LanguageFr {
		t.Fatalf("second get: %+v err=%v", again, err)
	}

	// Mise à jour.
	updated, err := prefs.Upsert(ctx, domain.UserPreference{UserID: user.ID, Language: domain.LanguageEn, Theme: domain.ThemeDark})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if updated.Language != domain.LanguageEn || updated.Theme != domain.ThemeDark {
		t.Fatalf("update non persisté : %+v", updated)
	}

	// Relecture.
	reread, err := prefs.GetOrCreate(ctx, user.ID)
	if err != nil || reread.Language != domain.LanguageEn || reread.Theme != domain.ThemeDark {
		t.Fatalf("relecture : %+v err=%v", reread, err)
	}
}
