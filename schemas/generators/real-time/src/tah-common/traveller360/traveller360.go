package traveller360

import (
	"time"
)

type AccpObject interface {
	ID() string
	LastUpdatedTime() time.Time
}

type Traveller struct {
	// Metadata
	ModelVersion  string    `json:"modelVersion"`
	Errors        []string  `json:"errors"`
	LastUpdated   time.Time `json:"lastUpdated"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
	Domain        string    `json:"domain"`
	MergedIn      string    `json:"mergedIn"`

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
	// should payment data be for an order, profile, separate history object or combination?

	// Object Type Records
	AirBookingRecords                 []AirBooking                 `json:"airBookingRecords"`
	AirLoyaltyRecords                 []AirLoyalty                 `json:"airLoyaltyRecords"`
	LoyaltyTxRecords                  []LoyaltyTx                  `json:"lotaltyTxRecords"`
	ClickstreamRecords                []Clickstream                `json:"clickstreamRecords"`
	EmailHistoryRecords               []EmailHistory               `json:"emailHistoryRecords"`
	HotelBookingRecords               []HotelBooking               `json:"hotelBookingRecords"`
	HotelLoyaltyRecords               []HotelLoyalty               `json:"hotelLoyaltyRecords"`
	HotelStayRecords                  []HotelStay                  `json:"hotelStayRecords"`
	PhoneHistoryRecords               []PhoneHistory               `json:"phoneHistoryRecords"`
	CustomerServiceInteractionRecords []CustomerServiceInteraction `json:"customerServiceInteractionRecords"`
	AncillaryServiceRecords           []AncillaryService           `json:"ancillaryServiceRecords"`

	ParsingErrors []string `json:"parsingErrors"`
}

type AirBooking struct {
	AccpObjectID        string    `json:"accpObjectID"`
	TravellerID         string    `json:"travellerId"`
	BookingID           string    `json:"bookingId"`
	SegmentID           string    `json:"segmentId"`
	From                string    `json:"from"`
	To                  string    `json:"to"`
	FlightNumber        string    `json:"flightNumber"`
	DepartureDate       string    `json:"departureDate"`
	DepartureTime       string    `json:"departureTime"`
	ArrivalDate         string    `json:"arrivalDate"`
	ArrivalTime         string    `json:"arrivalTime"`
	Channel             string    `json:"channel"`
	Status              string    `json:"status"`
	LastUpdated         time.Time `json:"lastUpdated"`
	LastUpdatedBy       string    `json:"lastUpdatedBy"`
	TotalPrice          float64   `json:"totalPrice"`
	TravellerPrice      float64   `json:"travellerPrice"`
	BookerID            string    `json:"bookerId"`
	CreationChannelID   string    `json:"creationChannelId"`
	LastUpdateChannelID string    `json:"lastUpdateChannelId"`
}

func (o AirBooking) ID() string {
	return o.AccpObjectID
}
func (o AirBooking) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type AirBookingByLastUpdated []AirBooking

func (a AirBookingByLastUpdated) Len() int      { return len(a) }
func (a AirBookingByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AirBookingByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type AirLoyalty struct {
	AccpObjectID     string    `json:"accpObjectID"`
	TravellerID      string    `json:"travellerId"`
	LoyaltyID        string    `json:"loyaltyId"`
	ProgramName      string    `json:"programName"`
	Miles            string    `json:"miles"`
	MilesToNextLevel string    `json:"milesToNextLevel"`
	Level            string    `json:"level"`
	Joined           time.Time `json:"joined"`
	LastUpdated      time.Time `json:"lastUpdated"`
	LastUpdatedBy    string    `json:"lastUpdatedBy"`
}

func (o AirLoyalty) ID() string {
	return o.AccpObjectID
}
func (o AirLoyalty) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type AirLoyaltyByLastUpdated []AirLoyalty

func (a AirLoyaltyByLastUpdated) Len() int      { return len(a) }
func (a AirLoyaltyByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AirLoyaltyByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type Clickstream struct {
	//Accp fields
	AccpObjectID string    `json:"accpObjectId"`
	TravellerID  string    `json:"travellerId"`
	LastUpdated  time.Time `json:"lastUpdated"`

	//General Clickstream
	SessionID        string    `json:"sessionId"`
	EventTimestamp   time.Time `json:"eventTimestamp"`
	EventType        string    `json:"eventType"`
	EventVersion     string    `json:"eventVersion"`
	ArrivalTimestamp time.Time `json:"arrivalTimestamp"`
	UserAgent        string    `json:"userAgent"`
	CustomEventName  string    `json:"customEventName"`

	// Customer
	CustomerBirthdate   time.Time `json:"customerBirthdate"`
	CustomerCountry     string    `json:"customerCountry"`
	CustomerEmail       string    `json:"customerEmail"`
	CustomerFirstName   string    `json:"customerFirstName"`
	CustomerGender      string    `json:"customerGender"`
	CustomerID          string    `json:"customerId"`
	CustomerLastName    string    `json:"customerLastName"`
	CustomerNationality string    `json:"customerNationality"`
	CustomerPhone       string    `json:"customerPhone"`
	CustomerType        string    `json:"customerType"`
	CustomerLoyaltyID   string    `json:"customerLoyaltyId"`
	LanguageCode        string    `json:"languageCode"`

	// Cross industry
	Currency            string `json:"currency"`
	Products            string `json:"products"`
	Quantities          string `json:"quantities"`
	ProductsPrices      string `json:"productsPrices"`
	EcommerceAction     string `json:"ecommerceAction"`
	OrderPaymentType    string `json:"orderPaymentType"`
	OrderPromoCode      string `json:"orderPromoCode"`
	PageName            string `json:"pageName"`
	PageTypeEnvironment string `json:"pageTypeEnvironment"`
	TransactionID       string `json:"transactionId"`
	BookingID           string `json:"bookingId"`
	GeofenceLatitude    string `json:"geofenceLatitude"`
	GeofenceLongitude   string `json:"geofenceLongitude"`
	GeofenceID          string `json:"geofenceId"`
	GeofenceName        string `json:"geofenceName"`
	PoiID               string `json:"poiId"`
	URL                 string `json:"url"`
	Custom              string `json:"custom"`

	//Air
	FareClass                       string    `json:"fareClass"`
	FareType                        string    `json:"fareType"`
	FlightSegmentsDepartureDateTime string    `json:"flightSegmentsDepartureDateTime"`
	FlightSegmentsArrivalDateTime   string    `json:"flightSegmentsArrivalDateTime"`
	FlightSegments                  string    `json:"flightSegments"`
	FlightSegmentSku                string    `json:"flightSegmentsSku"`
	FlightRoute                     string    `json:"flightRoute"`
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

	// Hotels
	RoomType          string    `json:"roomType"`
	RatePlan          string    `json:"ratePlan"`
	CheckinDate       time.Time `json:"checkinDate"`
	CheckoutDate      time.Time `json:"checkoutDate"`
	NumNights         int       `json:"numNights"`
	NumGuests         int       `json:"numGuests"`
	NumGuestsAdult    int       `json:"numGuestsAdult"`
	NumGuestsChildren int       `json:"numGuestsChildren"`
	HotelCode         string    `json:"hotelCode"`
	HotelCodeList     string    `json:"hotelCodeList"`
	HotelName         string    `json:"hotelName"`
	Destination       string    `json:"destination"`
}

func (o Clickstream) ID() string {
	return o.AccpObjectID
}
func (o Clickstream) LastUpdatedTime() time.Time {
	return o.EventTimestamp
}

type ClickstreamByLastUpdated []Clickstream

func (a ClickstreamByLastUpdated) Len() int      { return len(a) }
func (a ClickstreamByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ClickstreamByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type EmailHistory struct {
	AccpObjectID  string    `json:"accpObjectID"`
	TravellerID   string    `json:"travellerId"`
	Address       string    `json:"address"`
	Type          string    `json:"type"`
	LastUpdated   time.Time `json:"lastUpdated"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
}

func (o EmailHistory) ID() string {
	return o.AccpObjectID
}
func (o EmailHistory) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type EmailHistoryByLastUpdated []EmailHistory

func (a EmailHistoryByLastUpdated) Len() int      { return len(a) }
func (a EmailHistoryByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a EmailHistoryByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type HotelBooking struct {
	AccpObjectID          string    `json:"accpObjectID"`
	TravellerID           string    `json:"travellerId"`
	BookingID             string    `json:"bookingId"`
	HotelCode             string    `json:"hotelCode"`
	NumNights             int       `json:"numNights"`
	NumGuests             int       `json:"numGuests"`
	ProductID             string    `json:"productId"`
	CheckInDate           time.Time `json:"checkInDate"`
	RoomTypeCode          string    `json:"roomTypeCode"`
	RoomTypeName          string    `json:"roomTypeName"`
	RoomTypeDescription   string    `json:"roomTypeDescription"`
	RatePlanCode          string    `json:"ratePlanCode"`
	RatePlanName          string    `json:"ratePlanName"`
	RatePlanDescription   string    `json:"ratePlanDescription"`
	AttributeCodes        string    `json:"attributeCodes"`
	AttributeNames        string    `json:"attributeNames"`
	AttributeDescriptions string    `json:"attributeDescriptions"`
	AddOnCodes            string    `json:"addOnCodes"`
	AddOnNames            string    `json:"addOnNames"`
	AddOnDescriptions     string    `json:"addOnDescriptions"`
	LastUpdated           time.Time `json:"lastUpdated"`
	LastUpdatedBy         string    `json:"lastUpdatedBy"`
	PricePerTraveller     string    `json:"pricePerTraveller"`
	BookerID              string    `json:"bookerId"`
	Status                string    `json:"status"`
	TotalSegmentBeforeTax float64   `json:"totalSegmentBeforeTax"`
	TotalSegmentAfterTax  float64   `json:"totalSegmentAfterTax"`
	CreationChannelID     string    `json:"creationChannelId"`
	LastUpdateChannelID   string    `json:"lastUpdateChannelId"`
}

func (o HotelBooking) ID() string {
	return o.AccpObjectID
}
func (o HotelBooking) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type HotelBookingByLastUpdated []HotelBooking

func (a HotelBookingByLastUpdated) Len() int      { return len(a) }
func (a HotelBookingByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a HotelBookingByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type HotelLoyalty struct {
	AccpObjectID      string    `json:"accpObjectID"`
	TravellerID       string    `json:"travellerId"`
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

func (o HotelLoyalty) ID() string {
	return o.AccpObjectID
}
func (o HotelLoyalty) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type HotelLoyaltyByLastUpdated []HotelLoyalty

func (a HotelLoyaltyByLastUpdated) Len() int      { return len(a) }
func (a HotelLoyaltyByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a HotelLoyaltyByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type LoyaltyTx struct {
	AccpObjectID             string    `json:"accpObjectID"`
	TravellerID              string    `json:"travellerId"`
	LastUpdated              time.Time `json:"lastUpdated"`
	LastUpdatedBy            string    `json:"lastUpdatedBy"`
	PointsOffset             float64   `json:"pointsOffset"`
	PointUnit                string    `json:"pointUnit"`
	OriginPointsOffset       float64   `json:"originPointsOffset"`
	QualifyingPointsOffset   float64   `json:"qualifyingPointsOffset"`
	Source                   string    `json:"source"`
	Category                 string    `json:"category"`
	BookingDate              time.Time `json:"bookingDate"`
	OrderNumber              string    `json:"orderNumber"`
	ProductId                string    `json:"productId"`
	ExpireInDays             int       `json:"expireInDays"`
	Amount                   float64   `json:"amount"`
	AmountType               string    `json:"amountType"`
	VoucherQuantity          int       `json:"voucherQuantity"`
	CorporateReferenceNumber string    `json:"corporateReferenceNumber"`
	Promotions               string    `json:"promotions"`
	ActivityDay              time.Time `json:"activityDay"`
	Location                 string    `json:"location"`
	ToLoyaltyId              string    `json:"toLoyaltyId"`
	FromLoyaltyId            string    `json:"fromLoyaltyId"`
	OrganizationCode         string    `json:"organizationCode"`
	EventName                string    `json:"eventName"`
	DocumentNumber           string    `json:"documentNumber"`
	CorporateId              string    `json:"corporateId"`
	ProgramName              string    `json:"programName"`
}

func (o LoyaltyTx) ID() string {
	return o.AccpObjectID
}
func (o LoyaltyTx) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type LoyaltyTxByLastUpdated []LoyaltyTx

func (a LoyaltyTxByLastUpdated) Len() int      { return len(a) }
func (a LoyaltyTxByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a LoyaltyTxByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type HotelStay struct {
	AccpObjectID   string    `json:"accpObjectID"`
	TravellerID    string    `json:"travellerId"`
	StayID         string    `json:"stayId"`
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

func (o HotelStay) ID() string {
	return o.AccpObjectID
}
func (o HotelStay) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type HotelStayByLastUpdated []HotelStay

func (a HotelStayByLastUpdated) Len() int      { return len(a) }
func (a HotelStayByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a HotelStayByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type CustomerServiceInteraction struct {
	AccpObjectID    string    `json:"accpObjectID"`
	TravellerID     string    `json:"travellerId"`
	LastUpdated     time.Time `json:"lastUpdated"`
	LastUpdatedBy   string    `json:"lastUpdatedBy"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	Duration        int       `json:"duration"`
	SessionId       string    `json:"sessionId"`
	Channel         string    `json:"channel"`
	InteractionType string    `json:"interactionType"`
	Status          string    `json:"status"`
	LanguageCode    string    `json:"languageCode"`
	LanguageName    string    `json:"languageName"`
	Conversation    string    `json:"conversation"`
	SentimentScore  float64   `json:"sentimentScore"`
}

func (o CustomerServiceInteraction) ID() string {
	return o.AccpObjectID
}
func (o CustomerServiceInteraction) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type CustomerServiceInteractionByLastUpdated []CustomerServiceInteraction

func (a CustomerServiceInteractionByLastUpdated) Len() int      { return len(a) }
func (a CustomerServiceInteractionByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a CustomerServiceInteractionByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type PhoneHistory struct {
	AccpObjectID  string    `json:"accpObjectID"`
	TravellerID   string    `json:"travellerId"`
	Number        string    `json:"number"`
	CountryCode   string    `json:"countryCode"`
	Type          string    `json:"type"`
	LastUpdated   time.Time `json:"lastUpdated"`
	LastUpdatedBy string    `json:"lastUpdatedBy"`
}

func (o PhoneHistory) ID() string {
	return o.AccpObjectID
}
func (o PhoneHistory) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type PhoneHistoryByLastUpdated []PhoneHistory

func (a PhoneHistoryByLastUpdated) Len() int      { return len(a) }
func (a PhoneHistoryByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a PhoneHistoryByLastUpdated) Less(i, j int) bool {
	return a[i].LastUpdated.After(a[j].LastUpdated)
}

type AncillaryService struct {
	AccpObjectID             string    `json:"accpObjectID"`
	TravellerID              string    `json:"travellerId"`
	LastUpdated              time.Time `json:"lastUpdated"`
	LastUpdatedBy            string    `json:"lastUpdatedBy"`
	AncillaryType            string    `json:"ancillaryType"`
	BookingID                string    `json:"bookingId"`
	FlightNumber             string    `json:"flightNumber"`
	DepartureDate            string    `json:"departureDate"`
	BaggageType              string    `json:"baggageType"`
	PaxIndex                 int64     `json:"paxIndex"`
	Quantity                 int64     `json:"quantity"`
	Weight                   float64   `json:"weight"`
	DimentionsLength         float64   `json:"dimentionsLength"`
	DimentionsWidth          float64   `json:"dimentionsWidth"`
	DimentionsHeight         float64   `json:"dimentionsHeight"`
	PriorityBagDrop          bool      `json:"priorityBagDrop"`
	PriorityBagReturn        bool      `json:"priorityBagReturn"`
	LotBagInsurance          bool      `json:"lotBagInsurance"`
	ValuableBaggageInsurance bool      `json:"valuableBaggageInsurance"`
	HandsFreeBaggage         bool      `json:"handsFreeBaggage"`
	SeatNumber               string    `json:"seatNumber"`
	SeatZone                 string    `json:"seatZone"`
	NeighborFreeSeat         bool      `json:"neighborFreeSeat"`
	UpgradeAuction           bool      `json:"upgradeAuction"`
	ChangeType               string    `json:"changeType"`
	OtherAncilliaryType      string    `json:"otherAncilliaryType"`
	PriorityServiceType      string    `json:"priorityServiceType"`
	LoungeAccess             bool      `json:"loungeAccess"`
	Price                    float64   `json:"price"`
	Currency                 string    `json:"currency"`
}

func (o AncillaryService) ID() string {
	return o.AccpObjectID
}
func (o AncillaryService) LastUpdatedTime() time.Time {
	return o.LastUpdated
}

type AncillaryServiceByLastUpdated []AncillaryService

func (a AncillaryServiceByLastUpdated) Len() int      { return len(a) }
func (a AncillaryServiceByLastUpdated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AncillaryServiceByLastUpdated) Less(i, j int) bool {
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
	ID          string             `json:"id"`
	TotalPrice  string             `json:"totalPrice"`
	Currency    string             `json:"currency"`
	Itinerary   Itinerary          `json:"itinerary"`
	Return      Itinerary          `json:"return"`
	Ancillaries []AncillaryService `json:"ancillaries"`
	TS          time.Time          `json:"timestamp"`
	Source      string             `json:"source"`
}

type RestaurantOrder struct {
	ID          string             `json:"id"`
	TotalPrice  string             `json:"totalPrice"`
	Currency    string             `json:"currency"`
	Itinerary   Itinerary          `json:"itinerary"`
	Return      Itinerary          `json:"return"`
	Ancillaries []AncillaryService `json:"ancillaries"`
	TS          time.Time          `json:"timestamp"`
	Source      string             `json:"source"`
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

type Filter struct {
	ObjectName    string
	Fields        []string
	ExcludeFields bool
	Condition     string
}

type Match struct {
	ConfidenceScore float64 `json:"confidence"`
	ID              string  `json:"id"`
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	BirthDate       string  `json:"birthDate"`
	PhoneNumber     string  `json:"phone"`
	EmailAddress    string  `json:"email"`
}
