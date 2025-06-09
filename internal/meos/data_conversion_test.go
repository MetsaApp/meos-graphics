package meos

import (
	"testing"
	"time"

	"meos-graphics/internal/models"
	"meos-graphics/internal/state"
)

func TestAdapter_ConvertControls(t *testing.T) {
	config := NewConfig()
	appState := state.New()
	adapter := NewAdapter(config, appState)

	mopComplete := MOPComplete{
		Controls: []MOPControl{
			{ID: "100", Name: "Start"},
			{ID: "101", Name: "Control 1"},
			{ID: "102", Name: "Control 2"},
			{ID: "200", Name: "Finish"},
		},
	}

	controls := adapter.convertControls(mopComplete, true)

	if len(controls) != 4 {
		t.Errorf("Number of controls = %d, want %d", len(controls), 4)
	}

	// Verify control conversion
	expectedControls := []struct {
		id   int
		name string
	}{
		{100, "Start"},
		{101, "Control 1"},
		{102, "Control 2"},
		{200, "Finish"},
	}

	for i, expected := range expectedControls {
		if i >= len(controls) {
			break
		}
		if controls[i].ID != expected.id {
			t.Errorf("Control[%d].ID = %d, want %d", i, controls[i].ID, expected.id)
		}
		if controls[i].Name != expected.name {
			t.Errorf("Control[%d].Name = %q, want %q", i, controls[i].Name, expected.name)
		}
	}
}

func TestAdapter_ConvertClasses(t *testing.T) {
	config := NewConfig()
	appState := state.New()
	adapter := NewAdapter(config, appState)

	mopComplete := MOPComplete{
		Classes: []MOPClass{
			{ID: "1", Name: "Men Elite", Order: "10", Radio: "101,102,103"},
			{ID: "2", Name: "Women Elite", Order: "20", Radio: "101,103"},
			{ID: "3", Name: "Junior", Order: "30", Radio: ""},
		},
	}

	classes := adapter.convertClasses(mopComplete, true)

	if len(classes) != 3 {
		t.Errorf("Number of classes = %d, want %d", len(classes), 3)
	}

	// Verify first class
	if len(classes) > 0 {
		class := classes[0]
		if class.ID != 1 {
			t.Errorf("Class ID = %d, want %d", class.ID, 1)
		}
		if class.Name != "Men Elite" {
			t.Errorf("Class Name = %q, want %q", class.Name, "Men Elite")
		}
		if class.OrderKey != 10 {
			t.Errorf("Class OrderKey = %d, want %d", class.OrderKey, 10)
		}
		if len(class.RadioControls) != 3 {
			t.Errorf("Number of radio controls = %d, want %d", len(class.RadioControls), 3)
		}
	}

	// Verify class with empty radio controls
	if len(classes) > 2 {
		class := classes[2]
		if len(class.RadioControls) != 0 {
			t.Errorf("Class with empty radio should have 0 controls, got %d", len(class.RadioControls))
		}
	}
}

func TestAdapter_ConvertClubs(t *testing.T) {
	config := NewConfig()
	appState := state.New()
	adapter := NewAdapter(config, appState)

	mopComplete := MOPComplete{
		Organizations: []MOPOrg{
			{ID: "1", Name: "Test Club 1", Nationality: "SWE"},
			{ID: "2", Name: "Test Club 2", Nationality: "NOR"},
			{ID: "3", Name: "Test Club 3", Nationality: ""},
		},
	}

	clubs := adapter.convertClubs(mopComplete, true)

	if len(clubs) != 3 {
		t.Errorf("Number of clubs = %d, want %d", len(clubs), 3)
	}

	// Verify club conversion
	expectedClubs := []struct {
		id          int
		name        string
		countryCode string
	}{
		{1, "Test Club 1", "SWE"},
		{2, "Test Club 2", "NOR"},
		{3, "Test Club 3", ""},
	}

	for i, expected := range expectedClubs {
		if i >= len(clubs) {
			break
		}
		if clubs[i].ID != expected.id {
			t.Errorf("Club[%d].ID = %d, want %d", i, clubs[i].ID, expected.id)
		}
		if clubs[i].Name != expected.name {
			t.Errorf("Club[%d].Name = %q, want %q", i, clubs[i].Name, expected.name)
		}
		if clubs[i].CountryCode != expected.countryCode {
			t.Errorf("Club[%d].CountryCode = %q, want %q", i, clubs[i].CountryCode, expected.countryCode)
		}
	}
}

func TestAdapter_ConvertCompetitors(t *testing.T) {
	config := NewConfig()
	appState := state.New()
	adapter := NewAdapter(config, appState)

	// Set up competition time
	appState.SetEvent(&models.Event{
		Start: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
	})

	mopComplete := MOPComplete{
		Competitors: []MOPCompetitor{
			{
				ID:   "1",
				Card: "12345",
				Base: MOPBase{
					Org:         "1",
					Class:       "1",
					Status:      "1",
					StartTime:   "3600", // 360 seconds = 6 minutes after zero time
					RunningTime: "834",  // 83.4 seconds
					Text:        "John Doe",
				},
				Radio: "101,302;102,584;103,723",
			},
			{
				ID:   "2",
				Card: "",
				Base: MOPBase{
					Org:         "2",
					Class:       "1",
					Status:      "0",
					StartTime:   "3900",
					RunningTime: "",
					Text:        "Jane Smith",
				},
				Radio: "",
			},
			{
				ID:   "3",
				Card: "12347",
				Base: MOPBase{
					Org:         "1",
					Class:       "2",
					Status:      "3", // DNF
					StartTime:   "4200",
					RunningTime: "",
					Text:        "Mike Johnson",
				},
				Radio: "101,350",
			},
		},
	}

	competitors := adapter.convertCompetitors(mopComplete, true)

	if len(competitors) != 3 {
		t.Errorf("Number of competitors = %d, want %d", len(competitors), 3)
	}

	// Test finished competitor
	if len(competitors) > 0 {
		comp := competitors[0]
		if comp.ID != 1 {
			t.Errorf("Competitor ID = %d, want %d", comp.ID, 1)
		}
		if comp.Name != "John Doe" {
			t.Errorf("Competitor Name = %q, want %q", comp.Name, "John Doe")
		}
		if comp.Card != 12345 {
			t.Errorf("Competitor Card = %d, want %d", comp.Card, 12345)
		}
		if comp.Status != "1" {
			t.Errorf("Competitor Status = %q, want %q", comp.Status, "1")
		}
		if comp.FinishTime == nil {
			t.Error("Finished competitor should have finish time")
		} else {
			runTime := comp.FinishTime.Sub(comp.StartTime)
			expectedRunTime := 83*time.Second + 400*time.Millisecond
			if runTime != expectedRunTime {
				t.Errorf("Run time = %v, want %v", runTime, expectedRunTime)
			}
		}
		if len(comp.Splits) != 3 {
			t.Errorf("Number of splits = %d, want %d", len(comp.Splits), 3)
		}
	}

	// Test not started competitor
	if len(competitors) > 1 {
		comp := competitors[1]
		if comp.Card != 0 {
			t.Errorf("Competitor with empty card should have Card = 0, got %d", comp.Card)
		}
		if comp.Status != "0" {
			t.Errorf("Competitor Status = %q, want %q", comp.Status, "0")
		}
		if comp.FinishTime != nil {
			t.Error("Not started competitor should not have finish time")
		}
		if len(comp.Splits) != 0 {
			t.Errorf("Not started competitor should have 0 splits, got %d", len(comp.Splits))
		}
	}

	// Test DNF competitor
	if len(competitors) > 2 {
		comp := competitors[2]
		if comp.Status != "3" {
			t.Errorf("Competitor Status = %q, want %q", comp.Status, "3")
		}
		if comp.FinishTime != nil {
			t.Error("DNF competitor should not have finish time")
		}
		if len(comp.Splits) != 1 {
			t.Errorf("DNF competitor should have 1 split, got %d", len(comp.Splits))
		}
	}
}

func TestDecisecondsToTimes(t *testing.T) {
	tests := []struct {
		name        string
		deciseconds int
		expected    time.Duration
	}{
		{
			name:        "zero",
			deciseconds: 0,
			expected:    0,
		},
		{
			name:        "one decisecond",
			deciseconds: 1,
			expected:    100 * time.Millisecond,
		},
		{
			name:        "ten deciseconds (one second)",
			deciseconds: 10,
			expected:    1 * time.Second,
		},
		{
			name:        "834 deciseconds (83.4 seconds)",
			deciseconds: 834,
			expected:    83*time.Second + 400*time.Millisecond,
		},
		{
			name:        "36000 deciseconds (one hour)",
			deciseconds: 36000,
			expected:    1 * time.Hour,
		},
		{
			name:        "complex time with deciseconds",
			deciseconds: 7545,
			expected:    754*time.Second + 500*time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decisecondsToTimes(tt.deciseconds)
			if result != tt.expected {
				t.Errorf("decisecondsToTimes(%d) = %v, want %v", tt.deciseconds, result, tt.expected)
			}
		})
	}
}

func TestRadioTimeParsing(t *testing.T) {
	config := NewConfig()
	appState := state.New()
	adapter := NewAdapter(config, appState)

	// Set competition start time
	competitionStart := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	appState.SetEvent(&models.Event{Start: competitionStart})

	mopComplete := MOPComplete{
		Competitors: []MOPCompetitor{
			{
				ID:   "1",
				Card: "12345",
				Base: MOPBase{
					Org:         "1",
					Class:       "1",
					Status:      "1",
					StartTime:   "3600", // 360 seconds after zero time
					RunningTime: "834",
					Text:        "Test Competitor",
				},
				Radio: "101,302;102,584;103,723",
			},
			{
				ID:   "2",
				Card: "12346",
				Base: MOPBase{
					Org:         "1",
					Class:       "1",
					Status:      "1",
					StartTime:   "3600",
					RunningTime: "900",
					Text:        "Invalid Radio",
				},
				Radio: "101;102,abc;103", // Invalid format
			},
			{
				ID:   "3",
				Card: "12347",
				Base: MOPBase{
					Org:         "1",
					Class:       "1",
					Status:      "1",
					StartTime:   "3600",
					RunningTime: "1000",
					Text:        "No Radio",
				},
				Radio: "", // Empty radio
			},
		},
	}

	competitors := adapter.convertCompetitors(mopComplete, true)

	// Test first competitor with valid radio times
	if len(competitors) > 0 {
		comp := competitors[0]
		if len(comp.Splits) != 3 {
			t.Errorf("Competitor 1 splits = %d, want 3", len(comp.Splits))
		}

		// Check split times are sorted
		for i := 1; i < len(comp.Splits); i++ {
			if comp.Splits[i].PassingTime.Before(comp.Splits[i-1].PassingTime) {
				t.Error("Splits are not sorted by passing time")
			}
		}

		// Check specific split times
		if len(comp.Splits) >= 3 {
			expectedElapsed := []time.Duration{
				30*time.Second + 200*time.Millisecond,
				58*time.Second + 400*time.Millisecond,
				72*time.Second + 300*time.Millisecond,
			}
			for i, expected := range expectedElapsed {
				elapsed := comp.Splits[i].PassingTime.Sub(comp.StartTime)
				if elapsed != expected {
					t.Errorf("Split %d elapsed time = %v, want %v", i, elapsed, expected)
				}
			}
		}
	}

	// Test second competitor with invalid radio format
	if len(competitors) > 1 {
		comp := competitors[1]
		// "101;102,abc;103" - only "102,abc" is valid format (though abc parses as 0)
		if len(comp.Splits) != 1 {
			t.Errorf("Competitor 2 with invalid radio should have 1 split (102,abc), got %d", len(comp.Splits))
		}
		if len(comp.Splits) > 0 && comp.Splits[0].Control.ID != 102 {
			t.Errorf("Split control ID = %d, want 102", comp.Splits[0].Control.ID)
		}
	}

	// Test third competitor with empty radio
	if len(competitors) > 2 {
		comp := competitors[2]
		if len(comp.Splits) != 0 {
			t.Errorf("Competitor 3 with empty radio should have 0 splits, got %d", len(comp.Splits))
		}
	}
}

func TestUpdateEntities(t *testing.T) {
	// Test with controls
	currentControls := []models.Control{
		{ID: 1, Name: "Control 1"},
		{ID: 2, Name: "Control 2"},
		{ID: 3, Name: "Control 3"},
	}

	updateControls := []models.Control{
		{ID: 2, Name: "Control 2 Updated"},
		{ID: 4, Name: "Control 4"},
	}

	// Test complete update (replace all)
	result := updateEntities(currentControls, updateControls, true)
	if len(result) != 2 {
		t.Errorf("Complete update: expected %d entities, got %d", 2, len(result))
	}

	// Test incremental update
	result = updateEntities(currentControls, updateControls, false)
	if len(result) != 4 {
		t.Errorf("Incremental update: expected %d entities, got %d", 4, len(result))
	}

	// Verify control 2 was updated
	var control2Found bool
	for _, ctrl := range result {
		if ctrl.ID == 2 && ctrl.Name == "Control 2 Updated" {
			control2Found = true
			break
		}
	}
	if !control2Found {
		t.Error("Control 2 was not updated correctly")
	}

	// Verify control 4 was added
	var control4Found bool
	for _, ctrl := range result {
		if ctrl.ID == 4 {
			control4Found = true
			break
		}
	}
	if !control4Found {
		t.Error("Control 4 was not added")
	}
}

func TestAdapter_ConvertWithInvalidIDs(t *testing.T) {
	config := NewConfig()
	appState := state.New()
	adapter := NewAdapter(config, appState)

	// Test controls with invalid IDs
	mopComplete := MOPComplete{
		Controls: []MOPControl{
			{ID: "abc", Name: "Invalid ID Control"},
			{ID: "123", Name: "Valid ID Control"},
			{ID: "", Name: "Empty ID Control"},
		},
		Classes: []MOPClass{
			{ID: "xyz", Name: "Invalid ID Class", Order: "abc"},
			{ID: "1", Name: "Valid ID Class", Order: "10"},
		},
		Organizations: []MOPOrg{
			{ID: "", Name: "Empty ID Club"},
			{ID: "5", Name: "Valid ID Club"},
		},
		Competitors: []MOPCompetitor{
			{ID: "invalid", Card: "abc", Base: MOPBase{Text: "Invalid IDs"}},
			{ID: "10", Card: "12345", Base: MOPBase{Text: "Valid IDs"}},
		},
	}

	controls := adapter.convertControls(mopComplete, true)
	classes := adapter.convertClasses(mopComplete, true)
	clubs := adapter.convertClubs(mopComplete, true)
	competitors := adapter.convertCompetitors(mopComplete, true)

	// Verify that invalid IDs are converted to 0
	if len(controls) > 0 && controls[0].ID != 0 {
		t.Errorf("Invalid control ID should convert to 0, got %d", controls[0].ID)
	}
	if len(classes) > 0 && classes[0].ID != 0 {
		t.Errorf("Invalid class ID should convert to 0, got %d", classes[0].ID)
	}
	if len(clubs) > 0 && clubs[0].ID != 0 {
		t.Errorf("Invalid club ID should convert to 0, got %d", clubs[0].ID)
	}
	if len(competitors) > 0 && competitors[0].ID != 0 {
		t.Errorf("Invalid competitor ID should convert to 0, got %d", competitors[0].ID)
	}

	// Verify valid IDs are converted correctly
	if len(controls) > 1 && controls[1].ID != 123 {
		t.Errorf("Valid control ID = %d, want %d", controls[1].ID, 123)
	}
	if len(classes) > 1 && classes[1].ID != 1 {
		t.Errorf("Valid class ID = %d, want %d", classes[1].ID, 1)
	}
}
