// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	aurora "tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/cloudwatch"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	dynamo "tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/entityresolution"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/kinesis"
	sqs "tah/upt/source/tah-core/sqs"
	"tah/upt/source/ucp-common/src/feature"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

/*********************************************************************************
* This code manages a low cost profile storage component built using Amazon Aurora and DynamoDB
* The following changes need to be made before production
 */

const DOMAIN_CONFIG_PREFIX = "domain_"
const DOMAIN_CONFIG_SK = DOMAIN_CONFIG_PREFIX + "config"
const DOMAIN_CONFIG_ENV_KEY = "envName"
const DOMAIN_MAPPING_PREFIX = DOMAIN_CONFIG_PREFIX + "mapping_"
const INDEX_PROFILE = "PROFILE"
const INDEX_UNIQUE = "UNIQUE"
const OBJECT_TYPE_NAME_FIELD = "_object_type_name"

// Config
const PROFILE_BATCH_SIZE = 100

// Custom Metrics names

// PutProfileObject
const METRIC_PUT_OBJECT_RQ_COUNT = "PutProfileObjectRqCount"                     // PutProfileObject: count of incoming request
const METRIC_PUT_OBJECT_SUCCESS_COUNT = "PutProfileObjectSuccessCount"           // PutProfileObject: count of success
const METRIC_NAME_PREP_PROFILE_OBJECT = "PrepProfileLatency"                     // PutProfileObject: Latency measurement for control plane access and profile preparation before insert
const METRIC_NAME_UPSERT_PROFILE_OBJECT = "InsertProfileLatency"                 // PutProfileObject: Latency measurement for Insertion of profile in Amazon Aurora
const METRIC_NAME_UPDATE_DOWNSTREAM_CACHE = "UpdateDownstreamCache"              // PutProfileObject: Latency measurement for insertion of profile in downstream cache (dynamo/CP)
const METRIC_NAME_PUT_PROFILE_LATENCY = "PutProfileObjectLatency"                // PutProfileObject: Insertion of profile in downstream cache (dynamo/CP)
const METRIC_NAME_DYNAMO_UPDATE_LATENCY = "DynamoUpdateLatency"                  // PutProfileObject: Latency measurement for DynamoDB Cache update
const METRIC_NAME_CP_UPDATE_LATENCY = "CPUpdateLatency"                          // PutProfileObject: Latency measurement for CP Cache Update
const METRIC_NAME_PUT_OBJECT_MATCH_COUNT = "PutProfileObjectMatchCount"          // PutProfileObject: count of records pre-merged from an existing match
const METRIC_NAME_GET_PROFILE_SEARCH_RECORD = "PutProfileGetProfileSearchRecord" //	PutProfileObject: Latency measurement for finding most up to date connectID from Master Table

// Merge and Unmerge
const METRIC_NAME_MERGE_PROFILES = "MergeProfileLatency"                         // MergeProfile: Latency measurement for overall profile merge
const METRIC_NAME_UNMERGE_PROFILES = "UnmergeProfileLatency"                     // UnMergeProfile: Latency measurement for overall profile unmerge
const METRIC_NAME_MERGE_PROFILES_UPDATE_AURORA = "MergeProfileAuroraLatency"     // MergeProfile: Latency measurement for aurora update during profile merge
const METRIC_NAME_UNMERGE_PROFILES_UPDATE_AURORA = "UnmergeProfileAuroraLatency" // UnMergeProfile: Latency measurement for aurora update during profile unmerge
const METRIC_NAME_MERGE_PROFILES_UPDATE_DYNAMO = "MergeProfileUpdateDynamoLatency"
const METRIC_NAME_MERGE_PROFILES_DELETE_PROFILE_DYNAMO = "MergeProfileDeleteDynamoLatency"
const METRIC_NAME_MERGE_PROFILES_REINDEX = "MergeProfileReindexAuroraLatency"          //latency of the actual reindex process
const METRIC_NAME_MERGE_PROFILES_REINDEX_REMATCH = "MergeProfileReindexRematchLatency" //latency of the process to identify is a newly merged profile matches other profile and send it to the queue
const METRIC_NAME_DYNAMO_FETCH_PROFILE = "DynamoFetchProfileLatency"
const METRIC_NAME_DYNAMO_FIND_STARTING_WITH_LATENCY = "DynamoFindStartingWithLatency"
const METRIC_NAME_DYNAMO_FIND_STARTING_WITH_NUM_REC = "DynamoFindStartingWithRecords"
const METRIC_NAME_AURORA_UPDATE_MASTER_TABLE = "AuroraUpdateMasterTableLatency"
const METRIC_NAME_AURORA_UPSERT_MASTER_TABLE = "AuroraUpsertMasterTableLatency"
const METRIC_NAME_AURORA_PREMERGE_INSERT_PROFILE_HISTORY = "AuroraPreMergeInsertProfileHistoryLatency"
const METRIC_NAME_AURORA_UPDATE_SEARCH_TABLE = "AuroraUpdateSearchTableLatency"
const METRIC_NAME_AURORA_INSERT_PROFILE = "AuroraInsertProfileLatency"
const METRIC_NAME_AURORA_INSERT_OBJECT = "AuroraInsertProfileObjectLatency"
const METRIC_NAME_AURORA_DELETE_PROFILE = "AuroraDeleteProfileLatency"
const METRIC_NAME_AURORA_DELETE_PROFILE_INTERACTION = "AuroraDeleteProfileInteractionLatency"
const METRIC_NAME_AURORA_DELETE_PROFILE_OBJECT_HISTORY = "AuroraDeleteProfileInteractionHistoryLatency"
const METRIC_NAME_AURORA_INSERT_HISTORY = "AuroraInsertProfileHistoryLatency"
const METRIC_NAME_SEND_CHANGE_EVENT = "SendChangeEventLatency"

// Rule Based Identity Resolution
const METRIC_NAME_PUT_OBJECT_RULE_BASED = "PutProfileRuleBasedLatency" // Latency measurement for rule based matching execution
const METRIC_NAME_ADD_PROFILE_TO_INDEX_TABLE = "IndexProfileLatency"   // Latency measurement for update of the match table
const METRIC_NAME_RULE_EXECUTION_COUNT = "RuleExecutionCount"          // Count of rule execution by rule id and version
const METRIC_NAME_FIND_MATCHES = "FindMatchesLatency"

// Error Counts
const METRIC_AURORA_INSERT_ROLLBACK = "AuroraInsertRollbackCount" // Count of error inserting object into aurora
const METRIC_DYNAMO_INSERT_ERROR = "DynamoInsertErrorCount"       // Count of error inserting object into DynamoDB
const METRIC_CP_INSERT_ROLLBACK = "CPInsertErrorCount"            // Count of error inserting object into CP
const METRIC_AURORA_MERGE_ROLLBACK = "AuroraMergeRollbackCount"   // Count of error merging profile into aurora
// profile count
const TOTAL_PROFILE_COUNT = "TotalProfileCount"

// count table
const (
	METRIC_NAME_UPDATE_PROFILE_COUNT_INSERT  = "UpdateProfileCountOnInsertLatency"
	METRIC_NAME_UPDATE_OBJECT_COUNT_INSERT   = "UpdateObjectCountOnInsertLatency"
	METRIC_NAME_UPDATE_PROFILE_COUNT_DELETE  = "UpdateProfileCountOnDeleteLatency"
	METRIC_NAME_UPDATE_OBJECT_COUNT_DELETE   = "UpdateObjectCountOnDeleteLatency"
	METRIC_NAME_UPDATE_PROFILE_COUNT_MERGE   = "UpdateProfileCountOnMergeLatency"
	METRIC_NAME_UPDATE_PROFILE_COUNT_UNMERGE = "UpdateProfileCountOnUnmergeLatency"
	METRIC_NAME_UPDATE_PROFILE_COUNT_UPDATE  = "UpdateProfileCountOnUpdateLatency"
)

// field used for the profile id column in the db
const LC_PROFILE_ID_KEY = "profile_id"

// supported data types field mappings
const (
	MappingTypeString  string = "STRING"
	MappingTypeText    string = "TEXT"
	MappingTypeInteger string = "INTEGER"
)

// mapping between the profile field mapping type and the aurora column type
var TYPE_MAPPING = map[string]string{
	MappingTypeString:  "VARCHAR(255)",
	MappingTypeText:    "TEXT",
	MappingTypeInteger: "INTEGER",
}

type CustomerProfileConfigLowCost struct {
	DomainName            string
	Region                string
	KinesisClient         kinesis.IConfig
	EntityConfig          entityresolution.Config
	CustomerProfileConfig *customerprofiles.CustomerProfileConfig
	DomainConfig          LcDomainConfig
	GlueConfig            glue.Config
	MergeQueueClient      sqs.IConfig
	CPWriterClient        sqs.IConfig
	RoleArn               string
	S3BucketName          string
	Data                  DataHandler
	Config                ConfigHandler
	IR                    IdentityResolutionHandler
	Tx                    core.Transaction
	MetricLogger          *cloudwatch.MetricLogger
	features              feature.FeatureSet
}

type CustomerProfileInitOptions struct {
	GlueDb           string
	RoleArn          string
	S3Bucket         string
	MergeQueueClient sqs.IConfig
	AuroraReadOnly   *aurora.PostgresDBConfig
}

// interface implementation of LCS
type ICustomerProfileLowCostConfig interface {
	CreateMapping(name string, description string, fieldMappings []FieldMapping) error
	SetTx(tx core.Transaction)
	CreateDomainWithQueue(
		name string,
		kmsArn string,
		tags map[string]string,
		queueUrl string,
		matchS3Name string,
		options DomainOptions,
	) error
	ListDomains() ([]Domain, error)
	DeleteDomain() error
	GetProfile(id string, objectTypeNames []string, pagination []PaginationOptions) (profilemodel.Profile, error)
	SearchProfiles(key string, values []string) ([]profilemodel.Profile, error)
	AdvancedProfileSearch(searchCriteria []BaseCriteria) ([]profilemodel.Profile, error)
	DeleteProfile(id string, options ...DeleteProfileOptions) error
	GetDomain() (Domain, error)
	GetMappings() ([]ObjectMapping, error)
	GetProfileLevelFields() ([]string, error)
	GetObjectLevelFields() (map[string][]string, error)
	DeleteDomainByName(name string) error
	MergeProfiles(profileId string, profileIdToMerge string, mergeContext ProfileMergeContext) (string, error)
	MergeMany(pairs []ProfilePair) (string, error)
	SetDomain(domain string) error
	PutProfileObject(object, objectTypeName string) error
	RunRuleBasedIdentityResolution(object, objectTypeName, connectID string, ruleSet RuleSet) (bool, error)
	UnmergeProfiles(
		toUnmergeConnectID string,
		mergedIntoConnectID string,
		interactionIdToUnmerge string,
		interactionType string,
	) (string, error)
	FindCurrentParentOfProfile(domain, originalConnectId string) (string, error)
	IsThrottlingError(err error) bool
	GetProfileId(profileKey, profileId string) (string, error)
	GetSpecificProfileObject(objectKey string, objectId string, profileId string, objectTypeName string) (profilemodel.ProfileObject, error)
	ActivateIdResRuleSet() error
	ActivateCacheRuleSet() error
	SaveIdResRuleSet(rules []Rule) error
	SaveCacheRuleSet(rules []Rule) error
	ListIdResRuleSets(includeHistorical bool) ([]RuleSet, error)
	ListCacheRuleSets(includeHistorical bool) ([]RuleSet, error)
	GetActiveRuleSet() (RuleSet, error)
	EnableRuleBasedIdRes() error
	DisableRuleBasedIdRes() error
	BuildInteractionHistory(domain string, interactionId string, objectTypeName string, connectId string) ([]ProfileMergeContext, error)
	GetInteractionTable(
		domainName, objectType string,
		partition UuidPartition,
		lastId uuid.UUID,
		limit int,
	) ([]map[string]interface{}, error)
	ClearDynamoCache() error
	ClearCustomerProfileCache(tags map[string]string, matchS3Name string) error
	CacheProfile(connectID string, cacheMode CacheMode) error
	GetProfilePartition(partition UuidPartition) ([]string, error)
	GetUptID(profileId string) (string, error)
	UpdateVersion(version feature.FeatureSetVersion) error
	GetVersion() (feature.FeatureSetVersion, error)
	InvalidateStitchingRuleSetsCache() error
}

// Provides static check for the interface implementation without allocating memory at runtime
var _ ICustomerProfileLowCostConfig = &CustomerProfileConfigLowCost{}

func InitLowCost(
	region string,
	auroraClient *aurora.PostgresDBConfig,
	dynamoClientCfg *dynamo.DBConfig,
	kinesisCfg kinesis.IConfig,
	cpWriterQueueClient sqs.IConfig,
	metricNamespace string,
	solutionId string,
	solutionVersion string,
	logLevel core.LogLevel,
	options *CustomerProfileInitOptions,
	cache *cache.Cache,
) *CustomerProfileConfigLowCost {
	//this DynamoDB Client is to be used for dynamic table creation per domain

	//the one passed to the function is to handle config table for the service. To be likely created with the solution infrastructure script
	glueDbName := ""
	roleArn := ""
	s3BucketName := ""
	var auroraReadOnlyClient *aurora.PostgresDBConfig
	if options != nil {
		glueDbName = options.GlueDb
		roleArn = options.RoleArn
		s3BucketName = options.S3Bucket
		auroraReadOnlyClient = options.AuroraReadOnly
	}

	glueConfig := glue.Init(region, glueDbName, solutionId, solutionVersion)
	entityConfig := entityresolution.Init(region, glueDbName, solutionId, solutionVersion)
	customerProfileConfig := customerprofiles.Init(region, solutionId, solutionVersion)

	dynDataClient := dynamo.Init("", "", "", solutionId, solutionVersion)
	tx := core.NewTransaction("customerprofiles-lc", "", logLevel)
	logger := cloudwatch.NewMetricLogger(metricNamespace)
	lc := &CustomerProfileConfigLowCost{
		Data:                  DataHandler{AurSvc: auroraClient, AurReaderSvc: auroraReadOnlyClient, DynSvc: &dynDataClient, MetricLogger: logger},
		Config:                ConfigHandler{Svc: dynamoClientCfg, MetricLogger: logger, cache: cache},
		IR:                    IdentityResolutionHandler{AurSvc: auroraClient, AurReaderSvc: auroraReadOnlyClient, DynSvc: &dynDataClient, MetricLogger: logger},
		KinesisClient:         kinesisCfg,
		Region:                region,
		CustomerProfileConfig: customerProfileConfig,
		EntityConfig:          entityConfig,
		GlueConfig:            glueConfig,
		MergeQueueClient:      options.MergeQueueClient,
		CPWriterClient:        cpWriterQueueClient,
		MetricLogger:          logger,
		RoleArn:               roleArn,
		S3BucketName:          s3BucketName,
	}
	//setting default transaction in case this is not done later
	lc.SetTx(tx)
	return lc
}

func (c *CustomerProfileConfigLowCost) SetTx(tx core.Transaction) {
	tx.LogPrefix = "customerprofiles-lc"
	c.Tx = tx
	c.Data.SetTx(tx)
	c.Config.SetTx(tx)
	c.IR.SetTx(tx)
}

func (c *CustomerProfileConfigLowCost) SetDomain(domain string) error {
	c.Tx.Info("Setting Domain: %s", domain)
	c.DomainName = domain
	tName, pk, sk := buildDynamoProfileName(domain)
	c.Data.SetDynamoProfileTableProperties(tName, pk, sk)
	cfg, err := c.Config.GetDomainConfig(domain)
	if err != nil {
		c.Tx.Error("[SetDomain] Error while getting domain config: %v", err)
		return err
	}
	c.DomainConfig = cfg
	if cfg.CustomerProfileMode {
		c.CustomerProfileConfig.SetDomain(c.DomainName)
	}
	c.features, err = feature.InitFeatureSetVersion(
		feature.FeatureSetVersion{Version: cfg.DomainVersion, CompatibleVersion: cfg.CompatibleVersion},
	)
	if err != nil {
		c.Tx.Error("[SetDomain] error initializing feature set version: %v", err)
		return err
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) CreateMapping(name string, description string, fieldMappings []FieldMapping) error {
	c.Tx.Info("[CreateMapping] Creating mapping %s", name)
	if c.DomainName == "" {
		c.Tx.Error("[CreateMapping] Domain is not set")
		return fmt.Errorf("Domain is not set")
	}
	c.Tx.Debug("[CreateMapping] 0. validating mapping name")
	err := validateMappingName(name)
	if err != nil {
		c.Tx.Error("[CreateMapping] Mapping name is invalid: %v", err)
		return err
	}
	err = validateFieldMappings(fieldMappings)
	if err != nil {
		c.Tx.Error("[CreateMapping] Field mappings are invalid: %v", err)
		return err
	}
	objMapping := ObjectMapping{Fields: fieldMappings, Description: description, Name: name}
	c.Tx.Debug("[CreateMapping] 1. inserting mapping in DynamoDB")
	err = c.Config.SaveMapping(c.DomainName, objMapping)
	if err != nil {
		c.Tx.Error("[CreateMapping] Error while saving mapping to dynamo: %v", err)
		return err
	}
	c.Tx.Debug("[CreateMapping] 2. Creating Interaction Table")
	err = c.Data.CreateInteractionTable(c.DomainName, objMapping)
	if err != nil {
		c.Tx.Error("[CreateMapping] Error while creating interaction table: %v", err)
		return err
	}
	c.Tx.Debug("[CreateMapping] 3. Initializing interaction count")
	err = c.Data.InitializeCount(c.DomainName, []string{name})
	if err != nil {
		c.Tx.Error("error initializing interaction count for object %s: %v", name, err)
		return err
	}
	c.Tx.Debug("[CreateMapping] 4. Creating Interaction History Table")
	err = c.Data.CreateObjectHistoryTable(c.DomainName, name)
	if err != nil {
		c.Tx.Error("[CreateMapping] Error while creating interaction history table: %v", err)
		return err
	}
	c.Tx.Debug("[CreateMapping] 5. Updating master table with custom fields")
	err = c.Data.AddCustomAttributesFieldToSearchTable(c.DomainName, objMapping)
	if err != nil {
		c.Tx.Error("[CreateMapping] Error while creating interaction history table: %v", err)
		return err
	}
	c.Tx.Debug("[CreateMapping] 6. Possible CP update if in CP mode")
	err = c.CreateCPDomainMapping(c.DomainName, objMapping)
	if err != nil {
		c.Tx.Error("Error while creating CP domain mapping: %v", err)
		return err
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) CreateDomainWithQueue(
	name string,
	kmsArn string,
	tags map[string]string,
	queueUrl string,
	matchS3Name string,
	options DomainOptions,
) error {
	c.Tx.Info("[CreateDomainWithQueue] Creating domain %s", name)

	// Validate domain name and tags
	if err := c.validateDomainName(name); err != nil {
		return err
	}
	if err := c.validateDomainTags(tags); err != nil {
		return err
	}

	// Check downstream cache options
	customerProfileMode, dynamoMode, err := c.checkDownstreamCacheOptions(options)
	if err != nil {
		return err
	}

	// Set initial AiIdResolutionOn value
	options.AiIdResolutionOn = false

	cpDomainParams := CpDomainParams{
		queueUrl:  queueUrl,
		kmsArn:    kmsArn,
		streamArn: options.CpExportDestinationArn,
	}
	// Create domain config
	if err := c.createDomainConfig(name, matchS3Name, tags, queueUrl, options, customerProfileMode, dynamoMode); err != nil {
		return err
	}

	// Create profile caches
	if err := c.createProfileCaches(name, tags, matchS3Name, customerProfileMode, dynamoMode, cpDomainParams); err != nil {
		c.cleanupDomainData(name)
		return err
	}

	// Create profile master table
	if err := c.Data.CreateProfileMasterTable(name); err != nil {
		c.cleanupDomainData(name)
		return err
	}

	// Create search index table
	if err := c.Data.CreateProfileSearchIndexTable(name, []ObjectMapping{}, options); err != nil {
		c.cleanupDomainData(name)
		return err
	}

	// Create and initialize profile count table
	if err := c.createAndInitializeProfileCountTable(name); err != nil {
		c.cleanupDomainData(name)
		return err
	}

	// Create profile history table
	if err := c.Data.CreateProfileHistoryTable(name); err != nil {
		c.cleanupDomainData(name)
		return err
	}

	// Create IR table
	if err := c.IR.CreateIRTable(name); err != nil {
		c.cleanupDomainData(name)
		return err
	}
	c.SetDomain(name)
	return nil
}

func (c *CustomerProfileConfigLowCost) validateDomainName(name string) error {
	c.Tx.Debug("[CreateDomainWithQueue] 0.1 Validating domain name")
	err := ValidateDomainName(name)
	if err != nil {
		c.Tx.Error("[CreateDomainWithQueue] 0.1 Domain name is invalid: %v", err)
		return err
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) validateDomainTags(tags map[string]string) error {
	c.Tx.Debug("[CreateDomainWithQueue] 0.2 Validating tags")
	err := validateDomainTags(tags)
	if err != nil {
		c.Tx.Error("0.2 Domain tags are invalid: %v", err)
		return err
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) checkDownstreamCacheOptions(options DomainOptions) (bool, bool, error) {
	c.Tx.Debug("[CreateDomainWithQueue] 0.3 Checking downstream cache options")
	customerProfileMode, dynamoMode := c.getDownstreamCacheModes(options)

	if !customerProfileMode && !dynamoMode {
		return false, false, fmt.Errorf(
			"customer profile mode and dynamo mode cannot be both false, at least one downstream cache option must be chosen",
		)
	}

	return customerProfileMode, dynamoMode, nil
}

func (c *CustomerProfileConfigLowCost) getDownstreamCacheModes(options DomainOptions) (bool, bool) {
	var customerProfileMode, dynamoMode bool
	if options.CustomerProfileMode == nil {
		c.Tx.Info("[CreateDomainWithQueue] Customer profile mode is not provided. Setting it to default, false")
		customerProfileMode = false
	} else {
		customerProfileMode = *options.CustomerProfileMode
	}
	if options.DynamoMode == nil {
		c.Tx.Info("[CreateDomainWithQueue] Dynamo mode is not provided. Setting it to default, true")
		dynamoMode = true
	} else {
		dynamoMode = *options.DynamoMode
	}
	return customerProfileMode, dynamoMode
}

func (c *CustomerProfileConfigLowCost) createDomainConfig(
	name, matchS3Name string,
	tags map[string]string,
	queueUrl string,
	options DomainOptions,
	customerProfileMode, dynamoMode bool,
) error {
	c.Tx.Info("1. Creating config record in dynamo")
	c.Tx.Debug("1.1 Check if domain already exists")
	_, ok := c.Config.DomainConfigExists(name)
	if ok {
		c.Tx.Error("Domain already exists")
		return fmt.Errorf("Domain %s already exists", name)
	}
	c.Tx.Debug("[CreateDomainWithQueue] 1.2 Domain Does not exist. Creating it")

	// at this point, we do not set CompatibleVersion to an earlier version number
	// in the future we may use this field to allow older versions of the solution to load (but not migrate)
	// ... newer domains than existed when the older solution version was released, but this will require
	// ... intentional planning and testing
	featureSetVersion := feature.SimpleVersion(options.FeatureSetVersion)
	features, err := feature.InitFeatureSetVersion(featureSetVersion)
	if err != nil {
		return err
	}

	err = c.Config.CreateDomainConfig(LcDomainConfig{
		Name:                 name,
		AiIdResolutionBucket: matchS3Name,
		EnvName:              tags[DOMAIN_CONFIG_ENV_KEY],
		Options:              options,
		CustomerProfileMode:  customerProfileMode,
		DynamoMode:           dynamoMode,
		QueueUrl:             queueUrl,
		DomainVersion:        featureSetVersion.Version,
		CompatibleVersion:    featureSetVersion.CompatibleVersion,
	})
	if err != nil {
		return err
	}

	// the requested version may affect the remainder of the process of creating the domain
	// set it now. it is also set later by SetDomain
	c.features = features

	return nil
}

func (c *CustomerProfileConfigLowCost) createProfileCaches(
	name string,
	tags map[string]string,
	matchS3Name string,
	customerProfileMode, dynamoMode bool,
	cpDomainParams CpDomainParams,
) error {
	tx := c.Tx
	tx.Info("[CreateDomainWithQueue] 2. Creating profile caches")
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs error
	if customerProfileMode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := c.CreateCPCache(name, tags, matchS3Name, cpDomainParams)
			if err != nil {
				mu.Lock()
				errs = errors.Join(err, errs)
				mu.Unlock()
			}
		}()
	}
	if dynamoMode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := c.CreateDynamoCache(name, tags)
			if err != nil {
				mu.Lock()
				errs = errors.Join(err, errs)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return errs
}

func (c *CustomerProfileConfigLowCost) createAndInitializeProfileCountTable(name string) error {
	tx := c.Tx
	tx.Info("[CreateDomainWithQueue] 4. Creating and initializing profile count table")
	tx.Debug("[CreateDomainWithQueue]4.1 Creating profile count table")
	if err := c.Data.CreateRecordCountTable(name); err != nil {
		tx.Error("[TestInsertQuery] error creating profile count table %v", err)
		return err
	}
	tx.Debug("[CreateDomainWithQueue] 4.2 Initializing profile count")
	if err := c.Data.InitializeCount(name, []string{countTableProfileObject}); err != nil {
		tx.Error("[CreateDomainWithQueue] error initializing profile count %v", err)
		return err
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) cleanupDomainData(name string) {
	c.Tx.Error("Error creating domain data, cleaning up...")
	if err := c.DeleteDomainByName(name); err != nil {
		c.Tx.Error("Error deleting domain on create rollback")
	}
}

func (c *CustomerProfileConfigLowCost) ListDomains() ([]Domain, error) {
	c.Tx.Info("Listing domains")
	domainConfigs, err := c.Config.ListDomainConfigs()
	if err != nil {
		c.Tx.Error("Error while listing domains: %v", err)
		return []Domain{}, err
	}

	domains := []Domain{}
	for _, cfg := range domainConfigs {
		domains = append(domains, domainCfgToDomain(cfg))
	}

	return domains, nil
}

func (c *CustomerProfileConfigLowCost) DeleteDomain() error {
	c.Tx.Info("Deleting domain %v", c.DomainName)
	if !c.isDomainSet() {
		return fmt.Errorf("[DeleteDomain] domain is not set")
	}
	err := c.DeleteDomainByName(c.DomainName)
	if err != nil {
		return err
	}
	c.DomainName = ""
	c.DomainConfig = LcDomainConfig{}
	c.Data.SetDynamoProfileTableProperties("", "", "")
	return nil
}

// retrieve profile by it's profile_id (customer known) returns an error if not found
func (c *CustomerProfileConfigLowCost) GetProfileByProfileId(
	profileId string,
	objectTypeNames []string,
	pagination []PaginationOptions,
) (profilemodel.Profile, error) {
	profiles, err := c.SearchProfilesByProfileID(profileId)
	if err != nil {
		return profilemodel.Profile{}, err
	}
	if len(profiles) != 1 {
		return profilemodel.Profile{}, fmt.Errorf("no profile found for profile_id=%s", profileId)
	}
	return c.GetProfile(profiles[0].ProfileId, objectTypeNames, pagination)
}

// retrieve profile by its upt_id
func (c *CustomerProfileConfigLowCost) GetProfile(
	connectID string,
	objectTypeNames []string,
	pagination []PaginationOptions,
) (profilemodel.Profile, error) {
	allMappings, err := c.GetMappings()
	if err != nil {
		c.Tx.Error("[GetProfile] Error while fetching mappings: %v", err)
		return profilemodel.Profile{}, err
	}
	var profileLevelRow map[string]interface{}
	connectID, profileLevelRow, err = c.Data.GetProfileSearchRecord(c.DomainName, connectID)
	if err != nil {
		c.Tx.Error("[GetProfile] Error while profile record from master table: %v", err)
		return profilemodel.Profile{}, err
	}
	if len(profileLevelRow) == 0 {
		c.Tx.Warn("[GetProfile] No profile record found for connectID %s", connectID)
		return profilemodel.Profile{}, ErrProfileNotFound
	}
	if len(objectTypeNames) == 0 {
		c.Tx.Warn("[GetProfile] No objects requested, returning profile level data")
		profile, _, _ := c.Data.utils.auroraToProfile(profileLevelRow)
		return profile, nil
	}
	conn, err := c.Data.AurSvc.AcquireConnection(c.Tx)
	if err != nil {
		c.Tx.Error("[GetProfile] Error while acquiring connection: %v", err)
		return profilemodel.Profile{}, err
	}
	defer conn.Release()

	objectTypeNameToId := make(map[string]string, len(objectTypeNames))
	for _, objectType := range objectTypeNames {
		objectTypeNameToId[objectType] = ""
	}
	// we don't need object level priority here since we pass the profile level row from the search table
	prof, _, _, err := c.Data.buildProfile(
		conn,
		c.DomainName,
		connectID,
		objectTypeNameToId,
		allMappings,
		profileLevelRow,
		[]string{},
		pagination,
	)
	if err != nil {
		c.Tx.Error("[GetProfile] Error while building profile: %v", err)
		return profilemodel.Profile{}, err
	}
	return prof, nil
}

func (c *CustomerProfileConfigLowCost) SearchProfilesByProfileID(profileID string) ([]profilemodel.Profile, error) {
	c.Tx.Info("Searching profiles by profile id %s", profileID)
	profiles := []profilemodel.Profile{}
	rows, err := c.Data.SearchProfilesByProfileID(c.DomainName, profileID)
	if err != nil {
		c.Tx.Error("Error searching for profiles by ID: %v", err)
		return profiles, err
	}
	return c.Data.utils.auroraToProfiles(rows), nil
}

func (c *CustomerProfileConfigLowCost) SearchProfilesByObjectLevelField(
	objectTypeName, key string,
	value string,
) ([]profilemodel.Profile, error) {
	c.Tx.Info("Searching profiles by object type %s and key %s", objectTypeName, key)
	profiles := []profilemodel.Profile{}
	rows, err := c.Data.SearchProfilesByObjectLevelField(c.DomainName, objectTypeName, key, value)
	if err != nil {
		c.Tx.Error("[SearchProfilesByObjectLevelField] Error searching for profiles: %v", err)
		return profiles, err
	}
	return c.Data.utils.auroraToProfiles(rows), nil
}

// this function returns profile from index search table
func (c *CustomerProfileConfigLowCost) SearchProfilesByProfileLevelField(key string, value string) ([]profilemodel.Profile, error) {
	profiles := []profilemodel.Profile{}
	rows, err := c.Data.SearchProfiles(c.DomainName, key, value)
	if err != nil {
		c.Tx.Error("[SearchProfilesByProfileLevelField] Error searching for profiles: %v", err)
		return profiles, err
	}
	return c.Data.utils.auroraToProfiles(rows), nil
}

func (c *CustomerProfileConfigLowCost) SearchProfiles(key string, values []string) ([]profilemodel.Profile, error) {
	c.Tx.Info("Searching profiles by key %s and values %v", key, values)
	if len(values) == 0 {
		return []profilemodel.Profile{}, fmt.Errorf("no values provided")
	}
	if len(values) > 1 {
		c.Tx.Error("Search with multiple values for a given field not supported at the moment")
		return []profilemodel.Profile{}, fmt.Errorf("search with multiple values for a given field not supported at the moment")
	}
	allMappings, err := c.GetMappings()
	if err != nil {
		c.Tx.Error("Error while fetching mappings: %v", err)
		return []profilemodel.Profile{}, err
	}
	objectTypeName, err := c.objectTypeNameFromSearchKey(allMappings, key)
	if err != nil {
		c.Tx.Error("[SearchProfiles] Invalid search key: %v", err)
		return []profilemodel.Profile{}, err
	}

	if key == LC_PROFILE_ID_KEY {
		c.Tx.Info("Searching profiles by profile ID")
		return c.SearchProfilesByProfileID(values[0])
	}
	if objectTypeName == PROFILE_OBJECT_TYPE_PREFIX {
		c.Tx.Info("Searching profiles by profile level field")
		return c.SearchProfilesByProfileLevelField(key, values[0])
	}
	c.Tx.Info("Searching profiles by object level field")
	//removing object type prefix from the key
	objectSearchKey := strings.Replace(key, objectTypeName+".", "", 1)
	return c.SearchProfilesByObjectLevelField(objectTypeName, objectSearchKey, values[0])
}

func (c *CustomerProfileConfigLowCost) AdvancedProfileSearch(searchCriteria []BaseCriteria) ([]profilemodel.Profile, error) {
	c.Tx.Info("Advanced profile search")
	profiles := []profilemodel.Profile{}
	if len(searchCriteria) == 0 {
		c.Tx.Error("No search criteria provided")
		return profiles, nil
	}

	rows, err := c.Data.AdvancedSearchProfilesByProfileLevelField(c.DomainName, searchCriteria)
	if err != nil {
		c.Tx.Error("[AdvancedProfileSearch] Error searching for profiles: %v", err)
		return profiles, err
	}
	return c.Data.utils.auroraToProfiles(rows), nil
}

func (c *CustomerProfileConfigLowCost) SearchOnMultipleKeysProfile(sortKeyValueMapping map[string][]string) ([]string, error) {
	if len(sortKeyValueMapping) == 0 {
		return []string{}, errors.New("at least one search key should be provided provided")
	}
	var connectIDs []string
	mappingsByObjetTypeName := map[string]ObjectMapping{}
	var err error
	c.Tx.Debug("1.Fetching mappings")
	mappings, err := c.GetMappings()
	if err != nil {
		c.Tx.Error("Error while fetching mappings: %v", err)
		return []string{}, err
	}
	c.Tx.Debug("Organizing mappings by object type name")
	for _, mapping := range mappings {
		mappingsByObjetTypeName[mapping.Name] = mapping
	}

	columnArray := []string{}
	objectTypeArray := []string{}
	valuesArray := [][]string{}

	for sortKey, values := range sortKeyValueMapping {
		objectTypeName, key, err := c.processSearchKey(sortKey, mappingsByObjetTypeName)
		if err != nil {
			c.Tx.Error("[SearchOnMultipleKeysProfile] Invalid search key: %v", err)
			return []string{}, err
		}
		columnArray = append(columnArray, key)
		objectTypeArray = append(objectTypeArray, objectTypeName)
		valuesArray = append(valuesArray, values)
	}
	cids, err := c.Data.FindConnectIdsMultiSearch(c.DomainName, objectTypeArray, columnArray, valuesArray)
	if err != nil {
		c.Tx.Error("Error while fetching records: %v", err)
		return []string{}, err
	}
	connectIDs = append(connectIDs, cids...)
	return connectIDs, nil
}

// return hte object type name from the search key
func (c *CustomerProfileConfigLowCost) objectTypeNameFromSearchKey(mappings []ObjectMapping, key string) (string, error) {
	mappingsByObjetTypeName := map[string]ObjectMapping{}
	c.Tx.Debug("Organizing mappings by object type name")
	for _, mapping := range mappings {
		mappingsByObjetTypeName[mapping.Name] = mapping
	}

	//we use the mappings to identify where to locate the profile (customers can search using objectTypeName.key or key (if mapped at profile level))
	objectTypeName, _, err := c.processSearchKey(key, mappingsByObjetTypeName)
	if err != nil {
		c.Tx.Error("[objectTypeNameFromSearchKey] Invalid search key: %v", err)
		return "", err
	}
	return objectTypeName, err
}

func (c *CustomerProfileConfigLowCost) processSearchKey(key string, objectMappings map[string]ObjectMapping) (string, string, error) {
	c.Tx.Debug("Processing search key %s", key)
	c.Tx.Debug("Search key contains a object type name")
	segments := strings.Split(key, ".")
	if len(segments) > 3 {
		return "", key, fmt.Errorf("invalid search key %s. The key should contain at most 2 '.' separator", key)
	}
	if core.ContainsString(RESERVED_FIELDS, segments[0]) {
		c.Tx.Error("Reserved field %v used. must be a profile level attributes")
		if len(segments) == 2 {
			if !core.ContainsString([]string{"Address", "ShippingAddress", "MailingAddress", "BillingAddress", "Attributes"}, segments[0]) {
				return "", key, fmt.Errorf("invalid search key  '%v' Only addresses and custom attributes can be nested", key)
			}
		}
		return PROFILE_OBJECT_TYPE_PREFIX, key, nil
	}

	if len(segments) == 1 {
		c.Tx.Error("Search key does not start with a object type name nor with a reserve field")
		return "", key, fmt.Errorf(
			"invalid search key '%v' Search field should start with either a reserve field (profile level search) or a valid object type name followed by a '.' and a field name",
			key,
		)
	}
	c.Tx.Debug("Non-Reserved field %v used. must be an object level attributes. Validating against mappings")
	objectTypeName := segments[0]
	newKey := segments[1]
	objectIsValid := false
	keyIsValid := false
	for objectType, mappings := range objectMappings {
		if objectType == objectTypeName {
			objectIsValid = true
			for _, fieldMapping := range mappings.Fields {
				if fieldMapping.KeyOnly || fieldMapping.Searcheable {
					expectedName := strings.ReplaceAll(fieldMapping.Source, "_source.", "")
					if expectedName == newKey {
						keyIsValid = true
						break
					}
				}
			}
			//object found but no key
			break
		}
	}
	if !objectIsValid {
		return "", key, fmt.Errorf("invalid object type name in search key %s (Not present in mappings)", key)
	}
	if !keyIsValid {
		return "", key, fmt.Errorf("invalid search key %s (not present in mappings)", key)
	}
	return objectTypeName, newKey, nil
}

func (c *CustomerProfileConfigLowCost) DeleteProfile(id string, options ...DeleteProfileOptions) error {
	tx := core.NewTransaction(c.Tx.TransactionID, "", c.Tx.LogLevel)
	deleteHistory := false
	operatorId := ""
	deleteMsg := "[DeleteProfile] Deleting Profile: %s"
	for _, o := range options {
		if o.DeleteHistory {
			deleteHistory = true
			deleteMsg = deleteMsg + " and purging all history"
		}
		if o.OperatorID != "" {
			operatorId = o.OperatorID
		}
	}
	tx.Debug(deleteMsg, id)
	tx.Debug("[DeleteProfile] deleteHistory is set to: %b", deleteHistory)

	cfg := c.DomainConfig

	//	Delete from Interactions
	mappings, err := c.Config.ListMappings(c.DomainName)
	if err != nil {
		tx.Error("[DeleteProfile] Error getting mappings: %v", err)
		return err
	}

	conn, err := c.Data.AurSvc.AcquireConnection(tx)
	if err != nil {
		tx.Error("[DeleteProfile] Error acquiring connection from pool: %v", err)
		return err
	}
	defer conn.Release()

	connectIds, err := c.Data.getIdsForOriginalConnectIds(conn, c.DomainName, "connect_id", []string{id})
	if err != nil || len(connectIds) > 1 {
		conn.Tx.Error("Error finding profile: %v", err)
		return fmt.Errorf("error finding profile: %s", err)
	}
	if len(connectIds) == 0 {
		conn.Tx.Info("Most recent connect_id not found or profile already deleted, skipping delete")
		return nil
	}
	id = connectIds[0]

	tx.Debug("[DeleteProfile] Starting transaction")
	err = conn.StartTransaction()
	if err != nil {
		tx.Error("Error starting transaction: %v", err)
		return err
	}

	//	Delete from Interaction and Optionally History table
	for _, mapping := range mappings {
		err = c.Data.DeleteProfileObjectsFromInteractionTable(conn, c.DomainName, mapping.Name, id)
		if err != nil {
			tx.Error("[DeleteProfile] Error deleting profile from interaction table: %v; rolling back", err)
			return conn.RollbackTransaction()
		}

		if deleteHistory {
			err = c.Data.DeleteProfileObjectsFromObjectHistoryTable(conn, c.DomainName, mapping.Name, id)
			if err != nil {
				tx.Error("[DeleteProfile] Error deleting profile from object history table: %v; rolling back", err)
				conn.RollbackTransaction()
				return err
			}
		}
	}

	//	Handle Profile History Table
	err = c.Data.AddDeletionRecordToProfileHistoryTable(conn, c.DomainName, id, operatorId)
	if err != nil {
		tx.Error("[DeleteProfile] Error deleting profile from profile history table: %v; rolling back", err)
		conn.RollbackTransaction()
		return err
	}

	err = c.IR.deleteIndexes(conn, c.DomainName, []string{id})
	if err != nil {
		tx.Error("[DeleteProfile] Error deleting profile's rule indexes: %v; rolling back", err)
		conn.RollbackTransaction()
		return err
	}

	//	Delete from Profile Master Table
	err = c.Data.DeleteProfile(conn, c.DomainName, id)
	if err != nil {
		tx.Error("[DeleteProfile] Error deleting profile from master table: %v; rolling back", err)
		conn.RollbackTransaction()
		return err
	}

	_, err = conn.Query(c.Data.buildDeleteSearchRecordSql(c.DomainName), id)
	if err != nil {
		tx.Error("[DeleteProfile] Error deleting profile from search table: %v; rolling back", err)
		conn.RollbackTransaction()
		return err
	}

	tx.Debug("[DeleteProfile] Committing DBtransaction")
	err = conn.CommitTransaction()
	if err != nil {
		tx.Error("Error committing transaction: %v", err)
		return err
	}

	if cfg.DynamoMode {
		//	Delete from DynamoDB Cache
		err := c.Data.DeleteProfileFromCache(id)
		if err != nil {
			tx.Error("[DeleteProfile] Error deleting profile from cache: %v; rolling back", err)
			return err
		}
	}

	if cfg.CustomerProfileMode {
		//	Delete from Customer Profile
		cpId, err := c.CustomerProfileConfig.GetProfileId(customerprofiles.CONNECT_ID_ATTRIBUTE, id)
		if err != nil {
			tx.Error("[DeleteProfile] Error getting profile from Customer Profile: %v; rolling back", err)
			return err
		}
		err = c.CustomerProfileConfig.DeleteProfile(cpId)
		if err != nil {
			tx.Error("[DeleteProfile] Error deleting profile from Customer Profile: %v; rolling back", err)
			return err
		}
	}

	return nil
}

func (c *CustomerProfileConfigLowCost) GetDomain() (Domain, error) {
	if c.DomainConfig.Name == "" {
		c.Tx.Error("[GetDomain] domain not found")
		return Domain{}, ErrDomainNotFound
	}
	domain := domainCfgToDomain(c.DomainConfig)
	counts, err := c.Data.GetAllObjectCounts(c.DomainName)
	if err != nil {
		c.Tx.Error("[GetDomain] Error getting domain object counts: %v", err)
		return Domain{}, err
	}
	domain.AllCounts = counts
	domain.NProfiles = counts[countTableProfileObject]
	for obj, ct := range counts {
		if obj != countTableProfileObject {
			domain.NObjects = domain.NObjects + ct
		}
	}
	return domain, nil
}

func (c *CustomerProfileConfigLowCost) GetMappings() ([]ObjectMapping, error) {
	mappings, err := c.Config.ListMappings(c.DomainName)
	if IsRetryableError(err) {
		c.Tx.Debug("[GetMappings] Retryable error found, retrying ListMappings")
		mappings, err = c.Config.ListMappings(c.DomainName)
	}
	return mappings, err
}

func (c *CustomerProfileConfigLowCost) GetMappingsByName(domainName string) ([]ObjectMapping, error) {
	return c.Config.ListMappings(domainName)
}

func (c *CustomerProfileConfigLowCost) GetProfileLevelFields() ([]string, error) {
	return c.Config.GetMasterColumns(c.DomainName)
}

func (c *CustomerProfileConfigLowCost) GetObjectLevelFields() (map[string][]string, error) {
	if c.DomainName == "" {
		return nil, nil
	}
	return c.Config.GetObjectLevelColumns(c.DomainName)
}

func (c *CustomerProfileConfigLowCost) GetMapping(objType string) (ObjectMapping, error) {
	mappings, err := c.GetMappings()
	if err != nil {
		return ObjectMapping{}, err
	}
	for _, mapping := range mappings {
		if mapping.Name == objType {
			return mapping, nil
		}
	}
	return ObjectMapping{}, fmt.Errorf("mapping %s not found", objType)
}

func (c *CustomerProfileConfigLowCost) DeleteDomainByName(name string) error {
	c.Tx.Info("[DeleteDomainByName] Deleting domain %v", name)
	c.Tx.Debug("[DeleteDomainByName] 0. Retrieving Domain Config for domain: %v", name)
	domCfg, exists := c.Config.DomainConfigExists(name)
	if !exists {
		c.Tx.Error("[DeleteDomainByName] Domain %v does not exist", name)
		return fmt.Errorf("Domain %v does not exist", name)
	}
	c.Tx.Debug("[DeleteDomainByName] 0. Deleting mapping records")
	var errs error
	mappings, err := c.GetMappingsByName(name)
	if err != nil {
		c.Tx.Error("[DeleteDomainByName] Error while getting mappings: %v. continuing domain deletion", err)
		errs = errors.Join(errs, err)
	}
	for _, mapping := range mappings {
		err = c.Config.DeleteMapping(name, mapping.Name)
		if err != nil {
			c.Tx.Error("[DeleteDomainByName] Error while deleting mapping: %v. continuing domain deletion", err)
		}
		err = c.Data.DeleteInteractionTable(name, mapping.Name)
		if err != nil {
			c.Tx.Error("[DeleteDomainByName] Error while deleting interaction table: %v. continuing domain deletion", err)
		}
		err = c.Data.DeleteObjectHistoryTable(name, mapping.Name)
		if err != nil {
			c.Tx.Error("[DeleteDomainByName] Error while deleting object history table: %v. continuing domain deletion", err)
			errs = errors.Join(errs, err)
		}
	}
	c.Tx.Info("[DeleteDomainByName] 1. Deleting profile table")
	err = c.DeleteDomainProfileStore(name, domCfg.CustomerProfileMode, domCfg.DynamoMode)
	if err != nil {
		c.Tx.Error("[DeleteDomainByName] Error while deleting profile table: %v. continuing domain deletion", err)
		errs = errors.Join(errs, err)
	}
	c.Tx.Info("[DeleteDomainByName] 2. Deleting profile count table")
	err = c.Data.DeleteRecordCountTable(name)
	if err != nil {
		c.Tx.Error("[DeleteDomainByName] error deleting profile count table %v. continuing domain deletion", err)
	}
	c.Tx.Info("[DeleteDomainByName] 3. Deleting profile history table")
	err = c.Data.DeleteProfileHistoryTable(name)
	if err != nil {
		c.Tx.Error("Error while deleting profile history table: %v. continuing domain deletion", err)
		errs = errors.Join(errs, err)
	}
	c.Tx.Info("[DeleteDomainByName] 4. Deleting profile master table")
	err = c.Data.DeleteProfileMasterTable(name)
	if err != nil {
		c.Tx.Error("[DeleteDomainByName] Error while deleting profile table: %v. continuing domain deletion", err)
		errs = errors.Join(errs, err)
	}
	err = c.Data.DeleteProfileSearchIndexTable(name)
	if err != nil {
		c.Tx.Error("[DeleteDomainByName] Error while deleting profile search table: %v. continuing domain deletion", err)
		errs = errors.Join(errs, err)
	}
	c.Tx.Info("[DeleteDomainByName] 5. Deleting rule sets")
	err = c.Config.DeleteAllRuleSetsByDomain(name)
	if err != nil {
		c.Tx.Error("[DeleteDomainByName] Error while deleting rule sets for domain %s: %v. continuing domain deletion", name, err)
		errs = errors.Join(errs, err)
	}
	c.Tx.Info("[DeleteDomainByName] 6. Deleting config record")
	err = c.Config.DeleteDomainConfig(name)
	if err != nil {
		c.Tx.Error("Error while deleting domain config: %v. continuing domain deletion", err)
		errs = errors.Join(errs, err)
	}
	c.Tx.Info("[DeleteDomainByName] 7. Deleting rule base index")
	err = c.IR.DeleteIRTable(name)
	if err != nil {
		c.Tx.Error("Error while deleting rule based identity resolution table: %v. continuing domain deletion", err)
		errs = errors.Join(errs, err)
	}

	return errs
}

func (c *CustomerProfileConfigLowCost) MergeProfiles(
	connectId string,
	connectIdToMerge string,
	mergeContext ProfileMergeContext,
) (string, error) {
	mergeID := core.GenerateUniqueId()
	c.Tx.Info(
		"[MergeProfiles][%s] merging profile %s into %s with merge context: %+v and merge ID %s",
		mergeID,
		connectIdToMerge,
		connectId,
		mergeContext,
		mergeID,
	)
	c.Tx.Debug("[MergeProfiles][%s] 1. Getting control plane data", mergeID)
	startMergeProfile := time.Now()
	allObjectTypeNames := []string{}
	allMappings, err := c.GetMappings()
	if err != nil {
		c.Tx.Error("[MergeProfiles] Error while getting mapping: %v", err)
		return "FAILURE", err
	}
	for _, mapping := range allMappings {
		allObjectTypeNames = append(allObjectTypeNames, mapping.Name)
	}
	domainConfig, err := c.Config.GetDomainConfig(c.DomainName)
	if err != nil {
		c.Tx.Error("[MergeProfiles] Error while getting domain config: %v", err)
		return "FAILURE", err
	}
	startUpdateAurora := time.Now()

	c.Tx.Debug("[MergeProfiles][%s] Retrieving connectIdToMerge and validate it exists (for idempotency)", mergeID)
	var row map[string]interface{}
	connectIdToMerge, row, err = c.Data.GetProfileSearchRecord(c.DomainName, connectIdToMerge)
	if len(row) == 0 {
		c.Tx.Error("[MergeProfiles]Target profile does not exist and has not been merged before. Returning failure")
		return "FAILURE", fmt.Errorf("target profile %s does not exist", connectIdToMerge)
	}
	if err != nil {
		c.Tx.Error("[MergeProfiles] Error while getting profile search record: %v", err)
		return "FAILURE", err
	}
	if connectId == connectIdToMerge {
		c.Tx.Debug("[MergeProfiles] Profiles have been merged before, returning success to ensure idempotency")
		return "SUCCESS", nil
	}
	c.Tx.Debug("[MergeProfiles][%s] target profile exists!", mergeID)

	c.Tx.Debug("[MergeProfiles][%s] merging profiles in aurora", mergeID)
	err = c.Data.MergeProfiles(c.DomainName, connectId, connectIdToMerge, allObjectTypeNames, mergeContext, allMappings, mergeID)
	if err != nil {
		c.Tx.Error("[MergeProfiles] Error while merging profiles: %v. checking if profiles have already been merged", err)
		//we recheck for idempotency incase the profile has been merged bweetn our earlier check and now
		haveBeenMerge, err := c.HaveProfilesAlreadyBeenMerged(connectId, connectIdToMerge)
		if err != nil {
			c.Tx.Error("[MergeProfiles] Error while checking if profiles have already been merged: %v", err)
			return "FAILURE", err
		}
		if haveBeenMerge {
			c.Tx.Debug("[MergeProfiles] Profiles have been merged before, returning success to ensure idempotency")
			return "SUCCESS", nil
		}
		c.Tx.Debug("profiles have not been merged before. retrying (connection drop case)")
		err = c.Data.MergeProfiles(c.DomainName, connectId, connectIdToMerge, allObjectTypeNames, mergeContext, allMappings, mergeID)
		if err != nil {
			c.Tx.Error("Error while merging profiles (after retry): %v", err)
			return "FAILURE", fmt.Errorf("error while merging profiles: %w", err)
		}
	}
	c.LogLatencyMs(METRIC_NAME_MERGE_PROFILES_UPDATE_AURORA, float64(time.Since(startUpdateAurora).Milliseconds()))
	c.Tx.Info("[MergeProfiles][%s] successfully merged in aurora", mergeID)

	//getting row post-merge
	_, newRow, err := c.Data.GetProfileSearchRecord(c.DomainName, connectId)
	if err != nil {
		c.Tx.Error("[MergeProfiles] Error while getting profile search record: %v", err)
		return "FAILURE", err
	}

	//we don't pass the profile level row (even if it has been updated from Data.MergeProfiles)
	activeRuleSet, err := c.Config.getRuleSet(c.DomainName, RULE_SET_TYPE_ACTIVE, RULE_SK_PREFIX)
	if err != nil {
		return "FAILURE", err
	}

	//Creating profile records for both rule-base ID res and cache population
	conn, err := c.Data.AurSvc.AcquireConnection(c.Tx)
	if err != nil {
		c.Tx.Error("[MergeProfiles] Error while acquiring connection to run buildProfile: %v", err)
		return "FAILURE", err
	}
	defer conn.Release()
	objectTypeNameToId := make(map[string]string, len(allObjectTypeNames))
	for _, objectType := range allObjectTypeNames {
		objectTypeNameToId[objectType] = ""
	}
	combinedProfile, _, _, err := c.Data.buildProfile(
		conn,
		c.DomainName,
		connectId,
		objectTypeNameToId,
		allMappings,
		newRow,
		domainConfig.Options.ObjectTypePriority,
		[]PaginationOptions{},
	)
	if err != nil {
		return "FAILURE", err
	}

	var isMergeCandidate bool
	if len(activeRuleSet.Rules) > 0 {
		startReindexAfterMerge := time.Now()
		c.Tx.Debug("[MergeProfiles][%s] retreiving merged profile", mergeID)
		conn, err := c.Data.AurSvc.AcquireConnection(c.Tx)
		if err != nil {
			c.Tx.Error("Error while acquiring connection: %v", err)
			return "FAILURE", err
		}
		defer conn.Release()

		// Delete all rule indexes for old profiles that are now combined
		err = c.IR.ReindexProfileAfterMerge(c.DomainName, connectId, connectIdToMerge, map[string]string{}, combinedProfile, activeRuleSet)
		if err != nil {
			c.Tx.Error("Error while reindexing profile after merge: %v", err)
			return "FAILURE", err
		}
		c.LogLatencyMs(METRIC_NAME_MERGE_PROFILES_REINDEX, float64(time.Since(startReindexAfterMerge).Milliseconds()))

		startReindexAfterMergeRematch := time.Now()
		//check if the profile matches
		profileLevelRules, _ := c.IR.organizeRules(activeRuleSet.Rules)
		profileLevelRulesByIndex := map[string]Rule{}
		for _, rule := range profileLevelRules {
			profileLevelRulesByIndex[strconv.Itoa(rule.Index)] = rule
		}
		matches, err := c.IR.FindMatches(conn, c.DomainName, PROFILE_OBJECT_TYPE_NAME, map[string]string{}, combinedProfile, activeRuleSet)
		if err != nil {
			c.Tx.Error("Error while finding matches: %v", err)
			return "FAILURE", err
		}
		for _, match := range matches {
			if _, ok := profileLevelRulesByIndex[match.RuleIndex]; ok {
				isMergeCandidate = true
				err = c.sendRuleBasedMatches(matches, activeRuleSet)
				if err != nil {
					c.Tx.Error("Error while sending rule based matches: %v", err)
					return "FAILURE", err
				}
				break
			}
		}
		c.LogLatencyMs(METRIC_NAME_MERGE_PROFILES_REINDEX_REMATCH, float64(time.Since(startReindexAfterMergeRematch).Milliseconds()))
	}

	// Update low latency store(s) with new profile
	if !isMergeCandidate {
		// skip if we are sending profile to be merged again, since it will be updated momentarily
		var cacheMode CacheMode
		if domainConfig.CustomerProfileMode {
			cacheMode = cacheMode | CUSTOMER_PROFILES_MODE
		}
		if domainConfig.DynamoMode {
			cacheMode = cacheMode | DYNAMO_MODE
		}
		startUpdateDynamo := time.Now()
		err = c.UpdateDownstreamCache(combinedProfile, "", cacheMode, combinedProfile.ProfileObjects)
		if err != nil {
			return "FAILURE", err
		}
		c.LogLatencyMs(METRIC_NAME_MERGE_PROFILES_UPDATE_DYNAMO, float64(time.Since(startUpdateDynamo).Milliseconds()))
	}

	// Merged profile no longer exists, deleting from cache
	startDelete := time.Now()
	if domainConfig.CustomerProfileMode {
		err = c.DeleteProfileFromCustomerProfileCache(connectIdToMerge)
		if err != nil {
			c.Tx.Error("Error while deleting profile: %v", err)
			return "FAILURE", err
		}
	}
	if domainConfig.DynamoMode {
		err = c.Data.DeleteProfileFromCache(connectIdToMerge)
		if err != nil {
			return "FAILURE", err
		}
	}
	c.LogLatencyMs(METRIC_NAME_MERGE_PROFILES_DELETE_PROFILE_DYNAMO, float64(time.Since(startDelete).Milliseconds()))

	// Create merge event
	startSendEvents := time.Now()
	eventId := uuid.NewString()
	c.Tx.Debug("Sending change events: %s", eventId)
	mergeEventParams := ChangeEventParams{
		EventID:    eventId,
		EventType:  EventTypeMerged,
		Time:       startSendEvents,
		DomainName: c.DomainName,
		Profile:    combinedProfile,
		IsRealTime: true,
		MergeID:    eventId,
		MergedProfiles: []profilemodel.Profile{
			{ProfileId: connectIdToMerge},
		},
	}
	// Create delete event
	deleteEventParams := ChangeEventParams{
		EventID:    eventId,
		EventType:  EventTypeDeleted,
		Time:       startSendEvents,
		DomainName: c.DomainName,
		Profile: profilemodel.Profile{
			ProfileId: connectIdToMerge,
		},
		IsRealTime: true,
	}
	// Send merge and delete events to Kinesis
	err = SendChangeEvents(c.KinesisClient, mergeEventParams, deleteEventParams)
	if err != nil {
		c.Tx.Error("[MergeProfiles] Error sending change events to Kinesis %v", err)
	}
	c.LogLatencyMs(METRIC_NAME_SEND_CHANGE_EVENT, float64(time.Since(startSendEvents).Milliseconds()))

	c.LogLatencyMs(METRIC_NAME_MERGE_PROFILES, float64(time.Since(startMergeProfile).Milliseconds()))
	return "SUCCESS", nil
}

func (c *CustomerProfileConfigLowCost) HaveProfilesAlreadyBeenMerged(remainingProfile, mergedProfile string) (bool, error) {
	return c.Data.HaveBeenMergedAndNotUnmerged(c.DomainName, remainingProfile, mergedProfile)
}

func (c *CustomerProfileConfigLowCost) MergeMany(pairs []ProfilePair) (string, error) {
	c.Tx.Info("[MergeMany] Merging %d profile pairs", len(pairs))

	profilePairsBySourceID := organizeProfilePairsBySourceID(pairs) // what is the value of this?
	errs := []error{}
	inputs := []ProfilePair{}
	for _, input := range profilePairsBySourceID {
		inputs = append(inputs, input...)
	}
	batchSize := 10
	batchInputs := core.Chunk(core.InterfaceSlice(inputs), batchSize)
	for _, inputs := range batchInputs {
		var wg sync.WaitGroup
		mu := &sync.Mutex{}
		wg.Add(len(inputs))
		for _, input := range inputs {
			go func(mergeInput ProfilePair) {
				c.Tx.Debug("[MergeProfiles] ACCP Merge request: %+v", mergeInput)
				_, err := c.MergeProfiles(mergeInput.SourceID, mergeInput.TargetID, mergeInput.MergeContext)
				if err != nil {
					c.Tx.Error("[MergeProfiles] Error merging profiles: %v", err)
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
				wg.Done()
			}(input.(ProfilePair))
		}
		wg.Wait()
	}

	status := "SUCCESS"
	if len(errs) > 0 {
		c.Tx.Error("[MergeMany] FAILURE one or multiple merges failed: %v", errs)
		status = "FAILURE"
	}
	c.Tx.Info("[MergeMany] Merge profile complete with status %v", status)
	return status, core.ErrorsToError(errs)
}

func (c *CustomerProfileConfigLowCost) UnmergeProfiles(
	toUnmergeConnectID string,
	mergedIntoConnectID string,
	interactionIdToUnmerge string,
	interactionType string,
) (string, error) {
	tx := core.NewTransaction(c.Tx.TransactionID, "", c.Tx.LogLevel)
	tx.Info("[UnmergeProfiles] Starting unmerge operation")
	startUnmergeTime := time.Now()
	domainConfig, err := c.Config.GetDomainConfig(c.DomainName)
	if err != nil {
		tx.Error("[UnmergeProfiles] Error while getting domain config: %v", err)
		return "FAILURE", err
	}
	allObjectTypeNames := []string{}
	allMappings, err := c.GetMappings()
	if err != nil {
		tx.Error("[UnmergeProfiles] error while getting mapping: %v", err)
		return "FAILURE", err
	}
	for _, mapping := range allMappings {
		allObjectTypeNames = append(allObjectTypeNames, mapping.Name)
	}

	if toUnmergeConnectID == "" {
		if interactionIdToUnmerge == "" {
			return "FAILURE", fmt.Errorf("[UnmergeProfiles] unable to perform unmerge with the information given")
		}
		conn, err := c.Data.AurSvc.AcquireConnection(tx)
		if err != nil {
			tx.Error("[UnmergeProfiles] Error acquiring connection from pool: %v", err)
			return "FAILURE", err
		}
		defer conn.Release()
		// For the interaction to unmerge, find the profile it was originally ingested to
		toUnmergeConnectID, _, err = c.Data.getOriginalConnectIdForInteraction(
			conn,
			c.DomainName,
			interactionIdToUnmerge,
			interactionType,
			"",
		)
		if err != nil {
			tx.Error("[UnmergeProfiles] unable to find the original profile for the requested interaction %s", err)
			return "FAILURE", fmt.Errorf("[UnmergeProfiles] unable to find the original profile for the requested interaction %s", err)
		}
	}

	startUpdateAurora := time.Now()
	// Update aurora tables for unmerge

	// Rollback strategy
	// - Any changes made to aurora (interaction store) should propagate to dynamo (profile store)
	// - Incase updating dynamo fails and aurora succeeds, they will be out of sync
	//   and we should rollback changes made to aurora

	err = c.Data.UnmergeProfiles(
		c.DomainName,
		mergedIntoConnectID,
		toUnmergeConnectID,
		allObjectTypeNames,
		allMappings,
		domainConfig.Options.ObjectTypePriority,
	)
	if err != nil {
		return "FAILURE", fmt.Errorf("[UnmergeProfiles] error unmerging profiles %s", err)
	}
	c.LogLatencyMs(METRIC_NAME_UNMERGE_PROFILES_UPDATE_AURORA, float64(time.Since(startUpdateAurora).Milliseconds()))

	startUpdateDynamo := time.Now()

	// Get parent profile
	conn, err := c.Data.AurSvc.AcquireConnection(c.Tx)
	if err != nil {
		c.Tx.Error("[UnmergeProfiles] Error acquiring connection to run buildProfile: %v", err)
		return "FAILURE", err
	}
	defer conn.Release()
	objectTypeNameToId := make(map[string]string, len(allObjectTypeNames))
	for _, objectType := range allObjectTypeNames {
		objectTypeNameToId[objectType] = ""
	}
	mergedIntoProfile, _, _, err := c.Data.buildProfile(
		conn,
		c.DomainName,
		mergedIntoConnectID,
		objectTypeNameToId,
		allMappings,
		map[string]interface{}{},
		domainConfig.Options.ObjectTypePriority,
		[]PaginationOptions{},
	)
	if err != nil {
		return "FAILURE", fmt.Errorf("[UnmergeProfiles] error building parent profile, err: %s", err)
	}
	// Re-create the unmerged profile
	toUnmergeProfile, _, _, err := c.Data.buildProfile(
		conn,
		c.DomainName,
		toUnmergeConnectID,
		objectTypeNameToId,
		allMappings,
		map[string]interface{}{},
		domainConfig.Options.ObjectTypePriority,
		[]PaginationOptions{},
	)
	if err != nil {
		return "FAILURE", fmt.Errorf("[UnmergeProfiles] error building unmerged profile, err: %s", err)
	}
	res, ok := c.Config.DomainConfigExists(c.DomainName)
	if !ok {
		tx.Error("[CreateMapping] Domain name %v does not exists", c.DomainName)
		return "FAILURE", err
	}
	var cacheMode CacheMode
	if res.CustomerProfileMode {
		cacheMode = cacheMode | CUSTOMER_PROFILES_MODE
	}
	if res.DynamoMode {
		cacheMode = cacheMode | DYNAMO_MODE
	}

	// update parent profile with reduced set of interactions
	err = c.Data.DeleteProfileFromCache(mergedIntoConnectID)
	if err != nil {
		return "FAILURE", fmt.Errorf("[UnmergeProfiles] error updating parent profile %s", err)
	}
	err = c.UpdateDownstreamCache(mergedIntoProfile, "", cacheMode, mergedIntoProfile.ProfileObjects)
	if err != nil {
		c.Tx.Warn("[UnmergeProfiles] Error while updating mergedIntoProfile object in downstream cache: %v", err)
		return "FAILURE", fmt.Errorf("[UnmergeProfiles] error updating parent profile, err: %s", err)
	}
	err = c.UpdateDownstreamCache(toUnmergeProfile, "", cacheMode, toUnmergeProfile.ProfileObjects)
	if err != nil {
		c.Tx.Warn("[UnmergeProfiles] Error while updating toUnmergeProfile object in downstream cache: %v", err)
		return "FAILURE", fmt.Errorf("[UnmergeProfiles] error re-creating unmerged profile, err %s", err)
	}
	c.LogLatencyMs(METRIC_NAME_MERGE_PROFILES_UPDATE_DYNAMO, float64(time.Since(startUpdateDynamo).Milliseconds()))

	// 	Start sending to change processor
	startSendEvents := time.Now()
	eventId := uuid.NewString()
	tx.Debug("Sending change events: %s", eventId)

	unmergeEventParams := ChangeEventParams{
		EventID:        eventId,
		EventType:      EventTypeUnmerged,
		Time:           startSendEvents,
		DomainName:     c.DomainName,
		Profile:        profilemodel.Profile{ProfileId: mergedIntoConnectID},
		IsRealTime:     true,
		MergeID:        eventId,
		MergedProfiles: []profilemodel.Profile{{ProfileId: toUnmergeConnectID}},
	}
	// Send unmerge event to Kinesis
	err = SendChangeEvents(c.KinesisClient, unmergeEventParams)
	if err != nil {
		tx.Error("[UnmergeProfiles] Error sending change events to Kinesis %v", err)
	}
	c.LogLatencyMs(METRIC_NAME_SEND_CHANGE_EVENT, float64(time.Since(startSendEvents).Milliseconds()))

	c.LogLatencyMs(METRIC_NAME_UNMERGE_PROFILES, float64(time.Since(startUnmergeTime).Milliseconds()))
	return "SUCCESS", nil
}

func (c *CustomerProfileConfigLowCost) PutProfileObject(object, objectTypeName string) (err error) {
	if objectTypeName == "" {
		return fmt.Errorf("[PutProfileObject] object type name is required")
	}
	if object == "" {
		return fmt.Errorf("[PutProfileObject] object data is required")
	}
	c.MetricLogger.LogMetric(map[string]string{"domain": c.DomainName}, METRIC_PUT_OBJECT_RQ_COUNT, cloudwatch.Count, float64(1))
	tx := core.NewTransaction(c.Tx.TransactionID, "", c.Tx.LogLevel)

	//#region Prep Profile
	startPutProfile := time.Now()
	tx.Info("[PutProfileObject] Putting profile object %v", objectTypeName)
	tx.Debug("[PutProfileObject] Getting field mapping for %v", objectTypeName)
	tx.Debug("[PutProfileObject] Getting domain config")
	domainConfig, err := c.Config.GetDomainConfig(c.DomainName)
	if err != nil {
		tx.Error("[PutProfileObject] Error while getting domain config: %v", err)
		return err
	}
	allObjectTypeNames := []string{}
	mappings := ObjectMapping{}
	tx.Debug("[PutProfileObject] Getting domain mappings")
	allMappings, err := c.GetMappings()
	if err != nil {
		tx.Error("Error while getting mapping: %v", err)
		return err
	}
	mappingFound := false
	for _, mapping := range allMappings {
		allObjectTypeNames = append(allObjectTypeNames, mapping.Name)
		if mapping.Name == objectTypeName {
			mappings = mapping
			mappingFound = true
		}
	}
	if !mappingFound {
		return fmt.Errorf("no mapping found for object type %v, valid object types are %+v", objectTypeName, allObjectTypeNames)
	}

	tx.Debug("[PutProfileObject] Parsing json string %v", objectTypeName)
	objMap, err := c.parseProfileObjectString(object)
	if err != nil {
		tx.Error("Error parsing profile object: %v", err)
		return err
	}
	tx.Debug("[PutProfileObject] validating object")
	err = c.validateObject(objMap, objectTypeName, mappings)
	if err != nil {
		tx.Error("Invalid profile object of type: %v: %v", objectTypeName, err)
		return err
	}

	tx.Debug("[PutProfileObject] creating profile map")
	profMap, err := objectMapToProfileMap(objMap, mappings)
	if err != nil {
		tx.Error("Error converting object map to profile map: %v", err)
		return err
	}

	//we identify PROFILE index and set the profile_id column according ly on the object
	//the profile ID column is sed for internal indexing
	objMap = c.addProfileIdField(objMap, mappings)
	c.LogLatencyMs(METRIC_NAME_PREP_PROFILE_OBJECT, float64(time.Since(startPutProfile).Milliseconds()))
	//#endregion PrepProfile

	uniqueKey := getIndexKey(mappings, INDEX_UNIQUE)
	if uniqueKey == "" {
		tx.Error("Error while getting unique key for object type %s. something may be wrong with mappings", objectTypeName)
		return fmt.Errorf("error identifying unique key for object type %s", objectTypeName)
	}
	if objMap[uniqueKey] == "" {
		tx.Error("Error while getting unique key for object type %s. something may be wrong with mappings", objectTypeName)
		return fmt.Errorf("unique key for %s must have valid value", objectTypeName)
	}

	//#region InsertProfileObject
	startInsertProfileObject := time.Now()
	tx.Debug("[PutProfileObject] inserting into aurora")
	connectID, eventType, matches, activeRuleSet, err := c.InsertProfileObjectWithPreMatch(domainConfig, objectTypeName, uniqueKey, objMap, profMap, mappings, object)
	if err != nil {
		tx.Error("Error while inserting profile object: %v. retrying once (in case of connection drop)", err)
		c.Data.AurSvc.PrintPoolStats()
		connectID, eventType, matches, activeRuleSet, err = c.InsertProfileObjectWithPreMatch(domainConfig, objectTypeName, uniqueKey, objMap, profMap, mappings, object)
		if err != nil {
			tx.Error("Error while inserting profile object (after retry): %v. giving up", err)
			return err
		}
	}
	c.LogLatencyMs(METRIC_NAME_UPSERT_PROFILE_OBJECT, float64(time.Since(startInsertProfileObject).Milliseconds()))
	//#endregion InsertProfileObject

	//we retrieve the profile level data to avoid having to recompute the  profile from all object types
	//this can save large amount of queries to aurora.
	tx.Debug("[PutProfileObject] retrieving profile search record")
	startGetProfileSearchRecord := time.Now()
	_, rowLevelData, err := c.Data.GetProfileSearchRecord(c.DomainName, connectID)
	duration := time.Since(startGetProfileSearchRecord).Milliseconds()
	c.LogLatencyMs(METRIC_NAME_GET_PROFILE_SEARCH_RECORD, float64(duration))
	if err != nil {
		tx.Error("[PutProfileObject] Error while getting profile search record: %v", err)
		return err
	}
	tx.Debug("[PutProfileObject] creating profile ")
	// Build profile
	conn, err := c.Data.AurSvc.AcquireConnection(tx)
	if err != nil {
		tx.Warn("[PutProfileObject] Error while acquiring connection: %v", err)
		return err
	}
	defer conn.Release()
	profObj := c.Config.utils.objectMapToProfileObject(objectTypeName, objMap, uniqueKey)
	profile, _, _, err := c.Data.buildProfilePutProfileObject(
		conn,
		c.DomainName,
		connectID,
		profObj,
		allMappings,
		rowLevelData,
		[]string{},
	)
	if err != nil {
		tx.Warn("Error while building profile: %v", err)
		return err
	}
	//	index fully hydrated profile in rule_index table to ensure cascade profile level rule merges can be caught
	err = c.IR.IndexInIRTable(conn, domainConfig.Name, objectTypeName, objMap, profile, activeRuleSet)
	if err != nil {
		tx.Warn("[PutProfileObject] Error while indexing profile for rule based identity resolution")
	}

	//	Send matches to Merge Queue, if any remain
	var matchFound bool
	// matches, _ = c.IR.FindMatches(conn, domainConfig.Name, objectTypeName, objMap, profile, activeRuleSet)
	if len(activeRuleSet.Rules) > 0 && len(matches) > 0 {
		tx.Debug("[PutProfileObject] found %d rule based matches. sending to merger", len(matches))
		matchFound = true
		err = c.sendRuleBasedMatches(matches, activeRuleSet)
		if err != nil {
			c.Tx.Warn("[PutProfileObject]  Error while sending rule based matches to merger: %v", err)
		}
	} else {
		tx.Info("[PutProfileObject] No active rule set, skipping rule based id resolution")
	}

	objectId := objMap[uniqueKey]

	// Update profile in DynamoDB/Customer Profiles by default. However, we can skip if the profile
	// is expected to be merged, since the profile will be immediately updated again after the merge.
	if !matchFound {
		tx.Debug("[PutProfileObject] Updating profile cache")
		var cacheMode CacheMode
		if domainConfig.CustomerProfileMode {
			cacheMode = cacheMode | CUSTOMER_PROFILES_MODE
		}
		if domainConfig.DynamoMode {
			cacheMode = cacheMode | DYNAMO_MODE
		}
		startUpdateDynamo := time.Now()
		err = c.UpdateDownstreamCache(profile, objectId, cacheMode, []profilemodel.ProfileObject{profObj})
		if err != nil {
			tx.Warn("[PutProfileObject] Error while updating profile object in downstream cache: %v", err)
			return err
		}
		c.LogLatencyMs(METRIC_NAME_UPDATE_DOWNSTREAM_CACHE, float64(time.Since(startUpdateDynamo).Milliseconds()))
	} else {
		tx.Debug("[PutProfileObject] match found: Skipping update of profile cache")
	}

	// Send change event
	startSendEvent := time.Now()
	eventId := uuid.NewString()
	tx.Debug("[PutProfileObject] Sending change event: %s", eventId)
	eventParams := ChangeEventParams{
		EventID:        eventId,
		EventType:      eventType,
		Time:           startSendEvent,
		DomainName:     c.DomainName,
		ObjectTypeName: objectTypeName,
		ObjectID:       objectId,
		Profile:        profile,
		IsRealTime:     true,
	}
	err = SendChangeEvents(c.KinesisClient, eventParams)
	if err != nil {
		tx.Error("[PutProfileObject] Error sending change events to Kinesis %v", err)
	}
	c.LogLatencyMs(METRIC_NAME_SEND_CHANGE_EVENT, float64(time.Since(startSendEvent).Milliseconds()))
	c.LogLatencyMs(METRIC_NAME_PUT_PROFILE_LATENCY, float64(time.Since(startPutProfile).Milliseconds()))
	if err == nil {
		c.MetricLogger.LogMetric(map[string]string{"domain": c.DomainName}, METRIC_PUT_OBJECT_SUCCESS_COUNT, cloudwatch.Count, float64(1))
	}
	tx.Info("[PutProfileObject] successfully inserted object %s of type %s into profile %s", SENSITIVE_LOG_MASK, objectTypeName, profile.ProfileId)
	return err
}

func (c *CustomerProfileConfigLowCost) InsertProfileObjectWithPreMatch(
	domainConfig LcDomainConfig,
	objectTypeName, uniqueObjectKey string,
	objMap map[string]string,
	profMap map[string]string,
	mappings ObjectMapping,
	object string,
) (string, string, []RuleBasedMatch, RuleSet, error) {
	tx := core.NewTransaction(c.Tx.TransactionID, "", c.Tx.LogLevel)
	tx.Debug("[InsertProfileObjectWithPreMatch] Inserting object '%s' of type '%s' in domain %s", objMap["profile_id"], objectTypeName, domainConfig.Name)
	profileID := objMap["profile_id"]

	err := c.Data.validatePutProfileObjectInputs(profileID, objectTypeName, uniqueObjectKey, objMap, profMap, object)
	if err != nil {
		tx.Error("[InsertProfileObjectWithPreMatch] Input validation failed: %v", err)
		return "", "", []RuleBasedMatch{}, RuleSet{}, err
	}
	startTime := time.Now()
	var duration time.Duration

	//#region Configure Aurora Connections
	conn, err := c.Data.AurSvc.AcquireConnection(tx)
	if err != nil {
		tx.Warn("[InsertProfileObjectWithPreMatch] Error while acquiring connection: %v", err)
		return "", "", []RuleBasedMatch{}, RuleSet{}, err
	}
	defer conn.Release()
	// Attempting to configure readonly db instance
	var readonlyConn aurora.DBConnection
	if c.Data.AurReaderSvc != nil {
		tx.Debug("[InsertProfileObjectWithPreMatch] Using readonly endpoint")
		readonlyConn, err = c.Data.AurReaderSvc.AcquireConnection(tx)
		if err != nil {
			tx.Warn("[InsertProfileObjectWithPreMatch] Error while acquiring readonly connection: %v", err)
			return "", "", []RuleBasedMatch{}, RuleSet{}, err
		}
		defer readonlyConn.Release()
	} else {
		tx.Warn("[InsertProfileObjectWithPreMatch] Falling back to writer endpoint instead of reader endpoint")
		readonlyConn = conn
	}
	//#endregion Configure Aurora Connections

	//#region FindMatches
	var matches []RuleBasedMatch
	var activeRuleSet RuleSet
	var connectID string = core.UUID()
	var upsertType string = "insert"
	var inserted bool = false
	if domainConfig.Options.RuleBaseIdResolutionOn {
		activeRuleSet, err = c.Config.getRuleSet(domainConfig.Name, RULE_SET_TYPE_ACTIVE, RULE_SK_PREFIX)
		if err != nil {
			tx.Warn("Error while retrieving active rule set: %v", err)
		}
		if len(activeRuleSet.Rules) > 0 {
			profile := objectMapToProfile(objMap, mappings)
			profile.ProfileId = connectID
			startFindMatches := time.Now()
			matches, err = c.IR.FindMatches(readonlyConn, domainConfig.Name, objectTypeName, objMap, profile, activeRuleSet)
			if err != nil {
				tx.Warn("[InsertProfileObjectWithPreMatch] error finding matches: %v\nContinuing with no matches", err)
			}
			c.LogLatencyMs(METRIC_NAME_FIND_MATCHES, float64(time.Since(startFindMatches).Milliseconds()))

			if len(matches) > 0 {
				//	update master table
				startUpdateMasterTable := time.Now()
				var existingConnectID string
				existingConnectID, upsertType, err = c.Data.InsertPreMergedProfileInMasterTable(conn, tx, domainConfig.Name, matches[0].MatchIDs[0], connectID, profileID)
				inserted = true
				if err != nil {
					tx.Error("[InsertProfileObjectWithPreMatch] could not update profile in master table: %v")
					return "", "", []RuleBasedMatch{}, RuleSet{}, err
				}

				//	2 cases: existingConnectID == matches[0].MatchIDs[0], existingConnectID != matches[0].MatchIDs[0]
				if existingConnectID == matches[0].MatchIDs[0] {
					//	Pre-Merge occurred; log in profile history table
					c.MetricLogger.LogMetric(map[string]string{"domain": domainConfig.Name}, METRIC_NAME_PUT_OBJECT_MATCH_COUNT, cloudwatch.Count, float64(1))
					mergeID := core.GenerateUniqueId()
					tx.Debug("[InsertProfileObjectWithPreMatch][%s] Saving merge history during pre-merge", mergeID)
					query, args := insertProfileHistorySql(domainConfig.Name, matches[0].MatchIDs[0], connectID, ProfileMergeContext{
						Timestamp:              time.Now(),
						MergeType:              MergeTypeRule,
						ConfidenceUpdateFactor: 1,
						RuleID:                 matches[0].RuleIndex,
						RuleSetVersion:         fmt.Sprint(activeRuleSet.LatestVersion),
					})
					startSaveProfileHistory := time.Now()
					_, err = conn.Query(query, args...)
					duration = time.Since(startSaveProfileHistory)
					c.LogLatencyMs(METRIC_NAME_AURORA_PREMERGE_INSERT_PROFILE_HISTORY, float64(duration.Milliseconds()))
					if err != nil {
						tx.Error("[InsertProfileObjectWithPreMatch][%s] Error while saving merge history during pre-merge: %v", mergeID, err)
						return "", "", []RuleBasedMatch{}, RuleSet{}, err
					}

					//	Carry returned connectID (to allow different interactions with the same profileID) throughout
					//	the remainder of the ingestion process

					//	set connectID to first match
					connectID = matches[0].MatchIDs[0]

					//	First match is already pre-merged; ensure that what is sent to match queue is the un-used matches
					matches = matches[1:]
				} else {
					//	Carry returned `existingConnectID` throughout the remainder of the ingestion process
					//	Because the matches[0].MatchIDs[0] value is not used, all those matches must go through the merge queue
					connectID = existingConnectID

					//	Remove entries in `matches` where sourceID == connectID/existingConnectID
					for i, match := range matches {
						if match.SourceID == connectID {
							matches = append(matches[:i], matches[i+1:]...)
						}
					}
				}
				c.LogLatencyMs(METRIC_NAME_AURORA_UPSERT_MASTER_TABLE, float64(time.Since(startUpdateMasterTable).Milliseconds()))
			} else {
				//	insert profile into master table
				startProfileInsertInMasterTable := time.Now()

				//	pass generated connectID as both originalConnectID and connectID parameters
				//	overwrite connectID with returned value to handle ON CONFLICT (profileID) case
				//		1. No Conflict - returned connectID will match passed in connectID parameters
				//		2. Conflict on profileID - returned connectID will be updated to the existing connectID record in master table
				connectID, upsertType, err = c.Data.InsertPreMergedProfileInMasterTable(conn, tx, domainConfig.Name, connectID, connectID, profileID)
				duration := time.Since(startProfileInsertInMasterTable)
				inserted = true
				if err != nil {
					tx.Error("Error inserting profile in master table: %v", err)
					return "", "", []RuleBasedMatch{}, RuleSet{}, err
				}
				c.Data.MetricLogger.LogMetric(map[string]string{"domain": domainConfig.Name}, METRIC_NAME_AURORA_UPSERT_MASTER_TABLE, cloudwatch.Milliseconds, float64(duration.Milliseconds()))
			}
		}
	}

	if !inserted {
		connectID, upsertType, err = c.Data.InsertProfileInMasterTable(tx, domainConfig.Name, profileID)
		duration := time.Since(startTime)
		c.Data.MetricLogger.LogMetric(
			map[string]string{"domain": domainConfig.Name},
			METRIC_NAME_AURORA_UPSERT_MASTER_TABLE,
			cloudwatch.Milliseconds,
			float64(duration.Milliseconds()),
		)
		if err != nil {
			tx.Error("Error inserting profile in master table: %v", err)
			return "", "", []RuleBasedMatch{}, RuleSet{}, err
		}
		tx.Info("Successfully inserted profile ID %s into master table in %s ", profileID, duration)
	}
	//#endregion FindMatches

	//#region Insert into Search Table with potentially updated connectID
	//Insert profile level data + profile count update + object and history
	tx.Debug("[InsertProfileObjectWithPreMatch] starting transaction")
	err = conn.StartTransaction()
	if err != nil {
		tx.Error("Error starting aurora transaction: %v", err)
		return "", "", []RuleBasedMatch{}, RuleSet{}, fmt.Errorf("error starting aurora transaction: %s", err.Error())
	}
	tx.Debug(
		"[InsertProfileObjectWithPreMatch] operation: '%s' objectType: '%s' objectId: '%s' profileID: '%s' connectID: '%s' domain: '%s'",
		upsertType,
		objectTypeName,
		SENSITIVE_LOG_MASK,
		SENSITIVE_LOG_MASK,
		connectID,
		domainConfig.Name,
	)

	tx.Debug("[InsertProfileObjectWithPreMatch] upserting profile into search table")
	searchTableUpdateStartTime := time.Now()
	err = c.Data.UpsertProfileInSearchTable(conn, domainConfig.Name, connectID, objMap, profMap, objectTypeName)
	duration = time.Since(searchTableUpdateStartTime)
	conn.Tx.Debug("[InsertProfileObjectWithPreMatch] update search table duration:  %v", duration)
	c.Data.MetricLogger.LogMetric(
		map[string]string{"domain": domainConfig.Name},
		METRIC_NAME_AURORA_UPDATE_SEARCH_TABLE,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	duration = time.Since(startTime)
	tx.Debug("[InsertProfileObjectWithPreMatch] updateProfile (master + search table) duration:  %v", duration)
	c.Data.MetricLogger.LogMetric(
		map[string]string{"domain": domainConfig.Name},
		METRIC_NAME_AURORA_INSERT_PROFILE,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	if err != nil {
		err = NewRetryable(fmt.Errorf("database level error occurred while inserting into profile search table %w", err))
		return "", "", []RuleBasedMatch{}, RuleSet{}, c.Data.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("The following database-level error occurred while inserting into profile search table: %v", err), domainConfig.Name, err)
	}

	var eventType string
	if upsertType == "insert" {
		eventType = EventTypeCreated
		startTime := time.Now()
		tx.Info("[InsertProfileObjectWithPreMatch] New Profile created, incrementing profile count in domain %s", domainConfig.Name)
		err = c.Data.UpdateCount(conn, domainConfig.Name, countTableProfileObject, 1)
		if err != nil {
			return "", "", []RuleBasedMatch{}, RuleSet{}, c.Data.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("Error updating profile count: %v", err), domainConfig.Name, errors.New("Error updating profile count"))
		}
		duration := time.Since(startTime)
		tx.Debug("[InsertProfileObjectWithPreMatch] updateProfileCount duration:  %v", duration)
		c.Data.MetricLogger.LogMetric(
			map[string]string{"domain": domainConfig.Name},
			METRIC_NAME_UPDATE_PROFILE_COUNT_INSERT,
			cloudwatch.Milliseconds,
			float64(duration.Milliseconds()),
		)

	} else {
		eventType = EventTypeUpdated
	}

	tx.Info("[InsertProfileObjectWithPreMatch] Inserting object '%s' in interaction table", profileID)
	startTime = time.Now()
	query, args := insertObjectSql(domainConfig.Name, connectID, objectTypeName, uniqueObjectKey, objMap)
	objRes, err := conn.Query(query, args...)
	if err != nil {
		return "", "", []RuleBasedMatch{}, RuleSet{}, c.Data.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("[InsertProfileObjectWithPreMatch] The following database-level error occurred while inserting into object table %v: %v ", objectTypeName, err), domainConfig.Name, errors.New("a database-level error occurred while inserting object-level data"))
	}
	duration = time.Since(startTime)
	tx.Debug("[InsertProfileObjectWithPreMatch] %s duration:  %v", METRIC_NAME_AURORA_INSERT_OBJECT, duration)
	c.Data.MetricLogger.LogMetric(
		map[string]string{"domain": domainConfig.Name},
		METRIC_NAME_AURORA_INSERT_OBJECT,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)

	if len(objRes) == 0 {
		// this case should only happen if the db connection was lost, which we should retry
		err = NewRetryable(errors.New("[InsertProfileObjectWithPreMatch] upsert statement return was empty"))
		return "", "", []RuleBasedMatch{}, RuleSet{}, c.Data.rollbackInsert(conn, upsertType, connectID, "[InsertProfileObjectWithPreMatch] object upsert statement return was empty", domainConfig.Name, err)
	}
	if len(objRes) > 1 {
		return "", "", []RuleBasedMatch{}, RuleSet{}, c.Data.rollbackInsert(conn, upsertType, connectID, "[InsertProfileObjectWithPreMatch] object upsert statement returned more than one record", domainConfig.Name, errors.New("object upsert statement returned more than one record"))
	}
	upsertType, ok := objRes[0]["upsert_type"].(string)
	if !ok || upsertType != "insert" && upsertType != "update" {
		return "", "", []RuleBasedMatch{}, RuleSet{}, c.Data.rollbackInsert(conn, upsertType, connectID, "[InsertProfileObjectWithPreMatch] object upsert statement returned invalid upsert_type", domainConfig.Name, errors.New("object upsert statement returned invalid upsert_type"))
	}
	if upsertType == "insert" {
		startTime := time.Now()
		tx.Info("[InsertProfileObjectWithPreMatch] New Object created, incrementing object counts in domain %s", domainConfig.Name)
		err = c.Data.UpdateCount(conn, domainConfig.Name, objectTypeName, 1)
		if err != nil {
			return "", "", []RuleBasedMatch{}, RuleSet{}, c.Data.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("Error updating object count: %v", err), domainConfig.Name, errors.New("error updating object count"))
		}
		duration := time.Since(startTime)
		tx.Debug("updateObjectCount duration:  %v", duration)
		c.Data.MetricLogger.LogMetric(
			map[string]string{"domain": domainConfig.Name},
			METRIC_NAME_UPDATE_OBJECT_COUNT_INSERT,
			cloudwatch.Milliseconds,
			float64(duration.Milliseconds()),
		)

	}
	tx.Debug("[InsertProfileObjectWithPreMatch] Committing DB transaction")
	err = conn.CommitTransaction()
	if err != nil {
		return "", "", []RuleBasedMatch{}, RuleSet{}, c.Data.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("InsertProfileObjectWithPreMatch] error committing aurora transaction: %v ", err), domainConfig.Name, errors.New("error committing aurora transaction"))
	}
	tx.Info("[InsertProfileObjectWithPreMatch] operation successful: '%s' objectType: '%s' objectId: '%s' profileID: '%s' connectID: '%s' domain: '%s' pid: %d", upsertType, objectTypeName, SENSITIVE_LOG_MASK, SENSITIVE_LOG_MASK, connectID, domainConfig.Name)
	//#endregion

	return connectID, eventType, matches, activeRuleSet, err
}

/* 	Params:
*	object (string): Incoming Object to be processed
*	objectTypeName (string): Incoming Object Type
*	ruleSet (RuleSet): Rule Set to match on
 */
func (c *CustomerProfileConfigLowCost) RunRuleBasedIdentityResolution(
	object, objectTypeName, connectID string,
	ruleSet RuleSet,
) (bool, error) {
	tx := core.NewTransaction(c.Tx.TransactionID, "", c.Tx.LogLevel)
	tx.Info("[RunRuleBasedIdentityResolution] Processing object of type %v with connectId: %v", objectTypeName, connectID)
	tx.Info("[RunRuleBasedIdentityResolution] Rule set specified: %v", ruleSet)

	tx.Debug("[RunRuleBasedIdentityResolution] Getting domain config")
	domainCfg, err := c.Config.GetDomainConfig(c.DomainName)
	if err != nil {
		tx.Error("[RunRuleBasedIdentityResolution] Error while getting domain config: %v", err)
		return false, err
	}
	isRuleBasedEnabled := domainCfg.Options.RuleBaseIdResolutionOn
	if !isRuleBasedEnabled {
		tx.Info("[RunRuleBasedIdentityResolution] rule based matching is not enable. stopping")
		return false, nil
	}
	if objectTypeName == "" {
		tx.Error("[RunRuleBasedIdentityResolution] object type name is not provided. stopping")
		return false, fmt.Errorf("object type name is required")
	}
	if object == "" {
		tx.Error("[RunRuleBasedIdentityResolution] object data is not provided. stopping")
		return false, fmt.Errorf("object data is required")
	}
	tx.Debug("[RunRuleBasedIdentityResolution] Getting field mapping for %v", objectTypeName)
	mappings, err := c.GetMapping(objectTypeName)
	if err != nil {
		tx.Error("[RunRuleBasedIdentityResolution] Error while getting mapping: %v", err)
		return false, err
	}

	tx.Debug("[RunRuleBasedIdentityResolution] Parsing json string %v", objectTypeName)
	objMap, err := c.parseProfileObjectString(object)
	if err != nil {
		tx.Error("[RunRuleBasedIdentityResolution] Error parsing profile object: %v", err)
		return false, err
	}
	err = c.validateObject(objMap, objectTypeName, mappings)
	if err != nil {
		tx.Error("Invalid profile object of type: %v: %v", objectTypeName, err)
		return false, err
	}

	//we identify PROFILE index and set the profile_id column accordingly on the object
	//the profile ID column is sed for internal indexing
	objMap = c.addProfileIdField(objMap, mappings)

	uniqueKey := getIndexKey(mappings, INDEX_UNIQUE)
	if uniqueKey == "" {
		tx.Error(
			"[RunRuleBasedIdentityResolution] Error while getting unique key for object type %s. something may be wrong with mappings",
			objectTypeName,
		)
		return false, fmt.Errorf("error identifying unique key for object type %s", objectTypeName)
	}
	if objMap[uniqueKey] == "" {
		tx.Error(
			"[RunRuleBasedIdentityResolution] Error while getting unique key for object type %s. something may be wrong with mappings",
			objectTypeName,
		)
		return false, fmt.Errorf("unique key for %s must have valid value", objectTypeName)
	}

	//we should not rebuild the full profile here but just get the profile level data data

	//calling the GetProfile with no object type will just retrieve the profile level row in aurora and create a profile with no interaction
	tx.Debug("[RunRuleBasedIdentityResolution] Retreiving profile")
	profile, err := c.GetProfile(connectID, []string{}, []PaginationOptions{})
	if err != nil {
		tx.Error("[RunRuleBasedIdentityResolution] Error while getting profile: %v", err)
		return false, err
	}

	var isMergeCandidate bool
	startRuleBased := time.Now()
	tx.Debug("[RunRuleBasedIdentityResolution] Finding matches")
	if len(ruleSet.Rules) == 0 {
		tx.Info("[RunRuleBasedIdentityResolution] No rules in provided rule set, skipping rule based id resolution")
		return false, nil
	}

	// Acquire connection to writer
	conn, err := c.Data.AurSvc.AcquireConnection(tx)
	if err != nil {
		tx.Warn("[RunRuleBasedIdentityResolution] Error while acquiring connection: %v", err)
		return false, err
	}
	defer conn.Release()

	// Acquire connection to reader if available, otherwise fall back to the writer
	var readonlyConn aurora.DBConnection
	if c.Data.AurReaderSvc != nil {
		tx.Debug("[RunRuleBasedIdentityResolution] Using readonly endpoint")
		readonlyConn, err = c.Data.AurReaderSvc.AcquireConnection(tx)
		if err != nil {
			// Fall back to writer
			tx.Warn("[RunRuleBasedIdentityResolution] Error while acquiring readonly connection: %v. Falling back to writer connection.", err)
			readonlyConn = conn
		} else {
			defer readonlyConn.Release()
		}
	} else {
		tx.Warn("[RunRuleBasedIdentityResolution] readonly db connection not available, falling back to writer connection")
		readonlyConn = conn
	}

	matches, err := c.IR.FindMatches(readonlyConn, c.DomainName, objectTypeName, objMap, profile, ruleSet)
	if err != nil {
		tx.Warn("[RunRuleBasedIdentityResolution] Error while finding matches: %v", err)
		return false, err
	}
	if len(matches) > 0 {
		tx.Info("[RunRuleBasedIdentityResolution] found %d rule based matches. sending to merger", len(matches))
		isMergeCandidate = true
		err = c.sendRuleBasedMatches(matches, ruleSet)
		if err != nil {
			tx.Error("[RunRuleBasedIdentityResolution]  Error while sending rule based matches to merger: %v", err)
			return false, err
		}
	}
	c.LogLatencyMs(METRIC_NAME_PUT_OBJECT_RULE_BASED, float64(time.Since(startRuleBased).Milliseconds()))

	if !isMergeCandidate {
		tx.Debug("[RunRuleBasedIdentityResolution] isMergeCandidate = false: Adding profile to index table")
		startAddIndex := time.Now()
		err = c.IR.IndexInIRTable(conn, c.DomainName, objectTypeName, objMap, profile, ruleSet)
		if err != nil {
			tx.Warn("[RunRuleBasedIdentityResolution] Error while indexing profile for rule based identity resolution")
		}
		c.LogLatencyMs(METRIC_NAME_ADD_PROFILE_TO_INDEX_TABLE, float64(time.Since(startAddIndex).Milliseconds()))
	}
	return isMergeCandidate, err
}

// Converts json string into a map[string]string.
// We intentionally convert all values to strings for storage.
func (c *CustomerProfileConfigLowCost) parseProfileObjectString(str string) (map[string]string, error) {
	objMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(str), &objMap)
	if err != nil {
		return nil, err
	}

	strMap := make(map[string]string, len(objMap))
	for k, v := range objMap {
		strMap[k] = fmt.Sprintf("%v", v)
	}

	return strMap, nil
}

func (c *CustomerProfileConfigLowCost) addProfileIdField(obj map[string]string, mappings ObjectMapping) map[string]string {
	//the key should always exist in the mapping since it is valudated at mapping creation
	obj[LC_PROFILE_ID_KEY] = obj[getIndexKey(mappings, INDEX_PROFILE)]
	return obj
}

// Validate the object being processed contains only mapped fields.
func (c *CustomerProfileConfigLowCost) validateObject(obj map[string]string, objectTypeName string, mappings ObjectMapping) error {
	// Create set with mapped fields
	fields := map[string]interface{}{}
	for _, field := range mappings.Fields {
		fields[mappingToColumnName(field.Source)] = nil
	}

	// Validate each field in profile object is mapped
	unmappedFields := []string{}
	for objKey := range obj {
		if _, ok := fields[objKey]; !ok {
			unmappedFields = append(unmappedFields, objKey)
		}
	}

	if len(unmappedFields) > 0 {
		return fmt.Errorf("%s record contains fields that are not mapped: %v", objectTypeName, unmappedFields)
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) IsThrottlingError(err error) bool {
	// Not implemented
	// This was used to check if an error matched Amazon Connect Customer Profile's throttle error message.
	return false
}

// deprecated: replaced by Get UPT ID since this function has profileKey for compatibility with connect
// we don't need it
func (c *CustomerProfileConfigLowCost) GetProfileId(profileKey, profileId string) (string, error) {
	connectId, err := c.Data.FindConnectIDByProfileID(c.DomainName, profileId)
	if err != nil {
		c.Tx.Error("Error while fetching profile: %v", err)
		return "", err
	}
	return connectId, nil
}

// returns the upt_id (generated by UPT) from the profile_id (customer known id).
// returns an error if there is no upt_id for the given profile_id
func (c *CustomerProfileConfigLowCost) GetUptID(profileId string) (string, error) {
	connectId, err := c.Data.FindConnectIDByProfileID(c.DomainName, profileId)
	if err != nil {
		c.Tx.Error("Error while fetching profile: %v", err)
		return "", err
	}
	return connectId, nil
}

func (c *CustomerProfileConfigLowCost) GetSpecificProfileObject(
	objectKey string,
	objectId string,
	profileId string,
	objectTypeName string,
) (profilemodel.ProfileObject, error) {
	// need to identify table name
	// need to identify object key (need to validate or just use given key?)
	obj, err := c.Data.GetSpecificProfileObject(c.DomainName, objectKey, objectId, profileId, objectTypeName)
	if err != nil {
		c.Tx.Error("Error while fetching profile object: %v", err)
		return profilemodel.ProfileObject{}, err
	}
	return obj, nil
}

func (c *CustomerProfileConfigLowCost) GetProfileHistory(profileId string, domainName string, depth int) ([]ProfileHistory, error) {
	conn, err := c.Data.AurSvc.AcquireConnection(c.Tx)
	if err != nil {
		c.Tx.Error("Error acquiring connection: %v", err)
		return nil, err
	}
	defer conn.Release()
	history, err := c.Data.RetrieveProfileHistory(conn, profileId, domainName, depth)
	if err != nil {
		c.Tx.Error("Error retrieving profile history: %v", err)
		return nil, err
	}
	return history, nil
}

func (c *CustomerProfileConfigLowCost) LogLatencyMs(metricName string, value float64) {
	c.MetricLogger.LogMetric(map[string]string{"domain": c.DomainName}, metricName, cloudwatch.Milliseconds, value)
}

func (c *CustomerProfileConfigLowCost) ActivateIdResRuleSet() error {
	c.Tx.Info("[ActivateRuleSet] Starting ActivateRuleSet")
	if !c.isDomainSet() {
		return fmt.Errorf("[ActivateIdResRuleSet] domain is not set")
	}
	ruleBasedEnabled, err := c.isRuleBasedEnabled()
	if err != nil {
		return fmt.Errorf("error while checking if rule based matching is enabled to activate rule set: %v", err)
	}
	if !ruleBasedEnabled {
		return fmt.Errorf("rule based matching must be enabled")
	}
	oldActiveRuleSet, err := c.activateRuleSet(RULE_SK_PREFIX)
	if len(oldActiveRuleSet.Rules) > 0 {
		c.Tx.Debug("[ActivateRuleSet] emptying index index table for created for previous rule set")
		err = c.IR.EmptyIRTable(c.DomainName)
		if err != nil {
			c.Tx.Error("[ActivateRuleSet] Error while dropping old index table for rule set: %v", err)
			return err
		}
	}
	c.Config.invalidateCache(c.DomainName, controlPlaneCacheType_ActiveStitchingRuleSet)
	return err
}

func (c *CustomerProfileConfigLowCost) ActivateCacheRuleSet() error {
	_, err := c.activateRuleSet(RULE_SK_CACHE_PREFIX)
	c.Config.invalidateCache(c.DomainName, controlPlaneCacheType_ActiveCacheRuleSet)
	return err
}

func (c *CustomerProfileConfigLowCost) activateRuleSet(prefix string) (RuleSet, error) {
	c.Tx.Info("[ActivateRuleSet] Retrieving draft rule set")
	draftRuleSet, err := c.Config.getRuleSet(c.DomainName, RULE_SET_TYPE_DRAFT, prefix)
	if err != nil {
		c.Tx.Error("[ActivateRuleSet] Error while fetching draft rule set: %v", err)
		return RuleSet{}, err
	}
	if len(draftRuleSet.Rules) == 0 {
		c.Tx.Error("[ActivateRuleSet] No draft rule set found")
		return RuleSet{}, fmt.Errorf("no draft rule set found")
	}

	c.Tx.Debug("[ActivateRuleSet] Retrieving currently active rule set")
	oldActiveRuleSet, err := c.Config.getRuleSet(c.DomainName, RULE_SET_TYPE_ACTIVE, prefix)
	latestVersion := 0
	if err != nil {
		c.Tx.Error("[ActivateRuleSet] Error while fetching active rule set: %v", err)
		return RuleSet{}, err
	}
	if len(oldActiveRuleSet.Rules) > 0 {
		c.Tx.Debug("[ActivateRuleSet] Active rule set found, saving to historical")
		latestVersion = oldActiveRuleSet.LatestVersion
		err = c.deactivateRuleSet(oldActiveRuleSet, prefix)
		if err != nil {
			c.Tx.Error("[ActivateRuleSet] Error deactivating existing rule set: %v", err)
		}
	} else {
		c.Tx.Info("[ActivateRuleSet] No active rule set found")
	}

	c.Tx.Debug("[ActivateRuleSet] Creating new active rule set using draft rules")
	newActiveRuleSet := RuleSet{
		Name:                    RULE_SET_TYPE_ACTIVE,
		Rules:                   draftRuleSet.Rules,
		LastActivationTimestamp: time.Now(),
		LatestVersion:           latestVersion + 1,
	}
	err = c.Config.createRuleSet(c.DomainName, newActiveRuleSet, prefix)
	if err != nil {
		c.Tx.Error("[ActivateRuleSet] Error while creating new active rule set: %v", err)
		return RuleSet{}, err
	}
	c.Tx.Info("[ActivateRuleSet] Successfully activated new rule set, deleting draft rule set")

	err = c.Config.deleteRuleSet(c.DomainName, RULE_SET_TYPE_DRAFT, prefix)
	if err != nil {
		c.Tx.Error("[ActivateRuleSet] Error while deleting draft rule set: %v", err)
		return RuleSet{}, err
	}
	c.Tx.Info("[ActivateRuleSet] Successfully deleted draft rule set")
	return oldActiveRuleSet, nil
}

func (c *CustomerProfileConfigLowCost) SaveIdResRuleSet(rules []Rule) error {
	c.Tx.Info("[SaveRuleSet] Starting SaveRuleSet")
	if !c.isDomainSet() {
		return fmt.Errorf("[SaveIdResRuleSet] domain is not set")
	}
	ruleBasedEnabled, err := c.isRuleBasedEnabled()
	if err != nil {
		return fmt.Errorf("error while checking if rule based matching is enabled to save ruleset: %v", err)
	}
	if !ruleBasedEnabled {
		return fmt.Errorf("rule based matching must be enabled")
	}
	allMappings := make(map[string][]string)

	profileFields, err := c.GetProfileLevelFields()
	if err != nil {
		return err
	}
	allMappings["_profile"] = profileFields

	objectFields, err := c.GetObjectLevelFields()
	if err != nil {
		return err
	}
	for name, fieldNames := range objectFields {
		allMappings[name] = fieldNames
	}

	if len(rules) > MAX_RULE_SET_SIZE {
		return fmt.Errorf("rule set exceeds limit of %d rules", MAX_RULE_SET_SIZE)
	}
	for _, rule := range rules {
		conditions := rule.Conditions
		errMsgs := validateConditions(conditions, allMappings)
		if errMsgs != nil {
			return errMsgs
		}
	}

	return c.saveRuleSet(rules, RULE_SK_PREFIX)
}

func (c *CustomerProfileConfigLowCost) SaveCacheRuleSet(rules []Rule) error {
	allMappings := make(map[string][]string)

	profileFields, err := c.GetProfileLevelFields()
	if err != nil {
		return err
	}
	allMappings["_profile"] = profileFields

	objectFields, err := c.GetObjectLevelFields()
	if err != nil {
		return err
	}
	for name, fieldNames := range objectFields {
		allMappings[name] = fieldNames
	}

	if len(rules) > MAX_RULE_SET_SIZE {
		return fmt.Errorf("rule set exceeds limit of %d rules", MAX_RULE_SET_SIZE)
	}
	for _, rule := range rules {
		conditions := rule.Conditions
		errMsgs := validateCacheConditions(conditions, allMappings)
		if errMsgs != nil {
			return errMsgs
		}
	}
	return c.saveRuleSet(rules, RULE_SK_CACHE_PREFIX)
}

func (c *CustomerProfileConfigLowCost) saveRuleSet(rules []Rule, prefix string) error {
	c.Tx.Debug("[SaveRuleSet] Deleting existing rule set if it exists")
	err := c.Config.deleteRuleSet(c.DomainName, RULE_SET_TYPE_DRAFT, prefix)
	if err != nil {
		c.Tx.Error("[SaveRuleSet] Error while deleting rule set: %v", err)
		return err
	}
	ruleSet := RuleSet{
		Name:  RULE_SET_TYPE_DRAFT,
		Rules: rules,
	}
	err = c.Config.createRuleSet(c.DomainName, ruleSet, prefix)
	if err != nil {
		c.Tx.Error("[SaveRuleSet] Error while creating rule set: %v", err)
		return err
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) ListIdResRuleSets(includeHistorical bool) ([]RuleSet, error) {
	c.Tx.Info("[ListRuleSets] Starting ListRuleSets")
	if !c.isDomainSet() {
		return []RuleSet{}, fmt.Errorf("[ListIdResRuleSets] domain is not set")
	}
	ruleBasedEnabled, err := c.isRuleBasedEnabled()
	if err != nil {
		return []RuleSet{}, fmt.Errorf("error while checking if rule based matching is enabled to list rulesets: %v", err)
	}
	if !ruleBasedEnabled {
		return []RuleSet{}, fmt.Errorf("rule based matching is not enabled")
	}

	return c.listRuleSets(includeHistorical, RULE_SK_PREFIX)
}

func (c *CustomerProfileConfigLowCost) ListCacheRuleSets(includeHistorical bool) ([]RuleSet, error) {
	return c.listRuleSets(includeHistorical, RULE_SK_CACHE_PREFIX)
}

func (c *CustomerProfileConfigLowCost) listRuleSets(includeHistorical bool, prefix string) ([]RuleSet, error) {
	ruleSets := []RuleSet{}
	if includeHistorical {
		c.Tx.Debug("[ListRuleSets] Retrieving all rule sets")
		allRuleSets, err := c.Config.GetAllRuleSets(c.DomainName, prefix)
		if err != nil {
			c.Tx.Error("[ListRuleSets] Error while getting all rule sets: %v", err)
			return nil, err
		}
		c.Tx.Debug("[ListRuleSets] allRuleSets: %+v", allRuleSets)
		ruleSets = append(ruleSets, allRuleSets...)
		c.Tx.Debug("[ListRuleSets] ruleSets: %+v", ruleSets)
	} else {
		c.Tx.Debug("[ListRuleSets] Retrieving only active and draft rule sets")
		activeRuleSet, err := c.Config.getRuleSet(c.DomainName, RULE_SET_TYPE_ACTIVE, prefix)
		if err != nil {
			c.Tx.Error("[ListRuleSets] Error while getting active rule sets: %v", err)
			return nil, err
		}
		ruleSets = append(ruleSets, activeRuleSet)
		draftRuleSet, err := c.Config.getRuleSet(c.DomainName, RULE_SET_TYPE_DRAFT, prefix)
		if err != nil {
			c.Tx.Error("[ListRuleSets] Error while getting draft rule sets: %v", err)
			return nil, err
		}
		ruleSets = append(ruleSets, draftRuleSet)
	}
	return ruleSets, nil
}

func (c *CustomerProfileConfigLowCost) GetActiveRuleSet() (RuleSet, error) {
	c.Tx.Info("[GetActiveRuleSet] Starting GetActiveRuleSet")
	if !c.isDomainSet() {
		return RuleSet{}, fmt.Errorf("[GetActiveRuleSet] domain is not set")
	}
	ruleBasedEnabled, err := c.isRuleBasedEnabled()
	if err != nil {
		return RuleSet{}, fmt.Errorf("error while checking if rule based matching is enabled to get active ruleset: %v", err)
	}
	if !ruleBasedEnabled {
		return RuleSet{}, fmt.Errorf("rule based matching is not enabled")
	}

	c.Tx.Debug("[GetActiveRuleSet] Retrieving only active rule set")
	activeRuleSet, err := c.Config.getRuleSet(c.DomainName, RULE_SET_TYPE_ACTIVE, RULE_SK_PREFIX)
	if err != nil {
		c.Tx.Error("[GetActiveRuleSet] Error while getting active rule set: %v", err)
		return RuleSet{}, err
	}

	return activeRuleSet, nil
}

func (c *CustomerProfileConfigLowCost) isDomainSet() bool {
	return c.DomainName != ""
}

func (c *CustomerProfileConfigLowCost) isRuleBasedEnabled() (bool, error) {
	return c.DomainConfig.Options.RuleBaseIdResolutionOn, nil
}

// Enable rule based identity resolution by updating the domain config in DynamoDB
// to indicate rules should be used when we insert interaction records or merge profiles.
func (c *CustomerProfileConfigLowCost) EnableRuleBasedIdRes() error {
	c.Tx.Info("[EnableRuleBasedIdRes] Starting EnableRuleBasedIdRes")
	if !c.isDomainSet() {
		return fmt.Errorf("[EnableRuleBasedIdRes] domain is not set")
	}
	cfg, err := c.Config.GetDomainConfig(c.DomainName)
	if err != nil {
		return fmt.Errorf("[EnableRuleBasedIdRes] error retrieving domain config: %v", err)
	}
	if cfg.Options.RuleBaseIdResolutionOn {
		return fmt.Errorf("rule based matching is already enabled")
	}

	cfg.Options.RuleBaseIdResolutionOn = true
	err = c.Config.SaveDomainConfig(cfg)
	if err != nil {
		return fmt.Errorf("error saving updated domain config: %v", err)
	}

	return nil
}

// Disable rule based identity resolution by updating the domain config in DynamoDB
// to indicate rules should not be used when we insert interaction records or merge profiles.
//
// If there is an active rule set, we change it to a historical rule set
// that can still be viewed.
func (c *CustomerProfileConfigLowCost) DisableRuleBasedIdRes() error {
	c.Tx.Info("[DisableRuleBasedIdRes] Starting DisableRuleBasedIdRes")
	if !c.isDomainSet() {
		return fmt.Errorf("[DisableRuleBasedIdRes] domain is not set")
	}
	cfg, err := c.Config.GetDomainConfig(c.DomainName)
	if err != nil {
		return fmt.Errorf("DisableRuleBasedIdRes]error retrieving domain config: %v", err)
	}
	if !cfg.Options.RuleBaseIdResolutionOn {
		return fmt.Errorf("rule based matching is already disabled")
	}
	cfg.Options.RuleBaseIdResolutionOn = false
	err = c.Config.SaveDomainConfig(cfg)
	if err != nil {
		return fmt.Errorf("error saving updated domain config: %v", err)
	}

	activeRuleSet, err := c.Config.getRuleSet(c.DomainName, RULE_SET_TYPE_ACTIVE, RULE_SK_PREFIX)
	if err != nil {
		c.Tx.Error("[DisableRuleBasedIdRes] Error while getting active rule set: %v", err)
		return err
	}
	if len(activeRuleSet.Rules) > 0 {
		c.deactivateRuleSet(activeRuleSet, RULE_SK_PREFIX)
	}
	return nil
}

// Given an interaction (interactionId), this function returns the history of all merges/unmerges which affected the interaction
func (c *CustomerProfileConfigLowCost) BuildInteractionHistory(
	domain string,
	interactionId string,
	objectTypeName string,
	connectId string,
) ([]ProfileMergeContext, error) {
	conn, err := c.Data.AurSvc.AcquireConnection(c.Tx)
	if err != nil {
		c.Tx.Error("[BuildInteractionHistory] error acquiring connection: %s", err)
		return []ProfileMergeContext{}, err
	}
	defer conn.Release()
	// get original profile the interaction belonged to
	originalConnectId, profileId, err := c.Data.getOriginalConnectIdForInteraction(conn, domain, interactionId, objectTypeName, connectId)

	if err != nil {
		c.Tx.Error("[BuildInteractionHistory] error finding the connectId for interaction; %s", err)
		return []ProfileMergeContext{}, err
	}

	// get the interaction creation record
	interactionCreationEvent, err := c.Data.retrieveInteractionCreation(conn, domain, objectTypeName, interactionId, profileId)

	if err != nil {
		c.Tx.Error("[BuildInteractionHistory]  error finding the interaction creation; %s", err)
		return []ProfileMergeContext{}, err
	}

	// starting at the orignal connect id, gets all merges leading to the current parent profile
	interactionHistory, err := c.Data.getMergesLeadingToParent(conn, domain, originalConnectId)
	if err != nil {
		c.Tx.Error("[BuildInteractionHistory] error finding the merges; %s", err)
		return []ProfileMergeContext{}, err
	}

	return append([]ProfileMergeContext{interactionCreationEvent}, interactionHistory...), err

}

func (c *CustomerProfileConfigLowCost) deactivateRuleSet(ruleSet RuleSet, prefix string) error {
	latestVersion := ruleSet.LatestVersion
	historicalRuleSet := RuleSet{
		Name:                    "v" + strconv.Itoa(latestVersion),
		Rules:                   ruleSet.Rules,
		LastActivationTimestamp: ruleSet.LastActivationTimestamp,
	}
	err := c.Config.createRuleSet(c.DomainName, historicalRuleSet, prefix)
	if err != nil {
		return fmt.Errorf("error while creating historical copy of rule set: %v", err)
	}

	c.Tx.Info("[ActivateRuleSet] Successfully saved historical rule set, deleting old active rule set")
	err = c.Config.deleteRuleSet(c.DomainName, RULE_SET_TYPE_ACTIVE, prefix)
	if err != nil {
		return fmt.Errorf("error while deleting active rule set: %v", err)
	}

	return nil
}

func (c *CustomerProfileConfigLowCost) FindCurrentParentOfProfile(domain, originalConnectId string) (string, error) {
	conn, err := c.Data.AurSvc.AcquireConnection(c.Tx)
	if err != nil {
		return "", fmt.Errorf("[FindCurrentParentOfProfile] error acquiring connection: %s", err)
	}
	defer conn.Release()

	connectIds, err := c.Data.getIdsForOriginalConnectIds(conn, domain, "connect_id", []string{originalConnectId})
	if err != nil || len(connectIds) != 1 {
		return "", fmt.Errorf("[FindCurrentParentOfProfile] error finding parent for profile; %s", err)
	}
	return connectIds[0], nil
}

func organizeProfilePairsBySourceID(pairs []ProfilePair) map[string][]ProfilePair {
	profilePairsBySourceID := map[string][]ProfilePair{}
	for _, v := range pairs {
		profilePairsBySourceID[v.SourceID] = append(profilePairsBySourceID[v.SourceID], v)
	}
	return profilePairsBySourceID
}

// Temporary Function to retrieve Object Table
func (c *CustomerProfileConfigLowCost) GetInteractionTable(
	domain, objectType string,
	partition UuidPartition,
	lastId uuid.UUID,
	limit int,
) ([]map[string]interface{}, error) {
	objects, err := c.Data.ShowInteractionTablePartitionPage(domain, objectType, partition, lastId, limit)
	if err != nil {
		return nil, fmt.Errorf("[GetInteractionTable] error searching on interaction table %v; %s", objectType, err)
	}
	return objects, nil
}

func (c *CustomerProfileConfigLowCost) sendRuleBasedMatches(matches []RuleBasedMatch, activeRuleSet RuleSet) error {
	for _, match := range matches {
		mergeRq := MergeRequest{
			SourceID:      match.SourceID,
			TargetID:      match.MatchIDs[0],
			DomainName:    c.DomainName,
			TransactionID: c.Tx.TransactionID,
			Context: ProfileMergeContext{
				MergeType:              MergeTypeRule,
				ConfidenceUpdateFactor: 1.0,
				RuleID:                 match.RuleIndex,
				RuleSetVersion:         fmt.Sprintf("%d", activeRuleSet.LatestVersion),
			},
		}
		data, err := json.Marshal(mergeRq)
		if err != nil {
			c.Tx.Warn("Error marshaling merge request: %v", err)
			return err
		}
		err = c.MergeQueueClient.SendWithStringAttributes(string(data), map[string]string{})
		if err != nil {
			c.Tx.Warn("Error sending merge request to queue: %v", err)
			return err
		}
	}

	return nil
}

func (c *CustomerProfileConfigLowCost) ClearDynamoCache() error {
	cfg := c.DomainConfig
	if !cfg.DynamoMode {
		c.Tx.Error("dynamo cache is not enabled")
		return fmt.Errorf("dynamo cache is not enabled")
	}
	err := c.Data.DynSvc.ReplaceTable()
	if err != nil {
		c.Tx.Error("error replacing dynamo table: %v", err)
		return fmt.Errorf("error replacing dynamo table: %v", err)
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) ClearCustomerProfileCache(tags map[string]string, matchS3Name string) error {
	cfg := c.DomainConfig
	if !cfg.CustomerProfileMode {
		c.Tx.Error("CP cache is not enabled")
		return fmt.Errorf("CP cache is not enabled")
	}
	mappings, err := c.GetMappings()
	if err != nil {
		c.Tx.Error("error retrieving mappings")
		return fmt.Errorf("error retrieving mappings")
	}
	err = c.DeleteCPCacheByName(c.DomainName)
	if err != nil {
		c.Tx.Error("error deleting CP domain: %v", err)
		return fmt.Errorf("error deleting CP domain: %v", err)
	}
	//temporary solution, issue on CPs end regarding deletion and domain recreation
	//with same name. There is no way in CP to determine if a domain has actually been deleted
	//and everything happens in a different order from the users order of operations.
	//Issue on CP's side, they said they will fix
	//70s is a safe delay, according to them, set to 90s to give us a buffer, see ticket
	//for citation
	time.Sleep(90 * time.Second)
	cpDomainParams := CpDomainParams{
		kmsArn:    cfg.KmsArn,
		queueUrl:  cfg.QueueUrl,
		streamArn: cfg.Options.CpExportDestinationArn,
	}
	err = c.CreateCPCache(c.DomainName, tags, matchS3Name, cpDomainParams)
	if err != nil {
		c.Tx.Error("error creating CP domain: %v, deleting domain", err)
		c.DeleteCPCacheByName(c.DomainName)
		return fmt.Errorf("error creating CP domain: %v", err)
	}
	for _, mapping := range mappings {
		err = c.CreateCPDomainMapping(c.DomainName, mapping)
		if err != nil {
			c.Tx.Error("error recreating CP domain mappings, mapping: %v: %v", mapping.Name, err)
			err = errors.Join(fmt.Errorf("error recreating CP domain mappings, mapping: %v: %v", mapping.Name, err), err)
		}
	}
	return err
}

func (c *CustomerProfileConfigLowCost) CacheProfile(connectID string, cacheMode CacheMode) error {
	allObjectTypeNames := []string{}
	domainCfg, err := c.Config.GetDomainConfig(c.DomainName)
	if err != nil {
		c.Tx.Error("[CacheProfile] Error while getting domain config: %v", err)
		return fmt.Errorf("error while getting domain config: %v", err)
	}
	allMappings, err := c.GetMappings()
	if err != nil {
		c.Tx.Error("[CacheProfile] Error while getting mapping: %v", err)
		return fmt.Errorf("error while getting mapping: %v", err)
	}
	for _, mapping := range allMappings {
		allObjectTypeNames = append(allObjectTypeNames, mapping.Name)
	}

	c.Tx.Debug("[PutProfileObject] retrieving profile search record")
	var rowLevelData map[string]interface{}
	connectID, rowLevelData, err = c.Data.GetProfileSearchRecord(c.DomainName, connectID)
	if err != nil {
		c.Tx.Error("Error while getting profile search record: %v", err)
		return err
	}
	conn, err := c.Data.AurSvc.AcquireConnection(c.Tx)
	if err != nil {
		c.Tx.Error("[CacheProfile] Error while acquiring connection: %v", err)
		return fmt.Errorf("error while acquiring connection: %v", err)
	}
	defer conn.Release()
	objectTypeNameToId := make(map[string]string, len(allObjectTypeNames))
	for _, objectType := range allObjectTypeNames {
		objectTypeNameToId[objectType] = ""
	}
	profile, _, _, err := c.Data.buildProfile(
		conn,
		c.DomainName,
		connectID,
		objectTypeNameToId,
		allMappings,
		rowLevelData,
		domainCfg.Options.ObjectTypePriority,
		[]PaginationOptions{},
	)
	if err != nil {
		c.Tx.Error("Error while building profile: %v", err)
		return fmt.Errorf("error while building profile: %v", err)
	}
	return c.UpdateDownstreamCache(profile, "", cacheMode, profile.ProfileObjects)
}

func (c *CustomerProfileConfigLowCost) GetProfilePartition(partition UuidPartition) ([]string, error) {
	allRows, err := c.Data.GetProfilePartition(c.DomainName, partition.LowerBound.String(), partition.UpperBound.String())
	if err != nil {
		c.Tx.Error(
			"[GetProfilePartition] Error while getting profiles in partition %v to %v: %v",
			partition.LowerBound.String(),
			partition.UpperBound.String(),
			err,
		)
		return []string{}, fmt.Errorf(
			"[GetProfilePartition] Error while getting profiles in partition %v to %v: %v",
			partition.LowerBound.String(),
			partition.UpperBound.String(),
			err,
		)

	}
	connectIds := []string{}
	for _, row := range allRows {
		connectId, ok := row["connect_id"].(string)
		if !ok {
			return []string{}, fmt.Errorf("[GetProfilePartition] Error while getting connect ids")
		}
		connectIds = append(connectIds, connectId)
	}
	return connectIds, nil
}

func (c *CustomerProfileConfigLowCost) UpdateVersion(version feature.FeatureSetVersion) (err error) {
	oldVersion := c.DomainConfig.DomainVersion
	oldCompatibleVersion := c.DomainConfig.CompatibleVersion

	defer func() {
		if err != nil {
			c.Tx.Info("[UpdateVersion] Rolling back domain version: %v", err)
			c.DomainConfig.DomainVersion = oldVersion
			c.DomainConfig.CompatibleVersion = oldCompatibleVersion
		}
	}()

	c.DomainConfig.DomainVersion = version.Version
	c.DomainConfig.CompatibleVersion = version.CompatibleVersion

	err = c.Config.SaveDomainConfig(c.DomainConfig)
	if err != nil {
		c.Tx.Error("[UpdateVersion] Error updating domain version: %v", err)
		return err
	}

	return nil
}

func (c *CustomerProfileConfigLowCost) GetVersion() (version feature.FeatureSetVersion, err error) {
	config, err := c.Config.GetDomainConfig(c.DomainName)
	if err != nil {
		c.Tx.Error("[GetVersion] Error getting domain config: %v", err)
		return feature.FeatureSetVersion{}, err
	}
	return feature.FeatureSetVersion{
		Version:           config.DomainVersion,
		CompatibleVersion: config.CompatibleVersion,
	}, nil
}

func (c *CustomerProfileConfigLowCost) InvalidateStitchingRuleSetsCache() error {
	c.Config.invalidateCache(c.DomainName, controlPlaneCacheType_ActiveStitchingRuleSet)
	return nil
}
