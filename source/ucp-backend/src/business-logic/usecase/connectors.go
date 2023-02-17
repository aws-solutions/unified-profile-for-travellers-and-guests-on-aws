package usecase

import "os"

// Connectors
const (
	Hapi    string = "hapi"
	Tealium string = "tealium"
)

// Business Objects
const (
	HotelBooking string = "hotel-booking"
	AirBooking   string = "air-booking"
	GuestProfile string = "guest-profile"
	PaxProfile   string = "pax-profile"
	Clickstream  string = "clickstream"
	HotelStay    string = "hotel-stay"
)

// Map of Business Objects supported by each Connector
var connectorMap = map[string][]string{
	Hapi: {
		HotelBooking,
		HotelStay,
		GuestProfile,
	},
	Tealium: {
		Clickstream,
	},
}

var objectJobNameMap = map[string]string{
	HotelBooking: os.Getenv("HOTEL_BOOKING_JOB_NAME"),
	AirBooking:   os.Getenv("AIR_BOOKING_JOB_NAME"),
	GuestProfile: os.Getenv("GUEST_PROFILE_JOB_NAME"),
	PaxProfile:   os.Getenv("PAX_PROFILE_JOB_NAME"),
	Clickstream:  os.Getenv("CLICKSTREAM_JOB_NAME"),
	HotelStay:    os.Getenv("HOTEL_STAY_JOB_NAME"),
}

func GetObjectsForConnector(connectorId string) []string {
	return connectorMap[connectorId]
}

func GetJobNamesForObjects(businessObjects []string) []string {
	jobNames := []string{}
	for _, obj := range businessObjects {
		jobNames = append(jobNames, objectJobNameMap[obj])
	}
	return jobNames
}
