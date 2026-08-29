// Package ai fournit les adaptateurs des providers IA (OpenAI, HeyGen).
package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"afrilaunch/backend/internal/application/port"
)

const openAIDefaultBaseURL = "https://api.openai.com/v1"

// OpenAI implémente port.LLMProvider et port.ImageProvider (chat + images).
type OpenAI struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// NewOpenAI construit un client OpenAI.
func NewOpenAI(apiKey, baseURL string) *OpenAI {
	if baseURL == "" {
		baseURL = openAIDefaultBaseURL
	}
	return &OpenAI{apiKey: apiKey, baseURL: baseURL, http: &http.Client{Timeout: 120 * time.Second}}
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model          string           `json:"model"`
	Messages       []openAIMessage  `json:"messages"`
	MaxTokens      int              `json:"max_tokens,omitempty"`
	Temperature    *float64         `json:"temperature,omitempty"`
	Stream         bool             `json:"stream,omitempty"`
	ResponseFormat *json.RawMessage `json:"response_format,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *openAIError `json:"error,omitempty"`
}

type openAIImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Size   string `json:"size,omitempty"`
	N      int    `json:"n,omitempty"`
}

type openAIImageResponse struct {
	Data []struct {
		URL     string `json:"url,omitempty"`
		B64JSON string `json:"b64_json,omitempty"`
	} `json:"data"`
	Error *openAIError `json:"error,omitempty"`
}

type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func (e *openAIError) Error() string {
	if e == nil {
		return "openai error"
	}
	return e.Message
}

// Complete appelle /chat/completions.
func (o *OpenAI) Complete(ctx context.Context, req port.LLMRequest) (port.LLMResponse, error) {
	messages := make([]openAIMessage, 0, len(req.Messages)+1)
	if req.System != "" {
		messages = append(messages, openAIMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		messages = append(messages, openAIMessage{Role: m.Role, Content: m.Content})
	}

	body := openAIChatRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}
	if req.JSONMode {
		rf := json.RawMessage(`{"type":"json_object"}`)
		body.ResponseFormat = &rf
	}

	var out openAIChatResponse
	if err := o.post(ctx, "/chat/completions", body, &out); err != nil {
		return port.LLMResponse{}, err
	}
	if out.Error != nil {
		return port.LLMResponse{}, out.Error
	}
	if len(out.Choices) == 0 {
		return port.LLMResponse{}, fmt.Errorf("openai: empty choices")
	}

	return port.LLMResponse{
		Content:      out.Choices[0].Message.Content,
		Model:        req.Model,
		InputTokens:  out.Usage.PromptTokens,
		OutputTokens: out.Usage.CompletionTokens,
	}, nil
}

// openAIChatChunk est un fragment d'une réponse streamée.
type openAIChatChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// StreamComplete appelle /chat/completions en mode streaming (SSE) et émet
// chaque delta de texte via emit.
func (o *OpenAI) StreamComplete(ctx context.Context, req port.LLMRequest, emit func(string) error) error {
	messages := make([]openAIMessage, 0, len(req.Messages)+1)
	if req.System != "" {
		messages = append(messages, openAIMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		messages = append(messages, openAIMessage{Role: m.Role, Content: m.Content})
	}

	body := openAIChatRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr openAIChatResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Error != nil {
			return fmt.Errorf("openai: %s (status %d)", apiErr.Error.Message, resp.StatusCode)
		}
		return fmt.Errorf("openai: status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk openAIChatChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			if err := emit(chunk.Choices[0].Delta.Content); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

// Generate appelle /images/generations. Les modèles gpt-image renvoient du
// b64_json par défaut et ne supportent pas le paramètre `response_format`.
func (o *OpenAI) Generate(ctx context.Context, req port.ImageRequest) (port.Image, error) {
	body := openAIImageRequest{
		Model:  req.Model,
		Prompt: req.Prompt,
		Size:   req.Size,
		N:      1,
	}

	var out openAIImageResponse
	if err := o.post(ctx, "/images/generations", body, &out); err != nil {
		return port.Image{}, err
	}
	if out.Error != nil {
		return port.Image{}, out.Error
	}
	if len(out.Data) == 0 {
		return port.Image{}, fmt.Errorf("openai: empty image data")
	}

	img := port.Image{URL: out.Data[0].URL}
	if out.Data[0].B64JSON != "" {
		img.B64JSON = out.Data[0].B64JSON
		return img, nil
	}
	if out.Data[0].URL != "" {
		b64, err := o.downloadAsB64(ctx, out.Data[0].URL)
		if err != nil {
			return port.Image{}, err
		}
		img.B64JSON = b64
		return img, nil
	}
	return port.Image{}, fmt.Errorf("openai: empty image data")
}

// downloadAsB64 télécharge une image distante et la renvoie en base64.
func (o *OpenAI) downloadAsB64(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := o.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("openai: image download status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (o *OpenAI) post(ctx context.Context, path string, body, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr openAIChatResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Error != nil {
			return fmt.Errorf("openai: %s (status %d)", apiErr.Error.Message, resp.StatusCode)
		}
		return fmt.Errorf("openai: status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
