// Package apierror définit le format d'erreur unique de l'API, conforme
// à la RFC 9457 (Problem Details). Toute erreur HTTP renvoyée par
// l'application doit passer par ce type.
package apierror

// APIError est la représentation sérialisée d'une erreur (Problem Details).
type APIError struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail,omitempty"`
	Instance string       `json:"instance,omitempty"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// FieldError décrit une erreur liée à un champ spécifique (validation).
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implémente l'interface error.
func (e *APIError) Error() string { return e.Title }

// Validation renvoie une erreur 422 avec des erreurs de champ optionnelles.
func Validation(detail string, fields ...FieldError) *APIError {
	return &APIError{
		Type:   "https://afrilaunch.example/errors/validation",
		Title:  "Validation Error",
		Status: 422,
		Detail: detail,
		Errors: fields,
	}
}

// Unauthorized renvoie une erreur 401.
func Unauthorized(detail string) *APIError {
	return &APIError{Type: "https://afrilaunch.example/errors/unauthorized", Title: "Unauthorized", Status: 401, Detail: detail}
}

// Forbidden renvoie une erreur 403.
func Forbidden(detail string) *APIError {
	return &APIError{Type: "https://afrilaunch.example/errors/forbidden", Title: "Forbidden", Status: 403, Detail: detail}
}

// NotFound renvoie une erreur 404.
func NotFound(detail string) *APIError {
	return &APIError{Type: "https://afrilaunch.example/errors/not-found", Title: "Not Found", Status: 404, Detail: detail}
}

// Conflict renvoie une erreur 409.
func Conflict(detail string) *APIError {
	return &APIError{Type: "https://afrilaunch.example/errors/conflict", Title: "Conflict", Status: 409, Detail: detail}
}

// Business renvoie une erreur métier (422 par défaut, orientée action).
func Business(detail string, fields ...FieldError) *APIError {
	return &APIError{Type: "https://afrilaunch.example/errors/business", Title: "Business Error", Status: 422, Detail: detail, Errors: fields}
}

// Internal renvoie une erreur 500 sans détail technique exposé au client.
func Internal() *APIError {
	return &APIError{Type: "https://afrilaunch.example/errors/internal", Title: "Internal Server Error", Status: 500}
}

// TooManyRequests renvoie une erreur 429 (rate limiting).
func TooManyRequests(detail string) *APIError {
	return &APIError{Type: "https://afrilaunch.example/errors/rate-limit", Title: "Too Many Requests", Status: 429, Detail: detail}
}
