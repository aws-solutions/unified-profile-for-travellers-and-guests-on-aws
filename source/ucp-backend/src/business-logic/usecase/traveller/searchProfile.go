// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"errors"
	lcsmodel "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	adapter "tah/upt/source/ucp-common/src/adapter"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	travSvc "tah/upt/source/ucp-common/src/services/traveller"
	utils "tah/upt/source/ucp-common/src/utils/utils"

	"github.com/aws/aws-lambda-go/events"
)

var SEARCH_KEY_LAST_NAME = "LastName"
var SEARCH_KEY_FIRST_NAME = "FirstName"
var SEARCH_KEY_PERSONAL_EMAIL = "PersonalEmailAddress"
var SEARCH_KEY_BUSINESS_EMAIL = "BusinessEmailAddress"
var SEARCH_KEY_PHONE = "PhoneNumber"
var SEARCH_KEY_PROFILE_ID = "profile_id"

type SearchProfile struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewSearchProfile() *SearchProfile {
	return &SearchProfile{name: "SearchProfile"}
}

func (u *SearchProfile) Name() string {
	return u.name
}
func (u *SearchProfile) Tx() core.Transaction {
	return *u.tx
}
func (u *SearchProfile) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *SearchProfile) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *SearchProfile) Registry() *registry.Registry {
	return u.reg
}

func (u *SearchProfile) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	rw, err := registry.CreateRequest(u, req)
	rw.SearchRq = model.SearchRq{}
	rw.SearchRq.TravellerID = req.QueryStringParameters["travellerId"]
	rw.SearchRq.LastName = req.QueryStringParameters["lastName"]
	rw.SearchRq.Phone = req.QueryStringParameters["phone"]
	rw.SearchRq.Email = req.QueryStringParameters["email"]
	rw.SearchRq.Address1 = req.QueryStringParameters["addressAddress1"]
	rw.SearchRq.Address2 = req.QueryStringParameters["addressAddress2"]
	rw.SearchRq.City = req.QueryStringParameters["addressCity"]
	rw.SearchRq.State = req.QueryStringParameters["addressState"]
	rw.SearchRq.Province = req.QueryStringParameters["addressProvince"]
	rw.SearchRq.Country = req.QueryStringParameters["addressCountry"]
	rw.SearchRq.PostalCode = req.QueryStringParameters["addressPostalCode"]
	rw.SearchRq.AddressType = req.QueryStringParameters["addressType"]
	rw.SearchRq.Matches = req.QueryStringParameters["matches"]
	u.tx.Debug("Search Request: %v", rw)
	return rw, err
}

func (u *SearchProfile) AccessPermission() constant.AppPermission {
	return constant.SearchProfilePermission
}

func (u *SearchProfile) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	if err := validateString(rq.SearchRq.TravellerID); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.LastName); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.Phone); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.Email); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.Address1); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.Address2); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.City); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.State); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.Province); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.Country); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.PostalCode); err != nil {
		return err
	}
	if err := validateString(rq.SearchRq.AddressType); err != nil {
		return err
	}
	if rq.SearchRq.Matches != "" && (rq.SearchRq.Matches != "true" && rq.SearchRq.Matches != "false") {
		return errors.New("invalid request param. matches should be either empty, true or false")
	}
	return nil
}

func validateString(str string) error {
	invalidChars := []string{"\\", "/", "\b", "\f", "\n", "\r", "\t", "{", "}", "<", ">", "^", "="}
	if str == "" {
		return nil
	}
	if len(str) > 255 {
		return errors.New("invalid request param. should be smaller than 255 char")
	}
	if utils.ValidateString(str, invalidChars) != nil {
		return errors.New("search field value contains invalid characters")
	}
	return nil
}

func (u *SearchProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	//	Multiple fields should narrow down search (email AND lastName)
	//	Email and phone should search across all valid types (personalEmail or businessEmail)
	var profiles []profilemodel.Profile
	matchPairs := []commonModel.MatchPair{}
	totalMatchPairs := int64(0)
	addressType := req.SearchRq.AddressType
	var err error
	po := commonModel.PaginationOptions{}
	searchCriteria := []lcsmodel.BaseCriteria{}
	if len(req.Pagination) > 0 {
		po = req.Pagination[0]
		u.tx.Debug("Pagination details: %+v", po)
	}
	if req.SearchRq.TravellerID != "" {
		u.tx.Debug("Searching by travellerId")
		profiles, err = u.reg.Accp.SearchProfiles(SEARCH_KEY_PROFILE_ID, []string{req.SearchRq.TravellerID})
		if err != nil {
			u.tx.Error("Error searching by travellerId: %v", err)
			return model.ResponseWrapper{}, err
		}
	} else {
		if req.SearchRq.LastName != "" {
			u.tx.Debug("Adding search by lastName criterion")
			if len(searchCriteria) > 0 {
				searchCriteria = append(searchCriteria, lcsmodel.SearchOperator{Operator: lcsmodel.AND})
			}
			searchCriteria = append(searchCriteria, lcsmodel.SearchCriterion{Column: SEARCH_KEY_LAST_NAME, Operator: lcsmodel.EQ, Value: req.SearchRq.LastName})
			// profiles, err = u.reg.Accp.SearchProfiles(SEARCH_KEY_LAST_NAME, []string{req.SearchRq.LastName})
		}
		if req.SearchRq.Phone != "" {
			u.tx.Debug("Adding search by phone criterion")
			if len(searchCriteria) > 0 {
				searchCriteria = append(searchCriteria, lcsmodel.SearchOperator{Operator: lcsmodel.AND})
			}
			phoneCriteria := []lcsmodel.BaseCriteria{
				lcsmodel.SearchGroup{
					Criteria: []lcsmodel.BaseCriteria{
						lcsmodel.SearchCriterion{Column: SEARCH_KEY_PHONE, Operator: lcsmodel.EQ, Value: req.SearchRq.Phone},
						// lcsmodel.SearchOperator{Operator: lcsmodel.OR},
						// lcsmodel.SearchCriterion{Column: SEARCH_KEY_PHONE, Operator: lcsmodel.EQ, Value: req.SearchRq.Phone},
						// lcsmodel.SearchOperator{Operator: lcsmodel.OR},
						// lcsmodel.SearchCriterion{Column: SEARCH_KEY_PHONE, Operator: lcsmodel.EQ, Value: req.SearchRq.Phone},
					},
				},
			}
			searchCriteria = append(searchCriteria, phoneCriteria...)
			// profiles, err = u.reg.Accp.SearchProfiles(SEARCH_KEY_PHONE, []string{req.SearchRq.Phone})
		}
		if req.SearchRq.Email != "" {
			u.tx.Debug("Adding search by email criterion")
			if len(searchCriteria) > 0 {
				searchCriteria = append(searchCriteria, lcsmodel.SearchOperator{Operator: lcsmodel.AND})
			}
			emailCriteria := []lcsmodel.BaseCriteria{
				lcsmodel.SearchGroup{
					Criteria: []lcsmodel.BaseCriteria{
						lcsmodel.SearchCriterion{Column: SEARCH_KEY_PERSONAL_EMAIL, Operator: lcsmodel.EQ, Value: req.SearchRq.Email},
						lcsmodel.SearchOperator{Operator: lcsmodel.OR},
						lcsmodel.SearchCriterion{Column: SEARCH_KEY_BUSINESS_EMAIL, Operator: lcsmodel.EQ, Value: req.SearchRq.Email},
					},
				},
			}
			searchCriteria = append(searchCriteria, emailCriteria...)
			// profiles, err = u.reg.Accp.SearchProfiles(SEARCH_KEY_EMAIL, []string{req.SearchRq.Email})
		}
		if req.SearchRq.AddressType != "NoAddress" {
			addressCriteria := u.AddAddressCriteria(addressType, req)
			if len(addressCriteria) > 0 {
				u.tx.Debug("Adding search by address criteria")
				if len(searchCriteria) > 0 {
					searchCriteria = append(searchCriteria, lcsmodel.SearchOperator{Operator: lcsmodel.AND})
				}
				searchCriteria = append(searchCriteria, addressCriteria...)
			}
			// profiles, err = u.AddressSearch(addressType, req, profiles, err)
		}

		profiles, err = u.reg.Accp.AdvancedProfileSearch(searchCriteria)
		if err != nil {
			u.tx.Error("Error searching for profiles: %v", err)
			return model.ResponseWrapper{}, err
		}
	}
	if req.SearchRq.Matches == "true" {
		u.tx.Debug("Searching all matches")
		matchPairs, err = travSvc.FindAllMatches(*u.tx, *u.reg.MatchDB, po)
		if err != nil {
			u.tx.Error("Error searching for matches: %v", err)
			return model.ResponseWrapper{}, err
		}

		count, err := u.reg.MatchDB.GetItemCount()
		if err != nil {
			u.tx.Error("error getting match count: %v", err)
			return model.ResponseWrapper{}, errors.New("error getting match count")
		}
		estimateMatchCount := int64(po.Page * po.PageSize)
		totalMatchPairs = max(estimateMatchCount, count)
	}
	if err != nil {
		u.tx.Error("Error during search: %v", err)
		return model.ResponseWrapper{}, err
	}
	u.tx.Info("Profile search found %d results", len(profiles))
	travellers := adapter.ProfilesToTravellers(*u.tx, profiles)
	return model.ResponseWrapper{Profiles: &travellers, MatchPairs: &matchPairs, TotalMatchPairs: &totalMatchPairs}, nil
}

func (u *SearchProfile) AddAddressCriteria(addressType string, req model.RequestWrapper) (searchCriteria []lcsmodel.BaseCriteria) {
	if req.SearchRq.Address1 != "" {
		u.tx.Debug("Adding search by address1 criteria")
		searchCriteria = append(searchCriteria, lcsmodel.SearchCriterion{Column: addressType + ".Address1", Operator: lcsmodel.EQ, Value: req.SearchRq.Address1})
		// profiles, err = u.reg.Accp.SearchProfiles(addressType+".Address1", []string{req.SearchRq.Address1})
	}
	if req.SearchRq.Address2 != "" {
		u.tx.Debug("Adding search by address2 criteria")
		if len(searchCriteria) > 0 {
			searchCriteria = append(searchCriteria, lcsmodel.SearchOperator{Operator: lcsmodel.AND})
		}
		searchCriteria = append(searchCriteria, lcsmodel.SearchCriterion{Column: addressType + ".Address2", Operator: lcsmodel.EQ, Value: req.SearchRq.Address2})
		// profiles, err = u.reg.Accp.SearchProfiles(addressType+".Address2", []string{req.SearchRq.Address2})
	}
	if req.SearchRq.City != "" {
		u.tx.Debug("Adding search by city criteria")
		if len(searchCriteria) > 0 {
			searchCriteria = append(searchCriteria, lcsmodel.SearchOperator{Operator: lcsmodel.AND})
		}
		searchCriteria = append(searchCriteria, lcsmodel.SearchCriterion{Column: addressType + ".City", Operator: lcsmodel.EQ, Value: req.SearchRq.City})
		// profiles, err = u.reg.Accp.SearchProfiles(addressType+".City", []string{req.SearchRq.City})
	}
	if req.SearchRq.State != "" {
		u.tx.Debug("Adding search by state criteria")
		if len(searchCriteria) > 0 {
			searchCriteria = append(searchCriteria, lcsmodel.SearchOperator{Operator: lcsmodel.AND})
		}
		searchCriteria = append(searchCriteria, lcsmodel.SearchCriterion{Column: addressType + ".State", Operator: lcsmodel.EQ, Value: req.SearchRq.State})
		// profiles, err = u.reg.Accp.SearchProfiles(addressType+".State", []string{req.SearchRq.State})
	}
	if req.SearchRq.Province != "" {
		u.tx.Debug("Adding search by province criteria")
		if len(searchCriteria) > 0 {
			searchCriteria = append(searchCriteria, lcsmodel.SearchOperator{Operator: lcsmodel.AND})
		}
		searchCriteria = append(searchCriteria, lcsmodel.SearchCriterion{Column: addressType + ".Province", Operator: lcsmodel.EQ, Value: req.SearchRq.Province})
		// profiles, err = u.reg.Accp.SearchProfiles(addressType+".Province", []string{req.SearchRq.Province})
	}
	if req.SearchRq.PostalCode != "" {
		u.tx.Debug("Adding search by postal code criteria")
		if len(searchCriteria) > 0 {
			searchCriteria = append(searchCriteria, lcsmodel.SearchOperator{Operator: lcsmodel.AND})
		}
		searchCriteria = append(searchCriteria, lcsmodel.SearchCriterion{Column: addressType + ".PostalCode", Operator: lcsmodel.EQ, Value: req.SearchRq.PostalCode})
		// profiles, err = u.reg.Accp.SearchProfiles(addressType+".PostalCode", []string{req.SearchRq.PostalCode})
	}

	return searchCriteria
}

func (u *SearchProfile) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
