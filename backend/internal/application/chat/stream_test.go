package chat

import (
	"context"
	"strings"
	"testing"

	appai "afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

func testCtx() context.Context { return context.Background() }

func testConv() domain.Conversation {
	return domain.Conversation{ID: "conv-1", UserID: "user-1"}
}

// fakeLLM rejoue une séquence de deltas scriptée.
type fakeLLM struct {
	chunks []string
}

func (f *fakeLLM) Complete(ctx context.Context, req port.LLMRequest) (port.LLMResponse, error) {
	return port.LLMResponse{Content: strings.Join(f.chunks, "")}, nil
}

func (f *fakeLLM) StreamComplete(ctx context.Context, req port.LLMRequest, emit func(string) error) error {
	for _, c := range f.chunks {
		if err := emit(c); err != nil {
			return err
		}
	}
	return nil
}

func newTestService(chunks []string) *Service {
	aiSvc := appai.NewService(&fakeLLM{chunks: chunks}, nil, nil, nil, appai.NewModelRouter("m-research", "m-idea", "m-img"))
	return NewService(nil, nil, nil, nil, nil, aiSvc, nil)
}

func TestChatSystemPromptLanguage(t *testing.T) {
	fr := chatSystemPrompt(domain.LanguageFr)
	if !strings.Contains(fr, "French") {
		t.Error("la directive de langue française manque pour 'fr'")
	}
	en := chatSystemPrompt(domain.LanguageEn)
	if !strings.Contains(en, "English") {
		t.Error("la directive de langue anglaise manque pour 'en'")
	}
	// Langue inconnue : repli sur le français.
	if u := chatSystemPrompt("es"); !strings.Contains(u, "French") {
		t.Error("langue inconnue devrait retomber sur le français")
	}
}

func TestStreamAnswerPlainText(t *testing.T) {
	svc := newTestService([]string{"Salut ! Je peux t'", "aider à explorer le ma", "rché béninois."})

	res, err := svc.streamAnswer(testCtx(), testConv(), "msg-1", domain.LanguageFr, nil)
	if err != nil {
		t.Fatalf("streamAnswer: %v", err)
	}
	if res.search != nil || res.hasIdeas {
		t.Fatalf("aucun outil attendu : %+v", res)
	}
	want := "Salut ! Je peux t'aider à explorer le marché béninois."
	if res.visible != want {
		t.Errorf("visible = %q, want %q", res.visible, want)
	}
}

func TestStreamAnswerMarkerSplitAcrossDeltas(t *testing.T) {
	// Le "@" coupé entre deux deltas ne doit pas corrompre la détection.
	svc := newTestService([]string{"Voici mes idées @", "@IDEAS\n", `{"ideas":[{"title":"A","hook":"h","explanation":"e"}]}`, "\n@@END"})

	res, err := svc.streamAnswer(testCtx(), testConv(), "msg-1", domain.LanguageFr, nil)
	if err != nil {
		t.Fatalf("streamAnswer: %v", err)
	}
	if res.search != nil {
		t.Fatalf("pas de recherche attendue")
	}
	if !res.hasIdeas || len(res.ideas.Ideas) != 1 {
		t.Fatalf("idées non extraites : %+v", res)
	}
	if res.ideas.Ideas[0].Title != "A" {
		t.Errorf("title = %q", res.ideas.Ideas[0].Title)
	}
	if strings.Contains(res.visible, "@@") {
		t.Errorf("le marqueur ne doit pas fuiter dans le texte visible : %q", res.visible)
	}
	if res.visible != "Voici mes idées " {
		t.Errorf("visible = %q", res.visible)
	}
}

func TestStreamAnswerSearchFirstLine(t *testing.T) {
	svc := newTestService([]string{"@@SEA", `RCH {"query":"mobil-money","sector":"fintech","markets":["Bénin"]}`, "\n"})

	res, err := svc.streamAnswer(testCtx(), testConv(), "msg-1", domain.LanguageFr, nil)
	if err != nil {
		t.Fatalf("streamAnswer: %v", err)
	}
	if res.search == nil {
		t.Fatal("recherche non détectée")
	}
	if res.search.Query != "mobil-money" || res.search.Sector != "fintech" {
		t.Errorf("search = %+v", res.search)
	}
	if res.visible != "" {
		t.Errorf("la ligne de recherche ne doit pas être émise : %q", res.visible)
	}
}
