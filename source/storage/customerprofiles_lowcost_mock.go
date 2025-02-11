// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"fmt"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/ucp-common/src/feature"
	"time"

	"github.com/google/uuid"
)

type MockCustomerProfileLowCostConfig struct {
	Domain       *Domain
	Domains      *[]Domain
	Profile      *profilemodel.Profile
	Profiles     *[]profilemodel.Profile
	ProfilesByID map[string]profilemodel.Profile
	Mappings     *[]ObjectMapping
	Merged       []string
	Objects      []profilemodel.ProfileObject
	Failures     int //number of failures before success
	MaxFailures  int //number of failures before success
}

// Provides static check for the interface implementation without allocating memory at runtime
var _ ICustomerProfileLowCostConfig = &MockCustomerProfileLowCostConfig{}

func InitMock(
	domain *Domain,
	domains *[]Domain,
	profile *profilemodel.Profile,
	profiles *[]profilemodel.Profile,
	mappings *[]ObjectMapping,
) *MockCustomerProfileLowCostConfig {
	profilesById := map[string]profilemodel.Profile{}
	if profiles != nil && len(*profiles) > 0 {
		for _, p := range *profiles {
			profilesById[p.ProfileId] = p
		}
	}
	return &MockCustomerProfileLowCostConfig{
		Domain:       domain,
		Domains:      domains,
		Profile:      profile,
		Profiles:     profiles,
		ProfilesByID: profilesById,
		Mappings:     mappings,
		Merged:       []string{},
	}
}

func (c *MockCustomerProfileLowCostConfig) SetTx(tx core.Transaction) {
}

func (c *MockCustomerProfileLowCostConfig) CreateMapping(name string, description string, fieldMappings []FieldMapping) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) CreateDomainWithQueue(
	name string,
	kmsArn string,
	tags map[string]string,
	queueUrl string,
	matchS3Name string,
	options DomainOptions,
) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) ListDomains() ([]Domain, error) {
	return *c.Domains, nil
}

func (c *MockCustomerProfileLowCostConfig) DeleteDomain() error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) GetProfile(
	id string,
	objectTypeNames []string,
	pagination []PaginationOptions,
) (profilemodel.Profile, error) {
	if c.Failures < c.MaxFailures {
		c.Failures++
		return profilemodel.Profile{}, fmt.Errorf("mock failure %v our of %v", c.Failures, c.MaxFailures)
	}
	if len(c.ProfilesByID) > 0 {
		return c.ProfilesByID[id], nil
	}
	if c.Profile == nil {
		return profilemodel.Profile{}, ErrProfileNotFound
	}
	return *c.Profile, nil
}

func (c *MockCustomerProfileLowCostConfig) SearchProfiles(key string, values []string) ([]profilemodel.Profile, error) {
	if c.Failures < c.MaxFailures {
		c.Failures++
		return []profilemodel.Profile{}, fmt.Errorf("mock failure %v our of %v", c.Failures, c.MaxFailures)
	}
	return *c.Profiles, nil
}

func (c *MockCustomerProfileLowCostConfig) AdvancedProfileSearch(searchCriteria []BaseCriteria) ([]profilemodel.Profile, error) {
	if c.Failures < c.MaxFailures {
		c.Failures++
		return []profilemodel.Profile{}, fmt.Errorf("mock failure %v our of %v", c.Failures, c.MaxFailures)
	}
	return *c.Profiles, nil
}

func (c *MockCustomerProfileLowCostConfig) DeleteProfile(id string, options ...DeleteProfileOptions) error {
	c.Profiles = &[]profilemodel.Profile{}
	c.ProfilesByID = map[string]profilemodel.Profile{}
	c.Profile = nil
	return nil
}

func (c *MockCustomerProfileLowCostConfig) PutIntegration(
	flowNamePrefix, objectName, bucketName, bucketPrefix string,
	fieldMappings []FieldMapping,
	startTime time.Time,
) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) GetDomain() (Domain, error) {
	return *c.Domain, nil
}

func (c *MockCustomerProfileLowCostConfig) GetMappings() ([]ObjectMapping, error) {
	return *c.Mappings, nil
}

func (c *MockCustomerProfileLowCostConfig) GetProfileLevelFields() ([]string, error) {
	return []string{}, nil
}

func (c *MockCustomerProfileLowCostConfig) GetObjectLevelFields() (map[string][]string, error) {
	return map[string][]string{}, nil
}

func (c *MockCustomerProfileLowCostConfig) DeleteDomainByName(name string) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) MergeProfiles(
	profileId string,
	profileIdToMerge string,
	mergeContext ProfileMergeContext,
) (string, error) {
	c.Merged = append(c.Merged, profileId)
	c.Merged = append(c.Merged, profileIdToMerge)
	return "", nil
}

func (c *MockCustomerProfileLowCostConfig) MergeMany(pairs []ProfilePair) (string, error) {
	return "", nil
}

func (c *MockCustomerProfileLowCostConfig) SetDomain(domain string) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) PutProfileObject(object, objectTypeName string) error {
	return nil
}

func (m *MockCustomerProfileLowCostConfig) RunRuleBasedIdentityResolution(
	object, objectTypeName, connectID string,
	ruleSet RuleSet,
) (bool, error) {
	return false, nil
}

func (c *MockCustomerProfileLowCostConfig) IsThrottlingError(err error) bool {
	return false
}

func (c *MockCustomerProfileLowCostConfig) GetProfileId(profileKey, profileId string) (string, error) {
	if c.Profile != nil {
		return c.Profile.ProfileId, nil
	}
	return "", nil
}

func (c *MockCustomerProfileLowCostConfig) GetSpecificProfileObject(
	objectKey string,
	objectId string,
	profileId string,
	objectTypeName string,
) (profilemodel.ProfileObject, error) {
	for _, object := range c.Objects {
		if object.ID == objectId {
			return object, nil
		}
	}
	return profilemodel.ProfileObject{}, nil
}

func (c *MockCustomerProfileLowCostConfig) CreateEventStream(domainName, streamName, streamArn string) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) DeleteEventStream(domainName, streamName string) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) ActivateIdResRuleSet() error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) SaveIdResRuleSet([]Rule) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) ListIdResRuleSets(bool) ([]RuleSet, error) {
	return []RuleSet{}, nil
}

func (c *MockCustomerProfileLowCostConfig) ActivateCacheRuleSet() error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) SaveCacheRuleSet([]Rule) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) ListCacheRuleSets(bool) ([]RuleSet, error) {
	return []RuleSet{}, nil
}

func (c *MockCustomerProfileLowCostConfig) GetActiveRuleSet() (RuleSet, error) {
	return RuleSet{}, nil
}

func (c *MockCustomerProfileLowCostConfig) EnableRuleBasedIdRes() error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) DisableRuleBasedIdRes() error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) ToggleAiIdentityResolution(domainName string, s3BucketName string, roleArn string) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) EnableEntityResolution(
	domainName string,
	s3BucketName string,
	roleArn string,
	outputBucket string,
) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) DisableEntityResolution(domainName string) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) BuildInteractionHistory(
	domain string,
	interactionId string,
	objectTypeName string,
	connectId string,
) ([]ProfileMergeContext, error) {
	return []ProfileMergeContext{}, nil
}

func (c *MockCustomerProfileLowCostConfig) MergeProfilesLowCost(
	profileId string,
	profileIdToMerge string,
	mergeContext ProfileMergeContext,
) (string, error) {
	return "", nil
}

func (c *MockCustomerProfileLowCostConfig) UnmergeProfiles(
	toUnmergeConnectID string,
	mergedIntoConnectID string,
	interactionIdToUnmerge string,
	interactionType string,
) (string, error) {
	c.Merged = append(c.Merged, toUnmergeConnectID)
	c.Merged = append(c.Merged, mergedIntoConnectID)
	c.Merged = append(c.Merged, interactionIdToUnmerge)
	c.Merged = append(c.Merged, interactionType)
	return "", nil
}

func (c *MockCustomerProfileLowCostConfig) FindCurrentParentOfProfile(domain, originalConnectId string) (string, error) {
	return "", nil
}

func (c *MockCustomerProfileLowCostConfig) GetInteractionTable(
	domain, objectType string,
	partition UuidPartition,
	lastId uuid.UUID,
	limit int,
) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *MockCustomerProfileLowCostConfig) ClearDynamoCache() error {
	return nil
}

func (m *MockCustomerProfileLowCostConfig) ClearCustomerProfileCache(tags map[string]string, matchS3Name string) error {
	return nil
}

func (m *MockCustomerProfileLowCostConfig) CacheProfile(connectID string, cacheMode CacheMode) error {
	return nil
}

func (m *MockCustomerProfileLowCostConfig) GetProfilePartition(partition UuidPartition) ([]string, error) {
	return []string{}, nil
}

func (c *MockCustomerProfileLowCostConfig) GetUptID(profileId string) (string, error) {
	return profileId, nil
}

func (c *MockCustomerProfileLowCostConfig) UpdateVersion(feature.FeatureSetVersion) error {
	return nil
}

func (c *MockCustomerProfileLowCostConfig) GetVersion() (feature.FeatureSetVersion, error) {
	return feature.FeatureSetVersion{
		Version:           feature.LATEST_FEATURE_SET_VERSION,
		CompatibleVersion: feature.LATEST_FEATURE_SET_VERSION,
	}, nil
}

func (m *MockCustomerProfileLowCostConfig) InvalidateStitchingRuleSetsCache() error {
	return nil
}
