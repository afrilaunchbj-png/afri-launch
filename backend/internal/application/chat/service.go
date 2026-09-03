// Package chat implémente le copilote conversationnel : un chat multi-tours
// qui remplace le parcours statique opportunités → idées. Le streaming passe
// par le canal temps réel unique (port.EventPublisher), pas par la requête
// HTTP : POST /conversations/{id}/messages répond 202 et les deltas arrivent
// sur GET /api/v1/events.
package chat

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/application/jobs"
	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// turnTimeout borne la durée d'un tour (recherche en ligne + réponse).
const turnTimeout = 10 * time.Minute

// creditTTL est la durée de réservation des crédits d'un tour.
const creditTTL = 5 * time.Minute

// ideaEvent est la sérialisation JSON d'une idée (événements chat.completed).
type ideaEvent struct {
	ID               string  `json:"id"`
	OpportunityID    *string `json:"opportunity_id,omitempty"`
	ConversationID   *string `json:"conversation_id,omitempty"`
	Title            string  `json:"title"`
	Hook             string  `json:"hook"`
	Explanation      string  `json:"explanation"`
	Subtitle         string  `json:"subtitle"`
	Audience         string  `json:"audience"`
	Problem          string  `json:"problem"`
	Promise          string  `json:"promise"`
	Format           string  `json:"format"`
	EstimatedPrice   string  `json:"estimated_price"`
	Difficulty       string  `json:"difficulty"`
	MarketEvidence   string  `json:"market_evidence"`
	WhyNow           string  `json:"why_now"`
	CompetitiveAngle string  `json:"competitive_angle"`
	Status           string  `json:"status"`
}

func toIdeaEvent(i domain.ProductIdea) ideaEvent {
	return ideaEvent{
		ID: i.ID, OpportunityID: i.OpportunityID, ConversationID: i.ConversationID,
		Title: i.Title, Hook: i.Hook, Explanation: i.Explanation, Subtitle: i.Subtitle,
		Audience: i.Audience, Problem: i.Problem, Promise: i.Promise, Format: i.Format,
		EstimatedPrice: i.EstimatedPrice, Difficulty: i.Difficulty, MarketEvidence: i.MarketEvidence,
		WhyNow: i.WhyNow, CompetitiveAngle: i.CompetitiveAngle, Status: i.Status,
	}
}

// turnResult est le résultat d'une réponse streamée du modèle.
type turnResult struct {
	raw      string      // texte brut du modèle (avec marqueurs)
	visible  string      // texte émis à l'utilisateur (sans bloc idées)
	search   *searchArgs // ligne @@SEARCH détectée
	ideas    ideasPayload
	hasIdeas bool
}

// Service orchestre les conversations et les tours du copilote.
type Service struct {
	conversations port.ConversationRepository
	ideas         port.IdeaRepository
	opportunities port.OpportunityRepository
	credits       port.CreditRepository
	prefs         port.PreferenceRepository
	ai            *ai.Service
	events        port.EventPublisher

	// locks sérialisent les tours par conversation (2 envois rapides).
	lockMu sync.Mutex
	locks  map[string]*sync.Mutex
}

// NewService construit le service de chat.
func NewService(
	conversations port.ConversationRepository,
	ideas port.IdeaRepository,
	opportunities port.OpportunityRepository,
	credits port.CreditRepository,
	prefs port.PreferenceRepository,
	aiSvc *ai.Service,
	events port.EventPublisher,
) *Service {
	return &Service{
		conversations: conversations,
		ideas:         ideas,
		opportunities: opportunities,
		credits:       credits,
		prefs:         prefs,
		ai:            aiSvc,
		events:        events,
		locks:         make(map[string]*sync.Mutex),
	}
}

// Create ouvre une nouvelle conversation.
func (s *Service) Create(ctx context.Context, userID string) (domain.Conversation, error) {
	return s.conversations.Create(ctx, domain.Conversation{UserID: userID})
}

// List renvoie les conversations récentes de l'utilisateur.
func (s *Service) List(ctx context.Context, userID string, limit, offset int) ([]domain.Conversation, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.conversations.List(ctx, userID, limit, offset)
}

// Detail regroupe tout le contexte d'une conversation (hydratation page chat).
type Detail struct {
	Conversation domain.Conversation
	Opportunity  *domain.Opportunity
	Messages     []domain.ConversationMessage
	Ideas        []domain.ProductIdea
}

// Detail renvoie la conversation + son opportunité + messages + idées.
func (s *Service) Detail(ctx context.Context, userID, id string) (Detail, error) {
	conv, err := s.conversations.Get(ctx, userID, id)
	if err != nil {
		return Detail{}, err
	}
	d := Detail{Conversation: conv}

	if conv.OpportunityID != nil {
		if opp, err := s.opportunities.Get(ctx, *conv.OpportunityID); err == nil {
			d.Opportunity = &opp
		}
	}
	if d.Messages, err = s.conversations.ListMessages(ctx, conv.ID); err != nil {
		return Detail{}, err
	}
	if d.Ideas, err = s.ideas.ListByConversation(ctx, userID, conv.ID); err != nil {
		return Detail{}, err
	}
	return d, nil
}

// SendMessage persiste le message utilisateur et lance le tour du copilote
// en arrière-plan. Renvoie l'ID du message assistant en cours de génération.
func (s *Service) SendMessage(ctx context.Context, userID, convID, content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", domain.ErrInvalidInput
	}
	conv, err := s.conversations.Get(ctx, userID, convID)
	if err != nil {
		return "", err
	}

	userMsg := domain.ConversationMessage{
		ID:             uuid.NewString(),
		ConversationID: convID,
		UserID:         userID,
		Role:           domain.ConversationMessageUser,
		Content:        content,
	}
	if _, err := s.conversations.CreateMessage(ctx, userMsg); err != nil {
		return "", err
	}
	if conv.Title == "" {
		_, _ = s.conversations.SetTitle(ctx, convID, truncateTitle(content))
	}
	if _, err := s.conversations.Touch(ctx, convID); err != nil {
		return "", err
	}

	assistantID := uuid.NewString()
	s.publish(conv.UserID, port.EventChatStarted, map[string]any{
		"conversation_id": convID, "message_id": assistantID,
	})

	// Langue du compte : le copilote répond toujours dans cette langue.
	language := domain.LanguageFr
	if s.prefs != nil {
		if pref, err := s.prefs.GetOrCreate(ctx, userID); err == nil && pref.Language != "" {
			language = pref.Language
		}
	}

	// Le tour survit à la requête HTTP (réponse 202 immédiate).
	runCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), turnTimeout)
	go func() {
		defer cancel()
		mu := s.lockFor(convID)
		mu.Lock()
		defer mu.Unlock()
		s.runTurn(runCtx, conv, assistantID, language)
	}()
	return assistantID, nil
}

func (s *Service) lockFor(convID string) *sync.Mutex {
	s.lockMu.Lock()
	defer s.lockMu.Unlock()
	if s.locks[convID] == nil {
		s.locks[convID] = &sync.Mutex{}
	}
	return s.locks[convID]
}

// runTurn exécute une boucle agent bornée : réponse (streamée) ou recherche
// en ligne (outil) suivie d'une réponse. Maximum 2 rounds, 1 recherche/tour.
// language est la langue du compte (préférences) : toutes les sorties LLM
// (messages + champs d'idées) sont dans cette langue.
func (s *Service) runTurn(ctx context.Context, conv domain.Conversation, assistantID, language string) {
	history, err := s.conversations.ListMessages(ctx, conv.ID)
	if err != nil {
		s.failTurn(ctx, conv, assistantID, err)
		return
	}
	msgs := buildLLMMessages(history)

	searched := false
	for round := 0; round < 2; round++ {
		res, err := s.streamAnswer(ctx, conv, assistantID, language, msgs)
		if err != nil {
			s.failTurn(ctx, conv, assistantID, err)
			return
		}

		// Le modèle demande une recherche en ligne.
		if res.search != nil {
			if searched {
				msgs = append(msgs,
					port.LLMMessage{Role: "assistant", Content: res.raw},
					port.LLMMessage{Role: "user", Content: "A search has already been run this turn. Answer now with the data you already have."},
				)
				continue
			}
			searched = true
			ops, err := s.runSearch(ctx, conv, assistantID, *res.search, language)
			if err != nil {
				s.failTurn(ctx, conv, assistantID, err)
				return
			}
			// Les idées proposées ensuite doivent être liées à l'opportunité trouvée.
			if len(ops) > 0 {
				conv.OpportunityID = &ops[0].ID
			}
			msgs = append(msgs,
				port.LLMMessage{Role: "assistant", Content: res.raw},
				port.LLMMessage{Role: "user", Content: "RESEARCH RESULTS (verified, use these figures only):\n" + researchToolResult(ops, nil)},
			)
			continue
		}

		// Réponse finale : enregistrer le bloc d'idées éventuel.
		visible := strings.TrimSpace(res.visible)
		var created []domain.ProductIdea
		if res.hasIdeas && len(res.ideas.Ideas) > 0 {
			if created, err = s.createIdeas(ctx, conv, res.ideas.Ideas); err != nil {
				s.failTurn(ctx, conv, assistantID, err)
				return
			}
		}
		s.finishTurn(ctx, conv, assistantID, visible, created)
		return
	}
	s.failTurn(ctx, conv, assistantID, errors.New("le copilote a enchaîné trop d'outils"))
}

// errAbortStop interrompt proprement le flux LLM (cas contrôlés).
var errAbortStop = errors.New("chat: stop stream")

const (
	markerSearch = "@@SEARCH"
	markerIdeas  = "@@IDEAS"
)

// markerHold renvoie la longueur du suffixe de s qui pourrait être le début
// d'un marqueur d'outil (à retenir en attente du delta suivant).
func markerHold(s string) int {
	for l := len(s); l > 0; l-- {
		suffix := s[len(s)-l:]
		if strings.HasPrefix(markerSearch, suffix) || strings.HasPrefix(markerIdeas, suffix) {
			return l
		}
	}
	return 0
}

// streamAnswer streame la réponse du modèle et détecte les marqueurs d'outils.
// Machine à états : 0 = texte visible, 1 = capture du bloc @@IDEAS, 2 = ligne
// @@SEARCH (drainage puis arrêt du flux).
func (s *Service) streamAnswer(ctx context.Context, conv domain.Conversation, assistantID, language string, msgs []port.LLMMessage) (turnResult, error) {
	var res turnResult

	var (
		pending   string // texte pas encore émis (détection "@@" multi-delta)
		mode      int
		ideasBuf  strings.Builder
		searchBuf strings.Builder
	)
	const (
		modeText   = 0
		modeIdeas  = 1
		modeSearch = 2
	)

	publish := func(chunk string) {
		res.visible += chunk
		s.publish(conv.UserID, port.EventChatDelta, map[string]any{
			"conversation_id": conv.ID, "message_id": assistantID, "delta": chunk,
		})
	}

	emit := func(delta string) error {
		res.raw += delta
		switch mode {
		case modeIdeas:
			ideasBuf.WriteString(delta)
			if strings.HasSuffix(ideasBuf.String(), "@@END") {
				return errAbortStop
			}
			return nil
		case modeSearch:
			searchBuf.WriteString(delta)
			if strings.Contains(searchBuf.String(), "\n") {
				return errAbortStop
			}
			return nil
		}

		pending += delta

		// Bloc d'idées : tout ce qui suit @@IDEAS est capturé, pas émis.
		if idx := strings.Index(pending, "@@IDEAS"); idx >= 0 {
			publish(pending[:idx])
			ideasBuf.WriteString(pending[idx+len("@@IDEAS"):])
			pending = ""
			mode = modeIdeas
			return nil
		}
		// Ligne de recherche (normalement seule en tout début de réponse).
		if idx := strings.Index(pending, "@@SEARCH"); idx >= 0 {
			publish(pending[:idx])
			searchBuf.WriteString(pending[idx+len("@@SEARCH"):])
			pending = ""
			mode = modeSearch
			return nil
		}

		// Pas de marqueur complet : émettre tout sauf un éventuel suffixe
		// qui pourrait être le début d'un marqueur (@@SEARCH / @@IDEAS).
		if hold := markerHold(pending); hold < len(pending) {
			publish(pending[:len(pending)-hold])
			pending = pending[len(pending)-hold:]
		}
		return nil
	}

	err := s.ai.StreamMessages(ctx, ai.TaskIdeation, chatSystemPrompt(language), msgs, emit)
	if errors.Is(err, errAbortStop) {
		err = nil
	}
	if err != nil {
		return res, err
	}

	switch mode {
	case modeSearch:
		args, perr := parseSearchLine("@@SEARCH" + searchBuf.String())
		if perr != nil {
			return res, perr
		}
		res.search = &args
	case modeIdeas:
		res.ideas, res.hasIdeas = parseIdeasBlock(ideasBuf.String())
	default: // modeText : vider le tampon de garde
		publish(pending)
		pending = ""
	}
	return res, nil
}

// runSearch exécute la recherche en ligne (facturée) et crée les opportunités.
func (s *Service) runSearch(ctx context.Context, conv domain.Conversation, assistantID string, args searchArgs, language string) ([]domain.Opportunity, error) {
	s.publish(conv.UserID, port.EventChatTool, map[string]any{
		"conversation_id": conv.ID, "message_id": assistantID, "tool": "search", "status": "running",
	})

	cost, err := s.credits.GetGenerationCost(ctx, domain.OperationNicheResearch)
	if err != nil {
		return nil, err
	}
	reference := "chat-research:" + uuid.NewString()
	if _, err := s.credits.Reserve(ctx, conv.UserID, cost.Credits, domain.OperationNicheResearch, reference, creditTTL); err != nil {
		return nil, err
	}

	result, err := s.ai.Research(ctx, jobs.ResearchSystem, jobs.ResearchQuery(args.Query, args.Sector, args.Markets, language))
	if err != nil {
		_ = s.credits.Release(ctx, conv.UserID, reference)
		s.publish(conv.UserID, port.EventChatTool, map[string]any{
			"conversation_id": conv.ID, "message_id": assistantID, "tool": "search", "status": "failed", "error": err.Error(),
		})
		return nil, err
	}

	inputs, err := jobs.ParseResearchResult(result.Content)
	if err != nil {
		_ = s.credits.Release(ctx, conv.UserID, reference)
		return nil, err
	}

	ops := make([]domain.Opportunity, 0, len(inputs))
	for _, in := range inputs {
		opp, err := s.opportunities.Create(ctx, jobs.OpportunityFromResearch(in, args.Sector, language, conv.UserID, nil))
		if err != nil {
			_ = s.credits.Release(ctx, conv.UserID, reference)
			return nil, err
		}
		ops = append(ops, opp)
	}

	// La conversation suit l'opportunité la mieux scorée.
	if len(ops) > 0 {
		if _, err := s.conversations.SetOpportunity(ctx, conv.ID, &ops[0].ID); err != nil {
			slog.Warn("chat: set opportunity", "err", err)
		}
	}
	if _, err := s.credits.Consume(ctx, conv.UserID, reference); err != nil {
		slog.Warn("chat: consume research credits", "err", err)
	}

	ids := make([]string, 0, len(ops))
	for _, o := range ops {
		ids = append(ids, o.ID)
	}
	s.publish(conv.UserID, port.EventChatTool, map[string]any{
		"conversation_id": conv.ID, "message_id": assistantID, "tool": "search",
		"status": "completed", "opportunity_ids": ids, "count": len(ids),
	})
	return ops, nil
}

// createIdeas persiste les idées proposées par le copilote (facturé).
func (s *Service) createIdeas(ctx context.Context, conv domain.Conversation, ins []ideaInput) ([]domain.ProductIdea, error) {
	cost, err := s.credits.GetGenerationCost(ctx, domain.OperationIdeaGeneration)
	if err != nil {
		return nil, err
	}
	reference := "chat-ideas:" + uuid.NewString()
	if _, err := s.credits.Reserve(ctx, conv.UserID, cost.Credits, domain.OperationIdeaGeneration, reference, creditTTL); err != nil {
		return nil, err
	}

	created := make([]domain.ProductIdea, 0, len(ins))
	for _, in := range ins {
		idea, err := s.ideas.Create(ctx, domain.ProductIdea{
			UserID:           conv.UserID,
			OpportunityID:    conv.OpportunityID,
			ConversationID:   &conv.ID,
			Title:            in.Title,
			Hook:             in.Hook,
			Explanation:      in.Explanation,
			Subtitle:         in.Subtitle,
			Audience:         in.Audience,
			Problem:          in.Problem,
			Promise:          in.Promise,
			Format:           in.Format,
			EstimatedPrice:   in.EstimatedPrice,
			Difficulty:       in.Difficulty,
			MarketEvidence:   in.MarketEvidence,
			WhyNow:           in.WhyNow,
			CompetitiveAngle: in.CompetitiveAngle,
			Status:           domain.IdeaDraft,
		})
		if err != nil {
			_ = s.credits.Release(ctx, conv.UserID, reference)
			return nil, err
		}
		created = append(created, idea)
	}

	if _, err := s.credits.Consume(ctx, conv.UserID, reference); err != nil {
		slog.Warn("chat: consume idea credits", "err", err)
	}
	return created, nil
}

// finishTurn persiste le message assistant et publie l'événement de fin.
func (s *Service) finishTurn(ctx context.Context, conv domain.Conversation, assistantID, visible string, ideas []domain.ProductIdea) {
	payload := map[string]any{}
	if len(ideas) > 0 {
		ids := make([]string, 0, len(ideas))
		for _, i := range ideas {
			ids = append(ids, i.ID)
		}
		payload["idea_ids"] = ids
	}
	rawPayload, _ := json.Marshal(payload)

	msg, err := s.conversations.CreateMessage(ctx, domain.ConversationMessage{
		ID:             assistantID,
		ConversationID: conv.ID,
		UserID:         conv.UserID,
		Role:           domain.ConversationMessageAssistant,
		Content:        visible,
		Payload:        rawPayload,
	})
	if err != nil {
		slog.Error("chat: persist assistant message", "err", err)
		s.publish(conv.UserID, port.EventChatError, map[string]any{
			"conversation_id": conv.ID, "message_id": assistantID, "error": "échec de l'enregistrement de la réponse",
		})
		return
	}
	_, _ = s.conversations.Touch(ctx, conv.ID)

	ideaEvents := make([]ideaEvent, 0, len(ideas))
	for _, i := range ideas {
		ideaEvents = append(ideaEvents, toIdeaEvent(i))
	}
	s.publish(conv.UserID, port.EventChatCompleted, map[string]any{
		"conversation_id": conv.ID,
		"message": map[string]any{
			"id": msg.ID, "role": msg.Role, "content": msg.Content,
			"payload": json.RawMessage(rawPayload), "created_at": msg.CreatedAt,
		},
		"ideas": ideaEvents,
	})
}

func (s *Service) failTurn(ctx context.Context, conv domain.Conversation, assistantID string, turnErr error) {
	slog.Warn("chat: turn failed", "conversation", conv.ID, "err", turnErr)
	msg := turnErr.Error()
	if errors.Is(turnErr, domain.ErrInsufficient) {
		msg = "Crédits insuffisants pour cette action."
	}
	// Message assistant d'erreur persisté pour l'historique.
	_, _ = s.conversations.CreateMessage(ctx, domain.ConversationMessage{
		ID:             assistantID,
		ConversationID: conv.ID,
		UserID:         conv.UserID,
		Role:           domain.ConversationMessageAssistant,
		Content:        msg,
	})
	s.publish(conv.UserID, port.EventChatError, map[string]any{
		"conversation_id": conv.ID, "message_id": assistantID, "error": msg,
	})
}

func (s *Service) publish(userID, eventType string, data map[string]any) {
	if s.events == nil {
		return
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	s.events.Publish(userID, port.AppEvent{Type: eventType, Data: raw})
}

func truncateTitle(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 60 {
		s = s[:57] + "…"
	}
	return s
}
