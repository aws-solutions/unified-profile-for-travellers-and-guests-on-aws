// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"testing"

	customerprofileslcs "tah/upt/source/storage"

	"github.com/google/uuid"
)

func TestRebuildDynamoCacheMain(t *testing.T) {
	// Set up mocks
	cpMock := customerprofileslcs.InitMockV2()

	// Set Env Variables
	testDomainName := "test_domain_name"
	cacheType := "DYNAMO"
	partitionLowerBound := uuid.New()
	partitionUpperBound := uuid.New()
	lowerBoundString := partitionLowerBound.String()
	upperBoundString := partitionUpperBound.String()
	t.Setenv("DOMAIN_NAME", testDomainName)
	t.Setenv("CACHE_TYPE", cacheType)
	t.Setenv("PARTITION_LOWER_BOUND", lowerBoundString)
	t.Setenv("PARTITION_UPPER_BOUND", upperBoundString)

	partition := customerprofileslcs.UuidPartition{
		LowerBound: partitionLowerBound,
		UpperBound: partitionUpperBound,
	}

	testId1 := "test_id_1"
	testId2 := "test_id_2"
	testConnectIds := []string{testId1, testId2}

	cacheMode := customerprofileslcs.DYNAMO_MODE

	// Create Mocks
	cpMock.On("SetDomain", testDomainName).Return(nil)
	cpMock.On("GetProfilePartition", partition).Return(testConnectIds, nil)
	cpMock.On("CacheProfile", testId1, cacheMode).Return(nil)
	cpMock.On("CacheProfile", testId2, cacheMode).Return(nil)

	// Call Handler
	HandleWithServices(cpMock)
}

func TestRebuildCPCacheMain(t *testing.T) {
	// Set up mocks
	cpMock := customerprofileslcs.InitMockV2()

	// Set Env Variables
	testDomainName := "test_domain_name"
	cacheType := "CUSTOMER_PROFILES"
	partitionLowerBound := uuid.New()
	partitionUpperBound := uuid.New()
	lowerBoundString := partitionLowerBound.String()
	upperBoundString := partitionUpperBound.String()
	t.Setenv("DOMAIN_NAME", testDomainName)
	t.Setenv("CACHE_TYPE", cacheType)
	t.Setenv("PARTITION_LOWER_BOUND", lowerBoundString)
	t.Setenv("PARTITION_UPPER_BOUND", upperBoundString)

	partition := customerprofileslcs.UuidPartition{
		LowerBound: partitionLowerBound,
		UpperBound: partitionUpperBound,
	}

	testId1 := "test_id_1"
	testId2 := "test_id_2"
	testConnectIds := []string{testId1, testId2}

	cacheMode := customerprofileslcs.CUSTOMER_PROFILES_MODE

	// Create Mocks
	cpMock.On("SetDomain", testDomainName).Return(nil)
	cpMock.On("GetProfilePartition", partition).Return(testConnectIds, nil)
	cpMock.On("CacheProfile", testId1, cacheMode).Return(nil)
	cpMock.On("CacheProfile", testId2, cacheMode).Return(nil)

	// Call Handler
	HandleWithServices(cpMock)
}
