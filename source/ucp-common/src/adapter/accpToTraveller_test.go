// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"time"

	"reflect"
	// adapter "tah/upt/source/ucp-common/src/adapter"
	// constant "tah/upt/source/ucp-common/src/constant/admin"
	"testing"
)

func TestProfileToTraveller(t *testing.T) {
	profileByte, err := os.ReadFile("../../../test_data/accp_profiles/test_profile1.json")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	var profile profilemodel.Profile
	err = json.Unmarshal(profileByte, &profile)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	travellerByte, err := os.ReadFile("../../../test_data/accp_profiles/testTraveller_1.json")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	var traveller model.Traveller
	err = json.Unmarshal(travellerByte, &traveller)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	travellerRead := ProfilesToTravellers(core.NewTransaction(t.Name(), "", core.LogLevelDebug), []profilemodel.Profile{profile})
	tru := reflect.DeepEqual(travellerRead[0], traveller)
	if !tru {
		t1, _ := json.Marshal(traveller)
		t2, _ := json.Marshal(travellerRead[0])
		log.Println("EXPECTED: ", string(t1))
		log.Println("GOT: ", string(t2))
		t.Errorf("Expecting traveller has invalid schema")
	}

}

// /////////////////////////
// this SQL queries were originally used for Redshift integration which is no longer used in the solution. we keep the logic for future use.
// DO NOT update these manually field by field since this would be very cumbersome. The test TestSqlInsertStatement will output proposed query
// if failing. you can then copy these from the test output (between the COPY BELOW and COPY ABOVE statements).
// ///////////////////////////////
var traveller_sql = `INSERT INTO my_domain ("model_version", "last_updated_by", "domain", "merged_in", "connect_id", "traveller_id", "pssid", "gdsid", "pmsid", "crsid", "honorific", "first_name", "middle_name", "last_name", "gender", "pronoun", "job_title", "company_name", "phone_number", "mobile_phone_number", "home_phone_number", "business_phone_number", "personal_email_address", "business_email_address", "nationality_code", "nationality_name", "language_code", "language_name", "home_address_address1", "home_address_address2", "home_address_address3", "home_address_address4", "home_address_city", "home_address_state", "home_address_province", "home_address_postal_code", "home_address_country", "business_address_address1", "business_address_address2", "business_address_address3", "business_address_address4", "business_address_city", "business_address_state", "business_address_province", "business_address_postal_code", "business_address_country", "mailing_address_address1", "mailing_address_address2", "mailing_address_address3", "mailing_address_address4", "mailing_address_city", "mailing_address_state", "mailing_address_province", "mailing_address_postal_code", "mailing_address_country", "billing_address_address1", "billing_address_address2", "billing_address_address3", "billing_address_address4", "billing_address_city", "billing_address_state", "billing_address_province", "billing_address_postal_code", "billing_address_country") VALUES ('', '', '', '', 'abcdefgh', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '');`
var traveller_update_sql = `UPDATE my_domain SET "model_version" = '', "last_updated_by" = '', "domain" = '', "merged_in" = '', "connect_id" = 'abcdefgh', "traveller_id" = '', "pssid" = '', "gdsid" = '', "pmsid" = '', "crsid" = '', "honorific" = '', "first_name" = '', "middle_name" = '', "last_name" = '', "gender" = '', "pronoun" = '', "job_title" = '', "company_name" = '', "phone_number" = '', "mobile_phone_number" = '', "home_phone_number" = '', "business_phone_number" = '', "personal_email_address" = '', "business_email_address" = '', "nationality_code" = '', "nationality_name" = '', "language_code" = '', "language_name" = '', "home_address_address1" = '', "home_address_address2" = '', "home_address_address3" = '', "home_address_address4" = '', "home_address_city" = '', "home_address_state" = '', "home_address_province" = '', "home_address_postal_code" = '', "home_address_country" = '', "business_address_address1" = '', "business_address_address2" = '', "business_address_address3" = '', "business_address_address4" = '', "business_address_city" = '', "business_address_state" = '', "business_address_province" = '', "business_address_postal_code" = '', "business_address_country" = '', "mailing_address_address1" = '', "mailing_address_address2" = '', "mailing_address_address3" = '', "mailing_address_address4" = '', "mailing_address_city" = '', "mailing_address_state" = '', "mailing_address_province" = '', "mailing_address_postal_code" = '', "mailing_address_country" = '', "billing_address_address1" = '', "billing_address_address2" = '', "billing_address_address3" = '', "billing_address_address4" = '', "billing_address_city" = '', "billing_address_state" = '', "billing_address_province" = '', "billing_address_postal_code" = '', "billing_address_country" = '' WHERE connect_id = 'abcdefgh'`
var traveller_delete_sql = `DELETE FROM my_domain WHERE connect_id = 'abcdefgh'`
var traveller_create_table_sql = `CREATE TABLE my_domain ("model_version" VARCHAR(255),"last_updated_by" VARCHAR(255),"domain" VARCHAR(255),"merged_in" VARCHAR(255),"connect_id" VARCHAR(255),"traveller_id" VARCHAR(255),"pssid" VARCHAR(255),"gdsid" VARCHAR(255),"pmsid" VARCHAR(255),"crsid" VARCHAR(255),"honorific" VARCHAR(255),"first_name" VARCHAR(255),"middle_name" VARCHAR(255),"last_name" VARCHAR(255),"gender" VARCHAR(255),"pronoun" VARCHAR(255),"job_title" VARCHAR(255),"company_name" VARCHAR(255),"phone_number" VARCHAR(255),"mobile_phone_number" VARCHAR(255),"home_phone_number" VARCHAR(255),"business_phone_number" VARCHAR(255),"personal_email_address" VARCHAR(255),"business_email_address" VARCHAR(255),"nationality_code" VARCHAR(255),"nationality_name" VARCHAR(255),"language_code" VARCHAR(255),"language_name" VARCHAR(255),"home_address_address1" VARCHAR(255),"home_address_address2" VARCHAR(255),"home_address_address3" VARCHAR(255),"home_address_address4" VARCHAR(255),"home_address_city" VARCHAR(255),"home_address_state" VARCHAR(255),"home_address_province" VARCHAR(255),"home_address_postal_code" VARCHAR(255),"home_address_country" VARCHAR(255),"business_address_address1" VARCHAR(255),"business_address_address2" VARCHAR(255),"business_address_address3" VARCHAR(255),"business_address_address4" VARCHAR(255),"business_address_city" VARCHAR(255),"business_address_state" VARCHAR(255),"business_address_province" VARCHAR(255),"business_address_postal_code" VARCHAR(255),"business_address_country" VARCHAR(255),"mailing_address_address1" VARCHAR(255),"mailing_address_address2" VARCHAR(255),"mailing_address_address3" VARCHAR(255),"mailing_address_address4" VARCHAR(255),"mailing_address_city" VARCHAR(255),"mailing_address_state" VARCHAR(255),"mailing_address_province" VARCHAR(255),"mailing_address_postal_code" VARCHAR(255),"mailing_address_country" VARCHAR(255),"billing_address_address1" VARCHAR(255),"billing_address_address2" VARCHAR(255),"billing_address_address3" VARCHAR(255),"billing_address_address4" VARCHAR(255),"billing_address_city" VARCHAR(255),"billing_address_state" VARCHAR(255),"billing_address_province" VARCHAR(255),"billing_address_postal_code" VARCHAR(255),"billing_address_country" VARCHAR(255))`
var air_booking_sql = `INSERT INTO my_domain_air_booking ("accp_object_id", "overall_confidence_score", "traveller_id", "booking_id", "segment_id", "from", "to", "flight_number", "departure_date", "departure_time", "arrival_date", "arrival_time", "channel", "status", "last_updated", "last_updated_by", "total_price", "traveller_price", "booker_id", "creation_channel_id", "last_update_channel_id", "day_of_travel_email", "day_of_travel_phone", "first_name", "last_name", "profile_connect_id") VALUES ('', '0', '', '', '', '', '', '', '', '', '', '', '', '', '0001-01-01 00:00:00', '', '0', '0', '', '', '', '', '', '', '', '');`
var air_booking_update_sql = `UPDATE my_domain_air_booking SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "booking_id" = '', "segment_id" = '', "from" = '', "to" = '', "flight_number" = '', "departure_date" = '', "departure_time" = '', "arrival_date" = '', "arrival_time" = '', "channel" = '', "status" = '', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "total_price" = '0', "traveller_price" = '0', "booker_id" = '', "creation_channel_id" = '', "last_update_channel_id" = '', "day_of_travel_email" = '', "day_of_travel_phone" = '', "first_name" = '', "last_name" = '', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var air_booking_delete_sql = `DELETE FROM my_domain_air_booking WHERE accp_object_id = '' AND profile_connect_id = ''`
var air_booking_create_table_sql = `CREATE TABLE my_domain_air_booking ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"booking_id" VARCHAR(255),"segment_id" VARCHAR(255),"from" VARCHAR(255),"to" VARCHAR(255),"flight_number" VARCHAR(255),"departure_date" VARCHAR(255),"departure_time" VARCHAR(255),"arrival_date" VARCHAR(255),"arrival_time" VARCHAR(255),"channel" VARCHAR(255),"status" VARCHAR(255),"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"total_price" DECIMAL(10,2),"traveller_price" DECIMAL(10,2),"booker_id" VARCHAR(255),"creation_channel_id" VARCHAR(255),"last_update_channel_id" VARCHAR(255),"day_of_travel_email" VARCHAR(255),"day_of_travel_phone" VARCHAR(255),"first_name" VARCHAR(255),"last_name" VARCHAR(255),"profile_connect_id" VARCHAR(255))`
var hotel_booking_sql = `INSERT INTO my_domain_hotel_booking ("accp_object_id", "overall_confidence_score", "traveller_id", "booking_id", "hotel_code", "num_nights", "num_guests", "product_id", "check_in_date", "room_type_code", "room_type_name", "room_type_description", "rate_plan_code", "rate_plan_name", "rate_plan_description", "attribute_codes", "attribute_names", "attribute_descriptions", "add_on_codes", "add_on_names", "add_on_descriptions", "last_updated", "last_updated_by", "price_per_traveller", "booker_id", "status", "total_segment_before_tax", "total_segment_after_tax", "creation_channel_id", "last_update_channel_id", "first_name", "last_name", "profile_connect_id") VALUES ('', '0', '', '', '', '0', '0', '', '0001-01-01 00:00:00', '', '', '', '', '', '', '', '', '', '', '', '', '0001-01-01 00:00:00', '', '', '', '', '0', '0', '', '', '', '', '');`
var hotel_booking_update_sql = `UPDATE my_domain_hotel_booking SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "booking_id" = '', "hotel_code" = '', "num_nights" = '0', "num_guests" = '0', "product_id" = '', "check_in_date" = '0001-01-01 00:00:00', "room_type_code" = '', "room_type_name" = '', "room_type_description" = '', "rate_plan_code" = '', "rate_plan_name" = '', "rate_plan_description" = '', "attribute_codes" = '', "attribute_names" = '', "attribute_descriptions" = '', "add_on_codes" = '', "add_on_names" = '', "add_on_descriptions" = '', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "price_per_traveller" = '', "booker_id" = '', "status" = '', "total_segment_before_tax" = '0', "total_segment_after_tax" = '0', "creation_channel_id" = '', "last_update_channel_id" = '', "first_name" = '', "last_name" = '', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var hotel_booking_delete_sql = `DELETE FROM my_domain_hotel_booking WHERE accp_object_id = '' AND profile_connect_id = ''`
var hotel_booking_create_table_sql = `CREATE TABLE my_domain_hotel_booking ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"booking_id" VARCHAR(255),"hotel_code" VARCHAR(255),"num_nights" INTEGER,"num_guests" INTEGER,"product_id" VARCHAR(255),"check_in_date" TIMESTAMP,"room_type_code" VARCHAR(255),"room_type_name" VARCHAR(255),"room_type_description" VARCHAR(255),"rate_plan_code" VARCHAR(255),"rate_plan_name" VARCHAR(255),"rate_plan_description" VARCHAR(255),"attribute_codes" VARCHAR(255),"attribute_names" VARCHAR(255),"attribute_descriptions" VARCHAR(255),"add_on_codes" VARCHAR(255),"add_on_names" VARCHAR(255),"add_on_descriptions" VARCHAR(255),"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"price_per_traveller" VARCHAR(255),"booker_id" VARCHAR(255),"status" VARCHAR(255),"total_segment_before_tax" DECIMAL(10,2),"total_segment_after_tax" DECIMAL(10,2),"creation_channel_id" VARCHAR(255),"last_update_channel_id" VARCHAR(255),"first_name" VARCHAR(255),"last_name" VARCHAR(255),"profile_connect_id" VARCHAR(255))`
var air_loyalty_sql = `INSERT INTO my_domain_air_loyalty ("accp_object_id", "overall_confidence_score", "traveller_id", "loyalty_id", "program_name", "miles", "miles_to_next_level", "level", "joined", "last_updated", "last_updated_by", "enrollment_source", "currency", "amount", "account_status", "reason_for_close", "language_preference", "display_preference", "meal_preference", "seat_preference", "home_airport", "date_time_format_pref", "cabin_preference", "fare_type_preference", "expert_mode", "privacy_indicator", "car_preference_vendor", "car_preference_type", "special_accommodation1", "special_accommodation2", "special_accommodation3", "special_accommodation4", "special_accommodation5", "marketing_opt_ins", "renew_date", "next_bill_amount", "clear_enroll_date", "clear_renew_date", "clear_tier_level", "clear_next_bill_amount", "clear_is_active", "clear_auto_renew", "clear_has_biometrics", "clear_has_partner_pricing", "tsa_type", "tsa_seq_num", "tsa_number", "profile_connect_id") VALUES ('', '0', '', '', '', '', '', '', '0001-01-01 00:00:00', '0001-01-01 00:00:00', '', '', '', '0', '', '', '', '', '', '', '', '', '', '', 'false', '', '', '', '', '', '', '', '', '', '', '0', '', '', '', '0', 'false', 'false', 'false', 'false', '', '', '', '');`
var air_loyalty_update_sql = `UPDATE my_domain_air_loyalty SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "loyalty_id" = '', "program_name" = '', "miles" = '', "miles_to_next_level" = '', "level" = '', "joined" = '0001-01-01 00:00:00', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "enrollment_source" = '', "currency" = '', "amount" = '0', "account_status" = '', "reason_for_close" = '', "language_preference" = '', "display_preference" = '', "meal_preference" = '', "seat_preference" = '', "home_airport" = '', "date_time_format_pref" = '', "cabin_preference" = '', "fare_type_preference" = '', "expert_mode" = 'false', "privacy_indicator" = '', "car_preference_vendor" = '', "car_preference_type" = '', "special_accommodation1" = '', "special_accommodation2" = '', "special_accommodation3" = '', "special_accommodation4" = '', "special_accommodation5" = '', "marketing_opt_ins" = '', "renew_date" = '', "next_bill_amount" = '0', "clear_enroll_date" = '', "clear_renew_date" = '', "clear_tier_level" = '', "clear_next_bill_amount" = '0', "clear_is_active" = 'false', "clear_auto_renew" = 'false', "clear_has_biometrics" = 'false', "clear_has_partner_pricing" = 'false', "tsa_type" = '', "tsa_seq_num" = '', "tsa_number" = '', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var air_loyalty_delete_sql = `DELETE FROM my_domain_air_loyalty WHERE accp_object_id = '' AND profile_connect_id = ''`
var air_loyalty_create_table_sql = `CREATE TABLE my_domain_air_loyalty ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"loyalty_id" VARCHAR(255),"program_name" VARCHAR(255),"miles" VARCHAR(255),"miles_to_next_level" VARCHAR(255),"level" VARCHAR(255),"joined" TIMESTAMP,"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"enrollment_source" VARCHAR(255),"currency" VARCHAR(255),"amount" DECIMAL(10,2),"account_status" VARCHAR(255),"reason_for_close" VARCHAR(255),"language_preference" VARCHAR(255),"display_preference" VARCHAR(255),"meal_preference" VARCHAR(255),"seat_preference" VARCHAR(255),"home_airport" VARCHAR(255),"date_time_format_pref" VARCHAR(255),"cabin_preference" VARCHAR(255),"fare_type_preference" VARCHAR(255),"expert_mode" BOOLEAN,"privacy_indicator" VARCHAR(255),"car_preference_vendor" VARCHAR(255),"car_preference_type" VARCHAR(255),"special_accommodation1" VARCHAR(255),"special_accommodation2" VARCHAR(255),"special_accommodation3" VARCHAR(255),"special_accommodation4" VARCHAR(255),"special_accommodation5" VARCHAR(255),"marketing_opt_ins" VARCHAR(255),"renew_date" VARCHAR(255),"next_bill_amount" DECIMAL(10,2),"clear_enroll_date" VARCHAR(255),"clear_renew_date" VARCHAR(255),"clear_tier_level" VARCHAR(255),"clear_next_bill_amount" DECIMAL(10,2),"clear_is_active" BOOLEAN,"clear_auto_renew" BOOLEAN,"clear_has_biometrics" BOOLEAN,"clear_has_partner_pricing" BOOLEAN,"tsa_type" VARCHAR(255),"tsa_seq_num" VARCHAR(255),"tsa_number" VARCHAR(255),"profile_connect_id" VARCHAR(255))`
var hotel_loyalty_sql = `INSERT INTO my_domain_hotel_loyalty ("accp_object_id", "overall_confidence_score", "traveller_id", "loyalty_id", "program_name", "points", "units", "points_to_next_level", "level", "joined", "last_updated", "last_updated_by", "profile_connect_id") VALUES ('', '0', '', '', '', '', '', '', '', '0001-01-01 00:00:00', '0001-01-01 00:00:00', '', '');`
var hotel_loyalty_update_sql = `UPDATE my_domain_hotel_loyalty SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "loyalty_id" = '', "program_name" = '', "points" = '', "units" = '', "points_to_next_level" = '', "level" = '', "joined" = '0001-01-01 00:00:00', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var hotel_loyalty_delete_sql = `DELETE FROM my_domain_hotel_loyalty WHERE accp_object_id = '' AND profile_connect_id = ''`
var hotel_loyalty_create_table_sql = `CREATE TABLE my_domain_hotel_loyalty ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"loyalty_id" VARCHAR(255),"program_name" VARCHAR(255),"points" VARCHAR(255),"units" VARCHAR(255),"points_to_next_level" VARCHAR(255),"level" VARCHAR(255),"joined" TIMESTAMP,"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"profile_connect_id" VARCHAR(255))`
var clickstream_sql = `INSERT INTO my_domain_clickstream ("accp_object_id", "overall_confidence_score", "traveller_id", "last_updated", "session_id", "event_timestamp", "event_type", "event_version", "arrival_timestamp", "user_agent", "custom_event_name", "error", "customer_birthdate", "customer_country", "customer_email", "customer_first_name", "customer_gender", "customer_id", "customer_last_name", "customer_nationality", "customer_phone", "customer_type", "customer_loyalty_id", "language_code", "currency", "products", "quantities", "products_prices", "ecommerce_action", "order_payment_type", "order_promo_code", "page_name", "page_type_environment", "transaction_id", "booking_id", "geofence_latitude", "geofence_longitude", "geofence_id", "geofence_name", "poi_id", "url", "custom", "fare_class", "fare_type", "flight_segments_departure_date_time", "flight_segments_arrival_date_time", "flight_segments", "flight_segment_sku", "flight_route", "flight_numbers", "flight_market", "flight_type", "origin_date", "origin_date_time", "return_date", "return_date_time", "return_flight_route", "num_pax_adults", "num_pax_inf", "num_pax_children", "pax_type", "total_passengers", "length_of_stay", "origin", "selected_seats", "room_type", "rate_plan", "checkin_date", "checkout_date", "num_nights", "num_guests", "num_guests_adult", "num_guests_children", "hotel_code", "hotel_code_list", "hotel_name", "destination", "profile_connect_id") VALUES ('', '0', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '', '', '0001-01-01 00:00:00', '', '', '', '0001-01-01 00:00:00', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '', '0', '0', '0', '', '0', '0', '', '', '', '', '0001-01-01 00:00:00', '0001-01-01 00:00:00', '0', '0', '0', '0', '', '', '', '', '');`
var clickstream_update_sql = `UPDATE my_domain_clickstream SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "last_updated" = '0001-01-01 00:00:00', "session_id" = '', "event_timestamp" = '0001-01-01 00:00:00', "event_type" = '', "event_version" = '', "arrival_timestamp" = '0001-01-01 00:00:00', "user_agent" = '', "custom_event_name" = '', "error" = '', "customer_birthdate" = '0001-01-01 00:00:00', "customer_country" = '', "customer_email" = '', "customer_first_name" = '', "customer_gender" = '', "customer_id" = '', "customer_last_name" = '', "customer_nationality" = '', "customer_phone" = '', "customer_type" = '', "customer_loyalty_id" = '', "language_code" = '', "currency" = '', "products" = '', "quantities" = '', "products_prices" = '', "ecommerce_action" = '', "order_payment_type" = '', "order_promo_code" = '', "page_name" = '', "page_type_environment" = '', "transaction_id" = '', "booking_id" = '', "geofence_latitude" = '', "geofence_longitude" = '', "geofence_id" = '', "geofence_name" = '', "poi_id" = '', "url" = '', "custom" = '', "fare_class" = '', "fare_type" = '', "flight_segments_departure_date_time" = '', "flight_segments_arrival_date_time" = '', "flight_segments" = '', "flight_segment_sku" = '', "flight_route" = '', "flight_numbers" = '', "flight_market" = '', "flight_type" = '', "origin_date" = '', "origin_date_time" = '0001-01-01 00:00:00', "return_date" = '', "return_date_time" = '0001-01-01 00:00:00', "return_flight_route" = '', "num_pax_adults" = '0', "num_pax_inf" = '0', "num_pax_children" = '0', "pax_type" = '', "total_passengers" = '0', "length_of_stay" = '0', "origin" = '', "selected_seats" = '', "room_type" = '', "rate_plan" = '', "checkin_date" = '0001-01-01 00:00:00', "checkout_date" = '0001-01-01 00:00:00', "num_nights" = '0', "num_guests" = '0', "num_guests_adult" = '0', "num_guests_children" = '0', "hotel_code" = '', "hotel_code_list" = '', "hotel_name" = '', "destination" = '', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var clickstream_delete_sql = `DELETE FROM my_domain_clickstream WHERE accp_object_id = '' AND profile_connect_id = ''`
var clickstream_create_table_sql = `CREATE TABLE my_domain_clickstream ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"last_updated" TIMESTAMP,"session_id" VARCHAR(255),"event_timestamp" TIMESTAMP,"event_type" VARCHAR(255),"event_version" VARCHAR(255),"arrival_timestamp" TIMESTAMP,"user_agent" VARCHAR(255),"custom_event_name" VARCHAR(255),"error" VARCHAR(255),"customer_birthdate" TIMESTAMP,"customer_country" VARCHAR(255),"customer_email" VARCHAR(255),"customer_first_name" VARCHAR(255),"customer_gender" VARCHAR(255),"customer_id" VARCHAR(255),"customer_last_name" VARCHAR(255),"customer_nationality" VARCHAR(255),"customer_phone" VARCHAR(255),"customer_type" VARCHAR(255),"customer_loyalty_id" VARCHAR(255),"language_code" VARCHAR(255),"currency" VARCHAR(255),"products" VARCHAR(255),"quantities" VARCHAR(255),"products_prices" VARCHAR(255),"ecommerce_action" VARCHAR(255),"order_payment_type" VARCHAR(255),"order_promo_code" VARCHAR(255),"page_name" VARCHAR(255),"page_type_environment" VARCHAR(255),"transaction_id" VARCHAR(255),"booking_id" VARCHAR(255),"geofence_latitude" VARCHAR(255),"geofence_longitude" VARCHAR(255),"geofence_id" VARCHAR(255),"geofence_name" VARCHAR(255),"poi_id" VARCHAR(255),"url" VARCHAR(255),"custom" VARCHAR(255),"fare_class" VARCHAR(255),"fare_type" VARCHAR(255),"flight_segments_departure_date_time" VARCHAR(255),"flight_segments_arrival_date_time" VARCHAR(255),"flight_segments" VARCHAR(255),"flight_segment_sku" VARCHAR(255),"flight_route" VARCHAR(255),"flight_numbers" VARCHAR(255),"flight_market" VARCHAR(255),"flight_type" VARCHAR(255),"origin_date" VARCHAR(255),"origin_date_time" TIMESTAMP,"return_date" VARCHAR(255),"return_date_time" TIMESTAMP,"return_flight_route" VARCHAR(255),"num_pax_adults" INTEGER,"num_pax_inf" INTEGER,"num_pax_children" INTEGER,"pax_type" VARCHAR(255),"total_passengers" INTEGER,"length_of_stay" INTEGER,"origin" VARCHAR(255),"selected_seats" VARCHAR(255),"room_type" VARCHAR(255),"rate_plan" VARCHAR(255),"checkin_date" TIMESTAMP,"checkout_date" TIMESTAMP,"num_nights" INTEGER,"num_guests" INTEGER,"num_guests_adult" INTEGER,"num_guests_children" INTEGER,"hotel_code" VARCHAR(255),"hotel_code_list" VARCHAR(255),"hotel_name" VARCHAR(255),"destination" VARCHAR(255),"profile_connect_id" VARCHAR(255))`
var hotel_stay_sql = `INSERT INTO my_domain_hotel_stay ("accp_object_id", "overall_confidence_score", "traveller_id", "stay_id", "booking_id", "currency_code", "currency_name", "currency_symbol", "first_name", "last_name", "email", "phone", "start_date", "hotel_code", "type", "description", "amount", "date", "last_updated", "last_updated_by", "profile_connect_id") VALUES ('', '0', '', '', '', '', '', '', '', '', '', '', '0001-01-01 00:00:00', '', '', '', '', '0001-01-01 00:00:00', '0001-01-01 00:00:00', '', '');`
var hotel_stay_update_sql = `UPDATE my_domain_hotel_stay SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "stay_id" = '', "booking_id" = '', "currency_code" = '', "currency_name" = '', "currency_symbol" = '', "first_name" = '', "last_name" = '', "email" = '', "phone" = '', "start_date" = '0001-01-01 00:00:00', "hotel_code" = '', "type" = '', "description" = '', "amount" = '', "date" = '0001-01-01 00:00:00', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var hotel_stay_delete_sql = `DELETE FROM my_domain_hotel_stay WHERE accp_object_id = '' AND profile_connect_id = ''`
var hotel_stay_create_table_sql = `CREATE TABLE my_domain_hotel_stay ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"stay_id" VARCHAR(255),"booking_id" VARCHAR(255),"currency_code" VARCHAR(255),"currency_name" VARCHAR(255),"currency_symbol" VARCHAR(255),"first_name" VARCHAR(255),"last_name" VARCHAR(255),"email" VARCHAR(255),"phone" VARCHAR(255),"start_date" TIMESTAMP,"hotel_code" VARCHAR(255),"type" VARCHAR(255),"description" VARCHAR(255),"amount" VARCHAR(255),"date" TIMESTAMP,"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"profile_connect_id" VARCHAR(255))`
var customer_service_interaction_sql = `INSERT INTO my_domain_customer_service_interaction ("accp_object_id", "overall_confidence_score", "traveller_id", "last_updated", "last_updated_by", "start_time", "end_time", "duration", "session_id", "channel", "interaction_type", "status", "language_code", "language_name", "conversation", "conversation_summary", "cart_id", "booking_id", "loyalty_id", "sentiment_score", "campaign_job_id", "campaign_strategy", "campaign_program", "campaign_product", "campaign_name", "category", "subject", "loyalty_program_name", "is_voice_otp", "profile_connect_id") VALUES ('', '0', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '0001-01-01 00:00:00', '0', '', '', '', '', '', '', '', '', '', '', '', '0', '', '', '', '', '', '', '', '', 'false', '');`
var customer_service_interaction_update_sql = `UPDATE my_domain_customer_service_interaction SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "start_time" = '0001-01-01 00:00:00', "end_time" = '0001-01-01 00:00:00', "duration" = '0', "session_id" = '', "channel" = '', "interaction_type" = '', "status" = '', "language_code" = '', "language_name" = '', "conversation" = '', "conversation_summary" = '', "cart_id" = '', "booking_id" = '', "loyalty_id" = '', "sentiment_score" = '0', "campaign_job_id" = '', "campaign_strategy" = '', "campaign_program" = '', "campaign_product" = '', "campaign_name" = '', "category" = '', "subject" = '', "loyalty_program_name" = '', "is_voice_otp" = 'false', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var customer_service_interaction_delete_sql = `DELETE FROM my_domain_customer_service_interaction WHERE accp_object_id = '' AND profile_connect_id = ''`
var customer_service_interaction_create_table_sql = `CREATE TABLE my_domain_customer_service_interaction ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"start_time" TIMESTAMP,"end_time" TIMESTAMP,"duration" INTEGER,"session_id" VARCHAR(255),"channel" VARCHAR(255),"interaction_type" VARCHAR(255),"status" VARCHAR(255),"language_code" VARCHAR(255),"language_name" VARCHAR(255),"conversation" VARCHAR(255),"conversation_summary" VARCHAR(255),"cart_id" VARCHAR(255),"booking_id" VARCHAR(255),"loyalty_id" VARCHAR(255),"sentiment_score" DECIMAL(10,2),"campaign_job_id" VARCHAR(255),"campaign_strategy" VARCHAR(255),"campaign_program" VARCHAR(255),"campaign_product" VARCHAR(255),"campaign_name" VARCHAR(255),"category" VARCHAR(255),"subject" VARCHAR(255),"loyalty_program_name" VARCHAR(255),"is_voice_otp" BOOLEAN,"profile_connect_id" VARCHAR(255))`
var loyalty_transaction_sql = `INSERT INTO my_domain_loyalty_tx ("accp_object_id", "overall_confidence_score", "traveller_id", "last_updated", "last_updated_by", "points_offset", "point_unit", "origin_points_offset", "qualifying_points_offset", "source", "category", "booking_date", "order_number", "product_id", "expire_in_days", "amount", "amount_type", "voucher_quantity", "corporate_reference_number", "promotions", "activity_day", "location", "to_loyalty_id", "from_loyalty_id", "organization_code", "event_name", "document_number", "corporate_id", "program_name", "profile_connect_id") VALUES ('', '0', '', '0001-01-01 00:00:00', '', '0', '', '0', '0', '', '', '0001-01-01 00:00:00', '', '', '0', '0', '', '0', '', '', '0001-01-01 00:00:00', '', '', '', '', '', '', '', '', '');`
var loyalty_transaction_update_sql = `UPDATE my_domain_loyalty_tx SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "points_offset" = '0', "point_unit" = '', "origin_points_offset" = '0', "qualifying_points_offset" = '0', "source" = '', "category" = '', "booking_date" = '0001-01-01 00:00:00', "order_number" = '', "product_id" = '', "expire_in_days" = '0', "amount" = '0', "amount_type" = '', "voucher_quantity" = '0', "corporate_reference_number" = '', "promotions" = '', "activity_day" = '0001-01-01 00:00:00', "location" = '', "to_loyalty_id" = '', "from_loyalty_id" = '', "organization_code" = '', "event_name" = '', "document_number" = '', "corporate_id" = '', "program_name" = '', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var loyalty_transaction_delete_sql = `DELETE FROM my_domain_loyalty_tx WHERE accp_object_id = '' AND profile_connect_id = ''`
var loyalty_transaction_create_table_sql = `CREATE TABLE my_domain_loyalty_tx ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"points_offset" DECIMAL(10,2),"point_unit" VARCHAR(255),"origin_points_offset" DECIMAL(10,2),"qualifying_points_offset" DECIMAL(10,2),"source" VARCHAR(255),"category" VARCHAR(255),"booking_date" TIMESTAMP,"order_number" VARCHAR(255),"product_id" VARCHAR(255),"expire_in_days" INTEGER,"amount" DECIMAL(10,2),"amount_type" VARCHAR(255),"voucher_quantity" INTEGER,"corporate_reference_number" VARCHAR(255),"promotions" VARCHAR(255),"activity_day" TIMESTAMP,"location" VARCHAR(255),"to_loyalty_id" VARCHAR(255),"from_loyalty_id" VARCHAR(255),"organization_code" VARCHAR(255),"event_name" VARCHAR(255),"document_number" VARCHAR(255),"corporate_id" VARCHAR(255),"program_name" VARCHAR(255),"profile_connect_id" VARCHAR(255))`
var email_history_sql = `INSERT INTO my_domain_email_history ("accp_object_id", "overall_confidence_score", "traveller_id", "address", "type", "last_updated", "last_updated_by", "is_verified", "profile_connect_id") VALUES ('', '0', '', '', '', '0001-01-01 00:00:00', '', 'false', '');`
var email_history_update_sql = `UPDATE my_domain_email_history SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "address" = '', "type" = '', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "is_verified" = 'false', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var email_history_delete_sql = `DELETE FROM my_domain_email_history WHERE accp_object_id = '' AND profile_connect_id = ''`
var email_history_create_table_sql = `CREATE TABLE my_domain_email_history ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"address" VARCHAR(255),"type" VARCHAR(255),"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"is_verified" BOOLEAN,"profile_connect_id" VARCHAR(255))`
var phone_history_sql = `INSERT INTO my_domain_phone_history ("accp_object_id", "overall_confidence_score", "traveller_id", "number", "country_code", "type", "last_updated", "last_updated_by", "is_verified", "profile_connect_id") VALUES ('', '0', '', '', '', '', '0001-01-01 00:00:00', '', 'false', '');`
var phone_history_update_sql = `UPDATE my_domain_phone_history SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "number" = '', "country_code" = '', "type" = '', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "is_verified" = 'false', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var phone_history_delete_sql = `DELETE FROM my_domain_phone_history WHERE accp_object_id = '' AND profile_connect_id = ''`
var phone_history_create_table_sql = `CREATE TABLE my_domain_phone_history ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"number" VARCHAR(255),"country_code" VARCHAR(255),"type" VARCHAR(255),"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"is_verified" BOOLEAN,"profile_connect_id" VARCHAR(255))`
var ancillary_service_sql = `INSERT INTO my_domain_ancillary_service ("accp_object_id", "overall_confidence_score", "traveller_id", "last_updated", "last_updated_by", "ancillary_type", "booking_id", "flight_number", "departure_date", "baggage_type", "pax_index", "quantity", "weight", "dimentions_length", "dimentions_width", "dimentions_height", "priority_bag_drop", "priority_bag_return", "lot_bag_insurance", "valuable_baggage_insurance", "hands_free_baggage", "seat_number", "seat_zone", "neighbor_free_seat", "upgrade_auction", "change_type", "other_ancilliary_type", "priority_service_type", "lounge_access", "price", "currency", "profile_connect_id") VALUES ('', '0', '', '0001-01-01 00:00:00', '', '', '', '', '', '', '0', '0', '0', '0', '0', '0', 'false', 'false', 'false', 'false', 'false', '', '', 'false', 'false', '', '', '', 'false', '0', '', '');`
var ancillary_service_update_sql = `UPDATE my_domain_ancillary_service SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "ancillary_type" = '', "booking_id" = '', "flight_number" = '', "departure_date" = '', "baggage_type" = '', "pax_index" = '0', "quantity" = '0', "weight" = '0', "dimentions_length" = '0', "dimentions_width" = '0', "dimentions_height" = '0', "priority_bag_drop" = 'false', "priority_bag_return" = 'false', "lot_bag_insurance" = 'false', "valuable_baggage_insurance" = 'false', "hands_free_baggage" = 'false', "seat_number" = '', "seat_zone" = '', "neighbor_free_seat" = 'false', "upgrade_auction" = 'false', "change_type" = '', "other_ancilliary_type" = '', "priority_service_type" = '', "lounge_access" = 'false', "price" = '0', "currency" = '', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var ancillary_service_delete_sql = `DELETE FROM my_domain_ancillary_service WHERE accp_object_id = '' AND profile_connect_id = ''`
var ancillary_service_create_table_sql = `CREATE TABLE my_domain_ancillary_service ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"ancillary_type" VARCHAR(255),"booking_id" VARCHAR(255),"flight_number" VARCHAR(255),"departure_date" VARCHAR(255),"baggage_type" VARCHAR(255),"pax_index" INTEGER,"quantity" INTEGER,"weight" DECIMAL(10,2),"dimentions_length" DECIMAL(10,2),"dimentions_width" DECIMAL(10,2),"dimentions_height" DECIMAL(10,2),"priority_bag_drop" BOOLEAN,"priority_bag_return" BOOLEAN,"lot_bag_insurance" BOOLEAN,"valuable_baggage_insurance" BOOLEAN,"hands_free_baggage" BOOLEAN,"seat_number" VARCHAR(255),"seat_zone" VARCHAR(255),"neighbor_free_seat" BOOLEAN,"upgrade_auction" BOOLEAN,"change_type" VARCHAR(255),"other_ancilliary_type" VARCHAR(255),"priority_service_type" VARCHAR(255),"lounge_access" BOOLEAN,"price" DECIMAL(10,2),"currency" VARCHAR(255),"profile_connect_id" VARCHAR(255))`
var alternate_profile_ids_sql = `INSERT INTO my_domain_ancillary_service ("accp_object_id", "overall_confidence_score", "traveller_id", "last_updated", "last_updated_by", "ancillary_type", "booking_id", "flight_number", "departure_date", "baggage_type", "pax_index", "quantity", "weight", "dimentions_length", "dimentions_width", "dimentions_height", "priority_bag_drop", "priority_bag_return", "lot_bag_insurance", "valuable_baggage_insurance", "hands_free_baggage", "seat_number", "seat_zone", "neighbor_free_seat", "upgrade_auction", "change_type", "other_ancilliary_type", "priority_service_type", "lounge_access", "price", "currency", "profile_connect_id") VALUES ('', '0', '', '0001-01-01 00:00:00', '', '', '', '', '', '', '0', '0', '0', '0', '0', '0', 'false', 'false', 'false', 'false', 'false', '', '', 'false', 'false', '', '', '', 'false', '0', '', '');`
var alternate_profile_ids_update_sql = `UPDATE my_domain_ancillary_service SET "accp_object_id" = '', "overall_confidence_score" = '0', "traveller_id" = '', "last_updated" = '0001-01-01 00:00:00', "last_updated_by" = '', "ancillary_type" = '', "booking_id" = '', "flight_number" = '', "departure_date" = '', "baggage_type" = '', "pax_index" = '0', "quantity" = '0', "weight" = '0', "dimentions_length" = '0', "dimentions_width" = '0', "dimentions_height" = '0', "priority_bag_drop" = 'false', "priority_bag_return" = 'false', "lot_bag_insurance" = 'false', "valuable_baggage_insurance" = 'false', "hands_free_baggage" = 'false', "seat_number" = '', "seat_zone" = '', "neighbor_free_seat" = 'false', "upgrade_auction" = 'false', "change_type" = '', "other_ancilliary_type" = '', "priority_service_type" = '', "lounge_access" = 'false', "price" = '0', "currency" = '', "profile_connect_id" = '' WHERE accp_object_id = '' AND profile_connect_id = ''`
var alternate_profile_ids_delete_sql = `DELETE FROM my_domain_ancillary_service WHERE accp_object_id = '' AND profile_connect_id = ''`
var alternate_profile_ids_create_table_sql = `CREATE TABLE my_domain_ancillary_service ("accp_object_id" VARCHAR(255),"overall_confidence_score" DECIMAL(10,2),"traveller_id" VARCHAR(255),"last_updated" TIMESTAMP,"last_updated_by" VARCHAR(255),"ancillary_type" VARCHAR(255),"booking_id" VARCHAR(255),"flight_number" VARCHAR(255),"departure_date" VARCHAR(255),"baggage_type" VARCHAR(255),"pax_index" INTEGER,"quantity" INTEGER,"weight" DECIMAL(10,2),"dimentions_length" DECIMAL(10,2),"dimentions_width" DECIMAL(10,2),"dimentions_height" DECIMAL(10,2),"priority_bag_drop" BOOLEAN,"priority_bag_return" BOOLEAN,"lot_bag_insurance" BOOLEAN,"valuable_baggage_insurance" BOOLEAN,"hands_free_baggage" BOOLEAN,"seat_number" VARCHAR(255),"seat_zone" VARCHAR(255),"neighbor_free_seat" BOOLEAN,"upgrade_auction" BOOLEAN,"change_type" VARCHAR(255),"other_ancilliary_type" VARCHAR(255),"priority_service_type" VARCHAR(255),"lounge_access" BOOLEAN,"price" DECIMAL(10,2),"currency" VARCHAR(255),"profile_connect_id" VARCHAR(255))`

var traveller_delete_table_sql = "DROP TABLE test_domain;"
var air_booking_delete_table_sql = "DROP TABLE test_domain_air_booking;"
var hotel_booking_delete_table_sql = "DROP TABLE test_domain_hotel_booking;"
var air_loyalty_delete_table_sql = "DROP TABLE test_domain_air_loyalty;"
var hotel_loyalty_delete_table_sql = "DROP TABLE test_domain_hotel_loyalty;"
var email_history_delete_table_sql = "DROP TABLE test_domain_email_history;"
var phone_history_delete_table_sql = "DROP TABLE test_domain_phone_history;"
var clickstream_delete_table_sql = "DROP TABLE test_domain_clickstream;"
var hotel_stay_delete_table_sql = "DROP TABLE test_domain_hotel_stay;"
var customer_service_interaction_delete_table_sql = "DROP TABLE test_domain_customer_service_interaction;"
var loyalty_transaction_delete_table_sql = "DROP TABLE test_domain_loyalty_tx;"
var ancillary_service_delete_table_sql = "DROP TABLE test_domain_ancillary_service;"
var alternate_profile_id_delete_table_sql = "DROP TABLE test_domain_alternate_profile_ids;"

var expected_delete = []string{
	traveller_delete_table_sql,
	air_booking_delete_table_sql,
	hotel_booking_delete_table_sql,
	air_loyalty_delete_table_sql,
	hotel_loyalty_delete_table_sql,
	email_history_delete_table_sql,
	phone_history_delete_table_sql,
	clickstream_delete_table_sql,
	hotel_stay_delete_table_sql,
	customer_service_interaction_delete_table_sql,
	loyalty_transaction_delete_table_sql,
	ancillary_service_delete_table_sql,
	alternate_profile_id_delete_table_sql,
}

type Config struct {
	ProfileObjectConfig []ObjectConfig
}

type ObjectConfig struct {
	Name                string
	Object              model.AccpObject
	ExpectedInsert      string
	ExpectedUpdate      string
	ExpectedDelete      string
	ExpectedCreateTable string
}

var config = Config{
	ProfileObjectConfig: []ObjectConfig{
		{
			Name:                "air_booking",
			Object:              model.AirBooking{},
			ExpectedInsert:      air_booking_sql,
			ExpectedUpdate:      air_booking_update_sql,
			ExpectedDelete:      air_booking_delete_sql,
			ExpectedCreateTable: air_booking_create_table_sql,
		},
		{
			Name:                "hotel_booking",
			Object:              model.HotelBooking{},
			ExpectedInsert:      hotel_booking_sql,
			ExpectedUpdate:      hotel_booking_update_sql,
			ExpectedDelete:      hotel_booking_delete_sql,
			ExpectedCreateTable: hotel_booking_create_table_sql,
		},
		{
			Name:                "air_loyalty",
			Object:              model.AirLoyalty{},
			ExpectedInsert:      air_loyalty_sql,
			ExpectedUpdate:      air_loyalty_update_sql,
			ExpectedDelete:      air_loyalty_delete_sql,
			ExpectedCreateTable: air_loyalty_create_table_sql,
		},
		{
			Name:                "hotel_loyalty",
			Object:              model.HotelLoyalty{},
			ExpectedInsert:      hotel_loyalty_sql,
			ExpectedUpdate:      hotel_loyalty_update_sql,
			ExpectedDelete:      hotel_loyalty_delete_sql,
			ExpectedCreateTable: hotel_loyalty_create_table_sql,
		},
		{
			Name:                "clickstream",
			Object:              model.Clickstream{},
			ExpectedInsert:      clickstream_sql,
			ExpectedUpdate:      clickstream_update_sql,
			ExpectedDelete:      clickstream_delete_sql,
			ExpectedCreateTable: clickstream_create_table_sql,
		},
		{
			Name:                "hotel_stay",
			Object:              model.HotelStay{},
			ExpectedInsert:      hotel_stay_sql,
			ExpectedUpdate:      hotel_stay_update_sql,
			ExpectedDelete:      hotel_stay_delete_sql,
			ExpectedCreateTable: hotel_stay_create_table_sql,
		},
		{
			Name:                "customer_service_interaction",
			Object:              model.CustomerServiceInteraction{},
			ExpectedInsert:      customer_service_interaction_sql,
			ExpectedUpdate:      customer_service_interaction_update_sql,
			ExpectedDelete:      customer_service_interaction_delete_sql,
			ExpectedCreateTable: customer_service_interaction_create_table_sql,
		},
		{
			Name:                "loyalty_transaction",
			Object:              model.LoyaltyTx{},
			ExpectedInsert:      loyalty_transaction_sql,
			ExpectedUpdate:      loyalty_transaction_update_sql,
			ExpectedDelete:      loyalty_transaction_delete_sql,
			ExpectedCreateTable: loyalty_transaction_create_table_sql,
		},
		{
			Name:                "email_history",
			Object:              model.EmailHistory{},
			ExpectedInsert:      email_history_sql,
			ExpectedUpdate:      email_history_update_sql,
			ExpectedDelete:      email_history_delete_sql,
			ExpectedCreateTable: email_history_create_table_sql,
		},
		{
			Name:                "phone_history",
			Object:              model.PhoneHistory{},
			ExpectedInsert:      phone_history_sql,
			ExpectedUpdate:      phone_history_update_sql,
			ExpectedDelete:      phone_history_delete_sql,
			ExpectedCreateTable: phone_history_create_table_sql,
		},
		{
			Name:                "ancillary_service",
			Object:              model.AncillaryService{},
			ExpectedInsert:      ancillary_service_sql,
			ExpectedUpdate:      ancillary_service_update_sql,
			ExpectedDelete:      ancillary_service_delete_sql,
			ExpectedCreateTable: ancillary_service_create_table_sql,
		},
		{
			Name:                "alternate_profile_ids",
			Object:              model.AncillaryService{},
			ExpectedInsert:      alternate_profile_ids_sql,
			ExpectedUpdate:      alternate_profile_ids_update_sql,
			ExpectedDelete:      alternate_profile_ids_delete_sql,
			ExpectedCreateTable: alternate_profile_ids_create_table_sql,
		},
	},
}

func TestTrySetTimestamp(t *testing.T) {
	errs := []string{}
	parsedTime := trySetTimestamp("test_field_name", "2023-11-13T18:47:33.174453Z", &errs)
	if parsedTime.IsZero() {
		t.Errorf("time stamp should be %v and not %v", "2023-11-13T18:47:33.174453Z", t)
	}
	if len(errs) > 0 {
		t.Errorf("error array should have no new parsing error. have : %v", errs)
	}
	log.Printf("Parsed time: %v", parsedTime)
}

func TestTravellerToRedshift(t *testing.T) {
	log.Printf("Testing TestTravellerToRedshift")
	statements := TravellerToRedshift("my_test_domain", model.Traveller{
		ConnectID: "acbdef",
		AirBookingRecords: []model.AirBooking{
			{LastUpdated: time.Now()},
		}}, true, "create")
	log.Printf("TestTravellerToRedshift results: %+v", statements)
	if len(statements) != 2 {
		t.Errorf("invalid number of statements %v", len(statements))
	}
}

func TestSqlInsertStatement(t *testing.T) {
	tableName := "my_domain"
	statements := []string{}
	expected := []string{}

	fmt.Println("******************************************************************")
	fmt.Println("*******************    COPY BELOW        *************************")
	fmt.Println("******************************************************************")
	fmt.Println("")

	statement := GenerateProfileInsertStatement(tableName, model.Traveller{ConnectID: "abcdefgh"})
	fmt.Printf("var traveller_sql = `%s`\n", statement)
	statements = append(statements, statement)
	expected = append(expected, traveller_sql)

	statement = GenerateProfileUpdateStatement(tableName, model.Traveller{ConnectID: "abcdefgh"})
	fmt.Printf("var traveller_update_sql = `%s`\n", statement)
	statements = append(statements, statement)
	expected = append(expected, traveller_update_sql)

	statement = GenerateProfileDeleteStatement(tableName, model.Traveller{ConnectID: "abcdefgh"})
	fmt.Printf("var traveller_delete_sql = `%s`\n", statement)
	statements = append(statements, statement)
	expected = append(expected, traveller_delete_sql)

	statement = GenerateProfileCreateTableStatement(tableName)
	fmt.Printf("var traveller_create_table_sql = `%s`\n", statement)
	statements = append(statements, statement)
	expected = append(expected, traveller_create_table_sql)

	for _, objConfig := range config.ProfileObjectConfig {
		statement = GenerateProfileObjectInsertStatement(tableName, "", objConfig.Object)
		fmt.Printf("var "+objConfig.Name+"_sql = `%s`\n", statement)
		statements = append(statements, statement)
		expected = append(expected, objConfig.ExpectedInsert)

		statement = GenerateProfileObjectUpdateStatement(tableName, "", objConfig.Object)
		fmt.Printf("var "+objConfig.Name+"_update_sql = `%s`\n", statement)
		statements = append(statements, statement)
		expected = append(expected, objConfig.ExpectedUpdate)

		statement = GenerateProfileObjectDeleteStatement(tableName, "", objConfig.Object)
		fmt.Printf("var "+objConfig.Name+"_delete_sql = `%s`\n", statement)
		statements = append(statements, statement)
		expected = append(expected, objConfig.ExpectedDelete)

		statement = GenerateProfileObjectCreateTableStatement(tableName, objConfig.Object)
		fmt.Printf("var "+objConfig.Name+"_create_table_sql = `%s`\n", statement)
		statements = append(statements, statement)
		expected = append(expected, objConfig.ExpectedCreateTable)

	}

	fmt.Println("")
	fmt.Println("******************************************************************")
	fmt.Println("*******************    COPY ABOVE        *************************")
	fmt.Println("******************************************************************")

	for i, s := range statements {
		if s != expected[i] {
			t.Errorf("invalid statement %v\n should be %v", s, expected[i])
		}
	}
}

func TestGenerateCreateTableStatements(t *testing.T) {
	statements := GenerateCreateTableStatements("test_domain")
	if len(statements) != 12 {
		t.Errorf("GenerateCreateTableStatements should generate %v statements and not %v", 12, len(statements))
	}
}

func TestGenerateDeleteTableStatements(t *testing.T) {
	statements := GenerateDeleteTableStatements("test_domain")
	for i, s := range statements {
		if s != expected_delete[i] {
			t.Errorf("invalid statement %v\n should be %v", s, expected_delete[i])
		}
	}
}

func TestFormatFieldValues(t *testing.T) {
	values := formatFieldValues([]string{"test with value to'escape 123"})
	if len(values) == 0 {
		t.Errorf("formatFieldValues shoudl resturn a non empty array")
	}
	log.Printf("values: %v", values)
	if values[0] != "'test with value to\\'escape 123'" {
		t.Errorf("Shoudl return escaped values")
	}

}
