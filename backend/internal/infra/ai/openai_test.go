package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"afrilaunch/backend/internal/application/port"
)

func TestOpenAIComplete(t *testing.T) {
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %q, want /chat/completions", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing bearer auth")
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"bonjour"}}],"usage":{"prompt_tokens":11,"completion_tokens":7}}`))
	}))
	defer ts.Close()

	c := NewOpenAI("test-key", ts.URL)
	resp, err := c.Complete(context.Background(), port.LLMRequest{
		Model:    "gpt-5.6-terra",
		System:   "tu es utile",
		Messages: []port.LLMMessage{{Role: "user", Content: "salut"}},
		JSONMode: true,
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Content != "bonjour" || resp.InputTokens != 11 || resp.OutputTokens != 7 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if gotBody["model"] != "gpt-5.6-terra" {
		t.Errorf("model = %v", gotBody["model"])
	}
	if gotBody["response_format"] == nil {
		t.Errorf("expected response_format for JSONMode")
	}
}

func TestOpenAIGenerate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/generations" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"aW1n"}]}`))
	}))
	defer ts.Close()

	c := NewOpenAI("test-key", ts.URL)
	img, err := c.Generate(context.Background(), port.ImageRequest{Model: "gpt-image-2", Prompt: "couverture"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if img.B64JSON != "aW1n" {
		t.Fatalf("unexpected image: %+v", img)
	}
}

func TestOpenAIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limited","type":"rate_limit","code":"429"}}`))
	}))
	defer ts.Close()

	c := NewOpenAI("test-key", ts.URL)
	if _, err := c.Complete(context.Background(), port.LLMRequest{Model: "m"}); err == nil {
		t.Fatal("expected error for 429")
	}
}
