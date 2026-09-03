package chat_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	appai "afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/application/chat"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
	eventsinfra "afrilaunch/backend/internal/infra/events"
	"afrilaunch/backend/internal/infra/postgres"
)

// fakeChatLLM rejoue une réponse scriptée (streaming delta par delta).
type fakeChatLLM struct {
	chunks []string
}

func (f *fakeChatLLM) Complete(ctx context.Context, req port.LLMRequest) (port.LLMResponse, error) {
	content := ""
	for _, c := range f.chunks {
		content += c
	}
	return port.LLMResponse{Content: content}, nil
}

func (f *fakeChatLLM) StreamComplete(ctx context.Context, req port.LLMRequest, emit func(string) error) error {
	for _, c := range f.chunks {
		if err := emit(c); err != nil {
			return err
		}
	}
	return nil
}

// scriptedLLM rejoue plusieurs appels LLM (un par tour de boucle agent).
type scriptedLLM struct {
	calls [][]string // chunks par appel
	seq   int
}

func (f *scriptedLLM) Complete(ctx context.Context, req port.LLMRequest) (port.LLMResponse, error) {
	content := ""
	for _, c := range f.calls[min(f.seq, len(f.calls)-1)] {
		content += c
	}
	f.seq++
	return port.LLMResponse{Content: content}, nil
}

func (f *scriptedLLM) StreamComplete(ctx context.Context, req port.LLMRequest, emit func(string) error) error {
	chunks := f.calls[min(f.seq, len(f.calls)-1)]
	f.seq++
	for _, c := range chunks {
		if err := emit(c); err != nil {
			return err
		}
	}
	return nil
}

// fakeResearchProvider renvoie un résultat de recherche en ligne fixe.
type fakeResearchProvider struct{}

func (f *fakeResearchProvider) Research(ctx context.Context, req port.ResearchRequest) (port.ResearchResult, error) {
	content := `{"opportunities":[{"country":"Bénin","title":"Formation couture en ligne","summary":"Marché de la formation professionnelle au Bénin.","difficulty":"low","signal":"estimated","score":72,"scores":{"demand":70,"pain":65,"competition":55,"purchasing_power":50,"digital_fit":80,"evidence_strength":60},"evidence":[{"source":"test","title":"Rapport test","url":"https://example.test/rapport","metric":"pénétration mobile","value":"45%"}]}]}`
	return port.ResearchResult{Content: content, Sources: []port.Source{{Title: "Rapport test", URL: "https://example.test/rapport"}}}, nil
}

// TestChatTurnSearchFlow valide la boucle agent @@SEARCH : le copilote
// déclenche une recherche en ligne facturée, les opportunités sont créées
// et rattachées à la conversation, puis les idées sont proposées.
func TestChatTurnSearchFlow(t *testing.T) {
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
	ideas := postgres.NewIdeaRepository(store)
	opps := postgres.NewOpportunityRepository(store)
	conversations := postgres.NewConversationRepository(store)

	user, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: fmt.Sprintf("chat-search-%d@example.com", time.Now().UnixNano()), FullName: "Chat Search"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := credits.Grant(ctx, user.ID, 100, domain.OperationWelcomeBonus, "welcome:"+user.ID); err != nil {
		t.Fatalf("grant: %v", err)
	}

	ideasJSON := `{"ideas":[{"title":"Couture Pro","hook":"Apprenez la couture en ligne","explanation":"Formation vidéo pour débutantes."}]}`
	llm := &scriptedLLM{calls: [][]string{
		{`@@SEARCH {"query":"formation couture","sector":"education","markets":["Bénin"]}`, "\n"},
		{"D'après mes recherches, voici une idée ", "@@IDEAS\n", ideasJSON, "\n@@END"},
	}}
	aiSvc := appai.NewService(llm, nil, nil, &fakeResearchProvider{}, appai.NewModelRouter("m-research", "m-idea", "m-img"))

	broker := eventsinfra.NewBroker()
	chatSvc := chat.NewService(conversations, ideas, opps, credits, aiSvc, broker)

	conv, err := chatSvc.Create(ctx, user.ID)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	evtCh, cancel := broker.Subscribe(user.ID)
	defer cancel()

	if _, err := chatSvc.SendMessage(ctx, user.ID, conv.ID, "Trouve-moi une opportunité couture au Bénin"); err != nil {
		t.Fatalf("send message: %v", err)
	}

	// Collecter les événements jusqu'à chat.completed.
	var tools, completed []port.AppEvent
	deadline := time.After(10 * time.Second)
	for completed == nil {
		select {
		case evt, ok := <-evtCh:
			if !ok {
				t.Fatal("canal events fermé inattendu")
			}
			switch evt.Type {
			case port.EventChatTool:
				tools = append(tools, evt)
			case port.EventChatCompleted:
				completed = append(completed, evt)
			}
		case <-deadline:
			t.Fatal("timeout : chat.completed jamais reçu")
		}
	}

	// 2 événements tool : running puis completed.
	if len(tools) != 2 {
		t.Fatalf("expected 2 chat.tool events, got %d", len(tools))
	}
	var toolPayload struct {
		Tool           string   `json:"tool"`
		Status         string   `json:"status"`
		OpportunityIDs []string `json:"opportunity_ids"`
	}
	if err := json.Unmarshal(tools[1].Data, &toolPayload); err != nil {
		t.Fatalf("decode tool payload: %v", err)
	}
	if toolPayload.Status != "completed" || len(toolPayload.OpportunityIDs) != 1 {
		t.Fatalf("tool payload inattendu : %+v", toolPayload)
	}

	// L'opportunité est créée et rattachée à la conversation.
	convAfter, err := conversations.Get(ctx, user.ID, conv.ID)
	if err != nil {
		t.Fatalf("get conversation: %v", err)
	}
	if convAfter.OpportunityID == nil || *convAfter.OpportunityID != toolPayload.OpportunityIDs[0] {
		t.Fatalf("conversation opportunity = %v, want %q", convAfter.OpportunityID, toolPayload.OpportunityIDs[0])
	}
	opp, err := opps.Get(ctx, toolPayload.OpportunityIDs[0])
	if err != nil {
		t.Fatalf("get opportunity: %v", err)
	}
	if opp.Title != "Formation couture en ligne" || opp.UserID == nil || *opp.UserID != user.ID {
		t.Errorf("opportunity inattendue : %+v (user=%v)", opp.Title, opp.UserID)
	}

	// Les idées sont liées à l'opportunité trouvée.
	dbIdeas, err := ideas.ListByConversation(ctx, user.ID, conv.ID)
	if err != nil {
		t.Fatalf("list ideas: %v", err)
	}
	if len(dbIdeas) != 1 || dbIdeas[0].OpportunityID == nil || *dbIdeas[0].OpportunityID != opp.ID {
		t.Fatalf("idea non liée à l'opportunité : %+v", dbIdeas)
	}

	// Crédits : 5 (recherche) + 2 (idées) = 7 consommés.
	acc, err := credits.GetAccount(ctx, user.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if acc.Available() != 93 {
		t.Fatalf("expected available 93, got %d", acc.Available())
	}
}

// TestChatTurnIdeaFlow valide un tour complet du copilote contre une vraie
// base : messages persistés, idées créées liées à la conversation, crédits
// consommés, événement chat.completed diffusé sur le canal temps réel.
// Activé uniquement lorsque AFRILAUNCH_TEST_DB est défini.
func TestChatTurnIdeaFlow(t *testing.T) {
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
	ideas := postgres.NewIdeaRepository(store)
	opps := postgres.NewOpportunityRepository(store)
	conversations := postgres.NewConversationRepository(store)

	user, err := users.Upsert(ctx, domain.User{ID: uuid.NewString(), Email: fmt.Sprintf("chat-test-%d@example.com", time.Now().UnixNano()), FullName: "Chat Test"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := credits.Grant(ctx, user.ID, 100, domain.OperationWelcomeBonus, "welcome:"+user.ID); err != nil {
		t.Fatalf("grant: %v", err)
	}

	ideasJSON := `{"ideas":[` +
		`{"title":"Compta Express","hook":"Gérez vos comptes en 10 min/jour","explanation":"App mobile pour micro-commerçants.","status":"draft"},` +
		`{"title":"Stock Simple","hook":"Votre stock dans la poche","explanation":"Inventaire temps réel pour boutiques."},` +
		`{"title":"Prix Juste","hook":"Fixez le bon prix, enfin","explanation":"Comparateur de prix du marché."}` +
		`]}`
	llm := &fakeChatLLM{chunks: []string{
		"Voici 3 pistes pour vous lancer ",
		"@@IDEAS\n", ideasJSON, "\n@@END",
	}}
	aiSvc := appai.NewService(llm, nil, nil, nil, appai.NewModelRouter("m-research", "m-idea", "m-img"))

	broker := eventsinfra.NewBroker()
	chatSvc := chat.NewService(conversations, ideas, opps, credits, aiSvc, broker)

	conv, err := chatSvc.Create(ctx, user.ID)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	// S'abonner avant l'envoi pour attraper tous les événements.
	evtCh, cancel := broker.Subscribe(user.ID)
	defer cancel()

	if _, err := chatSvc.SendMessage(ctx, user.ID, conv.ID, "Propose des idées de produits"); err != nil {
		t.Fatalf("send message: %v", err)
	}

	// Attendre l'événement chat.completed (tour asynchrone).
	deadline := time.After(10 * time.Second)
	var completed port.AppEvent
	for done := false; !done; {
		select {
		case evt, ok := <-evtCh:
			if !ok {
				t.Fatal("canal events fermé inattendu")
			}
			if evt.Type == port.EventChatCompleted {
				completed = evt
				done = true
			}
		case <-deadline:
			t.Fatal("timeout : chat.completed jamais reçu")
		}
	}

	var payload struct {
		ConversationID string `json:"conversation_id"`
		Message        struct {
			ID      string `json:"id"`
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Ideas []struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"ideas"`
	}
	if err := json.Unmarshal(completed.Data, &payload); err != nil {
		t.Fatalf("decode completed payload: %v", err)
	}
	if payload.ConversationID != conv.ID {
		t.Errorf("conversation_id = %q, want %q", payload.ConversationID, conv.ID)
	}
	if len(payload.Ideas) != 3 {
		t.Fatalf("expected 3 ideas in event, got %d", len(payload.Ideas))
	}
	for _, i := range payload.Ideas {
		if i.Status != domain.IdeaDraft {
			t.Errorf("idea %q status = %q, want draft", i.Title, i.Status)
		}
	}

	// Le message assistant est persisté SANS le marqueur @@IDEAS.
	if got := payload.Message.Content; got != "Voici 3 pistes pour vous lancer" {
		t.Errorf("assistant content = %q", got)
	}

	// Les idées sont en base, rattachées à la conversation.
	dbIdeas, err := ideas.ListByConversation(ctx, user.ID, conv.ID)
	if err != nil {
		t.Fatalf("list ideas: %v", err)
	}
	if len(dbIdeas) != 3 {
		t.Fatalf("expected 3 ideas in DB, got %d", len(dbIdeas))
	}

	// Messages : 1 user + 1 assistant, titre auto.
	msgs, err := conversations.ListMessages(ctx, conv.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != domain.ConversationMessageUser || msgs[1].Role != domain.ConversationMessageAssistant {
		t.Fatalf("unexpected roles: %s, %s", msgs[0].Role, msgs[1].Role)
	}

	// Crédits : 2 consommés (idea_generation).
	acc, err := credits.GetAccount(ctx, user.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if acc.Available() != 98 {
		t.Fatalf("expected available 98 after idea generation, got %d", acc.Available())
	}

	// Historique rechargé via Detail (hydratation page chat).
	detail, err := chatSvc.Detail(ctx, user.ID, conv.ID)
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if detail.Conversation.Title == "" {
		t.Error("conversation title should be auto-set from first message")
	}
	if len(detail.Ideas) != 3 || len(detail.Messages) != 2 {
		t.Errorf("detail ideas=%d messages=%d", len(detail.Ideas), len(detail.Messages))
	}
}
