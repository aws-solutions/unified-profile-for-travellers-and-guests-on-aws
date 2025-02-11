// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"encoding/json"
	"fmt"
	"sync"
	customerprofileslcs "tah/upt/source/storage"
	cw "tah/upt/source/tah-core/cloudwatch"
	"tah/upt/source/tah-core/core"
	ddb "tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	"tah/upt/source/ucp-common/src/utils/utils"
	t "time"
)

type PurgeProfileDataUsecase struct {
	name       string
	domainName string
}

// Static Interface Check
var _ Usecase = (*PurgeProfileDataUsecase)(nil)

func InitPurgeProfileData() *PurgeProfileDataUsecase {
	return &PurgeProfileDataUsecase{
		name: "PurgeProfileData",
	}
}

// Name implements Usecase.
func (u *PurgeProfileDataUsecase) Name() string {
	return u.name
}

// Execute implements Usecase.
func (u *PurgeProfileDataUsecase) Execute(payload async.AsyncInvokePayload, services async.Services, tx core.Transaction) error {
	//	Retrieve payload details
	payloadJson, err := json.Marshal(payload.Body)
	if err != nil {
		tx.Error("[%s] Error marshalling payload: %v", u.name, err)
		return err
	}

	//	Convert json to PurgeProfileDataBody
	data := uc.PurgeProfileDataBody{}
	err = json.Unmarshal(payloadJson, &data)
	if err != nil {
		tx.Error("[%s] Error unmarshalling payload body: %v", u.name, err)
		return err
	}
	tx.Debug("[%s] Payload body: %+v", u.name, data)

	tx.Debug("[%s] Setting domain to ACCP config to %s", u.name, data.DomainName)
	services.AccpConfig.SetDomain(data.DomainName)
	u.domainName = data.DomainName

	//	Find records from PrivacySearch table
	var ddbRecords []ddb.DynamoRecord
	for _, connectId := range data.ConnectIds {
		ddbRecords = append(ddbRecords, ddb.DynamoRecord{
			Pk: data.DomainName,
			Sk: connectId,
		})
	}
	searchResults := []uc.PrivacySearchResult{}
	err = services.PrivacyResultsDB.GetMany(ddbRecords, &searchResults)
	if err != nil {
		tx.Error("[%s] Error getting Privacy Search Results: %v", err)
		return err
	}

	//	Set status of privacy search results records to PURGE_RUNNING
	//	This sets records with PK: domainName, SK: connectId|locationType to PURGE_RUNNING
	err = u.setPurgeStatus(
		searchResults,
		uc.PRIVACY_STATUS_PURGE_RUNNING,
		services,
		tx,
		uc.UpdatePurgeStatusOptions{
			UpdateMode: uc.LOCATION_RECORD,
		},
	)
	if err != nil {
		tx.Error("[%s] Error setting purge status: %v", u.name, err)
		return err
	}

	var gdprPurgeLog = cw.Init(services.Env["REGION"], services.Env["METRICS_SOLUTION_ID"], services.Env["METRICS_SOLUTION_VERSION"])
	logStreamName := fmt.Sprintf("%s-%s", t.Now().Format(t.DateOnly), tx.TransactionID)
	err = gdprPurgeLog.CreateLogStream(services.Env["GDPR_PURGE_LOG_GROUP_NAME"], logStreamName)
	if err != nil {
		tx.Error("[%s] Error creating log stream: %v", u.name, err)
		return err
	}
	err = gdprPurgeLog.SendLog(services.Env["GDPR_PURGE_LOG_GROUP_NAME"], logStreamName, fmt.Sprintf("ConnectIds: %v purged by: %s for domain: %s", data.ConnectIds, data.AgentCognitoId, data.DomainName))
	if err != nil {
		tx.Error("[%s] Error sending log: %v", u.name, err)
		return err
	}

	//	Setup WaitGroup, error
	var wg sync.WaitGroup
	var errors []error
	var errorMutex sync.Mutex

	//	Setup search funcs
	searchFunctions := map[uc.LocationType]func(searchResults []uc.PrivacySearchResult, services async.Services, tx core.Transaction, options ...uc.PurgeOptions) error{
		uc.S3LocationType:             u.purgeS3ChangeProcBucket,
		uc.ProfileStorageLocationType: u.purgeProfileStorage,
		uc.MatchDbLocationType:        u.purgeMatchDb,
	}

	purgeOptions := uc.PurgeOptions{
		OperatorID: data.AgentCognitoId,
	}

	//	Invoke goroutines for each defined search func
	for key, searchFunc := range searchFunctions {
		wg.Add(1)
		go func(key uc.LocationType, searchFunc func([]uc.PrivacySearchResult, async.Services, core.Transaction, ...uc.PurgeOptions) error) {
			defer wg.Done()
			err := searchFunc(searchResults, services, tx, purgeOptions)
			if err != nil {
				errorMutex.Lock()
				defer errorMutex.Unlock()
				errors = append(errors, err)
			}
		}(key, searchFunc)
	}

	//	Wait for all goroutines to finish
	wg.Wait()

	if len(errors) > 0 {
		tx.Error("[%s] Error(s) occurred during purge: %v", u.name, errors)
	}

	return core.ErrorsToError(errors)
}

func (u *PurgeProfileDataUsecase) purgeProfileStorage(searchResults []uc.PrivacySearchResult, services async.Services, tx core.Transaction, options ...uc.PurgeOptions) error {
	tx.Info("[%s] Purging CustomerProfile data", u.name)

	operatorId := ""
	for _, option := range options {
		if option.OperatorID != "" {
			operatorId = option.OperatorID
		}
	}

	var errors []error
	for _, searchResult := range searchResults {
		endingStatus := uc.PRIVACY_STATUS_PURGE_SUCCESS
		errorMessage := ""
		err := services.AccpConfig.DeleteProfile(searchResult.ConnectId, customerprofileslcs.DeleteProfileOptions{OperatorID: operatorId, DeleteHistory: true})
		if err != nil {
			tx.Error("[%s] Error deleting profile with connectId (%s): %v", u.name, searchResult.ConnectId, err)
			endingStatus = uc.PRIVACY_STATUS_PURGE_FAILED
			errorMessage = fmt.Sprintf("TxID: %s", tx.TransactionID)
			errors = append(errors, err)
		}
		err = u.setPurgeStatus(
			[]uc.PrivacySearchResult{searchResult},
			endingStatus,
			services,
			tx,
			uc.UpdatePurgeStatusOptions{
				UpdateMode:   uc.LOCATION_RECORD,
				LocationType: uc.ProfileStorageLocationType,
				ErrorMessage: errorMessage,
			},
		)
		if err != nil {
			tx.Error("[%s] Error setting PrivacyResults table status to PRIVACY_STATUS_PURGE_FAILED: %v", u.name, err)
			errors = append(errors, err)
		}
	}

	return core.ErrorsToError(errors)
}

func (u *PurgeProfileDataUsecase) purgeMatchDb(searchResults []uc.PrivacySearchResult, services async.Services, tx core.Transaction, options ...uc.PurgeOptions) error {
	tx.Info("[%s] Purging MatchDB data", u.name)
	deleteRequest := []map[string]string{}
	endingStatus := uc.PRIVACY_STATUS_PURGE_SUCCESS
	errorMessage := ""
	var errors []error
	for _, searchResult := range searchResults {
		matchResults := []model.MatchPair{}
		err := services.MatchDB.FindStartingWith(fmt.Sprintf("%s_%s", searchResult.DomainName, searchResult.ConnectId), "match_", &matchResults)
		if err != nil {
			tx.Error("[%s] Error finding match pairs: %v", u.name, err)
			errors = append(errors, err)
		}
		for _, match := range matchResults {
			deleteRequest = append(deleteRequest, map[string]string{
				services.MatchDB.PrimaryKey: match.Pk,
				services.MatchDB.SortKey:    match.Sk,
			})
			deleteRequest = append(deleteRequest, map[string]string{
				services.MatchDB.PrimaryKey: fmt.Sprintf("%s_%s", searchResult.DomainName, match.TargetProfileID),
				services.MatchDB.SortKey:    fmt.Sprintf("match_%s", searchResult.ConnectId),
			})
		}
	}
	err := services.MatchDB.DeleteMany(deleteRequest)
	if err != nil {
		tx.Error("[%s] Error deleting match pairs: %v", u.name, err)
		endingStatus = uc.PRIVACY_STATUS_PURGE_FAILED
		errorMessage = fmt.Sprintf("TxID: %s", tx.TransactionID)
		errors = append(errors, err)
	}
	err = u.setPurgeStatus(
		searchResults,
		endingStatus,
		services,
		tx,
		uc.UpdatePurgeStatusOptions{
			UpdateMode:   uc.LOCATION_RECORD,
			LocationType: uc.MatchDbLocationType,
			ErrorMessage: errorMessage,
		},
	)
	if err != nil {
		tx.Error("[%s] Error setting PrivacyResults table status to PRIVACY_STATUS_PURGE_FAILED: %v", u.name, err)
		errors = append(errors, err)
	}

	return core.ErrorsToError(errors)
}

func (u *PurgeProfileDataUsecase) purgeS3ChangeProcBucket(searchResults []uc.PrivacySearchResult, services async.Services, tx core.Transaction, options ...uc.PurgeOptions) error {
	tx.Info("[%s] Purging S3 data", u.name)
	resultMap := map[string][]string{}
	for _, searchResult := range searchResults {
		if _, ok := searchResult.LocationResults[uc.S3LocationType]; ok {
			for _, s3Location := range searchResult.LocationResults[uc.S3LocationType] {
				resultMap[s3Location] = append(resultMap[s3Location], searchResult.ConnectId)
			}
		}
	}

	tx.Debug("[%s] S3 Path -> ConnectId map: %+v", u.name, resultMap)
	//	Iterate over map's keys; invoke an async lambda usecase with KVP: S3Path -> []connectId
	var errors []error
	for s3Path, connectIds := range resultMap {
		body := uc.PurgeFromS3Body{
			S3Path:     s3Path,
			ConnectIds: connectIds,
			DomainName: u.domainName,
			TxID:       tx.TransactionID,
		}

		sqsMsg, err := json.Marshal(body)
		if err != nil {
			tx.Error("[%s] Error marshalling payload: %v", u.name, err)
			errors = append(errors, err)
		}
		err = services.S3ExciseSqsConfig.Send(string(sqsMsg))
		if err != nil {
			tx.Error("[%s] Error sending sqs message: %v", u.name, err)
			errors = append(errors, err)
		}
	}
	return core.ErrorsToError(errors)
}

func (u *PurgeProfileDataUsecase) setPurgeStatus(
	searchResults []uc.PrivacySearchResult,
	status uc.PrivacySearchStatus,
	services async.Services,
	tx core.Transaction,
	options ...uc.UpdatePurgeStatusOptions,
) error {
	//	Default options
	var updateMode uc.UpdateMode = uc.MAIN_RECORD
	var locationType uc.LocationType = ""
	var errMsg string = ""
	for _, o := range options {
		if o.UpdateMode != 0 {
			updateMode = o.UpdateMode
		}
		if o.LocationType != "" {
			locationType = o.LocationType
		}
		if o.ErrorMessage != "" {
			errMsg = o.ErrorMessage
		}
	}

	//	Update PrivacySearchResult entry for domainName + connectId; set status to PURGE_RUNNING
	var err error
	for _, searchResult := range searchResults {
		if updateMode&uc.MAIN_RECORD != 0 {
			err = services.PrivacyResultsDB.UpdateItems(
				searchResult.DomainName,
				searchResult.ConnectId,
				map[string]interface{}{
					utils.GetJSONTag(uc.PrivacySearchResult{}, "Status"):       string(status),
					utils.GetJSONTag(uc.PrivacySearchResult{}, "ErrorMessage"): errMsg,
					utils.GetJSONTag(uc.PrivacySearchResult{}, "TxID"):         tx.TransactionID,
				},
			)
			if err != nil {
				tx.Error("[%s] Error updating MainRecord in PrivacyResults table: %v", u.name, err)
				return err
			}
		}

		//	Upsert PrivacySearchResult entry for domainName + connectId|locationResult; set status to PURGE_RUNNING
		if updateMode&uc.LOCATION_RECORD != 0 {
			for locType, locations := range searchResult.LocationResults {
				if locationType == "" || locationType == locType {
					for _, location := range locations {
						err = services.PrivacyResultsDB.UpdateItems(
							searchResult.DomainName,
							fmt.Sprintf("%s|%s", searchResult.ConnectId, location),
							map[string]interface{}{
								utils.GetJSONTag(uc.PrivacySearchResult{}, "Status"):       string(status),
								utils.GetJSONTag(uc.PrivacySearchResult{}, "ErrorMessage"): errMsg,
								utils.GetJSONTag(uc.PrivacySearchResult{}, "TxID"):         tx.TransactionID,
							},
						)
						if err != nil {
							tx.Error("[%s] Error updating LocationRecord in PrivacyResults table: %v", u.name, err)
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
