// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/feature"
	"time"

	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockV2 struct {
	mock.Mock
}

var _ ICustomerProfileLowCostConfig = &MockV2{}

func InitMockV2() *MockV2 {
	return &MockV2{}
}

// CreateDomainWithQueue implements ICustomerProfileConfig.
func (m *MockV2) CreateDomainWithQueue(
	name string,
	kmsArn string,
	tags map[string]string,
	queueUrl string,
	matchS3Name string,
	options DomainOptions,
) error {
	args := m.Called(name, kmsArn, tags, queueUrl, matchS3Name, options)
	return args.Error(0)
}

// CreateEventStream implements ICustomerProfileConfig.
func (m *MockV2) CreateEventStream(domainName string, streamName string, streamArn string) error {
	args := m.Called(domainName, streamName, streamArn)
	return args.Error(0)
}

// CreateMapping implements ICustomerProfileConfig.
func (m *MockV2) CreateMapping(name string, description string, fieldMappings []FieldMapping) error {
	args := m.Called(name, description, fieldMappings)
	return args.Error(0)
}

// DeleteDomain implements ICustomerProfileConfig.
func (m *MockV2) DeleteDomain() error {
	args := m.Called()
	return args.Error(0)
}

// DeleteDomainByName implements ICustomerProfileConfig.
func (m *MockV2) DeleteDomainByName(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// DeleteEventStream implements ICustomerProfileConfig.
func (m *MockV2) DeleteEventStream(domainName string, streamName string) error {
	args := m.Called(domainName, streamName)
	return args.Error(0)
}

// DeleteProfile implements ICustomerProfileConfig.
func (m *MockV2) DeleteProfile(id string, options ...DeleteProfileOptions) error {
	args := m.Called(id, options)
	return args.Error(0)
}

// GetDomain implements ICustomerProfileConfig.
func (m *MockV2) GetDomain() (Domain, error) {
	args := m.Called()
	return args.Get(0).(Domain), args.Error(1)
}

// GetMappings implements ICustomerProfileConfig.
func (m *MockV2) GetMappings() ([]ObjectMapping, error) {
	args := m.Called()
	return args.Get(0).([]ObjectMapping), args.Error(1)
}

// GetProfile implements ICustomerProfileConfig.
func (m *MockV2) GetProfile(id string, objectTypeNames []string, pagination []PaginationOptions) (profilemodel.Profile, error) {
	args := m.Called(id, objectTypeNames, pagination)
	return args.Get(0).(profilemodel.Profile), args.Error(1)
}

// GetProfileId implements ICustomerProfileConfig.
func (m *MockV2) GetProfileId(profileKey string, profileId string) (string, error) {
	args := m.Called(profileKey, profileId)
	return args.String(0), args.Error(1)
}

// GetSpecificProfileObject implements ICustomerProfileConfig.
func (m *MockV2) GetSpecificProfileObject(
	objectKey string,
	objectId string,
	profileId string,
	objectTypeName string,
) (profilemodel.ProfileObject, error) {
	args := m.Called(objectKey, objectId, profileId, objectTypeName)
	return args.Get(0).(profilemodel.ProfileObject), args.Error(1)
}

// IsThrottlingError implements ICustomerProfileConfig.
func (m *MockV2) IsThrottlingError(err error) bool {
	args := m.Called(err)
	return args.Bool(0)
}

// ListDomains implements ICustomerProfileConfig.
func (m *MockV2) ListDomains() ([]Domain, error) {
	args := m.Called()
	return args.Get(0).([]Domain), args.Error(1)
}

// MergeMany implements ICustomerProfileConfig.
func (m *MockV2) MergeMany(pairs []ProfilePair) (string, error) {
	args := m.Called(pairs)
	return args.String(0), args.Error(1)
}

// MergeProfiles implements ICustomerProfileConfig.
func (m *MockV2) MergeProfiles(profileId string, profileIdToMerge string, mergeContext ProfileMergeContext) (string, error) {
	args := m.Called(profileId, profileIdToMerge, mergeContext)
	return args.String(0), args.Error(1)
}

// PutIntegration implements ICustomerProfileConfig.
func (m *MockV2) PutIntegration(
	flowNamePrefix string,
	objectName string,
	bucketName string,
	bucketPrefix string,
	fieldMappings []FieldMapping,
	startTime time.Time,
) error {
	args := m.Called(flowNamePrefix, objectName, bucketName, bucketPrefix, fieldMappings, startTime)
	return args.Error(0)
}

// PutProfileObject implements ICustomerProfileConfig.
func (m *MockV2) PutProfileObject(object string, objectTypeName string) error {
	args := m.Called(object, objectTypeName)
	return args.Error(0)
}

// RunRuleBasedIdentityResolution implements ICustomerProfileConfig.
func (m *MockV2) RunRuleBasedIdentityResolution(object, objectTypeName, connectID string, ruleSet RuleSet) (bool, error) {
	args := m.Called(object, objectTypeName, connectID, ruleSet)
	return args.Bool(0), args.Error(1)
}

// GetProfileLevelFields implements ICustomerProfileConfig.
func (m *MockV2) GetProfileLevelFields() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

// GetProfileLevelFields implements ICustomerProfileConfig.
func (m *MockV2) GetObjectLevelFields() (map[string][]string, error) {
	args := m.Called()
	return args.Get(0).(map[string][]string), args.Error(1)
}

// SearchProfiles implements ICustomerProfileConfig.
func (m *MockV2) SearchProfiles(key string, values []string) ([]profilemodel.Profile, error) {
	args := m.Called(key, values)
	return args.Get(0).([]profilemodel.Profile), args.Error(1)
}

// AdvancedProfileSearch implements ICustomerProfileConfig.
func (m *MockV2) AdvancedProfileSearch(searchCriteria []BaseCriteria) ([]profilemodel.Profile, error) {
	args := m.Called(searchCriteria)
	return args.Get(0).([]profilemodel.Profile), args.Error(1)
}

// SetDomain implements ICustomerProfileConfig.
func (m *MockV2) SetDomain(domain string) error {
	args := m.Called(domain)
	return args.Error(0)
}

// SetTx implements ICustomerProfileConfig.
func (m *MockV2) SetTx(tx core.Transaction) {
	m.Called(tx)
}

// ActivateRuleSet implements ICustomerProfileConfig.
func (m *MockV2) ActivateIdResRuleSet() error {
	args := m.Called()
	return args.Error(0)
}

// ListRuleSets implements ICustomerProfileConfig.
func (m *MockV2) ListIdResRuleSets(includeHistorical bool) ([]RuleSet, error) {
	args := m.Called(includeHistorical)
	return args.Get(0).([]RuleSet), args.Error(1)
}

func (m *MockV2) GetActiveRuleSet() (RuleSet, error) {
	args := m.Called()
	return args.Get(0).(RuleSet), args.Error(1)
}

// SaveRuleSet implements ICustomerProfileConfig.
func (m *MockV2) SaveIdResRuleSet(rules []Rule) error {
	args := m.Called(rules)
	return args.Error(0)
}

// ActivateCacheRuleSet implements ICustomerProfileConfig.
func (m *MockV2) ActivateCacheRuleSet() error {
	args := m.Called()
	return args.Error(0)
}

// ListRuleSetsCache implements ICustomerProfileConfig.
func (m *MockV2) ListCacheRuleSets(includeHistorical bool) ([]RuleSet, error) {
	args := m.Called(includeHistorical)
	return args.Get(0).([]RuleSet), args.Error(1)
}

// SaveCacheRuleSet implements ICustomerProfileConfig.
func (m *MockV2) SaveCacheRuleSet(rules []Rule) error {
	args := m.Called(rules)
	return args.Error(0)
}

// ToggleAiIdentityResolution implements ICustomerProfileConfig.
func (m *MockV2) ToggleAiIdentityResolution(domainName string, s3BucketName string, roleArn string) error {
	args := m.Called(domainName, s3BucketName, roleArn)
	return args.Error(0)
}

// EnableRuleBasedIdRes implements ICustomerProfileConfig.
func (m *MockV2) EnableRuleBasedIdRes() error {
	args := m.Called()
	return args.Error(0)
}

// DisableRuleBasedIdRes implements ICustomerProfileConfig.
func (m *MockV2) DisableRuleBasedIdRes() error {
	args := m.Called()
	return args.Error(0)
}

// EnableEntityResolution implements ICustomerProfileConfig.
func (m *MockV2) EnableEntityResolution(domainName string, s3BucketName string, roleArn string, outputBucket string) error {
	args := m.Called(domainName, s3BucketName, roleArn, outputBucket)
	return args.Error(0)
}

// DisableEntityResolution implements ICustomerProfileConfig.
func (m *MockV2) DisableEntityResolution(domainName string) error {
	args := m.Called(domainName)
	return args.Error(0)
}

func (m *MockV2) BuildInteractionHistory(domain string, interactionId string, objectTypeName string, connectId string) ([]ProfileMergeContext, error) {
	args := m.Called(domain, interactionId, objectTypeName)
	return args.Get(0).([]ProfileMergeContext), args.Error(1)
}

func (m *MockV2) UnmergeProfiles(
	toUnmergeConnectID string,
	mergedIntoConnectID string,
	interactionIdToUnmerge string,
	interactionType string,
) (string, error) {
	args := m.Called(toUnmergeConnectID, mergedIntoConnectID, interactionIdToUnmerge, interactionType)
	return args.String(0), args.Error(1)
}

func (m *MockV2) FindCurrentParentOfProfile(domain, originalConnectId string) (string, error) {
	args := m.Called(domain, originalConnectId)
	return args.String(0), args.Error(1)
}

func (m *MockV2) GetInteractionTable(
	domain, objectType string,
	partition UuidPartition,
	lastId uuid.UUID,
	limit int,
) ([]map[string]interface{}, error) {
	args := m.Called(domain, objectType, partition, lastId, limit)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockV2) ClearDynamoCache() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockV2) ClearCustomerProfileCache(tags map[string]string, matchS3Name string) error {
	args := m.Called(tags, matchS3Name)
	return args.Error(0)
}

func (m *MockV2) CacheProfile(connectID string, cacheMode CacheMode) error {
	args := m.Called(connectID, cacheMode)
	return args.Error(0)
}

func (m *MockV2) GetProfilePartition(partition UuidPartition) ([]string, error) {
	args := m.Called(partition)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockV2) GetUptID(profileId string) (string, error) {
	args := m.Called(profileId)
	return args.String(0), args.Error(1)
}

func (m *MockV2) UpdateVersion(version feature.FeatureSetVersion) error {
	args := m.Called(version)
	return args.Error(0)
}

func (m *MockV2) GetVersion() (feature.FeatureSetVersion, error) {
	args := m.Called()
	return args.Get(0).(feature.FeatureSetVersion), args.Error(1)
}

func (m *MockV2) InvalidateStitchingRuleSetsCache() error {
	args := m.Called()
	return args.Error(0)
}
