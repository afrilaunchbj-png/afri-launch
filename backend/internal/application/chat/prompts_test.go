package chat

import (
	"strings"
	"testing"

	"afrilaunch/backend/internal/domain"
)

func TestParseSearchLine(t *testing.T) {
	args, err := parseSearchLine(`@@SEARCH {"query":"formation en ligne","sector":"edtech","markets":["Bénin","Sénégal"]}`)
	if err != nil {
		t.Fatalf("parseSearchLine: %v", err)
	}
	if args.Query != "formation en ligne" {
		t.Errorf("query = %q", args.Query)
	}
	if args.Sector != "edtech" || len(args.Markets) != 2 {
		t.Errorf("args = %+v", args)
	}
}

func TestParseSearchLineInvalid(t *testing.T) {
	if _, err := parseSearchLine(`@@SEARCH pas du json`); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if _, err := parseSearchLine(`@@SEARCH {"query":""}`); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestParseIdeasBlock(t *testing.T) {
	block := `
{"ideas":[{"title":"Compta Express","hook":"Gérez vos comptes en 10 min/jour","explanation":"App mobile pour micro-commerçants."}]}
@@END
`
	ideas, ok := parseIdeasBlock(block)
	if !ok {
		t.Fatal("bloc idées non reconnu")
	}
	if len(ideas.Ideas) != 1 || ideas.Ideas[0].Title != "Compta Express" {
		t.Fatalf("ideas = %+v", ideas)
	}
}

func TestParseIdeasBlockEmpty(t *testing.T) {
	if _, ok := parseIdeasBlock(""); ok {
		t.Fatal("bloc vide ne devrait pas être reconnu")
	}
	if _, ok := parseIdeasBlock("pas du json"); ok {
		t.Fatal("JSON invalide ne devrait pas être reconnu")
	}
}

func TestBuildLLMMessagesTrimsHistory(t *testing.T) {
	history := make([]domain.ConversationMessage, 0, maxTurns+5)
	for i := 0; i < maxTurns+5; i++ {
		role := domain.ConversationMessageUser
		if i%2 == 1 {
			role = domain.ConversationMessageAssistant
		}
		history = append(history, domain.ConversationMessage{Role: role, Content: strings.Repeat("m", i+1)})
	}

	msgs := buildLLMMessages(history)
	if len(msgs) != maxTurns {
		t.Fatalf("len(msgs) = %d, want %d", len(msgs), maxTurns)
	}
	if msgs[0].Content != strings.Repeat("m", 6) {
		t.Errorf("le troncage doit garder les derniers messages")
	}
}
