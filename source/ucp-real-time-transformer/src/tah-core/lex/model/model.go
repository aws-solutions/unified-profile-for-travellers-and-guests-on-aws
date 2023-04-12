package model

var SLOT_SELECTION_STRATEGY_TOP_RESOLUTION = "TOP_RESOLUTION"

/**
{
    "enumerationValues": [
        {
            "value": "Long Beach Resort, Mauritius"
        },
        {
            "value": "Intercontinental Hotel Chicago Magnificent Mile",
            "synonyms": [
                "Intercontinental in Chicago"
            ]
        },
        {
            "value": "Hyatt Regency Miami",
            "synonyms": [
                "Hyatt Miami",
                "Hyatt in Miami",
                "Regency Miami"
            ]
        }
    ],
    "valueSelectionStrategy": "TOP_RESOLUTION",
    "name": "CloudrackPropertiesdev",
    "description": "List of Bookable properties (Shoulld be dynamically extracted from CRS at build time)"
}
*/
type SlotConfig struct {
	EnumerationValues      []SlotValue `json:"enumerationValues"`
	ValueSelectionStrategy string      `json:"valueSelectionStrategy"`
	Name                   string      `json:"name"`
	Description            string      `json:"description"`
	Checksum               string      `json:"checksum,omitempty"`
}
type SlotValue struct {
	Value    string   `json:"value"`
	Synonyms []string `json:"synonyms"`
}
