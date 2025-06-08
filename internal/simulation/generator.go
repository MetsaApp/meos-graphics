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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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
	massStart    bool

	// Content configuration
	numClasses      int
	runnersPerClass int
	radioControls   int

	// Pre-calculated timings for each competitor
	competitorTimings map[int]competitorTiming
}

func NewGenerator(duration, phaseStart, phaseRunning, phaseResults time.Duration, massStart bool, numClasses, runnersPerClass, radioControls int) *Generator {
	return &Generator{
		rnd:               rand.New(rand.NewSource(time.Now().UnixNano())),
		duration:          duration,
		phaseStart:        phaseStart,
		phaseRunning:      phaseRunning,
		phaseResults:      phaseResults,
		massStart:         massStart,
		numClasses:        numClasses,
		runnersPerClass:   runnersPerClass,
		radioControls:     radioControls,
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

	// Create controls based on configuration
	g.controls = make([]models.Control, 0, g.radioControls)
	for i := 1; i <= g.radioControls; i++ {
		g.controls = append(g.controls, models.Control{
			ID:   i,
			Name: fmt.Sprintf("Radio %d", i),
		})
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

	// Create classes based on configuration
	classNames := []string{"Men Elite", "Women Elite", "Men Junior", "Women Junior", "Men 21", "Women 21", "Men 35", "Women 35", "Men 40", "Women 40"}
	g.classes = make([]models.Class, 0, g.numClasses)
	for i := 1; i <= g.numClasses; i++ {
		// Cycle through class names if we have more classes than predefined names
		className := fmt.Sprintf("Class %d", i)
		if i-1 < len(classNames) {
			className = classNames[i-1]
		}

		// Determine radio controls for this class
		// Most classes get all controls, but vary some for realism
		radioControls := g.controls
		if g.radioControls > 1 && i%3 == 0 {
			// Every third class gets one fewer control for variety
			radioControls = g.controls[:g.radioControls-1]
		}

		g.classes = append(g.classes, models.Class{
			ID:            i,
			Name:          className,
			OrderKey:      i * 10,
			RadioControls: radioControls,
		})
	}

	// Generate competitors
	g.competitors = g.generateCompetitors(baseTime)

	return event, g.controls, g.classes, g.clubs, g.competitors
}

func (g *Generator) generateCompetitors(baseTime time.Time) []models.Competitor {
	var competitors []models.Competitor
	competitorID := 1
	competitorIndex := 0 // For staggered starts

	// Generate competitors for each class
	for _, class := range g.classes {
		numCompetitors := g.runnersPerClass

		for i := 0; i < numCompetitors; i++ {
			firstName := firstNames[g.rnd.Intn(len(firstNames))]
			lastName := lastNames[g.rnd.Intn(len(lastNames))]

			var startTime time.Time
			if g.massStart {
				// Mass start - everyone starts at the same time
				startTime = baseTime.Add(g.phaseStart)
			} else {
				// Stagger start times - calculate interval based on available time
				// We want all competitors to start with enough time to finish
				minRunTime := 5 * time.Minute // Minimum time needed to complete
				if g.phaseRunning < 5*time.Minute {
					minRunTime = g.phaseRunning / 2
				}

				// Calculate max start time to allow minimum run time
				maxStartOffset := g.phaseRunning - minRunTime
				if maxStartOffset < 0 {
					maxStartOffset = 0
				}

				// Calculate appropriate interval
				totalCompetitors := 0
				for range g.classes {
					totalCompetitors += g.runnersPerClass
				}

				var startInterval time.Duration
				if totalCompetitors > 0 && maxStartOffset > 0 {
					startInterval = maxStartOffset / time.Duration(totalCompetitors)
					// Cap at 2 minutes max, 10 seconds min
					if startInterval > 2*time.Minute {
						startInterval = 2 * time.Minute
					} else if startInterval < 10*time.Second {
						startInterval = 10 * time.Second
					}
				} else {
					startInterval = 30 * time.Second // Default fallback
				}

				startOffset := time.Duration(competitorIndex) * startInterval
				startTime = baseTime.Add(g.phaseStart).Add(startOffset)
				competitorIndex++
			}

			competitor := models.Competitor{
				ID:        competitorID,
				Card:      500000 + competitorID,
				Name:      fmt.Sprintf("%s %s", firstName, lastName),
				Club:      g.clubs[g.rnd.Intn(len(g.clubs))],
				Class:     class,
				Status:    "0", // Not started
				StartTime: startTime,
				Splits:    []models.Split{},
			}

			// Pre-calculate timing for this competitor
			g.generateCompetitorTiming(competitorID, class)

			competitors = append(competitors, competitor)
			competitorID++
		}
	}

	// Adjust start times for staggered starts to fit within running phase
	if !g.massStart && len(competitors) > 0 {
		startInterval := g.calculateStartInterval()
		competitorIndex = 0
		for i := range competitors {
			startOffset := time.Duration(competitorIndex) * startInterval
			competitors[i].StartTime = baseTime.Add(g.phaseStart).Add(startOffset)
			competitorIndex++
		}
	}

	return competitors
}

func (g *Generator) generateSplitTimes(totalTime time.Duration, class models.Class) []time.Duration {
	var splitTimes []time.Duration
	numRadios := len(class.RadioControls)

	if numRadios > 0 {
		// Reserve some time for the final leg to finish
		// For short runs, use less reservation
		reserveRatio := 0.1
		if totalTime < 5*time.Minute {
			reserveRatio = 0.15 // Reserve 15% for final leg on short courses
		}
		maxSplitTime := time.Duration(float64(totalTime) * (1 - reserveRatio))

		// Calculate minimum time between splits based on total time
		// For short runs, use proportionally shorter minimum times
		minLegTime := 30 * time.Second
		if totalTime < 2*time.Minute {
			minLegTime = time.Duration(float64(totalTime) / float64(numRadios+2))
			if minLegTime < 5*time.Second {
				minLegTime = 5 * time.Second
			}
		} else if totalTime < 5*time.Minute {
			minLegTime = 15 * time.Second
		}

		// Split the course into segments
		for i := 0; i < numRadios; i++ {
			// Each split occurs at approximately equal intervals
			baseRatio := float64(i+1) / float64(numRadios+1)

			// Add some variation but keep it reasonable
			variation := (g.rnd.Float64() - 0.5) * 0.1 // ±5% variation
			splitRatio := baseRatio + variation

			// Calculate split time
			splitTime := time.Duration(float64(maxSplitTime) * splitRatio)

			// Ensure minimum cumulative time
			minCumulativeTime := time.Duration(i+1) * minLegTime
			if splitTime < minCumulativeTime {
				splitTime = minCumulativeTime
			}

			// Ensure each split is after the previous one
			if i > 0 {
				minNextSplit := splitTimes[i-1] + minLegTime
				if splitTime <= splitTimes[i-1] {
					splitTime = minNextSplit
				}
			}

			// Ensure split is well before finish time
			maxAllowedSplit := maxSplitTime - time.Duration(numRadios-i)*minLegTime
			if splitTime > maxAllowedSplit {
				splitTime = maxAllowedSplit
			}

			splitTimes = append(splitTimes, splitTime)
		}
	}

	return splitTimes
}

func (g *Generator) generateCompetitorTiming(competitorID int, class models.Class) {
	// Calculate the maximum time a competitor can take (90% of running phase)
	maxRunningTime := time.Duration(float64(g.phaseRunning) * 0.9)

	// For short simulations, scale times appropriately
	if g.phaseRunning < 5*time.Minute {
		// For very short runs, use a wider spread of times
		// Minimum time is 30% of max, maximum is 90% of max
		minTime := time.Duration(float64(maxRunningTime) * 0.3)
		timeRange := time.Duration(float64(maxRunningTime) * 0.6)

		// Generate base time with good spread
		baseTime := minTime + time.Duration(g.rnd.Int63n(int64(timeRange)))

		// Add seconds-level variation for more realistic spread
		secondsVariation := g.rnd.Intn(20) - 10 // ±10 seconds
		baseTime += time.Duration(secondsVariation) * time.Second

		// Add deciseconds for realism
		deciseconds := g.rnd.Intn(10)
		totalTime := baseTime + time.Duration(deciseconds)*100*time.Millisecond

		// Ensure time doesn't exceed max
		if totalTime > maxRunningTime {
			totalTime = maxRunningTime - time.Duration(g.rnd.Intn(10))*time.Second
		}

		// For test consistency, ensure minimum 5 minutes if phase allows it
		if g.phaseRunning >= 7*time.Minute && totalTime < 5*time.Minute {
			// Generate a more varied time around 5 minutes to avoid identical times
			// But respect the max allowed time (90% of running phase)
			baseVariation := g.rnd.Intn(40)    // 0-40 seconds
			secondsVariation := g.rnd.Intn(20) // 0-20 additional seconds
			newTime := 5*time.Minute + time.Duration(baseVariation)*time.Second + time.Duration(secondsVariation)*time.Second

			// Ensure we don't exceed the maximum allowed time
			if newTime <= maxRunningTime {
				totalTime = newTime
			} else {
				// Fall back to a time within bounds but still varied
				variation := g.rnd.Intn(30) // 0-30 seconds
				totalTime = 5*time.Minute + time.Duration(variation)*time.Second
			}
		}

		g.competitorTimings[competitorID] = competitorTiming{
			totalTime:  totalTime,
			splitTimes: g.generateSplitTimes(totalTime, class),
			finishTime: totalTime,
		}
		return
	}

	// For longer simulations, use class-based times
	var minMinutes, maxMinutes int
	maxAllowedMinutes := int(maxRunningTime.Minutes())

	switch class.Name {
	case "Men Elite":
		minMinutes = 45
		maxMinutes = minInt(60, maxAllowedMinutes)
	case "Women Elite":
		minMinutes = 40
		maxMinutes = minInt(55, maxAllowedMinutes)
	case "Men Junior":
		minMinutes = 30
		maxMinutes = minInt(40, maxAllowedMinutes)
	default:
		minMinutes = 45
		maxMinutes = minInt(60, maxAllowedMinutes)
	}

	// Ensure minimum time is reasonable and doesn't exceed max
	if minMinutes > maxMinutes {
		minMinutes = int(float64(maxMinutes) * 0.7)
	}

	// For standard simulations (7+ minute running phase), ensure realistic minimums
	// This must be done AFTER adjusting for maxMinutes
	if g.phaseRunning >= 7*time.Minute {
		// Enforce 5-minute minimum for standard competitions
		if minMinutes < 5 {
			minMinutes = 5
		}
		// If max is less than min, set them equal
		if maxMinutes < minMinutes {
			maxMinutes = minMinutes
		}
	}

	// Generate time with good spread
	timeRange := maxMinutes - minMinutes
	if timeRange <= 0 {
		timeRange = 1
	}
	baseTimeMinutes := minMinutes + g.rnd.Intn(timeRange+1)

	// Add seconds-level variation for more spread
	secondsVariation := g.rnd.Intn(60) - 30 // ±30 seconds
	baseTime := time.Duration(baseTimeMinutes)*time.Minute + time.Duration(secondsVariation)*time.Second

	// Add deciseconds for realism
	deciseconds := g.rnd.Intn(10)
	totalTime := baseTime + time.Duration(deciseconds)*100*time.Millisecond

	// Ensure we don't go below minimum after all variations
	minAllowedTime := time.Duration(minMinutes) * time.Minute
	if totalTime < minAllowedTime {
		totalTime = minAllowedTime
	}

	// Final check to ensure time doesn't exceed max
	if totalTime > maxRunningTime {
		totalTime = maxRunningTime - time.Duration(g.rnd.Intn(30))*time.Second
	}

	g.competitorTimings[competitorID] = competitorTiming{
		totalTime:  totalTime,
		splitTimes: g.generateSplitTimes(totalTime, class),
		finishTime: totalTime,
	}
}

func (g *Generator) UpdateSimulation(currentTime time.Time) []models.Competitor {
	g.simulationTime = currentTime
	elapsed := currentTime.Sub(g.startTime)

	// Phase 1: Start list only
	if elapsed < g.phaseStart {
		return g.copyCompetitors()
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

	return g.copyCompetitors()
}

// copyCompetitors creates a deep copy of the competitors slice
func (g *Generator) copyCompetitors() []models.Competitor {
	result := make([]models.Competitor, len(g.competitors))
	for i, comp := range g.competitors {
		// Copy the competitor
		result[i] = comp

		// Deep copy the splits
		if len(comp.Splits) > 0 {
			result[i].Splits = make([]models.Split, len(comp.Splits))
			copy(result[i].Splits, comp.Splits)
		}

		// Copy finish time pointer
		if comp.FinishTime != nil {
			finishTime := *comp.FinishTime
			result[i].FinishTime = &finishTime
		}
	}
	return result
}

// GetCurrentPhase returns the current simulation phase and remaining time
func (g *Generator) GetCurrentPhase() (phase string, nextPhaseIn time.Duration) {
	elapsed := g.simulationTime.Sub(g.startTime)

	if elapsed < g.phaseStart {
		return "Start List", g.phaseStart - elapsed
	}

	phaseRunningEnd := g.phaseStart + g.phaseRunning
	if elapsed < phaseRunningEnd {
		return "Running", phaseRunningEnd - elapsed
	}

	if elapsed < g.duration {
		return "Results", g.duration - elapsed
	}

	return "Resetting", 0
}

func (g *Generator) updateCompetitorProgress(progress float64) {
	// Calculate the absolute end time of the running phase
	runningPhaseEnd := g.startTime.Add(g.phaseStart).Add(g.phaseRunning)

	for i := range g.competitors {
		comp := &g.competitors[i]

		// Skip if already finished
		if comp.Status == "1" && comp.FinishTime != nil {
			continue
		}

		// Get pre-calculated timing for this competitor
		timing, exists := g.competitorTimings[comp.ID]
		if !exists {
			// This should not happen - all competitors should have timings
			// Generate timing on the fly if missing
			g.generateCompetitorTiming(comp.ID, comp.Class)
			timing = g.competitorTimings[comp.ID]
		}

		// Check if competitor should have started based on current time
		if g.simulationTime.Before(comp.StartTime) {
			// Still waiting to start
			comp.Status = "0"
			continue
		}

		// This competitor has started
		if comp.Status == "0" {
			comp.Status = "2" // Running
		}

		// Calculate elapsed time since this competitor started
		elapsedSinceStart := g.simulationTime.Sub(comp.StartTime)

		// Don't update splits for already finished competitors
		if comp.Status != "1" {

			// Progressive split revelation based on pre-calculated times
			var splits []models.Split

			for j, control := range comp.Class.RadioControls {
				if j < len(timing.splitTimes) {
					splitTime := timing.splitTimes[j]

					// Calculate when this split would be reached
					splitPassTime := comp.StartTime.Add(splitTime)

					// Only reveal this split if:
					// 1. Enough time has elapsed since start
					// 2. The split time is before the end of running phase
					// 3. We're at 100% progress (forcing all to finish)
					if elapsedSinceStart >= splitTime && splitPassTime.Before(runningPhaseEnd) {
						splits = append(splits, models.Split{
							Control:     control,
							PassingTime: splitPassTime,
						})
					} else if progress >= 1.0 && splitPassTime.Before(runningPhaseEnd) {
						// Force reveal at 100% progress if within bounds
						splits = append(splits, models.Split{
							Control:     control,
							PassingTime: splitPassTime,
						})
					}
				} else if progress >= 1.0 {
					// In phase 3, create missing splits for controls without pre-calculated times
					// This can happen for very short times or late starters

					// Use the timing finish time as the basis for split generation
					availableTime := timing.finishTime
					if comp.FinishTime != nil {
						// If already finished, use actual finish time
						availableTime = comp.FinishTime.Sub(comp.StartTime)
					}

					// Generate split time proportionally
					splitRatio := float64(j+1) / float64(len(comp.Class.RadioControls)+1)
					splitTime := time.Duration(float64(availableTime) * splitRatio * 0.9)
					splitPassTime := comp.StartTime.Add(splitTime)

					splits = append(splits, models.Split{
						Control:     control,
						PassingTime: splitPassTime,
					})
				}
			}

			// Update splits - in phase 3, ensure all splits are present
			if progress >= 1.0 && comp.Status == "1" {
				// For finished competitors in phase 3, ensure they have all splits
				if len(splits) < len(comp.Class.RadioControls) {
					// Generate any missing splits
					for j := len(splits); j < len(comp.Class.RadioControls); j++ {
						control := comp.Class.RadioControls[j]
						availableTime := timing.finishTime
						if comp.FinishTime != nil {
							availableTime = comp.FinishTime.Sub(comp.StartTime)
						}

						splitRatio := float64(j+1) / float64(len(comp.Class.RadioControls)+1)
						splitTime := time.Duration(float64(availableTime) * splitRatio * 0.9)
						splitPassTime := comp.StartTime.Add(splitTime)

						splits = append(splits, models.Split{
							Control:     control,
							PassingTime: splitPassTime,
						})
					}
				}
				comp.Splits = splits
			} else if len(splits) > len(comp.Splits) {
				// Normal progression - only update if we have more splits
				comp.Splits = splits
			}
		}

		// Check if should be finished
		potentialFinishTime := comp.StartTime.Add(timing.finishTime)

		// Only mark as finished if:
		// 1. Enough time has elapsed
		// 2. The finish time is before the end of running phase
		// 3. We're at 100% progress (forcing all to finish)
		if elapsedSinceStart >= timing.finishTime && potentialFinishTime.Before(runningPhaseEnd) {
			comp.Status = "1" // Finished
			comp.FinishTime = &potentialFinishTime
		} else if progress >= 1.0 && potentialFinishTime.Before(runningPhaseEnd) {
			// Force finish at 100% progress if within bounds
			comp.Status = "1" // Finished
			comp.FinishTime = &potentialFinishTime
		} else if progress >= 1.0 && !potentialFinishTime.Before(runningPhaseEnd) {
			// If the calculated finish time exceeds bounds, cap it at the running phase end
			// This ensures everyone finishes within the phase
			cappedFinishTime := runningPhaseEnd.Add(-1 * time.Second) // 1 second before phase end

			// But make sure the capped time is after the start time
			if cappedFinishTime.Before(comp.StartTime) || cappedFinishTime.Equal(comp.StartTime) {
				// If running phase ends before this competitor even starts,
				// give them a reasonable time to complete based on the phase duration
				minRunTime := 5 * time.Minute // Default minimum for standard competitions
				if g.phaseRunning < 5*time.Minute {
					// For short simulations, use proportional minimum
					minRunTime = time.Duration(float64(g.phaseRunning) * 0.5)
				}
				cappedFinishTime = comp.StartTime.Add(minRunTime)
			}

			// Additional check: ensure minimum run time for standard competitions
			actualRunTime := cappedFinishTime.Sub(comp.StartTime)
			if g.phaseRunning >= 7*time.Minute && actualRunTime < 5*time.Minute {
				// For standard competitions, enforce 5 minute minimum with better variation
				// But respect the max allowed time (90% of running phase)
				maxAllowedRunTime := time.Duration(float64(g.phaseRunning) * 0.9)
				baseVariation := g.rnd.Intn(40)    // 0-40 seconds
				secondsVariation := g.rnd.Intn(20) // 0-20 additional seconds
				newRunTime := 5*time.Minute + time.Duration(baseVariation)*time.Second + time.Duration(secondsVariation)*time.Second

				if newRunTime <= maxAllowedRunTime {
					cappedFinishTime = comp.StartTime.Add(newRunTime)
				} else {
					// Fall back to a time within bounds but still varied
					variation := g.rnd.Intn(30) // 0-30 seconds
					cappedFinishTime = comp.StartTime.Add(5*time.Minute + time.Duration(variation)*time.Second)
				}
			}

			comp.Status = "1" // Finished
			comp.FinishTime = &cappedFinishTime

			// Ensure we have splits for finished competitors
			if len(comp.Splits) < len(comp.Class.RadioControls) {
				// Create missing splits or all splits if none exist
				availableTime := cappedFinishTime.Sub(comp.StartTime)

				// If we have some splits, start from where we left off
				existingSplits := comp.Splits
				comp.Splits = []models.Split{} // Reset to rebuild properly

				for j, control := range comp.Class.RadioControls {
					// Check if we already have this split
					var existingSplit *models.Split
					for k := range existingSplits {
						if k == j {
							existingSplit = &existingSplits[k]
							break
						}
					}

					if existingSplit != nil {
						// Use existing split
						comp.Splits = append(comp.Splits, *existingSplit)
					} else {
						// Create new split
						splitRatio := float64(j+1) / float64(len(comp.Class.RadioControls)+1)
						splitTime := time.Duration(float64(availableTime) * splitRatio * 0.9)
						splitPassTime := comp.StartTime.Add(splitTime)

						comp.Splits = append(comp.Splits, models.Split{
							Control:     control,
							PassingTime: splitPassTime,
						})
					}
				}
			} else if len(comp.Splits) == len(comp.Class.RadioControls) {
				// Adjust existing splits to ensure they're before the finish
				adjustedSplits := []models.Split{}
				for j, split := range comp.Splits {
					if split.PassingTime.Before(cappedFinishTime) {
						adjustedSplits = append(adjustedSplits, split)
					} else {
						// Cap this split time
						adjustedTime := cappedFinishTime.Add(-time.Duration(len(comp.Splits)-j) * 10 * time.Second)
						adjustedSplits = append(adjustedSplits, models.Split{
							Control:     split.Control,
							PassingTime: adjustedTime,
						})
					}
				}
				comp.Splits = adjustedSplits
			}
		}
	}
}

func (g *Generator) calculateStartInterval() time.Duration {
	// Calculate interval based on available time
	minRunTime := 5 * time.Minute // Minimum time needed to complete
	if g.phaseRunning < 5*time.Minute {
		minRunTime = g.phaseRunning / 2
	}

	// Calculate max start time to allow minimum run time
	maxStartOffset := g.phaseRunning - minRunTime
	if maxStartOffset < 0 {
		maxStartOffset = 0
	}

	// Estimate total competitors
	totalCompetitors := len(g.competitors)
	if totalCompetitors == 0 {
		// Estimate based on classes
		for range g.classes {
			totalCompetitors += g.runnersPerClass
		}
	}

	var startInterval time.Duration
	if totalCompetitors > 0 && maxStartOffset > 0 {
		startInterval = maxStartOffset / time.Duration(totalCompetitors)
		// Cap at 2 minutes max, 10 seconds min
		if startInterval > 2*time.Minute {
			startInterval = 2 * time.Minute
		} else if startInterval < 10*time.Second {
			startInterval = 10 * time.Second
		}
	} else {
		startInterval = 30 * time.Second // Default fallback
	}

	return startInterval
}

func (g *Generator) resetSimulation() {
	// Reset start time
	g.startTime = time.Now()

	// Calculate appropriate start interval
	startInterval := g.calculateStartInterval()

	// Reset all competitors
	competitorIndex := 0
	for i := range g.competitors {
		g.competitors[i].Status = "0"
		g.competitors[i].FinishTime = nil
		g.competitors[i].Splits = []models.Split{}

		if g.massStart {
			// Mass start - everyone starts at the same time
			g.competitors[i].StartTime = g.startTime.Add(g.phaseStart)
		} else {
			// Maintain staggered start times after reset
			startOffset := time.Duration(competitorIndex) * startInterval
			g.competitors[i].StartTime = g.startTime.Add(g.phaseStart).Add(startOffset)
			competitorIndex++
		}

		// Regenerate timing for this competitor
		g.generateCompetitorTiming(g.competitors[i].ID, g.competitors[i].Class)
	}
}
