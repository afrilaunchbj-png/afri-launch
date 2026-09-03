package chat

import (
	"encoding/json"
	"fmt"
	"strings"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

// chatSystem est la consigne système du copilote conversationnel.
// Protocole d'outils embarqué dans le flux texte (pas de function calling) :
//   - "@@SEARCH {json}" seul → le backend exécute la recherche en ligne puis
//     relance le modèle avec les résultats ;
//   - bloc "@@IDEAS … @@END" en fin de réponse → le backend crée les idées
//     product_ideas et les retire du texte visible par l'utilisateur.
const chatSystem = `You are AfriLaunch Copilot, a product-strategy copilot for African entrepreneurs.
Through conversation you help the user find a market opportunity (with verified data) and shape ONE product idea ready to become a digital product project (ebook + marketing assets + sales page).

Rules:
- NEVER invent statistics. Only mention figures that appear in research results provided to you, with their source. Otherwise reason qualitatively.
- Write in the user's language (default: French).
- Keep replies short and conversational (mobile chat). Ask at most one question per reply.
- Guide the user toward ONE validated idea: explore the market, propose ideas, challenge them, refine title and hook.

Tools:
- When current/verified market data is needed, your ENTIRE reply must be exactly one line:
@@SEARCH {"query":"...","sector":"...","markets":["..."]}
Then stop immediately. The system will run the online search and return verified results to you.
- When you propose 1-5 product ideas, finish your reply with exactly:
@@IDEAS
{"ideas":[{"title":"...","hook":"...","explanation":"...","subtitle":"...","audience":"...","problem":"...","promise":"...","format":"...","estimated_price":"...","difficulty":"...","market_evidence":"...","why_now":"...","competitive_angle":"..."}]}
@@END
Nothing after @@END. "hook" is a punchy one-line pitch with a clear, honest number ONLY if supported by evidence (never invent a statistic). Write every field in the conversation language.`

// searchArgs sont les arguments du marqueur @@SEARCH.
type searchArgs struct {
	Query   string   `json:"query"`
	Sector  string   `json:"sector"`
	Markets []string `json:"markets"`
}

// ideasPayload est le JSON du bloc @@IDEAS.
type ideasPayload struct {
	Ideas []ideaInput `json:"ideas"`
}

// ideaInput est une idée proposée par le copilote.
type ideaInput struct {
	Title            string `json:"title"`
	Hook             string `json:"hook"`
	Explanation      string `json:"explanation"`
	Subtitle         string `json:"subtitle"`
	Audience         string `json:"audience"`
	Problem          string `json:"problem"`
	Promise          string `json:"promise"`
	Format           string `json:"format"`
	EstimatedPrice   string `json:"estimated_price"`
	Difficulty       string `json:"difficulty"`
	MarketEvidence   string `json:"market_evidence"`
	WhyNow           string `json:"why_now"`
	CompetitiveAngle string `json:"competitive_angle"`
}

// researchToolResult est injecté dans le prompt après une recherche en ligne.
func researchToolResult(ops []domain.Opportunity, sources []port.Source) string {
	opsOut := make([]map[string]any, 0, len(ops))
	for _, o := range ops {
		opsOut = append(opsOut, map[string]any{
			"id":         o.ID,
			"country":    o.Country,
			"title":      o.Title,
			"summary":    o.Summary,
			"difficulty": o.Difficulty,
			"signal":     o.Signal,
			"score":      o.Score,
			"scores":     o.Scores,
			"evidence":   o.Evidence,
		})
	}
	b, _ := json.Marshal(map[string]any{"opportunities": opsOut})
	out := string(b)
	if len(sources) > 0 {
		src, _ := json.Marshal(sources)
		out += "\nSources: " + string(src)
	}
	return out
}

// parseSearchLine décode une ligne "@@SEARCH {json}".
func parseSearchLine(line string) (searchArgs, error) {
	var args searchArgs
	payload := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "@@SEARCH"))
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		return searchArgs{}, fmt.Errorf("marqueur @@SEARCH illisible: %w", err)
	}
	if strings.TrimSpace(args.Query) == "" {
		return searchArgs{}, fmt.Errorf("marqueur @@SEARCH sans requête")
	}
	return args, nil
}

// parseIdeasBlock décode le bloc capturé après @@IDEAS (jusqu'à @@END).
func parseIdeasBlock(block string) (ideasPayload, bool) {
	block = strings.TrimSpace(block)
	block = strings.TrimSuffix(block, "@@END")
	block = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(block), "```json"))
	block = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(block), "```"))
	if block == "" {
		return ideasPayload{}, false
	}
	var ideas ideasPayload
	if err := json.Unmarshal([]byte(block), &ideas); err != nil {
		return ideasPayload{}, false
	}
	return ideas, true
}

// maxTurns borne l'historique envoyé au LLM.
const maxTurns = 20

// buildLLMMessages construit la conversation multi-tours pour le LLM.
func buildLLMMessages(history []domain.ConversationMessage) []port.LLMMessage {
	if len(history) > maxTurns {
		history = history[len(history)-maxTurns:]
	}
	out := make([]port.LLMMessage, 0, len(history))
	for _, m := range history {
		if m.Role != domain.ConversationMessageUser && m.Role != domain.ConversationMessageAssistant {
			continue
		}
		out = append(out, port.LLMMessage{Role: m.Role, Content: m.Content})
	}
	return out
}
