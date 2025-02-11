package air

import (
	"encoding/json"
	"tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/core"
	"time"
)

var BOOKING_STATUS_CANCELED = "canceled"
var BOOKING_STATUS_CONFIRMED = "confirmed"
var BOOKING_STATUS_TENTATIVE = "tentative"
var BAGGAGE_TYPE_CARRY_ON = "carryOn"
var BAGGAGE_TYPE_CHECKED_BAG = "checkedBag"
var BAGGAGE_TYPE_OVERWEIGHT_BAG = "overweightBag"
var BAGGAGE_TYPE_OVERSIZED_BAG = "oversizedBag"
var BAGGAGE_TYPE_SPORTING_EQUIPMENT = "sportingEquipment"
var SEAT_ZONE_LIE_FLAT = "lieFlat"
var SEAT_ZONE_BUSINESS = "business"
var SEAT_ZONE_FIRST = "first"
var SEAT_ZONE_EXTRA_LEG_ROOM = "extraLegRoom"
var SEAT_ZONE_FRONT_AIRCRAFT = "FrontAircraft"
var SEAT_ZONE_ECONOMY = "economy"
var SEAT_ZONE_ECONOMY_PLUS = "economyPlus"
var CHANGE_TYPE_FLIGHT_CHANGE = "flightChange"
var CHANGE_TYPE_FLIGHT_CANCELLATION = "flightCancellation"
var CHANGE_TYPE_NAME_CHANGE = "nameChange"
var PRIORITY_BOARDING = "priorityBoarding"
var PRIORITY_SECURITY = "prioritySecurity"
var PRIORITY_BAG_DROP = "priorityBagDrop"
var PRIORITY_CHECKIN = "priorityCheckin"
var PRIORITY_BAG_RETURN = "priorityBagReturn"
var OTHER_ANCILLARY_TYPE_INFANT_CHARGE = "infantCharge"
var OTHER_ANCILLARY_TYPE_UNACCOMPANIED_MINOR = "unaccompaniedMinor"

type Booking struct {
	ObjectVersion       int                       `json:"objectVersion"`
	ModelVersion        string                    `json:"modelVersion"` //Version of the AWS travel and hospitality model
	ID                  string                    `json:"id"`           //Unique ID of the booking
	ExternalID          []common.ExternalID       `json:"externalIds,omitempty"`
	LastUpdatedOn       time.Time                 `json:"lastUpdatedOn"`
	CreatedOn           time.Time                 `json:"createdOn"`
	LastUpdatedBy       string                    `json:"lastUpdatedBy"`
	CreatedBy           string                    `json:"createdBy"`
	CreationChannelId   string                    `json:"creationChannelId"`
	LastUpdateChannelId string                    `json:"lastUpdateChannelId"`
	Price               BookingPrice              `json:"price"`
	Itinerary           Itinerary                 `json:"itinerary"`
	Return              Itinerary                 `json:"return"`
	PassengerInfo       PassengerInfo             `json:"passengerInfo"`
	PaymentInformation  common.PaymentInformation `json:"paymentInformation"`
	AncillaryServices   AncillaryServices         `json:"ancillaryServices"`
	Promocode           string                    `json:"promocode,omitempty"`
	MainEmail           string                    `json:"email"`
	MainPhone           string                    `json:"phone"`
	DayOfTravelEmail    string                    `json:"dayOfTravelEmail"`
	DayOfTravelPhone    string                    `json:"dayOfTravelPhone"`
	Status              string                    `json:"status"`
	ExtendedData        interface{}               `json:"extendedData,omitempty"`
}

func (booking Booking) MarshalJSON() ([]byte, error) {
	type Alias Booking
	return json.Marshal(&struct {
		Alias
		LastUpdatedOn string `json:"lastUpdatedOn"`
	}{
		Alias:         (Alias)(booking),
		LastUpdatedOn: booking.LastUpdatedOn.Format(core.INGEST_TIMESTAMP_FORMAT),
	})
}

type PassengerInfo struct {
	Passengers []PassengerProfile `json:"passengers"`
}

type Flight struct {
	DepartureDate string `json:"departureDate"`
	FlightNumber  string `json:"flightNumber"`
}

type Itinerary struct {
	From          string          `json:"from"`
	To            string          `json:"to"`
	DepartureDate string          `json:"departureDate"`
	DepartureTime string          `json:"departureTime"`
	ArrivalDate   string          `json:"arrivalDate"`
	ArrivalTime   string          `json:"arrivalTime"`
	Duration      string          `json:"duration,omitempty"`
	Segments      []FlightSegment `json:"segments,omitempty"`
	Inventory     Inventory       `json:"inventory"`
	Status        string          `json:"status"`
}

type BookingPrice struct {
	PriceItems []PricePerPassengerPerSegments `json:"priceItems,omitempty"`
	GrandTotal core.Float                     `json:"grandTotal"`
}

type PricePerPassengerPerSegments struct {
	SegmentRank    int    `json:"segmenRank"`     //refers to the rank of the rank of the segment
	PassengerIndex int    `json:"passengerIndex"` //refers to the index of the passenger in the passenger list
	FareClass      string `json:"fareClass"`
	Price          Price  `json:"price"`
}

type Price struct {
	Total core.Float `json:"total"`
	Base  core.Float `json:"base"`
	Fees  []Fee      `json:"fees,omitempty"`
}

type Fee struct {
	Amount core.Float `json:"amount"`
	Type   string     `json:"type"`
}

type FlightSegment struct {
	Rank          int       `json:"rank"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	DepartureDate string    `json:"departureDate"`
	DepartureTime string    `json:"departureTime"`
	ArrivalDate   string    `json:"arrivalDate"`
	ArrivalTime   string    `json:"arrivalTime"`
	FlightNumber  string    `json:"flightNumber"`
	Inventory     Inventory `json:"inventory"`
	Status        string    `json:"status"`
}

type ByRank []FlightSegment

func (a ByRank) Len() int           { return len(a) }
func (a ByRank) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRank) Less(i, j int) bool { return a[i].Rank < a[j].Rank }

type Inventory struct {
	TotalSeats           int                     `json:"totalSeats"`
	InventoryByFareClass []InventoryForFareClass `json:"inventoryByFareClass,omitempty"`
}

type InventoryForFareClass struct {
	FareClass string `json:"fareClass"`
	Number    int    `json:"number"`
}

type AncillaryServices struct {
	AncillaryServiceBaggage      []AncillaryServiceBaggage      `json:"ancillaryServiceBaggage,omitempty"`
	AncillaryServiceSeat         []AncillaryServiceSeat         `json:"ancillaryServiceSeat,omitempty"`
	AncillaryServiceFlightChange []AncillaryServiceFlightChange `json:"ancillaryServiceFlightChange,omitempty"`
	AncillaryServiceOther        []AncillaryServiceOther        `json:"ancillaryServiceOther,omitempty"`
	AncillaryServicePriority     []AncillaryServicePriority     `json:"ancillaryServicePriority,omitempty"`
	Price                        AncillaryServicePrice          `json:"price"`
}

type AncillaryServiceOther struct {
	ID                  string                `json:"id"`
	FlightNumber        string                `json:"flightNumber"`
	DepartureDate       string                `json:"departureDate"`
	OtherAncilliaryType string                `json:"otherAncilliaryType"`
	Price               AncillaryServicePrice `json:"price"`
	ExtendedData        interface{}           `json:"extendedData,omitempty"`
}

type AncillaryServicePriority struct {
	ID                  string                `json:"id"`
	FlightNumber        string                `json:"flightNumber"`
	DepartureDate       string                `json:"departureDate"`
	TransactionID       string                `json:"transactionId"`
	PriorityServiceType string                `json:"priorityServiceType"`
	LoungeAccess        bool                  `json:"loungeAccess"`
	Price               AncillaryServicePrice `json:"price"`
	ExtendedData        interface{}           `json:"extendedData,omitempty"`
}

type AncillaryServiceFlightChange struct {
	ID            string                `json:"id"`
	FlightNumber  string                `json:"flightNumber"`
	DepartureDate string                `json:"departureDate"`
	TransactionID string                `json:"transactionId"`
	ChangeType    string                `json:"changeType"`
	Price         AncillaryServicePrice `json:"price"`
	ExtendedData  interface{}           `json:"extendedData,omitempty"`
}

type AncillaryServiceSeat struct {
	ID               string                `json:"id"`
	FlightNumber     string                `json:"flightNumber"`
	DepartureDate    string                `json:"departureDate"`
	PaxIndex         int                   `json:"paxIndex"`
	SeatNumber       string                `json:"seatNumber"`
	SeatZone         string                `json:"seatZone"`
	Price            AncillaryServicePrice `json:"price"`
	NeighborFreeSeat bool                  `json:"neighborFreeSeat"`
	UpgradeAuction   bool                  `json:"upgradeAuction"`
	Available        bool                  `json:"available"`
	ExtendedData     interface{}           `json:"extendedData,omitempty"`
}

type AncillaryServiceBaggage struct {
	ID                       string                `json:"id"`
	FlightNumber             string                `json:"flightNumber"`
	DepartureDate            string                `json:"departureDate"`
	PaxIndex                 int                   `json:"paxIndex"`
	BaggageType              string                `json:"baggageType"`
	Quantity                 int                   `json:"quantity"`
	Weight                   core.Float            `json:"weight"`
	Dimensions               BagDimensions         `json:"dimensions"`
	PriorityBagDrop          bool                  `json:"priorityBagDrop"`
	PriorityBagReturn        bool                  `json:"priorityBagReturn"`
	LostBagInsurance         bool                  `json:"lostBagInsurance"`
	ValuableBaggageInsurance bool                  `json:"valuableBaggageInsurance"`
	HandsFreeBaggage         bool                  `json:"handsFreeBaggage"`
	Price                    AncillaryServicePrice `json:"price"`
	MaxAllowed               int                   `json:"maxAllowed"`
	ExtendedData             interface{}           `json:"extendedData,omitempty"`
}

type AncillaryServicePrice struct {
	ServiceName       string      `json:"serviceName"`
	AmountBeforeTaxes core.Float  `json:"amount,omitempty"`
	Currency          string      `json:"currency"`
	Taxes             []Fee       `json:"taxes,omitempty"`
	Fees              []Fee       `json:"fees,omitempty"`
	Total             core.Float  `json:"total"`
	ExtendedData      interface{} `json:"extendedData,omitempty"`
}

type BagDimensions struct {
	Length int `json:"length"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
