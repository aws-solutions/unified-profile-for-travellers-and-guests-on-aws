// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"context"
	"errors"
	"fmt"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	"tah/upt/source/ucp-common/src/feature"
	"tah/upt/source/ucp-common/src/migrate/migration"
)

const migrationErrMsg string = "error initializing migration: %v"

type Services struct {
	Accp customerprofiles.ICustomerProfileConfig
}

func InitMigrations(ctx context.Context, tx core.Transaction, services Services, accpEnabled bool) (map[int]migration.Migration, error) {
	if services.Accp == nil {
		return map[int]migration.Migration{}, errors.New("error initializing migration framework with nil ACCP client")
	}

	accpTravelerIdMappingMigration, err := migration.InitAccpTravelerIdMappingMigration(accpEnabled, services.Accp, tx)
	if err != nil {
		tx.Error(migrationErrMsg, err)
		return map[int]migration.Migration{}, err
	}

	alternateProfileIdMigration, err := migration.InitAlternateProfileIdMigration(tx)
	if err != nil {
		tx.Error(migrationErrMsg, err)
		return map[int]migration.Migration{}, err
	}

	accpExportStreamMigration, err := migration.InitAccpExportStreamMigration(tx)
	if err != nil {
		tx.Error(migrationErrMsg, err)
		return map[int]migration.Migration{}, err
	}

	migrations := map[int]migration.Migration{
		feature.AccpTravelerIdMappingIntroducedVersion: accpTravelerIdMappingMigration,
		feature.AlternateProfileIdIntroducedVersion:    alternateProfileIdMigration,
		feature.AccpExportStreamVersionIntroduced:      accpExportStreamMigration,
	}

	if len(migrations) != feature.LATEST_FEATURE_SET_VERSION-feature.INITIAL_FEATURE_SET_VERSION {
		return map[int]migration.Migration{}, fmt.Errorf(
			"expected %d migrations, got %d",
			feature.LATEST_FEATURE_SET_VERSION-1,
			len(migrations),
		)
	}
	i := feature.INITIAL_FEATURE_SET_VERSION + 1
	for i <= feature.LATEST_FEATURE_SET_VERSION {
		if _, ok := migrations[i]; !ok {
			return map[int]migration.Migration{}, fmt.Errorf("missing migration for version %d", i)
		}
		i++
	}

	return migrations, nil
}
