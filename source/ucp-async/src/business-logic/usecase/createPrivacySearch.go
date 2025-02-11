// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	storage "tah/upt/source/storage"
	"tah/upt/source/tah-core/athena"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	match "tah/upt/source/ucp-common/src/model/admin"
	model "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

type CreatePrivacySearchUsecase struct {
	name string
}

// Static Interface Check
var _ Usecase = &CreatePrivacySearchUsecase{}

func InitCreatePrivacySearch() *CreatePrivacySearchUsecase {
	return &CreatePrivacySearchUsecase{
		name: "CreatePrivacySearch",
	}
}

func (u *CreatePrivacySearchUsecase) Name() string {
	return u.name
}

func (u *CreatePrivacySearchUsecase) Execute(payload model.AsyncInvokePayload, services model.Services, tx core.Transaction) error {
	//	Retrieve payload details
	payloadJson, err := json.Marshal(payload.Body)
	if err != nil {
		tx.Error("[%s] Error marshalling payload: %v", u.name, err)
		return err
	}

	//	Convert json to CreatePrivacySearchBody
	data := uc.CreatePrivacySearchBody{}
	err = json.Unmarshal(payloadJson, &data)
	if err != nil {
		tx.Error("[%s] Error unmarshalling payload body: %v", u.name, err)
		return err
	}
	tx.Debug("[%s] Payload body: %+v", u.name, data)

	tx.Debug("[%s] Setting domain to ACCP config to %s", u.name, data.DomainName)
	services.AccpConfig.SetDomain(data.DomainName)

	var connectIdWg sync.WaitGroup
	var errorMutex sync.Mutex
	var errs error
	for _, connectId := range data.ConnectIds {
		connectIdWg.Add(1)
		go func(connectId string, domainName string, services model.Services, tx core.Transaction) {
			defer connectIdWg.Done()
			err := u.SearchForMatches(connectId, data.DomainName, services, tx)
			if err != nil {
				errorMutex.Lock()
				defer errorMutex.Unlock()
				errs = errors.Join(err)
			}
		}(connectId, data.DomainName, services, tx)

	}
	connectIdWg.Wait()

	return errs
}

func (u *CreatePrivacySearchUsecase) SearchForMatches(connectId string, domainName string, services model.Services, tx core.Transaction) error {
	var errorMutex sync.Mutex
	var localErrors []error

	//	Clear PrivacySearchResult records with PK: domainName, SK: connectId|locationType
	existingPurgeRecords := []uc.PrivacySearchResult{}
	err := services.PrivacyResultsDB.FindStartingWithAndFilterWithIndex(domainName, fmt.Sprintf("%s|", connectId), &existingPurgeRecords, db.QueryOptions{
		ProjectionExpression: db.DynamoProjectionExpression{
			Projection: []string{"domainName", "connectId"},
		},
	})
	if err != nil {
		tx.Error("[%s] Error finding existing purge records: %v", u.name, err)
		localErrors = append(localErrors, err)
		return core.ErrorsToError(localErrors)
	}
	deleteRequest := []map[string]string{}
	for _, purgeRecord := range existingPurgeRecords {
		deleteRequest = append(deleteRequest, map[string]string{
			services.PrivacyResultsDB.PrimaryKey: purgeRecord.DomainName,
			services.PrivacyResultsDB.SortKey:    purgeRecord.ConnectId,
		})
	}
	err = services.PrivacyResultsDB.DeleteMany(deleteRequest)
	if err != nil {
		tx.Error("[%s] Error deleting existing purge records: %v", u.name, err)
		localErrors = append(localErrors, err)
		return core.ErrorsToError(localErrors)
	}

	//	Add to search result table with pending status
	searchDate := time.Now()
	searchResult := uc.PrivacySearchResult{
		DomainName: domainName,
		ConnectId:  connectId,
		Status:     uc.PRIVACY_STATUS_SEARCH_RUNNING,
		SearchDate: searchDate,
	}
	_, err = services.PrivacyResultsDB.Save(searchResult)
	if err != nil {
		tx.Error("[%s] Error saving privacy search results for connectId: %s: %v", u.name, connectId, err)
		localErrors = append(localErrors, err)
		return core.ErrorsToError(localErrors)
	}

	// Setup WaitGroup, error, and mutex for goroutines
	var wg sync.WaitGroup
	var searchResultsMutex sync.Mutex

	//	Setup map for location storage results
	searchResults := map[uc.LocationType][]string{}

	//	Setup search funcs
	searchFunctions := map[uc.LocationType]func() ([]string, error){
		uc.S3LocationType: func() ([]string, error) {
			tx.Info("[%s] Starting search for s3 change processor results", u.name)
			results, err := u.SearchS3ChangeProcBucket(connectId, domainName, services, tx)
			if err != nil {
				tx.Error("[%s] Error searching output bucket: %v", u.name, err)
				return []string{}, err
			}
			return results, nil
		},
		uc.ProfileStorageLocationType: func() ([]string, error) {
			tx.Info("[%s] Starting search for profile storage results", u.name)
			results, err := u.SearchProfileStorage(connectId, services, tx)
			if err != nil {
				tx.Error("[%s] Error searching profile storage: %v", u.name, err)
				return []string{}, err
			}
			return results, nil
		},
		uc.MatchDbLocationType: func() ([]string, error) {
			tx.Info("[%s] Starting search for entity resolution results", u.name)
			results, err := u.SearchERMatches(connectId, domainName, services, tx)
			if err != nil {
				tx.Error("[%s] Error searching entity resolution matches: %v", u.name, err)
				return []string{}, err
			}
			return results, nil
		},
	}

	//	Invoke goroutines for each defined search func
	totalResultsRound := 0
	for key, searchFunc := range searchFunctions {
		wg.Add(1)
		go func(key uc.LocationType, f func() ([]string, error)) {
			defer wg.Done()
			results, err := f()
			if err != nil {
				tx.Error("[%s] Error invoking search func: %v", u.name, err)
				errorMutex.Lock()
				defer errorMutex.Unlock()
				localErrors = append(localErrors, err)
				return
			}
			searchResultsMutex.Lock()
			defer searchResultsMutex.Unlock()
			searchResults[key] = results
			totalResultsRound += len(results)
		}(key, searchFunc)
	}

	//	Wait for all goroutines to finish
	wg.Wait()
	endingStatus := uc.PRIVACY_STATUS_SEARCH_SUCCESS
	errorMessage := ""
	if len(localErrors) != 0 {
		endingStatus = uc.PRIVACY_STATUS_SEARCH_FAILED
		errorMessage = "TxID: " + tx.TransactionID
	}

	searchResult = uc.PrivacySearchResult{
		DomainName:        domainName,
		ConnectId:         connectId,
		Status:            endingStatus,
		SearchDate:        searchDate,
		LocationResults:   searchResults,
		TotalResultsFound: totalResultsRound,
		ErrorMessage:      errorMessage,
	}

	_, err = services.PrivacyResultsDB.Save(searchResult)
	if err != nil {
		tx.Error("[%s] Error saving privacy search results: %v", u.name, err)
		localErrors = append(localErrors, err)
	}

	return core.ErrorsToError(localErrors)
}

func (u *CreatePrivacySearchUsecase) SearchS3ChangeProcBucket(connectId string, domainName string, services model.Services, tx core.Transaction) ([]string, error) {
	//	Extract values from Environment
	glueDbName := services.Env["GLUE_DB"]
	glueTableName := services.Env["GLUE_EXPORT_TABLE_NAME"]
	glueWorkgroup := services.Env["ATHENA_WORKGROUP_NAME"]
	region := services.Env["REGION"]
	athenaOutputBucket := services.Env["ATHENA_OUTPUT_BUCKET"]
	solutionId := services.Env["METRICS_SOLUTION_ID"]
	solutionVersion := services.Env["METRICS_SOLUTION_VERSION"]
	getS3PathsPreparedStatementName := services.Env["GET_S3_PATHS_PREPARED_STATEMENT_NAME"]
	logLevel := services.Env["LOG_LEVEL"]
	athenaCfg := athena.Init(glueDbName, glueTableName, glueWorkgroup, region, athenaOutputBucket, solutionId, solutionVersion, core.LogLevelFromString(logLevel))

	//	Search Output Bucket
	queryResult, err := athenaCfg.ExecutePreparedStatement(getS3PathsPreparedStatementName, []*string{
		aws.String(connectId),
		aws.String(domainName),
	}, athena.Option{})
	if err != nil {
		tx.Error("[%s] Error executing prepared statement: %v", u.name, err)
		return nil, err
	}

	foundPaths := []string{}
	for i, row := range queryResult {
		//	Skip First Entry -- it's the table header
		if i != 0 {
			foundPaths = append(foundPaths, row["$path"])
		}
	}

	return foundPaths, nil
}

func (u *CreatePrivacySearchUsecase) SearchProfileStorage(connectId string, services model.Services, tx core.Transaction) ([]string, error) {
	profile, err := services.AccpConfig.GetProfile(connectId, []string{}, []storage.PaginationOptions{})
	if err != nil {
		if errors.Is(err, storage.ErrProfileNotFound) {
			tx.Warn("[%s] Profile not found for connectId: %s", u.name, connectId)
			return []string{}, nil
		}
		tx.Error("[%s] Error searching profiles: %v", u.name, err)
		return nil, err
	}
	foundPaths := []string{}
	foundPaths = append(foundPaths, "profileId: "+profile.ProfileId)
	return foundPaths, nil
}

func (u *CreatePrivacySearchUsecase) SearchERMatches(connectId string, domainName string, services model.Services, tx core.Transaction) ([]string, error) {
	matchPairs := []match.MatchPair{}
	err := services.MatchDB.FindByPk(fmt.Sprintf("%s_%s", domainName, connectId), &matchPairs)
	if err != nil {
		tx.Error("[%s] Error locating profiles: %v", u.name, err)
		return nil, err
	}

	foundPaths := []string{}
	for _, match := range matchPairs {
		foundPaths = append(foundPaths, fmt.Sprintf("%s://%s/%s", services.MatchDB.TableName, match.Pk, match.Sk))
	}
	return foundPaths, nil
}
