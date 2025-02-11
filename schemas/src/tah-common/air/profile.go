package air

import (
	"encoding/json"
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
	Language    common.Language `json:"language"`
	Nationality common.Country  `json:"nationality"`

	JobTitle    string `json:"jobTitle"`
	CompanyName string `json:"parentCompany"`

	LoyaltyPrograms     []LoyaltyProgram            `json:"loyaltyPrograms,omitempty"`
	IdentityProofs      []IdentityProof             `json:"identityProofs,omitempty"`
	AlternateProfileIDs []common.AlternateProfileID `json:"alternateProfileIds,omitempty"`
	ExtendedData        interface{}                 `json:"extendedData,omitempty"`
}

func (gp PassengerProfile) MarshalJSON() ([]byte, error) {
	type Alias PassengerProfile
	return json.Marshal(&struct {
		Alias
		LastUpdatedOn string `json:"lastUpdatedOn"`
	}{
		Alias:         (Alias)(gp),
		LastUpdatedOn: gp.LastUpdatedOn.Format(core.INGEST_TIMESTAMP_FORMAT),
	})
}

type IdentityProof struct {
	Id     string `json:"id"`
	TypeId string `json:"typeId"`
	Data   string `json:"data"`
}

type LoyaltyProgram struct {
	ID                       string             `json:"id"`
	ProgramName              string             `json:"programName"`
	Miles                    core.Float         `json:"miles"`
	MilesToNextLevel         core.Float         `json:"milesToNextLevel"`
	Level                    string             `json:"level"`
	Joined                   time.Time          `json:"joined"`
	Transactions             []common.LoyaltyTx `json:"transactions,omitempty"`
	EnrollmentSource         string             `json:"enrollmentSource"`
	Currency                 string             `json:"currency"`
	Amount                   core.Float         `json:"amount"`
	AccountStatus            string             `json:"accountStatus"`
	ReasonForClose           string             `json:"reasonForClose"`
	LanguagePreference       string             `json:"languagePreference"`
	DisplayPreference        string             `json:"displayPreference"`
	MealPreference           string             `json:"mealPreference"`
	SeatPreference           string             `json:"seatPreference"`
	HomeAirport              string             `json:"homeAirport"`
	DateTimeFormatPreference string             `json:"dateTimeFormatPreference"`
	CabinPreference          string             `json:"cabinPreference"`
	FareTypePreference       string             `json:"fareTypePreference"`
	ExpertMode               bool               `json:"expertMode"`
	PrivacyIndicator         string             `json:"privacyIndicator"`
	CarPreferenceVendor      string             `json:"carPreferenceVendor"`
	CarPreferenceType        string             `json:"carPreferenceType"`
	SpecialAccommodation1    string             `json:"specialAccommodation1"`
	SpecialAccommodation2    string             `json:"specialAccommodation2"`
	SpecialAccommodation3    string             `json:"specialAccommodation3"`
	SpecialAccommodation4    string             `json:"specialAccommodation4"`
	SpecialAccommodation5    string             `json:"specialAccommodation5"`
	MarketingOptIns          string             `json:"marketingOptIns"`
	RenewDate                string             `json:"renewDate"`
	NextBillAmount           core.Float         `json:"nextBillAmount"`
	ClearEnrollDate          string             `json:"clearEnrollDate"`
	ClearRenewDate           string             `json:"clearRenewDate"`
	ClearTierLevel           string             `json:"clearTierLevel"`
	ClearNextBillAmount      core.Float         `json:"clearNextBillAmount"`
	ClearIsActive            bool               `json:"clearIsActive"`
	ClearAutoRenew           bool               `json:"clearAutoRenew"`
	ClearHasBiometrics       bool               `json:"clearHasBiometrics"`
	ClearHasPartnerPricing   bool               `json:"clearHasPartnerPricing"`
	TsaType                  string             `json:"tsaType"`
	TsaSeqNum                string             `json:"tsaSeqNum"`
	TsaNumber                string             `json:"tsaNumber"`
	ExtendedData             interface{}        `json:"extendedData,omitempty"`
}
