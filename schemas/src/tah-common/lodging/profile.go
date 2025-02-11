package lodging

import (
	"encoding/json"
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

	Emails    []common.Email   `json:"emails,omitempty"`
	Phones    []common.Phone   `json:"phones,omitempty"`
	Addresses []common.Address `json:"addresses,omitempty"`

	Honorific   string          `json:"honorific,omitempty"`
	FirstName   string          `json:"firstName"`
	MiddleName  string          `json:"middleName,omitempty"`
	LastName    string          `json:"lastName"`
	Gender      string          `json:"gender,omitempty"`
	Pronoun     string          `json:"pronoun,omitempty"`
	DateOfBirth string          `json:"dateOfBirth,omitempty"`
	Language    common.Language `json:"language,omitempty"`
	Nationality common.Country  `json:"nationality,omitempty"`

	JobTitle    string `json:"jobTitle"`
	CompanyName string `json:"parentCompany"`

	LoyaltyPrograms     []LoyaltyProgram            `json:"loyaltyPrograms"`
	ExtendedData        interface{}                 `json:"extendedData,omitempty"`
	AlternateProfileIDs []common.AlternateProfileID `json:"alternateProfileIds,omitempty"`
}

func (gp GuestProfile) MarshalJSON() ([]byte, error) {
	type Alias GuestProfile
	return json.Marshal(&struct {
		Alias
		LastUpdatedOn string `json:"lastUpdatedOn"`
	}{
		Alias:         (Alias)(gp),
		LastUpdatedOn: gp.LastUpdatedOn.Format(core.INGEST_TIMESTAMP_FORMAT),
	})
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
	ExtendedData      interface{}        `json:"extendedData,omitempty"`
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
