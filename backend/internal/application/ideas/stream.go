package ideas

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"afrilaunch/backend/internal/application/ai"
	"afrilaunch/backend/internal/domain"
)

// ideaReviseSystem est la consigne système de l'itération sur une idée.
// Le LLM répond en texte lisible (diffusé en streaming), pas en JSON.
const ideaReviseSystem = `You are a product-idea coach. Given a product idea and the user's feedback,
revise the title, hook and explanation accordingly.
Respond ONLY with the revised fields in this exact format (no markdown, no JSON, no preamble):
TITLE: <revised title>
HOOK: <revised hook>
EXPLANATION: <revised explanation>
- "hook" is a punchy one-line pitch with a clear, honest number (only numbers already present in the idea/evidence — never invent).
- Keep each field on a single line.
- Keep the revised text in the same language as the current idea.`

// ideaRevisePrompt construit la consigne d'itération avec l'historique.
func ideaRevisePrompt(idea domain.ProductIdea, history []domain.IdeaMessage, feedback string) string {
	var b strings.Builder
	b.WriteString("Product idea to refine:\n")
	b.WriteString("TITLE: " + idea.Title + "\n")
	b.WriteString("HOOK: " + idea.Hook + "\n")
	b.WriteString("EXPLANATION: " + idea.Explanation + "\n\n")
	if len(history) > 0 {
		b.WriteString("Conversation history:\n")
		for _, m := range history {
			b.WriteString(m.Role + ": " + m.Content + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("Latest user feedback: " + feedback + "\n\n")
	b.WriteString("Revise the title, hook and explanation to address this feedback.")
	return b.String()
}

// parseRevision extrait titre/accroche/explication du texte streamé.
func parseRevision(text string) (title, hook, explanation string, err error) {
	lines := strings.Split(text, "\n")
	var current *string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "TITLE:"):
			title = strings.TrimSpace(strings.TrimPrefix(trimmed, "TITLE:"))
			current = &title
		case strings.HasPrefix(trimmed, "HOOK:"):
			hook = strings.TrimSpace(strings.TrimPrefix(trimmed, "HOOK:"))
			current = &hook
		case strings.HasPrefix(trimmed, "EXPLANATION:"):
			explanation = strings.TrimSpace(strings.TrimPrefix(trimmed, "EXPLANATION:"))
			current = &explanation
		default:
			if current != nil && trimmed != "" {
				*current += " " + trimmed
			}
		}
	}
	if title == "" && hook == "" && explanation == "" {
		return "", "", "", errors.New("réponse illisible du modèle")
	}
	return title, hook, explanation, nil
}

// StreamMessage enregistre le feedback, diffuse la révision en streaming et
// met à jour l'idée. emit reçoit chaque delta de texte.
func (s *Service) StreamMessage(ctx context.Context, userID, ideaID, content string, emit func(string) error) (domain.ProductIdea, error) {
	if content == "" {
		return domain.ProductIdea{}, domain.ErrInvalidInput
	}
	idea, err := s.ideas.Get(ctx, userID, ideaID)
	if err != nil {
		return domain.ProductIdea{}, err
	}
	history, err := s.ideaMessages.ListByIdea(ctx, ideaID)
	if err != nil {
		return domain.ProductIdea{}, err
	}

	if _, err := s.ideaMessages.Create(ctx, domain.IdeaMessage{
		IdeaID: ideaID, UserID: userID, Role: domain.IdeaMessageUser, Content: content,
	}); err != nil {
		return domain.ProductIdea{}, err
	}

	// Réservation de crédits pour l'itération.
	cost, err := s.credits.GetGenerationCost(ctx, domain.OperationIdeaGeneration)
	if err != nil {
		return domain.ProductIdea{}, err
	}
	reference := "revise:" + uuid.NewString()
	if _, err := s.credits.Reserve(ctx, userID, cost.Credits, domain.OperationIdeaGeneration, reference, 5*time.Minute); err != nil {
		return domain.ProductIdea{}, err
	}

	var buf strings.Builder
	err = s.ai.StreamComplete(ctx, ai.TaskIdeation, ideaReviseSystem, ideaRevisePrompt(idea, history, content), func(delta string) error {
		buf.WriteString(delta)
		return emit(delta)
	})
	if err != nil {
		_ = s.credits.Release(ctx, userID, reference)
		return domain.ProductIdea{}, err
	}

	title, hook, explanation, err := parseRevision(buf.String())
	if err != nil {
		_ = s.credits.Release(ctx, userID, reference)
		return domain.ProductIdea{}, err
	}

	updated, err := s.ideas.UpdateContent(ctx, domain.ProductIdea{
		ID: ideaID, Title: title, Hook: hook, Explanation: explanation,
	})
	if err != nil {
		_ = s.credits.Release(ctx, userID, reference)
		return domain.ProductIdea{}, err
	}

	if _, err := s.ideaMessages.Create(ctx, domain.IdeaMessage{
		IdeaID:  ideaID,
		UserID:  userID,
		Role:    domain.IdeaMessageAssistant,
		Content: "Titre: " + updated.Title + "\nAccroche: " + updated.Hook + "\nExplication: " + updated.Explanation,
	}); err != nil {
		_ = s.credits.Release(ctx, userID, reference)
		return domain.ProductIdea{}, err
	}

	_, _ = s.credits.Consume(ctx, userID, reference)
	return updated, nil
}
