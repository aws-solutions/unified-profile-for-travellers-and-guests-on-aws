// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/cognito"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/glue"
	common "tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/feature"
	accpMappings "tah/upt/source/ucp-common/src/model/accp-mappings"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	assetsSchema "tah/upt/source/ucp-common/src/model/assets-schema"
	model "tah/upt/source/ucp-common/src/model/async"
	uc "tah/upt/source/ucp-common/src/model/async/usecase"
	commonServices "tah/upt/source/ucp-common/src/services/admin"

	"golang.org/x/sync/errgroup"
)

type CreateDomainUsecase struct {
	name string
}

func InitCreateDomain() *CreateDomainUsecase {
	return &CreateDomainUsecase{
		name: "CreateDomain",
	}
}

func (u *CreateDomainUsecase) Name() string {
	return u.name
}

type CreateDomainConfig struct {
	CustomerProfileStorageMode bool
	DynamoStorageMode          bool
	FeatureSetVersion          int
	CpExportStream             string
}

/*
Create a domain and relevant tasks
1. Create Customer Profiles domain
2. Create mapping between business objects and Customer Profiles data store
3. Create integration (ACCP only)
4. Create "admin" group in Cognito
5. Create Glue table for each business object S3 bucket
6. Configure event stream export from ACCP to the Change Processor (ACCP only)
7. Retry Lambda
8. Add a partition to profile export bucket for the new domain
9. Add a default rule to auto merge Profiles with one or many Alternate Profile IDs

ACCP only steps are called regardless of mode (ACCP vs LCS), but they are not necessary for LCS.
The LCS functions just return a successful response.
*/
func (u *CreateDomainUsecase) Execute(payload model.AsyncInvokePayload, services model.Services, tx core.Transaction) error {
	// Convert payload from string to json data
	payloadJson, err := json.Marshal(payload.Body)
	if err != nil {
		tx.Error("[%s] Error marshalling payload %v", u.name, err)
		return err
	}

	// Convert json to CreateDomainBody
	data := uc.CreateDomainBody{}
	err = json.Unmarshal(payloadJson, &data)
	if err != nil {
		tx.Error("[%s] Error unmarshalling payload body %v", u.name, err)
		return err
	}
	tx.Debug("[%s] Payload body: %+v", u.name, data)

	// getting deployment modes
	customerProfileStorageModeString := services.Env["CUSTOMER_PROFILE_STORAGE_MODE"]
	dynamoStorageModeString := services.Env["DYNAMO_STORAGE_MODE"]
	cpExportStream := services.Env["CP_EXPORT_STREAM"]

	if data.RequestedVersion == 0 {
		// if unspecified, create the newest domain
		data.RequestedVersion = feature.LATEST_FEATURE_SET_VERSION
	}
	if data.RequestedVersion > feature.LATEST_FEATURE_SET_VERSION || data.RequestedVersion < feature.INITIAL_FEATURE_SET_VERSION {
		return fmt.Errorf("unknown feature set version requested: %d", data.RequestedVersion)
	}

	config := CreateDomainConfig{
		CustomerProfileStorageMode: customerProfileStorageModeString == "true",
		DynamoStorageMode:          dynamoStorageModeString == "true",
		FeatureSetVersion:          data.RequestedVersion,
		CpExportStream:             cpExportStream,
	}
	// If RequestedVersion is unspecified (indicated by go's zero value for int i.e 0), we will assume the feature set version to be latest
	if data.RequestedVersion == 0 {
		config.FeatureSetVersion = feature.LATEST_FEATURE_SET_VERSION
	}

	tx.Info("[%s] Customer Profile storage mode enabled: %t", u.name, config.CustomerProfileStorageMode)
	tx.Info("[%s] DynamoDB storage mode enabled: %t", u.name, config.DynamoStorageMode)
	return u.createDomain(services, data, config, tx)
}

/*
* Services used:
 1. AsyncLambdaConfig
 2. AccpConfig
 3. PortalConfigDB
 4. GlueConfig
 5. CognitoConfig
 6. RetryLambdaConfig
 7. ErrorDB
*/
func (u *CreateDomainUsecase) createDomain(
	services model.Services,
	data uc.CreateDomainBody,
	config CreateDomainConfig,
	tx core.Transaction,
) error {
	tx.Info("[%s] Creating domain %s", u.name, data.DomainName)
	tags, err := services.AsyncLambdaConfig.ListTags()
	if err != nil {
		tx.Error("[%s] error listing function tags: %v", u.name, err)
	}
	tags["envName"] = data.Env
	tags["aws_solution"] = "SO0244"
	filteredTags := filterTags(tags)
	options := customerprofiles.DomainOptions{
		AiIdResolutionOn:       false,
		RuleBaseIdResolutionOn: true,
		CustomerProfileMode:    &config.CustomerProfileStorageMode,
		DynamoMode:             &config.DynamoStorageMode,
		FeatureSetVersion:      config.FeatureSetVersion,
		CpExportDestinationArn: config.CpExportStream,
	}

	// Create domain
	err = services.AccpConfig.CreateDomainWithQueue(data.DomainName, data.KmsArn, filteredTags, data.AccpDlq, data.MatchBucketName, options)
	if err != nil {
		tx.Error("[%s] Error creating domain %s", u.name, err)
		return err
	}

	// Defer rollback: delete domain
	defer func() {
		if err != nil {
			tx.Info("[%s] Attempting to delete domain", u.name)
			rollbackErr := services.AccpConfig.DeleteDomain()
			if rollbackErr != nil {
				tx.Warn("[%s] Error deleting domain during attempted rollback: %v", u.name, rollbackErr)
			}
		}
	}()

	// errgroup to manage domain creation tasks
	mappingErrGroup := new(errgroup.Group)
	g := new(errgroup.Group)

	// Create mappings and integrations
	mappingErrGroup.Go(func() error {
		err := createMappingsAndIntegrations(services.AccpConfig, tx)
		if err != nil {
			tx.Error("[%s] Error creating mappings: %v", u.name, err)
		}
		return err
	})

	// Set default AI summary prompt
	g.Go(func() error {
		err := setDefaultPrompt(tx, services, data.DomainName)
		if err != nil {
			tx.Error("[%s] Error saving prompt configuration, ", u.name, err)
		}
		return err
	})

	// Set default AI match threshold
	g.Go(func() error {
		err := setDefaultAIMatchThreshold(tx, services, data.DomainName)
		if err != nil {
			tx.Error("[%s] Error saving AI match threshold, ", u.name, err)
		}
		return err
	})

	// Create Glue tables for each business object
	bizObjectBuckets := map[string]string{
		common.BIZ_OBJECT_HOTEL_BOOKING: data.HotelBookingBucket,
		common.BIZ_OBJECT_AIR_BOOKING:   data.AirBookingBucket,
		common.BIZ_OBJECT_GUEST_PROFILE: data.GuestProfileBucket,
		common.BIZ_OBJECT_PAX_PROFILE:   data.PaxProfileBucket,
		common.BIZ_OBJECT_CLICKSTREAM:   data.ClickstreamBucket,
		common.BIZ_OBJECT_STAY_REVENUE:  data.StayRevenueBucket,
		common.BIZ_OBJECT_CSI:           data.CSIBucket,
	}
	for _, businessObject := range common.BUSINESS_OBJECTS {
		businessObject := businessObject
		// Create Glue table for business object
		g.Go(func() error {
			err := createGlueTable(tx, services, data, businessObject, bizObjectBuckets)
			if err != nil {
				tx.Error("[%s] Error creating Glue table for %s: %v", u.name, businessObject, err)
			}
			return err
		})

		// Defer rollback: delete Glue tables
		defer func() {
			if err != nil {
				err = rollBackGlueTableOnFailure(tx, services, data, businessObject)
				if err != nil {
					tx.Warn("[%s] Error deleting Glue table during attempted rollback: %v", u.name, err)
				}
			}
		}()
	}

	// Create Cognito groups
	if isPermissionSystemEnabled(services.Env) {
		defer func() {
			if err != nil {
				tx.Info("[%s] Attempting to rollback cognito groups", u.name)
				deleteCognitoGroupError := deleteCognitoGroups(services.CognitoConfig, tx, u.name, data.DomainName)
				if deleteCognitoGroupError != nil {
					tx.Warn("[%s] Error deleting cognito groups during rollback: %v", u.name, deleteCognitoGroupError)
				}
			}
		}()

		g.Go(func() error {
			// User permissions are specific to each domain. We create an admin group
			// by default so users/admins can quickly start working in the new domain.
			tx.Info("[%s] Creating admin group", u.name)
			err := createCognitoGroup(tx, services.CognitoConfig, buildDataAccessGroupName(data.DomainName), "*/*")
			if err != nil {
				tx.Error("[%s] Error creating admin group: %v", u.name, err)
			}
			tx.Info("[%s] Creating groups for application", u.name)
			err = createAppAccessRoles(tx, services.CognitoConfig, data.DomainName)
			if err != nil {
				tx.Error("[%s] Error creating app access group: %v", u.name, err)
			}
			return err
		})
	}

	// Create Retry Lambda event mapping
	g.Go(func() error {
		tx.Info("[%s] Creating Retry Lambda Event Mapping (if not exists)", u.name)
		_, err := commonServices.CreateRetryLambdaEventSourceMapping(
			services.RetryLambdaConfig,
			services.ErrorDB,
			tx,
		)
		if err != nil {
			tx.Error("[%s] Error creating retry Lambda event mapping: %v", u.name, err)
		}
		return err
	})

	// Add domain partition to export table
	g.Go(func() error {
		tx.Info("[%s] Adding partitions to glue table created to read profiles export", u.name)
		//defined in common.go
		bucketName := services.Env["S3_EXPORT_BUCKET"]
		glueTableName := services.Env["GLUE_EXPORT_TABLE_NAME"]
		//s3 root path must be consistent with the firehose setting
		s3RootPath := services.Env["TRAVELER_S3_ROOT_PATH"]
		tx.Debug("[%s] S3_EXPORT_BUCKET: %v", u.name, bucketName)
		tx.Debug("[%s] GLUE_EXPORT_TABLE_NAME: %v", u.name, glueTableName)
		tx.Debug("[%s] TRAVELER_S3_ROOT_PATH: %v", u.name, s3RootPath)

		partition := glue.Partition{
			Values:   []string{data.DomainName},
			Location: "s3://" + bucketName + "/" + s3RootPath + "/domainname=" + data.DomainName,
		}
		tx.Debug("[%s] partitions details: %+v", u.name, partition)
		err := errors.Join(
			services.GlueConfig.AddParquetPartitionsToTable(glueTableName, []glue.Partition{partition})...)
		if err != nil {
			tx.Error("[%s] Error adding partitions to export table: %v", err)
		}
		return err
	})

	// Defer rollback: remove partitions
	defer func() {
		if err != nil {
			tx.Info("[%s] Attempting to remove partitions from export table", u.name)
			glueTableName := services.Env["GLUE_EXPORT_TABLE_NAME"]
			rollbackErr := services.GlueConfig.RemovePartitionsFromTable(
				glueTableName,
				[]glue.Partition{{Values: []string{data.DomainName}}},
			)
			if rollbackErr != nil {
				tx.Warn("[%s] Error removing export table partitions during attempted rollback: %v", u.name, rollbackErr)
			}
		}
	}()

	//	Adding the rule must be attempted after the mappings have finished
	err = mappingErrGroup.Wait()
	if err != nil {
		tx.Warn("[%s] Error creating domain, attempting rollback: %v", u.name, err)
	}
	g.Go(func() error {
		return createAlternateProfileIdMergeRule(services, tx)
	})

	err = g.Wait()
	if err != nil {
		tx.Warn("[%s] Error creating domain, attempting rollback: %v", u.name, err)
	}

	return err
}

func setDefaultPrompt(tx core.Transaction, services model.Services, domainName string) error {
	tx.Info("[CreateDomain] Setting default prompt")
	return commonServices.UpdateDomainSetting(
		tx,
		services.PortalConfigDB,
		commonServices.PORTAL_CONFIG_PROMPT_PK,
		domainName,
		INITIAL_CLAUDE_PROMPT,
		false,
	)
}

func setDefaultAIMatchThreshold(tx core.Transaction, services model.Services, domainName string) error {
	tx.Info("[CreateDomain] Setting default AI match threshold")
	return commonServices.UpdateDomainSetting(
		tx,
		services.PortalConfigDB,
		commonServices.PORTAL_CONFIG_MATCH_THRESHOLD_PK,
		domainName,
		"0.0",
		false,
	)
}

func createGlueTable(
	tx core.Transaction,
	services model.Services,
	data uc.CreateDomainBody,
	bizObject commonModel.BusinessObject,
	bucketMap map[string]string,
) error {
	tx.Info("[CreateDomain] Creating table for %s", bizObject.Name)
	schema, err := assetsSchema.LoadSchema(&tx, data.GlueSchemaPath, bizObject)
	if err != nil {
		tx.Error("[CreateDomain] Error loading schema: %v", err)
		return err
	}

	tableName := commonServices.BuildTableName(data.Env, bizObject, data.DomainName)
	glueDest := commonServices.BuildGlueDestination(bucketMap[bizObject.Name], data.DomainName)

	err = services.GlueConfig.CreateTable(
		tableName,
		glueDest,
		[]glue.PartitionKey{{Name: "year", Type: "int"}, {Name: "month", Type: "int"}, {Name: "day", Type: "int"}},
		schema,
	)
	if err != nil {
		tx.Error("[CreateDomain] Error creating table: %v", err)
		return err
	}

	return nil
}

func rollBackGlueTableOnFailure(
	tx core.Transaction,
	services model.Services,
	data uc.CreateDomainBody,
	bizObject commonModel.BusinessObject,
) error {
	tableName := commonServices.BuildTableName(data.Env, bizObject, data.DomainName)
	tx.Info("[CreateDomain] Attempting to delete Glue table %s", tableName)
	return services.GlueConfig.DeleteTable(tableName)
}

func createCognitoGroup(tx core.Transaction, cognitoClient cognito.ICognitoConfig, groupName string, groupDescription string) error {
	tx.Info("[CreateDomain] Creating admin group %s with description %s", groupName, groupDescription)
	err := cognitoClient.CreateGroup(groupName, groupDescription)
	if cognitoClient.IsGroupExistsException(err) {
		tx.Info("[CreateDomain] Group %s already exists, using existing group", groupName)
		return nil
	}
	return err
}

func createMappingsAndIntegrations(accpConfig customerprofiles.ICustomerProfileLowCostConfig, tx core.Transaction) error {
	for _, businessObject := range common.ACCP_RECORDS {
		accpRecName := businessObject.Name
		tx.Info("[CreateDomain] Creating mapping for %s", accpRecName)
		objMapping := accpMappings.ACCP_OBJECT_MAPPINGS[accpRecName]()
		err := accpConfig.CreateMapping(accpRecName, "Primary Mapping for the "+accpRecName+" object", objMapping.Fields)
		if err != nil {
			tx.Error("[CreateDomain] Error creating Mapping: %s. deleting domain", err)
			return err
		}
	}
	return nil
}

func createAppAccessRoles(tx core.Transaction, cognitoClient cognito.ICognitoConfig, domainName string) error {
	groupNamePrefix := buildAppAccessGroupPrefix(domainName)
	for roleName, rolePermission := range appAccessRoleMap {
		err := createCognitoGroup(tx, cognitoClient, buildAppAccessGroupName(groupNamePrefix, roleName, rolePermission), "")
		if err != nil {
			return err
		}
	}

	return nil
}

func filterTags(tags map[string]string) map[string]string {
	filteredTags := make(map[string]string)
	for key, value := range tags {
		if !strings.HasPrefix(key, "aws:") {
			filteredTags[key] = value
		}
	}
	return filteredTags
}

func createAlternateProfileIdMergeRule(services model.Services, tx core.Transaction) error {
	tx.Info("[CreateDomain] Adding Rule to Merge matching Alternate Profile IDs")
	rules := []customerprofiles.Rule{
		{
			Index:       0,
			Name:        "merge_same_alternate_profile_ids",
			Description: "Merge AlternateProfileID records whose name and value match.",
			Conditions: []customerprofiles.Condition{
				{
					Index:               0,
					ConditionType:       customerprofiles.CONDITION_TYPE_MATCH,
					IncomingObjectType:  common.ACCP_RECORD_ALTERNATE_PROFILE_ID,
					IncomingObjectField: "name",
					ExistingObjectType:  common.ACCP_RECORD_ALTERNATE_PROFILE_ID,
					ExistingObjectField: "name",
				},
				{
					Index:               1,
					ConditionType:       customerprofiles.CONDITION_TYPE_MATCH,
					IncomingObjectType:  common.ACCP_RECORD_ALTERNATE_PROFILE_ID,
					IncomingObjectField: "value",
					ExistingObjectType:  common.ACCP_RECORD_ALTERNATE_PROFILE_ID,
					ExistingObjectField: "value",
				},
			},
		},
	}
	addAltProfIdMergeRuleError := services.AccpConfig.SaveIdResRuleSet(rules)
	if addAltProfIdMergeRuleError != nil {
		tx.Error("[CreateDomain] Error creating Alternate Profile ID Default Merge Rule Set: %v", addAltProfIdMergeRuleError)
		return addAltProfIdMergeRuleError
	}
	addAltProfIdMergeRuleError = services.AccpConfig.ActivateIdResRuleSet()
	if addAltProfIdMergeRuleError != nil {
		tx.Error("[CreateDomain] Error activating ruleset with default alt prof id merge: %v", addAltProfIdMergeRuleError)
		return addAltProfIdMergeRuleError
	}
	return nil
}
