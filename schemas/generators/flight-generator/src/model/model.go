package model

import (
	"encoding/json"
	"tah/upt/source/tah-core/core"
)

var TIME_FORMAT = "15:04"
var DATE_FORMAT = "2006-01-02"
var DATETIME_FORMAT = "2006-01-02T15:04:05"

// Requests and Response to deserialize teh Json Payload into.
// ideally this should be a generic struct in teh core package with a specialised body but
// the design needs more thoughts so we will include it here for the moment
type RqWrapper struct {
	SubFunction      string           `json:"subFunction"`
	Id               string           `json:"id"`    //ID to be used for GET request
	OrgID            string           `json:"orgID"` //ID to be used for GET request
	ShoppingRequest  ShoppingRequest  `json:"request"`
	UpdateRequest    UpdateRequest    `json:"updateRequest"`
	PricingRequest   PricingRequest   `json:"pricingRequest"`
	AncillaryRequest AncillaryRequest `json:"ancillaryRequest"`
}

type ResWrapper struct {
	Error       core.ResError    `json:"error"`
	SubFunction string           `json:"subFunction"`
	Response    ShoppingResponse `json:"response"`
	TxID        string           `json:"transactionId"`
}

type AncillaryRequest struct {
}

type ShoppingRequest struct {
	OfferID     string   `json:"offerID"`
	FareClasses []string `json:"fareClasses"`
	Departure   string   `json:"departureDate"`
	Return      string   `json:"return"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	NPax        int      `json:"nPax"`
}
type PricingRequest struct {
	Offers []FlightOffer `json:"offers"`
	NPax   int           `json:"nPax"`
}

type UpdateRequest struct {
	Offers []FlightOffer `json:"offers"`
}

type ShoppingResponse struct {
	Offers            []FlightOffer     `json:"offers"`
	AncillaryServices AncillaryServices `json:"ancillaries"`
}

type FlightOffer struct {
	Price     OfferPrice `json:"price"`
	Inventory Inventory  `json:"availableSeats"`
	Itinerary Itinerary  `json:"itinerary"`
	Return    Itinerary  `json:"return"`
}
type Itinerary struct {
	From          string          `json:"from"`
	To            string          `json:"to"`
	DepartureDate string          `json:"departureDate"`
	DepartureTime string          `json:"departureTime"`
	ArrivalDate   string          `json:"arrivalDate"`
	ArrivalTime   string          `json:"arrivalTime"`
	Duration      string          `json:"duration"`
	Segments      []FlightSegment `json:"segments"`
	Price         OfferPrice      `json:"price"`
	Inventory     Inventory       `json:"inventory"`
	Status        string          `json:"status"`
}

type OfferPrice struct {
	Lower            float64             `json:"lower"`
	Currency         string              `json:"currency"`
	PriceByFareClass []PriceForFareClass `json:"priceByFareClass"`
	Price            Price               `json:"price"`
}

type PriceForFareClass struct {
	FareClass string `json:"fareClass"`
	Price     Price  `json:"price"`
}

type Price struct {
	Total      float64 `json:"total"`
	Base       float64 `json:"base"`
	Fees       []Fee   `json:"fees"`
	GrandTotal float64 `json:"grandTotal"`
}

type Fee struct {
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
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

type Flight struct {
	DepartureDate string `json:"departureDate"`
	FlightNumber  string `json:"flightNumber"`
}

type Inventory struct {
	TotalSeats           int                     `json:"totalSeats"`
	InventoryByFareClass []InventoryForFareClass `json:"inventoryByFareClass"`
}

type InventoryForFareClass struct {
	FareClass string `json:"fareClass"`
	Number    int    `json:"number"`
}

type ReferentialData struct {
	Tz map[string]string `json:"tz"`
}

type AncillaryServices struct {
	AncillaryServiceBaggage      []AncillaryServiceBaggage      `json:"ancillaryServiceBaggage"`
	AncillaryServiceSeat         []AncillaryServiceSeat         `json:"ancillaryServiceSeat"`
	AncillaryServiceFlightChange []AncillaryServiceFlightChange `json:"ancillaryServiceFlightChange"`
	AncillaryServiceOther        []AncillaryServiceOther        `json:"ancillaryServiceOther"`
	AncillaryServicePriority     []AncillaryServicePriority     `json:"ancillaryServicePriority"`
	Price                        AncillaryServicePrice          `json:"price"`
}

var OTHER_ANCILLARY_TYPE_INFANT_CHARGE = "infantCharge"
var OTHER_ANCILLARY_TYPE_UNACCOMPANIED_MINOR = "unaccompaniedMinor"

type AncillaryServiceOther struct {
	ID                  string                `json:"id"`
	OtherAncilliaryType string                `json:"otherAncilliaryType"`
	Price               AncillaryServicePrice `json:"price"`
}

var PRIORITY_BOARDING = "priorityBoarding"
var PRIORITY_SECURITY = "prioritySecurity"
var PRIORITY_BAG_DROP = "priorityBagDrop"
var PRIORITY_CHECKIN = "priorityCheckin"
var PRIORITY_BAG_RETURN = "priorityBagReturn"

type AncillaryServicePriority struct {
	ID                  string                `json:"id"`
	PriorityServiceType string                `json:"priorityServiceType"`
	LoungeAccess        bool                  `json:"loungeAccess"`
	Price               AncillaryServicePrice `json:"price"`
}

var CHANGE_TYPE_FLIGHT_CHANGE = "flightChange"
var CHANGE_TYPE_FLIGHT_CANCELLATION = "flightCancellation"
var CHANGE_TYPE_NAME_CHANGE = "nameChange"

type AncillaryServiceFlightChange struct {
	ID         string                `json:"id"`
	ChangeType string                `json:"changeType"`
	Price      AncillaryServicePrice `json:"price"`
}

var SEAT_ZONE_LIE_FLAT = "lieFlat"
var SEAT_ZONE_BUSINESS = "business"
var SEAT_ZONE_FIRST = "first"
var SEAT_ZONE_EXTRA_LEG_ROOM = "extraLegRoom"
var SEAT_ZONE_FRONT_AIRCRAFT = "FrontAircraft"
var SEAT_ZONE_ECONOMY = "economy"
var SEAT_ZONE_ECONOMY_PLUS = "economyPlus"

type AncillaryServiceSeat struct {
	ID               string                `json:"id"`
	PaxIndex         int                   `json:"paxIndex"`
	SeatNumber       string                `json:"seatNumber"`
	SeatZone         string                `json:"seatZone"`
	Price            AncillaryServicePrice `json:"price"`
	NeighborFreeSeat bool                  `json:"neighborFreeSeat"`
	UpgradeAuction   bool                  `json:"upgradeAuction"`
	Available        bool                  `json:"available"`
}

var BAGGAGE_TYPE_CARRY_ON = "carryOn"
var BAGGAGE_TYPE_CHECKED_BAG = "checkedBag"
var BAGGAGE_TYPE_OVERWEIGHT_BAG = "overweightBag"
var BAGGAGE_TYPE_OVERSIZED_BAG = "oversizedBag"
var BAGGAGE_TYPE_SPORTING_EQUIPMENT = "sportingEquipment"

type AncillaryServiceBaggage struct {
	ID                       string                `json:"id"`
	PaxIndex                 int                   `json:"paxIndex"`
	BaggageType              string                `json:"baggageType"`
	Quantity                 int                   `json:"quantity"`
	Weight                   float64               `json:"weight"`
	Dimentions               BagDimentions         `json:"dimentions"`
	PriorityBagDrop          bool                  `json:"priorityBagDrop"`
	PriorityBagReturn        bool                  `json:"priorityBagReturn"`
	LotBagInsurance          bool                  `json:"lotBagInsurance"`
	ValuableBaggageInsurance bool                  `json:"valuableBaggageInsurance"`
	HandsFreeBaggage         bool                  `json:"handsFreeBaggage"`
	Price                    AncillaryServicePrice `json:"price"`
	MaxAllowed               int                   `json:"maxAllowed"`
}

type AncillaryServicePrice struct {
	ServiceName       string  `json:"serviceName"`
	AmountBeforeTaxes float64 `json:"amount"`
	Currency          string  `json:"currency"`
	Taxes             []Fee   `json:"taxes"`
	Fees              []Fee   `json:"fees"`
	Total             float64 `json:"total"`
}

type BagDimentions struct {
	Length int `json:"length"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (res *ResWrapper) AddError(err error) {
	res.Error = core.BuildResError(err)
}

func (res *ResWrapper) SetTx(tx core.Transaction) {
	res.TxID = tx.TransactionID
}

func (p ResWrapper) Decode(dec json.Decoder) (error, core.JSONObject) {
	return dec.Decode(&p), p
}
