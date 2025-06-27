package service

import (
	"fmt"
	"sort"
	"time"

	"meos-graphics/internal/i18n"
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
	Name      string `json:"name"`
	Club      string `json:"club"`
	StartTime string `json:"startTime"` // Formatted as HH:mm
}

// ResultEntry represents a competitor's result
type ResultEntry struct {
	Name        string `json:"name"`
	Club        string `json:"club"`
	Status      string `json:"status"`
	RunningTime string `json:"runningTime,omitempty"` // Formatted duration string
	Position    int    `json:"position,omitempty"`
	Difference  string `json:"difference,omitempty"` // Formatted duration from leader
}

// SplitTime represents a split time at a control
type SplitTime struct {
	Position       int     `json:"position,omitempty"`
	Name           string  `json:"name"`
	Club           string  `json:"club"`
	ElapsedTime    *string `json:"elapsedTime,omitempty"`
	TimeDifference *string `json:"timeDifference,omitempty"`
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
	for _, comp := range competitors {
		startList = append(startList, StartListEntry{
			Name:      comp.Name,
			Club:      comp.Club.Name,
			StartTime: comp.StartTime.Format("15:04"),
		})
	}

	return startList, nil
}

// GetResults returns the results for a specific class
func (s *Service) GetResults(classID int) ([]ResultEntry, error) {
	competitors := s.state.GetCompetitorsByClass(classID)
	currentTime := time.Now()

	var results []ResultEntry
	var finishedCompetitors []models.Competitor
	var dnfCompetitors []models.Competitor
	var dnsCompetitors []models.Competitor
	var runningCompetitors []models.Competitor
	var waitingCompetitors []models.Competitor

	// Categorize competitors
	for _, comp := range competitors {
		switch comp.Status {
		case "1": // Approved/Finished
			if comp.FinishTime != nil {
				finishedCompetitors = append(finishedCompetitors, comp)
			}
		case "3": // Miss Punch
			dnfCompetitors = append(dnfCompetitors, comp)
		case "4": // Not Finished (DNF)
			dnfCompetitors = append(dnfCompetitors, comp)
		case "5": // Disqualified
			dnfCompetitors = append(dnfCompetitors, comp)
		case "6": // Max. Time
			dnfCompetitors = append(dnfCompetitors, comp)
		case "20": // Not Started
			if !comp.StartTime.IsZero() && currentTime.Before(comp.StartTime) {
				waitingCompetitors = append(waitingCompetitors, comp)
			} else {
				dnsCompetitors = append(dnsCompetitors, comp)
			}
		case "21": // Cancelled
			dnsCompetitors = append(dnsCompetitors, comp)
		case "99": // Not Competing
			dnsCompetitors = append(dnsCompetitors, comp)
		case "0": // Unknown - need to determine based on start time
			if !comp.StartTime.IsZero() {
				if currentTime.Before(comp.StartTime) {
					// Has start time but hasn't started yet
					waitingCompetitors = append(waitingCompetitors, comp)
				} else if comp.FinishTime == nil {
					// Started but not finished
					runningCompetitors = append(runningCompetitors, comp)
				}
			} else {
				// No start time - treat as DNS
				dnsCompetitors = append(dnsCompetitors, comp)
			}
		default:
			// For any other status codes, check if they're running based on time
			if !comp.StartTime.IsZero() && currentTime.After(comp.StartTime) && comp.FinishTime == nil {
				runningCompetitors = append(runningCompetitors, comp)
			}
		}
	}

	// Sort finished competitors by time, then by name for ties
	sort.Slice(finishedCompetitors, func(i, j int) bool {
		if finishedCompetitors[i].FinishTime == nil || finishedCompetitors[j].FinishTime == nil {
			return false
		}
		timeI := finishedCompetitors[i].FinishTime.Sub(finishedCompetitors[i].StartTime)
		timeJ := finishedCompetitors[j].FinishTime.Sub(finishedCompetitors[j].StartTime)
		if timeI == timeJ {
			// For tied times, sort alphabetically by name
			return finishedCompetitors[i].Name < finishedCompetitors[j].Name
		}
		return timeI < timeJ
	})

	// Build results with positions
	position := 1
	var winnerTime time.Duration

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

		// Calculate position considering ties
		if i > 0 {
			prevRunTime := finishedCompetitors[i-1].FinishTime.Sub(finishedCompetitors[i-1].StartTime)
			if runTime != prevRunTime {
				// Not a tie, update position to current index + 1
				position = i + 1
			}
			// If times are equal, keep the same position
		}

		result := ResultEntry{
			Name:        comp.Name,
			Club:        comp.Club.Name,
			Status:      i18n.GetInstance().GetStatusDescription("1"),
			RunningTime: timeStr,
			Position:    position,
		}
		if timeBehind != nil {
			result.Difference = *timeBehind
		}
		results = append(results, result)
	}

	// Add DNF competitors
	for _, comp := range dnfCompetitors {
		var status string
		switch comp.Status {
		case "3":
			status = i18n.GetInstance().GetStatusDescription(comp.Status) // Miss Punch
		case "4":
			status = i18n.GetInstance().GetStatusDescription(comp.Status) // Did Not Finish
		case "5":
			status = i18n.GetInstance().GetStatusDescription(comp.Status) // Disqualified
		case "6":
			status = i18n.GetInstance().GetStatusDescription(comp.Status) // Over Time (Max. Time)
		default:
			status = i18n.GetInstance().GetStatusDescription("4")
		}
		results = append(results, ResultEntry{
			Name:   comp.Name,
			Club:   comp.Club.Name,
			Status: status,
		})
	}

	// Add running competitors (sorted by start time)
	sort.Slice(runningCompetitors, func(i, j int) bool {
		return runningCompetitors[i].StartTime.Before(runningCompetitors[j].StartTime)
	})
	for _, comp := range runningCompetitors {
		results = append(results, ResultEntry{
			Name:   comp.Name,
			Club:   comp.Club.Name,
			Status: i18n.GetInstance().GetStatusDescription("1001"),
		})
	}

	// Add waiting competitors (not yet started)
	for _, comp := range waitingCompetitors {
		results = append(results, ResultEntry{
			Name:   comp.Name,
			Club:   comp.Club.Name,
			Status: i18n.GetInstance().GetStatusDescription("1000"),
		})
	}

	// Add DNS competitors (Did Not Start - set by organizers)
	for _, comp := range dnsCompetitors {
		var status string
		switch comp.Status {
		case "21":
			status = i18n.GetInstance().GetStatusDescription(comp.Status)
		case "99":
			status = i18n.GetInstance().GetStatusDescription(comp.Status)
		default:
			status = i18n.GetInstance().GetStatusDescription("20")
		}
		results = append(results, ResultEntry{
			Name:   comp.Name,
			Club:   comp.Club.Name,
			Status: status,
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
		return nil, fmt.Errorf("class not found")
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

		// Sort by elapsed time, then by name for ties
		sort.Slice(splitEntries, func(i, j int) bool {
			if splitEntries[i].elapsed == splitEntries[j].elapsed {
				// For tied times, sort alphabetically by name
				return splitEntries[i].competitor.Name < splitEntries[j].competitor.Name
			}
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

			// Calculate position considering ties
			if i > 0 {
				if entry.elapsed != splitEntries[i-1].elapsed {
					// Not a tie, update position to current index + 1
					position = i + 1
				}
				// If times are equal, keep the same position
			}

			standing.Standings = append(standing.Standings, SplitTime{
				Position:       position,
				Name:           entry.competitor.Name,
				Club:           entry.competitor.Club.Name,
				ElapsedTime:    &elapsedStr,
				TimeDifference: timeBehind,
			})
		}

		// Add competitors without this split (but not for finish)
		if control.ID != -1 {
			for _, comp := range competitors {
				if comp.Status == "3" { // Miss Punch
					found := false
					for _, entry := range splitEntries {
						if entry.competitor.ID == comp.ID {
							found = true
							break
						}
					}
					if !found {
						standing.Standings = append(standing.Standings, SplitTime{
							Name: comp.Name,
							Club: comp.Club.Name,
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
