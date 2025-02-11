// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Framework for performing migrations on domains
package migrate

import (
	"context"
	"errors"
	"fmt"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/feature"
	"tah/upt/source/ucp-common/src/migrate/migration"
)

// Action to perform if a multi-step migration fails
type OnFailureOption int

const (
	// Attempt to rollback the domain to its initial version before the migration began
	RollbackOnFailure OnFailureOption = iota
	// Abort the migration
	AbortOnFailure
)

type MigrationFramework struct {
	version    feature.FeatureSetVersion
	migrations map[int]migration.Migration
	tx         core.Transaction
	onFailure  OnFailureOption
	lcs        customerprofileslcs.ICustomerProfileLowCostConfig
}

type migrationDirection int

const (
	up migrationDirection = iota
	down
)

// LCS must have the domain set to the target domain
func InitMigrationFramework(
	lcs customerprofileslcs.ICustomerProfileLowCostConfig,
	migrations map[int]migration.Migration,
	tx core.Transaction,
	onFailure OnFailureOption,
) (*MigrationFramework, error) {
	if lcs == nil {
		return &MigrationFramework{}, errors.New("error initializing migration framework with nil ACCP client")
	}

	version, err := lcs.GetVersion()
	if err != nil {
		return &MigrationFramework{}, fmt.Errorf("error getting domain version while initializing migration framework: %w", err)
	}

	// prevent initialization if the domain is outside the scope of known versions
	bounds := getBounds(migrations)
	if version.Version < bounds.LowerBound-1 {
		return &MigrationFramework{}, fmt.Errorf(
			"error initializing migration framework with version %d, earliest version is %d",
			version,
			bounds.LowerBound-1,
		)
	}
	if version.Version > bounds.UpperBound {
		return &MigrationFramework{}, fmt.Errorf(
			"error initializing migration framework with version %d, newest version is %d",
			version,
			bounds.UpperBound,
		)
	}

	return &MigrationFramework{
		version:    version,
		migrations: migrations,
		tx:         tx,
		onFailure:  onFailure,
		lcs:        lcs,
	}, nil
}

type migrationBounds struct {
	LowerBound int
	UpperBound int
}

func getBounds(migrations map[int]migration.Migration) migrationBounds {
	bounds := migrationBounds{}
	first := true
	for i := range migrations {
		if first {
			bounds.LowerBound = i
			bounds.UpperBound = i
			first = false
		} else {
			if i < bounds.LowerBound {
				bounds.LowerBound = i
			}
			if i > bounds.UpperBound {
				bounds.UpperBound = i
			}
		}
	}
	return bounds
}

func (mf *MigrationFramework) Migrate(ctx context.Context, target int) error {
	// don't start a migration unless we're sure we have each incremental migration
	if missingMigrations := getMissingMigrations(mf.version.Version, target, mf.migrations); len(missingMigrations) > 0 {
		return fmt.Errorf(
			"error starting migration from version %d to %d, missing migrations: %v",
			mf.version.Version,
			target,
			missingMigrations,
		)
	}

	mf.tx.Info("beginning migration from version %d to version %d", mf.version.Version, target)

	originalVersion := mf.version.Version
	rollingBack := false
	var errorThatCausedRollback error
	for mf.version.Version != target {
		var direction migrationDirection
		var nextVersion int
		var nextMigration migration.Migration
		var migrationFound bool

		if target > mf.version.Version {
			direction = up
			nextVersion = mf.version.Version + 1
			nextMigration, migrationFound = mf.migrations[nextVersion]
			if !migrationFound {
				// something went wrong, this should have been sanity-checked
				return fmt.Errorf("no migration found for version %d", nextVersion)
			}
		} else {
			direction = down
			nextVersion = mf.version.Version - 1
			nextMigration, migrationFound = mf.migrations[mf.version.Version]
			if !migrationFound {
				// something went wrong, this should have been sanity-checked
				return fmt.Errorf("no migration found for version %d", mf.version.Version)
			}
		}

		mf.tx.Info("beginning migration to version %d", nextVersion)
		// todo: get the actual version from the migration itself

		nextFeatureSetVersion, err := performMigration(ctx, mf.tx, nextMigration, direction)
		// todo: sanity check on version returned from migration
		if err != nil {
			if mf.onFailure == RollbackOnFailure {
				if !rollingBack {
					mf.tx.Error("error migrating, attempting rollback: %v", err)
					errorThatCausedRollback = err
					rollingBack = true
					target = originalVersion
					continue
				} else {
					return errors.Join(fmt.Errorf("error attempting to roll back migration: %v", err), errorThatCausedRollback)
				}
			} else {
				return fmt.Errorf("failed migration: %w", err)
			}
		}

		mf.tx.Info("completed migration, updating domain configuration to version %d", nextVersion)

		err = mf.lcs.UpdateVersion(nextFeatureSetVersion)
		if err != nil {
			return fmt.Errorf("error updating domain version after migration: %v", err)
		}

		mf.version = nextFeatureSetVersion
	}

	if rollingBack {
		return errorThatCausedRollback
	} else {
		mf.tx.Info("completed migration to version %d", target)
	}

	return nil
}

func performMigration(
	ctx context.Context,
	tx core.Transaction,
	migration migration.Migration,
	direction migrationDirection,
) (version feature.FeatureSetVersion, err error) {
	defer func() {
		if r := recover(); r != nil {
			tx.Info("recovered from panic in migration: %v", r)
			if castError, ok := r.(error); ok {
				err = castError
			} else {
				panic(r)
			}
		}
	}()

	switch direction {
	case up:
		version, err = migration.Up(ctx)
	case down:
		version, err = migration.Down(ctx)
	default:
		return version, fmt.Errorf("invalid migration direction: %d, expected %d (up) or %d (down)", direction, up, down)
	}

	if err != nil {
		return version, fmt.Errorf("failed migration: %w", err)
	}
	return version, nil
}

func getMissingMigrations(originalVersion, targetVersion int, migrations map[int]migration.Migration) []int {
	var lowerVersion, higherVersion int
	if targetVersion > originalVersion {
		lowerVersion = originalVersion
		higherVersion = targetVersion
	} else {
		lowerVersion = targetVersion
		higherVersion = originalVersion
	}
	var result []int
	version := lowerVersion + 1
	for version <= higherVersion {
		if _, ok := migrations[version]; !ok {
			result = append(result, version)
		}
		version++
	}
	return result
}
