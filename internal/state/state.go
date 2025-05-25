package state

import (
	"sync"

	"meos-graphics/internal/logger"
	"meos-graphics/internal/models"
)

type State struct {
	mu              sync.RWMutex
	Event           *models.Event
	Controls        []models.Control
	Classes         []models.Class
	Clubs           []models.Club
	Competitors     []models.Competitor
	updateCallbacks []func()
}

func New() *State {
	return &State{
		Controls:    []models.Control{},
		Classes:     []models.Class{},
		Clubs:       []models.Club{},
		Competitors: []models.Competitor{},
	}
}

func (s *State) Lock() {
	s.mu.Lock()
}

func (s *State) Unlock() {
	s.mu.Unlock()
}

func (s *State) GetEvent() *models.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Event
}

func (s *State) SetEvent(event *models.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Event = event
}

func (s *State) GetControls() []models.Control {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Control, len(s.Controls))
	copy(result, s.Controls)
	return result
}

func (s *State) GetClasses() []models.Class {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Class, len(s.Classes))
	copy(result, s.Classes)
	return result
}

func (s *State) GetClubs() []models.Club {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Club, len(s.Clubs))
	copy(result, s.Clubs)
	return result
}

func (s *State) GetCompetitors() []models.Competitor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Competitor, len(s.Competitors))
	copy(result, s.Competitors)
	return result
}

func (s *State) GetCompetitorsByClass(classID int) []models.Competitor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.Competitor
	for _, comp := range s.Competitors {
		if comp.Class.ID == classID {
			result = append(result, comp)
		}
	}
	return result
}

func (s *State) GetCompetitor(id int) *models.Competitor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, comp := range s.Competitors {
		if comp.ID == id {
			return &comp
		}
	}
	return nil
}

// OnUpdate registers a callback to be called when the state is updated
func (s *State) OnUpdate(callback func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateCallbacks = append(s.updateCallbacks, callback)
}

// notifyUpdate calls all registered update callbacks
func (s *State) notifyUpdate() {
	s.mu.RLock()
	callbacks := make([]func(), len(s.updateCallbacks))
	copy(callbacks, s.updateCallbacks)
	s.mu.RUnlock()

	for _, cb := range callbacks {
		cb()
	}
}

// UpdateFromMeOS updates the state with new data from MeOS and notifies listeners only if data changed
func (s *State) UpdateFromMeOS(event *models.Event, controls []models.Control, classes []models.Class, clubs []models.Club, competitors []models.Competitor) {
	s.mu.Lock()

	// Check if data has actually changed
	hasChanges := false

	// Check event changes
	if !hasChanges && (s.Event == nil && event != nil || s.Event != nil && event == nil) {
		hasChanges = true
	}
	if !hasChanges && s.Event != nil && event != nil {
		if s.Event.Name != event.Name || s.Event.Organizer != event.Organizer || s.Event.Start != event.Start {
			hasChanges = true
		}
	}

	// Check basic length changes
	if !hasChanges && len(s.Controls) != len(controls) {
		hasChanges = true
	}
	if !hasChanges && len(s.Classes) != len(classes) {
		hasChanges = true
	}
	if !hasChanges && len(s.Clubs) != len(clubs) {
		hasChanges = true
	}
	if !hasChanges && len(s.Competitors) != len(competitors) {
		hasChanges = true
	}

	// For competitors, create a map for efficient lookup and check all fields
	if !hasChanges && len(s.Competitors) == len(competitors) {
		// Create map of current competitors for O(1) lookup
		currentMap := make(map[int]*models.Competitor)
		for i := range s.Competitors {
			currentMap[s.Competitors[i].ID] = &s.Competitors[i]
		}

		// Check each new competitor against current state
		for _, newComp := range competitors {
			current, exists := currentMap[newComp.ID]
			if !exists {
				hasChanges = true
				break
			}

			// Check all relevant fields
			if current.Status != newComp.Status ||
				current.Card != newComp.Card ||
				current.Name != newComp.Name ||
				current.StartTime != newComp.StartTime ||
				current.Class.ID != newComp.Class.ID ||
				current.Club.ID != newComp.Club.ID {
				hasChanges = true
				logger.DebugLogger.Printf("Competitor %d changed: status=%v->%v, card=%v->%v, name=%v->%v, start=%v->%v, class=%v->%v, club=%v->%v",
					newComp.ID, current.Status, newComp.Status, current.Card, newComp.Card,
					current.Name, newComp.Name, current.StartTime, newComp.StartTime,
					current.Class.ID, newComp.Class.ID, current.Club.ID, newComp.Club.ID)
				break
			}

			// Check finish time
			if (current.FinishTime == nil) != (newComp.FinishTime == nil) {
				hasChanges = true
				logger.DebugLogger.Printf("Competitor %d finish time changed: %v -> %v", newComp.ID, current.FinishTime, newComp.FinishTime)
				break
			}
			if current.FinishTime != nil && newComp.FinishTime != nil && *current.FinishTime != *newComp.FinishTime {
				hasChanges = true
				logger.DebugLogger.Printf("Competitor %d finish time changed: %v -> %v", newComp.ID, *current.FinishTime, *newComp.FinishTime)
				break
			}

			// Check splits
			if len(current.Splits) != len(newComp.Splits) {
				hasChanges = true
				logger.DebugLogger.Printf("Competitor %d splits count changed: %d -> %d", newComp.ID, len(current.Splits), len(newComp.Splits))
				break
			}
			for j := range newComp.Splits {
				if j >= len(current.Splits) ||
					current.Splits[j].Control.ID != newComp.Splits[j].Control.ID ||
					current.Splits[j].PassingTime != newComp.Splits[j].PassingTime {
					hasChanges = true
					logger.DebugLogger.Printf("Competitor %d split %d changed", newComp.ID, j)
					break
				}
			}
			if hasChanges {
				break
			}
		}
	}

	// Update the state
	s.Event = event
	s.Controls = controls
	s.Classes = classes
	s.Clubs = clubs
	s.Competitors = competitors

	s.mu.Unlock()

	// Only notify if there were changes
	if hasChanges {
		logger.DebugLogger.Println("State changed, notifying update callbacks")
		s.notifyUpdate()
	} else {
		logger.DebugLogger.Println("No state changes detected")
	}
}
