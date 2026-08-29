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

func TestOpenAIStreamComplete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %q, want /chat/completions", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(
			"data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n" +
				"data: {\"choices\":[{\"delta\":{\"content\":\" world\"}}]}\n\n" +
				"data: [DONE]\n\n",
		))
	}))
	defer ts.Close()

	c := NewOpenAI("test-key", ts.URL)
	var got string
	err := c.StreamComplete(context.Background(), port.LLMRequest{Model: "gpt-5.6-luna", Messages: []port.LLMMessage{{Role: "user", Content: "hi"}}}, func(delta string) error {
		got += delta
		return nil
	})
	if err != nil {
		t.Fatalf("StreamComplete: %v", err)
	}
	if got != "Hello world" {
		t.Fatalf("streamed = %q, want %q", got, "Hello world")
	}
}

func TestOpenAIResearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Errorf("path = %q, want /responses", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if tools, ok := body["tools"].([]any); !ok || len(tools) == 0 {
			t.Errorf("expected web_search tool in body: %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output_text": "{\"opportunities\":[]}",
			"output": [
				{"type":"message","content":[{"type":"output_text","text":"ignored","annotations":[{"type":"url_citation","url":"https://example.com/a","title":"Source A"}]}]},
				{"type":"web_search_call","output":[{"url":"https://example.com/b","title":"Source B"}]}
			]
		}`))
	}))
	defer ts.Close()

	c := NewOpenAI("test-key", ts.URL)
	res, err := c.Research(context.Background(), port.ResearchRequest{
		Model:  "gpt-5.6-terra",
		System: "be concise",
		Query:  "mobile money in Senegal",
	})
	if err != nil {
		t.Fatalf("Research: %v", err)
	}
	if res.Content != "{\"opportunities\":[]}" {
		t.Errorf("content = %q", res.Content)
	}
	if len(res.Sources) != 2 {
		t.Fatalf("sources = %+v, want 2", res.Sources)
	}
	if res.Sources[0].URL != "https://example.com/a" || res.Sources[1].URL != "https://example.com/b" {
		t.Errorf("unexpected sources: %+v", res.Sources)
	}
}
