package state

import (
	"sync"

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
	
	// Simple change detection - could be optimized further
	if !hasChanges && (s.Event == nil && event != nil || s.Event != nil && event == nil) {
		hasChanges = true
	}
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
	
	// For competitors, check if any have different status or finish times
	if !hasChanges && len(s.Competitors) == len(competitors) {
		for i := range competitors {
			if i >= len(s.Competitors) {
				hasChanges = true
				break
			}
			if s.Competitors[i].Status != competitors[i].Status {
				hasChanges = true
				break
			}
			if (s.Competitors[i].FinishTime == nil) != (competitors[i].FinishTime == nil) {
				hasChanges = true
				break
			}
			if len(s.Competitors[i].Splits) != len(competitors[i].Splits) {
				hasChanges = true
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
		s.notifyUpdate()
	}
}
