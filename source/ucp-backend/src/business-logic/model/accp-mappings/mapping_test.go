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

	testMapping(BuildAirBookingMapping().GetSourceNames(), "../../../../test_assets/air_booking_fields.csv", t)
	testMapping(BuildAirLoyaltyMapping().GetSourceNames(), "../../../../test_assets/air_loyalty_fields.csv", t)
	testMapping(BuildClickstreamMapping().GetSourceNames(), "../../../../test_assets/clickstream_fields.csv", t)
	testMapping(BuildEmailHistoryMapping().GetSourceNames(), "../../../../test_assets/email_history_fields.csv", t)
	testMapping(BuildGuestProfileMapping().GetSourceNames(), "../../../../test_assets/guest_profile_fields.csv", t)
	testMapping(BuildHotelBookingMapping().GetSourceNames(), "../../../../test_assets/hotel_booking_fields.csv", t)
	testMapping(BuildHotelLoyaltyMapping().GetSourceNames(), "../../../../test_assets/hotel_loyalty_fields.csv", t)
	testMapping(BuildHotelStayMapping().GetSourceNames(), "../../../../test_assets/hotel_stay_fields.csv", t)
	testMapping(BuildPassengerProfileMapping().GetSourceNames(), "../../../../test_assets/pax_profile_fields.csv", t)
	testMapping(BuildPhoneHistoryMapping().GetSourceNames(), "../../../../test_assets/phone_history_fields.csv", t)
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
	missings := []string{}
	for _, v := range expectedFields {
		key := v[0]
		_, ok := sourceMap[key]
		if ok {
			delete(sourceMap, key)
		} else {

			missings = append(missings, key)
		}
	}
	if len(missings) > 0 {
		toaAdd := ""
		for _, key := range missings {
			toaAdd += `{
				Type:    "STRING",
				Source:  "_source.` + key + `",
				Target:  "` + key + `",
				KeyOnly: true,
			},
			`
		}
		t.Errorf("***************************\n %v: %v mapping missings\n**********************\n\n Add the following \n%v", filePath, len(missings), toaAdd)
	}
	if len(sourceMap) > 0 {
		for k := range sourceMap {
			t.Errorf("[%s] Field %v found but not expected", filePath, k)
		}
	}
}
