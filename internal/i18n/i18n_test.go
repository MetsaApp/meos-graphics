package i18n

import (
	"testing"
)

func TestGetInstance(t *testing.T) {
	// Test singleton pattern
	instance1 := GetInstance()
	instance2 := GetInstance()

	if instance1 != instance2 {
		t.Error("GetInstance should return the same instance")
	}

	// Test default language
	if instance1.GetLanguage() != English {
		t.Errorf("Default language should be English, got %s", instance1.GetLanguage())
	}
}

func TestSetGetLanguage(t *testing.T) {
	translator := GetInstance()

	// Set to Danish
	translator.SetLanguage(Danish)
	if translator.GetLanguage() != Danish {
		t.Errorf("Expected language to be Danish, got %s", translator.GetLanguage())
	}

	// Set back to English
	translator.SetLanguage(English)
	if translator.GetLanguage() != English {
		t.Errorf("Expected language to be English, got %s", translator.GetLanguage())
	}
}

func TestGetStatusDescription(t *testing.T) {
	translator := GetInstance()

	tests := []struct {
		name       string
		language   Language
		statusCode string
		expected   string
	}{
		// English tests
		{"English Unknown", English, "0", "Unknown"},
		{"English Approved", English, "1", "Approved"},
		{"English Miss Punch", English, "3", "Miss Punch"},
		{"English Not Finished", English, "4", "Not Finished"},
		{"English Disqualified", English, "5", "Disqualified"},
		{"English Max Time", English, "6", "Max. Time"},
		{"English Not Started", English, "20", "Not Started"},
		{"English Cancelled", English, "21", "Cancelled"},
		{"English Not Competing", English, "99", "Not Competing"},
		{"English Waiting to Start", English, "1000", "Waiting to Start"},
		{"English Running", English, "1001", "Running"},
		{"English Invalid Code", English, "999", "Unknown"},

		// Danish tests
		{"Danish Unknown", Danish, "0", "Ukendt"},
		{"Danish Approved", Danish, "1", "Godkendt"},
		{"Danish Miss Punch", Danish, "3", "Fejlstempel"},
		{"Danish Not Finished", Danish, "4", "Ikke Gennemført"},
		{"Danish Disqualified", Danish, "5", "Diskvalificeret"},
		{"Danish Max Time", Danish, "6", "Max. Tid"},
		{"Danish Not Started", Danish, "20", "Ikke Startet"},
		{"Danish Cancelled", Danish, "21", "Annulleret"},
		{"Danish Not Competing", Danish, "99", "Deltager Ikke"},
		{"Danish Waiting to Start", Danish, "1000", "Venter på Start"},
		{"Danish Running", Danish, "1001", "Løber"},
		{"Danish Invalid Code", Danish, "999", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator.SetLanguage(tt.language)
			result := translator.GetStatusDescription(tt.statusCode)
			if result != tt.expected {
				t.Errorf("GetStatusDescription(%s) = %s, want %s", tt.statusCode, result, tt.expected)
			}
		})
	}
}


func TestParseLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected Language
	}{
		{"en", English},
		{"da", Danish},
		{"fr", English}, // Unsupported language defaults to English
		{"", English},   // Empty string defaults to English
		{"EN", English}, // Case doesn't match, defaults to English
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLanguage(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLanguage(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLanguageString(t *testing.T) {
	tests := []struct {
		language Language
		expected string
	}{
		{English, "en"},
		{Danish, "da"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.language.String()
			if result != tt.expected {
				t.Errorf("Language.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	translator := GetInstance()

	// Test concurrent reads and writes
	done := make(chan bool)
	iterations := 100

	// Writer goroutine
	go func() {
		for i := 0; i < iterations; i++ {
			if i%2 == 0 {
				translator.SetLanguage(English)
			} else {
				translator.SetLanguage(Danish)
			}
		}
		done <- true
	}()

	// Reader goroutines
	for i := 0; i < 3; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				_ = translator.GetLanguage()
				_ = translator.GetStatusDescription("1")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 4; i++ {
		<-done
	}
}