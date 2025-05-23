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

type Generator struct {
	startTime      time.Time
	simulationTime time.Time
	classes        []models.Class
	clubs          []models.Club
	controls       []models.Control
	competitors    []models.Competitor
	rnd            *rand.Rand
}

func NewGenerator() *Generator {
	return &Generator{
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
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
				Status:    "0", // Not started
				StartTime: baseTime, // Everyone starts at the same time
				Splits:    []models.Split{},
			}
			
			competitors = append(competitors, competitor)
			competitorID++
		}
	}

	return competitors
}

func (g *Generator) UpdateSimulation(currentTime time.Time) []models.Competitor {
	elapsed := currentTime.Sub(g.startTime)
	
	// Phase 1: 0-3 minutes - Only start list
	if elapsed < 3*time.Minute {
		return g.competitors
	}
	
	// Phase 2: 3-10 minutes - Competitors running and finishing
	if elapsed >= 3*time.Minute && elapsed < 10*time.Minute {
		progress := float64(elapsed-3*time.Minute) / float64(7*time.Minute)
		g.updateCompetitorProgress(progress)
	}
	
	// Phase 3: 10-15 minutes - All finished, results stable
	if elapsed >= 10*time.Minute && elapsed < 15*time.Minute {
		// Ensure all competitors are finished
		g.updateCompetitorProgress(1.0)
	}
	
	// Reset after 15 minutes
	if elapsed >= 15*time.Minute {
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
		
		// Determine if this competitor should have finished by now
		// Spread finishes across the 7-minute window
		competitorProgress := float64(i) / float64(len(g.competitors))
		
		if progress >= competitorProgress {
			// This competitor should be running or finished
			if comp.Status == "0" {
				comp.Status = "9" // Running
			}
			
			// Generate times with realistic variations
			baseMinutes := 35 + g.rnd.Intn(15) // 35-50 minutes total time
			baseSeconds := g.rnd.Intn(60)
			baseDeciseconds := g.rnd.Intn(10)
			
			totalDeciseconds := (baseMinutes*60+baseSeconds)*10 + baseDeciseconds
			
			// Calculate split times
			var splits []models.Split
			var accumulatedTime int
			
			for j, control := range comp.Class.RadioControls {
				// Each split is roughly 1/(n+1) of total time with variation
				splitRatio := float64(j+1) / float64(len(comp.Class.RadioControls)+1)
				targetTime := int(float64(totalDeciseconds) * splitRatio)
				
				// Add variation
				variation := g.rnd.Intn(200) - 100 // +/- 10 seconds
				splitTime := targetTime + variation
				
				if splitTime <= accumulatedTime {
					splitTime = accumulatedTime + 100 // At least 10 seconds
				}
				
				accumulatedTime = splitTime
				
				passingTime := comp.StartTime.Add(time.Duration(splitTime) * 100 * time.Millisecond)
				
				splits = append(splits, models.Split{
					Control:     control,
					PassingTime: passingTime,
				})
			}
			
			comp.Splits = splits
			
			// Check if should be finished
			finishProgress := competitorProgress + 0.1 // Finish slightly after last split
			if progress >= finishProgress {
				comp.Status = "1" // Finished
				finishTime := comp.StartTime.Add(time.Duration(totalDeciseconds) * 100 * time.Millisecond)
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
	}
	
	// Reset start time
	g.startTime = time.Now()
}