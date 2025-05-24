package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"meos-graphics/internal/models"
)

var (
	firstNames = []string{"Emma", "Oliver", "Sophia", "Liam", "Isabella", "Noah", "Mia", "Lucas", "Charlotte", "Ethan",
		"Amelia", "Mason", "Harper", "Elijah", "Evelyn", "James", "Abigail", "Benjamin", "Emily", "William"}
	lastNames = []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
		"Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin"}
	clubNames = []string{"OK Silva", "OK Pan", "OK Linné", "Lynx OK", "OK Orion", "OK Ravinen", "OK Kompassen", "OK Denseln",
		"Stora Tuna OK", "OK Kåre", "Sävedalens AIK", "Göteborg-Majorna OK", "Matteus SI", "Järfälla OK", "OK Södertörn"}
)

type competitorTiming struct {
	totalTime  time.Duration
	splitTimes []time.Duration // Cumulative times for each split
	finishTime time.Duration
}

type Generator struct {
	startTime      time.Time
	simulationTime time.Time
	classes        []models.Class
	clubs          []models.Club
	controls       []models.Control
	competitors    []models.Competitor
	rnd            *rand.Rand

	// Timing configuration
	duration     time.Duration
	phaseStart   time.Duration
	phaseRunning time.Duration
	phaseResults time.Duration

	// Pre-calculated timings for each competitor
	competitorTimings map[int]competitorTiming
}

func NewGenerator(duration, phaseStart, phaseRunning, phaseResults time.Duration) *Generator {
	return &Generator{
		rnd:               rand.New(rand.NewSource(time.Now().UnixNano())),
		duration:          duration,
		phaseStart:        phaseStart,
		phaseRunning:      phaseRunning,
		phaseResults:      phaseResults,
		competitorTimings: make(map[int]competitorTiming),
	}
}

func (g *Generator) GenerateInitialData(baseTime time.Time) (models.Event, []models.Control, []models.Class, []models.Club, []models.Competitor) {
	g.startTime = baseTime
	g.simulationTime = baseTime

	// Create event
	event := models.Event{
		Name:      "Simulation Event",
		Organizer: "MeOS Graphics Simulator",
		Start:     baseTime,
	}

	// Create controls
	g.controls = []models.Control{
		{ID: 1, Name: "Radio 1"},
		{ID: 2, Name: "Radio 2"},
		{ID: 3, Name: "Radio 3"},
	}

	// Create clubs
	g.clubs = make([]models.Club, 0, len(clubNames))
	for i, name := range clubNames {
		g.clubs = append(g.clubs, models.Club{
			ID:          i + 1,
			Name:        name,
			CountryCode: "SWE",
		})
	}

	// Create classes
	g.classes = []models.Class{
		{
			ID:            1,
			Name:          "Men Elite",
			OrderKey:      10,
			RadioControls: g.controls,
		},
		{
			ID:            2,
			Name:          "Women Elite",
			OrderKey:      20,
			RadioControls: g.controls,
		},
		{
			ID:            3,
			Name:          "Men Junior",
			OrderKey:      30,
			RadioControls: g.controls[:2], // Only 2 radio controls
		},
	}

	// Generate competitors
	g.competitors = g.generateCompetitors(baseTime)

	return event, g.controls, g.classes, g.clubs, g.competitors
}

func (g *Generator) generateCompetitors(baseTime time.Time) []models.Competitor {
	var competitors []models.Competitor
	competitorID := 1

	// Generate competitors for each class
	for _, class := range g.classes {
		numCompetitors := 15 + g.rnd.Intn(10) // 15-25 competitors per class

		for i := 0; i < numCompetitors; i++ {
			firstName := firstNames[g.rnd.Intn(len(firstNames))]
			lastName := lastNames[g.rnd.Intn(len(lastNames))]

			competitor := models.Competitor{
				ID:        competitorID,
				Card:      500000 + competitorID,
				Name:      fmt.Sprintf("%s %s", firstName, lastName),
				Club:      g.clubs[g.rnd.Intn(len(g.clubs))],
				Class:     class,
				Status:    "0",      // Not started
				StartTime: baseTime, // Everyone starts at the same time
				Splits:    []models.Split{},
			}

			// Pre-calculate timing for this competitor
			g.generateCompetitorTiming(competitorID, class)

			competitors = append(competitors, competitor)
			competitorID++
		}
	}

	return competitors
}

func (g *Generator) generateCompetitorTiming(competitorID int, class models.Class) {
	// Base time varies by class (in minutes)
	var baseTimeMinutes int
	switch class.Name {
	case "Men Elite":
		baseTimeMinutes = 45 + g.rnd.Intn(15) // 45-60 minutes
	case "Women Elite":
		baseTimeMinutes = 40 + g.rnd.Intn(15) // 40-55 minutes
	case "Men Junior":
		baseTimeMinutes = 30 + g.rnd.Intn(10) // 30-40 minutes
	default:
		baseTimeMinutes = 45 + g.rnd.Intn(15)
	}

	totalTime := time.Duration(baseTimeMinutes) * time.Minute

	// Calculate split times for radio controls
	var splitTimes []time.Duration
	numRadios := len(class.RadioControls)

	if numRadios > 0 {
		// Split the course into segments
		for i := 0; i < numRadios; i++ {
			// Each split occurs at approximately equal intervals with some variation
			baseRatio := float64(i+1) / float64(numRadios+1)
			variation := (g.rnd.Float64() - 0.5) * 0.1 // ±5% variation
			splitRatio := baseRatio + variation

			// Ensure splits are in order and reasonable
			if splitRatio < 0.1 {
				splitRatio = 0.1
			}
			if splitRatio > 0.9 {
				splitRatio = 0.9
			}

			splitTime := time.Duration(float64(totalTime) * splitRatio)

			// Ensure each split is after the previous one
			if i > 0 && splitTime <= splitTimes[i-1] {
				splitTime = splitTimes[i-1] + 30*time.Second
			}

			splitTimes = append(splitTimes, splitTime)
		}
	}

	g.competitorTimings[competitorID] = competitorTiming{
		totalTime:  totalTime,
		splitTimes: splitTimes,
		finishTime: totalTime,
	}
}

func (g *Generator) UpdateSimulation(currentTime time.Time) []models.Competitor {
	g.simulationTime = currentTime
	elapsed := currentTime.Sub(g.startTime)

	// Phase 1: Start list only
	if elapsed < g.phaseStart {
		return g.competitors
	}

	// Phase 2: Competitors running and finishing
	phaseRunningEnd := g.phaseStart + g.phaseRunning
	if elapsed >= g.phaseStart && elapsed < phaseRunningEnd {
		progress := float64(elapsed-g.phaseStart) / float64(g.phaseRunning)
		g.updateCompetitorProgress(progress)
	}

	// Phase 3: All finished, results stable
	if elapsed >= phaseRunningEnd && elapsed < g.duration {
		// Ensure all competitors are finished
		g.updateCompetitorProgress(1.0)
	}

	// Reset after full cycle
	if elapsed >= g.duration {
		g.resetSimulation()
	}

	return g.competitors
}

func (g *Generator) updateCompetitorProgress(progress float64) {
	for i := range g.competitors {
		comp := &g.competitors[i]

		// Skip if already finished
		if comp.Status == "1" && comp.FinishTime != nil {
			continue
		}

		// Get pre-calculated timing for this competitor
		timing, exists := g.competitorTimings[comp.ID]
		if !exists {
			continue
		}

		// Determine if this competitor should have finished by now
		// Spread finishes across the 7-minute window
		competitorProgress := float64(i) / float64(len(g.competitors))

		if progress >= competitorProgress {
			// This competitor should be running or finished
			if comp.Status == "0" {
				comp.Status = "2" // Running
			}

			// Progressive split revelation based on pre-calculated times
			var splits []models.Split
			elapsedSinceStart := g.simulationTime.Sub(comp.StartTime)

			for j, control := range comp.Class.RadioControls {
				if j < len(timing.splitTimes) {
					splitTime := timing.splitTimes[j]

					// Only reveal this split if the competitor should have reached it
					if elapsedSinceStart >= splitTime || progress >= 1.0 {
						passingTime := comp.StartTime.Add(splitTime)
						splits = append(splits, models.Split{
							Control:     control,
							PassingTime: passingTime,
						})
					}
				}
			}

			comp.Splits = splits

			// Check if should be finished
			finishProgress := competitorProgress + 0.1 // Finish slightly after starting
			if progress >= finishProgress && (elapsedSinceStart >= timing.finishTime || progress >= 1.0) {
				comp.Status = "1" // Finished
				finishTime := comp.StartTime.Add(timing.finishTime)
				comp.FinishTime = &finishTime
			}
		}
	}
}

func (g *Generator) resetSimulation() {
	// Reset all competitors
	for i := range g.competitors {
		g.competitors[i].Status = "0"
		g.competitors[i].FinishTime = nil
		g.competitors[i].Splits = []models.Split{}

		// Regenerate timing for this competitor
		g.generateCompetitorTiming(g.competitors[i].ID, g.competitors[i].Class)
	}

	// Reset start time
	g.startTime = time.Now()
}
