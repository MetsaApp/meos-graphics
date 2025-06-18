package i18n

import (
	"sync"
)

// Language represents a supported language
type Language string

const (
	// English language
	English Language = "en"
	// Danish language
	Danish Language = "da"
)

// Translator handles translations for status codes and strings
type Translator struct {
	mu       sync.RWMutex
	language Language
}

// statusDescriptions maps status codes to full descriptions
var statusDescriptions = map[Language]map[string]string{
	English: {
		"0":    "Unknown",
		"1":    "Approved",
		"3":    "Miss Punch",
		"4":    "Not Finished",
		"5":    "Disqualified",
		"6":    "Max. Time",
		"20":   "Not Started",
		"21":   "Cancelled",
		"99":   "Not Competing",
		"1000": "Waiting to Start",
		"1001": "Running",
	},
	Danish: {
		"0":    "Ukendt",
		"1":    "Godkendt",
		"3":    "Fejlstempel",
		"4":    "Ikke Gennemført",
		"5":    "Diskvalificeret",
		"6":    "Max. Tid",
		"20":   "Ikke Startet",
		"21":   "Annulleret",
		"99":   "Deltager Ikke",
		"1000": "Venter på Start",
		"1001": "Løber",
	},
}

var (
	instance *Translator
	once     sync.Once
)

// GetInstance returns the singleton translator instance
func GetInstance() *Translator {
	once.Do(func() {
		instance = &Translator{
			language: English, // Default to English
		}
	})
	return instance
}

// SetLanguage sets the language for translations
func (t *Translator) SetLanguage(lang Language) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.language = lang
}

// GetLanguage returns the current language
func (t *Translator) GetLanguage() Language {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.language
}

// GetStatusDescription returns the full status description for a given status code
func (t *Translator) GetStatusDescription(statusCode string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if translations, ok := statusDescriptions[t.language]; ok {
		if desc, ok := translations[statusCode]; ok {
			return desc
		}
	}

	// Fallback to English if translation not found
	if t.language != English {
		if translations, ok := statusDescriptions[English]; ok {
			if desc, ok := translations[statusCode]; ok {
				return desc
			}
		}
	}

	return "Unknown"
}

// ParseLanguage converts a string to Language type
func ParseLanguage(lang string) Language {
	switch lang {
	case "da":
		return Danish
	case "en":
		return English
	default:
		return English // Default to English
	}
}

// String returns the string representation of a Language
func (l Language) String() string {
	return string(l)
}

