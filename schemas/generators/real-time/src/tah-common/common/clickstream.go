package common

import (
	core "tah/upt/schemas/src/tah-common/core"
	"time"
)

// initial v1.0.0
var EVENT_START_SESSION string = "start_session"                    // The traveler initiates a new session by accessing the travel booking site or Mobile App. A session ID is created.
var EVENT_LOGIN string = "login"                                    // The traveler identify themselves by signin-in
var EVENT_LOGOUT string = "logout"                                  // The traveler signs out
var EVENT_SEARCH_DESTINATION string = "search_destination"          // The traveler searches for a destination (could be a county, a city, an Airport or a specific hotel property). Typically, a geolocation service.
var EVENT_SEARCH_EXPERIENCE string = "search_experience"            // The traveler searches for a travel experience (could be a cruise, a tour, an adventure vacation…). This search is not specific to a destination.
var EVENT_SELECT_DESTINATION string = "select_destination"          // The traveler selects one of the destinations returned by the search service used.
var EVENT_SELECT_EXPERIENCE string = "select_experience"            // The traveler selects one of the experiences returned by the search service used.
var EVENT_SELECT_ORIGIN string = "select_origin"                    // The traveler selects an origin for the trip (departure airport, car rental pick-up location…).
var EVENT_SELECT_N_TRAVELLER string = "select_n_traveller"          // The traveler selects the number of travelers in the trip.
var EVENT_SELECT_START_DATE string = "select_start_date"            // The traveler selects a start date for the trip (check-in date for a hotel, departure date for a flight, pick-up date for a car rental or a cruise).
var EVENT_SELECT_END_DATE string = "select_end_date"                // The traveler selects an end date for the trip (return date for a flight, or a car rental, checkout date for a hotel…).
var EVENT_SELECT_START_TIME string = "select_start_time"            // Optional time to be used for hotel, alternate accommodation check-in, or car rental pick-up.
var EVENT_SELECT_END_TIME string = "select_end_time"                // Optional time to be used for hotel, alternate accommodation checkout, or car rental return.
var EVENT_SELECT_N_ROOM string = "select_n_rooms"                   // The traveler selects a number of rooms for a stay or cabins for a cruise.
var EVENT_SELECT_N_NIGHT string = "select_n_nights"                 // The traveler selects a number of nights for a hotel stay or a cruise.
var EVENT_SEARCH_FLIGHT string = "search_flight"                    // Search for a flight.
var EVENT_SEARCH_MULTI_AVAIL string = "search_multi_availability"   // Search for availability for multiple properties.
var EVENT_SEARCH_SINGLE_AVAIL string = "search_single_availability" // Search for availability in one property.
var EVENT_VIEW_PRODUUCT string = "view_product"                     // The traveler views a product.
var EVENT_SELECT_PRODUCT string = "select_product"                  // The traveler selects a product from the search results.
var EVENT_START_BOOKING string = "start_booking"                    // The traveler starts a booking.
var EVENT_UPDATE_BOOKING string = "update_booking"                  // The traveler updates a booking.
var EVENT_CONFIRM_BOOKING string = "confirm_booking"                // The traveler confirms a booking.
var EVENT_CANCEL_BOOKING string = "cancel_booking"                  // The traveler cancels a booking.
var EVENT_RETRIEVE_BOOKING string = "retrieve_booking"              // Retrieve a booking.
var EVENT_SEARCH_BOOKING string = "search_booking"                  // Search for a booking.
var EVENT_IGNORE_BOOKING string = "ignore_booking"                  // Ignore a booking.
var EVENT_REGRET string = "regret"                                  // Regret event.
var EVENT_SEARCH_ADD_ON string = "search_add_on"                    // Search for an add-on to add to the reservation.
var EVENT_VIEW_ADD_ON string = "view_add_on"                        // View an add-on.
var EVENT_SELECT_ADD_ON string = "select_add_on"                    // Select an add-on from the search results.
var EVENT_REMOVE_ADD_ON string = "remove_add_on"                    // Remove an add-on from the reservation.
var EVENT_CUSTOM string = "custom"                                  // Custom event.
var EVENT_CROSS_GEOFENCE string = "cross_geofence"                  // Cross geofence event.
// new v1.1.0
var EVENT_CONFIRM_PURCHASE string = "confirm_purchase" // Confirm purchase
var EVENT_EMPTY_CART string = "empty_cart"             // empty cart
// new v2.0.0
var EVENT_SIGN_UP string = "sign_up"                               // The traveler signs up for loyalty
var EVENT_CHECK_IN string = "check_in"                             // The traveler checks in online
var EVENT_ATTRIBUTE_ORIGIN string = "origin"                       // origin airport code (ex: ATL)
var EVENT_ATTRIBUTE_DESTINATION string = "destination"             // destination airport code (ex: BOS)
var EVENT_ATTRIBUTE_MARKETING_CARRIER string = "marketing_carrier" // marketing carrier (ex UA)
var EVENT_ATTRIBUTE_LENGTH_OF_STAY string = "length_of_stay"       // length of stay (ex: 2)
var EVENT_ATTRIBUTE_ERROR string = "error"                         // traveler facing error happened (ex: "We were unable to assign some or all of your seats. Please visit Manage Reservations to try selecting seats again.,Manage Reservations")
var EVENT_ATTRIBUTE_SELECTED_SEATS string = "selected_seats"       //selected seats (ex: [Choose seat|37E,13B|Choose seat])

// / Common Event attributes for all segments
// TRAVELLER
var EVENT_ATTRIBUTE_CUSTOMER_BIRTHDATE = "customer_birthdate"     // Contains the customer's date of birth. "1981-03-03T00:00:00"
var EVENT_ATTRIBUTE_CUSTOMER_COUNTRY = "customer_country"         // Contains the customer's country. "MX"
var EVENT_ATTRIBUTE_CUSTOMER_EMAIL = "customer_email"             // Contains the customer's email address. "email@example.com"
var EVENT_ATTRIBUTE_CUSTOMER_FIRST_NAME = "customer_first_name"   // Contains the customer's first name. "John"
var EVENT_ATTRIBUTE_CUSTOMER_GENDER = "customer_gender"           // Contains the customer's gender. "Male"
var EVENT_ATTRIBUTE_CUSTOMER_ID = "customer_id"                   // Contains the customer's id. "123"
var EVENT_ATTRIBUTE_CUSTOMER_LAST_NAME = "customer_last_name"     // Contains the customer's last name. "Doe"
var EVENT_ATTRIBUTE_CUSTOMER_NATIONALITY = "customer_nationality" // Contains the customer's nationality. "MX"
var EVENT_ATTRIBUTE_CUSTOMER_PHONE = "customer_phone"             // Contains the customer's phone number. "527523695215"
var EVENT_ATTRIBUTE_CUSTOMER_TYPE = "customer_type"               // Contains the customer's type. "Anonymous"
var EVENT_ATTRIBUTE_LANGUAGE_CODE = "language_code"               // Contains the language code. "en-US"
var EVENT_ATTRIBUTE_LOYALTY_ID = "loyalty_id"                     // Contains the loyalty program identifier
// ORDER
var EVENT_ATTRIBUTE_CURRENCY = "currency"                           // Contains the currency. "MXN"
var EVENT_ATTRIBUTE_ECOMMERCE_ACTION = "ecommerce_action"           // Contains the ecommerce action. "Booking Flow"
var EVENT_ATTRIBUTE_ORDER_PAYMENT_TYPE = "order_payment_type"       // Payment Type "Visa;VI,Visa;VI,Visa;VI"
var EVENT_ATTRIBUTE_ORDER_PROMO_CODE = "order_promo_code"           // Promo Code "VOLH50"
var EVENT_ATTRIBUTE_PAGE_NAME = "page_name"                         // Page Name "FlightSearch"
var EVENT_ATTRIBUTE_PAGE_TYPE_ENVIRONMENT = "page_type_environment" // Page Type "test"
var EVENT_ATTRIBUTE_TRANSACTION_ID = "transaction_id"               // ID of the purchase transaction
var EVENT_ATTRIBUTE_BOOKING_ID = "booking_id"                       // Airline PNR (ex X7Z1TI) or Hotel Booking Id
var EVENT_ATTRIBUTE_URL = "url"                                     // Value of the page URL. (airline.com/flight/search, brand.com)

// GEOFENCE
var EVENT_ATTRIBUTE_GEOFENCE_LATITUDE = "geofence_latitude"   // Contains the geofence latitude. "19.4"
var EVENT_ATTRIBUTE_GEOFENCE_LONGITUDE = "geofence_longitude" // Contains the geofence longitude. "99.4"
var EVENT_ATTRIBUTE_GEOFENCE_ID = "geofence_id"               // Contains the geofence id. "1"
var EVENT_ATTRIBUTE_GEOFENCE_NAME = "geofence_name"           // Contains the geofence name. "1"
var EVENT_ATTRIBUTE_POI_ID = "poi_id"                         // Place ID in geolocation API

// OTHER
var EVENT_ATTRIBUTE_CUSTOM_EVENT_NAME = "custom_event_name" // if event name is different than tealium_event
var EVENT_ATTRIBUTE_CUSTOM = "custom"                       // reserved mapping for custom attribute names

type ClickEvent struct {
	ModelVersion     string           `json:"modelVersion"`
	EventType        string           `json:"event_type"`
	EventTimestamp   time.Time        `json:"event_timestamp"`
	ArrivalTimestamp int64            `json:"arrival_timestamp"`
	EventVersion     string           `json:"event_version"`
	Application      EventApplication `json:"application"`
	Client           EventClient      `json:"client"`
	Device           EventDevice      `json:"device"`
	Session          EventSession     `json:"session"`
	Attributes       []EventAttribute `json:"attributes"`
	Endpoint         EventEndpoint    `json:"endpoint"`
	AwsAccountID     string           `json:"awsAccountId"`
	UserID           string           `json:"userId"` //if known, this field identifies the user unambigiously (can come from a CDP or similar application)
}

var EVENT_ATTRIBUTE_TYPE_STRING = "string"
var EVENT_ATTRIBUTE_TYPE_STRINGS = "strings"
var EVENT_ATTRIBUTE_TYPE_NUMBER = "number"
var EVENT_ATTRIBUTE_TYPE_NUMBERS = "numbers"

type EventAttribute struct {
	Type         string       `json:"type"`
	Name         string       `json:"name"`
	StringValue  string       `json:"stringValue"`
	StringValues []string     `json:"stringValues"`
	NumValue     core.Float   `json:"numValue"`
	NumValues    []core.Float `json:"numValues"`
}

type EventApplication struct {
	ID string `json:"id"`
}
type EventClient struct {
	ID string `json:"id"`
}
type EventDevice struct {
	UserAgent string `json:"useragent"` // "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
	IpAddress string `json:"ipAddress"` // "The ID Address if the device
}
type EventSession struct {
	ID string `json:"id"`
}
type EventEndpoint struct {
	ID string `json:"id"`
}

func (p ClickEvent) Version() string {
	return p.ModelVersion
}
