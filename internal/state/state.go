package state

import (
	"sync"

	"meos-graphics/internal/models"
)

type State struct {
	mu          sync.RWMutex
	Event       *models.Event
	Controls    []models.Control
	Classes     []models.Class
	Clubs       []models.Club
	Competitors []models.Competitor
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
