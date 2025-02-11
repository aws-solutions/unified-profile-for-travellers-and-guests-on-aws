// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package e2epipeline

import (
	"fmt"
	"tah/upt/schemas/generators/real-time/src/example"
	"tah/upt/schemas/src/tah-common/air"
	"tah/upt/schemas/src/tah-common/common"
	csi "tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/lodging"
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func TestPutProfileCache(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000
	// Constants
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}
	kinesisCfg := kinesis.Init(infraConfig.CpExportStream, envConfig.Region, "", "", core.LogLevelDebug)
	kinesisCfg.InitConsumer("TRIM_HORIZON")

	domainName, uptHandler, dynamoCache, cpCache, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	initialTravellerId := "testTravellerOne"
	hotelRecords, err := generators.GenerateSimpleHotelStayRecords(initialTravellerId, domainName, 5)
	if err != nil {
		t.Fatalf("Error generating hotel record: %v", err)
	}
	guestRecord, err := generators.GenerateSimpleGuestProfileRecords(initialTravellerId, domainName, 1)
	if err != nil {
		t.Fatalf("Error generating guest record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, hotelRecords)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestRecord)

	err = util.WaitForExpectedCondition(func() bool {
		profile, exists := util.GetProfileWithTravellerId(t, uptHandler, domainName, initialTravellerId, []string{"hotel_stay"})
		if !exists {
			return exists
		}
		if len(profile.HotelStayRecords) != 5 {
			t.Logf("Expected 5 hotel stay records, got %d", len(profile.HotelStayRecords))
			return false
		}
		_, profileCp, exists := util.GetProfileFromAccp(t, cpCache, domainName, profile.ConnectID, []string{"hotel_stay_revenue_items"})
		if !exists {
			return exists
		}
		dynamoProfile, dynamoObjects, exists := util.GetProfileRecordsFromDynamo(t, dynamoCache, domainName, profile.ConnectID, []string{"hotel_stay_revenue_items"})
		if !exists {
			return exists
		}
		if profile.FirstName == "" {
			t.Logf("Profile first name should not be empty")
			return false
		}
		if profile.FirstName != profileCp.FirstName {
			t.Logf("Expected %s, got %s", profile.FirstName, profileCp.FirstName)
			return false
		}
		if profile.FirstName != dynamoProfile.FirstName {
			t.Logf("Expected %s, got %s", profile.FirstName, dynamoProfile.FirstName)
			return false
		}
		dynamoHotelObjects, ok := dynamoObjects["hotel_stay_revenue_items"]
		if !ok {
			t.Logf("Expected hotel_stay_revenue_items to be present in dynamo objects")
			return false
		}
		if len(dynamoHotelObjects) != 5 {
			t.Logf("Expected 5 dynamo objects, got %d", len(dynamoObjects))
			return false
		}
		// Search Using GSI
		indexResults := []customerprofileslcs.DynamoTravelerIndex{}
		err = dynamoCache.FindByGsi(initialTravellerId, customerprofileslcs.DDB_GSI_NAME, customerprofileslcs.DDB_GSI_PK, &indexResults, db.FindByGsiOptions{Limit: 1})
		if err != nil {
			t.Logf("GSI did not return a value")
			return false
		}
		if len(indexResults) != 1 {
			t.Logf("GSI search should return 1 value, found %v", len(indexResults))
			return false
		}
		dynamoGsiProfile, dynamoGsiObjects, exists := util.GetProfileRecordsFromDynamo(t, dynamoCache, domainName, indexResults[0].ConnectId, []string{"hotel_stay_revenue_items"})
		if !exists {
			return exists
		}
		if profile.FirstName != dynamoGsiProfile.FirstName {
			t.Logf("Expected %s, got %s", profile.FirstName, dynamoGsiProfile.FirstName)
			return false
		}
		dynamoHotelGsiObjects, ok := dynamoGsiObjects["hotel_stay_revenue_items"]
		if !ok {
			t.Logf("Expected hotel_stay_revenue_items to be present in dynamo objects")
			return false
		}
		if len(dynamoHotelGsiObjects) != 5 {
			t.Logf("Expected 5 dynamo objects, got %d", len(dynamoHotelGsiObjects))
			return false
		}
		return true
	}, 12, 5*time.Second)
	if err != nil {
		t.Errorf("Error waiting for condition: %v", err)
	}

	records, err := kinesisCfg.FetchRecords(10)
	if err != nil {
		t.Errorf("Error fetching records from kinesis: %v", err)
	}
	if len(records) == 0 {
		t.Errorf("Expected records in the stream, got %d", len(records))
	}
}

func TestPutHotelStay(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, _, cpConfig, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	commonAccpId := "testAccpId"
	initialTravellerId := "testTravellerOne"
	recordOne, err := generators.GenerateSpecificHotelStayRecords(initialTravellerId, commonAccpId, domainName)
	if err != nil {
		t.Fatalf("Error generating hotel record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, recordOne)

	err = util.WaitForExpectedCondition(func() bool {
		connectIdOne, err := uptHandler.GetProfileId(domainName, initialTravellerId)
		if err != nil {
			t.Logf("Error getting connect id: %v", err)
			return false
		}
		cpId1, err := cpConfig.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, connectIdOne)
		if err != nil {
			t.Logf("Error getting accp id for %v: %v", connectIdOne, err)
			return false
		}
		if cpId1 == "" || connectIdOne == "" {
			t.Logf("Expected connect ids to be present, got %s and %s", connectIdOne, cpId1)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}

	mappings, err := cpConfig.GetMappings()
	if err != nil {
		t.Fatalf("Error getting mappings: %v", err)
	}
	for _, mapping := range mappings {
		if len(mapping.Fields) < 10 {
			t.Fatalf("Expected at least 10 fields in mapping, got %d", len(mapping.Fields))
		}
	}
}

func TestSameEmailPhoneDifferentType(t *testing.T) {
	t.Parallel()
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, db, _, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	//	Commong func that can be used across all test cases
	expectedFunc := func(t *testing.T, id string) bool {
		connectIdOne, err := uptHandler.GetProfileId(domainName, id)
		if err != nil {
			t.Logf("[%s] Error getting connect id: %v", t.Name(), err)
			return false
		}
		_, profileObjects, success := util.GetProfileRecordsFromDynamo(t, db, domainName, connectIdOne, []string{"email_history", "phone_history"})
		if !success {
			t.Logf("[%s] Error getting profile from dynamo: %v", t.Name(), err)
			return false
		}
		if len(profileObjects["email_history"]) != 2 {
			t.Logf("[%s] Expected 2 email history records, got %d", t.Name(), len(profileObjects["email_history"]))
			return false
		}
		if len(profileObjects["phone_history"]) != 3 {
			t.Logf("[%s] Expected 3 phone history records, got %d", t.Name(), len(profileObjects["phone_history"]))
			return false
		}
		for _, emailHistory := range profileObjects["email_history"] {
			if emailHistory.ID != fmt.Sprintf("%s%s", emailHistory.Attributes["type"], emailHistory.Attributes["address"]) {
				t.Logf("[%s] expected: %s%s; got: %s", t.Name(), emailHistory.Attributes["type"], emailHistory.Attributes["address"], emailHistory.ID)
				return false
			}
		}
		for _, phoneHistory := range profileObjects["phone_history"] {
			if phoneHistory.ID != fmt.Sprintf("%s%s", phoneHistory.Attributes["type"], phoneHistory.Attributes["number"]) {
				t.Logf("[%s] expected: %s%s; got: %s", t.Name(), phoneHistory.Attributes["type"], phoneHistory.Attributes["number"], phoneHistory.ID)
				return false
			}
		}
		return true
	}

	//

	//	Define Test Cases
	tests := []struct {
		name         string
		objectType   string
		skipTest     bool
		generateFunc func() (obj interface{}, id string)
	}{
		{
			name:       "Air Booking Object Type",
			objectType: "air_booking",
			skipTest:   false,
			generateFunc: func() (obj interface{}, id string) {
				airBooking := generators.GenerateRandomSimpleAirBooking(domainName, []generators.FieldOverrides{
					{
						Weight: 100000,
						Fields: []generators.FieldOverride{
							{
								Name: "Emails",
								Value: []common.Email{
									{
										Address: "bob@example.com",
										Type:    common.EMAIL_TYPE_PERSONAL,
									},
									{
										Address: "bob@example.com",
										Type:    common.EMAIL_TYPE_BUSINESS,
									},
								},
							},
							{
								Name: "Phones",
								Value: []common.Phone{
									{
										Number: "8679305",
										Type:   common.PHONE_TYPE_BUSINESS,
									},
									{
										Number: "8679305",
										Type:   common.PHONE_TYPE_HOME,
									},
									{
										Number: "8679305",
										Type:   common.PHONE_TYPE_MOBILE,
									},
								},
							},
						},
					},
				}...)
				return airBooking, airBooking.PassengerInfo.Passengers[0].ID
			},
		},
		{
			name:       "Hotel Booking Object Type",
			objectType: "hotel_booking",
			skipTest:   false,
			generateFunc: func() (obj interface{}, id string) {
				hotelBooking := generators.GenerateHotelBooking(true, true, []generators.FieldOverrides{
					{
						Weight: 100000,
						Fields: []generators.FieldOverride{
							{
								Name: "Emails",
								Value: []common.Email{
									{
										Address: "larry@example.com",
										Type:    common.EMAIL_TYPE_PERSONAL,
									},
									{
										Address: "larry@example.com",
										Type:    common.EMAIL_TYPE_BUSINESS,
									},
								},
							},
							{
								Name: "Phones",
								Value: []common.Phone{
									{
										Number: "18002255669",
										Type:   common.PHONE_TYPE_BUSINESS,
									},
									{
										Number: "18002255669",
										Type:   common.PHONE_TYPE_HOME,
									},
									{
										Number: "18002255669",
										Type:   common.PHONE_TYPE_MOBILE,
									},
								},
							},
						},
					},
				}...)
				return hotelBooking, hotelBooking.Holder.ID
			},
		},
		{
			name:       "Guest Profile Object Type",
			objectType: "guest_profile",
			skipTest:   false,
			generateFunc: func() (obj interface{}, id string) {
				guest := example.CreateGuestProfileExample(1, true)
				guest.Emails = []common.Email{
					{
						Address: "margaret@example.com",
						Type:    common.EMAIL_TYPE_PERSONAL,
					},
					{
						Address: "margaret@example.com",
						Type:    common.EMAIL_TYPE_BUSINESS,
					},
				}
				guest.Phones = []common.Phone{
					{
						Number: "18332541111",
						Type:   common.PHONE_TYPE_BUSINESS,
					},
					{
						Number: "18332541111",
						Type:   common.PHONE_TYPE_HOME,
					},
					{
						Number: "18332541111",
						Type:   common.PHONE_TYPE_MOBILE,
					},
				}

				return guest, guest.ID
			},
		},
		{
			name:       "Passenger Profile Object Type",
			objectType: "pax_profile",
			skipTest:   false,
			generateFunc: func() (obj interface{}, id string) {
				passenger := example.CreatePassengerProfileExample(1)
				passenger.Emails = []common.Email{
					{
						Address: "sally@example.com",
						Type:    common.EMAIL_TYPE_PERSONAL,
					},
					{
						Address: "sally@example.com",
						Type:    common.EMAIL_TYPE_BUSINESS,
					},
				}
				passenger.Phones = []common.Phone{
					{
						Number: "18882258875",
						Type:   common.PHONE_TYPE_BUSINESS,
					},
					{
						Number: "18882258875",
						Type:   common.PHONE_TYPE_HOME,
					},
					{
						Number: "18882258875",
						Type:   common.PHONE_TYPE_MOBILE,
					},
				}
				return passenger, passenger.ID
			},
		},
	}

	//	Iterate through test cases
	for _, test := range tests {
		test := test // quirk we need to deal with until we use Go 1.22 https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/loopclosure
		if test.skipTest {
			t.Logf("Skipping test %s", test.name)
			continue
		}
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			obj, id := test.generateFunc()
			rec, err := generators.GetKinesisRecord(domainName, test.objectType, obj)
			if err != nil {
				t.Fatalf("[%s] error generating %s record: %v", t.Name(), test.objectType, err)
			}
			util.SendKinesisBatch(t, realTimeStream, 1000, []kinesis.Record{rec})

			err = util.WaitForExpectedCondition(func() bool {
				return expectedFunc(t, id)
			}, 6, 5*time.Second)
			if err != nil {
				t.Fatalf("[%s] error waiting for condition: %v", t.Name(), err)
			}
		})
	}
}

// this test validates the "extended data" feature
// for all object types, we have an extended data field, which allows
// us to add whatever data we want so long as it is in JSON form. If you manually
// map that data (which is required), that data should be mapped and exist on the profile.
// This test creates objects with extended data, manually changes the mappings, then
// validates the addition data exists at profile level.
func TestExtendedData(t *testing.T) {
	t.Parallel()
	kinesisRate := 1000
	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	//extended data objects
	extendDataCsi := map[string]interface{}{
		"language": "English",
	}

	extendDataClickstream := map[string]interface{}{
		"sessionId": "sess_123456",
		"events": []map[string]interface{}{
			{
				"timestamp": "2024-03-15T14:22:31Z",
				"eventType": "page_view",
				"path":      "/hotels/search",
			},
		},
		"deviceType": "mobile",
	}

	extendDataAirBooking := map[string]interface{}{
		"pnr": "ABC123",
		"segments": []map[string]interface{}{
			{
				"from":     "SFO",
				"to":       "NRT",
				"flightNo": "NH007",
			},
		},
		"status": "confirmed",
	}

	extendDataPhoneHistory := map[string]interface{}{
		"call": map[string]interface{}{
			"duration": 360,
			"type":     "support",
			"topic":    "booking_modification",
		},
	}

	extendDataEmailHistory := map[string]interface{}{
		"messages": map[string]interface{}{
			"messageId": "msg_987654",
			"subject":   "Booking Confirmation",
			"category":  "transactional",
		},
	}

	extendDataAncillaryService := map[string]interface{}{
		"service":    "airport_transfer",
		"bookingRef": "TRF789",
		"status":     "confirmed",
		"amount":     75.50,
		"currency":   "USD",
	}

	extendDataAirLoyalty := map[string]interface{}{
		"airline":    "Star Alliance",
		"number":     "SA987654",
		"tier":       "Gold",
		"miles":      85000,
		"validUntil": "2024-12-31",
	}
	extendDataGuestProfile := map[string]interface{}{
		"personalInfo": map[string]interface{}{
			"demographics": map[string]interface{}{
				"age":        35,
				"occupation": "Engineer",
			},
		},
	}

	domainName, _, _, cpConfig, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	//Modifying our mappings to map extended data to attribute fields

	err = PutModifiedMapping(cpConfig, "customer_service_interaction", "language", "csi")
	if err != nil {
		t.Fatalf("Error putting modified mapping: %v", err)
	}
	err = PutModifiedMapping(cpConfig, "email_history", "messages.subject", "email_history")
	if err != nil {
		t.Fatalf("Error putting modified mapping: %v", err)
	}
	err = PutModifiedMapping(cpConfig, "phone_history", "call.duration", "phone_history")
	if err != nil {
		t.Fatalf("Error putting modified mapping: %v", err)
	}
	err = PutModifiedMapping(cpConfig, "guest_profile", "personalInfo.demographics.age", "guest_profile")
	if err != nil {
		t.Fatalf("Error putting modified mapping: %v", err)
	}
	err = PutModifiedMapping(cpConfig, "clickstream", "deviceType", "clickstream")
	if err != nil {
		t.Fatalf("Error putting modified mapping: %v", err)
	}
	err = PutModifiedMapping(cpConfig, "ancillary_service", "bookingRef", "ancillary_service")
	if err != nil {
		t.Fatalf("Error putting modified mapping: %v", err)
	}
	err = PutModifiedMapping(cpConfig, "air_loyalty", "tier", "air_loyalty")
	if err != nil {
		t.Fatalf("Error putting modified mapping: %v", err)
	}

	//sending the records for all object types
	allRecords := []kinesis.Record{}
	travellerId := "testTravellerOne"
	csiRecord, err := generators.GenerateSpecificCsiRecord(travellerId, domainName, func(gp *csi.CustomerServiceInteraction) {
		gp.ExtendedData = extendDataCsi
	})
	if err != nil {
		t.Fatalf("Error generating csi record: %v", err)
	}

	guestProfileRecord, err := generators.GenerateSpecificGuestProfileRecord(travellerId, domainName, func(gp *lodging.GuestProfile) {
		gp.Emails = []common.Email{
			{
				Address:      "bob@example.com",
				Type:         common.EMAIL_TYPE_PERSONAL,
				ExtendedData: extendDataEmailHistory,
			},
		}
		gp.Phones = []common.Phone{
			{
				Number:       "8679305",
				Type:         common.PHONE_TYPE_BUSINESS,
				ExtendedData: extendDataPhoneHistory,
			},
		}
		gp.ExtendedData = extendDataGuestProfile
	},
	)
	if err != nil {
		t.Fatalf("Error generating guest profile record: %v", err)
	}

	paxBooking := generators.GenerateAirBooking(1)
	paxBooking.AncillaryServices = air.AncillaryServices{
		AncillaryServiceBaggage: []air.AncillaryServiceBaggage{
			{
				ID:           "sample_ancillary_id",
				BaggageType:  "Carry On",
				ExtendedData: extendDataAncillaryService,
			},
		},
	}
	paxBooking.ExtendedData = extendDataAirBooking
	if len(paxBooking.PassengerInfo.Passengers) == 0 {
		t.Fatalf("Error generating air booking record: no passengers")
	}
	paxPassenger := &paxBooking.PassengerInfo.Passengers[0]
	firstName := "testFirstName"
	middleName := "testMiddleName"
	lastName := "testLastName"
	paxPassenger.FirstName = firstName
	paxPassenger.MiddleName = middleName
	paxPassenger.LastName = lastName
	paxPassenger.LoyaltyPrograms = []air.LoyaltyProgram{
		{
			ID:           "sample_object_id",
			ProgramName:  "testProgram",
			ExtendedData: extendDataAirLoyalty,
		},
	}
	paxBookingRecord, err := generators.GetKinesisRecord(domainName, "air_booking", paxBooking)
	if err != nil {
		t.Fatalf("Error generating pax booking record: %v", err)
	}

	clickRecord, err := generators.BuildSpecificClickstreamRecord(travellerId, domainName, func(click *common.ClickEvent) {
		click.ExtendedData = extendDataClickstream
	})
	if err != nil {
		t.Fatalf("Error generating click record: %v", err)
	}

	allRecords = append(allRecords, csiRecord...)
	allRecords = append(allRecords, guestProfileRecord...)
	allRecords = append(allRecords, paxBookingRecord)
	allRecords = append(allRecords, clickRecord)

	util.SendKinesisBatch(t, realTimeStream, kinesisRate, allRecords)

	//we now check that when we grab the data from CP, the mappings worked
	err = util.WaitForExpectedCondition(func() bool {
		prof, boo := util.GetProfileFromAccpWithTravellerId(t, cpConfig, domainName, travellerId)
		if !boo {
			t.Logf("Error getting profile from accp: %v", err)
			return false
		}
		fullProf, err := cpConfig.GetProfile(prof.ProfileId,
			[]string{"email_history",
				"phone_history",
				"guest_profile",
				"clickstream",
				"customer_service_interaction",
			},
			[]customerprofiles.PaginationOptions{})
		if err != nil {
			t.Logf("Error getting profile from accp: %v", err)
			return false
		}
		if len(fullProf.ProfileObjects) != 5 {
			t.Logf("Expected 5 profile objects, got %d", len(fullProf.ProfileObjects))
			return false
		}
		csi, ok := fullProf.Attributes["csi"]
		if !ok {
			t.Logf("Expected csi to be present")
			return false
		}
		if csi != "English" {
			t.Logf("Expected csi to be English, got %s", csi)
			return false
		}
		phone_history, ok := fullProf.Attributes["phone_history"]
		if !ok {
			t.Logf("Expected phone history to be present")
			return false
		}
		if phone_history != "360" {
			t.Logf("Expected phone history to be 360, got %s", csi)
			return false
		}
		email_history, ok := fullProf.Attributes["email_history"]
		if !ok {
			t.Logf("Expected email history to be present")
			return false
		}
		if email_history != "Booking Confirmation" {
			t.Logf("Expected email to be Booking Confirmation, got %s", csi)
			return false
		}
		guest_data, ok := fullProf.Attributes["guest_profile"]
		if !ok {
			t.Logf("Expected guest profile to be present")
			return false
		}
		if guest_data != "35" {
			t.Logf("Expected guest data to be 35, got %s", guest_data)
			return false
		}
		click, ok := fullProf.Attributes["clickstream"]
		if !ok {
			t.Logf("Expected clickstream to be present")
			return false
		}
		if click != "mobile" {
			t.Logf("Expected clickstream to be mobile, got %s", guest_data)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}

	err = util.WaitForExpectedCondition(func() bool {
		airProf, err := cpConfig.SearchProfiles("_fullName", []string{firstName + " " + middleName + " " + lastName})
		if err != nil {
			t.Logf("Error getting profile from accp: %v", err)
			return false
		}
		if len(airProf) == 0 {
			t.Logf("Expected air profile to be present")
			return false
		}
		fullAirProf, err := cpConfig.GetProfile(airProf[0].ProfileId,
			[]string{"pax_profile",
				"air_loyalty",
				"air_booking",
				"ancillary_service"},
			[]customerprofiles.PaginationOptions{})
		if err != nil {
			t.Logf("Error getting profile from accp: %v", err)
			return false
		}
		ancillary_service, ok := fullAirProf.Attributes["ancillary_service"]
		if !ok {
			t.Logf("Expected ancillary service to be present")
			return false
		}
		if ancillary_service != "TRF789" {
			t.Logf("Expected ancillary service to be TRF789, got %s", ancillary_service)
			return false
		}
		air_loyalty, ok := fullAirProf.Attributes["air_loyalty"]
		if !ok {
			t.Logf("Expected air loyalty to be present")
			return false
		}
		if air_loyalty != "Gold" {
			t.Logf("Expected air loyalty to be Gold, got %s", ancillary_service)
			return false
		}
		return true
	}, 6, 5*time.Second)
	if err != nil {
		t.Fatalf("Error waiting for condition: %v", err)
	}
}

func PutModifiedMapping(cpConfig *customerprofiles.CustomerProfileConfig, objectTypeName, key, attributeValue string) error {
	err := cpConfig.CreateMapping(objectTypeName, "test", []customerprofiles.FieldMapping{
		{
			Type:    "STRING",
			Source:  "_source.timestamp",
			Target:  "timestamp",
			Indexes: []string{},
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.traveller_id",
			Target:      "traveller_id",
			Indexes:     []string{},
			Searcheable: true,
			KeyOnly:     true,
		},
		{
			Type:        "STRING",
			Source:      "_source." + customerprofiles.CONNECT_ID_ATTRIBUTE,
			Target:      "_profile.Attributes." + customerprofiles.CONNECT_ID_ATTRIBUTE,
			Indexes:     []string{"PROFILE"},
			Searcheable: true,
		},
		{
			Type:    "STRING",
			Source:  "_source." + customerprofiles.UNIQUE_OBJECT_ATTRIBUTE,
			Target:  customerprofiles.UNIQUE_OBJECT_ATTRIBUTE,
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.extended_data." + key,
			Target:      "_profile.Attributes." + attributeValue,
			Indexes:     []string{},
			Searcheable: true,
		},
	})
	return err
}
