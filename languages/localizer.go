package languages

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed translations.json
var translationFile embed.FS

type translationMap map[string]map[string]string

type Languages struct {
	currentLanguage  string
	fallbackLanguage string
	activeTexts      map[string]string
}

func New(lang string) (*Languages, error) {
	lang = strings.ToLower(strings.TrimSpace(lang))

	if len(lang) < 2 {
		return nil, fmt.Errorf("invalid language: %q", lang)
	}

	loc := &Languages{
		currentLanguage:  lang[:2],
		fallbackLanguage: "en",
		activeTexts:      make(map[string]string),
	}

	if err := loc.loadTranslations(); err != nil {
		return nil, err
	}

	return loc, nil
}

func (loc *Languages) loadTranslations() error {
	data, err := translationFile.ReadFile("translations.json")
	if err != nil {
		return fmt.Errorf("failed to read translations.json: %w", err)
	}

	var allTranslations translationMap

	if err := json.Unmarshal(data, &allTranslations); err != nil {
		return fmt.Errorf("failed to parse translations JSON: %w", err)
	}

	for key, locales := range allTranslations {
		if text, exists := locales[loc.currentLanguage]; exists {
			loc.activeTexts[key] = text
			continue
		}

		if text, exists := locales[loc.fallbackLanguage]; exists {
			loc.activeTexts[key] = text
			continue
		}

		loc.activeTexts[key] = fmt.Sprintf("[%s]", key)
	}

	return nil
}

func (loc *Languages) Get(key string) string {
	if text, exists := loc.activeTexts[key]; exists {
		return text
	}

	return fmt.Sprintf("[%s]", key)
}

func (loc *Languages) CurrentLanguage() string {
	return loc.currentLanguage
}