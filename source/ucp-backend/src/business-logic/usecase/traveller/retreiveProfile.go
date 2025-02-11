// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	adminModel "tah/upt/source/ucp-common/src/model/admin"
	travModel "tah/upt/source/ucp-common/src/model/traveler"
	travSvc "tah/upt/source/ucp-common/src/services/traveller"
	utils "tah/upt/source/ucp-common/src/utils/utils"

	"github.com/aws/aws-lambda-go/events"
)

const (
	QUERY_PARAM_PAGES      string = "pages"
	QUERY_PARAM_PAGE_SIZES string = "pageSizes"
	QUERY_PARAM_OBJECTS    string = "objects"
)

type RetreiveProfile struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewRetreiveProfile() *RetreiveProfile {
	return &RetreiveProfile{name: "RetreiveProfile"}
}

func (u *RetreiveProfile) Name() string {
	return u.name
}
func (u *RetreiveProfile) Tx() core.Transaction {
	return *u.tx
}
func (u *RetreiveProfile) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *RetreiveProfile) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *RetreiveProfile) Registry() *registry.Registry {
	return u.reg
}

func (u *RetreiveProfile) AccessPermission() constant.AppPermission {
	return constant.SearchProfilePermission
}

func (u *RetreiveProfile) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	pages := utils.ParseIntArray(strings.Split(req.QueryStringParameters[QUERY_PARAM_PAGES], ","))
	pageSizes := utils.ParseIntArray(strings.Split(req.QueryStringParameters[QUERY_PARAM_PAGE_SIZES], ","))
	objects := strings.Split(req.QueryStringParameters[QUERY_PARAM_OBJECTS], ",")

	paginationOptions, err := buildPaginationOptions(pages, pageSizes, objects)
	if err != nil {
		return model.RequestWrapper{}, err
	}
	rw := model.RequestWrapper{
		ID:         req.PathParameters["id"],
		Pagination: paginationOptions,
	}
	return rw, nil
}

func buildPaginationOptions(pages, pageSizes []int, objects []string) ([]adminModel.PaginationOptions, error) {
	if len(pages) != len(pageSizes) || len(pages) != len(objects) || len(pageSizes) != len(objects) {
		return []adminModel.PaginationOptions{}, errors.New("invalid multi-object pagination options. number of query param mismatch between objects, pages and pageSizes")
	}
	pos := []adminModel.PaginationOptions{}
	for i, obj := range objects {
		if obj != "" {
			pos = append(pos, adminModel.PaginationOptions{
				Page:       pages[i],
				PageSize:   pageSizes[i],
				ObjectType: obj,
			})
		}
	}
	return pos, nil
}

func (u *RetreiveProfile) ValidateRequest(rq model.RequestWrapper) error {
	isValid := map[string]bool{}
	for _, obj := range constant.ACCP_RECORDS {
		isValid[obj.Name] = true
	}
	u.tx.Debug("[%v] Validating request", u.Name())
	for _, po := range rq.Pagination {
		if !isValid[po.ObjectType] {
			return fmt.Errorf("invalid object type %v in pagination. Supported values are %v", po.ObjectType, constant.AccpRecordsNames())
		}
	}
	return nil
}

func (u *RetreiveProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	traveller, err := travSvc.RetreiveTraveller(*u.tx, u.reg.Accp, req.ID, toLcsPagination(req.Pagination))
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	//we only paginate after because we want to control the sort order of items by timestamp
	metadata := BuildPaginationMetadata(traveller)
	//traveller = paginate(traveller, req.Pagination)
	matches, err := u.ProfileToMatches(req.ID, req)
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	return model.ResponseWrapper{Profiles: &[]travModel.Traveller{traveller}, Matches: &matches, PaginationMetadata: &metadata}, err
}

func toLcsPagination(pag []adminModel.PaginationOptions) []customerprofiles.PaginationOptions {
	lo := []customerprofiles.PaginationOptions{}
	for _, po := range pag {
		lo = append(lo, customerprofiles.PaginationOptions{
			Page:       po.Page,
			PageSize:   po.PageSize,
			ObjectType: po.ObjectType,
		})
	}
	return lo
}

func (u *RetreiveProfile) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func (u *RetreiveProfile) ProfileToMatches(id string, req model.RequestWrapper) ([]travModel.Match, error) {
	matches := []travModel.Match{}
	domainName := u.reg.Env["ACCP_DOMAIN_NAME"]

	matchPairs, err := travSvc.SearchMatches(*u.reg.MatchDB, domainName, id)
	if err != nil {
		u.tx.Error("Error searching for matches: %v", err)
		return matches, err
	}
	for _, m := range matchPairs {
		profile, err := u.reg.Accp.GetProfile(m.TargetProfileID, []string{}, []customerprofiles.PaginationOptions{})
		if err != nil {
			u.tx.Error("Profile not found for match %s. Deleting match", m.TargetProfileID)
			err = travSvc.DeleteMatch(*u.reg.MatchDB, domainName, id, m.TargetProfileID)
			if err != nil {
				u.tx.Error("Error cleaning up matches after merge", err)
			}
			continue
		}
		score, err := strconv.ParseFloat(m.Score, 64)
		if err != nil {
			u.tx.Error("Error converting score to float: ", err)
			return matches, err
		}
		matches = append(matches, travModel.Match{
			ConfidenceScore: score,
			ID:              profile.ProfileId,
			FirstName:       profile.FirstName,
			LastName:        profile.LastName,
			BirthDate:       profile.BirthDate,
			PhoneNumber:     profile.PhoneNumber,
			EmailAddress:    profile.EmailAddress,
		})
	}

	return matches, nil
}

func BuildPaginationMetadata(traveller travModel.Traveller) model.PaginationMetadata {
	metadata := model.PaginationMetadata{
		AirBookingRecords:                 model.Metadata{TotalRecords: int(traveller.NAirBookingRecords)},
		AirLoyaltyRecords:                 model.Metadata{TotalRecords: int(traveller.NAirLoyaltyRecords)},
		ClickstreamRecords:                model.Metadata{TotalRecords: int(traveller.NClickstreamRecords)},
		EmailHistoryRecords:               model.Metadata{TotalRecords: int(traveller.NEmailHistoryRecords)},
		HotelBookingRecords:               model.Metadata{TotalRecords: int(traveller.NHotelBookingRecords)},
		HotelLoyaltyRecords:               model.Metadata{TotalRecords: int(traveller.NHotelLoyaltyRecords)},
		HotelStayRecords:                  model.Metadata{TotalRecords: int(traveller.NHotelStayRecords)},
		PhoneHistoryRecords:               model.Metadata{TotalRecords: int(traveller.NPhoneHistoryRecords)},
		CustomerServiceInteractionRecords: model.Metadata{TotalRecords: int(traveller.NCustomerServiceInteractionRecords)},
		LoyaltyTxRecords:                  model.Metadata{TotalRecords: int(traveller.NLoyaltyTxRecords)},
		AncillaryServiceRecords:           model.Metadata{TotalRecords: int(traveller.NAncillaryServiceRecords)},
	}
	return metadata
}
