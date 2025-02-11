// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"

	"github.com/aws/aws-lambda-go/events"
)

type ListPrivacySearchesData struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

// Static interface check
var _ registry.Usecase = &ListPrivacySearchesData{}

func NewListPrivacySearches() *ListPrivacySearchesData {
	return &ListPrivacySearchesData{
		name: "ListPrivacySearches",
	}
}

func (u *ListPrivacySearchesData) CreateRequest(event events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return model.RequestWrapper{
		ListPrivacySearches: model.ListPrivacySearchesRq{
			DomainName: u.reg.Env[constants.ACCP_DOMAIN_NAME_ENV_VAR],
		},
	}, nil
}

func (u *ListPrivacySearchesData) CreateResponse(response model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, response)
}

// Name implements registry.Usecase.
func (u *ListPrivacySearchesData) Name() string {
	return u.name
}

// Registry implements registry.Usecase.
func (u *ListPrivacySearchesData) Registry() *registry.Registry {
	return u.reg
}

// Run implements registry.Usecase.
func (u *ListPrivacySearchesData) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	searchResults := []uc.PrivacySearchResult{}
	u.tx.Info("Running privacy search or domain '%s'", req.ListPrivacySearches.DomainName)
	err := u.reg.PrivacyDB.FindByPk(req.ListPrivacySearches.DomainName, &searchResults)
	if err != nil {
		u.tx.Error("[%s] Error unmarshalling privacy searches: %s", u.name, err.Error())
		return model.ResponseWrapper{}, err
	}
	u.tx.Debug("Found search %d results for privacy search on domain '%s'", len(searchResults), req.ListPrivacySearches.DomainName)
	//	searchResults should have records that match: PK: domainName SK: connectId
	//	as well as records that match: PK: domainName SK: connectId|locationType
	mainResultsMap := make(map[string]*uc.PrivacySearchResult)
	locationResultsSlice := []uc.PrivacySearchResult{}
	for _, searchResult := range searchResults {
		if strings.Contains(searchResult.ConnectId, "|") {
			locationResultsSlice = append(locationResultsSlice, searchResult)
		} else {
			//	Only return a subset of the fields to reduce unnecessary traffic to the UI
			mainResultsMap[searchResult.ConnectId] = &uc.PrivacySearchResult{
				DomainName:        searchResult.DomainName,
				ConnectId:         searchResult.ConnectId,
				Status:            searchResult.Status,
				SearchDate:        searchResult.SearchDate,
				TotalResultsFound: searchResult.TotalResultsFound,
				ErrorMessage:      searchResult.ErrorMessage,
			}
		}
	}

	for _, searchResult := range locationResultsSlice {
		//	Overwrite the mainResultsMap's status with the locationType's status if it exists
		//	Status should overwrite with the following priority:
		//	(Highest) Failed status
		//	Running status
		//	Invoked status
		//	(Lowest) Success status
		if mainResult, ok := mainResultsMap[strings.Split(searchResult.ConnectId, "|")[0]]; ok {
			mainResult.Status = getHigherPriorityStatus(mainResult.Status, searchResult.Status)
			mainResult.ErrorMessage = strings.Join([]string{mainResult.ErrorMessage, searchResult.ErrorMessage}, "\n")
		}
	}

	returnSlice := []uc.PrivacySearchResult{}
	for _, result := range mainResultsMap {
		returnSlice = append(returnSlice, *result)
	}
	u.tx.Info("Returning  %d results for privacy search on domain '%s'", len(returnSlice), req.ListPrivacySearches.DomainName)
	return model.ResponseWrapper{
		PrivacySearchResults: &returnSlice,
	}, nil
}

// SetRegistry implements registry.Usecase.
func (u *ListPrivacySearchesData) SetRegistry(r *registry.Registry) {
	u.reg = r
}

// SetTx implements registry.Usecase.
func (u *ListPrivacySearchesData) SetTx(tx *core.Transaction) {
	u.tx = tx
}

// Tx implements registry.Usecase.
func (u *ListPrivacySearchesData) Tx() core.Transaction {
	return *u.tx
}

func (u *ListPrivacySearchesData) AccessPermission() admin.AppPermission {
	return admin.ListPrivacySearchPermission
}

// ValidateRequest implements registry.Usecase.
func (u *ListPrivacySearchesData) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%s] Validating Request", u.name)

	return nil
}

func getHigherPriorityStatus(currentStatus uc.PrivacySearchStatus, otherStatus uc.PrivacySearchStatus) uc.PrivacySearchStatus {
	currentStatusPriority := uc.PrivacySearchStatusPriority[currentStatus]
	otherStatusPriority := uc.PrivacySearchStatusPriority[otherStatus]
	if currentStatusPriority > otherStatusPriority {
		return currentStatus
	}
	return otherStatus
}
