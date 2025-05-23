package models

import (
	"testing"
	"time"
)

func TestControl_GetID(t *testing.T) {
	control := Control{
		ID:   123,
		Name: "Test Control",
	}
	
	if id := control.GetID(); id != 123 {
		t.Errorf("GetID() = %d, want %d", id, 123)
	}
}

func TestClass_GetID(t *testing.T) {
	class := Class{
		ID:            456,
		Name:          "Test Class",
		OrderKey:      10,
		RadioControls: []Control{},
	}
	
	if id := class.GetID(); id != 456 {
		t.Errorf("GetID() = %d, want %d", id, 456)
	}
}

func TestClub_GetID(t *testing.T) {
	club := Club{
		ID:          789,
		Name:        "Test Club",
		CountryCode: "SWE",
	}
	
	if id := club.GetID(); id != 789 {
		t.Errorf("GetID() = %d, want %d", id, 789)
	}
}

func TestCompetitor_GetID(t *testing.T) {
	competitor := Competitor{
		ID:        101,
		Card:      12345,
		Name:      "Test Competitor",
		Status:    "1",
		StartTime: time.Now(),
		Splits:    []Split{},
		Club:      Club{ID: 1, Name: "Club"},
		Class:     Class{ID: 1, Name: "Class"},
	}
	
	if id := competitor.GetID(); id != 101 {
		t.Errorf("GetID() = %d, want %d", id, 101)
	}
}

func TestEntity_Interface(t *testing.T) {
	// Test that all types implement the Entity interface
	var entities []Entity = []Entity{
		Control{ID: 1},
		Class{ID: 2},
		Club{ID: 3},
		Competitor{ID: 4},
	}
	
	expectedIDs := []int{1, 2, 3, 4}
	
	for i, entity := range entities {
		if id := entity.GetID(); id != expectedIDs[i] {
			t.Errorf("Entity %d GetID() = %d, want %d", i, id, expectedIDs[i])
		}
	}
}