package state

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"meos-graphics/internal/models"
	"meos-graphics/internal/testhelpers"
)

func TestState_New(t *testing.T) {
	s := New()

	if s == nil {
		t.Fatal("New() returned nil")
	}

	// Verify initial state
	if s.Event != nil {
		t.Error("Initial Event should be nil")
	}
	if len(s.Controls) != 0 {
		t.Errorf("Initial Controls length = %d, want 0", len(s.Controls))
	}
	if len(s.Classes) != 0 {
		t.Errorf("Initial Classes length = %d, want 0", len(s.Classes))
	}
	if len(s.Clubs) != 0 {
		t.Errorf("Initial Clubs length = %d, want 0", len(s.Clubs))
	}
	if len(s.Competitors) != 0 {
		t.Errorf("Initial Competitors length = %d, want 0", len(s.Competitors))
	}
}

func TestState_EventGetSet(t *testing.T) {
	s := New()

	// Test getting nil event
	event := s.GetEvent()
	if event != nil {
		t.Error("GetEvent() should return nil initially")
	}

	// Test setting event
	testEvent := testhelpers.CreateTestEvent()
	s.SetEvent(testEvent)

	// Test getting event
	retrievedEvent := s.GetEvent()
	if retrievedEvent == nil {
		t.Fatal("GetEvent() returned nil after SetEvent")
	}
	if retrievedEvent.Name != testEvent.Name {
		t.Errorf("Event name = %q, want %q", retrievedEvent.Name, testEvent.Name)
	}

	// Test setting nil event
	s.SetEvent(nil)
	if s.GetEvent() != nil {
		t.Error("GetEvent() should return nil after setting nil")
	}
}

func TestState_GetControls(t *testing.T) {
	s := New()

	// Test empty controls
	controls := s.GetControls()
	if len(controls) != 0 {
		t.Errorf("GetControls() length = %d, want 0", len(controls))
	}

	// Add controls
	testControls := []models.Control{
		testhelpers.CreateTestControl(1, "Start"),
		testhelpers.CreateTestControl(2, "Control 1"),
		testhelpers.CreateTestControl(3, "Finish"),
	}
	s.Controls = testControls

	// Get controls
	retrievedControls := s.GetControls()
	if len(retrievedControls) != len(testControls) {
		t.Errorf("GetControls() length = %d, want %d", len(retrievedControls), len(testControls))
	}

	// Verify it's a copy
	if len(retrievedControls) > 0 {
		retrievedControls[0].Name = "Modified"
		if s.Controls[0].Name == "Modified" {
			t.Error("GetControls() should return a copy, not the original slice")
		}
	}
}

func TestState_GetClasses(t *testing.T) {
	s := New()

	// Test empty classes
	classes := s.GetClasses()
	if len(classes) != 0 {
		t.Errorf("GetClasses() length = %d, want 0", len(classes))
	}

	// Add classes
	control1 := testhelpers.CreateTestControl(101, "Control 1")
	control2 := testhelpers.CreateTestControl(102, "Control 2")
	testClasses := []models.Class{
		testhelpers.CreateTestClass(1, "Men Elite", 10, control1, control2),
		testhelpers.CreateTestClass(2, "Women Elite", 20, control1),
	}
	s.Classes = testClasses

	// Get classes
	retrievedClasses := s.GetClasses()
	if len(retrievedClasses) != len(testClasses) {
		t.Errorf("GetClasses() length = %d, want %d", len(retrievedClasses), len(testClasses))
	}

	// Verify it's a copy
	if len(retrievedClasses) > 0 {
		retrievedClasses[0].Name = "Modified"
		if s.Classes[0].Name == "Modified" {
			t.Error("GetClasses() should return a copy, not the original slice")
		}
	}
}

func TestState_GetClubs(t *testing.T) {
	s := New()

	// Test empty clubs
	clubs := s.GetClubs()
	if len(clubs) != 0 {
		t.Errorf("GetClubs() length = %d, want 0", len(clubs))
	}

	// Add clubs
	testClubs := []models.Club{
		testhelpers.CreateTestClub(1, "Test Club 1", "SWE"),
		testhelpers.CreateTestClub(2, "Test Club 2", "NOR"),
	}
	s.Clubs = testClubs

	// Get clubs
	retrievedClubs := s.GetClubs()
	if len(retrievedClubs) != len(testClubs) {
		t.Errorf("GetClubs() length = %d, want %d", len(retrievedClubs), len(testClubs))
	}

	// Verify it's a copy
	if len(retrievedClubs) > 0 {
		retrievedClubs[0].Name = "Modified"
		if s.Clubs[0].Name == "Modified" {
			t.Error("GetClubs() should return a copy, not the original slice")
		}
	}
}

func TestState_GetCompetitors(t *testing.T) {
	s := New()

	// Test empty competitors
	competitors := s.GetCompetitors()
	if len(competitors) != 0 {
		t.Errorf("GetCompetitors() length = %d, want 0", len(competitors))
	}

	// Add competitors
	club := testhelpers.CreateTestClub(1, "Test Club", "SWE")
	class := testhelpers.CreateTestClass(1, "Men Elite", 10)
	testCompetitors := []models.Competitor{
		testhelpers.CreateTestCompetitor(1, "John Doe", club, class),
		testhelpers.CreateTestCompetitor(2, "Jane Smith", club, class),
	}
	s.Competitors = testCompetitors

	// Get competitors
	retrievedCompetitors := s.GetCompetitors()
	if len(retrievedCompetitors) != len(testCompetitors) {
		t.Errorf("GetCompetitors() length = %d, want %d", len(retrievedCompetitors), len(testCompetitors))
	}

	// Verify it's a copy
	if len(retrievedCompetitors) > 0 {
		retrievedCompetitors[0].Name = "Modified"
		if s.Competitors[0].Name == "Modified" {
			t.Error("GetCompetitors() should return a copy, not the original slice")
		}
	}
}

func TestState_GetCompetitorsByClass(t *testing.T) {
	s := New()

	// Create test data
	club := testhelpers.CreateTestClub(1, "Test Club", "SWE")
	class1 := testhelpers.CreateTestClass(1, "Men Elite", 10)
	class2 := testhelpers.CreateTestClass(2, "Women Elite", 20)

	competitors := []models.Competitor{
		testhelpers.CreateTestCompetitor(1, "John Doe", club, class1),
		testhelpers.CreateTestCompetitor(2, "Jane Smith", club, class2),
		testhelpers.CreateTestCompetitor(3, "Mike Johnson", club, class1),
		testhelpers.CreateTestCompetitor(4, "Sarah Wilson", club, class2),
	}
	s.Competitors = competitors

	// Test getting competitors by class 1
	class1Competitors := s.GetCompetitorsByClass(1)
	if len(class1Competitors) != 2 {
		t.Errorf("GetCompetitorsByClass(1) length = %d, want 2", len(class1Competitors))
	}

	// Verify correct competitors returned
	for _, comp := range class1Competitors {
		if comp.Class.ID != 1 {
			t.Errorf("Competitor class ID = %d, want 1", comp.Class.ID)
		}
	}

	// Test getting competitors by class 2
	class2Competitors := s.GetCompetitorsByClass(2)
	if len(class2Competitors) != 2 {
		t.Errorf("GetCompetitorsByClass(2) length = %d, want 2", len(class2Competitors))
	}

	// Test non-existent class
	noCompetitors := s.GetCompetitorsByClass(999)
	if len(noCompetitors) != 0 {
		t.Errorf("GetCompetitorsByClass(999) length = %d, want 0", len(noCompetitors))
	}
}

func TestState_GetCompetitor(t *testing.T) {
	s := New()

	// Create test data
	club := testhelpers.CreateTestClub(1, "Test Club", "SWE")
	class := testhelpers.CreateTestClass(1, "Men Elite", 10)

	competitors := []models.Competitor{
		testhelpers.CreateTestCompetitor(1, "John Doe", club, class),
		testhelpers.CreateTestCompetitor(2, "Jane Smith", club, class),
		testhelpers.CreateTestCompetitor(3, "Mike Johnson", club, class),
	}
	s.Competitors = competitors

	// Test getting existing competitor
	comp := s.GetCompetitor(2)
	if comp == nil {
		t.Fatal("GetCompetitor(2) returned nil")
	}
	if comp.ID != 2 {
		t.Errorf("Competitor ID = %d, want 2", comp.ID)
	}
	if comp.Name != "Jane Smith" {
		t.Errorf("Competitor name = %q, want %q", comp.Name, "Jane Smith")
	}

	// Test getting non-existent competitor
	noComp := s.GetCompetitor(999)
	if noComp != nil {
		t.Error("GetCompetitor(999) should return nil")
	}
}

func TestState_ConcurrentReads(t *testing.T) {
	s := New()

	// Set up test data
	s.SetEvent(testhelpers.CreateTestEvent())
	s.Controls = []models.Control{
		testhelpers.CreateTestControl(1, "Start"),
		testhelpers.CreateTestControl(2, "Finish"),
	}
	s.Classes = []models.Class{
		testhelpers.CreateTestClass(1, "Men Elite", 10),
	}
	s.Clubs = []models.Club{
		testhelpers.CreateTestClub(1, "Test Club", "SWE"),
	}
	s.Competitors = []models.Competitor{
		testhelpers.CreateTestCompetitor(1, "John Doe", s.Clubs[0], s.Classes[0]),
	}

	// Run concurrent reads
	var wg sync.WaitGroup
	numGoroutines := 100
	numIterations := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Perform various read operations
				_ = s.GetEvent()
				_ = s.GetControls()
				_ = s.GetClasses()
				_ = s.GetClubs()
				_ = s.GetCompetitors()
				_ = s.GetCompetitorsByClass(1)
				_ = s.GetCompetitor(1)
			}
		}()
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

func TestState_ConcurrentReadWrite(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	numReaders := 50
	numWriters := 10
	numIterations := 100

	// Start readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Perform read operations
				_ = s.GetEvent()
				_ = s.GetControls()
				_ = s.GetClasses()
				_ = s.GetClubs()
				_ = s.GetCompetitors()
				_ = s.GetCompetitorsByClass(id % 3)
				_ = s.GetCompetitor(id)

				// Small delay to allow for more interleaving
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Start writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Perform write operations
				event := testhelpers.CreateTestEvent()
				event.Name = fmt.Sprintf("Event %d-%d", id, j)
				s.SetEvent(event)

				// Direct writes (simulating what adapter does)
				s.Lock()
				s.Controls = []models.Control{
					testhelpers.CreateTestControl(id, fmt.Sprintf("Control %d", id)),
				}
				s.Classes = []models.Class{
					testhelpers.CreateTestClass(id, fmt.Sprintf("Class %d", id), id*10),
				}
				s.Clubs = []models.Club{
					testhelpers.CreateTestClub(id, fmt.Sprintf("Club %d", id), "SWE"),
				}
				s.Competitors = []models.Competitor{
					testhelpers.CreateTestCompetitor(id, fmt.Sprintf("Competitor %d", id), s.Clubs[0], s.Classes[0]),
				}
				s.Unlock()

				// Small delay to allow for more interleaving
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

func TestState_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	s := New()

	// Set up initial data
	numClasses := 10
	numClubs := 20
	numCompetitorsPerClass := 50

	classes := make([]models.Class, numClasses)
	for i := 0; i < numClasses; i++ {
		classes[i] = testhelpers.CreateTestClass(i+1, fmt.Sprintf("Class %d", i+1), i*10)
	}

	clubs := make([]models.Club, numClubs)
	for i := 0; i < numClubs; i++ {
		clubs[i] = testhelpers.CreateTestClub(i+1, fmt.Sprintf("Club %d", i+1), "SWE")
	}

	competitors := make([]models.Competitor, 0, numClasses*numCompetitorsPerClass)
	compID := 1
	for _, class := range classes {
		for j := 0; j < numCompetitorsPerClass; j++ {
			club := clubs[j%numClubs]
			comp := testhelpers.CreateTestCompetitor(compID, fmt.Sprintf("Competitor %d", compID), club, class)
			competitors = append(competitors, comp)
			compID++
		}
	}

	s.Lock()
	s.Classes = classes
	s.Clubs = clubs
	s.Competitors = competitors
	s.Unlock()

	// Run stress test with many concurrent operations
	var wg sync.WaitGroup
	var readCount int64
	var writeCount int64

	// Many readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				classID := (id+j)%numClasses + 1
				_ = s.GetCompetitorsByClass(classID)
				atomic.AddInt64(&readCount, 1)
			}
		}(i)
	}

	// Some writers updating competitor status
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.Lock()
				// Simulate updating competitor status
				if len(s.Competitors) > 0 {
					idx := (id*100 + j) % len(s.Competitors)
					s.Competitors[idx].Status = "1"
					finishTime := time.Now()
					s.Competitors[idx].FinishTime = &finishTime
				}
				s.Unlock()
				atomic.AddInt64(&writeCount, 1)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Stress test completed: %d reads, %d writes", readCount, writeCount)
}

func TestState_LockUnlock(t *testing.T) {
	s := New()

	// Test that Lock() and Unlock() work correctly
	done := make(chan bool)

	// Lock in main goroutine
	s.Lock()

	// Try to lock in another goroutine
	go func() {
		s.Lock()
		defer s.Unlock()
		done <- true
	}()

	// Give the goroutine time to block
	select {
	case <-done:
		t.Error("Second Lock() should have blocked")
	case <-time.After(100 * time.Millisecond):
		// Expected behavior - the goroutine is blocked
	}

	// Unlock and verify the goroutine can proceed
	s.Unlock()

	select {
	case <-done:
		// Expected behavior - the goroutine acquired the lock
	case <-time.After(100 * time.Millisecond):
		t.Error("Second Lock() should have succeeded after Unlock()")
	}
}

func BenchmarkState_GetCompetitors(b *testing.B) {
	s := New()

	// Set up test data
	club := testhelpers.CreateTestClub(1, "Test Club", "SWE")
	class := testhelpers.CreateTestClass(1, "Men Elite", 10)

	competitors := make([]models.Competitor, 1000)
	for i := 0; i < 1000; i++ {
		competitors[i] = testhelpers.CreateTestCompetitor(i+1, fmt.Sprintf("Competitor %d", i+1), club, class)
	}
	s.Competitors = competitors

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = s.GetCompetitors()
	}
}

func BenchmarkState_GetCompetitorsByClass(b *testing.B) {
	s := New()

	// Set up test data with multiple classes
	club := testhelpers.CreateTestClub(1, "Test Club", "SWE")
	classes := make([]models.Class, 10)
	for i := 0; i < 10; i++ {
		classes[i] = testhelpers.CreateTestClass(i+1, fmt.Sprintf("Class %d", i+1), i*10)
	}

	competitors := make([]models.Competitor, 1000)
	for i := 0; i < 1000; i++ {
		class := classes[i%10]
		competitors[i] = testhelpers.CreateTestCompetitor(i+1, fmt.Sprintf("Competitor %d", i+1), club, class)
	}
	s.Competitors = competitors

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = s.GetCompetitorsByClass((i % 10) + 1)
	}
}

func BenchmarkState_ConcurrentReads(b *testing.B) {
	s := New()

	// Set up test data
	s.SetEvent(testhelpers.CreateTestEvent())
	s.Controls = []models.Control{testhelpers.CreateTestControl(1, "Start")}
	s.Classes = []models.Class{testhelpers.CreateTestClass(1, "Men Elite", 10)}
	s.Clubs = []models.Club{testhelpers.CreateTestClub(1, "Test Club", "SWE")}
	s.Competitors = []models.Competitor{
		testhelpers.CreateTestCompetitor(1, "John Doe", s.Clubs[0], s.Classes[0]),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = s.GetEvent()
			_ = s.GetCompetitors()
			_ = s.GetCompetitorsByClass(1)
		}
	})
}
