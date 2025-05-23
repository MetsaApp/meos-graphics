package models

import "time"

type Entity interface {
	GetID() int
}

type Event struct {
	Name      string
	Organizer string
	Start     time.Time
}

type Control struct {
	ID   int
	Name string
}

func (c Control) GetID() int {
	return c.ID
}

type Class struct {
	ID            int
	OrderKey      int
	RadioControls []Control
	Name          string
}

func (c Class) GetID() int {
	return c.ID
}

type Club struct {
	ID          int
	CountryCode string
	Name        string
}

func (c Club) GetID() int {
	return c.ID
}

type Competitor struct {
	ID         int
	Card       int
	Club       Club
	Class      Class
	Status     string
	StartTime  time.Time
	FinishTime *time.Time
	Name       string
	Splits     []Split
}

func (c Competitor) GetID() int {
	return c.ID
}

type Split struct {
	Control     Control
	PassingTime time.Time
}