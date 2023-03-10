package model

import "time"

type Traveller struct {
	//Admin fields
	ModelVersion string    `json:"modelVersion"`
	TS           time.Time `json:"lastUpdated"`
	UpdatedBy    string    `json:"updatedBy"`

	//IDs
	ID          string              `json:"unique_id"`
	ExternalIDs []ExternalProfileID `json:"externalIDs"`

	//Main Identifiers
	Gender     string    `json:"gender"`
	Pronouns   []string  `json:"pronouns"`
	Title      string    `json:"title"`
	LastName   string    `json:"lastName"`
	MiddleName string    `json:"middleName"`
	FirstName  string    `json:"firstName"`
	BirthDate  time.Time `json:"birthDate"`

	//PII
	Emails     []TraceableString `json:"emails"`
	FirstNames []TraceableString `json:"firstNames"`
	LastNames  []TraceableString `json:"lastNames"`
	Phones     []TraceableString `json:"phones"`
	Addresses  []Address         `json:"addresses"`
	Companies  []Company         `json:"company"`
	JobTitles  []TraceableString `json:"jobTitles"`

	//Loyalty
	LoyaltyProfiles []LoyaltyProfileSummary `json:"loyaltyProfiles"`

	//Transactions
	RestaurantOrders []RestaurantOrder      `json:"restaurantOrders"`
	HotelBookings    []HotelBookingSummary  `json:"hotelBookings"`
	AirBookings      []AirBookingSummary    `json:"airbookings"`
	CruiseBookings   []CruiseBookingSummary `json:"cruiseBookings"`
	CarBookings      []CarBookingSummary    `json:"carBookings"`
	RailBookings     []RailBookingSummary   `json:"railBookings"`
	//Interests
	Searches []Search `json:"searches"`
	
}

type TraceableString struct {
	Value  string    `json:"value"`
	TS     time.Time `json:"timestamp"`
	Source string    `json:"source"`
}

type ExternalProfileID struct {
	System string `json:"system"`
	ID     string `json:"id"`
}

type HotelBookingSummary struct {
	ID         string    `json:"id"`
	StartDate  string    `json:"startDate"`
	TotalPrice float64   `json:"totalPrice"`
	Currency   string    `json:"currency"`
	Products   string    `json:"products"`
	NNight     int64     `json:"nNight"`
	HotelCode  string    `json:"hotelCode"`
	Channel    string    `json:"channel"`
	NGuests    int64     `json:"nGuests"`
	TS         time.Time `json:"timestamp"`
	Source     string    `json:"source"`
}
type AirBookingSummary struct {
	ID          string      `json:"id"`
	TotalPrice  string      `json:"totalPrice"`
	Currency    string      `json:"currency"`
	Itinerary   Itinerary   `json:"itinerary"`
	Return      Itinerary   `json:"return"`
	Ancillaries []Ancillary `json:"ancillaries"`
	TS          time.Time   `json:"timestamp"`
	Source      string      `json:"source"`
}

type RestaurantOrder struct {
	ID          string      `json:"id"`
	TotalPrice  string      `json:"totalPrice"`
	Currency    string      `json:"currency"`
	Itinerary   Itinerary   `json:"itinerary"`
	Return      Itinerary   `json:"return"`
	Ancillaries []Ancillary `json:"ancillaries"`
	TS          time.Time   `json:"timestamp"`
	Source      string      `json:"source"`
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
	ID          string    `json:"id"`
	StartDate   string    `json:"startDate"`
	Total_price string    `json:"total_price"`
	Products    string    `json:"products"`
	NNight      string    `json:"nNight"`
	Hotel_code  string    `json:"hotel_code"`
	Channel     string    `json:"channel"`
	NGuests     string    `json:"nGuests"`
	TS          time.Time `json:"timestamp"`
	Source      string    `json:"source"`
}

type CarBookingSummary struct {
	ID          string    `json:"id"`
	StartDate   string    `json:"startDate"`
	Total_price string    `json:"total_price"`
	Products    string    `json:"products"`
	NNight      string    `json:"nNight"`
	Hotel_code  string    `json:"hotel_code"`
	Channel     string    `json:"channel"`
	NGuests     string    `json:"nGuests"`
	TS          time.Time `json:"timestamp"`
	Source      string    `json:"source"`
}

type RailBookingSummary struct {
	ID          string    `json:"id"`
	StartDate   string    `json:"startDate"`
	Total_price string    `json:"total_price"`
	Products    string    `json:"products"`
	NNight      string    `json:"nNight"`
	Hotel_code  string    `json:"hotel_code"`
	Channel     string    `json:"channel"`
	NGuests     string    `json:"nGuests"`
	TS          time.Time `json:"timestamp"`
	Source      string    `json:"source"`
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
	Birthdate         time.Time           `json:"birthdate"`
	Points            float64             `json:"points"`
	Status            string              `json:"status"`
	Joined            time.Time           `json:"joined"`
	HotelPreferences  []HotelPreferences  `json:"hotelPreferences"`
	AirPreferences    []AirPreferences    `json:"airPreferences"`
	CruisePreferences []CruisePreferences `json:"cruisePreferences"`
	CarPreferences    []CarPreferences    `json:"carPreferences"`
	RailPreferences   []RailPreferences   `json:"railPreferences"`
}

type HotelPreferences struct {
	TS     time.Time `json:"timestamp"`
	Source string    `json:"source"`
}
type AirPreferences struct {
	TS     time.Time `json:"timestamp"`
	Source string    `json:"source"`
}
type CruisePreferences struct {
	TS     time.Time `json:"timestamp"`
	Source string    `json:"source"`
}
type CarPreferences struct {
	TS     time.Time `json:"timestamp"`
	Source string    `json:"source"`
}
type RailPreferences struct {
	TS     time.Time `json:"timestamp"`
	Source string    `json:"source"`
}

type Address struct {
	StateCode    string    `json:"stateCode"`
	PostalCode   string    `json:"postalCode"`
	City         string    `json:"city"`
	Country      string    `json:"country"`
	State        string    `json:"state"`
	AddressLine1 string    `json:"addressLine1"`
	TS           time.Time `json:"timestamp"`
	Source       string    `json:"source"`
}

type Company struct {
	ID      string    `json:"id"`   //ID in internal system (used for lookup in CRM or other)
	Name    string    `json:"name"` //Name of the company
	Address Address   `json:"address"`
	TS      time.Time `json:"timestamp"`
	Source  string    `json:"source"`
}
