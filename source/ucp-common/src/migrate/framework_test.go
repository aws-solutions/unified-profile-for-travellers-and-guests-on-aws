// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"context"
	"errors"
	"slices"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/feature"
	"tah/upt/source/ucp-common/src/migrate/migration"
	"testing"
)

type DummyMigration struct {
	versionIntroduced int
}

func (m *DummyMigration) VersionIntroduced() int { return m.versionIntroduced }

func (m *DummyMigration) Up(ctx context.Context) (feature.FeatureSetVersion, error) {
	return feature.SimpleVersion(m.VersionIntroduced()), nil
}

func (m *DummyMigration) Down(ctx context.Context) (feature.FeatureSetVersion, error) {
	return feature.SimpleVersion(m.VersionIntroduced() - 1), nil
}

var _ migration.Migration = &DummyMigration{}

func InitDummyMigration(versionIntroduced int) *DummyMigration {
	return &DummyMigration{versionIntroduced: versionIntroduced}
}

func makeDummyMigrationMap(versions ...int) map[int]migration.Migration {
	migrations := map[int]migration.Migration{}
	for _, i := range versions {
		migrations[i] = InitDummyMigration(i)
	}
	return migrations
}

func TestMigrationFrameworkUp(t *testing.T) {
	t.Parallel()

	mockLcs := customerprofileslcs.InitMockV2()
	defer mockLcs.AssertExpectations(t)
	initialVersion := 4
	mockLcs.On("GetVersion").Return(feature.SimpleVersion(initialVersion), nil)

	migrations := makeDummyMigrationMap(5, 6, 7)
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	fw, err := InitMigrationFramework(mockLcs, migrations, tx, AbortOnFailure)
	if err != nil {
		t.Fatalf("error initializing migration framework: %v", err)
	}

	mockLcs.On("UpdateVersion", feature.SimpleVersion(5)).Return(nil)
	mockLcs.On("UpdateVersion", feature.SimpleVersion(6)).Return(nil)
	mockLcs.On("UpdateVersion", feature.SimpleVersion(7)).Return(nil)

	err = fw.Migrate(context.Background(), 7)
	if err != nil {
		t.Fatalf("error migrating to target version: %v", err)
	}
}

func TestMigrationFrameworkDown(t *testing.T) {
	t.Parallel()

	mockLcs := customerprofileslcs.InitMockV2()
	defer mockLcs.AssertExpectations(t)
	initialVersion := 11
	mockLcs.On("GetVersion").Return(feature.SimpleVersion(initialVersion), nil)

	migrations := makeDummyMigrationMap(9, 10, 11, 12)
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	fw, err := InitMigrationFramework(mockLcs, migrations, tx, AbortOnFailure)
	if err != nil {
		t.Fatalf("error initializing migration framework: %v", err)
	}

	mockLcs.On("UpdateVersion", feature.SimpleVersion(10)).Return(nil)
	mockLcs.On("UpdateVersion", feature.SimpleVersion(9)).Return(nil)

	err = fw.Migrate(context.Background(), 9)
	if err != nil {
		t.Fatalf("error migrating to target version: %v", err)
	}
}

func TestMigrationFrameworkInvalidAccp(t *testing.T) {
	t.Parallel()

	migrations := map[int]migration.Migration{2: InitDummyMigration(2)}

	_, err := InitMigrationFramework(nil, migrations, core.NewTransaction("", "", core.LogLevelDebug), AbortOnFailure)
	if err == nil {
		t.Error("expected error initializing migration framework")
	}
}

func TestMigrationFrameworkGetVersionError(t *testing.T) {
	t.Parallel()

	mockLcs := customerprofileslcs.InitMockV2()
	defer mockLcs.AssertExpectations(t)
	mockLcs.On("GetVersion").Return(feature.FeatureSetVersion{}, errors.New("error"))

	migrations := makeDummyMigrationMap(10)

	_, err := InitMigrationFramework(mockLcs, migrations, core.NewTransaction("", "", core.LogLevelDebug), RollbackOnFailure)
	if err == nil {
		t.Error("expected error initializing migration framework")
	}
}

func TestMigrationFrameworkGetVersionTooOld(t *testing.T) {
	t.Parallel()

	mockLcs := customerprofileslcs.InitMockV2()
	defer mockLcs.AssertExpectations(t)
	mockLcs.On("GetVersion").Return(feature.SimpleVersion(4), nil)

	migrations := makeDummyMigrationMap(7)

	_, err := InitMigrationFramework(mockLcs, migrations, core.NewTransaction("", "", core.LogLevelDebug), AbortOnFailure)
	if err == nil {
		t.Error("expected error initializing migration framework")
	}
}

func TestMigrationFrameworkGetVersionTooNew(t *testing.T) {
	t.Parallel()

	mockLcs := customerprofileslcs.InitMockV2()
	defer mockLcs.AssertExpectations(t)
	mockLcs.On("GetVersion").Return(feature.SimpleVersion(12), nil)

	migrations := makeDummyMigrationMap(7)

	_, err := InitMigrationFramework(mockLcs, migrations, core.NewTransaction("", "", core.LogLevelDebug), AbortOnFailure)
	if err == nil {
		t.Error("expected error initializing migration framework")
	}
}

type FailingMigration struct {
	versionIntroduced int
}

func (m *FailingMigration) VersionIntroduced() int { return m.versionIntroduced }

func (m *FailingMigration) Up(ctx context.Context) (feature.FeatureSetVersion, error) {
	return feature.SimpleVersion(m.VersionIntroduced() - 1), errors.New("error")
}

func (m *FailingMigration) Down(ctx context.Context) (feature.FeatureSetVersion, error) {
	return feature.SimpleVersion(m.VersionIntroduced()), errors.New("error")
}

var _ migration.Migration = &FailingMigration{}

func InitFailingMigration(versionIntroduced int) *FailingMigration {
	return &FailingMigration{versionIntroduced: versionIntroduced}
}

func TestMigrationFrameworkAbort(t *testing.T) {
	t.Parallel()

	mockLcs := customerprofileslcs.InitMockV2()
	defer mockLcs.AssertExpectations(t)
	initialVersion := 1
	mockLcs.On("GetVersion").Return(feature.SimpleVersion(initialVersion), nil)

	migrations := map[int]migration.Migration{
		2: InitDummyMigration(2),
		3: InitFailingMigration(3),
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	mockLcs.On("UpdateVersion", feature.SimpleVersion(2)).Return(nil)

	fw, err := InitMigrationFramework(mockLcs, migrations, tx, AbortOnFailure)
	if err != nil {
		t.Fatalf("error initializing migration framework: %v", err)
	}

	err = fw.Migrate(context.Background(), 3)
	if err == nil {
		t.Fatal("expected error migrating")
	}
}

func TestMigrationFrameworkRollback(t *testing.T) {
	t.Parallel()

	mockLcs := customerprofileslcs.InitMockV2()
	defer mockLcs.AssertExpectations(t)
	initialVersion := 1
	mockLcs.On("GetVersion").Return(feature.SimpleVersion(initialVersion), nil)

	migrations := map[int]migration.Migration{
		2: InitDummyMigration(2),
		3: InitFailingMigration(3),
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	mockLcs.On("UpdateVersion", feature.SimpleVersion(2)).Return(nil)
	mockLcs.On("UpdateVersion", feature.SimpleVersion(1)).Return(nil)

	fw, err := InitMigrationFramework(mockLcs, migrations, tx, RollbackOnFailure)
	if err != nil {
		t.Fatalf("error initializing migration framework: %v", err)
	}

	err = fw.Migrate(context.Background(), 3)
	if err == nil {
		t.Fatal("expected error migrating")
	}
}

// todo: migration cases
// - migration cases w/mocks
// - migration panic recovery

func makeNilMigrationMap(versions ...int) map[int]migration.Migration {
	migrations := map[int]migration.Migration{}
	for _, i := range versions {
		migrations[i] = nil
	}
	return migrations
}

func TestGetBounds(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		migrations         map[int]migration.Migration
		expectedLowerBound int
		expectedUpperBound int
	}{
		"simple continuous": {
			migrations:         makeNilMigrationMap(0, 1, 2, 3),
			expectedLowerBound: 0,
			expectedUpperBound: 3,
		},
		"discontinuous": {
			migrations:         makeNilMigrationMap(1, 5, 6, 7),
			expectedLowerBound: 1,
			expectedUpperBound: 7,
		},
		"negative": {
			migrations:         makeNilMigrationMap(-13, -12, -11),
			expectedLowerBound: -13,
			expectedUpperBound: -11,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			bounds := getBounds(c.migrations)
			if bounds.LowerBound != c.expectedLowerBound {
				t.Errorf("expect lower bound to be %d, got %d", c.expectedLowerBound, bounds.LowerBound)
			}
			if bounds.UpperBound != c.expectedUpperBound {
				t.Errorf("expect upper bound to be %d, got %d", c.expectedUpperBound, bounds.UpperBound)
			}
		})
	}
}

func TestGetMissingMigrations(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		migrations map[int]migration.Migration
		initial    int
		target     int
		missing    []int
	}{
		"up simple": {
			migrations: makeNilMigrationMap(1, 2, 3, 4, 5),
			initial:    0,
			target:     5,
			missing:    []int{},
		},
		"up single missing": {
			migrations: makeNilMigrationMap(3, 4, 5, 6, 8),
			initial:    4,
			target:     8,
			missing:    []int{7},
		},
		"up too low": {
			migrations: makeNilMigrationMap(7, 8, 9),
			initial:    5,
			target:     8,
			missing:    []int{6},
		},
		"up too high": {
			migrations: makeNilMigrationMap(3, 4, 5),
			initial:    4,
			target:     6,
			missing:    []int{6},
		},
		"up multiple missing": {
			migrations: makeNilMigrationMap(10, 11, 13, 15, 17),
			initial:    10,
			target:     17,
			missing:    []int{12, 14, 16},
		},
		"down simple": {
			migrations: makeNilMigrationMap(91, 92, 93, 94),
			initial:    94,
			target:     90,
			missing:    []int{},
		},
		"down single missing": {
			migrations: makeNilMigrationMap(7, 8, 10, 11),
			initial:    11,
			target:     6,
			missing:    []int{9},
		},
		"down too low": {
			migrations: makeNilMigrationMap(7, 8, 9),
			initial:    8,
			target:     5,
			missing:    []int{6},
		},
		"down too high": {
			migrations: makeNilMigrationMap(3, 4, 5),
			initial:    6,
			target:     4,
			missing:    []int{6},
		},
		"down multiple missing": {
			migrations: makeNilMigrationMap(10, 11, 13, 15, 17),
			initial:    17,
			target:     10,
			missing:    []int{12, 14, 16},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			missing := getMissingMigrations(c.initial, c.target, c.migrations)
			if !slices.Equal(missing, c.missing) {
				t.Errorf("expect missing: %v, got: %v", c.missing, missing)
			}
		})
	}
}
