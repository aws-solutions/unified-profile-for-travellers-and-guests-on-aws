// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	"encoding/csv"
	"os"
	"testing"
)

func TestMappings(t *testing.T) {
	AirBookingMapping := BuildAirBookingMapping()
	t.Logf("AirBookingMapping: %v", AirBookingMapping)
	PhoneHistoryMapping := BuildPhoneHistoryMapping()
	t.Logf("PhoneHistoryMapping: %v", PhoneHistoryMapping)
	EmailHistoryMapping := BuildEmailHistoryMapping()
	t.Logf("EmailHistoryMapping: %v", EmailHistoryMapping)
	AirLoyaltyMapping := BuildAirLoyaltyMapping()
	t.Logf("AirLoyaltyMapping: %v", AirLoyaltyMapping)
	ClickstreamMapping := BuildClickstreamMapping()
	t.Logf("ClickstreamMapping: %v", ClickstreamMapping)
	GuestProfileMapping := BuildGuestProfileMapping()
	t.Logf("GuestProfileMapping: %v", GuestProfileMapping)
	HotelLoyaltyMapping := BuildHotelLoyaltyMapping()
	t.Logf("HotelLoyaltyMapping: %v", HotelLoyaltyMapping)
	HotelStayMapping := BuildHotelStayMapping()
	t.Logf("HotelStayMapping: %v", HotelStayMapping)
	HotelBookingMapping := BuildHotelBookingMapping()
	t.Logf("HotelBookingMapping: %v", HotelBookingMapping)
	PassengerProfileMapping := BuildPassengerProfileMapping()
	t.Logf("PassengerProfileMapping: %v", PassengerProfileMapping)
	CustomerServiceInteractionMapping := BuildCustomerServiceInteractionMapping()
	t.Logf("CustomerServiceInteractionMapping: %v", CustomerServiceInteractionMapping)
	LoyaltyTransactionMapping := BuildLoyaltyTransactionMapping()
	t.Logf("LoyaltyTransactionMapping: %v", LoyaltyTransactionMapping)
	AncillaryServiceMapping := BuildAncillaryServiceMapping()
	t.Logf("AncillaryServiceMapping: %v", AncillaryServiceMapping)
	AlternateProfileIdsMapping := BuildAlternateProfileIdMapping()
	t.Logf("AlternateProfileIdsMapping: %v", AlternateProfileIdsMapping)

	testMapping(BuildAirBookingMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/air_booking_fields.csv", t)
	testMapping(BuildAirLoyaltyMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/air_loyalty_fields.csv", t)
	testMapping(BuildClickstreamMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/clickstream_fields.csv", t)
	testMapping(BuildEmailHistoryMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/email_history_fields.csv", t)
	testMapping(BuildGuestProfileMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/guest_profile_fields.csv", t)
	testMapping(BuildHotelBookingMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/hotel_booking_fields.csv", t)
	testMapping(BuildHotelLoyaltyMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/hotel_loyalty_fields.csv", t)
	testMapping(BuildHotelStayMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/hotel_stay_fields.csv", t)
	testMapping(BuildPassengerProfileMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/pax_profile_fields.csv", t)
	testMapping(BuildPhoneHistoryMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/phone_history_fields.csv", t)
	testMapping(BuildCustomerServiceInteractionMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/customer_service_interaction_fields.csv", t)
	testMapping(BuildLoyaltyTransactionMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/loyalty_transaction_fields.csv", t)
	testMapping(BuildAncillaryServiceMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/ancillary_service_fields.csv", t)
	testMapping(BuildAlternateProfileIdMapping().GetSourceNames(), "../../../../ucp-backend/test_assets/alternate_profile_ids_fields.csv", t)
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
