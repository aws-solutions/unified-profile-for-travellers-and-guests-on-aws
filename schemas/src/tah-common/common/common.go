package common

import (
	"encoding/json"
	core "tah/upt/schemas/src/tah-common/core"
	"time"
)

const EXTERNAL_ID_ORIGINATING_SYSTEM_PMS = "pms"
const EXTERNAL_ID_ORIGINATING_SYSTEM_CRS = "crs"
const EXTERNAL_ID_ORIGINATING_SYSTEM_GDS = "gds"
const EXTERNAL_ID_ORIGINATING_SYSTEM_PSS = "pss"
const EXTERNAL_ID_ORIGINATING_SYSTEM_POS = "pos"

type Comment struct {
	Type                 string        `json:"type,omitempty"` //the type of comment
	Language             Language      `json:"language,omitempty"`
	Title                string        `json:"title,omitempty"`
	Text                 string        `json:"text"`
	Context              []ContextItem `json:"context,omitempty"` //context in which the comment was added (property code, airport check-in...)
	CreatedDateTime      time.Time     `json:"createdDateTime"`
	CreatedBy            string        `json:"createdBy"`
	LastModifiedDateTime time.Time     `json:"lastModifiedDateTime,omitempty"`
	LastModifiedBy       string        `json:"lastModifiedBy,omitempty"`
	IsTravelerViewable   bool          `json:"isTravelerViewable,omitempty"` //is the comment vieweable by the traveller
}

type ContextItem struct {
	ContextType  string `json:"contextType,omitempty"`
	ContextValue string `json:"contextValue,omitempty"`
}

type Language struct {
	Code string `json:"code"`
	Name string `json:"name,omitempty"`
}

type ExternalID struct {
	ID                string `json:"id"`                          //value of teh ID
	IdName            string `json:"IdName,omitempty"`            //Name of the ID in teh originating system (ConfirmationNumber, RecordLocator...)
	OriginatingSystem string `json:"originatingSystem,omitempty"` //Originaing system
}

type CancellationReason struct {
	Reason  string  `json:"reason"`
	Comment Comment `json:"comment,omitempty"`
}

type AlternateProfileID struct {
	Name          string    `json:"name"`                  //	Name of the AltProfileID, e.g. "airline"
	Value         string    `json:"value"`                 //	Actual value of the ID
	Description   string    `json:"description,omitempty"` //	Optional more descriptive name
	LastUpdatedOn time.Time `json:"lastUpdatedOn"`
}

func (altProfId AlternateProfileID) MarshalJSON() ([]byte, error) {
	type Alias AlternateProfileID
	return json.Marshal(&struct {
		Alias
		LastUpdatedOn string `json:"lastUpdatedOn"`
	}{
		Alias:         (Alias)(altProfId),
		LastUpdatedOn: altProfId.LastUpdatedOn.Format(core.INGEST_TIMESTAMP_FORMAT),
	})
}
