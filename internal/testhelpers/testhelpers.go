package testhelpers

import (
	"time"

	"meos-graphics/internal/models"
)

// CreateTestEvent creates a test event with default values
func CreateTestEvent() *models.Event {
	return &models.Event{
		Name:      "Test Competition",
		Organizer: "Test Organizer",
		Start:     time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
	}
}

// CreateTestControl creates a test control
func CreateTestControl(id int, name string) models.Control {
	return models.Control{
		ID:   id,
		Name: name,
	}
}

// CreateTestClass creates a test class with optional radio controls
func CreateTestClass(id int, name string, orderKey int, radioControls ...models.Control) models.Class {
	return models.Class{
		ID:            id,
		Name:          name,
		OrderKey:      orderKey,
		RadioControls: radioControls,
	}
}

// CreateTestClub creates a test club
func CreateTestClub(id int, name, countryCode string) models.Club {
	return models.Club{
		ID:          id,
		Name:        name,
		CountryCode: countryCode,
	}
}

// CreateTestCompetitor creates a test competitor
func CreateTestCompetitor(id int, name string, club models.Club, class models.Class) models.Competitor {
	return models.Competitor{
		ID:        id,
		Name:      name,
		Card:      id * 100,
		Club:      club,
		Class:     class,
		Status:    "0", // Not started
		StartTime: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Splits:    []models.Split{},
	}
}

// CreateFinishedCompetitor creates a test competitor with finish time
func CreateFinishedCompetitor(id int, name string, club models.Club, class models.Class, runTimeDeciseconds int) models.Competitor {
	startTime := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	finishTime := startTime.Add(DecisecondsToTime(runTimeDeciseconds))
	return models.Competitor{
		ID:         id,
		Name:       name,
		Card:       id * 100,
		Club:       club,
		Class:      class,
		Status:     "1", // OK/Finished
		StartTime:  startTime,
		FinishTime: &finishTime,
		Splits:     []models.Split{},
	}
}

// CreateTestSplit creates a test split for a control
func CreateTestSplit(control models.Control, elapsedDeciseconds int, startTime time.Time) models.Split {
	return models.Split{
		Control:     control,
		PassingTime: startTime.Add(DecisecondsToTime(elapsedDeciseconds)),
	}
}

// MOPCompleteXML returns a sample MOPComplete XML for testing
func MOPCompleteXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<MOPComplete nextdifference="abc123">
    <competition date="2024-01-01" organizer="Test Organizer" zerotime="10:00:00">Test Competition</competition>
    <ctrl id="100">Start</ctrl>
    <ctrl id="101">Control 1</ctrl>
    <ctrl id="102">Control 2</ctrl>
    <ctrl id="103">Control 3</ctrl>
    <ctrl id="200">Finish</ctrl>
    <cls id="1" ord="10" radio="101,102,103">Men Elite</cls>
    <cls id="2" ord="20" radio="101,103">Women Elite</cls>
    <org id="1" nat="SWE">Test Club 1</org>
    <org id="2" nat="NOR">Test Club 2</org>
    <cmp id="1" card="12345">
        <base org="1" cls="1" stat="1" st="3600" rt="834">John Doe</base>
        <radio>101,302;102,584;103,723</radio>
        <input it="4434" tstat="1"/>
    </cmp>
    <cmp id="2" card="12346">
        <base org="2" cls="1" stat="0" st="3900" rt="">Jane Smith</base>
    </cmp>
    <cmp id="3" card="12347">
        <base org="1" cls="2" stat="3" st="4200" rt="">Mike Johnson</base>
    </cmp>
</MOPComplete>`
}

// MOPDiffXML returns a sample MOPDiff XML for testing
func MOPDiffXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<MOPDiff nextdifference="def456">
    <cmp id="2" card="12346">
        <base org="2" cls="1" stat="1" st="3900" rt="912">Jane Smith</base>
        <radio>101,334;102,612;103,801</radio>
        <input it="4812" tstat="1"/>
    </cmp>
    <cmp id="4" card="12348">
        <base org="2" cls="2" stat="0" st="4500" rt="">New Competitor</base>
    </cmp>
</MOPDiff>`
}

// InvalidXML returns invalid XML for error testing
func InvalidXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<MOPComplete nextdifference="invalid">
    <competition>Unclosed tag
</MOPComplete>`
}

// EmptyMOPCompleteXML returns a minimal valid MOPComplete XML
func EmptyMOPCompleteXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<MOPComplete nextdifference="empty123">
    <competition date="2024-01-01" organizer="Test" zerotime="10:00:00">Empty Competition</competition>
</MOPComplete>`
}

// MOPCompleteWithMissingFieldsXML returns MOPComplete with missing optional fields
func MOPCompleteWithMissingFieldsXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<MOPComplete nextdifference="missing123">
    <competition date="2024-01-01" zerotime="10:00:00">Missing Organizer</competition>
    <ctrl id="100">Start</ctrl>
    <cls id="1" ord="10">Class Without Radio</cls>
    <org id="1">Club Without Country</org>
    <cmp id="1">
        <base org="1" cls="1" stat="1" st="3600">Competitor Without Card</base>
    </cmp>
</MOPComplete>`
}

// DecisecondsToTime converts deciseconds to time.Duration
func DecisecondsToTime(deciseconds int) time.Duration {
	return time.Duration(deciseconds) * 100 * time.Millisecond
}

// TimeToDeciseconds converts time.Duration to deciseconds
func TimeToDeciseconds(d time.Duration) int {
	return int(d.Milliseconds() / 100)
}
