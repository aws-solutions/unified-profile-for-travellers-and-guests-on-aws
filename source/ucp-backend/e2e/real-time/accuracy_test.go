// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"tah/upt/source/ucp-common/src/utils/config"
	"tah/upt/source/ucp-common/src/utils/test"
	"testing"
)

var rootDirectory, _ = config.AbsolutePathToProjectRootDir()
var testFilePath = rootDirectory + "source/test_data"
var testConfigs = []TestConfig{
	{
		Files: []string{testFilePath + "/real_time/guest_profile_customer_1.jsonl"},
		Profiles: map[string][]test.ExpectedField{
			"112779536": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "ROBERT"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "DOWNEY"},
				{ObjectType: "_profile", FieldName: "birthDate", FieldValue: "1888-11-03"},
				{ObjectType: "_profile", FieldName: "homePhoneNumber", FieldValue: "4844797454"},
				{ObjectType: "_profile", FieldName: "personalEmailAddress", FieldValue: "user1@example.com"},
				{ObjectType: "_profile", FieldName: "companyName", FieldValue: "NN"},
				{ObjectType: "_profile", FieldName: "addressLine1", FieldValue: "VESTBYDGVEGEN 300"},
				{ObjectType: "_profile", FieldName: "addressLine2", FieldValue: "NA"},
				{ObjectType: "_profile", FieldName: "city", FieldValue: "FREKHAUG"},
				{ObjectType: "_profile", FieldName: "stateCode", FieldValue: "MA"},
				{ObjectType: "_profile", FieldName: "postalCode", FieldValue: "5918"},
				{ObjectType: "_profile", FieldName: "countryCode", FieldValue: "NO"},
				{ObjectType: "_profile", FieldName: "bizAddressLine1", FieldValue: "MARIDALSVEIEN 14"},
				{ObjectType: "_profile", FieldName: "bizAddressLine2", FieldValue: "NA"},
				{ObjectType: "_profile", FieldName: "bizCity", FieldValue: "OSLO"},
				{ObjectType: "_profile", FieldName: "bizPostalCode", FieldValue: "0423"},
				{ObjectType: "_profile", FieldName: "bizCountryCode", FieldValue: "NO"},
				{ObjectType: "_profile", FieldName: "bizStateCode", FieldValue: "NY"},
				{ObjectType: "_profile", FieldName: "languageCode", FieldValue: "EN"},
				{ObjectType: "_profile", FieldName: "honorific", FieldValue: "MR"},
				{ObjectType: "hotel_loyalty", ObjectID: "22841965", FieldName: "level", FieldValue: "Blue"},
				{ObjectType: "hotel_loyalty", ObjectID: "22841965", FieldName: "joined", FieldValue: "2013-06-22T00:00:00Z"},
				{ObjectType: "alternate_profile_id", ObjectID: "port_MnXN3JLXMUELnPuqcz", FieldName: "description", FieldValue: "Wes Anderson whatever everyday chicharrones deep v deep v church-key chillwave hoodie drinking."},
			},
			//	This entry tests that the same profile can be searched by its ProfileID, as well as any AlternateProfileId
			"port_MnXN3JLXMUELnPuqcz": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "ROBERT"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "DOWNEY"},
				{ObjectType: "_profile", FieldName: "birthDate", FieldValue: "1888-11-03"},
				{ObjectType: "_profile", FieldName: "homePhoneNumber", FieldValue: "4844797454"},
				{ObjectType: "_profile", FieldName: "personalEmailAddress", FieldValue: "user1@example.com"},
				{ObjectType: "_profile", FieldName: "companyName", FieldValue: "NN"},
				{ObjectType: "_profile", FieldName: "addressLine1", FieldValue: "VESTBYDGVEGEN 300"},
				{ObjectType: "_profile", FieldName: "addressLine2", FieldValue: "NA"},
				{ObjectType: "_profile", FieldName: "city", FieldValue: "FREKHAUG"},
				{ObjectType: "_profile", FieldName: "stateCode", FieldValue: "MA"},
				{ObjectType: "_profile", FieldName: "postalCode", FieldValue: "5918"},
				{ObjectType: "_profile", FieldName: "countryCode", FieldValue: "NO"},
				{ObjectType: "_profile", FieldName: "bizAddressLine1", FieldValue: "MARIDALSVEIEN 14"},
				{ObjectType: "_profile", FieldName: "bizAddressLine2", FieldValue: "NA"},
				{ObjectType: "_profile", FieldName: "bizCity", FieldValue: "OSLO"},
				{ObjectType: "_profile", FieldName: "bizPostalCode", FieldValue: "0423"},
				{ObjectType: "_profile", FieldName: "bizCountryCode", FieldValue: "NO"},
				{ObjectType: "_profile", FieldName: "bizStateCode", FieldValue: "NY"},
				{ObjectType: "_profile", FieldName: "languageCode", FieldValue: "EN"},
				{ObjectType: "_profile", FieldName: "honorific", FieldValue: "MR"},
				{ObjectType: "hotel_loyalty", ObjectID: "22841965", FieldName: "level", FieldValue: "Blue"},
				{ObjectType: "hotel_loyalty", ObjectID: "22841965", FieldName: "joined", FieldValue: "2013-06-22T00:00:00Z"},
				{ObjectType: "alternate_profile_id", ObjectID: "port_MnXN3JLXMUELnPuqcz", FieldName: "description", FieldValue: "Wes Anderson whatever everyday chicharrones deep v deep v church-key chillwave hoodie drinking."},
			},
			"115791287": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "CHRIS"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "EVANS"},
				{ObjectType: "_profile", FieldName: "birthDate", FieldValue: "1888-01-26"},
				{ObjectType: "_profile", FieldName: "homePhoneNumber", FieldValue: "5964798339"},
				{ObjectType: "_profile", FieldName: "personalEmailAddress", FieldValue: "CHRIS.EVANS@example.com"},
				{ObjectType: "hotel_loyalty", ObjectID: "23140152", FieldName: "level", FieldValue: "Silver"},
				{ObjectType: "hotel_loyalty", ObjectID: "23140152", FieldName: "joined", FieldValue: "2007-06-28T00:00:00Z"},
			},
		},
	},
	{
		Files: []string{testFilePath + "/real_time/hotel_booking_customer_1.jsonl"},
		Profiles: map[string][]test.ExpectedField{
			"396422331": {
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "JACKSON"},
				{ObjectType: "_profile", FieldName: "homePhoneNumber", FieldValue: "6068707944"},
				{ObjectType: "_profile", FieldName: "businessPhoneNumber", FieldValue: "265-636-5055"},
				{ObjectType: "_profile", FieldName: "companyName", FieldValue: "MJ Lodging"},
				{ObjectType: "_profile", FieldName: "addressLine1", FieldValue: "171 e clinton st"},
				{ObjectType: "_profile", FieldName: "city", FieldValue: "Equality"},
				{ObjectType: "_profile", FieldName: "postalCode", FieldValue: "62934"},
				{ObjectType: "_profile", FieldName: "stateCode", FieldValue: "IL"},
				{ObjectType: "_profile", FieldName: "countryCode", FieldValue: "US"},
				{ObjectType: "_profile", FieldName: "bizAddressLine1", FieldValue: "9111 East 32nd St. North"},
				{ObjectType: "_profile", FieldName: "bizAddressLine2", FieldValue: "Suite 100"},
				{ObjectType: "_profile", FieldName: "bizCity", FieldValue: "Wichita"},
				{ObjectType: "_profile", FieldName: "bizPostalCode", FieldValue: "67226-2614"},
				{ObjectType: "_profile", FieldName: "bizStateCode", FieldValue: "KS"},
				{ObjectType: "_profile", FieldName: "bizCountryCode", FieldValue: "US"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "roomTypeCode", FieldValue: "SNK"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "ratePlanCode", FieldValue: "LCLC"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "status", FieldValue: "RESERVED"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "totalBeforeTax", FieldValue: "77"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "totalAfterTax", FieldValue: "91.82"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "nNights", FieldValue: "1"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "nGuests", FieldValue: "1"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "checkInDate", FieldValue: "2022-08-02"},
				{ObjectType: "alternate_profile_id", ObjectID: "keyboard_MpRiXMNqxYdFv9xf", FieldName: "description", FieldValue: "Selfies blue bottle heirloom pork belly organic selfies brunch pabst park sartorial."},
				{ObjectType: "alternate_profile_id", ObjectID: "matrix_B09Tq7eLKhnHGT9LU", FieldName: "description", FieldValue: "Occupy mumblecore messenger bag chia cronut cardigan whatever scenester dreamcatcher lumber."},
			},
			"keyboard_MpRiXMNqxYdFv9xf": {
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "JACKSON"},
				{ObjectType: "_profile", FieldName: "homePhoneNumber", FieldValue: "6068707944"},
				{ObjectType: "_profile", FieldName: "businessPhoneNumber", FieldValue: "265-636-5055"},
				{ObjectType: "_profile", FieldName: "companyName", FieldValue: "MJ Lodging"},
				{ObjectType: "_profile", FieldName: "addressLine1", FieldValue: "171 e clinton st"},
				{ObjectType: "_profile", FieldName: "city", FieldValue: "Equality"},
				{ObjectType: "_profile", FieldName: "postalCode", FieldValue: "62934"},
				{ObjectType: "_profile", FieldName: "stateCode", FieldValue: "IL"},
				{ObjectType: "_profile", FieldName: "countryCode", FieldValue: "US"},
				{ObjectType: "_profile", FieldName: "bizAddressLine1", FieldValue: "9111 East 32nd St. North"},
				{ObjectType: "_profile", FieldName: "bizAddressLine2", FieldValue: "Suite 100"},
				{ObjectType: "_profile", FieldName: "bizCity", FieldValue: "Wichita"},
				{ObjectType: "_profile", FieldName: "bizPostalCode", FieldValue: "67226-2614"},
				{ObjectType: "_profile", FieldName: "bizStateCode", FieldValue: "KS"},
				{ObjectType: "_profile", FieldName: "bizCountryCode", FieldValue: "US"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "roomTypeCode", FieldValue: "SNK"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "ratePlanCode", FieldValue: "LCLC"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "status", FieldValue: "RESERVED"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "totalBeforeTax", FieldValue: "77"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "totalAfterTax", FieldValue: "91.82"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "nNights", FieldValue: "1"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "nGuests", FieldValue: "1"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "checkInDate", FieldValue: "2022-08-02"},
				{ObjectType: "alternate_profile_id", ObjectID: "keyboard_MpRiXMNqxYdFv9xf", FieldName: "description", FieldValue: "Selfies blue bottle heirloom pork belly organic selfies brunch pabst park sartorial."},
				{ObjectType: "alternate_profile_id", ObjectID: "matrix_B09Tq7eLKhnHGT9LU", FieldName: "description", FieldValue: "Occupy mumblecore messenger bag chia cronut cardigan whatever scenester dreamcatcher lumber."},
			},
			"matrix_B09Tq7eLKhnHGT9LU": {
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "JACKSON"},
				{ObjectType: "_profile", FieldName: "homePhoneNumber", FieldValue: "6068707944"},
				{ObjectType: "_profile", FieldName: "businessPhoneNumber", FieldValue: "265-636-5055"},
				{ObjectType: "_profile", FieldName: "companyName", FieldValue: "MJ Lodging"},
				{ObjectType: "_profile", FieldName: "addressLine1", FieldValue: "171 e clinton st"},
				{ObjectType: "_profile", FieldName: "city", FieldValue: "Equality"},
				{ObjectType: "_profile", FieldName: "postalCode", FieldValue: "62934"},
				{ObjectType: "_profile", FieldName: "stateCode", FieldValue: "IL"},
				{ObjectType: "_profile", FieldName: "countryCode", FieldValue: "US"},
				{ObjectType: "_profile", FieldName: "bizAddressLine1", FieldValue: "9111 East 32nd St. North"},
				{ObjectType: "_profile", FieldName: "bizAddressLine2", FieldValue: "Suite 100"},
				{ObjectType: "_profile", FieldName: "bizCity", FieldValue: "Wichita"},
				{ObjectType: "_profile", FieldName: "bizPostalCode", FieldValue: "67226-2614"},
				{ObjectType: "_profile", FieldName: "bizStateCode", FieldValue: "KS"},
				{ObjectType: "_profile", FieldName: "bizCountryCode", FieldValue: "US"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "roomTypeCode", FieldValue: "SNK"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "ratePlanCode", FieldValue: "LCLC"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "status", FieldValue: "RESERVED"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "totalBeforeTax", FieldValue: "77"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "totalAfterTax", FieldValue: "91.82"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "nNights", FieldValue: "1"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "nGuests", FieldValue: "1"},
				{ObjectType: "hotel_booking", ObjectID: "TN776|826764986|1000|2022-08-02|lDb7YDd2EeiFaga6pL6yNg", FieldName: "checkInDate", FieldValue: "2022-08-02"},
				{ObjectType: "alternate_profile_id", ObjectID: "keyboard_MpRiXMNqxYdFv9xf", FieldName: "description", FieldValue: "Selfies blue bottle heirloom pork belly organic selfies brunch pabst park sartorial."},
				{ObjectType: "alternate_profile_id", ObjectID: "matrix_B09Tq7eLKhnHGT9LU", FieldName: "description", FieldValue: "Occupy mumblecore messenger bag chia cronut cardigan whatever scenester dreamcatcher lumber."},
			},
		},
	},
	{
		Files: []string{testFilePath + "/real_time/clickstream_customer_1.jsonl"},
		Profiles: map[string][]test.ExpectedField{
			"6FpjO7XiF3urkvTD8S468k": {
				{ObjectType: "clickstream", ObjectID: "6FpjO7XiF3urkvTD8S468k" + "_" + "2023-07-03T10:54:57.943Z", FieldName: "bookingId", FieldValue: "78214857"},
				{ObjectType: "clickstream", ObjectID: "6FpjO7XiF3urkvTD8S468k" + "_" + "2023-07-03T10:54:57.943Z", FieldName: "eventType", FieldValue: "confirm_booking"},
				{ObjectType: "clickstream", ObjectID: "6FpjO7XiF3urkvTD8S468k" + "_" + "2023-07-03T10:54:57.943Z", FieldName: "eventTimestamp", FieldValue: "2023-07-03T10:54:57.943Z"},
			},
			"3W10sBWGQlekfdH_FtNmHQ": {
				{ObjectType: "clickstream", ObjectID: "3W10sBWGQlekfdH_FtNmHQ" + "_" + "2023-07-09T02:56:11.727Z", FieldName: "eventType", FieldValue: "login"},
				{ObjectType: "clickstream", ObjectID: "3W10sBWGQlekfdH_FtNmHQ" + "_" + "2023-07-09T02:56:11.727Z", FieldName: "customerId", FieldValue: "334398523"},
			},
			"7ccqusDISSe-_8MPtehnVA": {
				{ObjectType: "clickstream", ObjectID: "7ccqusDISSe-_8MPtehnVA" + "_" + "2023-07-03T11:01:07.995Z", FieldName: "eventType", FieldValue: "view_product"},
				{ObjectType: "clickstream", ObjectID: "7ccqusDISSe-_8MPtehnVA" + "_" + "2023-07-03T11:01:07.995Z", FieldName: "hotelCode", FieldValue: "OH298"},
			},
			"cmqPNhl2IhukfT8zPTvyL.": {
				{ObjectType: "clickstream", ObjectID: "cmqPNhl2IhukfT8zPTvyL." + "_" + "2023-07-03T11:01:57.533Z", FieldName: "eventType", FieldValue: "view_product"},
				{ObjectType: "clickstream", ObjectID: "cmqPNhl2IhukfT8zPTvyL." + "_" + "2023-07-03T11:01:57.533Z", FieldName: "url", FieldValue: "https://www.example.com/any-town/any-product"},
			},
			"KLB2kfeGRjWiaXO-HA2HXA": {
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "eventType", FieldValue: "search_destination"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "ratePlanCode", FieldValue: "RACK"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "checkinDate", FieldValue: "2023-07-03"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "checkoutDate", FieldValue: "2023-07-07"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "numNights", FieldValue: "4"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "numGuestsAdult", FieldValue: "1"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "numGuestsChildren", FieldValue: "1"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "destination", FieldValue: "Washington"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "destination", FieldValue: "Washington"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "quantities", FieldValue: "1"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "poiId", FieldValue: "391082"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "geofenceLatitude", FieldValue: "38.89037"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "geofenceLongitude", FieldValue: "-77.03196"},
				{ObjectType: "clickstream", ObjectID: "KLB2kfeGRjWiaXO-HA2HXA" + "_" + "2023-07-02T10:52:29.448Z", FieldName: "languageCode", FieldValue: "EN"},
			},
			"4503604466619643": {
				{ObjectType: "clickstream", ObjectID: "4503604466619643" + "_" + "2024-06-25T00:07:27.000000Z", FieldName: "eventType", FieldValue: "sign_up"},
				{ObjectType: "clickstream", ObjectID: "4503604466619643" + "_" + "2024-06-25T00:08:15.000000Z", FieldName: "eventType", FieldValue: "login"},
				{ObjectType: "clickstream", ObjectID: "4503604466619643" + "_" + "2024-06-25T00:08:15.000000Z", FieldName: "customerId", FieldValue: "251696080"},
				{ObjectType: "alternate_profile_id", ObjectID: "feed_zF77wi0D8PBoaSuD2TZc", FieldName: "description", FieldValue: "Carry banjo pitchfork pinterest lumber cornhole gluten-free knausgaard pug deep v."},
			},
			"feed_zF77wi0D8PBoaSuD2TZc": {
				{ObjectType: "clickstream", ObjectID: "4503604466619643" + "_" + "2024-06-25T00:07:27.000000Z", FieldName: "eventType", FieldValue: "sign_up"},
				{ObjectType: "clickstream", ObjectID: "4503604466619643" + "_" + "2024-06-25T00:08:15.000000Z", FieldName: "eventType", FieldValue: "login"},
				{ObjectType: "clickstream", ObjectID: "4503604466619643" + "_" + "2024-06-25T00:08:15.000000Z", FieldName: "customerId", FieldValue: "251696080"},
				{ObjectType: "alternate_profile_id", ObjectID: "feed_zF77wi0D8PBoaSuD2TZc", FieldName: "description", FieldValue: "Carry banjo pitchfork pinterest lumber cornhole gluten-free knausgaard pug deep v."},
			},
			"4503604466620250": {
				{ObjectType: "clickstream", ObjectID: "4503604466620250" + "_" + "2024-06-25T00:00:45.000000Z", FieldName: "eventType", FieldValue: "search_flight"},
				{ObjectType: "clickstream", ObjectID: "4503604466620250" + "_" + "2024-06-25T00:00:45.000000Z", FieldName: "fareClass", FieldValue: "ECONOMY"},
				{ObjectType: "clickstream", ObjectID: "4503604466620250" + "_" + "2024-06-25T00:00:45.000000Z", FieldName: "destination", FieldValue: "SRQ"},
				{ObjectType: "clickstream", ObjectID: "4503604466620250" + "_" + "2024-06-25T00:00:45.000000Z", FieldName: "originDate", FieldValue: "2024-06-25"},
				{ObjectType: "clickstream", ObjectID: "4503604466620250" + "_" + "2024-06-25T00:00:45.000000Z", FieldName: "paxType", FieldValue: "1"},
				{ObjectType: "clickstream", ObjectID: "4503604466620250" + "_" + "2024-06-25T00:00:45.000000Z", FieldName: "origin", FieldValue: "HTS"},
				{ObjectType: "clickstream", ObjectID: "4503604466620250" + "_" + "2024-06-25T00:00:45.000000Z", FieldName: "lengthOfStay", FieldValue: "2"},
				{ObjectType: "clickstream", ObjectID: "4503604466620250" + "_" + "2024-06-25T00:00:45.000000Z", FieldName: "returnDate", FieldValue: "2024-06-27"},
			},
			"4503604466623956": {
				{ObjectType: "clickstream", ObjectID: "4503604466623956" + "_" + "2024-06-25T00:04:23.000000Z", FieldName: "eventType", FieldValue: "confirm_booking"},
				{ObjectType: "clickstream", ObjectID: "4503604466623956" + "_" + "2024-06-25T00:04:23.000000Z", FieldName: "orderPaymentType", FieldValue: "MC"},
				{ObjectType: "clickstream", ObjectID: "4503604466623956" + "_" + "2024-06-25T00:04:23.000000Z", FieldName: "transactionId", FieldValue: "3CFEF984-6F63-45BD-B97D-2D06467DAE10"},
				{ObjectType: "clickstream", ObjectID: "4503604466623956" + "_" + "2024-06-25T00:04:23.000000Z", FieldName: "productsPrices", FieldValue: "82600"},
				{ObjectType: "clickstream", ObjectID: "4503604466623956" + "_" + "2024-06-25T00:04:23.000000Z", FieldName: "error", FieldValue: "We were unable to assign some or all of your seats. Please visit Manage Reservations to try selecting seats again.,Manage Reservations"},
				{ObjectType: "clickstream", ObjectID: "4503604466623956" + "_" + "2024-06-25T00:04:23.000000Z", FieldName: "selectedSeats", FieldValue: "[Choose seat|37E,13B|Choose seat]"},
				{ObjectType: "clickstream", ObjectID: "4503604466623956" + "_" + "2024-06-25T00:04:23.000000Z", FieldName: "bookingId", FieldValue: "ENJX07"},
			},
		},
	},
	{
		Files: []string{testFilePath + "/real_time/ancillary.jsonl"},
		Profiles: map[string][]test.ExpectedField{
			"6850319251": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Shakira"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Maggio"},
				{ObjectType: "_profile", FieldName: "birthDate", FieldValue: "1981-06-25"},
				{ObjectType: "_profile", FieldName: "homePhoneNumber", FieldValue: "7461799285"},
				{ObjectType: "_profile", FieldName: "personalEmailAddress", FieldValue: "madiegrady@example.com"},
				{ObjectType: "_profile", FieldName: "companyName", FieldValue: "Arrive Labs"},
				{ObjectType: "_profile", FieldName: "bizAddressLine1", FieldValue: "43811 Drives ton"},
				{ObjectType: "_profile", FieldName: "bizAddressLine2", FieldValue: "more content line 2"},
				{ObjectType: "_profile", FieldName: "bizCity", FieldValue: "Mesa"},
				{ObjectType: "_profile", FieldName: "bizPostalCode", FieldValue: "31765"},
				{ObjectType: "_profile", FieldName: "bizCountryCode", FieldValue: "GE"},
				{ObjectType: "_profile", FieldName: "bizProvinceCode", FieldValue: "QC"},
				{ObjectType: "_profile", FieldName: "languageCode", FieldValue: "ro"},
				{ObjectType: "_profile", FieldName: "honorific", FieldValue: "Cllr"},
				{ObjectType: "air_loyalty", ObjectID: "7214879726", FieldName: "level", FieldValue: "bronze"},
				{ObjectType: "air_loyalty", ObjectID: "7214879726", FieldName: "joined", FieldValue: "2022-06-03T04:31:08Z"},
				{ObjectType: "air_loyalty", ObjectID: "7214879726", FieldName: "miles", FieldValue: "8684.233"},
				{ObjectType: "air_loyalty", ObjectID: "7214879726", FieldName: "programName", FieldValue: "elite"},
				{ObjectType: "ancillary_service", ObjectID: "5WYGQK", FieldName: "seatNumber", FieldValue: "E35"},
				{ObjectType: "ancillary_service", ObjectID: "5WYGQK", FieldName: "seatZone", FieldValue: "FrontAircraft"},
				{ObjectType: "ancillary_service", ObjectID: "5WYGQK", FieldName: "total", FieldValue: "22.604013"},
				{ObjectType: "ancillary_service", ObjectID: "5WYGQK", FieldName: "currency", FieldValue: "USD"},
				{ObjectType: "ancillary_service", ObjectID: "2GLN5E", FieldName: "seatNumber", FieldValue: "A39"},
				{ObjectType: "ancillary_service", ObjectID: "2GLN5E", FieldName: "seatZone", FieldValue: "FrontAircraft"},
				{ObjectType: "ancillary_service", ObjectID: "2GLN5E", FieldName: "total", FieldValue: "22.604013"},
				{ObjectType: "ancillary_service", ObjectID: "2GLN5E", FieldName: "currency", FieldValue: "USD"},
				{ObjectType: "ancillary_service", ObjectID: "2GLN5E", FieldName: "neighborFreeSeat", FieldValue: "true"},
				{ObjectType: "ancillary_service", ObjectID: "2GLN5E", FieldName: "upgradeAuction", FieldValue: "false"},
				{ObjectType: "ancillary_service", ObjectID: "UK9519", FieldName: "changeType", FieldValue: "Flight Date change"},
				{ObjectType: "ancillary_service", ObjectID: "UK9519", FieldName: "total", FieldValue: "253.37029"},
				{ObjectType: "ancillary_service", ObjectID: "UK9519", FieldName: "currency", FieldValue: "USD"},
				{ObjectType: "ancillary_service", ObjectID: "1AJ040", FieldName: "otherAncilliaryType", FieldValue: "unaccompaniedMinor"},
				{ObjectType: "ancillary_service", ObjectID: "1AJ040", FieldName: "total", FieldValue: "192.2113"},
				{ObjectType: "ancillary_service", ObjectID: "1AJ040", FieldName: "currency", FieldValue: "USD"},
				{ObjectType: "ancillary_service", ObjectID: "9ADJPS", FieldName: "priorityServiceType", FieldValue: "priorityBoarding"},
				{ObjectType: "ancillary_service", ObjectID: "9ADJPS", FieldName: "loungeAccess", FieldValue: "true"},
				{ObjectType: "ancillary_service", ObjectID: "9ADJPS", FieldName: "total", FieldValue: "250.94392"},
				{ObjectType: "ancillary_service", ObjectID: "9ADJPS", FieldName: "currency", FieldValue: "USD"},
			},
			"0982031183": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Leopold"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Wilderman"},
				{ObjectType: "_profile", FieldName: "personalEmailAddress", FieldValue: "air_traveller_1@example.com"},
				{ObjectType: "_profile", FieldName: "businessEmailAddress", FieldValue: "hillardosinski@example.com"},
				{ObjectType: "alternate_profile_id", ObjectID: "bus_ml68a7jHbkDOLU", FieldName: "description", FieldValue: "Godard austin tofu 90's kitsch XOXO schlitz chillwave fixie wayfarers."},
				{ObjectType: "alternate_profile_id", ObjectID: "pixel_0wevTwwQBiXze33a", FieldName: "description", FieldValue: "Five dollar toast PBR&B twee PBR&B heirloom cornhole schlitz paleo cliche sustainable."},
			},
			"bus_ml68a7jHbkDOLU": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Leopold"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Wilderman"},
				{ObjectType: "_profile", FieldName: "personalEmailAddress", FieldValue: "air_traveller_1@example.com"},
				{ObjectType: "_profile", FieldName: "businessEmailAddress", FieldValue: "hillardosinski@example.com"},
				{ObjectType: "alternate_profile_id", ObjectID: "bus_ml68a7jHbkDOLU", FieldName: "description", FieldValue: "Godard austin tofu 90's kitsch XOXO schlitz chillwave fixie wayfarers."},
				{ObjectType: "alternate_profile_id", ObjectID: "pixel_0wevTwwQBiXze33a", FieldName: "description", FieldValue: "Five dollar toast PBR&B twee PBR&B heirloom cornhole schlitz paleo cliche sustainable."},
			},
			"pixel_0wevTwwQBiXze33a": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Leopold"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Wilderman"},
				{ObjectType: "_profile", FieldName: "personalEmailAddress", FieldValue: "air_traveller_1@example.com"},
				{ObjectType: "_profile", FieldName: "businessEmailAddress", FieldValue: "hillardosinski@example.com"},
				{ObjectType: "alternate_profile_id", ObjectID: "bus_ml68a7jHbkDOLU", FieldName: "description", FieldValue: "Godard austin tofu 90's kitsch XOXO schlitz chillwave fixie wayfarers."},
				{ObjectType: "alternate_profile_id", ObjectID: "pixel_0wevTwwQBiXze33a", FieldName: "description", FieldValue: "Five dollar toast PBR&B twee PBR&B heirloom cornhole schlitz paleo cliche sustainable."},
			},
		},
	},
	{
		Files: []string{testFilePath + "/real_time/csi.jsonl"},
		Profiles: map[string][]test.ExpectedField{
			"d55a8191-bbf9-48e6-a2ad-cac5cd649500": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Sophie"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Bernier"},
				{ObjectType: "alternate_profile_id", ObjectID: "albatross_a8z0iRYB99Yl34", FieldName: "description", FieldValue: "Kitsch wayfarers cray gentrify humblebrag organic lumber shoreditch disrupt."},
				{ObjectType: "alternate_profile_id", ObjectID: "program_xGZhaUV1nDK", FieldName: "description", FieldValue: "Butcher keffiyeh bespoke art party heirloom artisan actually bushwick locavore try-hard."},
				{ObjectType: "alternate_profile_id", ObjectID: "harddrive_46YkXJQqR7n", FieldName: "description", FieldValue: "Cleanse roof twee hoodie +1 austin hashtag master park cold-pressed."},
			},
			"albatross_a8z0iRYB99Yl34": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Sophie"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Bernier"},
				{ObjectType: "alternate_profile_id", ObjectID: "albatross_a8z0iRYB99Yl34", FieldName: "description", FieldValue: "Kitsch wayfarers cray gentrify humblebrag organic lumber shoreditch disrupt."},
				{ObjectType: "alternate_profile_id", ObjectID: "program_xGZhaUV1nDK", FieldName: "description", FieldValue: "Butcher keffiyeh bespoke art party heirloom artisan actually bushwick locavore try-hard."},
				{ObjectType: "alternate_profile_id", ObjectID: "harddrive_46YkXJQqR7n", FieldName: "description", FieldValue: "Cleanse roof twee hoodie +1 austin hashtag master park cold-pressed."},
			},
			"program_xGZhaUV1nDK": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Sophie"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Bernier"},
				{ObjectType: "alternate_profile_id", ObjectID: "albatross_a8z0iRYB99Yl34", FieldName: "description", FieldValue: "Kitsch wayfarers cray gentrify humblebrag organic lumber shoreditch disrupt."},
				{ObjectType: "alternate_profile_id", ObjectID: "program_xGZhaUV1nDK", FieldName: "description", FieldValue: "Butcher keffiyeh bespoke art party heirloom artisan actually bushwick locavore try-hard."},
				{ObjectType: "alternate_profile_id", ObjectID: "harddrive_46YkXJQqR7n", FieldName: "description", FieldValue: "Cleanse roof twee hoodie +1 austin hashtag master park cold-pressed."},
			},
			"harddrive_46YkXJQqR7n": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Sophie"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Bernier"},
				{ObjectType: "alternate_profile_id", ObjectID: "albatross_a8z0iRYB99Yl34", FieldName: "description", FieldValue: "Kitsch wayfarers cray gentrify humblebrag organic lumber shoreditch disrupt."},
				{ObjectType: "alternate_profile_id", ObjectID: "program_xGZhaUV1nDK", FieldName: "description", FieldValue: "Butcher keffiyeh bespoke art party heirloom artisan actually bushwick locavore try-hard."},
				{ObjectType: "alternate_profile_id", ObjectID: "harddrive_46YkXJQqR7n", FieldName: "description", FieldValue: "Cleanse roof twee hoodie +1 austin hashtag master park cold-pressed."},
			},
			"84608747-9fa9a3da-19c2-4ce0-bd09-800d1ea8ecbb": {
				{ObjectType: "customer_service_interaction", ObjectID: "84608747-9fa9a3da-19c2-4ce0-bd09-800d1ea8ecbb", FieldName: "bookingId", FieldValue: "2V25EB"},
				{ObjectType: "customer_service_interaction", ObjectID: "84608747-9fa9a3da-19c2-4ce0-bd09-800d1ea8ecbb", FieldName: "conversationSummary", FieldValue: "Customer's name:\nN/A\n\nCall reason:\nCustomer wants to locate their reservation using the ticket number.\n\nResolution steps:\n- Agent asks for customer's confirmation or reference number, full name, and a summary of the assistance needed.\n- Agent informs customer that their agent will message them as soon as possible.\n- Customer provides the ticket number and the full name of the passenger and the email address associated with the reservation.\n- Agent confirms that the reservation is still not found.\n- Agent checks the reservation on the AnyCompany2 website and finds that it is still confirm.\n- Agent transfers customer to support to further assist with their reservation.\n\nConversation results:\n- Customer's reservation is not found and they are transferred to support.\n- Agent apologizes for the inconvenience and informs customer that the confirmation number needs to be generated before the flight is confirmed.\n"},
			},
		},
	},
	{
		Files: []string{testFilePath + "/real_time/hotel_stay.jsonl"},
		Profiles: map[string][]test.ExpectedField{
			"2950645095": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Myrtie"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Sanford"},
				{ObjectType: "alternate_profile_id", ObjectID: "driver_rOZ394Wf7t9DXJPCVmu", FieldName: "description", FieldValue: "Food truck chillwave franzen franzen meditation organic dreamcatcher truffaut next level tumblr."},
			},
			"driver_rOZ394Wf7t9DXJPCVmu": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Myrtie"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Sanford"},
				{ObjectType: "alternate_profile_id", ObjectID: "driver_rOZ394Wf7t9DXJPCVmu", FieldName: "description", FieldValue: "Food truck chillwave franzen franzen meditation organic dreamcatcher truffaut next level tumblr."},
			},
		},
	},
	{
		Files: []string{testFilePath + "/real_time/pax_profile.jsonl"},
		Profiles: map[string][]test.ExpectedField{
			"6207056237": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Sylvester"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Turcotte"},
				{ObjectType: "alternate_profile_id", ObjectID: "alarm_Q6DX2U8fJpHUTPN", FieldName: "description", FieldValue: "Freegan dreamcatcher pork belly bushwick beard authentic trust fund waistcoat cred forage."},
			},
			"alarm_Q6DX2U8fJpHUTPN": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Sylvester"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Turcotte"},
				{ObjectType: "alternate_profile_id", ObjectID: "alarm_Q6DX2U8fJpHUTPN", FieldName: "description", FieldValue: "Freegan dreamcatcher pork belly bushwick beard authentic trust fund waistcoat cred forage."},
			},
			"251696080": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Doris"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Walls"},
				{ObjectType: "alternate_profile_id", ObjectID: "alarm_lO7PD1LrKhZ63Q2", FieldName: "description", FieldValue: "Freegan dreamcatcher pork belly bushwick beard authentic trust fund waistcoat cred forage."},
			},
			"alarm_lO7PD1LrKhZ63Q2": {
				{ObjectType: "_profile", FieldName: "firstName", FieldValue: "Doris"},
				{ObjectType: "_profile", FieldName: "lastName", FieldValue: "Walls"},
				{ObjectType: "alternate_profile_id", ObjectID: "alarm_lO7PD1LrKhZ63Q2", FieldName: "description", FieldValue: "Freegan dreamcatcher pork belly bushwick beard authentic trust fund waistcoat cred forage."},
			},
		},
	},
}

func TestCustomer1(t *testing.T) {
	// Load configs
	envConfig, infraCfg, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("[%s] Error loading configs: %v", t.Name(), err)
	}
	services, err := InitServices()
	if err != nil {
		t.Fatalf("[%s] Error initializing services: %v", t.Name(), err)
	}
	adminConfig, err := Setup(t, services, envConfig, infraCfg)
	if err != nil {
		t.Fatalf("[%s] Could not Setup test environment: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = Cleanup(t, adminConfig, services, envConfig, infraCfg)
		if err != nil {
			t.Errorf("[%s] Could not Cleanup test environment: %v", t.Name(), err)
		}
	})
	reqHelper := test.EndToEndRequest{
		AccessToken: adminConfig.AccessToken,
		BaseUrl:     infraCfg.ApiBaseUrl,
		DomainName:  adminConfig.DomainName,
	}
	for _, testConfig := range testConfigs {
		err := IngestFromFiles(t, testConfig.Files, adminConfig, services)
		if err != nil {
			t.Fatalf("[%s] Error ingesting files: %v", t.Name(), err)
		}
	}
	for _, testConfig := range testConfigs {
		for profileId, expectedFields := range testConfig.Profiles {
			refs := buildProfileObjectRefs(expectedFields)
			profile, err := waitForIngestionCompletion(reqHelper, profileId, refs, 120)
			if err == nil {
				for _, expected := range expectedFields {
					test.CheckExpected(t, profile, expected)
				}
			} else {
				t.Fatalf("[%s] Error waiting for profile creation: %v", t.Name(), err)
			}
		}
	}
}
