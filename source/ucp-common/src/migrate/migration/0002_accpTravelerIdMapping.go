// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"context"
	"errors"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	common "tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/feature"
)

type AccpTravelerIdMappingMigration struct {
	accpEnabled bool
	accp        customerprofiles.ICustomerProfileConfig
	tx          core.Transaction
}

var _ Migration = &AccpTravelerIdMappingMigration{}

func InitAccpTravelerIdMappingMigration(
	accpMode bool,
	accp customerprofiles.ICustomerProfileConfig,
	tx core.Transaction,
) (*AccpTravelerIdMappingMigration, error) {
	if accp == nil {
		return nil, errors.New("error initializing migration with nil ACCP client")
	}
	return &AccpTravelerIdMappingMigration{
		accpEnabled: accpMode,
		accp:        accp,
		tx:          tx,
	}, nil
}

func (*AccpTravelerIdMappingMigration) VersionIntroduced() int {
	return feature.AccpTravelerIdMappingIntroducedVersion
}

func (m *AccpTravelerIdMappingMigration) Up(ctx context.Context) (version feature.FeatureSetVersion, err error) {
	versionIntroduced := m.VersionIntroduced()
	previousVersion := versionIntroduced - 1
	// dtodo: is there a way to handle this automatically, separately from the actual migration logic?

	version = feature.SimpleVersion(versionIntroduced)
	if !m.accpEnabled {
		m.tx.Info("ACCP mode disabled, nothing to do")
		return version, nil
	}

	defer func() {
		if err != nil {
			m.tx.Error("error during migration, rolling back")
			rollbackErr := putOldMappings(ctx, m.tx, m.accp)
			if rollbackErr != nil {
				m.tx.Error("error rolling back migration: %v", rollbackErr)
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	err = putUpdatedMappings(ctx, m.tx, m.accp)
	if err != nil {
		return feature.SimpleVersion(previousVersion), err
	}

	return version, nil
}

func (m *AccpTravelerIdMappingMigration) Down(ctx context.Context) (version feature.FeatureSetVersion, err error) {
	versionIntroduced := m.VersionIntroduced()
	previousVersion := versionIntroduced - 1
	// dtodo: is there a way to handle this automatically, separately from the actual migration logic?

	version = feature.SimpleVersion(previousVersion)
	if !m.accpEnabled {
		m.tx.Info("ACCP mode disabled, nothing to do")
		return version, nil
	}

	defer func() {
		if err != nil {
			m.tx.Error("error during migration, rolling back")
			rollbackErr := putUpdatedMappings(ctx, m.tx, m.accp)
			if rollbackErr != nil {
				m.tx.Error("error rolling back migration: %v", rollbackErr)
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	err = putOldMappings(ctx, m.tx, m.accp)
	if err != nil {
		return feature.SimpleVersion(versionIntroduced), err
	}

	return version, nil
}

func putUpdatedMappings(_ context.Context, tx core.Transaction, accp customerprofiles.ICustomerProfileConfig) error {
	fieldMappings := customerprofiles.BuildGenericIntegrationFieldMappingWithTravelerId()

	for _, businessObject := range common.ACCP_RECORDS {
		accpRecName := businessObject.Name
		tx.Debug("updating mapping for %s", accpRecName)
		err := accp.CreateMapping(accpRecName, "Primary Mapping for the "+accpRecName+" object", fieldMappings)
		if err != nil {
			tx.Error("error updating mapping: %v", err)
			return err
		}
	}

	return nil
}

func putOldMappings(_ context.Context, tx core.Transaction, accp customerprofiles.ICustomerProfileConfig) error {
	fieldMappings := customerprofiles.BuildGenericIntegrationFieldMapping()

	for _, businessObject := range common.ACCP_RECORDS {
		accpRecName := businessObject.Name
		tx.Debug("updating mapping for %s", accpRecName)
		err := accp.CreateMapping(accpRecName, "Primary Mapping for the "+accpRecName+" object", fieldMappings)
		if err != nil {
			tx.Error("error updating mapping: %v", err)
			return err
		}
	}

	return nil
}
