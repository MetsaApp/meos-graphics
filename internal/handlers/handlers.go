package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"meos-graphics/internal/models"
	"meos-graphics/internal/state"
)

type Handler struct {
	state *state.State
}

func New(appState *state.State) *Handler {
	return &Handler{
		state: appState,
	}
}

func (h *Handler) GetClasses(c *gin.Context) {
	classes := h.state.GetClasses()

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

	c.JSON(http.StatusOK, classInfos)
}

func (h *Handler) GetStartList(c *gin.Context) {
	var classID int
	if _, err := fmt.Sscanf(c.Param("classId"), "%d", &classID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
		return
	}

	competitors := h.state.GetCompetitorsByClass(classID)

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

	c.JSON(http.StatusOK, startList)
}

func (h *Handler) GetResults(c *gin.Context) {
	var classID int
	if _, err := fmt.Sscanf(c.Param("classId"), "%d", &classID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
		return
	}

	competitors := h.state.GetCompetitorsByClass(classID)

	var results []ResultEntry
	var finishedCompetitors []models.Competitor
	var dnfCompetitors []models.Competitor
	var dnsCompetitors []models.Competitor

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
		default:
			// Running or other status - skip
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
	classes := h.state.GetClasses()
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
				if split.Control.ID == rc.ID {
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
		}

		results = append(results, ResultEntry{
			Position:   position,
			Name:       comp.Name,
			Club:       comp.Club.Name,
			StartTime:  comp.StartTime,
			FinishTime: comp.FinishTime,
			Time:       &timeStr,
			Status:     "OK",
			TimeBehind: timeBehind,
			RadioTimes: radioTimes,
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

	// Add DNS competitors
	for _, comp := range dnsCompetitors {
		results = append(results, ResultEntry{
			Name:      comp.Name,
			Club:      comp.Club.Name,
			StartTime: comp.StartTime,
			Status:    "DNS",
		})
	}

	c.JSON(http.StatusOK, results)
}

func (h *Handler) GetSplits(c *gin.Context) {
	var classID int
	if _, err := fmt.Sscanf(c.Param("classId"), "%d", &classID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
		return
	}

	// Get class info
	var className string
	var radioControls []models.Control
	classes := h.state.GetClasses()
	for _, class := range classes {
		if class.ID == classID {
			className = class.Name
			radioControls = class.RadioControls
			break
		}
	}

	if className == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
		return
	}

	competitors := h.state.GetCompetitorsByClass(classID)

	response := SplitsResponse{
		ClassName: className,
		Splits:    []SplitStanding{},
	}

	// Process each control (including finish)
	allControls := append(radioControls, models.Control{ID: -1, Name: "Finish"})

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
				Position:    position,
				Name:        entry.competitor.Name,
				Club:        entry.competitor.Club.Name,
				SplitTime:   entry.splitTime,
				ElapsedTime: &elapsedStr,
				TimeBehind:  timeBehind,
				Status:      "OK",
			})
			position++
		}

		// Add competitors without this split
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

		response.Splits = append(response.Splits, standing)
	}

	c.JSON(http.StatusOK, response)
}

func formatDuration(d time.Duration) string {
	totalSeconds := d.Seconds()
	hours := int(totalSeconds) / 3600
	minutes := (int(totalSeconds) % 3600) / 60
	seconds := int(totalSeconds) % 60
	deciseconds := int((totalSeconds - float64(int(totalSeconds))) * 10)

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%d", hours, minutes, seconds, deciseconds)
	}
	return fmt.Sprintf("%d:%02d.%d", minutes, seconds, deciseconds)
}
