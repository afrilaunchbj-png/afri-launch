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

const heyGenDefaultBaseURL = "https://api.heygen.com"

// HeyGen implémente port.VideoProvider (vidéo avatar).
type HeyGen struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// NewHeyGen construit un client HeyGen.
func NewHeyGen(apiKey, baseURL string) *HeyGen {
	if baseURL == "" {
		baseURL = heyGenDefaultBaseURL
	}
	return &HeyGen{apiKey: apiKey, baseURL: baseURL, http: &http.Client{Timeout: 30 * time.Second}}
}

type heyGenCreateRequest struct {
	Type        string `json:"type"`
	AvatarID    string `json:"avatar_id"`
	Script      string `json:"script,omitempty"`
	VoiceID     string `json:"voice_id,omitempty"`
	AspectRatio string `json:"aspect_ratio,omitempty"`
	Resolution  string `json:"resolution,omitempty"`
	Title       string `json:"title,omitempty"`
}

type heyGenCreateResponse struct {
	Data struct {
		VideoID string `json:"video_id"`
		Status  string `json:"status"`
	} `json:"data"`
	Error *heyGenError `json:"error,omitempty"`
}

type heyGenVideoDetail struct {
	ID             string  `json:"id"`
	Status         string  `json:"status"`
	VideoURL       *string `json:"video_url"`
	FailureMessage *string `json:"failure_message"`
}

type heyGenGetResponse struct {
	Data  heyGenVideoDetail `json:"data"`
	Error *heyGenError      `json:"error,omitempty"`
}

type heyGenError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *heyGenError) Error() string {
	if e == nil {
		return "heygen error"
	}
	return e.Message
}

// Submit crée une vidéo avatar (POST /v3/videos).
func (h *HeyGen) Submit(ctx context.Context, req port.VideoRequest) (string, error) {
	body := heyGenCreateRequest{
		Type:        "avatar",
		AvatarID:    req.AvatarID,
		Script:      req.Script,
		VoiceID:     req.VoiceID,
		AspectRatio: req.AspectRatio,
		Resolution:  req.Resolution,
		Title:       req.Title,
	}

	var out heyGenCreateResponse
	if err := h.do(ctx, http.MethodPost, "/v3/videos", body, &out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", out.Error
	}
	if out.Data.VideoID == "" {
		return "", fmt.Errorf("heygen: empty video_id")
	}
	return out.Data.VideoID, nil
}

// Status interroge l'état d'une vidéo (GET /v3/videos/{id}).
func (h *HeyGen) Status(ctx context.Context, videoID string) (port.VideoResult, error) {
	var out heyGenGetResponse
	if err := h.do(ctx, http.MethodGet, "/v3/videos/"+videoID, nil, &out); err != nil {
		return port.VideoResult{}, err
	}
	if out.Error != nil {
		return port.VideoResult{}, out.Error
	}

	result := port.VideoResult{
		ID:     out.Data.ID,
		Status: port.VideoStatus(out.Data.Status),
	}
	if out.Data.VideoURL != nil {
		result.URL = *out.Data.VideoURL
	}
	if out.Data.FailureMessage != nil {
		result.Error = *out.Data.FailureMessage
	}
	return result, nil
}

func (h *HeyGen) do(ctx context.Context, method, path string, body, out any) error {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, h.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", h.apiKey)

	resp, err := h.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr heyGenGetResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Error != nil {
			return fmt.Errorf("heygen: %s (status %d)", apiErr.Error.Message, resp.StatusCode)
		}
		return fmt.Errorf("heygen: status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
