package admin

import "time"

type BusinessObject struct {
	Name string
}

type AccpRecord struct {
	Name string
}

type Domain struct {
	Name            string          `json:"customerProfileDomain"`
	NObjects        int64           `json:"numberOfObjects"`
	NProfiles       int64           `json:"numberOfProfiles"`
	MatchingEnabled bool            `json:"matchingEnabled"`
	Created         time.Time       `json:"created"`
	LastUpdated     time.Time       `json:"lastUpdated"`
	Mappings        []ObjectMapping `json:"mappings"`
	Integrations    []Integration   `json:"integrations"`
}

type ObjectMapping struct {
	Name      string         `json:"name"`
	Timestamp string         `json:"timestamp"`
	Fields    []FieldMapping `json:"fields"`
}

type FieldMapping struct {
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	Indexes     []string `json:"indexes"`
	Searcheable bool     `json:"searchable"`
	IsTimestamp bool     `json:"isTimestamp"`
}

type Integration struct {
	Source         string    `json:"source"`
	Target         string    `json:"target"`
	Status         string    `json:"status"`
	StatusMessage  string    `json:"statuMessage"`
	LastRun        time.Time `json:"lastRun"`
	LastRunStatus  string    `json:"lastRunStatus"`
	LastRunMessage string    `json:"lastRunMessage"`
	Trigger        string    `json:"trigger"`
}
