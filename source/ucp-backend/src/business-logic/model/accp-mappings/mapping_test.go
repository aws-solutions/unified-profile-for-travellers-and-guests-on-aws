package accpmappings

import (
	"encoding/csv"
	"log"
	"os"
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

	testMapping(BuildHotelBookingMapping().GetSourceNames(), "../../../../test_assets/hotel_booking_fields.csv", t)
}

// Compare mapped fields to expected fields from a CSV file.
// CSV test data is created from the ETL expected output, which should have all
// the standard fields, and then removing fields that are intentionally excluded.
//
// Any extra or missing fields will be printed out in the error log.
func testMapping(sourceNames []string, filePath string, t *testing.T) {
	// Create map of source fields
	sourceMap := make(map[string]struct{})
	for _, v := range sourceNames {
		sourceMap[v] = struct{}{}
	}

	// Load expected fields from CSV
	f, err := os.Open(filePath)
	if err != nil {
		t.Errorf("Error opening test file %v", err)
	}
	reader := csv.NewReader(f)
	expectedFields, err := reader.ReadAll()
	if err != nil {
		t.Errorf("Error reading test file %v", err)
	}
	f.Close()

	// Compare actual vs expected fields
	for _, v := range expectedFields {
		key := v[0]
		_, ok := sourceMap[key]
		if ok {
			delete(sourceMap, key)
		} else {
			t.Errorf("Field %v not found", key)
		}
	}
	if len(sourceMap) > 0 {
		for k := range sourceMap {
			t.Errorf("Field %v found but not expected", k)
		}
	}
}
