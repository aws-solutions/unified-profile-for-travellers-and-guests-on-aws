package air

import (
	"tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/core"
	"time"
)

type PassengerProfile struct {
	ModelVersion  string    `json:"modelVersion"`
	ID            string    `json:"id"`
	LastUpdatedOn time.Time `json:"lastUpdatedOn"`
	CreatedOn     time.Time `json:"createdOn"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
	CreatedBy     string    `json:"createdBy"`
	IsBooker      bool      `json:"isBooker"` //when set to true within a booking, this attribute indicates that this passenger is the booker

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
	IdentityProofs  []IdentityProof  `json:"identityProofs"`
}

type IdentityProof struct {
	Id     string `json:"id"`
	TypeId string `json:"typeId"`
	Data   string `json:"data"`
}

type LoyaltyProgram struct {
	ID               string             `json:"id"`
	ProgramName      string             `json:"programName"`
	Miles            core.Float         `json:"miles"`
	MilesToNextLevel core.Float         `json:"milesToNextLevel"`
	Level            string             `json:"level"`
	Joined           time.Time          `json:"joined"`
	Transactions     []common.LoyaltyTx `json:"transactions"`
}
