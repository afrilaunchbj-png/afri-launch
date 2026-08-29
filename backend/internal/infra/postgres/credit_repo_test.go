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

// TestCreditLedgerLifecycle valide le cycle reserve → consume|release et son
// idempotence, contre une vraie base PostgreSQL.
// Activé uniquement lorsque AFRILAUNCH_TEST_DB est défini.
func TestCreditLedgerLifecycle(t *testing.T) {
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

	email := fmt.Sprintf("ledger-test-%d@example.com", time.Now().UnixNano())
	user, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: email, FullName: "Ledger Test"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Compte initial + bonus de bienvenue.
	if _, err := credits.GetOrCreateAccount(ctx, user.ID, 0); err != nil {
		t.Fatalf("get or create account: %v", err)
	}
	if _, err := credits.Grant(ctx, user.ID, 100, domain.OperationWelcomeBonus, "welcome:"+user.ID); err != nil {
		t.Fatalf("grant: %v", err)
	}

	acc, err := credits.GetAccount(ctx, user.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if acc.Available() != 100 {
		t.Fatalf("expected available 100, got %d", acc.Available())
	}

	// Reserve (idempotent).
	ref := "gen:" + user.ID
	res1, err := credits.Reserve(ctx, user.ID, 30, domain.OperationEbookGen, ref, time.Hour)
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	acc, _ = credits.GetAccount(ctx, user.ID)
	if acc.Available() != 70 || acc.Reserved != 30 {
		t.Fatalf("after reserve: available=%d reserved=%d", acc.Available(), acc.Reserved)
	}

	// Re-réservation avec la même référence : idempotent, pas de double blocage.
	res2, err := credits.Reserve(ctx, user.ID, 30, domain.OperationEbookGen, ref, time.Hour)
	if err != nil {
		t.Fatalf("reserve (idempotent): %v", err)
	}
	if res1.ID != res2.ID {
		t.Fatalf("expected same reservation, got %s vs %s", res1.ID, res2.ID)
	}
	acc, _ = credits.GetAccount(ctx, user.ID)
	if acc.Reserved != 30 {
		t.Fatalf("idempotent reserve changed reserved to %d", acc.Reserved)
	}

	// Réservation au-delà du disponible.
	if _, err := credits.Reserve(ctx, user.ID, 999, domain.OperationVideoGen, "gen-big:"+user.ID, time.Hour); err != domain.ErrInsufficient {
		t.Fatalf("expected ErrInsufficient, got %v", err)
	}

	// Consume (idempotent).
	tx1, err := credits.Consume(ctx, user.ID, ref)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if tx1.Amount != 30 || tx1.Type != domain.TransactionDebit {
		t.Fatalf("unexpected consume tx: %+v", tx1)
	}
	acc, _ = credits.GetAccount(ctx, user.ID)
	if acc.Available() != 70 || acc.Reserved != 0 {
		t.Fatalf("after consume: available=%d reserved=%d", acc.Available(), acc.Reserved)
	}

	tx2, err := credits.Consume(ctx, user.ID, ref)
	if err != nil {
		t.Fatalf("consume (idempotent): %v", err)
	}
	if tx1.ID != tx2.ID {
		t.Fatalf("expected same transaction, got %s vs %s", tx1.ID, tx2.ID)
	}
	acc, _ = credits.GetAccount(ctx, user.ID)
	if acc.Available() != 70 {
		t.Fatalf("idempotent consume changed balance to %d", acc.Available())
	}

	// Release.
	ref2 := "gen-release:" + user.ID
	if _, err := credits.Reserve(ctx, user.ID, 20, domain.OperationIdeaGeneration, ref2, time.Hour); err != nil {
		t.Fatalf("reserve 2: %v", err)
	}
	if err := credits.Release(ctx, user.ID, ref2); err != nil {
		t.Fatalf("release: %v", err)
	}
	acc, _ = credits.GetAccount(ctx, user.ID)
	if acc.Available() != 70 || acc.Reserved != 0 {
		t.Fatalf("after release: available=%d reserved=%d", acc.Available(), acc.Reserved)
	}

	// Release idempotent.
	if err := credits.Release(ctx, user.ID, ref2); err != nil {
		t.Fatalf("release (idempotent): %v", err)
	}
}
