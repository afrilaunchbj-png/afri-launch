package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"afrilaunch/backend/internal/application/port"
)

// openAIResponsesRequest est le corps de POST /v1/responses (Responses API).
type openAIResponsesRequest struct {
	Model        string       `json:"model"`
	Instructions string       `json:"instructions,omitempty"`
	Input        string       `json:"input"`
	Tools        []openAITool `json:"tools,omitempty"`
}

type openAITool struct {
	Type string `json:"type"`
}

type openAIResponsesResponse struct {
	OutputText string             `json:"output_text"`
	Output     []openAIOutputItem `json:"output"`
	Error      *openAIError       `json:"error,omitempty"`
}

type openAIOutputItem struct {
	Type    string              `json:"type"`
	Content []openAIContentPart `json:"content"`
	Output  []json.RawMessage   `json:"output"`
}

type openAIContentPart struct {
	Type        string             `json:"type"`
	Text        string             `json:"text"`
	Annotations []openAIAnnotation `json:"annotations"`
}

type openAIAnnotation struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

// Research appelle l'API Responses avec l'outil web_search, puis renvoie le
// texte produit et les sources (citations) remontées.
func (o *OpenAI) Research(ctx context.Context, req port.ResearchRequest) (port.ResearchResult, error) {
	body := openAIResponsesRequest{
		Model:        req.Model,
		Instructions: req.System,
		Input:        req.Query,
		Tools:        []openAITool{{Type: "web_search"}},
	}

	var out openAIResponsesResponse
	if err := o.post(ctx, "/responses", body, &out); err != nil {
		return port.ResearchResult{}, err
	}
	if out.Error != nil {
		return port.ResearchResult{}, out.Error
	}

	content := out.OutputText
	if content == "" {
		for _, item := range out.Output {
			for _, p := range item.Content {
				if p.Type == "output_text" {
					content += p.Text
				}
			}
		}
	}

	var sources []port.Source
	seen := map[string]bool{}
	for _, item := range out.Output {
		for _, p := range item.Content {
			for _, a := range p.Annotations {
				if a.Type == "url_citation" && a.URL != "" && !seen[a.URL] {
					seen[a.URL] = true
					sources = append(sources, port.Source{Title: a.Title, URL: a.URL})
				}
			}
		}
		for _, raw := range item.Output {
			var m map[string]any
			if json.Unmarshal(raw, &m) != nil {
				continue
			}
			if u, ok := m["url"].(string); ok && u != "" && !seen[u] {
				seen[u] = true
				title, _ := m["title"].(string)
				sources = append(sources, port.Source{Title: title, URL: u})
			}
		}
	}

	if content == "" {
		return port.ResearchResult{}, fmt.Errorf("openai: empty research output")
	}
	return port.ResearchResult{Content: content, Sources: sources}, nil
}
