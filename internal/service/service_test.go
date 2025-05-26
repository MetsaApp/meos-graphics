package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"meos-graphics/internal/models"
	"meos-graphics/internal/state"
)

func TestGetResults_SharedPositions(t *testing.T) {
	appState := state.New()
	svc := New(appState)

	// Create test data with tied competitors
	now := time.Now()
	startTime := now.Add(-time.Hour)

	testClass := models.Class{
		ID:            1,
		Name:          "Elite",
		RadioControls: []models.Control{},
	}

	competitors := []models.Competitor{
		{
			ID:        1,
			Name:      "Runner A",
			Card:      101,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(45*time.Minute + 23*time.Second + 500*time.Millisecond)
				return &t
			}(),
			Status: "1", // OK
			Class:  testClass,
			Club:   models.Club{Name: "Club A"},
		},
		{
			ID:        2,
			Name:      "Runner B",
			Card:      102,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(45*time.Minute + 23*time.Second + 500*time.Millisecond)
				return &t
			}(), // Same time as Runner A
			Status: "1", // OK
			Class:  testClass,
			Club:   models.Club{Name: "Club B"},
		},
		{
			ID:        3,
			Name:      "Runner C",
			Card:      103,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(46*time.Minute + 12*time.Second + 300*time.Millisecond)
				return &t
			}(),
			Status: "1", // OK
			Class:  testClass,
			Club:   models.Club{Name: "Club C"},
		},
		{
			ID:        4,
			Name:      "Runner D",
			Card:      104,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(47*time.Minute + 1*time.Second + 900*time.Millisecond)
				return &t
			}(),
			Status: "1", // OK
			Class:  testClass,
			Club:   models.Club{Name: "Club D"},
		},
	}

	// Update state with test data
	appState.UpdateFromMeOS(nil, []models.Control{}, []models.Class{testClass}, []models.Club{}, competitors)

	result, err := svc.GetResults(1)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check that runners with identical times share position 1
	assert.Equal(t, 1, result[0].Position)
	assert.Equal(t, "Runner A", result[0].Name)
	assert.Equal(t, 1, result[1].Position)
	assert.Equal(t, "Runner B", result[1].Name)

	// Check that the next runner gets position 3 (skipping position 2)
	assert.Equal(t, 3, result[2].Position)
	assert.Equal(t, "Runner C", result[2].Name)

	// Check that the last runner gets position 4
	assert.Equal(t, 4, result[3].Position)
	assert.Equal(t, "Runner D", result[3].Name)
}

func TestGetResults_MultipleSharedPositions(t *testing.T) {
	appState := state.New()
	svc := New(appState)

	// Create test data with multiple groups of tied competitors
	now := time.Now()
	startTime := now.Add(-time.Hour)

	testClass := models.Class{
		ID:            1,
		Name:          "Elite",
		RadioControls: []models.Control{},
	}

	competitors := []models.Competitor{
		{
			ID:        1,
			Name:      "Runner A",
			Card:      101,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(45*time.Minute + 23*time.Second + 500*time.Millisecond)
				return &t
			}(),
			Status: "1",
			Class:  testClass,
			Club:   models.Club{Name: "Club A"},
		},
		{
			ID:        2,
			Name:      "Runner B",
			Card:      102,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(45*time.Minute + 51*time.Second + 200*time.Millisecond)
				return &t
			}(),
			Status: "1",
			Class:  testClass,
			Club:   models.Club{Name: "Club B"},
		},
		{
			ID:        3,
			Name:      "Runner C",
			Card:      103,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(46*time.Minute + 12*time.Second + 300*time.Millisecond)
				return &t
			}(),
			Status: "1",
			Class:  testClass,
			Club:   models.Club{Name: "Club C"},
		},
		{
			ID:        4,
			Name:      "Runner D",
			Card:      104,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(46*time.Minute + 45*time.Second + 800*time.Millisecond)
				return &t
			}(),
			Status: "1",
			Class:  testClass,
			Club:   models.Club{Name: "Club D"},
		},
		{
			ID:        5,
			Name:      "Runner E",
			Card:      105,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(47*time.Minute + 1*time.Second + 900*time.Millisecond)
				return &t
			}(),
			Status: "1",
			Class:  testClass,
			Club:   models.Club{Name: "Club E"},
		},
		{
			ID:        6,
			Name:      "Runner F",
			Card:      106,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(47*time.Minute + 23*time.Second + 400*time.Millisecond)
				return &t
			}(),
			Status: "1",
			Class:  testClass,
			Club:   models.Club{Name: "Club F"},
		},
		{
			ID:        7,
			Name:      "Runner G",
			Card:      107,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(47*time.Minute + 23*time.Second + 400*time.Millisecond)
				return &t
			}(), // Same as F
			Status: "1",
			Class:  testClass,
			Club:   models.Club{Name: "Club G"},
		},
		{
			ID:        8,
			Name:      "Runner H",
			Card:      108,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(47*time.Minute + 23*time.Second + 400*time.Millisecond)
				return &t
			}(), // Same as F and G
			Status: "1",
			Class:  testClass,
			Club:   models.Club{Name: "Club H"},
		},
		{
			ID:        9,
			Name:      "Runner I",
			Card:      109,
			StartTime: startTime,
			FinishTime: func() *time.Time {
				t := startTime.Add(48*time.Minute + 15*time.Second + 700*time.Millisecond)
				return &t
			}(),
			Status: "1",
			Class:  testClass,
			Club:   models.Club{Name: "Club I"},
		},
	}

	appState.UpdateFromMeOS(nil, []models.Control{}, []models.Class{testClass}, []models.Club{}, competitors)

	result, err := svc.GetResults(1)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the expected positions from the issue example
	assert.Equal(t, 1, result[0].Position)
	assert.Equal(t, 2, result[1].Position)
	assert.Equal(t, 3, result[2].Position)
	assert.Equal(t, 4, result[3].Position)
	assert.Equal(t, 5, result[4].Position)
	assert.Equal(t, 6, result[5].Position) // F
	assert.Equal(t, 6, result[6].Position) // G (tied with F)
	assert.Equal(t, 6, result[7].Position) // H (tied with F and G)
	assert.Equal(t, 9, result[8].Position) // I (position 9, skipping 7 and 8)
}

func TestGetSplits_SharedPositions(t *testing.T) {
	appState := state.New()
	svc := New(appState)

	// Create test data with tied split times
	now := time.Now()
	startTime := now.Add(-time.Hour)

	testClass := models.Class{
		ID:            1,
		Name:          "Elite",
		RadioControls: []models.Control{{ID: 1, Name: "Control 1"}},
	}

	control1 := models.Control{ID: 1, Name: "Control 1"}

	competitors := []models.Competitor{
		{
			ID:        1,
			Name:      "Runner A",
			Card:      101,
			StartTime: startTime,
			Status:    "2", // Running
			Class:     testClass,
			Club:      models.Club{Name: "Club A"},
			Splits: []models.Split{
				{
					Control:     control1,
					PassingTime: startTime.Add(10 * time.Minute), // 10:00.0 elapsed
				},
			},
		},
		{
			ID:        2,
			Name:      "Runner B",
			Card:      102,
			StartTime: startTime,
			Status:    "2", // Running
			Class:     testClass,
			Club:      models.Club{Name: "Club B"},
			Splits: []models.Split{
				{
					Control:     control1,
					PassingTime: startTime.Add(10 * time.Minute), // 10:00.0 elapsed (same as A)
				},
			},
		},
		{
			ID:        3,
			Name:      "Runner C",
			Card:      103,
			StartTime: startTime,
			Status:    "2", // Running
			Class:     testClass,
			Club:      models.Club{Name: "Club C"},
			Splits: []models.Split{
				{
					Control:     control1,
					PassingTime: startTime.Add(11*time.Minute + 30*time.Second), // 11:30.0 elapsed
				},
			},
		},
	}

	appState.UpdateFromMeOS(nil, []models.Control{}, []models.Class{testClass}, []models.Club{}, competitors)

	result, err := svc.GetSplits(1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Splits, 2) // Control 1 and Finish

	// Check split standings at control 1 (first in list)
	control1Standings := result.Splits[0].Standings
	assert.Equal(t, "Control 1", result.Splits[0].ControlName)
	assert.Len(t, control1Standings, 3)

	// Runners A and B should share position 1
	assert.Equal(t, 1, control1Standings[0].Position)
	assert.Equal(t, 1, control1Standings[1].Position)

	// Runner C should have position 3
	assert.Equal(t, 3, control1Standings[2].Position)
}

func TestGetResults_TiedCompetitorsAlphabeticalOrder(t *testing.T) {
	appState := state.New()
	svc := New(appState)

	// Create test data with tied competitors to check alphabetical ordering
	now := time.Now()
	startTime := now.Add(-time.Hour)

	testClass := models.Class{
		ID:            1,
		Name:          "Elite",
		RadioControls: []models.Control{},
	}

	competitors := []models.Competitor{
		{
			ID:         1,
			Name:       "Zebra Runner",
			Card:       101,
			StartTime:  startTime,
			FinishTime: func() *time.Time { t := startTime.Add(45 * time.Minute); return &t }(),
			Status:     "1",
			Class:      testClass,
			Club:       models.Club{Name: "Club Z"},
		},
		{
			ID:         2,
			Name:       "Alpha Runner",
			Card:       102,
			StartTime:  startTime,
			FinishTime: func() *time.Time { t := startTime.Add(45 * time.Minute); return &t }(), // Same time
			Status:     "1",
			Class:      testClass,
			Club:       models.Club{Name: "Club A"},
		},
		{
			ID:         3,
			Name:       "Charlie Runner",
			Card:       103,
			StartTime:  startTime,
			FinishTime: func() *time.Time { t := startTime.Add(45 * time.Minute); return &t }(), // Same time
			Status:     "1",
			Class:      testClass,
			Club:       models.Club{Name: "Club C"},
		},
		{
			ID:         4,
			Name:       "Beta Runner",
			Card:       104,
			StartTime:  startTime,
			FinishTime: func() *time.Time { t := startTime.Add(45 * time.Minute); return &t }(), // Same time
			Status:     "1",
			Class:      testClass,
			Club:       models.Club{Name: "Club B"},
		},
	}

	appState.UpdateFromMeOS(nil, []models.Control{}, []models.Class{testClass}, []models.Club{}, competitors)

	result, err := svc.GetResults(1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 4)

	// All should have position 1
	for i := 0; i < 4; i++ {
		assert.Equal(t, 1, result[i].Position)
	}

	// Check alphabetical order by name
	assert.Equal(t, "Alpha Runner", result[0].Name)
	assert.Equal(t, "Beta Runner", result[1].Name)
	assert.Equal(t, "Charlie Runner", result[2].Name)
	assert.Equal(t, "Zebra Runner", result[3].Name)
}
