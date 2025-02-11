// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofiles

import (
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"time"

	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

var _ ICustomerProfileConfig = &Mock{}

func InitMock() *Mock {
	return &Mock{}
}

// CreateDomainWithQueue implements ICustomerProfileConfig.
func (m *Mock) CreateDomainWithQueue(
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
func (m *Mock) CreateEventStream(domainName string, streamName string, streamArn string) error {
	args := m.Called(domainName, streamName, streamArn)
	return args.Error(0)
}

// CreateMapping implements ICustomerProfileConfig.
func (m *Mock) CreateMapping(name string, description string, fieldMappings []FieldMapping) error {
	args := m.Called(name, description, fieldMappings)
	return args.Error(0)
}

// DeleteDomain implements ICustomerProfileConfig.
func (m *Mock) DeleteDomain() error {
	args := m.Called()
	return args.Error(0)
}

// DeleteDomainByName implements ICustomerProfileConfig.
func (m *Mock) DeleteDomainByName(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// DeleteEventStream implements ICustomerProfileConfig.
func (m *Mock) DeleteEventStream(domainName string, streamName string) error {
	args := m.Called(domainName, streamName)
	return args.Error(0)
}

// DeleteProfile implements ICustomerProfileConfig.
func (m *Mock) DeleteProfile(id string, options ...DeleteProfileOptions) error {
	args := m.Called(id, options)
	return args.Error(0)
}

// GetDomain implements ICustomerProfileConfig.
func (m *Mock) GetDomain() (Domain, error) {
	args := m.Called()
	return args.Get(0).(Domain), args.Error(1)
}

// GetIntegrations implements ICustomerProfileConfig.
func (m *Mock) GetIntegrations() ([]Integration, error) {
	args := m.Called()
	return args.Get(0).([]Integration), args.Error(1)
}

// GetMappings implements ICustomerProfileConfig.
func (m *Mock) GetMappings() ([]ObjectMapping, error) {
	args := m.Called()
	return args.Get(0).([]ObjectMapping), args.Error(1)
}

// GetProfile implements ICustomerProfileConfig.
func (m *Mock) GetProfile(
	id string,
	objectTypeNames []string,
	pagination []PaginationOptions,
) (profilemodel.Profile, error) {
	args := m.Called(id, objectTypeNames, pagination)
	return args.Get(0).(profilemodel.Profile), args.Error(1)
}

// GetProfileId implements ICustomerProfileConfig.
func (m *Mock) GetProfileId(profileKey string, profileId string) (string, error) {
	args := m.Called(profileKey, profileId)
	return args.String(0), args.Error(1)
}

// GetSpecificProfileObject implements ICustomerProfileConfig.
func (m *Mock) GetSpecificProfileObject(
	objectKey string,
	objectId string,
	profileId string,
	objectTypeName string,
) (profilemodel.ProfileObject, error) {
	args := m.Called(objectKey, objectId, profileId, objectTypeName)
	return args.Get(0).(profilemodel.ProfileObject), args.Error(1)
}

// IsThrottlingError implements ICustomerProfileConfig.
func (m *Mock) IsThrottlingError(err error) bool {
	args := m.Called(err)
	return args.Bool(0)
}

// ListDomains implements ICustomerProfileConfig.
func (m *Mock) ListDomains() ([]Domain, error) {
	args := m.Called()
	return args.Get(0).([]Domain), args.Error(1)
}

// MergeMany implements ICustomerProfileConfig.
func (m *Mock) MergeMany(pairs []ProfilePair) (string, error) {
	args := m.Called(pairs)
	return args.String(0), args.Error(1)
}

// MergeProfiles implements ICustomerProfileConfig.
func (m *Mock) MergeProfiles(profileId string, profileIdToMerge string) (string, error) {
	args := m.Called(profileId, profileIdToMerge)
	return args.String(0), args.Error(1)
}

// PutIntegration implements ICustomerProfileConfig.
func (m *Mock) PutIntegration(
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

// PutProfileAsObject implements ICustomerProfileConfig.
func (m *Mock) PutProfileAsObject(p profilemodel.Profile) error {
	args := m.Called(p)
	return args.Error(0)
}

// PutProfileObject implements ICustomerProfileConfig.
func (m *Mock) PutProfileObject(object string, objectTypeName string) error {
	args := m.Called(object, objectTypeName)
	return args.Error(0)
}

// PutProfileObjectFromLcs implements ICustomerProfileConfig.
func (m *Mock) PutProfileObjectFromLcs(p profilemodel.ProfileObject, lcsId string) error {
	args := m.Called(p, lcsId)
	return args.Error(0)
}

// SearchProfiles implements ICustomerProfileConfig.
func (m *Mock) SearchProfiles(key string, values []string) ([]profilemodel.Profile, error) {
	args := m.Called(key, values)
	return args.Get(0).([]profilemodel.Profile), args.Error(1)
}

// SetDomain implements ICustomerProfileConfig.
func (m *Mock) SetDomain(domain string) {
	m.Called(domain)
}

// SetTx implements ICustomerProfileConfig.
func (m *Mock) SetTx(tx core.Transaction) {
	m.Called(tx)
}
