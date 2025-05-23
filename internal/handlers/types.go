package handlers

import "time"

type ClassInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	OrderKey int    `json:"orderKey"`
}

type StartListEntry struct {
	StartNumber int       `json:"startNumber"`
	Name        string    `json:"name"`
	Club        string    `json:"club"`
	StartTime   time.Time `json:"startTime"`
	Card        int       `json:"card"`
}

type ResultEntry struct {
	Position   int         `json:"position"`
	Name       string      `json:"name"`
	Club       string      `json:"club"`
	StartTime  time.Time   `json:"startTime"`
	FinishTime *time.Time  `json:"finishTime,omitempty"`
	Time       *string     `json:"time,omitempty"`
	Status     string      `json:"status"`
	TimeBehind *string     `json:"timeBehind,omitempty"`
	RadioTimes []RadioTime `json:"radioTimes,omitempty"`
}

type RadioTime struct {
	ControlName string `json:"controlName"`
	ElapsedTime string `json:"elapsedTime"`
	SplitTime   string `json:"splitTime"`
}

type SplitTime struct {
	Position    int        `json:"position"`
	Name        string     `json:"name"`
	Club        string     `json:"club"`
	SplitTime   *time.Time `json:"splitTime,omitempty"`
	ElapsedTime *string    `json:"elapsedTime,omitempty"`
	TimeBehind  *string    `json:"timeBehind,omitempty"`
	Status      string     `json:"status"`
}

type SplitStanding struct {
	ControlID   int         `json:"controlId"`
	ControlName string      `json:"controlName"`
	Standings   []SplitTime `json:"standings"`
}

type SplitsResponse struct {
	ClassName string          `json:"className"`
	Splits    []SplitStanding `json:"splits"`
}
