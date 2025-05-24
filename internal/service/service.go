package service

import (
	"fmt"
	"sort"
	"time"

	"meos-graphics/internal/models"
	"meos-graphics/internal/state"
)

// Service contains the business logic for competition data
type Service struct {
	state *state.State
}

// New creates a new service instance
func New(appState *state.State) *Service {
	return &Service{
		state: appState,
	}
}

// ClassInfo represents basic class information
type ClassInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	OrderKey int    `json:"orderKey"`
}

// StartListEntry represents an entry in the start list
type StartListEntry struct {
	StartNumber int       `json:"startNumber"`
	Name        string    `json:"name"`
	Club        string    `json:"club"`
	StartTime   time.Time `json:"startTime"`
	Card        int       `json:"card"`
}

// RadioTime represents a radio control passing time
type RadioTime struct {
	ControlName string `json:"controlName"`
	ElapsedTime string `json:"elapsedTime"`
	SplitTime   string `json:"splitTime"`
}

// ResultEntry represents a competitor's result
type ResultEntry struct {
	Position       int         `json:"position,omitempty"`
	Name           string      `json:"name"`
	Club           string      `json:"club"`
	StartTime      time.Time   `json:"startTime"`
	FinishTime     *time.Time  `json:"finishTime,omitempty"`
	Time           *string     `json:"time,omitempty"`
	Status         string      `json:"status"`
	TimeDifference *string     `json:"timeDifference,omitempty"`
	RadioTimes     []RadioTime `json:"radioTimes,omitempty"`
}

// SplitTime represents a split time at a control
type SplitTime struct {
	Position       int        `json:"position,omitempty"`
	Name           string     `json:"name"`
	Club           string     `json:"club"`
	SplitTime      *time.Time `json:"splitTime,omitempty"`
	ElapsedTime    *string    `json:"elapsedTime,omitempty"`
	TimeDifference *string    `json:"timeDifference,omitempty"`
	Status         string     `json:"status"`
}

// SplitStanding represents standings at a control
type SplitStanding struct {
	ControlID   int         `json:"controlId"`
	ControlName string      `json:"controlName"`
	Standings   []SplitTime `json:"standings"`
}

// SplitsResponse represents the full splits response
type SplitsResponse struct {
	ClassName string          `json:"className"`
	Splits    []SplitStanding `json:"splits"`
}

// GetClasses returns all competition classes sorted by order key
func (s *Service) GetClasses() []ClassInfo {
	classes := s.state.GetClasses()

	var classInfos []ClassInfo
	for _, class := range classes {
		classInfos = append(classInfos, ClassInfo{
			ID:       class.ID,
			Name:     class.Name,
			OrderKey: class.OrderKey,
		})
	}

	// Sort by OrderKey
	sort.Slice(classInfos, func(i, j int) bool {
		return classInfos[i].OrderKey < classInfos[j].OrderKey
	})

	return classInfos
}

// GetStartList returns the start list for a specific class
func (s *Service) GetStartList(classID int) ([]StartListEntry, error) {
	competitors := s.state.GetCompetitorsByClass(classID)
	if len(competitors) == 0 {
		return []StartListEntry{}, nil
	}

	// Sort by start time first
	sort.Slice(competitors, func(i, j int) bool {
		return competitors[i].StartTime.Before(competitors[j].StartTime)
	})

	var startList []StartListEntry
	for i, comp := range competitors {
		startList = append(startList, StartListEntry{
			StartNumber: i + 1,
			Name:        comp.Name,
			Club:        comp.Club.Name,
			StartTime:   comp.StartTime,
			Card:        comp.Card,
		})
	}

	return startList, nil
}

// GetResults returns the results for a specific class
func (s *Service) GetResults(classID int) ([]ResultEntry, error) {
	competitors := s.state.GetCompetitorsByClass(classID)

	var results []ResultEntry
	var finishedCompetitors []models.Competitor
	var dnfCompetitors []models.Competitor
	var dnsCompetitors []models.Competitor
	var runningCompetitors []models.Competitor

	// Categorize competitors
	for _, comp := range competitors {
		switch comp.Status {
		case "1": // OK/Finished
			if comp.FinishTime != nil {
				finishedCompetitors = append(finishedCompetitors, comp)
			}
		case "3", "4": // DNF or MP (MisPunch)
			dnfCompetitors = append(dnfCompetitors, comp)
		case "0": // DNS/Not started
			dnsCompetitors = append(dnsCompetitors, comp)
		case "2": // Running
			runningCompetitors = append(runningCompetitors, comp)
		default:
			// Other status - skip
		}
	}

	// Sort finished competitors by time
	sort.Slice(finishedCompetitors, func(i, j int) bool {
		if finishedCompetitors[i].FinishTime == nil || finishedCompetitors[j].FinishTime == nil {
			return false
		}
		timeI := finishedCompetitors[i].FinishTime.Sub(finishedCompetitors[i].StartTime)
		timeJ := finishedCompetitors[j].FinishTime.Sub(finishedCompetitors[j].StartTime)
		return timeI < timeJ
	})

	// Build results with positions
	position := 1
	var winnerTime time.Duration

	// Get class radio controls
	var radioControls []models.Control
	classes := s.state.GetClasses()
	for _, class := range classes {
		if class.ID == classID {
			radioControls = class.RadioControls
			break
		}
	}

	for i, comp := range finishedCompetitors {
		runTime := comp.FinishTime.Sub(comp.StartTime)
		timeStr := formatDuration(runTime)

		var timeBehind *string
		if i == 0 {
			winnerTime = runTime
		} else {
			behind := runTime - winnerTime
			behindStr := "+" + formatDuration(behind)
			timeBehind = &behindStr
		}

		// Build radio times
		var radioTimes []RadioTime
		var prevTime time.Duration

		for _, rc := range radioControls {
			for _, split := range comp.Splits {
				if split.Control.ID != rc.ID {
					continue
				}
				elapsed := split.PassingTime.Sub(comp.StartTime)
				splitTime := elapsed - prevTime
				radioTimes = append(radioTimes, RadioTime{
					ControlName: split.Control.Name,
					ElapsedTime: formatDuration(elapsed),
					SplitTime:   formatDuration(splitTime),
				})
				prevTime = elapsed
				break
			}
		}

		results = append(results, ResultEntry{
			Position:       position,
			Name:           comp.Name,
			Club:           comp.Club.Name,
			StartTime:      comp.StartTime,
			FinishTime:     comp.FinishTime,
			Time:           &timeStr,
			Status:         "OK",
			TimeDifference: timeBehind,
			RadioTimes:     radioTimes,
		})
		position++
	}

	// Add DNF competitors
	for _, comp := range dnfCompetitors {
		status := "DNF"
		if comp.Status == "4" {
			status = "MP" // Mispunch
		}
		results = append(results, ResultEntry{
			Name:      comp.Name,
			Club:      comp.Club.Name,
			StartTime: comp.StartTime,
			Status:    status,
		})
	}

	// Add running competitors (sorted by start time)
	sort.Slice(runningCompetitors, func(i, j int) bool {
		return runningCompetitors[i].StartTime.Before(runningCompetitors[j].StartTime)
	})
	for _, comp := range runningCompetitors {
		results = append(results, ResultEntry{
			Name:      comp.Name,
			Club:      comp.Club.Name,
			StartTime: comp.StartTime,
			Status:    "Running",
		})
	}

	// Add DNS competitors
	for _, comp := range dnsCompetitors {
		results = append(results, ResultEntry{
			Name:      comp.Name,
			Club:      comp.Club.Name,
			StartTime: comp.StartTime,
			Status:    "DNS",
		})
	}

	return results, nil
}

// GetSplits returns split standings for a specific class
func (s *Service) GetSplits(classID int) (*SplitsResponse, error) {
	// Get class info
	var className string
	var radioControls []models.Control
	classes := s.state.GetClasses()
	for _, class := range classes {
		if class.ID == classID {
			className = class.Name
			radioControls = class.RadioControls
			break
		}
	}

	if className == "" {
		return nil, fmt.Errorf("Class not found")
	}

	competitors := s.state.GetCompetitorsByClass(classID)

	response := &SplitsResponse{
		ClassName: className,
		Splits:    []SplitStanding{},
	}

	// Process each control (including finish)
	allControls := radioControls
	allControls = append(allControls, models.Control{ID: -1, Name: "Finish"})

	for _, control := range allControls {
		standing := SplitStanding{
			ControlID:   control.ID,
			ControlName: control.Name,
			Standings:   []SplitTime{},
		}

		var splitEntries []struct {
			competitor models.Competitor
			splitTime  *time.Time
			elapsed    time.Duration
		}

		// Collect split times for this control
		for _, comp := range competitors {
			if comp.Status == "0" {
				continue // Skip DNS
			}

			var splitTime *time.Time
			var elapsed time.Duration

			if control.ID == -1 { // Finish
				if comp.FinishTime != nil {
					splitTime = comp.FinishTime
					elapsed = splitTime.Sub(comp.StartTime)
					splitEntries = append(splitEntries, struct {
						competitor models.Competitor
						splitTime  *time.Time
						elapsed    time.Duration
					}{comp, splitTime, elapsed})
				}
			} else {
				// Find split for this control
				for _, split := range comp.Splits {
					if split.Control.ID == control.ID {
						splitTime = &split.PassingTime
						elapsed = splitTime.Sub(comp.StartTime)
						splitEntries = append(splitEntries, struct {
							competitor models.Competitor
							splitTime  *time.Time
							elapsed    time.Duration
						}{comp, splitTime, elapsed})
						break
					}
				}
			}
		}

		// Sort by elapsed time
		sort.Slice(splitEntries, func(i, j int) bool {
			return splitEntries[i].elapsed < splitEntries[j].elapsed
		})

		// Build standings with positions
		position := 1
		var leaderTime time.Duration

		for i, entry := range splitEntries {
			elapsedStr := formatDuration(entry.elapsed)

			var timeBehind *string
			if i == 0 {
				leaderTime = entry.elapsed
			} else {
				behind := entry.elapsed - leaderTime
				behindStr := "+" + formatDuration(behind)
				timeBehind = &behindStr
			}

			standing.Standings = append(standing.Standings, SplitTime{
				Position:       position,
				Name:           entry.competitor.Name,
				Club:           entry.competitor.Club.Name,
				SplitTime:      entry.splitTime,
				ElapsedTime:    &elapsedStr,
				TimeDifference: timeBehind,
				Status:         "OK",
			})
			position++
		}

		// Add competitors without this split (but not for finish)
		if control.ID != -1 {
			for _, comp := range competitors {
				if comp.Status == "3" { // DNF
					found := false
					for _, entry := range splitEntries {
						if entry.competitor.ID == comp.ID {
							found = true
							break
						}
					}
					if !found {
						standing.Standings = append(standing.Standings, SplitTime{
							Name:   comp.Name,
							Club:   comp.Club.Name,
							Status: "DNF",
						})
					}
				}
			}
		}

		response.Splits = append(response.Splits, standing)
	}

	return response, nil
}

// formatDuration formats a duration with deciseconds
func formatDuration(d time.Duration) string {
	// Convert to deciseconds to avoid floating point precision issues
	totalDeciseconds := d.Milliseconds() / 100

	hours := totalDeciseconds / 36000
	minutes := (totalDeciseconds % 36000) / 600
	seconds := (totalDeciseconds % 600) / 10
	deciseconds := totalDeciseconds % 10

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%d", hours, minutes, seconds, deciseconds)
	}
	return fmt.Sprintf("%d:%02d.%d", minutes, seconds, deciseconds)
}
