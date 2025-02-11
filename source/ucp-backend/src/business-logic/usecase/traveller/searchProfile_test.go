// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"errors"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	utils "tah/upt/source/ucp-common/src/utils/utils"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestSearch(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	testProfile := profilemodel.Profile{
		ProfileId:    "1",
		FirstName:    "Hello",
		LastName:     "World",
		PhoneNumber:  "1234567890",
		EmailAddress: "something",
		ShippingAddress: profilemodel.Address{
			State: "MA",
		},
	}
	searchResults := []profilemodel.Profile{testProfile}
	searchCriteria := []customerprofiles.BaseCriteria{
		customerprofiles.SearchCriterion{Column: SEARCH_KEY_LAST_NAME, Operator: customerprofiles.EQ, Value: "World"},
	}
	mock := customerprofiles.InitMockV2()
	mock.On("AdvancedProfileSearch", searchCriteria).Return(searchResults, nil)
	mock.On("SearchProfiles", "profile_id", []string{"1"}).Return(searchResults, nil)
	reg := registry.Registry{Accp: mock}
	req := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"travellerId": "1",
		},
	}

	search := NewSearchProfile()
	search.SetTx(&tx)
	search.SetRegistry(&reg)
	search.reg.SetAppAccessPermission(admin.SearchProfilePermission)

	if search.Name() != "SearchProfile" {
		t.Errorf("Expected SearchProfile, got %s", search.Name())
	}
	if search.Tx().TransactionID == "" {
		t.Errorf("Expected transaction ID, got empty string")
	}
	if search.Registry().Accp == nil {
		t.Errorf("Expected Accp object, got nil")
	}
	reqWrapper, err := search.CreateRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	err = search.ValidateRequest(reqWrapper)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	resWrapper, err := search.Run(reqWrapper)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	res, err := search.CreateResponse(resWrapper)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", res.StatusCode)
	}

	//	Test search by lastName
	req = events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"lastName": "World",
		},
	}
	reqWrapper, err = search.CreateRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	err = search.ValidateRequest(reqWrapper)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	resWrapper, err = search.Run(reqWrapper)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	profiles := utils.SafeDereference(resWrapper.Profiles)
	if len(profiles) == 0 {
		t.Errorf("Expected results; got none")
	}
	foundProfile := profiles[0]
	if foundProfile.LastName != "World" {
		t.Errorf("Expected %s; got %s", "World", foundProfile.LastName)
	}
	res, err = search.CreateResponse(resWrapper)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", res.StatusCode)
	}

	//	Test by non-existent travelerID and lastName
	mock.On("SearchProfiles", "profile_id", []string{"4"}).Return([]profilemodel.Profile{}, nil)
	req = events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"travellerId": "4",
			"lastName":    "World",
		},
	}
	reqWrapper, err = search.CreateRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	err = search.ValidateRequest(reqWrapper)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	resWrapper, err = search.Run(reqWrapper)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	profiles = utils.SafeDereference(resWrapper.Profiles)
	if len(profiles) != 0 {
		t.Errorf("Expected no results, got %d", len(profiles))
	}
	res, err = search.CreateResponse(resWrapper)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", res.StatusCode)
	}

}

type ValidationTests struct {
	Rq    model.RequestWrapper
	Error error
}

func TestSearchValidation(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	reg := registry.Registry{}
	search := NewSearchProfile()
	search.SetTx(&tx)
	search.SetRegistry(&reg)
	search.reg.SetAppAccessPermission(admin.SearchProfilePermission)
	TestSearchValidation := []ValidationTests{
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{TravellerID: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{LastName: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{Phone: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{Email: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{Address1: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{Address2: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{City: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{State: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{Province: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{Country: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{PostalCode: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{AddressType: "1=1"}}, Error: errors.New("search field value contains invalid characters")},
		{Rq: model.RequestWrapper{SearchRq: model.SearchRq{Matches: "blabla"}}, Error: errors.New("search field value contains invalid characters")},
	}
	for _, test := range TestSearchValidation {
		err := search.ValidateRequest(test.Rq)
		if errors.Is(test.Error, err) {
			t.Fatalf("Validation for request %+v expected '%v', got '%v'", test.Rq, test.Error, err)
		}
	}

}
