package meos

import (
	"encoding/xml"
	"fmt"
	"time"
)

type MOPComplete struct {
	XMLName        xml.Name `xml:"MOPComplete"`
	NextDifference string   `xml:"nextdifference,attr,omitempty"`
	Competition    MOPCompetition
	Controls       []MOPControl    `xml:"ctrl"`
	Classes        []MOPClass      `xml:"cls"`
	Organizations  []MOPOrg        `xml:"org"`
	Teams          []MOPTeam       `xml:"tm"`
	Competitors    []MOPCompetitor `xml:"cmp"`
}

type MOPDiff struct {
	XMLName        xml.Name `xml:"MOPDiff"`
	NextDifference string   `xml:"nextdifference,attr"`
	Competition    *MOPCompetition
	Controls       []MOPControl    `xml:"ctrl"`
	Classes        []MOPClass      `xml:"cls"`
	Organizations  []MOPOrg        `xml:"org"`
	Teams          []MOPTeam       `xml:"tm"`
	Competitors    []MOPCompetitor `xml:"cmp"`
}

type MOPCompetition struct {
	XMLName   xml.Name `xml:"competition"`
	Date      string   `xml:"date,attr"`
	Organizer string   `xml:"organizer,attr"`
	Homepage  string   `xml:"homepage,attr"`
	ZeroTime  string   `xml:"zerotime,attr"`
	Name      string   `xml:",chardata"`
}

type MOPControl struct {
	XMLName xml.Name `xml:"ctrl"`
	ID      string   `xml:"id,attr"`
	Name    string   `xml:",chardata"`
}

type MOPClass struct {
	XMLName xml.Name `xml:"cls"`
	ID      string   `xml:"id,attr"`
	Order   string   `xml:"ord,attr"`
	Radio   string   `xml:"radio,attr"`
	Name    string   `xml:",chardata"`
}

type MOPOrg struct {
	XMLName     xml.Name `xml:"org"`
	ID          string   `xml:"id,attr"`
	Nationality string   `xml:"nat,attr,omitempty"`
	Name        string   `xml:",chardata"`
}

type MOPTeam struct {
	XMLName xml.Name `xml:"tm"`
	ID      string   `xml:"id,attr"`
	Base    MOPBase
	Results string `xml:"r"`
}

type MOPBase struct {
	XMLName     xml.Name `xml:"base"`
	Org         string   `xml:"org,attr"`
	Class       string   `xml:"cls,attr"`
	Status      string   `xml:"stat,attr"`
	StartTime   string   `xml:"st,attr"`
	RunningTime string   `xml:"rt,attr"`
	Bib         string   `xml:"bib,attr"`
	Text        string   `xml:",chardata"`
}

type MOPCompetitor struct {
	XMLName xml.Name `xml:"cmp"`
	ID      string   `xml:"id,attr"`
	Card    string   `xml:"card,attr"`
	Base    MOPBase
	Radio   string `xml:"radio"`
	Input   MOPInput
}

type MOPInput struct {
	XMLName    xml.Name `xml:"input"`
	InputTime  string   `xml:"it,attr"`
	TimeStatus string   `xml:"tstat,attr"`
}

func (c *MOPCompetitor) StartTime() int {
	return parseInt(c.Base.StartTime)
}

func (c *MOPCompetitor) RunningTime() int {
	return parseInt(c.Base.RunningTime)
}

func parseInt(s string) int {
	var i int
	// Ignore error - returns 0 on parse failure which is acceptable for our use case
	_, _ = fmt.Sscanf(s, "%d", &i)
	return i
}

func (c *MOPCompetition) Time() time.Time {
	dateTimeStr := fmt.Sprintf("%sT%s", c.Date, c.ZeroTime)
	parsedTime, _ := time.Parse("2006-01-02T15:04:05", dateTimeStr)
	return parsedTime
}
