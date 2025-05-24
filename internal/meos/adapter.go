package meos

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"meos-graphics/internal/logger"
	"meos-graphics/internal/models"
	"meos-graphics/internal/state"
)

type Adapter struct {
	client            *http.Client
	config            *Config
	state             *state.State
	connected         bool
	mu                sync.RWMutex
	stopChan          chan struct{}
	currentDifference string
}

func NewAdapter(config *Config, appState *state.State) *Adapter {
	return &Adapter{
		client:            &http.Client{Timeout: 3 * time.Second},
		config:            config,
		state:             appState,
		stopChan:          make(chan struct{}),
		currentDifference: "zero",
	}
}

func (a *Adapter) Connect() error {
	// Get current difference without holding the lock during processing
	a.mu.RLock()
	difference := a.currentDifference
	a.mu.RUnlock()

	_, err := a.fetchAndProcessData(difference)
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.connected = true
	a.mu.Unlock()
	return nil
}

func (a *Adapter) StartPolling() error {
	a.mu.RLock()
	if !a.connected {
		a.mu.RUnlock()
		return fmt.Errorf("not connected to MeOS")
	}
	a.mu.RUnlock()

	go func() {
		ticker := time.NewTicker(a.config.PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-a.stopChan:
				return
			case <-ticker.C:
				a.mu.RLock()
				difference := a.currentDifference
				a.mu.RUnlock()

				updated, err := a.fetchAndProcessData(difference)
				if err != nil {
					logger.ErrorLogger.Printf("Error fetching/processing data: %v", err)
				} else if updated {
					logger.DebugLogger.Printf("Data updated from MeOS (difference: %s)", difference)
				}
			}
		}
	}()

	return nil
}

func (a *Adapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.connected {
		close(a.stopChan)
		a.connected = false
	}
	return nil
}

func (a *Adapter) fetchAndProcessData(difference string) (bool, error) {
	protocol := "http"
	if a.config.HTTPS {
		protocol = "https"
	}

	var baseURL string
	if a.config.PortStr == "none" {
		baseURL = fmt.Sprintf("%s://%s/meos", protocol, a.config.Hostname)
	} else {
		baseURL = fmt.Sprintf("%s://%s:%d/meos", protocol, a.config.Hostname, a.config.Port)
	}

	values := url.Values{}
	values.Add("difference", difference)
	_url := baseURL + "?" + values.Encode()

	resp, err := a.client.Get(_url)
	if err != nil {
		return false, fmt.Errorf("failed to connect to MeOS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	return a.processData(data)
}

func (a *Adapter) processData(data []byte) (bool, error) {
	type xmlRoot struct {
		XMLName        xml.Name
		NextDifference string `xml:"nextdifference,attr"`
	}

	var root xmlRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return false, fmt.Errorf("failed to parse XML: %w", err)
	}

	if root.NextDifference == a.currentDifference {
		return false, nil
	}

	logger.DebugLogger.Printf("Processing MeOS data update: %s -> %s", a.currentDifference, root.NextDifference)

	// Update the difference key before processing to avoid holding the lock during processing
	a.mu.Lock()
	a.currentDifference = root.NextDifference
	a.mu.Unlock()

	var source interface{}
	var isMOPComplete bool

	if root.XMLName.Local == "MOPComplete" {
		var mopComplete MOPComplete
		if err := xml.Unmarshal(data, &mopComplete); err != nil {
			return false, fmt.Errorf("failed to parse MOPComplete: %w", err)
		}
		source = mopComplete
		isMOPComplete = true
		logger.InfoLogger.Printf("Received MOPComplete with %d controls, %d classes, %d clubs, %d competitors",
			len(mopComplete.Controls), len(mopComplete.Classes), len(mopComplete.Organizations), len(mopComplete.Competitors))

		if a.state.GetEvent() == nil {
			a.state.SetEvent(&models.Event{})
		}
		event := a.state.GetEvent()
		event.Name = mopComplete.Competition.Name
		event.Organizer = mopComplete.Competition.Organizer
		event.Start = mopComplete.Competition.Time()
		a.state.SetEvent(event)

	} else if root.XMLName.Local == "MOPDiff" {
		var mopDiff MOPDiff
		if err := xml.Unmarshal(data, &mopDiff); err != nil {
			return false, fmt.Errorf("failed to parse MOPDiff: %w", err)
		}
		source = mopDiff
		isMOPComplete = false
		logger.DebugLogger.Printf("Received MOPDiff with %d controls, %d classes, %d clubs, %d competitors",
			len(mopDiff.Controls), len(mopDiff.Classes), len(mopDiff.Organizations), len(mopDiff.Competitors))

		if a.state.GetEvent() == nil {
			a.state.SetEvent(&models.Event{})
		}
		if mopDiff.Competition != nil {
			event := a.state.GetEvent()
			if mopDiff.Competition.Name != "" {
				event.Name = mopDiff.Competition.Name
			}
			if mopDiff.Competition.Organizer != "" {
				event.Organizer = mopDiff.Competition.Organizer
			}
			if !mopDiff.Competition.Time().IsZero() {
				event.Start = mopDiff.Competition.Time()
			}
			a.state.SetEvent(event)
		}
	} else {
		return false, fmt.Errorf("unknown XML root element: %s", root.XMLName.Local)
	}

	// Convert all data before locking global state
	newControls := a.convertControls(source, isMOPComplete)
	newClasses := a.convertClasses(source, isMOPComplete)
	newClubs := a.convertClubs(source, isMOPComplete)
	newCompetitors := a.convertCompetitors(source, isMOPComplete)

	// Get current state for updating
	currentEvent := a.state.GetEvent()
	currentControls := a.state.GetControls()
	currentClasses := a.state.GetClasses()
	currentClubs := a.state.GetClubs()
	currentCompetitors := a.state.GetCompetitors()

	// Update entities
	updatedControls := updateEntities(currentControls, newControls, isMOPComplete)
	updatedClubs := updateEntities(currentClubs, newClubs, isMOPComplete)
	updatedClasses := updateEntities(currentClasses, newClasses, isMOPComplete)
	updatedCompetitors := updateEntities(currentCompetitors, newCompetitors, isMOPComplete)

	// Resolve radio controls for classes
	for i := range updatedClasses {
		resolvedRadioControls := []models.Control{}
		for _, rc := range updatedClasses[i].RadioControls {
			for _, ctrl := range updatedControls {
				if ctrl.ID == rc.ID {
					resolvedRadioControls = append(resolvedRadioControls, ctrl)
					break
				}
			}
		}
		updatedClasses[i].RadioControls = resolvedRadioControls
	}

	// Resolve references for competitors
	for i := range updatedCompetitors {
		// Resolve club reference
		for _, club := range updatedClubs {
			if club.ID == updatedCompetitors[i].Club.ID {
				updatedCompetitors[i].Club = club
				break
			}
		}
		// Resolve class reference
		for _, class := range updatedClasses {
			if class.ID == updatedCompetitors[i].Class.ID {
				updatedCompetitors[i].Class = class
				break
			}
		}
		// Resolve split control references
		for j := range updatedCompetitors[i].Splits {
			for _, ctrl := range updatedControls {
				if ctrl.ID == updatedCompetitors[i].Splits[j].Control.ID {
					updatedCompetitors[i].Splits[j].Control = ctrl
					break
				}
			}
		}
	}

	// Update state atomically and notify listeners
	a.state.UpdateFromMeOS(currentEvent, updatedControls, updatedClasses, updatedClubs, updatedCompetitors)

	return true, nil
}

func updateEntities[T models.Entity](current, updates []T, isComplete bool) []T {
	if isComplete {
		return append([]T{}, updates...)
	}

	result := append([]T{}, current...)

	for _, update := range updates {
		found := false
		for i, existing := range result {
			if existing.GetID() == update.GetID() {
				result[i] = update
				found = true
				break
			}
		}

		if !found {
			result = append(result, update)
		}
	}

	return result
}

func (a *Adapter) convertControls(source interface{}, isComplete bool) []models.Control {
	if isComplete {
		if complete, ok := source.(MOPComplete); ok {
			return a.convertControlList(complete.Controls)
		}
	} else {
		if diff, ok := source.(MOPDiff); ok {
			return a.convertControlList(diff.Controls)
		}
	}
	return []models.Control{}
}

func (a *Adapter) convertClasses(source interface{}, isComplete bool) []models.Class {
	if isComplete {
		if complete, ok := source.(MOPComplete); ok {
			return a.convertClassList(complete.Classes)
		}
	} else {
		if diff, ok := source.(MOPDiff); ok {
			return a.convertClassList(diff.Classes)
		}
	}
	return []models.Class{}
}

func (a *Adapter) convertClubs(source interface{}, isComplete bool) []models.Club {
	if isComplete {
		if complete, ok := source.(MOPComplete); ok {
			return a.convertClubList(complete.Organizations)
		}
	} else {
		if diff, ok := source.(MOPDiff); ok {
			return a.convertClubList(diff.Organizations)
		}
	}
	return []models.Club{}
}

func (a *Adapter) convertCompetitors(source interface{}, isComplete bool) []models.Competitor {
	if isComplete {
		if complete, ok := source.(MOPComplete); ok {
			return a.convertCompetitorList(complete.Competitors)
		}
	} else {
		if diff, ok := source.(MOPDiff); ok {
			return a.convertCompetitorList(diff.Competitors)
		}
	}
	return []models.Competitor{}
}

func (a *Adapter) convertControlList(controls []MOPControl) []models.Control {
	result := make([]models.Control, 0, len(controls))
	for _, ctrl := range controls {
		result = append(result, models.Control{
			ID:   parseInt(ctrl.ID),
			Name: ctrl.Name,
		})
	}
	return result
}

func (a *Adapter) convertClassList(classes []MOPClass) []models.Class {
	result := make([]models.Class, 0, len(classes))
	for _, cls := range classes {
		class := models.Class{
			ID:            parseInt(cls.ID),
			OrderKey:      parseInt(cls.Order),
			Name:          cls.Name,
			RadioControls: []models.Control{},
		}

		if cls.Radio != "" {
			idStrs := strings.Split(cls.Radio, ",")

			for _, idStr := range idStrs {
				idStr = strings.TrimSpace(idStr)
				id := parseInt(idStr)

				// Radio controls will be resolved later after all controls are loaded
				// For now, just store the ID
				class.RadioControls = append(class.RadioControls, models.Control{ID: id})
			}
		}

		result = append(result, class)
	}
	return result
}

func (a *Adapter) convertClubList(orgs []MOPOrg) []models.Club {
	result := make([]models.Club, 0, len(orgs))
	for _, org := range orgs {
		result = append(result, models.Club{
			ID:          parseInt(org.ID),
			CountryCode: org.Nationality,
			Name:        org.Name,
		})
	}
	return result
}

func (a *Adapter) convertCompetitorList(cmps []MOPCompetitor) []models.Competitor {
	result := make([]models.Competitor, 0, len(cmps))
	for _, cmp := range cmps {
		// Store IDs for now, will be resolved later
		clubID := parseInt(cmp.Base.Org)
		club := models.Club{ID: clubID}

		classID := parseInt(cmp.Base.Class)
		class := models.Class{ID: classID}

		competitor := models.Competitor{
			ID:     parseInt(cmp.ID),
			Card:   parseInt(cmp.Card),
			Name:   cmp.Base.Text,
			Status: cmp.Base.Status,
			Club:   club,
			Class:  class,
			Splits: []models.Split{},
		}

		startTimeDeciseconds := cmp.StartTime()

		event := a.state.GetEvent()
		if event != nil && startTimeDeciseconds > 0 {
			eventStart := event.Start
			compStartSeconds := eventStart.Hour()*3600 + eventStart.Minute()*60 + eventStart.Second()
			compStartDeciseconds := compStartSeconds * 10

			if startTimeDeciseconds != compStartDeciseconds {
				seconds := startTimeDeciseconds / 10
				hours := (seconds / 3600) % 24
				minutes := (seconds % 3600) / 60
				secs := seconds % 60

				nanos := (startTimeDeciseconds % 10) * 100000000

				startTime := time.Date(
					eventStart.Year(),
					eventStart.Month(),
					eventStart.Day(),
					hours,
					minutes,
					secs,
					nanos,
					eventStart.Location(),
				)

				competitor.StartTime = startTime

				runningTimeDeciseconds := cmp.RunningTime()
				if runningTimeDeciseconds > 0 {
					runningDuration := decisecondsToTimes(runningTimeDeciseconds)
					finishTime := startTime.Add(runningDuration)
					competitor.FinishTime = &finishTime
				}

				if cmp.Radio != "" {
					splitPairs := strings.Split(cmp.Radio, ";")
					for _, pair := range splitPairs {
						parts := strings.Split(pair, ",")
						if len(parts) != 2 {
							continue
						}

						controlID := parseInt(parts[0])
						splitTimeDeciseconds := parseInt(parts[1])

						// Create control with just ID for now
						control := models.Control{ID: controlID}

						splitDuration := decisecondsToTimes(splitTimeDeciseconds)
						passingTime := startTime.Add(splitDuration)

						split := models.Split{
							Control:     control,
							PassingTime: passingTime,
						}

						competitor.Splits = append(competitor.Splits, split)
					}

					sort.Slice(competitor.Splits, func(i, j int) bool {
						return competitor.Splits[i].PassingTime.Before(competitor.Splits[j].PassingTime)
					})
				}
			}
		}

		result = append(result, competitor)
	}
	return result
}

func decisecondsToTimes(deciseconds int) time.Duration {
	seconds := deciseconds / 10
	nanoseconds := (deciseconds % 10) * 100000000
	return time.Duration(seconds)*time.Second + time.Duration(nanoseconds)*time.Nanosecond
}
