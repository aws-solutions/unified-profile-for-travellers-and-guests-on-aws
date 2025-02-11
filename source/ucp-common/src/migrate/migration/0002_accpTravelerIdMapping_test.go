// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"context"
	"errors"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	"tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/feature"
	"testing"

	"github.com/stretchr/testify/mock"
)

// intentionally duplicated as a point-in-time snapshot of behavior
// if we change what business objects we create mappings for, we will need to decide how this migration should work
var businessObjectNames = []string{
	admin.ACCP_RECORD_AIR_BOOKING,
	admin.ACCP_RECORD_EMAIL_HISTORY,
	admin.ACCP_RECORD_PHONE_HISTORY,
	admin.ACCP_RECORD_AIR_LOYALTY,
	admin.ACCP_RECORD_LOYALTY_TX,
	admin.ACCP_RECORD_CLICKSTREAM,
	admin.ACCP_RECORD_GUEST_PROFILE,
	admin.ACCP_RECORD_HOTEL_LOYALTY,
	admin.ACCP_RECORD_HOTEL_BOOKING,
	admin.ACCP_RECORD_PAX_PROFILE,
	admin.ACCP_RECORD_HOTEL_STAY_MAPPING,
	admin.ACCP_RECORD_CSI,
	admin.ACCP_RECORD_ANCILLARY,
	admin.ACCP_RECORD_ALTERNATE_PROFILE_ID,
}

const connectIdAttribute = "external_connect_id"
const uniqueIdAttribute = "external_unique_object_key"

// intentionally duplicated as a point-in-time snapshot of behavior
// if we change what mappings we create, we will need to decide how this migration should work
var updatedMappings = []customerprofiles.FieldMapping{
	{
		Type:    "STRING",
		Source:  "_source.timestamp",
		Target:  "timestamp",
		Indexes: []string{},
		KeyOnly: true,
	},
	{
		Type:        "STRING",
		Source:      "_source.traveller_id",
		Target:      "traveller_id",
		Indexes:     []string{},
		Searcheable: true,
		KeyOnly:     true,
	},
	{
		Type:        "STRING",
		Source:      "_source." + connectIdAttribute,
		Target:      "_profile.Attributes." + connectIdAttribute,
		Indexes:     []string{"PROFILE"},
		Searcheable: true,
	},
	{
		Type:    "STRING",
		Source:  "_source." + uniqueIdAttribute,
		Target:  uniqueIdAttribute,
		Indexes: []string{"UNIQUE"},
		KeyOnly: true,
	},
}

// intentionally duplicated as a point-in-time snapshot of behavior
// if we change what mappings we create, we will need to decide how this migration should work
var oldMappings = []customerprofiles.FieldMapping{
	{
		Type:    "STRING",
		Source:  "_source.timestamp",
		Target:  "timestamp",
		Indexes: []string{},
		KeyOnly: true,
	},
	{
		Type:        "STRING",
		Source:      "_source." + connectIdAttribute,
		Target:      "_profile.Attributes." + connectIdAttribute,
		Indexes:     []string{"PROFILE"},
		Searcheable: true,
	},
	{
		Type:    "STRING",
		Source:  "_source." + uniqueIdAttribute,
		Target:  uniqueIdAttribute,
		Indexes: []string{"UNIQUE"},
		KeyOnly: true,
	},
}

func TestAccpTravelerIdMappingMigrationInvalidAccp(t *testing.T) {
	t.Parallel()

	_, err := InitAccpTravelerIdMappingMigration(true, nil, core.NewTransaction("", "", core.LogLevelDebug))
	if err == nil {
		t.Error("expected error initializing migration")
	}
}

func TestAccpTravelerIdMappingMigration(t *testing.T) {
	t.Parallel()

	mockAccp := customerprofiles.InitMock()
	defer mockAccp.AssertExpectations(t)

	for _, name := range businessObjectNames {
		mockAccp.On("CreateMapping", name, mock.AnythingOfType("string"), updatedMappings).Return(nil)
	}

	migration, err := InitAccpTravelerIdMappingMigration(true, mockAccp, core.NewTransaction("", "", core.LogLevelDebug))
	if err != nil {
		t.Fatalf("error initializing migration: %v", err)
	}

	version, err := migration.Up(context.Background())
	if err != nil {
		t.Fatalf("error running up migration: %v", err)
	}

	expectedVersion := feature.SimpleVersion(migration.VersionIntroduced())
	if version != expectedVersion {
		t.Errorf("expected version %v, got version %v", expectedVersion, version)
	}

	for _, name := range businessObjectNames {
		mockAccp.On("CreateMapping", name, mock.AnythingOfType("string"), oldMappings).Return(nil)
	}

	version, err = migration.Down(context.Background())
	if err != nil {
		t.Fatalf("error running down migration: %v", err)
	}

	expectedVersion = feature.SimpleVersion(migration.VersionIntroduced() - 1)
	if version != expectedVersion {
		t.Errorf("expected version %v, got %v", expectedVersion, version)
	}
}

func TestAccpTravelerIdMappingMigrationDisabled(t *testing.T) {
	t.Parallel()

	mockAccp := customerprofiles.InitMock()
	defer mockAccp.AssertExpectations(t)

	migration, err := InitAccpTravelerIdMappingMigration(false, mockAccp, core.NewTransaction("", "", core.LogLevelDebug))
	if err != nil {
		t.Fatalf("error initializing migration: %v", err)
	}

	version, err := migration.Up(context.Background())
	if err != nil {
		t.Fatalf("error running up migration: %v", err)
	}

	expectedVersion := feature.SimpleVersion(migration.VersionIntroduced())
	if version != expectedVersion {
		t.Errorf("expected version %v, got version %v", expectedVersion, version)
	}

	version, err = migration.Down(context.Background())
	if err != nil {
		t.Fatalf("error running down migration: %v", err)
	}

	expectedVersion = feature.SimpleVersion(migration.VersionIntroduced() - 1)
	if version != expectedVersion {
		t.Errorf("expected version %v, got %v", expectedVersion, version)
	}
}

func TestAccpTravelerIdMappingMigrationUpRollback(t *testing.T) {
	t.Parallel()

	mockAccp := customerprofiles.InitMock()
	defer mockAccp.AssertExpectations(t)

	mockAccp.On("CreateMapping", mock.AnythingOfType("string"), mock.AnythingOfType("string"), updatedMappings).Return(errors.New("error"))
	for _, name := range businessObjectNames {
		mockAccp.On("CreateMapping", name, mock.AnythingOfType("string"), oldMappings).Return(nil)
	}

	migration, err := InitAccpTravelerIdMappingMigration(true, mockAccp, core.NewTransaction("", "", core.LogLevelDebug))
	if err != nil {
		t.Fatalf("error initializing migration: %v", err)
	}

	version, err := migration.Up(context.Background())
	if err == nil {
		t.Fatal("expected error running up migration")
	}

	expectedVersion := feature.SimpleVersion(migration.VersionIntroduced() - 1)
	if version != expectedVersion {
		t.Errorf("expected version %v, got version %v", expectedVersion, version)
	}
}

func TestAccpTravelerIdMappingMigrationDownRollback(t *testing.T) {
	t.Parallel()

	mockAccp := customerprofiles.InitMock()
	defer mockAccp.AssertExpectations(t)

	mockAccp.On("CreateMapping", mock.AnythingOfType("string"), mock.AnythingOfType("string"), oldMappings).Return(errors.New("error"))
	for _, name := range businessObjectNames {
		mockAccp.On("CreateMapping", name, mock.AnythingOfType("string"), updatedMappings).Return(nil)
	}

	migration, err := InitAccpTravelerIdMappingMigration(true, mockAccp, core.NewTransaction("", "", core.LogLevelDebug))
	if err != nil {
		t.Fatalf("error initializing migration: %v", err)
	}

	version, err := migration.Down(context.Background())
	if err == nil {
		t.Fatal("expected error running up migration")
	}

	expectedVersion := feature.SimpleVersion(migration.VersionIntroduced())
	if version != expectedVersion {
		t.Errorf("expected version %v, got version %v", expectedVersion, version)
	}
}
