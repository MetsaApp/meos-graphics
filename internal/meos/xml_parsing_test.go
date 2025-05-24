package meos

import (
	"encoding/xml"
	"testing"
	"time"

	"meos-graphics/internal/state"
	"meos-graphics/internal/testhelpers"
)

func TestXMLParsing_MOPComplete(t *testing.T) {
	var mopComplete MOPComplete
	err := xml.Unmarshal([]byte(testhelpers.MOPCompleteXML()), &mopComplete)
	if err != nil {
		t.Fatalf("Failed to unmarshal MOPComplete: %v", err)
	}

	// Test competition parsing
	if mopComplete.Competition.Name != "Test Competition" {
		t.Errorf("Competition name = %q, want %q", mopComplete.Competition.Name, "Test Competition")
	}
	if mopComplete.Competition.Organizer != "Test Organizer" {
		t.Errorf("Competition organizer = %q, want %q", mopComplete.Competition.Organizer, "Test Organizer")
	}
	if mopComplete.Competition.Date != "2024-01-01" {
		t.Errorf("Competition date = %q, want %q", mopComplete.Competition.Date, "2024-01-01")
	}
	if mopComplete.Competition.ZeroTime != "10:00:00" {
		t.Errorf("Competition zerotime = %q, want %q", mopComplete.Competition.ZeroTime, "10:00:00")
	}

	// Test controls parsing
	if len(mopComplete.Controls) != 5 {
		t.Errorf("Number of controls = %d, want %d", len(mopComplete.Controls), 5)
	}
	if len(mopComplete.Controls) > 0 && mopComplete.Controls[0].ID != "100" {
		t.Errorf("First control ID = %q, want %q", mopComplete.Controls[0].ID, "100")
	}

	// Test classes parsing
	if len(mopComplete.Classes) != 2 {
		t.Errorf("Number of classes = %d, want %d", len(mopComplete.Classes), 2)
	}
	if len(mopComplete.Classes) > 0 {
		if mopComplete.Classes[0].Radio != "101,102,103" {
			t.Errorf("Class radio controls = %q, want %q", mopComplete.Classes[0].Radio, "101,102,103")
		}
	}

	// Test organizations parsing
	if len(mopComplete.Organizations) != 2 {
		t.Errorf("Number of organizations = %d, want %d", len(mopComplete.Organizations), 2)
	}

	// Test competitors parsing
	if len(mopComplete.Competitors) != 3 {
		t.Errorf("Number of competitors = %d, want %d", len(mopComplete.Competitors), 3)
	}
	if len(mopComplete.Competitors) > 0 {
		comp := mopComplete.Competitors[0]
		if comp.ID != "1" {
			t.Errorf("Competitor ID = %q, want %q", comp.ID, "1")
		}
		if comp.Card != "12345" {
			t.Errorf("Competitor card = %q, want %q", comp.Card, "12345")
		}
		if comp.Base.Status != "1" {
			t.Errorf("Competitor status = %q, want %q", comp.Base.Status, "1")
		}
		if comp.Radio != "101,302;102,584;103,723" {
			t.Errorf("Competitor radio = %q, want %q", comp.Radio, "101,302;102,584;103,723")
		}
	}
}

func TestXMLParsing_MOPDiff(t *testing.T) {
	var mopDiff MOPDiff
	err := xml.Unmarshal([]byte(testhelpers.MOPDiffXML()), &mopDiff)
	if err != nil {
		t.Fatalf("Failed to unmarshal MOPDiff: %v", err)
	}

	if mopDiff.NextDifference != "def456" {
		t.Errorf("NextDifference = %q, want %q", mopDiff.NextDifference, "def456")
	}

	if len(mopDiff.Competitors) != 2 {
		t.Errorf("Number of competitors = %d, want %d", len(mopDiff.Competitors), 2)
	}

	// Competition should be nil in this diff
	if mopDiff.Competition != nil {
		t.Error("Competition should be nil in MOPDiff")
	}
}

func TestXMLParsing_InvalidXML(t *testing.T) {
	var mopComplete MOPComplete
	err := xml.Unmarshal([]byte(testhelpers.InvalidXML()), &mopComplete)
	if err == nil {
		t.Error("Expected error parsing invalid XML, got nil")
	}
}

func TestXMLParsing_EmptyMOPComplete(t *testing.T) {
	var mopComplete MOPComplete
	err := xml.Unmarshal([]byte(testhelpers.EmptyMOPCompleteXML()), &mopComplete)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty MOPComplete: %v", err)
	}

	if mopComplete.Competition.Name != "Empty Competition" {
		t.Errorf("Competition name = %q, want %q", mopComplete.Competition.Name, "Empty Competition")
	}
	if len(mopComplete.Controls) != 0 {
		t.Errorf("Number of controls = %d, want %d", len(mopComplete.Controls), 0)
	}
	if len(mopComplete.Classes) != 0 {
		t.Errorf("Number of classes = %d, want %d", len(mopComplete.Classes), 0)
	}
	if len(mopComplete.Competitors) != 0 {
		t.Errorf("Number of competitors = %d, want %d", len(mopComplete.Competitors), 0)
	}
}

func TestXMLParsing_MissingOptionalFields(t *testing.T) {
	var mopComplete MOPComplete
	err := xml.Unmarshal([]byte(testhelpers.MOPCompleteWithMissingFieldsXML()), &mopComplete)
	if err != nil {
		t.Fatalf("Failed to unmarshal MOPComplete with missing fields: %v", err)
	}

	// Competition should have empty organizer
	if mopComplete.Competition.Organizer != "" {
		t.Errorf("Competition organizer = %q, want empty", mopComplete.Competition.Organizer)
	}

	// Class should have empty radio
	if len(mopComplete.Classes) > 0 && mopComplete.Classes[0].Radio != "" {
		t.Errorf("Class radio = %q, want empty", mopComplete.Classes[0].Radio)
	}

	// Organization should have empty nationality
	if len(mopComplete.Organizations) > 0 && mopComplete.Organizations[0].Nationality != "" {
		t.Errorf("Organization nationality = %q, want empty", mopComplete.Organizations[0].Nationality)
	}

	// Competitor should have empty card
	if len(mopComplete.Competitors) > 0 && mopComplete.Competitors[0].Card != "" {
		t.Errorf("Competitor card = %q, want empty", mopComplete.Competitors[0].Card)
	}
}

func TestMOPCompetition_Time(t *testing.T) {
	tests := []struct {
		name     string
		comp     MOPCompetition
		expected time.Time
	}{
		{
			name: "valid date and time",
			comp: MOPCompetition{
				Date:     "2024-01-01",
				ZeroTime: "10:00:00",
			},
			expected: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "invalid date format",
			comp: MOPCompetition{
				Date:     "01-01-2024",
				ZeroTime: "10:00:00",
			},
			expected: time.Time{}, // Zero time
		},
		{
			name: "invalid time format",
			comp: MOPCompetition{
				Date:     "2024-01-01",
				ZeroTime: "10:00",
			},
			expected: time.Time{}, // Zero time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.comp.Time()
			if !got.Equal(tt.expected) {
				t.Errorf("MOPCompetition.Time() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"-123", -123},
		{"abc", 0},
		{"", 0},
		{"123abc", 123},
		{"1.23", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseInt(tt.input)
			if got != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAdapter_ProcessData_MOPComplete(t *testing.T) {
	config := &Config{
		Hostname:     "localhost",
		Port:         2009,
		PortStr:      "2009",
		PollInterval: 1 * time.Second,
	}
	appState := state.New()
	adapter := NewAdapter(config, appState)

	// Process MOPComplete data
	updated, err := adapter.processData([]byte(testhelpers.MOPCompleteXML()))
	if err != nil {
		t.Fatalf("processData failed: %v", err)
	}
	if !updated {
		t.Error("Expected data to be updated")
	}

	// Verify event was set
	event := appState.GetEvent()
	if event == nil {
		t.Fatal("Event should not be nil")
	}
	if event.Name != "Test Competition" {
		t.Errorf("Event name = %q, want %q", event.Name, "Test Competition")
	}

	// Verify controls were added
	controls := appState.GetControls()
	if len(controls) != 5 {
		t.Errorf("Number of controls = %d, want %d", len(controls), 5)
	}

	// Verify classes were added
	classes := appState.GetClasses()
	if len(classes) != 2 {
		t.Errorf("Number of classes = %d, want %d", len(classes), 2)
	}
	if len(classes) > 0 && len(classes[0].RadioControls) != 3 {
		t.Errorf("Number of radio controls = %d, want %d", len(classes[0].RadioControls), 3)
	}

	// Verify clubs were added
	clubs := appState.GetClubs()
	if len(clubs) != 2 {
		t.Errorf("Number of clubs = %d, want %d", len(clubs), 2)
	}

	// Verify competitors were added
	competitors := appState.GetCompetitors()
	if len(competitors) != 3 {
		t.Errorf("Number of competitors = %d, want %d", len(competitors), 3)
	}

	// Check first competitor details
	if len(competitors) > 0 {
		comp := competitors[0]
		if comp.Name != "John Doe" {
			t.Errorf("Competitor name = %q, want %q", comp.Name, "John Doe")
		}
		if comp.Status != "1" {
			t.Errorf("Competitor status = %q, want %q", comp.Status, "1")
		}
		if len(comp.Splits) != 3 {
			t.Errorf("Number of splits = %d, want %d", len(comp.Splits), 3)
		}
		if comp.FinishTime == nil {
			t.Error("Competitor finish time should not be nil")
		}
	}
}

func TestAdapter_ProcessData_MOPDiff(t *testing.T) {
	config := &Config{
		Hostname:     "localhost",
		Port:         2009,
		PortStr:      "2009",
		PollInterval: 1 * time.Second,
	}
	appState := state.New()
	adapter := NewAdapter(config, appState)

	// First process MOPComplete to set up initial state
	_, err := adapter.processData([]byte(testhelpers.MOPCompleteXML()))
	if err != nil {
		t.Fatalf("Initial processData failed: %v", err)
	}

	// Then process MOPDiff
	updated, err := adapter.processData([]byte(testhelpers.MOPDiffXML()))
	if err != nil {
		t.Fatalf("processData for diff failed: %v", err)
	}
	if !updated {
		t.Error("Expected data to be updated")
	}

	// Verify competitor 2 was updated
	competitors := appState.GetCompetitors()
	var comp2Found bool
	for _, comp := range competitors {
		if comp.ID == 2 {
			comp2Found = true
			if comp.Status != "1" {
				t.Errorf("Competitor 2 status = %q, want %q", comp.Status, "1")
			}
			if comp.FinishTime == nil {
				t.Error("Competitor 2 finish time should not be nil after update")
			}
			break
		}
	}
	if !comp2Found {
		t.Error("Competitor 2 not found after diff update")
	}

	// Verify new competitor was added
	if len(competitors) != 4 {
		t.Errorf("Number of competitors = %d, want %d", len(competitors), 4)
	}
}

func TestAdapter_ProcessData_NoUpdate(t *testing.T) {
	config := &Config{
		Hostname:     "localhost",
		Port:         2009,
		PortStr:      "2009",
		PollInterval: 1 * time.Second,
	}
	appState := state.New()
	adapter := NewAdapter(config, appState)

	// Process same data twice
	adapter.processData([]byte(testhelpers.MOPCompleteXML()))

	// Second call should not update (same nextdifference)
	updated, err := adapter.processData([]byte(testhelpers.MOPCompleteXML()))
	if err != nil {
		t.Fatalf("processData failed: %v", err)
	}
	if updated {
		t.Error("Expected no update for same nextdifference")
	}
}

func TestAdapter_ProcessData_InvalidXML(t *testing.T) {
	config := &Config{
		Hostname:     "localhost",
		Port:         2009,
		PortStr:      "2009",
		PollInterval: 1 * time.Second,
	}
	appState := state.New()
	adapter := NewAdapter(config, appState)

	_, err := adapter.processData([]byte(testhelpers.InvalidXML()))
	if err == nil {
		t.Error("Expected error for invalid XML")
	}
}

func TestAdapter_ProcessData_UnknownRootElement(t *testing.T) {
	config := &Config{
		Hostname:     "localhost",
		Port:         2009,
		PortStr:      "2009",
		PollInterval: 1 * time.Second,
	}
	appState := state.New()
	adapter := NewAdapter(config, appState)

	unknownXML := `<?xml version="1.0" encoding="UTF-8"?>
<UnknownElement nextdifference="abc123">
    <data>test</data>
</UnknownElement>`

	_, err := adapter.processData([]byte(unknownXML))
	if err == nil {
		t.Error("Expected error for unknown root element")
	}
	if err != nil && err.Error() != "unknown XML root element: UnknownElement" {
		t.Errorf("Unexpected error message: %v", err)
	}
}
