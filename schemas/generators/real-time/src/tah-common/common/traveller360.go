package common

import core "tah/upt/schemas/src/tah-common/core"

type Traveller struct {
	ModelVersion    string                  `json:"modelVersion"`
	Chain           string                  `json:"chain"`
	ID              string                  `json:"unique_id"`
	Email           string                  `json:"email"`
	Emails          []string                `json:"emails"`
	FirstNames      []string                `json:"firstName"`
	LastName        string                  `json:"last_name"`
	FirstName       string                  `json:"first_name"`
	LastNames       []string                `json:"lastName"`
	Phones          []string                `json:"phones"`
	HotelBookings   []HotelBookingSummary   `json:"hotelBookings"`
	AirBookings     []AirBookingSummary     `json:"airbookings"`
	CruiseBookings  []CruiseBookingSummary  `json:"cruiseBookings"`
	CarBookings     []CarBookingSummary     `json:"carBookings"`
	RailBookings    []RailBookingSummary    `json:"railBookings"`
	Searches        []Search                `json:"searches"`
	LoyaltyProfiles []LoyaltyProfileSummary `json:"loyaltyProfiles"`
	Addresses       []Address               `json:"addresses"`
	Companies       []string                `json:"parentCompanies"`
}

type HotelBookingSummary struct {
	ID         string `json:"id"`
	StartDate  string `json:"startDate"`
	TotalPrice string `json:"totalPrice"`
	Currency   string `json:"currency"`

	Products   string `json:"products"`
	NNight     string `json:"nNight"`
	Hotel_code string `json:"hotel_code"`
	Channel    string `json:"channel"`
	NGuests    string `json:"nGuests"`
}
type AirBookingSummary struct {
	ID          string      `json:"id"`
	TotalPrice  string      `json:"totalPrice"`
	Currency    string      `json:"currency"`
	Itinerary   Itinerary   `json:"itinerary"`
	Return      Itinerary   `json:"return"`
	Ancillaries []Ancillary `json:"ancillaries"`
}

type Ancillary struct {
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
	Status        string          `json:"status"`
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
type Inventory struct {
	TotalSeats           int                     `json:"totalSeats"`
	InventoryByFareClass []InventoryForFareClass `json:"inventoryByFareClass"`
}

type InventoryForFareClass struct {
	FareClass string `json:"fareClass"`
	Number    int    `json:"number"`
}

type CruiseBookingSummary struct {
	ID          string `json:"id"`
	StartDate   string `json:"startDate"`
	Total_price string `json:"total_price"`
	Products    string `json:"products"`
	NNight      string `json:"nNight"`
	Hotel_code  string `json:"hotel_code"`
	Channel     string `json:"channel"`
	NGuests     string `json:"nGuests"`
}

type CarBookingSummary struct {
	ID          string `json:"id"`
	StartDate   string `json:"startDate"`
	Total_price string `json:"total_price"`
	Products    string `json:"products"`
	NNight      string `json:"nNight"`
	Hotel_code  string `json:"hotel_code"`
	Channel     string `json:"channel"`
	NGuests     string `json:"nGuests"`
}

type RailBookingSummary struct {
	ID          string `json:"id"`
	StartDate   string `json:"startDate"`
	Total_price string `json:"total_price"`
	Products    string `json:"products"`
	NNight      string `json:"nNight"`
	Hotel_code  string `json:"hotel_code"`
	Channel     string `json:"channel"`
	NGuests     string `json:"nGuests"`
}

type Search struct {
	Date        string `json:"date"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Hotel       string `json:"hotel"`
	Sailing     string `json:"sailing"`
}

type LoyaltyProfileSummary struct {
	ID                string              `json:"id"`
	Program           string              `json:"program"`
	Gender            string              `json:"gender"`
	FirstName         string              `json:"firstName"`
	MiddleName        string              `json:"middleName"`
	LastName          string              `json:"lastName"`
	Birthdate         string              `json:"birthdate"`
	Points            core.Float          `json:"points"`
	Status            string              `json:"status"`
	HotelPreferences  []HotelPreferences  `json:"hotelPreferences"`
	AirPreferences    []AirPreferences    `json:"airPreferences"`
	CruisePreferences []CruisePreferences `json:"cruisePreferences"`
	CarPreferences    []CarPreferences    `json:"carPreferences"`
	RailPreferences   []RailPreferences   `json:"railPreferences"`
}

type HotelPreferences struct{}
type AirPreferences struct{}
type CruisePreferences struct{}
type CarPreferences struct{}
type RailPreferences struct{}
