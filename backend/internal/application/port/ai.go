package port

import "context"

// LLMMessage est un message d'une conversation.
type LLMMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// LLMRequest est une demande de complétion.
type LLMRequest struct {
	Model       string
	System      string
	Messages    []LLMMessage
	MaxTokens   int
	Temperature *float64
	JSONMode    bool
}

// LLMResponse est le résultat d'une complétion, avec les métriques
// d'observabilité (master.md §25).
type LLMResponse struct {
	Content      string
	Model        string
	InputTokens  int
	OutputTokens int
}

// LLMProvider est le port d'accès aux modèles de langage (OpenAI).
type LLMProvider interface {
	Complete(ctx context.Context, req LLMRequest) (LLMResponse, error)
}

// ImageRequest est une demande de génération d'image.
type ImageRequest struct {
	Model   string
	Prompt  string
	Size    string // ex. 1024x1024
	Quality string
}

// Image est une image générée (b64 ou URL).
type Image struct {
	B64JSON string
	URL     string
}

// ImageProvider est le port d'accès aux modèles image (OpenAI).
type ImageProvider interface {
	Generate(ctx context.Context, req ImageRequest) (Image, error)
}

// VideoRequest est une demande de génération de vidéo avatar (HeyGen).
type VideoRequest struct {
	AvatarID    string
	Script      string
	VoiceID     string
	AspectRatio string // 16:9, 9:16, 1:1, auto
	Resolution  string // 720p, 1080p
	Title       string
}

// VideoStatus est l'état d'un job vidéo.
type VideoStatus string

const (
	VideoPending    VideoStatus = "pending"
	VideoProcessing VideoStatus = "processing"
	VideoCompleted  VideoStatus = "completed"
	VideoFailed     VideoStatus = "failed"
)

// VideoResult est l'état courant d'une vidéo.
type VideoResult struct {
	ID     string
	Status VideoStatus
	URL    string
	Error  string
}

// VideoProvider est le port d'accès à la génération vidéo (HeyGen).
type VideoProvider interface {
	Submit(ctx context.Context, req VideoRequest) (videoID string, err error)
	Status(ctx context.Context, videoID string) (VideoResult, error)
}
