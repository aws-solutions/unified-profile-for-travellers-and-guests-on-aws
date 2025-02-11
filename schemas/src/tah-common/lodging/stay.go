package lodging

import (
	"encoding/json"
	common "tah/upt/schemas/src/tah-common/common"
	core "tah/upt/schemas/src/tah-common/core"
	"time"
)

type Stay struct {
	ObjectVersion       int                         `json:"objectVersion"`
	ModelVersion        string                      `json:"modelVersion"`
	ID                  string                      `json:"id"`
	ReservationID       string                      `json:"bookingId"`
	GuestID             string                      `json:"guestId"`
	LastUpdatedOn       time.Time                   `json:"lastUpdatedOn"`
	CreatedOn           time.Time                   `json:"createdOn"`
	LastUpdatedBy       string                      `json:"lastUpdatedBy"`
	CreatedBy           string                      `json:"createdBy"`
	Currency            common.Currency             `json:"currency"`
	FirstName           string                      `json:"firstName"`
	LastName            string                      `json:"lastName"`
	Email               string                      `json:"email"`
	Phone               string                      `json:"phone"`
	StartDate           string                      `json:"startDate"`
	HotelCode           string                      `json:"hotelCode"`
	Revenue             []StayRevenueItem           `json:"revenue"`
	AlternateProfileIDs []common.AlternateProfileID `json:"alternateProfileIds,omitempty"`
	ExtendedData        interface{}                 `json:"extendedData,omitempty"`
}

func (stay Stay) MarshalJSON() ([]byte, error) {
	type Alias Stay
	return json.Marshal(&struct {
		Alias
		LastUpdatedOn string `json:"lastUpdatedOn"`
		CreatedOn     string `json:"createdOn"`
	}{
		Alias:         (Alias)(stay),
		LastUpdatedOn: stay.LastUpdatedOn.Format(core.INGEST_TIMESTAMP_FORMAT),
		CreatedOn:     stay.CreatedOn.Format(core.INGEST_TIMESTAMP_FORMAT),
	})
}

type StayRevenueItem struct {
	Type         string          `json:"type"`
	Description  string          `json:"description"`
	Currency     common.Currency `json:"currency"`
	Amount       core.Float      `json:"amount"`
	Date         time.Time       `json:"date"`
	ExtendedData interface{}     `json:"extendedData,omitempty"`
}

////////////////////
// Importable interface: struct implementing this interface can be serizalized for S3 import
//////////////////////

func (s Stay) DataID() string {
	return s.ID
}

func (s Stay) LastUpdaded() time.Time {
	return s.LastUpdatedOn
}

////////////////////
// OPERATIONS
///////////////////

type StaySearchRs struct {
	Stays []Stay `json:"stays"`
}
