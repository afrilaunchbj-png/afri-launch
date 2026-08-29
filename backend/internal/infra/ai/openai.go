// Package ai fournit les adaptateurs des providers IA (OpenAI, HeyGen).
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size,omitempty"`
	N              int    `json:"n,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
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

// Generate appelle /images/generations (retour b64_json).
func (o *OpenAI) Generate(ctx context.Context, req port.ImageRequest) (port.Image, error) {
	body := openAIImageRequest{
		Model:          req.Model,
		Prompt:         req.Prompt,
		Size:           req.Size,
		N:              1,
		ResponseFormat: "b64_json",
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

	return port.Image{B64JSON: out.Data[0].B64JSON, URL: out.Data[0].URL}, nil
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
