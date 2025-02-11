package common

import "time"

const EXTERNAL_ID_ORIGINATING_SYSTEM_PMS = "pms"
const EXTERNAL_ID_ORIGINATING_SYSTEM_CRS = "crs"
const EXTERNAL_ID_ORIGINATING_SYSTEM_GDS = "gds"
const EXTERNAL_ID_ORIGINATING_SYSTEM_PSS = "pss"
const EXTERNAL_ID_ORIGINATING_SYSTEM_POS = "pos"

type Comment struct {
	Type                 string        `json:"type"` //the type of comment
	Language             Language      `json:"language"`
	Title                string        `json:"title"`
	Text                 string        `json:"text"`
	Context              []ContextItem `json:"context"` //context in which the comment was added (property code, airport check-in...)
	CreatedDateTime      time.Time     `json:"createdDateTime"`
	CreatedBy            string        `json:"createdBy"`
	LastModifiedDateTime time.Time     `json:"lastModifiedDateTime"`
	LastModifiedBy       string        `json:"lastModifiedBy"`
	IsTravellerViewable  bool          `json:"isTravellerViewable"` //is the comment vieweable by the traveller
}

type ContextItem struct {
	ContextType  string `json:"contextType"`
	ContextValue string `json:"contextValue"`
}

type Language struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ExternalID struct {
	ID                string `json:"id"`                //value of teh ID
	IdName            string `json:"IdName"`            //Name of the ID in teh originating system (ConfirmationNumber, RecordLocator...)
	OriginatingSystem string `json:"originatingSystem"` //Originaing system
}

type CancellationReason struct {
	Reason  string  `json:"reason"`
	Comment Comment `json:"comment"`
}
