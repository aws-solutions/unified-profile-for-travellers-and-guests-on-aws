package lodging

import (
	common "tah/upt/schemas/src/tah-common/common"
	core "tah/upt/schemas/src/tah-common/core"
	"time"
)

type GuestProfile struct {
	ModelVersion  string    `json:"modelVersion"`
	ID            string    `json:"id"`
	LastUpdatedOn time.Time `json:"lastUpdatedOn"`
	CreatedOn     time.Time `json:"createdOn"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
	CreatedBy     string    `json:"createdBy"`

	Emails    []common.Email   `json:"emails"`
	Phones    []common.Phone   `json:"phones"`
	Addresses []common.Address `json:"addresses"`

	Honorific   string          `json:"honorific"`
	FirstName   string          `json:"firstName"`
	MiddleName  string          `json:"middleName"`
	LastName    string          `json:"lastName"`
	Gender      string          `json:"gender"`
	Pronoun     string          `json:"pronoun"`
	DateOfBirth string          `json:"dateOfBirth"`
	Language    common.Language `json:"language"`
	Nationality common.Country  `json:"nationality"`

	JobTitle    string `json:"jobTitle"`
	CompanyName string `json:"parentCompany"`

	LoyaltyPrograms []LoyaltyProgram `json:"loyaltyPrograms"`
}

////////////////////
// Importable interface: struct implementing this interface can be serizalized for S3 import
//////////////////////

func (b GuestProfile) LastUpdaded() time.Time {
	return b.LastUpdatedOn
}

func (b GuestProfile) DataID() string {
	return b.ID
}

type LoyaltyProgram struct {
	ID                string             `json:"id"`
	ProgramName       string             `json:"programName"`
	Points            core.Float         `json:"points"`
	PointUnit         string             `json:"pointUnit"`
	PointsToNextLevel core.Float         `json:"pointsToNextLevel"`
	Level             string             `json:"level"`
	Joined            time.Time          `json:"joined"`
	Transactions      []common.LoyaltyTx `json:"transactions"`
}

func (p GuestProfile) Version() string {
	return p.ModelVersion
}

////////////////////
// OPERATIONS
///////////////////

type GuestProfilesSearchRs struct {
	GuestProfiles []GuestProfile `json:"guestProfiles"`
}
