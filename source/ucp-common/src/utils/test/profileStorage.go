// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	customerprofiles "tah/upt/source/storage"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/utils"
)

/////////////////////////////
// DEPRECATED: this and use the upt SDK instead
/////////////////////////////
/*
Test utilities for e2e testing Profile Storage (supports ACCP and LCS)

Historically, we were able to interact directly with Amazon Connect Customer Profiles to
set up and validate our test cases. With the move to LCS, we introduced an Aurora cluster and
cannot interact directly with the data source as easily.

Instead, these utility methods set up and search domains/profiles by interacting with the deployed backend.
*/

type EndToEndRequest struct {
	AccessToken string
	BaseUrl     string
	DomainName  string
}

func CreateDomainWithQueue(req EndToEndRequest, name string, idResolutionOn bool, kmsArn string, tags map[string]string, queueUrl string, matchS3Name string) error {
	path := "/api/ucp/admin"
	method := "POST"
	body := createDomainBody(name)

	_, err := request(req.BaseUrl, path, method, body, req.AccessToken, req.DomainName)
	if err != nil {
		return fmt.Errorf("error creating domain: %v", err)
	}
	return nil
}

func ListDomains(req EndToEndRequest) ([]model.Domain, error) {
	path := "/api/ucp/admin"
	method := "GET"
	body := "" // empty request body

	res, err := request(req.BaseUrl, path, method, body, req.AccessToken, req.DomainName)
	if err != nil {
		return nil, fmt.Errorf("error listing domains: %v", err)
	}
	return res.UCPConfig.Domains, nil
}

func DeleteDomain(req EndToEndRequest) error {
	path := "/api/ucp/admin/" + req.DomainName
	method := "DELETE"
	body := "" // empty request body

	_, err := request(req.BaseUrl, path, method, body, req.AccessToken, req.DomainName)
	if err != nil {
		return fmt.Errorf("error deleting domain: %v", err)
	}
	return nil
}

func SearchProfiles(req EndToEndRequest, key string, values []string) ([]traveler.Traveller, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("invalid number of values") // we currently only support single value search
	}
	path := "/api/ucp/profile"
	query := fmt.Sprintf("?%s=%s", key, values[0])
	method := "GET"
	body := "" // empty request body

	res, err := request(req.BaseUrl, path+query, method, body, req.AccessToken, req.DomainName)
	if err != nil {
		return nil, fmt.Errorf("error searching profiles: %v", err)
	}
	return utils.SafeDereference(res.Profiles), nil
}

func GetProfile(req EndToEndRequest, id string, objectTypeNames []string, pagination []customerprofiles.PaginationOptions) (traveler.Traveller, error) {
	path := "/api/ucp/profile/" + id
	method := "GET"
	body := "" // empty request body

	res, err := request(req.BaseUrl, path, method, body, req.AccessToken, req.DomainName)
	if err != nil {
		return traveler.Traveller{}, fmt.Errorf("error getting profile: %v", err)
	}
	profiles := utils.SafeDereference(res.Profiles)
	if len(profiles) != 1 {
		return traveler.Traveller{}, nil
	}
	return profiles[0], nil
}

func GetProfileId(req EndToEndRequest, profileKey, profileId string) (string, error) {
	profiles, err := SearchProfiles(req, profileKey, []string{profileId})
	if err != nil {
		return "", err
	}

	if len(profiles) != 1 {
		return "", fmt.Errorf("profile not found")
	}

	return profiles[0].ConnectID, nil
}

func GetSpecificProfileObject(req EndToEndRequest, objectId string, profileId string, objectTypeName string) (profilemodel.ProfileObject, error) {
	profile, err := GetProfile(req, profileId, admin.AccpRecordsNames(), []customerprofiles.PaginationOptions{})
	if err != nil {
		return profilemodel.ProfileObject{}, err
	}

	switch objectTypeName {
	case admin.ACCP_RECORD_AIR_LOYALTY:
		for _, obj := range profile.AirLoyaltyRecords {
			if obj.AccpObjectID == objectId {
				ai, err := buildAttributeInterface(obj)
				if err != nil {
					return profilemodel.ProfileObject{}, err
				}
				return profilemodel.ProfileObject{
					ID:                  obj.AccpObjectID,
					Type:                admin.ACCP_RECORD_AIR_LOYALTY,
					AttributesInterface: ai,
				}, nil
			}
		}
	case admin.ACCP_RECORD_ANCILLARY:
		for _, obj := range profile.AncillaryServiceRecords {
			if obj.AccpObjectID == objectId {
				ai, err := buildAttributeInterface(obj)
				if err != nil {
					return profilemodel.ProfileObject{}, err
				}
				return profilemodel.ProfileObject{
					ID:                  obj.AccpObjectID,
					Type:                admin.ACCP_RECORD_ANCILLARY,
					AttributesInterface: ai,
				}, nil
			}
		}
	case admin.ACCP_RECORD_CLICKSTREAM:
		for _, obj := range profile.ClickstreamRecords {
			if obj.AccpObjectID == objectId {
				ai, err := buildAttributeInterface(obj)
				if err != nil {
					return profilemodel.ProfileObject{}, err
				}
				return profilemodel.ProfileObject{
					ID:                  obj.AccpObjectID,
					Type:                admin.ACCP_RECORD_CLICKSTREAM,
					AttributesInterface: ai,
				}, nil
			}
		}
	case admin.ACCP_RECORD_HOTEL_BOOKING:
		for _, obj := range profile.HotelBookingRecords {
			if obj.AccpObjectID == objectId {
				ai, err := buildAttributeInterface(obj)
				if err != nil {
					return profilemodel.ProfileObject{}, err
				}
				return profilemodel.ProfileObject{
					ID:                  obj.AccpObjectID,
					Type:                admin.ACCP_RECORD_HOTEL_BOOKING,
					AttributesInterface: ai,
				}, nil
			}
		}
	case admin.ACCP_RECORD_HOTEL_LOYALTY:
		for _, obj := range profile.HotelLoyaltyRecords {
			if obj.AccpObjectID == objectId {
				ai, err := buildAttributeInterface(obj)
				if err != nil {
					return profilemodel.ProfileObject{}, err
				}
				return profilemodel.ProfileObject{
					ID:                  obj.AccpObjectID,
					Type:                admin.ACCP_RECORD_HOTEL_LOYALTY,
					AttributesInterface: ai,
				}, nil
			}
		}
	case admin.ACCP_RECORD_CSI:
		for _, obj := range profile.CustomerServiceInteractionRecords {
			if obj.AccpObjectID == objectId {
				ai, err := buildAttributeInterface(obj)
				if err != nil {
					return profilemodel.ProfileObject{}, err
				}
				return profilemodel.ProfileObject{
					ID:                  obj.AccpObjectID,
					Type:                admin.ACCP_RECORD_CSI,
					AttributesInterface: ai,
				}, nil
			}
		}
	case admin.ACCP_RECORD_ALTERNATE_PROFILE_ID:
		for _, obj := range profile.AlternateProfileIDs {
			if obj.AccpObjectID == objectId {
				ai, err := buildAttributeInterface(obj)
				if err != nil {
					return profilemodel.ProfileObject{}, err
				}
				return profilemodel.ProfileObject{
					ID:                  obj.AccpObjectID,
					Type:                admin.ACCP_RECORD_ALTERNATE_PROFILE_ID,
					AttributesInterface: ai,
				}, nil
			}
		}
	}

	return profilemodel.ProfileObject{}, fmt.Errorf("[GetSpecificProfileObject] object not found: %s", objectTypeName)
}

func request(base string, path string, method string, body string, token string, domainName string) (model.ResponseWrapper, error) {
	var request *http.Request
	var rw model.ResponseWrapper
	var err error
	url := base + path

	log.Printf("Rest API request %s %s with body: '%s'", method, url, body)

	if body != "" {
		request, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
		if err != nil {
			return rw, err
		}
	} else {
		request, err = http.NewRequest(method, url, nil)
		if err != nil {
			return rw, err
		}
	}

	// Set headers
	request.Header.Set("Authorization", "Bearer "+token)
	if domainName != "" {
		request.Header.Set("customer-profiles-domain", domainName)
	}
	if method == "POST" || method == "PUT" {
		request.Header.Set("Content-Type", "application/json")
	}

	log.Printf("Rest API request headers: %+v", request.Header)

	// Perform request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return rw, err
	}
	defer response.Body.Close()

	// Validate successful response
	if response.StatusCode != 200 {
		log.Printf("Rest API error code: %d", response.StatusCode)
		res, err := io.ReadAll(response.Body)
		log.Printf("Rest API Error Response  body: %s", string(res))
		if err != nil {
			return rw, err
		}
		return rw, fmt.Errorf("error with request %v %v", response.StatusCode, response.Status)
	}

	// Read response body
	res, err := io.ReadAll(response.Body)
	if err != nil {
		return rw, err
	}
	log.Printf("Rest API response: %s", string(res))

	// Update

	err = json.Unmarshal(res, &rw)
	if err != nil {
		return rw, err
	}

	return rw, nil
}

func createDomainBody(name string) string {
	return fmt.Sprintf(`
	{
		"domain": {
			"customerProfileDomain": "%s"
		}
	}
	`, name)
}

func buildAttributeInterface(obj interface{}) (map[string]interface{}, error) {
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)

	// Ensure input is a struct
	if objType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct")
	}

	result := make(map[string]interface{})

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i).Interface()
		fieldName := field.Name

		result[fieldName] = fieldValue
	}

	return result, nil
}
