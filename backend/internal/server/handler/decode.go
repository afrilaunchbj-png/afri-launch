package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"afrilaunch/backend/internal/server/apierror"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// decodeJSON décode et valide structurellement le corps JSON de la requête.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return apierror.Validation("Corps de requête invalide ou mal formé.")
	}
	return nil
}

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// validEmail vérifie un format d'email raisonnable.
func validEmail(s string) bool {
	return len(s) <= 254 && emailRe.MatchString(s)
}

// isValidationError indique si err est déjà une erreur API (à ne pas re-mapper).
func isValidationError(err error) bool {
	var apiErr *apierror.APIError
	return errors.As(err, &apiErr)
}
