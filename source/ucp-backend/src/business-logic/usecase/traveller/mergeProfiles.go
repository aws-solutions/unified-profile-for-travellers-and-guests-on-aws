// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"encoding/json"
	"errors"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	async "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	travSvc "tah/upt/source/ucp-common/src/services/traveller"
	"time"

	"github.com/google/uuid"

	"github.com/aws/aws-lambda-go/events"
)

type MergeProfile struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewMergeProfile() *MergeProfile {
	return &MergeProfile{name: "MergeProfile"}
}

func (u *MergeProfile) Name() string {
	return u.name
}
func (u *MergeProfile) Tx() core.Transaction {
	return *u.tx
}
func (u *MergeProfile) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *MergeProfile) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *MergeProfile) Registry() *registry.Registry {
	return u.reg
}

func (u *MergeProfile) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *MergeProfile) AccessPermission() constant.AppPermission {
	return constant.MergeProfilePermission
}

func (u *MergeProfile) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("Validating request for MergeProfile")
	if u.reg.DataAccessPermission != constant.DATA_ACCESS_PERMISSION_ALL {
		return errors.New("only administrator can merge profiles")
	}
	if len(rq.MergeRq) == 0 {
		return errors.New("at least one profile pair must be provided for merge request")
	}
	if len(rq.MergeRq) > 50 {
		return errors.New("a maximum of 50 profile merges can be requested in a single request")
	}
	return nil
}

func (u *MergeProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("Merging %v profile pairs", len(req.MergeRq))
	var err0, err1 error
	rs := model.ResponseWrapper{}
	domain := u.reg.Env["ACCP_DOMAIN_NAME"]
	if domain == "" {
		return model.ResponseWrapper{}, errors.New("ACCP_DOMAIN_NAME not set")
	}

	ignoreRqs := []commonModel.MergeRq{}
	mergeRqs := []commonModel.MergeRq{}
	for _, mergeRq := range req.MergeRq {
		if !mergeRq.FalsePositive {
			mergeRqs = append(mergeRqs, mergeRq)
		} else {
			ignoreRqs = append(ignoreRqs, mergeRq)
		}
	}
	if len(ignoreRqs) > 0 {
		u.tx.Info("Marking %v matchPairs as false positive", len(ignoreRqs))
		err0 = travSvc.MarkFalsePositives(*u.tx, u.reg.MatchDB, domain, ignoreRqs)
	}
	if len(mergeRqs) > 0 {
		u.tx.Info("[MergeInProgress] request contains %v pairs", len(mergeRqs))
		u.tx.Debug("Filtering out overlaping matchPairs")
		mergeRqs = u.filterOutOverlapingRq(mergeRqs)
		u.tx.Info("There are %v pairs remaining to mark as MergeInProgress", len(mergeRqs))
		u.tx.Debug("Identifying addintional match pairs to be marked as merge in progress (same target profile id)")
		toMarkeAsMergeInProgress := u.matchPairsToBeMarkedAsMergeInProgress(mergeRqs, domain)
		u.tx.Info("Marking %v matchPairs as MergeInProgress", len(toMarkeAsMergeInProgress))
		err1 = travSvc.MarkMergeInProgress(*u.tx, u.reg.MatchDB, domain, toMarkeAsMergeInProgress)
		//profiles that are merged without having gone through the AI matching process will not have match pairs ain the DB
		//but should still be merged successfully
		if err1 != nil {
			u.tx.Warn("Warning: could not mark some matchpairs as MergInProgress. This may be an ad-hox merging. error: %+v", err1)
		}
		u.tx.Info("Starting merge for %v matchPairs", len(mergeRqs))
		rs, err1 = u.MergeAsync(domain, mergeRqs)
	}
	return rs, core.ErrorsToError([]error{err0, err1})
}

// this function identify all the matchpairs containing the same target profile id then the provided list
func (u *MergeProfile) matchPairsToBeMarkedAsMergeInProgress(mergeRqs []commonModel.MergeRq, domain string) []commonModel.MergeRq {
	u.tx.Debug("Identifying match pairs to be matched as merge in progress (all the ones sharing one of target profiles in the request)")
	finalRq := []commonModel.MergeRq{}
	duplicates := map[string]bool{}
	for _, rq := range mergeRqs {
		if !duplicates[rq.SourceProfileID+rq.TargetProfileID] {
			finalRq = append(finalRq, rq)
			duplicates[rq.SourceProfileID+rq.TargetProfileID] = true
			duplicates[rq.TargetProfileID+rq.SourceProfileID] = true
		}
		u.tx.Debug("Identifying all matches for target profile %v", rq.TargetProfileID)
		matches, err := travSvc.SearchMatches(*u.reg.MatchDB, domain, rq.TargetProfileID)
		if err != nil {
			u.tx.Error("Could not identify matches for target profile %v. error: %v", rq.TargetProfileID, err)
		}
		for _, match := range matches {
			if !duplicates[rq.TargetProfileID+match.TargetProfileID] && !duplicates[match.TargetProfileID+rq.TargetProfileID] {
				finalRq = append(finalRq, commonModel.MergeRq{
					SourceProfileID: rq.TargetProfileID,
					TargetProfileID: match.TargetProfileID,
				})
			} else {
				u.tx.Info("Match pair %v already exists for target profile", rq.TargetProfileID+match.TargetProfileID)
			}
		}
	}
	return finalRq
}

// this function identifies all the request that overlap with the target profile id and remove them from the list
func (u *MergeProfile) filterOutOverlapingRq(mergeRqs []commonModel.MergeRq) []commonModel.MergeRq {
	u.tx.Debug("Filtering out overlaping matchPairs")
	filetered := []commonModel.MergeRq{}
	isFiltered := map[string]bool{}
	for i, rqToMerge := range mergeRqs {
		if !isFiltered[rqToMerge.TargetProfileID+rqToMerge.SourceProfileID] {
			filetered = append(filetered, rqToMerge)
			for j := i + 1; j < len(mergeRqs); j++ {
				potentialOverlapRq := mergeRqs[j]
				if rqToMerge.TargetProfileID == potentialOverlapRq.SourceProfileID ||
					(rqToMerge.TargetProfileID == potentialOverlapRq.TargetProfileID && rqToMerge.SourceProfileID != potentialOverlapRq.SourceProfileID) {
					u.tx.Debug("Request %+v overlaps with request %+v. Removing %+v", rqToMerge, potentialOverlapRq, potentialOverlapRq)
					isFiltered[potentialOverlapRq.TargetProfileID+potentialOverlapRq.SourceProfileID] = true
				}
			}
		}
	}
	return filetered
}

func (u *MergeProfile) MergeAsync(domain string, rq []commonModel.MergeRq) (model.ResponseWrapper, error) {
	eventId := uuid.NewString()
	payload := u.buildAsyncPayload(domain, eventId, rq)
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		u.tx.Error("[%s] Error marshalling async payload: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	_, err = u.reg.AsyncLambda.InvokeAsync(payloadJson)
	if err != nil {
		u.tx.Error("[%s] Error invoking async lambda: %v", u.Name(), err)
		return model.ResponseWrapper{}, err
	}

	asyncEvent := async.AsyncEvent{
		EventID:     eventId,
		Usecase:     async.USECASE_MERGE_PROFILES,
		Status:      async.EVENT_STATUS_INVOKED,
		LastUpdated: time.Now(),
	}
	_, err = u.reg.ConfigDB.Save(asyncEvent)
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	res := model.ResponseWrapper{
		AsyncEvent: &asyncEvent,
	}
	return res, nil
}

func (u *MergeProfile) buildAsyncPayload(domain string, eventId string, rq []commonModel.MergeRq) async.AsyncInvokePayload {
	body := uc.MergeProfilesBody{
		Rq:     rq,
		Domain: domain,
	}
	payload := async.AsyncInvokePayload{
		EventID:       eventId,
		Usecase:       async.USECASE_MERGE_PROFILES,
		TransactionID: u.tx.TransactionID,
		Body:          body,
	}
	return payload
}

func (u *MergeProfile) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
