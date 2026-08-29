package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"afrilaunch/backend/internal/application/port"
)

func TestHeyGenSubmit(t *testing.T) {
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/videos" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "hg-key" {
			t.Errorf("missing X-Api-Key")
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"video_id":"v_abc","status":"waiting"}}`))
	}))
	defer ts.Close()

	h := NewHeyGen("hg-key", ts.URL)
	id, err := h.Submit(context.Background(), port.VideoRequest{
		AvatarID:    "avatar-1",
		Script:      "Bonjour",
		VoiceID:     "voice-1",
		AspectRatio: "9:16",
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if id != "v_abc" {
		t.Fatalf("video id = %q", id)
	}
	if gotBody["type"] != "avatar" || gotBody["avatar_id"] != "avatar-1" || gotBody["aspect_ratio"] != "9:16" {
		t.Errorf("unexpected body: %+v", gotBody)
	}
}

func TestHeyGenStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/videos/v_abc" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":"v_abc","status":"completed","video_url":"https://files/heygen.mp4"}}`))
	}))
	defer ts.Close()

	h := NewHeyGen("hg-key", ts.URL)
	res, err := h.Status(context.Background(), "v_abc")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if res.Status != port.VideoCompleted || res.URL == "" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestHeyGenFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":"v_abc","status":"failed","failure_message":"render timeout"}}`))
	}))
	defer ts.Close()

	h := NewHeyGen("hg-key", ts.URL)
	res, err := h.Status(context.Background(), "v_abc")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if res.Status != port.VideoFailed || res.Error == "" {
		t.Fatalf("unexpected result: %+v", res)
	}
}
