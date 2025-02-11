// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"strconv"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/utils"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestRetrieveProfile(t *testing.T) {

	segment1 := createAirBookingObject("1", "RDU", "BOS", "CONFIRMED")
	segment2 := createAirBookingObject("2", "BOS", "JFK", "CONFIRMED")
	segment3 := createAirBookingObject("3", "JFK", "RDU", "CONFIRMED")
	segment4 := createAirBookingObject("4", "RDU", "SEA", "CONFIRMED")

	retrieve(t, "*/*", "ucp_retreive_test-"+core.GenerateUniqueId(), 5,
		[]model.Traveller{
			{
				ConnectID:          "mock-traveller",
				FirstName:          "First",
				LastName:           "Last",
				NAirBookingRecords: 4,
				AirBookingRecords: []model.AirBooking{
					segment1,
					segment2,
					segment3,
					segment4,
				},
			},
		},
		[]profilemodel.ProfileObject{
			createProfileObject("mock-air-booking-1", "air_booking", segment1, 4),
			createProfileObject("mock-air-booking-2", "air_booking", segment2, 4),
			createProfileObject("mock-air-booking-3", "air_booking", segment3, 4),
			createProfileObject("mock-air-booking-4", "air_booking", segment4, 4),
		})

	retrievePaginatedData(t, "*/*", "ucp_page_retreive_test", 2,
		[]model.Traveller{
			{
				ConnectID:          "mock-traveller",
				FirstName:          "First",
				LastName:           "Last",
				NAirBookingRecords: 4,
				AirBookingRecords: []model.AirBooking{
					segment1,
					segment2,
					segment3,
					segment4,
				},
			},
		},
		[]profilemodel.ProfileObject{
			createProfileObject("mock-air-booking-3", "air_booking", segment3, 4),
			createProfileObject("mock-air-booking-4", "air_booking", segment4, 4),
		})
}

func testSetup(t *testing.T, permissionString string, tableName string, pageSize string, profileObjects []profilemodel.ProfileObject) (events.APIGatewayProxyRequest, *RetreiveProfile) {
	username := "test-username"
	domain := "test"
	groups := []cognito.Group{
		{
			Name:        "ucp-" + domain + "-group",
			Description: permissionString,
		},
		{
			Name:        "other",
			Description: "other",
		},
	}
	cognitoService := cognito.InitMock(nil, &groups)

	profile := profilemodel.Profile{
		ProfileId:      "mock-traveller",
		FirstName:      "First",
		LastName:       "Last",
		ProfileObjects: profileObjects,
	}
	accpService := customerprofiles.InitMock(nil, nil, &profile, nil, &[]customerprofiles.ObjectMapping{
		{Name: "air_booking"},
		{Name: "air_loyalty"},
		{Name: "hotel_loyalty"},
		{Name: "hotel_booking"},
		{Name: "hotel_stay_revenue_items"},
		{Name: "customer_service_interaction"},
		{Name: "ancillary_service"},
		{Name: "alternate_profile_id"},
		{Name: "guest_profile"},
		{Name: "pax_profile"},
		{Name: "email_history"},
		{Name: "phone_history"},
		{Name: "loyalty_transaction"},
		{Name: "clickstream"},
	})

	dbConfig, err := db.InitWithNewTable(tableName, "domain_sourceProfileId", "match_targetProfileId", "", "")
	if err != nil {
		t.Errorf("Error initializing db: %s", err)
	}
	t.Cleanup(func() { dbConfig.DeleteTable(tableName) })
	dbConfig.WaitForTableCreation()

	services := registry.ServiceHandlers{
		Cognito: cognitoService,
		Accp:    accpService,
		MatchDB: &dbConfig,
	}
	reg := registry.NewRegistry("test-region", core.LogLevelDebug, services)
	req := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": "mock-traveller",
		},
		QueryStringParameters: map[string]string{
			QUERY_PARAM_PAGES:      "0",
			QUERY_PARAM_PAGE_SIZES: pageSize,
			QUERY_PARAM_OBJECTS:    constants.BIZ_OBJECT_AIR_BOOKING,
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"claims": map[string]interface{}{
					"username": username,
				},
			},
		},
	}

	retrieveProfile := NewRetreiveProfile()
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	retrieveProfile.SetTx(&tx)
	retrieveProfile.SetRegistry(&reg)
	retrieveProfile.reg.AddEnv("ACCP_DOMAIN_NAME", domain)
	retrieveProfile.reg.AddEnv("DYNAMO_TABLE_MATCH", "")

	actualName := retrieveProfile.Name()
	expectedName := "RetreiveProfile"
	if actualName != expectedName {
		t.Errorf("[TestRetrieveProfile] Error: expected request name %v but received %v", expectedName, actualName)
	}

	return req, retrieveProfile
}

func retrieve(t *testing.T, permissionString string, tableName string, pageSize int, expectedProfiles []model.Traveller, profileObjects []profilemodel.ProfileObject) {
	req, retrieveProfile := testSetup(t, permissionString, tableName, strconv.Itoa(pageSize), profileObjects)

	reqWrapper, err := retrieveProfile.CreateRequest(req)
	if err != nil {
		t.Errorf("[TestRetrieveProfile] Error creating request: %v", err)
	}

	err = retrieveProfile.ValidateRequest(reqWrapper)
	if err != nil {
		t.Errorf("[TestRetrieveProfile] Invalid request: %v", err)
	}

	res, err := retrieveProfile.Run(reqWrapper)
	if err != nil {
		t.Errorf("[TestRetrieveProfile] Error running request: %v", err)
	}
	profiles := utils.SafeDereference(res.Profiles)
	if len(profiles) != len(expectedProfiles) ||
		profiles[0].FirstName != expectedProfiles[0].FirstName ||
		profiles[0].LastName != expectedProfiles[0].LastName ||
		profiles[0].ConnectID != expectedProfiles[0].ConnectID ||
		len(profiles[0].AirBookingRecords) != len(expectedProfiles[0].AirBookingRecords) ||
		res.PaginationMetadata.AirBookingRecords.TotalRecords != len(expectedProfiles[0].AirBookingRecords) {
		t.Errorf("[TestRetrieveProfile] Error: expected profiles %v but received %v", expectedProfiles, profiles)
	}
}

func retrievePaginatedData(t *testing.T, permissionString string, tableName string, pageSize int, expectedProfiles []model.Traveller, profileObjects []profilemodel.ProfileObject) {
	req, retrieveProfile := testSetup(t, permissionString, tableName, strconv.Itoa(pageSize), profileObjects)

	reqWrapper, err := retrieveProfile.CreateRequest(req)
	if err != nil {
		t.Errorf("[TestRetrieveProfile] Error creating request: %v", err)
	}

	err = retrieveProfile.ValidateRequest(reqWrapper)
	if err != nil {
		t.Errorf("[TestRetrieveProfile] Invalid request: %v", err)
	}

	res, err := retrieveProfile.Run(reqWrapper)
	if err != nil {
		t.Errorf("[TestRetrieveProfile] Error running request: %v", err)
	}

	if len((utils.SafeDereference(res.Profiles))[0].AirBookingRecords) != pageSize ||
		res.PaginationMetadata.AirBookingRecords.TotalRecords != len(expectedProfiles[0].AirBookingRecords) {
		t.Errorf("[TestRetrieveProfile] Error: expected profiles %v but received %v", expectedProfiles, res.Profiles)
	}
}

func structToMap(obj model.AirBooking) map[string]string {
	return map[string]string{
		"segment_id": obj.SegmentID,
		"from":       obj.From,
		"to":         obj.To,
		"status":     obj.Status,
	}
}

func createProfileObject(id string, typ string, segment model.AirBooking, totalCount int64) profilemodel.ProfileObject {
	return profilemodel.ProfileObject{
		ID:         id,
		Type:       typ,
		Attributes: structToMap(segment),
		TotalCount: totalCount,
	}
}

func createAirBookingObject(segmentNum string, from string, to string, status string) model.AirBooking {
	return model.AirBooking{
		SegmentID: "segment-" + segmentNum,
		From:      from,
		To:        to,
		Status:    status,
	}
}
