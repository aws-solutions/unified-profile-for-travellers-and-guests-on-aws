package accpmappings

import (
	"log"
	"testing"
)

func TestMappings(t *testing.T) {
	AirBookingMapping := BuildAirBookingMapping()
	log.Printf("AirBookingMapping: %v", AirBookingMapping)
	PhoneHistoryMapping := BuildPhoneHistoryMapping()
	log.Printf("PhoneHistoryMapping: %v", PhoneHistoryMapping)
	EmailHistoryMapping := BuildEmailHistoryMapping()
	log.Printf("EmailHistoryMapping: %v", EmailHistoryMapping)
	AirLoyaltyMapping := BuildAirLoyaltyMapping()
	log.Printf("AirLoyaltyMapping: %v", AirLoyaltyMapping)
	ClickstreamMapping := BuildClickstreamMapping()
	log.Printf("ClickstreamMapping: %v", ClickstreamMapping)
	GuestProfileMapping := BuildGuestProfileMapping()
	log.Printf("GuestProfileMapping: %v", GuestProfileMapping)
	HotelLoyaltyMapping := BuildHotelLoyaltyMapping()
	log.Printf("HotelLoyaltyMapping: %v", HotelLoyaltyMapping)
	HotelStayMapping := BuildHotelStayMapping()
	log.Printf("HotelStayMapping: %v", HotelStayMapping)
	HotelBookingMapping := BuildHotelBookingMapping()
	log.Printf("HotelBookingMapping: %v", HotelBookingMapping)
	PassengerProfileMapping := BuildPassengerProfileMapping()
	log.Printf("PassengerProfileMapping: %v", PassengerProfileMapping)
}
