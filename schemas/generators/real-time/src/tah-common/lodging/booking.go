package lodging

import (
	common "tah/upt/schemas/src/tah-common/common"
	core "tah/upt/schemas/src/tah-common/core"
	"time"
)

var SEARCH_KEY_START_DATE string = "startDate" //HTTP QueryParam forr Start Date in search
var SEARCH_KEY_END_DATE string = "endDate"
var SEARCH_KEY_CREATION_DATE string = "created"
var SEARCH_KEY_LAST_UPDATED_AFTER string = "lastUpdatedAfter"
var SEARCH_KEY_LAST_UPDATED_BEFORE string = "lastUpdatedBefore"
var SEARCH_KEY_HOTEL_CODE string = "hotelCode"
var SEARCH_KEY_CHAIN_CODE string = "chainCode"

var BOOKING_CHANNEL_HEADER = "booking-channel-id" //HTTP Header what the channel ID for the booking is stored

var STATUS_PENDING string = "pending"                           //the reservation has been requestted by the inventory is in courtesy hold but not guaranteed yet,
var STATUS_CONFIRMED string = "confirmed"                       //the reservatiion is confirmed and the inventory is held. Reservation could still be autocancelled based on guarantee conditions
var STATUS_CANCELLED string = "cancelled"                       //reservation is cancelled, inventoory is no longer held
var STATUS_CANCELLATION_PENDING string = "cancellation_pending" //Cancellation has been requested  but not executed yet.
var STATUS_IGNORED string = "ignored"                           //reservation was pending but has not been confired. courtesy inventoory hold expired
var STATUS_NO_SHOW string = "noshow"                            //guest did not show up for the reservation
var STATUS_WAITLISTED string = "waitlisted"                     //guest has been wiatlisted on a product that is noo lonoger available
var STATUS_IN_HOUSE string = "in_house"                         //Guest has checked in
var STATUS_CHECKED_OUT string = "checked_out"                   //Guest has checked out,
var STATUS_UNKNOWN string = "unknown"                           //unknown status. indicate a potential issue in the system.

// //////////////////////
// BUSINESS OBJECT
// ///////////////////////
type Booking struct {
	ObjectVersion       int                       `json:"objectVersion"`
	ModelVersion        string                    `json:"modelVersion"` //Version of the AWS travel and hospitality model
	ID                  string                    `json:"id"`           //Unique ID of the booking
	ExternalID          []common.ExternalID       `json:"externalIds"`
	LastUpdatedOn       time.Time                 `json:"lastUpdatedOn"`
	CreatedOn           time.Time                 `json:"createdOn"`
	LastUpdatedBy       string                    `json:"lastUpdatedBy"`
	CreatedBy           string                    `json:"createdBy"`
	CreationChannelId   string                    `json:"creationChannelId"`
	LastUpdateChannelId string                    `json:"lastUpdateChannelId"`
	HotelCode           string                    `json:"hotelCode"`
	NNights             int64                     `json:"nNights"`
	NGuests             int64                     `json:"nGuests"`
	StartDate           string                    `json:"startDate"`
	EndDate             string                    `json:"endDate"`
	Holder              GuestProfile              `json:"holder"`
	PaymentInformation  common.PaymentInformation `json:"paymentInformation"`
	ContextId           string                    `json:"contextId"`
	GroupId             string                    `json:"groupId"`
	Status              string                    `json:"status"`
	Currency            common.Currency           `json:"currency"`
	CancelReason        common.CancellationReason `json:"cancelReason"`
	Segments            []BookingSegment          `json:"segments"`
	Comments            []common.Comment          `json:"comments"`
}

type BookingSegment struct {
	ID                 string                    `json:"id"`
	HotelCode          string                    `json:"hotelCode"`
	NNights            int64                     `json:"nNights"`
	NGuests            int64                     `json:"nGuests"`
	StartDate          string                    `json:"startDate"`
	Products           []Product                 `json:"products"`
	Holder             GuestProfile              `json:"holder"`
	AdditionalGuests   []GuestProfile            `json:"additionalGuests"`
	PaymentInformation common.PaymentInformation `json:"paymentInformation"`
	GroupId            string                    `json:"groupId"`
	Status             string                    `json:"status"`
	Price              BookingPrice              `json:"price"`
}

type BookingPrice struct {
	PricePerNight                   []AmountPerNightItem        `json:"pricePerNight"`
	TaxePerNight                    []AmountPerNightItem        `json:"taxePerNight"`
	PricePerStay                    []AmountPerProductItem      `json:"pricePerStay"`
	TaxPerStay                      []AmountPerProductItem      `json:"taxPerStay"`
	BusinessRules                   []BusinessRule              `json:"businessRules"`
	TotalPricePerNightBeforeTaxes   []TotalAmountPerNightItem   `json:"totalPricePerNightBeforeTaxes"`
	TotalPricePerNightAfterTaxes    []TotalAmountPerNightItem   `json:"totalPricePerNightAfterTaxes"`
	TotalPricePerProductBeforeTaxes []TotalAmountPerProductItem `json:"totalPricePerProductBeforeTaxes"`
	TotalPricePerProductAfterTaxes  []TotalAmountPerProductItem `json:"totalPricePerProductAfterTaxes"`
	TotalBeforeTax                  core.Float                  `json:"totalBeforeTax"`
	TotalAfterTax                   core.Float                  `json:"totalAfterTax"`
}

type TotalAmountPerNightItem struct {
	Date     string          `json:"date"`
	Label    string          `json:"label"` //label of the amount for display
	Amount   core.Float      `json:"amount"`
	Currency common.Currency `json:"currency"`
}
type TotalAmountPerProductItem struct {
	ProductID          string          `json:"productId"`
	ProductDescription string          `json:"productDescription"`
	Label              string          `json:"label"` //label of the amount for display
	Amount             core.Float      `json:"amount"`
	Currency           common.Currency `json:"currency"`
}

type AmountPerNightItem struct {
	Date             string                 `json:"date"`
	Label            string                 `json:"label"` //label of the amount for display
	AmountPerProduct []AmountPerProductItem `json:"amountPerProduct"`
	Currency         common.Currency        `json:"currency"`
}

type AmountPerProductItem struct {
	ProductID          string          `json:"productId"`
	ProductDescription string          `json:"productDescription"`
	Amount             core.Float      `json:"amount"`
	Label              string          `json:"label"` //label of the amount for display
	Currency           common.Currency `json:"currency"`
}

type BusinessRule struct {
	Id        string             `json:"pricePerNight"`
	AppliesTo BusinessRuleScope  `json:"appliesTo"`
	Effect    BusinessRuleEffect `json:"effect"`
	On        []DateRange        `json:"on"`
	NotOn     []DateRange        `json:"notOn"`
}
type BusinessRuleScope struct {
	Selectors BusinessObjectSelector `json:"selectors"`
}
type BusinessObjectSelector struct {
	ObjectType string `json:"objectType"`
	Field      string `json:"field"`
	Op         string `json:"op"`
	Value      string `json:"value"`
}

type BusinessRuleEffect struct {
	Type            string     `json:"type"`
	Available       bool       `json:"available"`
	PriceImpact     core.Float `json:"priceImpact"`
	PriceUpdateUnit string     `json:"priceUpdateUnit"`
}
type DateRange struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Dow  []string `json:"dow"`
}

// ////////////////////////
// Product ID shouold be built using the folloowing logic:
// <ROOM_TYPE_CODE>-<RATE_PLAN_CODE>[|<ATTR_CODE1>-<ATTR_CODE2>...][|<ADD_ON_CODE1>-<ADD_ON_CODE2>...]
// where al codes are alphabetically sorted
// /////////////////////
type Product struct {
	ID          string      `json:"id"`          //unique identifier fo the product
	Description string      `json:"description"` //unique identifier fo the product
	RoomType    RoomType    `json:"roomType"`    //Room type code in the system
	RatePlan    RatePlan    `json:"ratePlan"`    //Room type code in the system
	Attributes  []Attribute `json:"attributes"`  //for systems supporting product attribut pricing, a lits of attribute codes
	AddOns      []AddOn     `json:"addOns"`      //a list of add-ons to the room product (Non room inventory)
}

type RoomType struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type RatePlan struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type Attribute struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type AddOn struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

////////////////////
// Utility functions
//////////////////////

// returns check In date
func (b Booking) CheckIn() time.Time {
	ci, _ := time.Parse(core.DATE_FORMAT, b.StartDate)
	return ci
}

// returns checkout date
func (b Booking) CheckOut() time.Time {
	return b.CheckIn().AddDate(0, 0, int(b.NNights))
}

// returns true if booking is in the past
func (b Booking) IsPast() bool {
	now := time.Now()
	checkIn, _ := time.Parse(core.DATE_FORMAT, b.StartDate)
	return checkIn.Before(now)
}

////////////////////
// Importable interface: struct implementing this interface can be serizalized for S3 import
//////////////////////

func (b Booking) DataID() string {
	return b.ID
}

func (b Booking) LastUpdaded() time.Time {
	return b.LastUpdatedOn
}

////////////////
// Model interface
/////////////////

func (p Booking) Version() string {
	return p.ModelVersion
}

////////////////////
// OPERATIONS
///////////////////

type BookingSearchRs struct {
	Bookings []Booking `json:"bookings"`
}
