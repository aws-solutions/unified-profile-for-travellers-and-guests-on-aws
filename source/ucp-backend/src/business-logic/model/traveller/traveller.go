package model

import "time"

type Traveller struct {
	// Metadata
	ModelVersion string
	Errors       []string

	// Profile IDs
	ConnectID   string
	TravellerID string
	PSSID       string
	GDSID       string
	PMSID       string
	CRSID       string

	// Profile Data
	Honorific   string
	FirstName   string
	MiddleName  string
	LastName    string
	Gender      string
	Pronoun     string // should be an array of pronouns?
	BirthDate   time.Time
	JobTitle    string
	CompanyName string

	// Contact Info
	PhoneNumber          string
	MobilePhoneNumber    string
	HomePhoneNumber      string
	BusinessPhoneNumber  string
	PersonalEmailAddress string
	BusinessEmailAddress string
	NationalityCode      string
	NationalityName      string
	LanguageCode         string
	LanguageName         string

	// Addresses
	HomeAddress     Address
	BusinessAddress Address
	MailingAddress  Address
	BillingAddress  Address

	// Payment Info
	// TODO: should payment data be for an order, profile, separate history object or combination?

	// Object Type Records
	AirBookingRecords   []AirBooking
	AirLoyaltyRecords   []AirLoyalty
	ClickstreamRecords  []Clickstream
	EmailHistoryRecords []EmailHistory
	HotelBookingRecords []HotelBooking
	HotelLoyaltyRecords []HotelLoyalty
	HotelStayRecords    []HotelStay
	PhoneHistoryRecords []PhoneHistory

	ParsingErrors []string
}

type AirBooking struct {
	BookingID     string
	SegmentID     string
	From          string
	To            string
	FlightNumber  string
	DepartureDate string
	DepartureTime string
	ArrivalDate   string
	ArrivalTime   string
	Channel       string
	Status        string
	Price         string
	LastUpdated   time.Time
	LastUpdatedBy string
}

type AirLoyalty struct {
	LoyaltyID        string
	ProgramName      string
	Miles            string
	MilesToNextLevel string
	Level            string
	Joined           time.Time
	LastUpdated      time.Time
	LastUpdatedBy    string
}

type Clickstream struct {
	SessionID                       string
	EventTimestamp                  time.Time
	EventType                       string
	EventVersion                    string
	ArrivalTimestamp                time.Time
	UserAgent                       string
	Products                        string
	FareClass                       string
	FareType                        string
	FlightSegmentsDepartureDateTime time.Time
	FlightNumbers                   string
	FlightMarket                    string
	FlightType                      string
	OriginDate                      string
	OriginDateTime                  time.Time
	ReturnDate                      string
	ReturnDateTime                  time.Time
	ReturnFlightRoute               string
	NumPaxAdults                    int
	NumPaxInf                       int
	NumPaxChildren                  int
	PaxType                         string
	TotalPassengers                 int
	LastUpdated                     time.Time
	LastUpdatedBy                   string
}

type EmailHistory struct {
	Address       string
	Type          string
	LastUpdated   time.Time
	LastUpdatedBy string
}

type HotelBooking struct {
	BookingID             string
	HotelCode             string
	NumNights             int
	NumGuests             int
	ProductID             string
	CheckInDate           time.Time
	RoomTypeCode          string
	RoomTypeName          string
	RoomTypeDescription   string
	AttributeCodes        string
	AttributeNames        string
	AttributeDescriptions string
	LastUpdated           time.Time
	LastUpdatedBy         string
}

type HotelLoyalty struct {
	LoyaltyID         string
	ProgramName       string
	Points            string
	Units             string
	PointsToNextLevel string
	Level             string
	Joined            time.Time
	LastUpdated       time.Time
	LastUpdatedBy     string
}

type HotelStay struct {
	ID             string
	BookingID      string
	CurrencyCode   string
	CurrencyName   string
	CurrencySymbol string
	FirstName      string
	LastName       string
	Email          string
	Phone          string
	StartDate      time.Time
	HotelCode      string
	Type           string
	Description    string
	Amount         string
	Date           time.Time
	LastUpdated    time.Time
	LastUpdatedBy  string
}

type PhoneHistory struct {
	Number        string
	CountryCode   string
	Type          string
	LastUpdated   time.Time
	LastUpdatedBy string
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
	Address1   string
	Address2   string
	Address3   string
	Address4   string
	City       string
	State      string
	Province   string
	PostalCode string
	Country    string
}

type Company struct {
	ID      string    `json:"id"`   //ID in internal system (used for lookup in CRM or other)
	Name    string    `json:"name"` //Name of the company
	Address Address   `json:"address"`
	TS      time.Time `json:"timestamp"`
	Source  string    `json:"source"`
}
