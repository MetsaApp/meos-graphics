package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"meos-graphics/internal/models"
	"meos-graphics/internal/service"
	"meos-graphics/internal/state"
	"meos-graphics/internal/testhelpers"
)

func setupTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/classes", h.GetClasses)
	router.GET("/classes/:classId/startlist", h.GetStartList)
	router.GET("/classes/:classId/results", h.GetResults)
	router.GET("/classes/:classId/splits", h.GetSplits)
	return router
}

func TestHandler_GetClasses(t *testing.T) {
	// Set up state with test data
	s := state.New()
	s.Lock()
	s.Classes = []models.Class{
		testhelpers.CreateTestClass(1, "Men Elite", 10),
		testhelpers.CreateTestClass(2, "Women Elite", 20),
		testhelpers.CreateTestClass(3, "Junior Men", 15),
	}
	s.Unlock()

	h := New(s)
	router := setupTestRouter(h)

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/classes", nil)
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var classes []service.ClassInfo
	err := json.Unmarshal(w.Body.Bytes(), &classes)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(classes) != 3 {
		t.Errorf("Number of classes = %d, want 3", len(classes))
	}

	// Verify sorting by OrderKey
	if len(classes) >= 3 {
		if classes[0].OrderKey != 10 || classes[1].OrderKey != 15 || classes[2].OrderKey != 20 {
			t.Error("Classes not sorted correctly by OrderKey")
		}
	}
}

func TestHandler_GetClasses_Empty(t *testing.T) {
	s := state.New()
	h := New(s)
	router := setupTestRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/classes", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var classes []service.ClassInfo
	err := json.Unmarshal(w.Body.Bytes(), &classes)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(classes) != 0 {
		t.Errorf("Number of classes = %d, want 0", len(classes))
	}
}

func TestHandler_GetStartList(t *testing.T) {
	// Set up state with test data
	s := state.New()
	club1 := testhelpers.CreateTestClub(1, "Test Club 1", "SWE")
	club2 := testhelpers.CreateTestClub(2, "Test Club 2", "NOR")
	class := testhelpers.CreateTestClass(1, "Men Elite", 10)

	startTime1 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	startTime2 := time.Date(2024, 1, 1, 11, 2, 0, 0, time.UTC)
	startTime3 := time.Date(2024, 1, 1, 11, 4, 0, 0, time.UTC)

	s.Lock()
	s.Competitors = []models.Competitor{
		{
			ID:        2,
			Name:      "Jane Smith",
			Card:      200,
			Club:      club2,
			Class:     class,
			Status:    "0",
			StartTime: startTime2,
		},
		{
			ID:        1,
			Name:      "John Doe",
			Card:      100,
			Club:      club1,
			Class:     class,
			Status:    "0",
			StartTime: startTime1,
		},
		{
			ID:        3,
			Name:      "Mike Johnson",
			Card:      300,
			Club:      club1,
			Class:     class,
			Status:    "0",
			StartTime: startTime3,
		},
	}
	s.Unlock()

	h := New(s)
	router := setupTestRouter(h)

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/classes/1/startlist", nil)
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var startList []service.StartListEntry
	err := json.Unmarshal(w.Body.Bytes(), &startList)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(startList) != 3 {
		t.Errorf("Number of start list entries = %d, want 3", len(startList))
	}

	// Verify sorting by start time
	if len(startList) >= 3 {
		if startList[0].Name != "John Doe" {
			t.Errorf("First starter = %q, want %q", startList[0].Name, "John Doe")
		}
		// Check start time formatting
		if startList[0].StartTime != "11:00" {
			t.Errorf("First start time = %q, want %q", startList[0].StartTime, "11:00")
		}
	}
}

func TestHandler_GetStartList_InvalidClassID(t *testing.T) {
	s := state.New()
	h := New(s)
	router := setupTestRouter(h)

	// Test with invalid class ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/classes/invalid/startlist", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Invalid class ID" {
		t.Errorf("Error message = %q, want %q", response["error"], "Invalid class ID")
	}
}

func TestHandler_GetResults(t *testing.T) {
	// Set up state with test data
	s := state.New()
	club1 := testhelpers.CreateTestClub(1, "Test Club 1", "SWE")
	club2 := testhelpers.CreateTestClub(2, "Test Club 2", "NOR")

	control1 := testhelpers.CreateTestControl(101, "Control 1")
	control2 := testhelpers.CreateTestControl(102, "Control 2")
	class := testhelpers.CreateTestClass(1, "Men Elite", 10, control1, control2)

	startTime := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)

	// Create competitors with different statuses
	comp1 := testhelpers.CreateFinishedCompetitor(1, "John Doe", club1, class, 834) // 83.4 seconds
	comp1.Splits = []models.Split{
		testhelpers.CreateTestSplit(control1, 302, startTime), // 30.2 seconds
		testhelpers.CreateTestSplit(control2, 584, startTime), // 58.4 seconds
	}

	comp2 := testhelpers.CreateFinishedCompetitor(2, "Jane Smith", club2, class, 912) // 91.2 seconds
	comp2.Splits = []models.Split{
		testhelpers.CreateTestSplit(control1, 334, startTime), // 33.4 seconds
		testhelpers.CreateTestSplit(control2, 612, startTime), // 61.2 seconds
	}

	comp3 := testhelpers.CreateTestCompetitor(3, "Mike Johnson", club1, class)
	comp3.Status = "3" // DNF

	comp4 := testhelpers.CreateTestCompetitor(4, "Sarah Wilson", club2, class)
	comp4.Status = "21" // Cancelled (will show as DNS category)

	comp5 := testhelpers.CreateTestCompetitor(5, "Tom Brown", club1, class)
	comp5.Status = "4" // MP (Mispunch)

	s.Lock()
	s.Classes = []models.Class{class}
	s.Competitors = []models.Competitor{comp1, comp2, comp3, comp4, comp5}
	s.Unlock()

	h := New(s)
	router := setupTestRouter(h)

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/classes/1/results", nil)
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var results []service.ResultEntry
	err := json.Unmarshal(w.Body.Bytes(), &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Number of results = %d, want 5", len(results))
	}

	// Check first place
	if len(results) > 0 {
		if results[0].Position != 1 {
			t.Errorf("First position = %d, want 1", results[0].Position)
		}
		if results[0].Name != "John Doe" {
			t.Errorf("First place name = %q, want %q", results[0].Name, "John Doe")
		}
		if results[0].RunningTime != "1:23.4" {
			t.Errorf("First place time = %q, want %q", results[0].RunningTime, "1:23.4")
		}
		if results[0].Difference != "" {
			t.Error("First place should not have time difference")
		}
	}

	// Check second place
	if len(results) > 1 {
		if results[1].Position != 2 {
			t.Errorf("Second position = %d, want 2", results[1].Position)
		}
		if results[1].Difference != "+0:07.8" {
			t.Errorf("Second place time difference = %q, want %q", results[1].Difference, "+0:07.8")
		}
	}

	// Check DNF
	dnfFound := false
	mpFound := false
	dnsFound := false
	for _, result := range results {
		switch result.Status {
		case "Not Finished", "Disqualified", "Max. Time":
			dnfFound = true
			if result.Position != 0 {
				t.Error("Not Finished/Disqualified/Max. Time should not have position")
			}
			if result.RunningTime != "" {
				t.Error("Not Finished/Disqualified/Max. Time should not have running time")
			}
		case "Miss Punch":
			mpFound = true
		case "Not Started", "Cancelled", "Not Competing":
			dnsFound = true
		}
	}

	if !dnfFound {
		t.Error("Not Finished competitor not found in results")
	}
	if !mpFound {
		t.Error("Miss Punch competitor not found in results")
	}
	if !dnsFound {
		t.Error("Not Started competitor not found in results")
	}
}

func TestHandler_GetResults_EmptyClass(t *testing.T) {
	s := state.New()
	class := testhelpers.CreateTestClass(1, "Men Elite", 10)
	s.Lock()
	s.Classes = []models.Class{class}
	s.Unlock()

	h := New(s)
	router := setupTestRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/classes/1/results", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var results []service.ResultEntry
	err := json.Unmarshal(w.Body.Bytes(), &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Number of results = %d, want 0", len(results))
	}
}

func TestHandler_GetSplits(t *testing.T) {
	// Set up state with test data
	s := state.New()
	club1 := testhelpers.CreateTestClub(1, "Test Club 1", "SWE")
	club2 := testhelpers.CreateTestClub(2, "Test Club 2", "NOR")

	control1 := testhelpers.CreateTestControl(101, "Control 1")
	control2 := testhelpers.CreateTestControl(102, "Control 2")
	class := testhelpers.CreateTestClass(1, "Men Elite", 10, control1, control2)

	startTime := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)

	// Create competitors
	comp1 := testhelpers.CreateFinishedCompetitor(1, "John Doe", club1, class, 834)
	comp1.Splits = []models.Split{
		testhelpers.CreateTestSplit(control1, 302, startTime),
		testhelpers.CreateTestSplit(control2, 584, startTime),
	}

	comp2 := testhelpers.CreateFinishedCompetitor(2, "Jane Smith", club2, class, 912)
	comp2.Splits = []models.Split{
		testhelpers.CreateTestSplit(control1, 334, startTime),
		testhelpers.CreateTestSplit(control2, 612, startTime),
	}

	comp3 := testhelpers.CreateTestCompetitor(3, "Mike Johnson", club1, class)
	comp3.Status = "3" // DNF
	comp3.Splits = []models.Split{
		testhelpers.CreateTestSplit(control1, 350, startTime),
	}

	s.Lock()
	s.Classes = []models.Class{class}
	s.Competitors = []models.Competitor{comp1, comp2, comp3}
	s.Unlock()

	h := New(s)
	router := setupTestRouter(h)

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/classes/1/splits", nil)
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response service.SplitsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ClassName != "Men Elite" {
		t.Errorf("Class name = %q, want %q", response.ClassName, "Men Elite")
	}

	// Should have 3 splits: Control 1, Control 2, and Finish
	if len(response.Splits) != 3 {
		t.Errorf("Number of splits = %d, want 3", len(response.Splits))
	}

	// Check Control 1 standings
	if len(response.Splits) > 0 {
		split1 := response.Splits[0]
		if split1.ControlName != "Control 1" {
			t.Errorf("First split control name = %q, want %q", split1.ControlName, "Control 1")
		}
		if len(split1.Standings) != 3 {
			t.Errorf("Control 1 standings = %d, want 3", len(split1.Standings))
		}
		if len(split1.Standings) > 0 {
			if split1.Standings[0].Name != "John Doe" {
				t.Errorf("Control 1 leader = %q, want %q", split1.Standings[0].Name, "John Doe")
			}
			if split1.Standings[0].ElapsedTime == nil {
				t.Error("Control 1 leader time should not be nil")
			} else if *split1.Standings[0].ElapsedTime != "0:30.2" {
				t.Errorf("Control 1 leader time = %q, want %q", *split1.Standings[0].ElapsedTime, "0:30.2")
			}
		}
	}

	// Check Finish standings
	if len(response.Splits) > 2 {
		finishSplit := response.Splits[2]
		if finishSplit.ControlName != "Finish" {
			t.Errorf("Last split control name = %q, want %q", finishSplit.ControlName, "Finish")
		}
		// Only finished competitors should be in finish standings
		if len(finishSplit.Standings) != 2 {
			t.Errorf("Finish standings = %d, want 2", len(finishSplit.Standings))
		}
	}
}

func TestHandler_GetSplits_ClassNotFound(t *testing.T) {
	s := state.New()
	h := New(s)
	router := setupTestRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/classes/999/splits", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusNotFound)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "class not found" {
		t.Errorf("Error message = %q, want %q", response["error"], "class not found")
	}
}

func TestHandler_RadioTimeCalculation(t *testing.T) {
	// Set up state with test data
	s := state.New()
	club := testhelpers.CreateTestClub(1, "Test Club", "SWE")

	control1 := testhelpers.CreateTestControl(101, "Control 1")
	control2 := testhelpers.CreateTestControl(102, "Control 2")
	control3 := testhelpers.CreateTestControl(103, "Control 3")
	class := testhelpers.CreateTestClass(1, "Men Elite", 10, control1, control2, control3)

	startTime := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)

	// Create competitor with splits
	comp := testhelpers.CreateFinishedCompetitor(1, "John Doe", club, class, 1234) // 123.4 seconds
	comp.Splits = []models.Split{
		{Control: control1, PassingTime: startTime.Add(302 * 100 * time.Millisecond)}, // 30.2s
		{Control: control2, PassingTime: startTime.Add(584 * 100 * time.Millisecond)}, // 58.4s
		{Control: control3, PassingTime: startTime.Add(923 * 100 * time.Millisecond)}, // 92.3s
	}

	s.Lock()
	s.Classes = []models.Class{class}
	s.Competitors = []models.Competitor{comp}
	s.Unlock()

	h := New(s)
	router := setupTestRouter(h)

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/classes/1/results", nil)
	router.ServeHTTP(w, req)

	var results []service.ResultEntry
	err := json.Unmarshal(w.Body.Bytes(), &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Radio times are no longer included in results endpoint
	// as per the API cleanup requirement
}

func TestHandler_ConcurrentRequests(t *testing.T) {
	// Set up state with test data
	s := state.New()
	club := testhelpers.CreateTestClub(1, "Test Club", "SWE")
	class := testhelpers.CreateTestClass(1, "Men Elite", 10)

	competitors := make([]models.Competitor, 100)
	for i := 0; i < 100; i++ {
		competitors[i] = testhelpers.CreateFinishedCompetitor(i+1, fmt.Sprintf("Competitor %d", i+1), club, class, 800+i*10)
	}

	s.Lock()
	s.Classes = []models.Class{class}
	s.Competitors = competitors
	s.Unlock()

	h := New(s)
	router := setupTestRouter(h)

	// Run concurrent requests
	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Make multiple requests
			endpoints := []string{
				"/classes",
				"/classes/1/startlist",
				"/classes/1/results",
				"/classes/1/splits",
			}

			for _, endpoint := range endpoints {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", endpoint, nil)
				router.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Request to %s returned status %d", endpoint, w.Code)
				}
			}
		}()
	}

	wg.Wait()
}
