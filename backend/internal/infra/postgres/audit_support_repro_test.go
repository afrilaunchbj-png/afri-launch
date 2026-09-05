package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	"afrilaunch/backend/internal/infra/postgres"
)

// TestAuditLogsList reproduit le 500 de GET /admin/audit-logs (filtre vide).
func TestAuditLogsList(t *testing.T) {
	url := os.Getenv("AFRILAUNCH_TEST_DB")
	if url == "" {
		t.Skip("AFRILAUNCH_TEST_DB non défini")
	}
	ctx := context.Background()
	pool, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)
	adminRepo := postgres.NewAdminRepository(store)

	// Écrit une entrée puis la relit avec filtre vide (cas de la page admin).
	users := postgres.NewUserRepository(store)
	user, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: "audit-" + uuid.NewString() + "@test.local", FullName: "Audit Test"})
	if err != nil {
		t.Fatal(err)
	}
	auditRepo := postgres.NewAuditRepository(store)
	if err := auditRepo.Log(ctx, domain.AuditLog{
		UserID: user.ID, Action: domain.AuditProjectCreated, Entity: "project", EntityID: uuid.NewString(),
		Metadata: map[string]any{"title": "x"},
	}); err != nil {
		t.Fatal(err)
	}

	logs, total, err := adminRepo.ListAuditLogs(ctx, port.AuditFilter{}, 20, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs filtre vide: %v", err) // ← 500 constaté
	}
	if total < 1 || len(logs) < 1 {
		t.Fatalf("journal vide: total=%d len=%d", total, len(logs))
	}

	if _, _, err := adminRepo.ListAuditLogs(ctx, port.AuditFilter{UserID: user.ID, Action: "", Entity: ""}, 20, 0); err != nil {
		t.Fatalf("ListAuditLogs filtre userId: %v", err)
	}
}

// TestTicketDetailAttachments reproduit le 500 de GET /admin/tickets/{id}.
func TestTicketDetailAttachments(t *testing.T) {
	url := os.Getenv("AFRILAUNCH_TEST_DB")
	if url == "" {
		t.Skip("AFRILAUNCH_TEST_DB non défini")
	}
	ctx := context.Background()
	pool, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	store := postgres.NewStore(pool)
	users := postgres.NewUserRepository(store)
	supportRepo := postgres.NewSupportRepository(store)

	user, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: "sup-" + uuid.NewString() + "@test.local", FullName: "Support Test"})
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := supportRepo.Create(ctx, domain.SupportTicket{UserID: user.ID, Subject: "Bug", Message: "Test"})
	if err != nil {
		t.Fatal(err)
	}

	// Attachment uploadé puis rattaché au ticket.
	att, err := supportRepo.CreateAttachment(ctx, domain.SupportAttachment{
		UserID: user.ID, Filename: "capture.png", StorageKey: "support/test/" + uuid.NewString() + ".png",
		ContentType: "image/png", SizeBytes: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := supportRepo.BindAttachments(ctx, user.ID, []string{att.ID}, ticket.ID, ""); err != nil {
		t.Fatalf("bind: %v", err)
	}

	files, err := supportRepo.ListAttachmentsByTicket(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("ListAttachmentsByTicket: %v", err) // ← 500 constaté
	}
	if len(files) != 1 || files[0].TicketID != ticket.ID {
		t.Fatalf("files = %+v", files)
	}

	// Message + rattachement au message (pièce jointe dédiée : bind à usage unique).
	msg, err := supportRepo.AddMessage(ctx, ticket.ID, domain.TicketMessage{AuthorID: user.ID, Content: "voir fichier"})
	if err != nil {
		t.Fatal(err)
	}
	att2, err := supportRepo.CreateAttachment(ctx, domain.SupportAttachment{
		UserID: user.ID, Filename: "log.pdf", StorageKey: "support/test/" + uuid.NewString() + ".pdf",
		ContentType: "application/pdf", SizeBytes: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := supportRepo.BindAttachments(ctx, user.ID, []string{att2.ID}, "", msg.ID); err != nil {
		t.Fatalf("bind message: %v", err)
	}
	byMsg, err := supportRepo.ListAttachmentsByMessages(ctx, []string{msg.ID})
	if err != nil {
		t.Fatalf("ListAttachmentsByMessages: %v", err)
	}
	if len(byMsg) != 1 || byMsg[0].MessageID != msg.ID {
		t.Fatalf("byMsg = %+v", byMsg)
	}
}
