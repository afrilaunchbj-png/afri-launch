package domain

import "time"

// Langues supportées par l'application (i18n).
const (
	LanguageFr = "fr"
	LanguageEn = "en"
)

// Thèmes de l'interface.
const (
	ThemeLight  = "light"
	ThemeDark   = "dark"
	ThemeSystem = "system"
)

// UserPreference regroupe les préférences d'un utilisateur (langue, thème).
// La langue est injectée dans les prompts du copilote : il répond toujours
// dans la langue du compte.
type UserPreference struct {
	UserID    string
	Language  string
	Theme     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SupportedLanguages liste les langues acceptées côté API.
func SupportedLanguages() []string { return []string{LanguageFr, LanguageEn} }

// SupportedThemes liste les thèmes acceptés côté API.
func SupportedThemes() []string { return []string{ThemeLight, ThemeDark, ThemeSystem} }

func Contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}
