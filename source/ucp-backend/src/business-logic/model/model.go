package model

import (
	core "tah/core/core"
	"time"
)

type SearchRq struct {
	Email     string
	LastName  string
	Phone     string
	LoyaltyID string
}

type LinkIndustryConnectorRq struct {
	AgwUrl        string
	TokenEndpoint string
	ClientId      string
	ClientSecret  string
	BucketArn     string
}

type LinkIndustryConnectorRes struct {
	GlueRoleArn  string
	BucketPolicy string
}

type CreateConnectorCrawlerRq struct {
	GlueRoleArn string
	BucketPath  string
}

type UCPRequest struct {
	ID       string
	SearchRq SearchRq
	Domain   Domain
}

type UCPConfig struct {
	Domains []Domain `json:"domains"`
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

type ResWrapper struct {
	Profiles        []Traveller       `json:"profiles"`
	IngestionErrors []IngestionErrors `json:"ingestionErrors"`
	TotalErrors     int64             `json:"totalErrors"`
	UCPConfig       UCPConfig         `json:"config"`
	Matches         []Match           `json:"matches"`
	Error           core.ResError     `json:"error"`
}

type Match struct {
	ConfidenceScore float64 `json:"confidence"`
	ID              string  `json:"id"`
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	BirthDate       string  `json:"birthDate"`
	PhoneNumber     string  `json:"phone"`
	EmailAddress    string  `json:"email"`
}

type IngestionErrors struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type ObjectMapping struct {
	Name   string         `json:"name"`
	Fields []FieldMapping `json:"fields"`
}

type FieldMapping struct {
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	Indexes     []string `json:"indexes"`
	Searcheable bool     `json:"searchable"`
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
