package model

import (
	"time"
)

type Traveller struct {
	// Metadata
	ModelVersion string   `json:"modelVersion"`
	Errors       []string `json:"errors"`

	// Profile IDs
	ConnectID   string `json:"connectId"`
	TravellerID string `json:"travellerId"`
	PSSID       string `json:"pssId"`
	GDSID       string `json:"gdsId"`
	PMSID       string `json:"pmsId"`
	CRSID       string `json:"crsId"`

	// Profile Data
	Honorific   string    `json:"honorific"`
	FirstName   string    `json:"firstName"`
	MiddleName  string    `json:"middleName"`
	LastName    string    `json:"lastName"`
	Gender      string    `json:"gender"`
	Pronoun     string    `json:"pronoun"`
	BirthDate   time.Time `json:"birthDate"`
	JobTitle    string    `json:"jobTitle"`
	CompanyName string    `json:"companyName"`

	// Contact Info
	PhoneNumber          string `json:"phoneNumber"`
	MobilePhoneNumber    string `json:"mobilePhoneNumber"`
	HomePhoneNumber      string `json:"homePhoneNumber"`
	BusinessPhoneNumber  string `json:"businessPhoneNumber"`
	PersonalEmailAddress string `json:"personalEmailAddress"`
	BusinessEmailAddress string `json:"businessEmailAddress"`
	NationalityCode      string `json:"nationalityCode"`
	NationalityName      string `json:"nationalityName"`
	LanguageCode         string `json:"languageCode"`
	LanguageName         string `json:"languageName"`

	// Addresses
	HomeAddress     Address `json:"homeAddress"`
	BusinessAddress Address `json:"businessAddress"`
	MailingAddress  Address `json:"mailingAddress"`
	BillingAddress  Address `json:"billingAddress"`

	// Payment Info
	// TODO: should payment data be for an order, profile, separate history object or combination?

	// Object Type Records
	AirBookingRecords   []AirBooking   `json:"airBookingRecords"`
	AirLoyaltyRecords   []AirLoyalty   `json:"airLoyaltyRecords"`
	ClickstreamRecords  []Clickstream  `json:"clickstreamRecords"`
	EmailHistoryRecords []EmailHistory `json:"emailHistoryRecords"`
	HotelBookingRecords []HotelBooking `json:"hotelBookingRecords"`
	HotelLoyaltyRecords []HotelLoyalty `json:"hotelLoyaltyRecords"`
	HotelStayRecords    []HotelStay    `json:"hotelStayRecords"`
	PhoneHistoryRecords []PhoneHistory `json:"phoneHistoryRecords"`

	ParsingErrors []string `json:"parsingErrors"`
}

type AirBooking struct {
	BookingID     string    `json:"bookingId"`
	SegmentID     string    `json:"segmentId"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	FlightNumber  string    `json:"flightNumber"`
	DepartureDate string    `json:"departureDate"`
	DepartureTime string    `json:"departureTime"`
	ArrivalDate   string    `json:"arrivalDate"`
	ArrivalTime   string    `json:"arrivalTime"`
	Channel       string    `json:"channel"`
	Status        string    `json:"status"`
	Price         string    `json:"price"`
	LastUpdated   time.Time `json:"lastUpdated"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
}
type AirBookingByLastUpdated []AirBooking

func (a AirBookingByLastUpdated) Len() int      { return len(a) }
func (a AirBookingByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AirBookingByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type AirLoyalty struct {
	LoyaltyID        string    `json:"loyaltyId"`
	ProgramName      string    `json:"programName"`
	Miles            string    `json:"miles"`
	MilesToNextLevel string    `json:"milesToNextLevel"`
	Level            string    `json:"level"`
	Joined           time.Time `json:"joined"`
	LastUpdated      time.Time `json:"lastUpdated"`
	LastUpdatedBy    string    `json:"lastUpdatedBy"`
}

type AirLoyaltyByLastUpdated []AirLoyalty

func (a AirLoyaltyByLastUpdated) Len() int      { return len(a) }
func (a AirLoyaltyByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AirLoyaltyByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type Clickstream struct {
	SessionID                       string    `json:"sessionId"`
	EventTimestamp                  time.Time `json:"eventTimestamp"`
	EventType                       string    `json:"eventType"`
	EventVersion                    string    `json:"eventVersion"`
	ArrivalTimestamp                time.Time `json:"arrivalTimestamp"`
	UserAgent                       string    `json:"userAgent"`
	Products                        string    `json:"products"`
	FareClass                       string    `json:"fareClass"`
	FareType                        string    `json:"fareType"`
	FlightSegmentsDepartureDateTime time.Time `json:"flightSegmentsDepartureDateTime"`
	FlightNumbers                   string    `json:"flightNumbers"`
	FlightMarket                    string    `json:"flightMarket"`
	FlightType                      string    `json:"flightType"`
	OriginDate                      string    `json:"originDate"`
	OriginDateTime                  time.Time `json:"originDateTime"`
	ReturnDate                      string    `json:"returnDate"`
	ReturnDateTime                  time.Time `json:"returnDateTime"`
	ReturnFlightRoute               string    `json:"returnFlightRoute"`
	NumPaxAdults                    int       `json:"numPaxAdults"`
	NumPaxInf                       int       `json:"numPaxInf"`
	NumPaxChildren                  int       `json:"numPaxChildren"`
	PaxType                         string    `json:"paxType"`
	TotalPassengers                 int       `json:"totalPassengers"`
	LastUpdated                     time.Time `json:"lastUpdated"`
	LastUpdatedBy                   string    `json:"lastUpdatedBy"`
}

type ClickstreamByLastUpdated []Clickstream

func (a ClickstreamByLastUpdated) Len() int      { return len(a) }
func (a ClickstreamByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ClickstreamByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type EmailHistory struct {
	Address       string    `json:"address"`
	Type          string    `json:"type"`
	LastUpdated   time.Time `json:"lastUpdated"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
}
type EmailHistoryByLastUpdated []EmailHistory

func (a EmailHistoryByLastUpdated) Len() int      { return len(a) }
func (a EmailHistoryByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a EmailHistoryByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type HotelBooking struct {
	BookingID             string    `json:"bookingId"`
	HotelCode             string    `json:"hotelCode"`
	NumNights             int       `json:"numNights"`
	NumGuests             int       `json:"numGuests"`
	ProductID             string    `json:"productId"`
	CheckInDate           time.Time `json:"checkInDate"`
	RoomTypeCode          string    `json:"roomTypeCode"`
	RoomTypeName          string    `json:"roomTypeName"`
	RoomTypeDescription   string    `json:"roomTypeDescription"`
	AttributeCodes        string    `json:"attributeCodes"`
	AttributeNames        string    `json:"attributeNames"`
	AttributeDescriptions string    `json:"attributeDescriptions"`
	LastUpdated           time.Time `json:"lastUpdated"`
	LastUpdatedBy         string    `json:"lastUpdatedBy"`
}
type HotelBookingByLastUpdated []HotelBooking

func (a HotelBookingByLastUpdated) Len() int      { return len(a) }
func (a HotelBookingByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a HotelBookingByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type HotelLoyalty struct {
	LoyaltyID         string    `json:"loyaltyId"`
	ProgramName       string    `json:"programName"`
	Points            string    `json:"points"`
	Units             string    `json:"units"`
	PointsToNextLevel string    `json:"pointsToNextLevel"`
	Level             string    `json:"level"`
	Joined            time.Time `json:"joined"`
	LastUpdated       time.Time `json:"lastUpdated"`
	LastUpdatedBy     string    `json:"lastUpdatedBy"`
}
type HotelLoyaltyByLastUpdated []HotelLoyalty

func (a HotelLoyaltyByLastUpdated) Len() int      { return len(a) }
func (a HotelLoyaltyByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a HotelLoyaltyByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type HotelStay struct {
	ID             string    `json:"id"`
	BookingID      string    `json:"bookingId"`
	CurrencyCode   string    `json:"currencyCode"`
	CurrencyName   string    `json:"currencyName"`
	CurrencySymbol string    `json:"currencySymbol"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	StartDate      time.Time `json:"startDate"`
	HotelCode      string    `json:"hotelCode"`
	Type           string    `json:"type"`
	Description    string    `json:"description"`
	Amount         string    `json:"amount"`
	Date           time.Time `json:"date"`
	LastUpdated    time.Time `json:"lastUpdated"`
	LastUpdatedBy  string    `json:"lastUpdatedBy"`
}
type HotelStayByLastUpdated []HotelStay

func (a HotelStayByLastUpdated) Len() int      { return len(a) }
func (a HotelStayByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a HotelStayByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type PhoneHistory struct {
	Number        string    `json:"number"`
	CountryCode   string    `json:"countryCode"`
	Type          string    `json:"type"`
	LastUpdated   time.Time `json:"lastUpdated"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
}
type PhoneHistoryByLastUpdated []PhoneHistory

func (a PhoneHistoryByLastUpdated) Len() int      { return len(a) }
func (a PhoneHistoryByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a PhoneHistoryByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
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
